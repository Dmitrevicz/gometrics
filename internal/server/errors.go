package server

// http error messages
const (
	ErrMsgMethodOnlyGET    = "Only GET requests are allowed"
	ErrMsgMethodOnlyPOST   = "Only POST requests are allowed"
	ErrMsgWrongMetricType  = "Wrong metric type"
	ErrMsgEmptyMetricName  = "Empty metric name"
	ErrMsgWrongMetricValue = "Wrong metric value"
	ErrMsgNegativeCounter  = "Counter value must not be negative"
)
