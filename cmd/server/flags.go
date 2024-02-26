package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Dmitrevicz/gometrics/internal/server/config"
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
func parseConfig(cfg *config.Config) (err error) {

	// retrieve cofig file path first from flag/env
	if err = parseFlags(cfg); err != nil {
		return err
	}

	if cfg.ConfigPath != "" {
		if err := config.ParseFromFile(cfg, cfg.ConfigPath); err != nil {
			return fmt.Errorf("error parsing config from file: %w", err)
		}

		// reset default flag set to avoid "flag redefined" panic
		// (because I define flags directly on config fields)
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		// read from flags and env again to achieve required priority
		if err = parseFlags(cfg); err != nil {
			return err
		}
	}

	return nil
}

// parseFlags parses Config fields from flags or env.
// Environment variables will overwrite flags parameters.
func parseFlags(cfg *config.Config) error {

	// config flag definition with shorthand
	flagNameConfig := "config"
	usageConfig := "path to config file"
	flag.StringVar(&cfg.ConfigPath, flagNameConfig, "", usageConfig)
	flag.StringVar(&cfg.ConfigPath, flagNameConfig[:1], "",
		fmt.Sprintf("%s (shorthand for -%s)", usageConfig, flagNameConfig),
	)

	// other flags
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "TCP address for the server to listen on")
	flag.StringVar(&cfg.LogLevel, "loglvl", cfg.LogLevel, "logger level")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "file path for metrics data to be dumped in")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "data source name to connect to database")
	flag.StringVar(&cfg.Key, "k", cfg.Key, "hash key")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "path to file with public key to be used in messages encryption")
	flag.IntVar(&cfg.StoreInterval, "i", cfg.StoreInterval, "interval in seconds for current metrics data to be dumped into file")
	flag.BoolVar(&cfg.Restore, "r", cfg.Restore, "shows if data restore from file should be made")

	flag.Parse()

	// get from env if exist
	if e, ok := os.LookupEnv("ADDRESS"); ok {
		cfg.ServerAddress = e
	}

	if e, ok := os.LookupEnv("LOG_LVL"); ok {
		cfg.LogLevel = e
	}

	if e, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		cfg.FileStoragePath = e
	}

	if e, ok := os.LookupEnv("DATABASE_DSN"); ok {
		cfg.DatabaseDSN = e
	}

	if e, ok := os.LookupEnv("KEY"); ok {
		cfg.Key = e
	}

	if e, ok := os.LookupEnv("CRYPTO_KEY"); ok {
		cfg.CryptoKey = e
	}

	if e, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		v, err := strconv.Atoi(e)
		if err != nil {
			return errors.New("bad env \"STORE_INTERVAL\": " + err.Error())
		}
		cfg.StoreInterval = v
	}

	if e, ok := os.LookupEnv("RESTORE"); ok {
		v, err := strconv.ParseBool(e)
		if err != nil {
			return errors.New("bad env \"RESTORE\": " + err.Error())
		}
		cfg.Restore = v
	}

	cfg.FileStoragePath = strings.TrimSpace(cfg.FileStoragePath)

	return nil
}
