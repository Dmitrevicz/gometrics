package server

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Dmitrevicz/gometrics/internal/model"
	"github.com/Dmitrevicz/gometrics/internal/server/config"
	"github.com/Dmitrevicz/gometrics/pkg/encryptor"
	"github.com/stretchr/testify/require"
)

func TestGzipCompression(t *testing.T) {
	server := New(config.NewTesting())

	requestBody := `{
		"id": "testGauge",
		"type": "gauge",
		"value": 4.5
	}`

	successBody := requestBody

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest("POST", "/update/", buf)
		r.Header.Set("Content-Encoding", "gzip")

		w := httptest.NewRecorder()
		server.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()

		require.Equal(t, http.StatusOK, res.StatusCode)

		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		require.JSONEq(t, successBody, string(b))
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)
		r := httptest.NewRequest("POST", "/update/", buf)
		r.Header.Set("Accept-Encoding", "gzip")

		w := httptest.NewRecorder()
		server.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()

		require.Equal(t, http.StatusOK, res.StatusCode)

		zr, err := gzip.NewReader(res.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.JSONEq(t, successBody, string(b))
	})

	t.Run("send_and_accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest("POST", "/update/", buf)
		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Accept-Encoding", "gzip")

		w := httptest.NewRecorder()
		server.ServeHTTP(w, r)

		res := w.Result()
		defer res.Body.Close()

		require.Equal(t, http.StatusOK, res.StatusCode)

		zr, err := gzip.NewReader(res.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.JSONEq(t, successBody, string(b))
	})
}

// TestDecryptRSA tests if server handles encrypted payload.
func TestDecryptRSA(t *testing.T) {
	// generate files containing encryption keys
	pub, priv := prepareTestDecryptRSAKeyFiles(t)

	// create encryptor; decryptor will be created by the server
	encrypter, err := encryptor.NewEncryptor(pub)
	require.NoError(t, err, "failed to create Encryptor")

	reqURL := "/updates/"
	gaugeValue := 42.420
	counterValue := 42

	// content to be encrypted and sent
	reqBody := fmt.Sprintf(`[
		{"id":"TestEncMetric1","type":"%s","value":%f},
		{"id":"TestEncMetric2","type":"%s","delta":%d}]`,
		model.MetricTypeGauge, gaugeValue,
		model.MetricTypeCounter, counterValue,
	)

	encryptedBody, err := encrypter.Encrypt([]byte(reqBody))
	require.NoError(t, err, "failed to encrypt request payload")

	r := httptest.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(encryptedBody))
	r.Header.Set(EncryptionHeader, "1")

	w := httptest.NewRecorder()

	cfg := config.NewTesting()
	cfg.CryptoKey = priv // path to private key

	server := New(cfg)
	server.ServeHTTP(w, r)

	res := w.Result()
	defer res.Body.Close()

	responseBody, err := io.ReadAll(res.Body)
	require.NoError(t, err, "failed to read response body")

	require.Equal(t, http.StatusOK, res.StatusCode,
		"encrypted request failed, status: '%s', response: '%s'",
		res.Status, string(responseBody),
	)
}

func prepareTestDecryptRSAKeyFiles(t *testing.T) (pub, priv string) {
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
	fPriv, err := os.CreateTemp(dir, "test_genkey_*")
	require.NoError(t, err, "failed to create temporary private key file: %v", err)
	defer fPriv.Close()

	_, err = fPriv.Write(privatePEM)
	require.NoError(t, err, "failed to write private key to temporary file: %v", err)

	// write public key to file
	fPub, err := os.CreateTemp(dir, "test_genkey_pub*")
	require.NoError(t, err, "failed to create temporary public key file: %v", err)
	defer fPub.Close()

	_, err = fPub.Write(publicPEM)
	require.NoError(t, err, "failed to write public key to temporary file: %v", err)

	return fPub.Name(), fPriv.Name()
}
