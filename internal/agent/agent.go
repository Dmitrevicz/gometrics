package agent

import (
	"log"

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

func New(pollInterval, reportInterval int) *Agent {
	log.Printf("intervals (in seconds) - poll: %d, report: %d\n", pollInterval, reportInterval)
	poller := NewPoller(pollInterval)
	return &Agent{
		poller: poller,
		sender: NewSender(reportInterval, poller),
	}
}

// Start initiates timers
func (a *Agent) Start() {
	log.Println("Agent is starting its timers...")

	go a.poller.Start()
	go a.sender.Start()
}
