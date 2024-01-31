package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Dmitrevicz/gometrics/internal/model"
	"github.com/Dmitrevicz/gometrics/internal/server/config"
	"github.com/Dmitrevicz/gometrics/internal/storage"
)

type handlerTestCase struct {
	Name        string
	Method      string
	Path        string
	StatusCode  func(code int) bool
	RequestBody []byte
}

func prepareTestMetrics(s storage.Storage) ([]model.MetricCounter, []model.MetricGauge, error) {
	size := 1
	counters := make([]model.MetricCounter, 0, size)
	gauges := make([]model.MetricGauge, 0, size)
	for i := 1; i <= size; i++ {
		counters = append(counters, model.MetricCounter{
			Name:  "TestCounter" + strconv.Itoa(i),
			Value: model.Counter(i),
		})
		gauges = append(gauges, model.MetricGauge{
			Name:  "TestGauge" + strconv.Itoa(i),
			Value: model.Gauge(i),
		})
	}

	if err := s.Counters().BatchUpdate(counters); err != nil {
		return nil, nil, err
	}

	if err := s.Gauges().BatchUpdate(gauges); err != nil {
		return nil, nil, err
	}

	return counters, gauges, nil
}

func prepareCases(cs []model.MetricCounter, gs []model.MetricGauge) ([]handlerTestCase, error) {
	defaultCodeCheck := func(code int) bool {
		return code >= 400
	}

	// metrics := make([]model.Metrics, 0, len(cs)+len(gs))
	// for _, counter := range cs {
	// 	val := counter.Value
	// 	metrics = append(metrics, model.Metrics{
	// 		ID:    counter.Name,
	// 		MType: model.MetricTypeCounter,
	// 		Delta: (*int64)(&val),
	// 	})
	// }
	// for _, gauge := range gs {
	// 	val := gauge.Value
	// 	metrics = append(metrics, model.Metrics{
	// 		ID:    gauge.Name,
	// 		MType: model.MetricTypeGauge,
	// 		Value: (*float64)(&val),
	// 	})
	// }

	// updateBatchedBody, err := json.Marshal(&metrics)
	// if err != nil {
	// 	return nil, err
	// }

	return []handlerTestCase{
		{
			Name:       "Ping",
			Path:       "/ping",
			Method:     http.MethodGet,
			StatusCode: defaultCodeCheck,
		},
		{
			Name:        "GetValue",
			Path:        "/value/",
			Method:      http.MethodPost,
			StatusCode:  defaultCodeCheck,
			RequestBody: []byte(fmt.Sprintf("{\"id\":\"%s\",\"type\":\"%s\"}", gs[0].Name, model.MetricTypeGauge)),
		},
		{
			Name:       "UpdateBatched/Counter",
			Path:       "/updates/",
			Method:     http.MethodPost,
			StatusCode: defaultCodeCheck,
			// RequestBody: updateBatchedBody,
			RequestBody: []byte(fmt.Sprintf("[{\"id\":\"%s\",\"type\":\"%s\",\"delta\":%d}]", cs[0].Name, model.MetricTypeCounter, cs[0].Value)),
		},
		{
			Name:        "UpdateBatched/Gauge",
			Path:        "/updates/",
			Method:      http.MethodPost,
			StatusCode:  defaultCodeCheck,
			RequestBody: []byte(fmt.Sprintf("[{\"id\":\"%s\",\"type\":\"%s\",\"value\":%f}]", gs[0].Name, model.MetricTypeGauge, float64(gs[0].Value))),
		},
	}, nil
}

func BenchmarkHandlers(b *testing.B) {
	b.ReportAllocs()

	server := New(config.NewTesting())

	counters, gauges, err := prepareTestMetrics(server.Storage)
	if err != nil {
		b.Fatalf("Error preparing test metrics for the benchmark: %v", err)
	}
	tests, err := prepareCases(counters, gauges)
	if err != nil {
		b.Fatalf("Error preparing test cases for the benchmark: %v", err)
	}

	b.ResetTimer()

	for _, tc := range tests {
		b.Run(tc.Name, func(b *testing.B) {
			var (
				r       *http.Request
				w       *httptest.ResponseRecorder
				res     *http.Response
				reqBody *bytes.Reader
			)

			b.ResetTimer()
			b.StopTimer()

			for i := 0; i < b.N; i++ {
				reqBody = bytes.NewReader(tc.RequestBody)
				r = httptest.NewRequest(tc.Method, tc.Path, reqBody)
				w = httptest.NewRecorder()
				b.StartTimer()

				server.ServeHTTP(w, r)

				b.StopTimer()
				res = w.Result()
				body, err := io.ReadAll(res.Body)
				if err != nil {
					b.Errorf("Response body wasn't read, code %s, err: %v", res.Status, err)
				}
				// XXX: statictest - "response body must be closed"
				// statictest ругается на незакрытое тело, но зачем его
				// закрывать, если httptest.NewRecorder().Result() возвращает
				// NopCloser Body?? Пришлось городить какую-то грязь.
				// Что ли прямо в цикле бенча ставить defer res.Body.Close()?
				res.Body.Close()

				if tc.StatusCode(w.Code) {
					// body, err := io.ReadAll(res.Body)
					// if err != nil {
					// 	b.Errorf("Response body wasn't read, code %s, err: %v", res.Status, err)
					// }

					b.Errorf("Unexpected response code: %s, body: %s", res.Status, body)
				}
			}
		})
	}
}
