package agent

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Dmitrevicz/gometrics/internal/model"
)

type sender struct {
	reportInterval int
	poller         *poller
	client         *http.Client
}

func NewSender(reportInterval int, poller *poller) *sender {
	return &sender{
		reportInterval: reportInterval,
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
		s.Send(s.poller.AcquireMetrics())
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
		err := s.sendGauge(name, gauge)
		if err != nil {
			log.Println("Got error while sending gauge update request: " + err.Error())
		}
	}

	for name, counter := range metrics.Counters {
		// TODO: make it async
		err := s.sendCounter(name, counter)
		if err != nil {
			log.Println("Got error while sending counter update request: " + err.Error())
		}
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
	url := fmt.Sprintf("http://localhost:8080/update/gauge/%s/%f", name, value)
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
		return fmt.Errorf("unexpecteed response status code: %s, body: %s", resp.Status, string(body))
	}

	return nil
}

func (s *sender) sendCounter(name string, value model.Counter) error {
	url := fmt.Sprintf("http://localhost:8080/update/counter/%s/%d", name, value)
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
		return fmt.Errorf("unexpecteed response status code: %s, body: %s", resp.Status, string(body))
	}

	return nil
}
