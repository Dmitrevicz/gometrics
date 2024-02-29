// Package main represents entry point for http server service.
//
// Server service stores runtime metrics gathered by the agent service.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
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
	if err := parseConfig(cfg); err != nil {
		log.Fatalln("failed parsing config:", err)
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
