package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Dmitrevicz/gometrics/internal/agent/config"
	"github.com/Dmitrevicz/gometrics/internal/model"
	"github.com/Dmitrevicz/gometrics/internal/retry"
	"github.com/Dmitrevicz/gometrics/internal/server"
	"github.com/Dmitrevicz/gometrics/pkg/encryptor"
	"golang.org/x/sync/errgroup"
)

type sender struct {
	reportInterval int
	url            string
	key            string
	hostIP         string
	batch          bool
	poller         *poller
	gopsutilPoller *gopsutilPoller
	client         *http.Client
	encryptor      *encryptor.Encryptor

	quit  chan struct{}
	timer *time.Timer

	// Задание 15-го инкремента реализовал через семафор
	//
	// > "Количество одновременно исходящих запросов на сервер нужно ограничивать «сверху»"
	Semaphore *Semaphore
}

func NewSender(cfg *config.Config, poller *poller, gopsutilPoller *gopsutilPoller) (*sender, error) {
	if cfg.RateLimit < 1 {
		cfg.RateLimit = 1
	}

	if cfg.HostIP == "" {
		log.Println("Couldn't detect host IP - X-Real-IP header will not be set on outgoing requests!")
	}

	var (
		encrypt *encryptor.Encryptor
		err     error
	)

	if cfg.CryptoKey != "" {
		encrypt, err = encryptor.NewEncryptor(cfg.CryptoKey)
		if err != nil {
			return nil, err
		}
	} else {
		// logger.Log.Warn("Empty CRYPTO_KEY was provided - encryption will be disabled!")
		log.Println("Empty CRYPTO_KEY was provided - encryption will be disabled!")
	}

	return &sender{
		reportInterval: cfg.ReportInterval,
		url:            cfg.ServerURL,
		key:            cfg.Key,
		hostIP:         cfg.HostIP,
		batch:          cfg.Batch,
		poller:         poller,
		gopsutilPoller: gopsutilPoller,
		client:         NewClientDefault(),
		Semaphore:      NewSemaphore(cfg.RateLimit),
		encryptor:      encrypt,
		quit:           make(chan struct{}),
	}, nil
}

func (s *sender) Start() {
	log.Println("Sender started")

	var ts time.Time
	sleepDuration := time.Second * time.Duration(s.reportInterval)
	s.timer = time.NewTimer(sleepDuration)

	for {
		select {
		case ts = <-s.timer.C:
			metrics := s.poller.AcquireMetrics()
			metrics.Merge(s.gopsutilPoller.AcquireMetrics())

			// iteration-12:
			// > Научите агент работать с использованием нового API (отправлять метрики батчами).
			if s.batch {
				s.SendBatched(metrics)
			} else {
				s.Send(metrics)
			}

			fmt.Println("send fired:", time.Since(ts))

			// I don't calculate delta-time, because current behaviour
			// is good enough right now.
			s.timer.Reset(sleepDuration)

		case <-s.quit:
			// stop the timer
			if !s.timer.Stop() {
				// drain the chanel (might not be needed here, but leave it be
				// as a kind of exercise)
				<-s.timer.C
			}

			log.Println("Sender timer stopped")
			return
		}
	}
}

// Shutdown stops sender's ticker and sends current data to server.
func (s *sender) Shutdown(ctx context.Context) error {
	log.Println("Stopping Sender")

	s.stop()

	wait := make(chan error, 1)
	go func() {
		metrics := s.poller.AcquireMetrics()
		metrics.Merge(s.gopsutilPoller.AcquireMetrics())
		s.SendBatched(metrics)
		close(wait)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("sender Shutdown failed: %v", ctx.Err())
	case <-wait:
		return nil
	}
}

// stop stops sender's timer.
func (s *sender) stop() {
	close(s.quit)
	log.Println("Sender stopped")
}

