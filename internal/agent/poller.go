package agent

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/Dmitrevicz/gometrics/internal/model"
)

type poller struct {
	pollInterval int

	stat runtime.MemStats // gauges will be polling here

	PollCount   model.Counter // additional custom counter value
	RandomValue model.Gauge   // additional custom gauge value
	LastPoll    time.Time

	// mu sync.RWMutex // (maybe stat structure should be protected with mutex)
}

func NewPoller(pollInterval int) *poller {
	return &poller{
		pollInterval: pollInterval,
	}
}

func (p *poller) Start() {
	log.Println("Poller started")

	sleepDuration := time.Second * time.Duration(p.pollInterval)
	var ts time.Time

	for {
		ts = time.Now()
		p.Poll()
		fmt.Println("poll fired:", p.LastPoll, time.Since(ts))

		// lesson of the 2nd increment asked to use time.Sleep
		time.Sleep(sleepDuration)
	}
}

// Poll updates metrics data every pollInterval seconds
func (p *poller) Poll() {
	// p.mu.Lock()
	// defer p.mu.Unlock()

	// update main metrics data
	runtime.ReadMemStats(&p.stat)

	// update additional metrics
	p.RandomValue = model.Gauge(rand.Float64())
	// p.PollCount++ // tests fail when trying: go test -race ./...
	atomic.AddInt64((*int64)(&p.PollCount), 1) // tests fail on -race flag without this

	p.LastPoll = time.Now()
}

// AcquireMetrics prepares struct of metrics data ready to be sent to server
func (p *poller) AcquireMetrics() (s *Metrics) {
	// (?)
	// Func receiver by value was used so p.stat structure will be safely copied
	// before usage. (Even though struct itself is pretty big, so might be worth
	// considering pointers&locks usage instead)

	// p.mu.RLock()
	// defer p.mu.RUnlock()

	s = new(Metrics)
	s.Counters = map[string]model.Counter{
		"PollCount": p.PollCount, // additional custom counter value
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
