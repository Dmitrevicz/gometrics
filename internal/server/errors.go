package server

import "errors"

// http error messages (or msg to be logged)
const (
	ErrMsgNothingFound     = "Nothing found"
	ErrMsgMethodOnlyGET    = "Only GET requests are allowed"
	ErrMsgMethodOnlyPOST   = "Only POST requests are allowed"
	ErrMsgWrongMetricType  = "Wrong metric type"
	ErrMsgEmptyMetricName  = "Empty metric name"
	ErrMsgWrongMetricValue = "Wrong metric value"
	ErrMsgNegativeCounter  = "Counter value must not be negative"
	ErrMsgTemplateExec     = "Error executing template"
	ErrMsgStorageFail      = "Storage error"
	ErrMsgDumperFail       = "Dumper failed"
)

// statictest туле очень не понравились ошибки начинающиеся с большой буквы
var (
	ErrWrongMetricType  = errors.New("wrong metric type")
	ErrEmptyMetricName  = errors.New("empty metric name")
	ErrWrongMetricValue = errors.New("wrong metric value")
	ErrNegativeCounter  = errors.New("counter value must not be negative")
)
