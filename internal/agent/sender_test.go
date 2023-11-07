package agent

import (
	"net/http/httptest"
	"testing"

	"github.com/Dmitrevicz/gometrics/internal/model"
	"github.com/Dmitrevicz/gometrics/internal/server"
	"github.com/Dmitrevicz/gometrics/internal/server/config"
	"github.com/stretchr/testify/require"
)

func TestSender(t *testing.T) {
	cfg := config.NewTesting()

	ts := httptest.NewServer(server.New(cfg))
	defer ts.Close()

	poller := NewPoller(0)
	sender := NewSender(0, ts.URL, poller)

	gaugeValue := model.Gauge(42.420)
	counterValue := model.Counter(42)

	t.Run("sendGauge", func(t *testing.T) {
		err := sender.sendGauge("testGauge", gaugeValue)
		require.NoError(t, err)
	})

	t.Run("sendCounter", func(t *testing.T) {
		err := sender.sendCounter("testCounter", counterValue)
		require.NoError(t, err)
	})

	t.Run("sendMetrics", func(t *testing.T) {
		err := sender.sendMetrics("testCounter2", counterValue)
		require.NoError(t, err)

		err = sender.sendMetrics("testGauge2", gaugeValue)
		require.NoError(t, err)

		err = sender.sendMetrics("unknownMetric", false)
		require.Error(t, err)
	})
}
