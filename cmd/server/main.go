// Package main represents entry point for http server service.
//
// Server service stores runtime metrics gathered by the agent service.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Dmitrevicz/gometrics/internal/logger"
	"github.com/Dmitrevicz/gometrics/internal/server"
	"github.com/Dmitrevicz/gometrics/internal/server/config"
	"go.uber.org/zap"
)

// build info
var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

// printVersion prints build info to std.
// Build info might be provided by linker flags while building, e.g.:
//
//	go build -ldflags "-X main.buildVersion=v1.0.1 \
//	-X 'main.buildDate=$(date +'%Y/%m/%d %H:%M:%S')'" main.go
func printVersion() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		buildVersion, buildDate, buildCommit,
	)
}

func main() {
	printVersion()

	cfg := config.New()
	if err := parseFlags(cfg); err != nil {
		log.Fatalln("failed parsing flags:", err)
	}

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalln("failed initializing logger:", err)
	}
	defer logger.Sync()

	logger.Log.Sugar().Infof("Server config: %+v", cfg)

	srv := server.New(cfg)
	s := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: srv,
	}

	if err := srv.Dumper.Start(); err != nil {
		logger.Log.Fatal("dumper start failed", zap.Error(err))
	}

	go func() {
		logger.Log.Info("Starting Server",
			zap.String("addr", s.Addr),
			zap.String("loglvl", logger.Log.Level().String()),
		)

		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("", zap.Error(err))
		}
	}()

	waitShutdown(s, srv.Dumper)
}

// waitShutdown implements graceful shutdown.
func waitShutdown(s *http.Server, dumper *server.Dumper) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	logger.Log.Info("Server caught os signal. Starting shutdown...\n",
		zap.String("signal", sig.String()),
	)

	dumper.Quit()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		logger.Log.Fatal("Shutdown got error",
			zap.Error(err),
		)
	}

	logger.Log.Info("Server was stopped")
}

// parseFlags parses Config fields from flags or env.
// Environment variables will overwrite flags parameters.
func parseFlags(cfg *config.Config) error {
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "TCP address for the server to listen on")
	flag.StringVar(&cfg.LogLevel, "loglvl", cfg.LogLevel, "logger level")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "file path for metrics data to be dumped in")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "data source name to connect to database")
	flag.StringVar(&cfg.Key, "k", cfg.Key, "hash key")
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
