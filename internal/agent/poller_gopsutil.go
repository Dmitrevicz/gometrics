package agent

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/Dmitrevicz/gometrics/internal/model"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// data gathered from gopsutil lib
type gopsutilStat struct {
	TotalMemory uint64
	FreeMemory  uint64
	CPUsUtilz   []float64 // percents of CPUs utilizations
}

// gopsutilPoller
//
// > [Iteration 15] Добавьте ещё одну горутину, которая будет использовать пакет
// gopsutil и собирать дополнительные метрики типа gauge:
// - TotalMemory,
// - FreeMemory,
// - CPUutilization1 (точное количество — по числу CPU, определяемому во время исполнения).
type gopsutilPoller struct {
	pollInterval int

	stat gopsutilStat // gauges will be polling here

	lastPoll time.Time // for debugging

	quit chan struct{}

	// polled data must be protected because accessed from separate goroutines
	mu sync.RWMutex
}

func NewGopsutilPoller(pollInterval int) *gopsutilPoller {
	return &gopsutilPoller{
		pollInterval: pollInterval,
		quit:         make(chan struct{}),
	}
}

func (p *gopsutilPoller) Start() {
	log.Println("Poller started (gopsutil)")

	sleepDuration := time.Second * time.Duration(p.pollInterval)
	var (
		ts  time.Time
		err error
	)

	for {
		select {
		case <-p.quit:
			log.Println("Poller timer stopped (gopsutil)")
			return
		default:
		}

		ts = time.Now()
		if err = p.Poll(); err != nil {
			log.Println("Poll failed for gopsutil, err:", err)
		}
		fmt.Println("poll fired (gopsutil):", p.lastPoll, time.Since(ts))

		time.Sleep(sleepDuration)
	}
}

// Stop stops poller's timer.
func (p *gopsutilPoller) Stop() {
	close(p.quit)
	log.Println("Poller stopped (gopsutil)")
}

// Poll updates metrics data every pollInterval seconds
func (p *gopsutilPoller) Poll() error {
	stats, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	cpus, err := cpu.Percent(0, true)
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.stat.TotalMemory = stats.Total
	p.stat.FreeMemory = stats.Free
	p.stat.CPUsUtilz = cpus
	p.lastPoll = time.Now()

	return nil
}

// AcquireMetrics prepares struct of metrics data ready to be sent to server
func (p *gopsutilPoller) AcquireMetrics() (s *Metrics) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	s = new(Metrics)
	s.Gauges = make(map[string]model.Gauge, len(p.stat.CPUsUtilz)+2)

	for i, cpu := range p.stat.CPUsUtilz {
		s.Gauges["CPUutilization"+strconv.Itoa(i)] = model.Gauge(cpu)
	}

	s.Gauges["TotalMemory"] = model.Gauge(p.stat.TotalMemory)
	s.Gauges["FreeMemory"] = model.Gauge(p.stat.FreeMemory)

	return s
}
