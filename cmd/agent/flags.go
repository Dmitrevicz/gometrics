package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Dmitrevicz/gometrics/internal/agent/config"
)

// TODO: might use some 3rd-party packages to parse parameters into config
// fields by specifying struct tags, e.g.
//  - github.com/caarlos0/env/v10
//  - github.com/integrii/flaggy
//
//  struct Config {
//      Addr `json:"addr" env:"ADDRESS" flag:"address" flag-short:"a"`
//  }

// parseConfig parses config from env/flags/file.
//
// Had to do parse flags/envs twice this way to achieve
// required priority of parameters parsed:
//  1. env (highest priority, overrides everything else below)
//  2. command-line arguments (flags)
//  3. config file
func parseConfig(cfg *config.Config) {

	// retrieve cofig file path first from flag/env
	parseFlags(cfg)
	parseEnvs(cfg)

	if cfg.ConfigPath != "" {
		if err := config.ParseFromFile(cfg, cfg.ConfigPath); err != nil {
			log.Fatalln("Error parsing config from file: ", err)
		}

		// reset default flag set to avoid "flag redefined" panic
		// (because I define flags directly on config fields)
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		// read from flags and env again to achieve required priority
		parseFlags(cfg)
		parseEnvs(cfg)
	}
}

// parseFlags parses config from flags.
func parseFlags(cfg *config.Config) {

	// config flag definition with shorthand
	flagNameConfig := "config"
	usageConfig := "path to config file"
	flag.StringVar(&cfg.ConfigPath, flagNameConfig, "", usageConfig)
	flag.StringVar(&cfg.ConfigPath, flagNameConfig[:1], "",
		fmt.Sprintf("%s (shorthand for -%s)", usageConfig, flagNameConfig),
	)

	// other flags
	// flag.StringVar(&urlServer, "a", "http://localhost:8080", "api endpoint address")
	flag.StringVar(&cfg.GRPCServerURL, "grpc", cfg.GRPCServerURL, "server address that gRPC client must call to")
	flag.StringVar(&cfg.Key, "k", cfg.Key, "hash key")
	flag.IntVar(&cfg.PollInterval, "p", cfg.PollInterval, "poll interval in seconds")
	flag.IntVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "report interval in seconds")
	flag.IntVar(&cfg.RateLimit, "l", cfg.RateLimit, "rate limit (number of max concurrent senders)")
	flag.BoolVar(&cfg.Batch, "batch", cfg.Batch, "send metrics update request in single batch")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "path to file with private key to be used in messages encryption")

	// XXX: [Workaround]
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

// parseEnvs parses environment variables to populate Config.
// Overwrites cfg fields when appropriate env variables exist.
func parseEnvs(cfg *config.Config) {
	var err error

	if e, ok := os.LookupEnv("ADDRESS"); ok {
		cfg.ServerURL = e

		// XXX: [Workaround]
		// buggy auto-tests workaround (same as for flags)
		e = strings.TrimSpace(e)
		if !(strings.Contains(e, "http://") || strings.Contains(e, "https://")) {
			log.Printf("Provided ENV ADDRESS=\"%s\" lacks protocol scheme, attempt to fix it will be made\n", e)
			cfg.ServerURL = "http://" + e
		}
	}

	if e, ok := os.LookupEnv("GRPC"); ok {
		cfg.GRPCServerURL = e
	}

	if e, ok := os.LookupEnv("KEY"); ok {
		cfg.Key = e
	}

	if e, ok := os.LookupEnv("CRYPTO_KEY"); ok {
		cfg.CryptoKey = e
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
