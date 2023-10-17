package agent

import (
	"net/http/httptest"
	"testing"

	"github.com/Dmitrevicz/gometrics/internal/model"
	"github.com/Dmitrevicz/gometrics/internal/server"
	"github.com/stretchr/testify/require"
)

func TestSender(t *testing.T) {
	ts := httptest.NewServer(server.New())
	defer ts.Close()

	poller := NewPoller(0)
	sender := NewSender(0, ts.URL, poller)

	t.Run("sendGauge", func(t *testing.T) {
		err := sender.sendGauge("testGauge", model.Gauge(42.420))
		require.NoError(t, err)
	})

	t.Run("sendCounter", func(t *testing.T) {
		err := sender.sendCounter("testCounter", model.Counter(42))
		require.NoError(t, err)
	})
}
