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
	"github.com/Dmitrevicz/gometrics/pkg/hostip"
)

// Metrics data to be sent to the server.
type Metrics struct {
	Gauges   map[string]model.Gauge
	Counters map[string]model.Counter
}

// Merge adds values of maps from m2 to m.
// Panic will be thrown if m is nil.
func (m *Metrics) Merge(m2 *Metrics) {
	for name, value := range m2.Gauges {
		m.Gauges[name] = value
	}

	for name, value := range m2.Counters {
		m.Counters[name] = value
	}
}

func (m *Metrics) Len() int {
	return len(m.Counters) + len(m.Gauges)
}

// Agent is responsible for gathering and sending metrics to server.
type Agent struct {
	poller         *poller
	gopsutilPoller *gopsutilPoller

	// sender *sender
	sender MetricsSender
}

// New creates new agent service.
func New(cfg *config.Config) (*Agent, error) {
	log.Printf("intervals (in seconds) - poll: %d, report: %d\n", cfg.PollInterval, cfg.ReportInterval)
	log.Printf("url: \"%s\"\n", cfg.ServerURL)

	poller := NewPoller(cfg.PollInterval)
	gopsutilPoller := NewGopsutilPoller(cfg.PollInterval)

	var err error
	if err = detectHostIP(cfg); err != nil {
		return nil, err
	}

	var sender MetricsSender
	if cfg.GRPCServerURL != "" {
		sender, err = NewSenderGRPC(cfg, poller, gopsutilPoller)
	} else {
		sender, err = NewSender(cfg, poller, gopsutilPoller)
	}
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

// detectHostIP tries to detect host IP dynamically when Config.HostIP is empty.
func detectHostIP(cfg *config.Config) (err error) {
	if cfg.HostIP == "" {
		// try to detect host IP
		cfg.HostIP, err = hostip.FindHostIP()
		if err != nil {
			// return fmt.Errorf("attempt to detect host IP failed with error: %v", err)
			log.Printf("attempt to detect host IP failed with error: %v", err)
			return nil
		}

		log.Printf("Host IP detected dynamically: %s\n", cfg.HostIP)
	} else {
		log.Printf("Host IP retrieved from config file: %s\n", cfg.HostIP)
	}

	return nil
}