func (s *sender) Send(metrics Metrics) {
	log.Println("Metrics report started")

	// if metrics == nil {
	// 	log.Println("Error sending metrics: got nil as *Metrics, skip")
	// 	return
	// }

	ts := time.Now()
	g := new(errgroup.Group)

	for name, gauge := range metrics.Gauges {
		name := name
		gauge := gauge
		// make it async
		g.Go(func() error {
			err := s.sendMetrics(name, gauge)
			if err != nil {
				return fmt.Errorf("gauge update request failed: %w", err)
			}
			return nil
		})
	}

	for name, counter := range metrics.Counters {
		name := name
		counter := counter
		// make it async
		g.Go(func() error {
			err := s.sendMetrics(name, counter)
			if err != nil {
				return fmt.Errorf("counter update request failed: %w", err)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		log.Println("Got error while sending metric update request: " + err.Error())
	}

	log.Printf("Metrics have been sent (%d in %v)\n", len(metrics.Counters)+len(metrics.Gauges), time.Since(ts))
}

func (s *sender) prepareMetricsBatch(metrics Metrics) (batch []model.Metrics) {
	batch = make([]model.Metrics, 0, len(metrics.Gauges)+len(metrics.Counters))

	for name, val := range metrics.Gauges {
		val := val
		gauge := model.Metrics{
			MType: model.MetricTypeGauge,
			ID:    name,
			Value: (*float64)(&val),
		}
		batch = append(batch, gauge)
	}

	for name, val := range metrics.Counters {
		val := val
		counter := model.Metrics{
			MType: model.MetricTypeCounter,
			ID:    name,
			Delta: (*int64)(&val),
		}
		batch = append(batch, counter)
	}

	return batch
}

// SendBatched sends metrics to server in batch.
//
// > Научите агент работать с использованием нового API (отправлять метрики батчами).
func (s *sender) SendBatched(metrics Metrics) {
	log.Println("Metrics report started (batched)")

	// if metrics == nil {
	// 	log.Println("Error sending metrics: got nil as *Metrics, skip")
	// 	return
	// }

	ts := time.Now()

	batch := s.prepareMetricsBatch(metrics)

	retry := retry.NewRetrier(time.Second, 3)
	if err := retry.Do("send batched metrics", func() error {
		return s.sendBatched(batch)
	}); err != nil {
		log.Println("Got error while sending batched update request: " + err.Error())
	}

	log.Printf("Metrics have been sent (%d in %v)\n", len(metrics.Counters)+len(metrics.Gauges), time.Since(ts))
}

// DefaultHTTPClientTimeoutSeconds - custom default http client timeout in seconds.
const DefaultHTTPClientTimeoutSeconds = 10

// NewClientDefault returns *http.Client with DefaultHTTPClientTimeoutSeconds timeout set.
func NewClientDefault() *http.Client {
	return &http.Client{
		Timeout: time.Second * DefaultHTTPClientTimeoutSeconds,
	}
}

func (s *sender) sendGauge(name string, value model.Gauge) error {
	url := fmt.Sprintf("%s/update/gauge/%s/%f", s.url, name, value)
	resp, err := s.client.Post(url, "text/plain", nil)
	if err != nil {
		return fmt.Errorf("error while doing the request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error while reading the response bytes: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected response status code: %s, body: %s", resp.Status, string(body))
	}

	return nil
}

func (s *sender) sendCounter(name string, value model.Counter) error {
	url := fmt.Sprintf("%s/update/counter/%s/%d", s.url, name, value)
	resp, err := s.client.Post(url, "text/plain", nil)
	if err != nil {
		return fmt.Errorf("error while doing the request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error while reading the response bytes: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected response status code: %s, body: %s", resp.Status, string(body))
	}

	return nil
}

// sendMetrics - value must be either of type model.Counter or model.Gauge,
// otherwise error will be returned.
func (s *sender) sendMetrics(name string, value any) error {
	s.Semaphore.Acquire()
	defer s.Semaphore.Release()

	// configure struct to be sent in request body
	metrics := model.Metrics{
		ID: name,
	}

	switch v := value.(type) {
	case model.Gauge:
		f := float64(v)
		metrics.Value = &f
		metrics.MType = model.MetricTypeGauge
	case model.Counter:
		d := int64(v)
		metrics.Delta = &d
		metrics.MType = model.MetricTypeCounter
	default:
		return errors.New("unexpected metric value type")
	}

	b, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("error while preparing body for request: %w", err)
	}

	// gzip
	buf, err := s.compress(b)
	if err != nil {
		return fmt.Errorf("data compression failed: %w", err)
	}

	// prepare request
	url := s.url + "/update/"
	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return fmt.Errorf("error preparing the request: %w", err)
	}

	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

	if s.hostIP != "" {
		req.Header.Set(server.XRealIPHeader, s.hostIP)
	}

	// do request
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("error while doing the request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error while reading the response bytes: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected response status code: %s, body: %s", resp.Status, string(body))
	}

	return nil
}

// compress uses gzip compression.
func (s *sender) compress(b []byte) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)

	zb := gzip.NewWriter(buf)
	_, err := zb.Write(b)
	if err != nil {
		return buf, err
	}

	err = zb.Close()
	if err != nil {
		return buf, err
	}

	return buf, nil
}

// sendBatched sends metrics batch to server.
//
// > Научите агент работать с использованием нового API (отправлять метрики батчами).
//
// TODO: Maybe somehow remake encryption as middleware or smth...
func (s *sender) sendBatched(batch []model.Metrics) error {
	s.Semaphore.Acquire()
	defer s.Semaphore.Release()

	b, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("error while preparing body for request: %w", err)
	}

	// gzip
	buf, err := s.compress(b)
	if err != nil {
		return fmt.Errorf("data compression failed: %w", err)
	}

	// XXX: What should go first - encryption or compression?
	if s.encryptor != nil {
		// TODO: Should also sign payload before encryption.
		// I guess, HashSHA256 header acts as some sort of signature, for now.
		// (hash was added in previous task increments)

		// encrypt payload
		ciphertext, err := s.encryptor.Encrypt(buf.Bytes())
		if err != nil {
			return fmt.Errorf("data encryption failed: %w", err)
		}

		// write encrypted data to the same bytes buffer
		buf.Reset()
		if _, err = buf.Write(ciphertext); err != nil {
			return fmt.Errorf("error writing encrypted payload to the buffer: %w", err)
		}
	}

	// prepare request
	url := s.url + "/updates/"
	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return fmt.Errorf("error preparing the request: %w", err)
	}

	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

	if s.hostIP != "" {
		req.Header.Set(server.XRealIPHeader, s.hostIP)
	}

	if s.encryptor != nil {
		// show via header that content is encrypted
		req.Header.Set(server.EncryptionHeader, "1")
	}

	if s.key != "" {
		// create body hash
		hasher := hmac.New(sha256.New, []byte(s.key))
		_, err = hasher.Write(b)
		if err != nil {
			return fmt.Errorf("error creating hash for request body: %w", err)
		}
		hash := hex.EncodeToString(hasher.Sum(nil))
		req.Header.Set(server.HashHeader, hash)
	}

	// do request
	resp, err := s.client.Do(req)
	if err != nil {
		return model.NewRetriableError(fmt.Errorf("error while doing the request: %w", err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error while reading the response bytes: %w", err)
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("unexpected response status code: '%s', body: '%s'", resp.Status, string(body))
		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			return model.NewRetriableError(err)
		}
		return err
	}

	return nil
}
