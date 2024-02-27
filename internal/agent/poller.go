package agent

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/Dmitrevicz/gometrics/internal/model"
)

// poller updates metrics data every pollInterval seconds.
type poller struct {
	pollInterval int

	stat runtime.MemStats // gauges will be polling here

	pollCount   model.Counter // additional custom counter value
	RandomValue model.Gauge   // additional custom gauge value
	LastPoll    time.Time

	quit  chan struct{}
	timer *time.Timer

	// polled data must be protected because accessed from separate goroutines
	mu sync.RWMutex
}

// NewPoller returns a poller that should be used to gather metrics data every
// pollInterval seconds.
func NewPoller(pollInterval int) *poller {
	return &poller{
		pollInterval: pollInterval,
		quit:         make(chan struct{}),
	}
}

// Start starts updating metrics data every pollInterval seconds.
func (p *poller) Start() {
	log.Println("Poller started")

	ts := time.Now()

	// poll at the very start and then repeat after every sleepDuration
	p.Poll()
	fmt.Println("poll fired:", p.LastPoll, time.Since(ts))

	sleepDuration := time.Second * time.Duration(p.pollInterval)
	p.timer = time.NewTimer(sleepDuration)

	for {
		select {
		case ts = <-p.timer.C:
			p.Poll()
			fmt.Println("poll fired:", p.LastPoll, time.Since(ts))

			// I don't calculate delta-time, because current behaviour
			// is good enough right now.
			p.timer.Reset(sleepDuration)
		case <-p.quit:
			// stop the timer
			if !p.timer.Stop() {
				// drain the chanel (might not be needed here, but leave it be
				// as a kind of exercise)
				<-p.timer.C
			}

			log.Println("Poller timer stopped")
			return
		}
	}
}

// Stop stops poller's timer.
func (p *poller) Stop() {
	close(p.quit)
	log.Println("Poller stopped")
}

// Poll retrieves metrics data from runtime.
func (p *poller) Poll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// update main metrics data
	runtime.ReadMemStats(&p.stat)

	// update additional metrics
	p.RandomValue = model.Gauge(rand.Float64())
	p.pollCount++ // tests failed here when run with -race flag and there were no mutex or any other syncronization

	p.LastPoll = time.Now()
}

// PollCount returns current counter value representing number of polls.
func (p *poller) PollCount() (pc model.Counter) {
	p.mu.RLock()
	pc = p.pollCount
	p.mu.RUnlock()

	return
}

// AcquireMetrics prepares struct of metrics data ready to be sent to server.
func (p *poller) AcquireMetrics() (s Metrics) {
	// (?)
	// Func receiver by value was used so p.stat structure will be safely copied
	// before usage. (Even though struct itself is pretty big, so might be worth
	// considering pointers&locks usage instead)

	p.mu.RLock()
	defer p.mu.RUnlock()

	// s = new(Metrics)
	s.Counters = map[string]model.Counter{
		"PollCount": p.pollCount, // additional custom counter value
	}

	s.Gauges = map[string]model.Gauge{
		"Alloc":         model.Gauge(p.stat.Alloc),
		"BuckHashSys":   model.Gauge(p.stat.BuckHashSys),
		"Frees":         model.Gauge(p.stat.Frees),
		"GCCPUFraction": model.Gauge(p.stat.GCCPUFraction),
		"GCSys":         model.Gauge(p.stat.GCSys),
		"HeapAlloc":     model.Gauge(p.stat.HeapAlloc),
		"HeapIdle":      model.Gauge(p.stat.HeapIdle),
		"HeapInuse":     model.Gauge(p.stat.HeapInuse),
		"HeapObjects":   model.Gauge(p.stat.HeapObjects),
		"HeapReleased":  model.Gauge(p.stat.HeapReleased),
		"HeapSys":       model.Gauge(p.stat.HeapSys),
		"LastGC":        model.Gauge(p.stat.LastGC),
		"Lookups":       model.Gauge(p.stat.Lookups),
		"MCacheInuse":   model.Gauge(p.stat.MCacheInuse),
		"MCacheSys":     model.Gauge(p.stat.MCacheSys),
		"MSpanInuse":    model.Gauge(p.stat.MSpanInuse),
		"MSpanSys":      model.Gauge(p.stat.MSpanSys),
		"Mallocs":       model.Gauge(p.stat.Mallocs),
		"NextGC":        model.Gauge(p.stat.NextGC),
		"NumForcedGC":   model.Gauge(p.stat.NumForcedGC),
		"NumGC":         model.Gauge(p.stat.NumGC),
		"OtherSys":      model.Gauge(p.stat.OtherSys),
		"PauseTotalNs":  model.Gauge(p.stat.PauseTotalNs),
		"StackInuse":    model.Gauge(p.stat.StackInuse),
		"StackSys":      model.Gauge(p.stat.StackSys),
		"Sys":           model.Gauge(p.stat.Sys),
		"TotalAlloc":    model.Gauge(p.stat.TotalAlloc),
		"RandomValue":   model.Gauge(p.RandomValue), // additional custom gauge value
	}

	return s
}
