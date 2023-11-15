package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/Dmitrevicz/gometrics/internal/agent"
	"github.com/Dmitrevicz/gometrics/internal/logger"
)

// server address for metrics to be sent to
var urlServer string = "http://localhost:8080"

// interval in seconds
var (
	pollInterval   int
	reportInterval int
)

// Shows if metrics update request should be sent in single batch.
// Default is true
var batch = true

func main() {
	parseFlags()
	checkEnvs()

	if err := logger.Initialize(""); err != nil {
		log.Fatalln("failed initializing logger:", err)
	}
	defer logger.Sync()

	agent := agent.New(pollInterval, reportInterval, urlServer, batch)
	agent.Start()

	waitExit()
}

func waitExit() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	s := <-quit
	log.Printf("Agent was stopped with signal: %v\n", s)
}

func parseFlags() {
	// flag.StringVar(&urlServer, "a", "http://localhost:8080", "api endpoint address")
	flag.IntVar(&pollInterval, "p", 2, "poll interval in seconds")
	flag.IntVar(&reportInterval, "r", 10, "report interval in seconds")
	flag.BoolVar(&batch, "batch", batch, "send metrics update request in single batch")

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

func checkEnvs() {
	var err error

	if e, ok := os.LookupEnv("ADDRESS"); ok {
		urlServer = e

		// buggy auto-tests workaround (same as for flags)
		e = strings.TrimSpace(e)
		if !(strings.Contains(e, "http://") || strings.Contains(e, "https://")) {
			log.Printf("Provided ENV ADDRESS=\"%s\" lacks protocol scheme, attempt to fix it will be made\n", e)
			urlServer = "http://" + e
		}
	}

	if e, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		reportInterval, err = strconv.Atoi(e)
		if err != nil {
			log.Fatalln("Error parsing REPORT_INTERVAL from env: ", err)
			return
		}
	}

	if e, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		pollInterval, err = strconv.Atoi(e)
		if err != nil {
			log.Fatalln("Error parsing POLL_INTERVAL from env: ", err)
			return
		}
	}

	if e, ok := os.LookupEnv("BATCH"); ok {
		v, err := strconv.ParseBool(e)
		if err != nil {
			log.Fatalln("Error parsing BATCH from env: ", err)
			return
		}
		batch = v
	}
}
