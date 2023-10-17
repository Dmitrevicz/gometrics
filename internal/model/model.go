package model

import "strconv"

// metric value
type (
	Gauge   float64
	Counter int64
)

// FromString parses gauge value from string s
func (g Gauge) FromString(s string) (Gauge, error) {
	v, err := strconv.ParseFloat(s, 64)

	return Gauge(v), err
}

// FromString parses counter value from string s
func (g Counter) FromString(s string) (Counter, error) {
	v, err := strconv.ParseInt(s, 10, 64)

	return Counter(v), err
}
