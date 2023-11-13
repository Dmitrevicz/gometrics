package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Dmitrevicz/gometrics/internal/model"
)

type sender struct {
	reportInterval int
	url            string
	batch          bool
	poller         *poller
	client         *http.Client
}

func NewSender(reportInterval int, url string, batch bool, poller *poller) *sender {
	return &sender{
		reportInterval: reportInterval,
		url:            url,
		batch:          batch,
		poller:         poller,
		client:         NewClientDefault(),
	}
}

func (s *sender) Start() {
	log.Println("Sender started")

	sleepDuration := time.Second * time.Duration(s.reportInterval)
	var ts time.Time

	for {
		// lesson of the 2nd increment asked to use time.Sleep
		time.Sleep(sleepDuration)

		ts = time.Now()

		// iteration-12:
		// > Научите агент работать с использованием нового API (отправлять метрики батчами).
		if s.batch {
			s.SendBatched(s.poller.AcquireMetrics())
		} else {
			s.Send(s.poller.AcquireMetrics())
		}

		fmt.Println("send fired:", time.Since(ts))
	}
}

func (s *sender) Send(metrics *Metrics) {
	log.Println("Metrics report started")

	if metrics == nil {
		log.Println("Error sending metrics: got nil as *Metrics, skip")
		return
	}

	ts := time.Now()

	for name, gauge := range metrics.Gauges {
		// TODO: make it async
		err := s.sendMetrics(name, gauge)
		if err != nil {
			log.Println("Got error while sending gauge update request: " + err.Error())
		}
	}

	for name, counter := range metrics.Counters {
		// TODO: make it async
		err := s.sendMetrics(name, counter)
		if err != nil {
			log.Println("Got error while sending counter update request: " + err.Error())
		}
	}

	log.Printf("Metrics have been sent (%d in %v)\n", len(metrics.Counters)+len(metrics.Gauges), time.Since(ts))
}

func (s *sender) prepareMetricsBatch(metrics *Metrics) (batch []model.Metrics) {
	batch = make([]model.Metrics, 0, len(metrics.Gauges)+len(metrics.Counters))

	for name, val := range metrics.Gauges {
		gauge := model.Metrics{
			MType: model.MetricTypeGauge,
			ID:    name,
			Value: (*float64)(&val),
		}
		batch = append(batch, gauge)
	}

	for name, val := range metrics.Counters {
		counter := model.Metrics{
			MType: model.MetricTypeCounter,
			ID:    name,
			Delta: (*int64)(&val),
		}
		batch = append(batch, counter)
	}

	return batch
}

// SendBatched
//
// > Научите агент работать с использованием нового API (отправлять метрики батчами).
func (s *sender) SendBatched(metrics *Metrics) {
	log.Println("Metrics report started (batched)")

	if metrics == nil {
		log.Println("Error sending metrics: got nil as *Metrics, skip")
		return
	}

	ts := time.Now()

	batch := s.prepareMetricsBatch(metrics)

	err := s.sendBatched(batch)
	if err != nil {
		log.Println("Got error while sending batched update request: " + err.Error())
	}

	log.Printf("Metrics have been sent (%d in %v)\n", len(metrics.Counters)+len(metrics.Gauges), time.Since(ts))
}

// DefaultHTTPClientTimeoutSeconds - custom default http client timeout in seconds
const DefaultHTTPClientTimeoutSeconds = 10

// NewClientDefault returns *http.Client with DefaultHTTPClientTimeoutSeconds timeout set
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

// > Научите агент работать с использованием нового API (отправлять метрики батчами).
func (s *sender) sendBatched(batch []model.Metrics) error {
	b, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("error while preparing body for request: %w", err)
	}

	// gzip
	buf, err := s.compress(b)
	if err != nil {
		return fmt.Errorf("data compression failed: %w", err)
	}

	// prepare request
	url := s.url + "/updates/"
	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return fmt.Errorf("error preparing the request: %w", err)
	}

	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

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
