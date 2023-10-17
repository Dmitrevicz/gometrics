package agent

import (
	"testing"
	"time"

	"github.com/Dmitrevicz/gometrics/internal/model"
	"github.com/stretchr/testify/suite"
)

type AgentPollerSuit struct {
	suite.Suite
}

func TestPoller(t *testing.T) {
	suite.Run(t, new(AgentPollerSuit))
}

func (s *AgentPollerSuit) TestPoll() {
	p := NewPoller(0)

	prevPollTime := p.LastPoll

	// check initial values of poller fields
	s.Assert().Empty(p.stat, "initial poller stat structure is expected to have zero-values")
	s.Assert().LessOrEqual(p.LastPoll, time.Now(), "initial poller LastPoll timestamp is wrong")
	s.Assert().Zero(p.PollCount, "initial poller PollCount value must be 0")

	// no need to continue if poller initialization have already failed
	if s.T().Failed() {
		return
	}

	// trigger metrics data update
	p.Poll()

	// check that metrics data was updated after the Poll()
	s.Assert().NotEmpty(p.stat, "poller stat structure wasn't updated after Poll")
	s.Assert().WithinRange(p.LastPoll, prevPollTime, time.Now(), "wrong timestamp after Poll call")
	s.Assert().Equal(model.Counter(1), p.PollCount, "wrong poller PollCount value after poll call")
}

// TestStart checks Polls being called within timer
func (s *AgentPollerSuit) TestStartTicker() {
	const (
		expectedPollCount = 2
		pollInterval      = 2
		checkTickInterval = (time.Second * pollInterval) / 2
		// 2*checkTickInterval is added to be sure that PollCount will 100% have enough time to reach expected value
		durationTillTimout = time.Second*pollInterval*expectedPollCount + 2*checkTickInterval
	)

	p := NewPoller(pollInterval)
	go p.Start()

	s.Assert().Eventuallyf(
		func() bool {
			return p.PollCount == expectedPollCount
		},
		durationTillTimout,
		checkTickInterval,
		"PollCount didn't reach %d in expected period of time",
		expectedPollCount,
	)
}

func (s *AgentPollerSuit) TestAcquireMetrics() {
	p := NewPoller(0)
	p.Poll()
	m := p.AcquireMetrics()

	metrics := struct {
		got  []string
		want []string // elements from the lesson task
	}{
		got: make([]string, 0, len(m.Counters)+len(m.Gauges)),
		want: []string{
			"Alloc",
			"BuckHashSys",
			"Frees",
			"GCCPUFraction",
			"GCSys",
			"HeapAlloc",
			"HeapIdle",
			"HeapInuse",
			"HeapObjects",
			"HeapReleased",
			"HeapSys",
			"LastGC",
			"Lookups",
			"MCacheInuse",
			"MCacheSys",
			"MSpanInuse",
			"MSpanSys",
			"Mallocs",
			"NextGC",
			"NumForcedGC",
			"NumGC",
			"OtherSys",
			"PauseTotalNs",
			"StackInuse",
			"StackSys",
			"Sys",
			"TotalAlloc",
			"RandomValue",
			"PollCount",
		},
	}

	// get list of metrics names aqcuired
	for name := range m.Gauges {
		metrics.got = append(metrics.got, name)
	}
	for name := range m.Counters {
		metrics.got = append(metrics.got, name)
	}

	s.Assert().Len(metrics.got, len(metrics.want))
	s.Assert().ElementsMatch(metrics.want, metrics.got)
}
