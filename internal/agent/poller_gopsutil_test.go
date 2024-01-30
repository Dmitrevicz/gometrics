package agent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type AgentGopsutilPollerSuit struct {
	suite.Suite
}

func TestGopsutilPoller(t *testing.T) {
	suite.Run(t, new(AgentGopsutilPollerSuit))
}

func (s *AgentGopsutilPollerSuit) TestPoll() {
	p := NewGopsutilPoller(0)

	prevPollTime := p.lastPoll

	// check initial values of poller fields
	s.Assert().Empty(p.stat, "initial poller stat structure is expected to have zero-values")
	s.Assert().LessOrEqual(p.lastPoll, time.Now(), "initial poller LastPoll timestamp is wrong")

	// no need to continue if poller initialization have already failed
	if s.T().Failed() {
		return
	}

	// trigger metrics data update
	err := p.Poll()
	s.Require().NoError(err, "poller Poll failed")

	// check that metrics data was updated after the Poll()
	s.Assert().NotEmpty(p.stat, "poller stat structure wasn't updated after Poll")
	s.Assert().NotEmpty(p.stat.CPUsUtilz, "poller cpu stat wasn't updated after Poll")
	s.Assert().WithinRange(p.lastPoll, prevPollTime, time.Now(), "wrong timestamp after Poll call")
}

// TestStart checks Polls being called within timer
func (s *AgentGopsutilPollerSuit) TestStartTicker() {
	const (
		expectedGaugesCount = 3 // can be more than this
		pollInterval        = 1 // poll interval in seconds
		checkTickInterval   = (time.Second * pollInterval) / 2
		durationTillTimeout = time.Second*pollInterval*3 + checkTickInterval
	)

	p := NewGopsutilPoller(pollInterval)
	go p.Start()

	s.Assert().Eventuallyf(
		func() bool {
			m := p.AcquireMetrics()
			return len(m.Gauges) >= expectedGaugesCount
		},
		durationTillTimeout,
		checkTickInterval,
		"poller stats wasn't updated in expected period of time",
	)
}

func (s *AgentGopsutilPollerSuit) TestAcquireMetrics() {
	p := NewGopsutilPoller(0)

	err := p.Poll()
	s.Require().NoError(err)

	m := p.AcquireMetrics()

	metrics := struct {
		got  []string
		want []string // elements from the lesson task
	}{
		got: make([]string, 0, len(m.Counters)+len(m.Gauges)),
		want: []string{
			"TotalMemory",
			"FreeMemory",
			"CPUutilization0",
		},
	}

	// get list of metrics names aqcuired
	for name := range m.Gauges {
		metrics.got = append(metrics.got, name)
	}
	for name := range m.Counters {
		metrics.got = append(metrics.got, name)
	}

	s.Assert().GreaterOrEqual(len(metrics.got), len(metrics.want), "not enough metrics collected")

	for _, want := range metrics.want {
		s.Assert().Contains(metrics.got, want)
	}
}
