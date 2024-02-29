package agent

import (
	"net/http/httptest"
	"os"
	"testing"

	configAgent "github.com/Dmitrevicz/gometrics/internal/agent/config"
	"github.com/Dmitrevicz/gometrics/internal/model"
	"github.com/Dmitrevicz/gometrics/internal/server"
	configServer "github.com/Dmitrevicz/gometrics/internal/server/config"
	"github.com/Dmitrevicz/gometrics/pkg/encryptor"
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

	sender, err := NewSender(cfgAgent, poller, gopsutilPoller)
	require.NoError(t, err)

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

func TestSender_sendBatchedEncrypted(t *testing.T) {
	// generate files containing encryption keys
	pub, priv := prepareTestSenderRSAKeyFiles(t)

	hashKey := "8bb5929d212764f7923ae9998fa18aa46ca4ee8b1cfd319b"

	cfgServer := configServer.NewTesting()
	cfgServer.CryptoKey = priv
	cfgServer.Key = hashKey

	ts := httptest.NewServer(server.New(cfgServer))
	defer ts.Close()

	cfgAgent := &configAgent.Config{
		ServerURL:      ts.URL,
		CryptoKey:      pub,
		Key:            hashKey,
		PollInterval:   0,
		ReportInterval: 0,
		Batch:          false,
	}

	poller := NewPoller(cfgAgent.PollInterval)
	gopsutilPoller := NewGopsutilPoller(cfgAgent.PollInterval)

	sender, err := NewSender(cfgAgent, poller, gopsutilPoller)
	require.NoError(t, err, "failed to create sender instance")

	var (
		// gaugeValue   float64 = 42.420 // I hate statictest
		gaugeValue         = 42.420
		counterValue int64 = 42
	)

	metricsBatch := []model.Metrics{
		{
			ID:    "TestSenderEncrMetric1",
			MType: model.MetricTypeGauge,
			Value: &gaugeValue,
		},
		{
			ID:    "TestSenderEncrMetric2",
			MType: model.MetricTypeCounter,
			Delta: &counterValue,
		},
	}

	err = sender.sendBatched(metricsBatch)
	require.NoError(t, err, "failed to send metrics")
}

func prepareTestSenderRSAKeyFiles(t *testing.T) (pub, priv string) {
	t.Helper()

	// generate rsa private key
	privateKey, err := encryptor.GenerateKeys(2048)
	require.NoError(t, err, "failed to generate private key: %v", err)

	// encode keys to PEM format
	privatePEM, err := encryptor.FormatPrivateKey(privateKey)
	require.NoError(t, err, "failed to encode private key to PEM format: %v", err)
	publicPEM, err := encryptor.FormatPublicKey(&privateKey.PublicKey)
	require.NoError(t, err, "failed to encode private key to PEM format: %v", err)

	// write keys in temporary directory
	dir := t.TempDir()

	// write private key to file
	fPriv, err := os.CreateTemp(dir, "test_sender_genkey_*")
	require.NoError(t, err, "failed to create temporary private key file: %v", err)
	defer fPriv.Close()

	_, err = fPriv.Write(privatePEM)
	require.NoError(t, err, "failed to write private key to temporary file: %v", err)

	// write public key to file
	fPub, err := os.CreateTemp(dir, "test_sender_genkey_pub*")
	require.NoError(t, err, "failed to create temporary public key file: %v", err)
	defer fPub.Close()

	_, err = fPub.Write(publicPEM)
	require.NoError(t, err, "failed to write public key to temporary file: %v", err)

	return fPub.Name(), fPriv.Name()
}
