package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/Dmitrevicz/gometrics/internal/model"
	"github.com/Dmitrevicz/gometrics/internal/server/config"
)

// ExampleHandlers_UpdateBatch shows an example of UpdateBatch handler.
func ExampleHandlers_UpdateBatch() {

	reqURL := "/updates/" // handler endpoint url
	gaugeValue := 42.420
	counterValue := 42

	reqBody := bytes.NewBufferString(fmt.Sprintf(`[
		{"id":"TestMetric1","type":"%s","value":%f},
		{"id":"TestMetric2","type":"%s","delta":%d}
	]`, model.MetricTypeGauge, gaugeValue, model.MetricTypeCounter, counterValue))

	r := httptest.NewRequest(http.MethodPost, reqURL, reqBody)
	w := httptest.NewRecorder()

	// creating test server
	server := New(config.NewTesting())
	server.ServeHTTP(w, r)

	// print response status code
	fmt.Printf("code: %d\n", w.Code)

	// handle unexpected errors
	if w.Code != http.StatusOK {
		res := w.Result()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Printf("Unexpected code: %s, error reading response body: %v\n", res.Status, err)
		}
		fmt.Printf("Unexpected code: %s, body: %s\n", res.Status, body)
	}

	// Output:
	// code: 200
}

// ExampleHandlers_GetMetricByJSON shows an example of GetMetricByJSON handler.
func ExampleHandlers_GetMetricByJSON() {

	reqURL := "/value/" // handler endpoint url
	counterValue := int64(42)

	want := model.Metrics{
		ID:    "ExampleCounterName",
		MType: model.MetricTypeCounter,
		Delta: &counterValue,
	}

	// creating test server
	server := New(config.NewTesting())

	// populate expected data
	if err := server.Storage.Counters().Set(want.ID, model.Counter(*want.Delta)); err != nil {
		fmt.Printf("Error populating storage before test: %v\n", err)
		return
	}

	reqBody := bytes.NewBufferString(fmt.Sprintf(`{
		"id":"%s",
		"type":"%s"
	}`, want.ID, want.MType))

	r := httptest.NewRequest(http.MethodPost, reqURL, reqBody)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, r)

	// print response status code
	fmt.Printf("code: %d\n", w.Code) // outputs 1st line of success

	res := w.Result()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Status code: %s, error reading response body: %v\n", res.Status, err)
		return
	}

	if w.Code != http.StatusOK {
		fmt.Printf("Unexpected code: %s, body: %s\n", res.Status, body)
		return
	}

	got := model.Metrics{}
	err = json.Unmarshal(body, &got)
	if err != nil {
		fmt.Printf("Status code: %s, error unmarshaling response body: %v\n", res.Status, err)
		return
	}

	// compare want and got response
	if got.MType != want.MType || got.ID != want.ID || *(got.Delta) != *(want.Delta) {
		fmt.Printf("BAD: got != want. Unexpected response body: %s\n", body)
		return
	}

	fmt.Println("OK: got == want") // outputs 2nd line of success

	// Output:
	// code: 200
	// OK: got == want
}
