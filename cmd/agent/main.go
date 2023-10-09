package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/Dmitrevicz/gometrics/internal/agent"
)

// server address for metrics to be sent to
var urlServer string = "http://localhost:8080"

// interval in seconds
var (
	pollInterval   int
	reportInterval int
)

func main() {
	parseFlags()

	agent := agent.New(pollInterval, reportInterval, urlServer)
	agent.Start()

	c := make(chan struct{})
	<-c
}

func parseFlags() {
	// flag.StringVar(&urlServer, "a", "http://localhost:8080", "api endpoint address")
	flag.IntVar(&pollInterval, "p", 2, "poll interval in seconds")
	flag.IntVar(&reportInterval, "r", 10, "report interval in seconds")

	// have to implement a workaround to trick buggy autotests
	flag.Func("a", fmt.Sprintf("api endpoint address (default %s)", urlServer), func(s string) error {
		s = strings.TrimSpace(s)
		if s == "" {
			return errors.New("url must not be set as empty string")
		}

		// autotests always fail because url is provided without protocol scheme
		if !(strings.Contains(s, "http://") || strings.Contains(s, "https://")) {
			log.Printf("Provided flag -a=\"%s\" lacks protocol scheme, attempt to fix it will be made\n", s)
			urlServer = "http://" + s
			return nil
		}

		return nil
	})

	flag.Parse()
}
