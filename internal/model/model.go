// Package model contains structures used by services.
//
// Also holds custom errors definitions.
package model

import "strconv"

// metric type name
const (
	MetricTypeGauge   = "gauge"
	MetricTypeCounter = "counter"
)

// metric value
type (
	Gauge   float64
	Counter int64
)

// FromString parses gauge value from string s.
func (g Gauge) FromString(s string) (Gauge, error) {
	v, err := strconv.ParseFloat(s, 64)

	return Gauge(v), err
}

// FromString parses counter value from string s.
func (g Counter) FromString(s string) (Counter, error) {
	v, err := strconv.ParseInt(s, 10, 64)

	return Counter(v), err
}

// Metrics - struct from the lesson.
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

// MetricGauge represents metric of specific Gauge type.
type MetricGauge struct {
	Name  string
	Value Gauge
}

// MetricCounter represents metric of specific Counter type.
type MetricCounter struct {
	Name  string
	Value Counter
}
