package agent

import (
	"net/http/httptest"
	"testing"

	configAgent "github.com/Dmitrevicz/gometrics/internal/agent/config"
	"github.com/Dmitrevicz/gometrics/internal/model"
	"github.com/Dmitrevicz/gometrics/internal/server"
	configServer "github.com/Dmitrevicz/gometrics/internal/server/config"
	"github.com/stretchr/testify/require"
)

func TestSender(t *testing.T) {
	cfgServer := configServer.NewTesting()

	ts := httptest.NewServer(server.New(cfgServer))
	defer ts.Close()

	cfgAgent := &configAgent.Config{
		ServerURL:      ts.URL,
		Key:            "",
		PollInterval:   0,
		ReportInterval: 0,
		Batch:          false,
	}
	poller := NewPoller(cfgAgent.PollInterval)
	gopsutilPoller := NewGopsutilPoller(cfgAgent.PollInterval)
	sender := NewSender(cfgAgent, poller, gopsutilPoller)

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
