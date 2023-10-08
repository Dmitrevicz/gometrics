package server

// http error messages
const (
	ErrMsgNothingFound     = "Nothing found"
	ErrMsgMethodOnlyGET    = "Only GET requests are allowed"
	ErrMsgMethodOnlyPOST   = "Only POST requests are allowed"
	ErrMsgWrongMetricType  = "Wrong metric type"
	ErrMsgEmptyMetricName  = "Empty metric name"
	ErrMsgWrongMetricValue = "Wrong metric value"
	ErrMsgNegativeCounter  = "Counter value must not be negative"
	ErrMsgTemplateExec     = "Error executing template"
)
