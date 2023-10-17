package main

import (
	"github.com/Dmitrevicz/gometrics/internal/agent"
)

// intervals in seconds
const (
	pollInterval   = 2
	reportInterval = 10
)

// server address for metrics to be sent to
const url = "http://localhost:8080"

func main() {
	agent := agent.New(pollInterval, reportInterval, url)
	agent.Start()

	c := make(chan struct{})
	<-c
}
