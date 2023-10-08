package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandlers_UpdateGauge(t *testing.T) {
	pathMain := "/update"

	type RequestData struct {
		method     string
		metricType string
		metricName string
		value      interface{}
	}

	data := struct {
		good, bad RequestData
	}{
		good: RequestData{
			method:     http.MethodPost,
			metricType: "gauge",
			metricName: "Alloc",
			value:      42.420,
		},
		bad: RequestData{
			method:     http.MethodPatch,
			metricType: "broken-type",
			metricName: "%20",
			value:      "broken-value",
		},
	}

	tests := []struct {
		name       string
		wantCode   int
		method     string
		metricType string
		metricName string
		value      interface{}
		specialURL string
	}{
		{
			name:       "correct",
			wantCode:   http.StatusOK,
			method:     data.good.method,
			metricType: data.good.metricType,
			metricName: data.good.metricName,
			value:      data.good.value,
		},
		{
			name:       "incorrect-metric-value",
			wantCode:   http.StatusBadRequest,
			method:     data.good.method,
			metricType: data.good.metricType,
			metricName: data.good.metricName,
			value:      data.bad.value,
		},
		{
			name:       "incorrect-metric-value-empty",
			wantCode:   http.StatusBadRequest,
			method:     data.good.method,
			metricType: data.good.metricType,
			metricName: data.good.metricName,
			value:      "",
		},
		{
			name:       "incorrect-metric-type",
			wantCode:   http.StatusBadRequest,
			method:     data.good.method,
			metricType: data.bad.metricType,
			metricName: data.good.metricName,
			value:      data.good.value,
		},
		{
			name:       "incorrect-metric-type-2",
			wantCode:   http.StatusBadRequest,
			method:     data.good.method,
			specialURL: "/",
		},
		{
			name:       "incorrect-metric-name",
			wantCode:   http.StatusNotFound,
			method:     data.good.method,
			metricType: data.good.metricType,
			metricName: data.bad.metricName,
			value:      data.good.value,
		},
		{
			name:       "incorrect-method",
			wantCode:   http.StatusNotFound,
			method:     data.bad.method,
			metricType: data.good.metricType,
			metricName: data.good.metricName,
			value:      data.good.value,
		},
	}

	server := New()

	for _, tc := range tests {
		t.Run(tc.method, func(t *testing.T) {
			reqURL := pathMain + tc.specialURL
			if reqURL == pathMain {
				reqURL += fmt.Sprintf("/%s/%s/%v", tc.metricType, tc.metricName, tc.value)
			}

			r := httptest.NewRequest(tc.method, reqURL, nil)
			w := httptest.NewRecorder()

			server.ServeHTTP(w, r)

			assert.Equalf(t, tc.wantCode, w.Code, "Код ответа не совпадает с ожидаемым. Method: %s, URL: %s", r.Method, reqURL)
		})
	}
}

func TestHandlers_UpdateCounter(t *testing.T) {
	pathMain := "/update"

	type RequestData struct {
		method     string
		metricType string
		metricName string
		value      interface{}
	}

	data := struct {
		good, bad RequestData
	}{
		good: RequestData{
			method:     http.MethodPost,
			metricType: "counter",
			metricName: "PollCount",
			value:      42,
		},
		bad: RequestData{
			method:     http.MethodPatch,
			metricType: "broken-type",
			metricName: "%20",
			value:      "broken-value",
		},
	}

	tests := []struct {
		name       string
		wantCode   int
		method     string
		metricType string
		metricName string
		value      interface{}
		specialURL string
	}{
		{
			name:       "correct",
			wantCode:   http.StatusOK,
			method:     data.good.method,
			metricType: data.good.metricType,
			metricName: data.good.metricName,
			value:      data.good.value,
		},
		{
			name:       "incorrect-metric-value",
			wantCode:   http.StatusBadRequest,
			method:     data.good.method,
			metricType: data.good.metricType,
			metricName: data.good.metricName,
			value:      data.bad.value,
		},
		{
			name:       "incorrect-metric-value-empty",
			wantCode:   http.StatusBadRequest,
			method:     data.good.method,
			metricType: data.good.metricType,
			metricName: data.good.metricName,
			value:      "",
		},
		{
			name:       "incorrect-metric-value-neg",
			wantCode:   http.StatusBadRequest,
			method:     data.good.method,
			metricType: data.good.metricType,
			metricName: data.good.metricName,
			value:      -42,
		},
		{
			name:       "incorrect-metric-type",
			wantCode:   http.StatusBadRequest,
			method:     data.good.method,
			metricType: data.bad.metricType,
			metricName: data.good.metricName,
			value:      data.good.value,
		},
		{
			name:       "incorrect-metric-type-2",
			wantCode:   http.StatusBadRequest,
			method:     data.good.method,
			specialURL: "/",
		},
		{
			name:       "incorrect-metric-name",
			wantCode:   http.StatusNotFound,
			method:     data.good.method,
			metricType: data.good.metricType,
			metricName: data.bad.metricName,
			value:      data.good.value,
		},
		{
			name:       "incorrect-method",
			wantCode:   http.StatusNotFound,
			method:     data.bad.method,
			metricType: data.good.metricType,
			metricName: data.good.metricName,
			value:      data.good.value,
		},
	}

	server := New()

	for _, tc := range tests {
		t.Run(tc.method, func(t *testing.T) {
			reqURL := pathMain + tc.specialURL
			if reqURL == pathMain {
				reqURL += fmt.Sprintf("/%s/%s/%v", tc.metricType, tc.metricName, tc.value)
			}

			r := httptest.NewRequest(tc.method, reqURL, nil)
			w := httptest.NewRecorder()

			server.ServeHTTP(w, r)

			assert.Equalf(t, tc.wantCode, w.Code, "Код ответа не совпадает с ожидаемым. Method: %s, URL: %s", r.Method, reqURL)
		})
	}
}
