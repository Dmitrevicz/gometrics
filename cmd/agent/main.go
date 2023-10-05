package main

import (
	"github.com/Dmitrevicz/gometrics/internal/agent"
)

// intervals in seconds
const (
	pollInterval   = 2
	reportInterval = 10
)

func main() {
	agent := agent.New(pollInterval, reportInterval)
	agent.Start()

	c := make(chan struct{})
	<-c
}
