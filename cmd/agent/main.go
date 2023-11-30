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
	"github.com/Dmitrevicz/gometrics/internal/agent/config"
	"github.com/Dmitrevicz/gometrics/internal/logger"
)

func main() {
	cfg := config.New()
	parseFlags(cfg)
	checkEnvs(cfg)

	if err := logger.Initialize(""); err != nil {
		log.Fatalln("failed initializing logger:", err)
	}
	defer logger.Sync()

	logger.Log.Sugar().Infof("Agent config: %+v", cfg)

	agent := agent.New(cfg)
	agent.Start()

	waitExit()
}

func waitExit() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	s := <-quit
	log.Printf("Agent was stopped with signal: %v\n", s)
}

func parseFlags(cfg *config.Config) {
	// flag.StringVar(&urlServer, "a", "http://localhost:8080", "api endpoint address")
	flag.StringVar(&cfg.Key, "k", cfg.Key, "hash key")
	flag.IntVar(&cfg.PollInterval, "p", cfg.PollInterval, "poll interval in seconds")
	flag.IntVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "report interval in seconds")
	flag.IntVar(&cfg.RateLimit, "l", cfg.RateLimit, "rate limit (number of max concurrent senders)")
	flag.BoolVar(&cfg.Batch, "batch", cfg.Batch, "send metrics update request in single batch")

	// have to implement a workaround to trick buggy autotests
	flag.Func("a", fmt.Sprintf("api endpoint address (default %s)", cfg.ServerURL), func(s string) error {
		s = strings.TrimSpace(s)
		if s == "" {
			return errors.New("url must not be set as empty string")
		}

		// autotests always fail because url is provided without protocol scheme
		if !(strings.Contains(s, "http://") || strings.Contains(s, "https://")) {
			log.Printf("Provided flag -a=\"%s\" lacks protocol scheme, attempt to fix it will be made\n", s)
			cfg.ServerURL = "http://" + s
			return nil
		}

		return nil
	})

	flag.Parse()
}

func checkEnvs(cfg *config.Config) {
	var err error

	if e, ok := os.LookupEnv("ADDRESS"); ok {
		cfg.ServerURL = e

		// buggy auto-tests workaround (same as for flags)
		e = strings.TrimSpace(e)
		if !(strings.Contains(e, "http://") || strings.Contains(e, "https://")) {
			log.Printf("Provided ENV ADDRESS=\"%s\" lacks protocol scheme, attempt to fix it will be made\n", e)
			cfg.ServerURL = "http://" + e
		}
	}

	if e, ok := os.LookupEnv("KEY"); ok {
		cfg.Key = e
	}

	if e, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		cfg.ReportInterval, err = strconv.Atoi(e)
		if err != nil {
			log.Fatalln("Error parsing REPORT_INTERVAL from env: ", err)
			return
		}
	}

	if e, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		cfg.PollInterval, err = strconv.Atoi(e)
		if err != nil {
			log.Fatalln("Error parsing POLL_INTERVAL from env: ", err)
			return
		}
	}

	if e, ok := os.LookupEnv("RATE_LIMIT"); ok {
		cfg.RateLimit, err = strconv.Atoi(e)
		if err != nil {
			log.Fatalln("Error parsing RATE_LIMIT from env: ", err)
			return
		}
	}

	if e, ok := os.LookupEnv("BATCH"); ok {
		v, err := strconv.ParseBool(e)
		if err != nil {
			log.Fatalln("Error parsing BATCH from env: ", err)
			return
		}
		cfg.Batch = v
	}
}
