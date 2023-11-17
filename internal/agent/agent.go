package agent

import (
	"log"

	"github.com/Dmitrevicz/gometrics/internal/agent/config"
	"github.com/Dmitrevicz/gometrics/internal/model"
)

// Metrics data to be sent to the server
type Metrics struct {
	Gauges   map[string]model.Gauge
	Counters map[string]model.Counter
}

// Agent is responsible for gathering and sending metrics to server
type Agent struct {
	poller *poller
	sender *sender
}

func New(cfg *config.Config) *Agent {
	log.Printf("intervals (in seconds) - poll: %d, report: %d\n", cfg.PollInterval, cfg.ReportInterval)
	log.Printf("url: \"%s\"\n", cfg.ServerURL)

	poller := NewPoller(cfg.PollInterval)
	return &Agent{
		poller: poller,
		sender: NewSender(cfg.ReportInterval, cfg.ServerURL, cfg.Batch, poller),
	}
}

// Start initiates timers
func (a *Agent) Start() {
	log.Println("Agent is starting its timers...")

	go a.poller.Start()
	go a.sender.Start()
}
