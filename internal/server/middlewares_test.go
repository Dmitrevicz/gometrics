package server

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Dmitrevicz/gometrics/internal/server/config"
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
