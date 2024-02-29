// Package agent represents agent service that gathers runtime metrics.
//
// Package contains poller and sender.
// Poller is used to periodically gather runtime metrics data.
// Sender sends data, gathered by the poller, to the server.
package agent

import (
	"context"
	"log"

	"github.com/Dmitrevicz/gometrics/internal/agent/config"
	"github.com/Dmitrevicz/gometrics/internal/model"
)

// Metrics data to be sent to the server.
type Metrics struct {
	Gauges   map[string]model.Gauge
	Counters map[string]model.Counter
}

// Merge adds values of maps from m2 to m1.
// Panic will be thrown if m1 is nil.
func (m1 *Metrics) Merge(m2 *Metrics) {
	for name, value := range m2.Gauges {
		m1.Gauges[name] = value
	}

	for name, value := range m2.Counters {
		m1.Counters[name] = value
	}
}

// Agent is responsible for gathering and sending metrics to server.
type Agent struct {
	poller         *poller
	gopsutilPoller *gopsutilPoller

	sender *sender
}

// New creates new agent service.
func New(cfg *config.Config) (*Agent, error) {
	log.Printf("intervals (in seconds) - poll: %d, report: %d\n", cfg.PollInterval, cfg.ReportInterval)
	log.Printf("url: \"%s\"\n", cfg.ServerURL)

	poller := NewPoller(cfg.PollInterval)
	gopsutilPoller := NewGopsutilPoller(cfg.PollInterval)

	sender, err := NewSender(cfg, poller, gopsutilPoller)
	if err != nil {
		return nil, err
	}

	return &Agent{
		poller:         poller,
		gopsutilPoller: gopsutilPoller,
		sender:         sender,
	}, nil
}

// Start initiates agent timers.
func (a *Agent) Start() {
	log.Println("Agent is starting its timers...")

	go a.poller.Start()
	go a.gopsutilPoller.Start() // > "Добавьте ещё одну горутину, которая будет использовать пакет gopsutil"
	go a.sender.Start()
}

// Shutdown implements graceful shutdown.
// Shutdown stops poller and sender timers and sends current data to server.
func (a *Agent) Shutdown(ctx context.Context) (err error) {
	a.poller.Stop()
	a.gopsutilPoller.Stop()

	return a.sender.Shutdown(ctx)
}
