package agent

import (
	"testing"
)

func BenchmarkPoller(b *testing.B) {
	b.Run("Poll", benchPollerPoll)
	b.Run("Acquire", benchPollerAcquire)
}

func benchPollerPoll(b *testing.B) {
	p := NewPoller(0)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p.Poll()
	}
}

func benchPollerAcquire(b *testing.B) {
	p := NewPoller(0)
	p.Poll()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p.AcquireMetrics()
	}
}
