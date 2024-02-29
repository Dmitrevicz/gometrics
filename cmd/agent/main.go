// Package main represents entry point for the agent service.
//
// Agent service periodically gathers runtime metrics and sends them to server.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Dmitrevicz/gometrics/internal/agent"
	"github.com/Dmitrevicz/gometrics/internal/agent/config"
	"github.com/Dmitrevicz/gometrics/internal/logger"
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
	parseConfig(cfg)

	if err := logger.Initialize(""); err != nil {
		log.Fatalln("failed initializing logger:", err)
	}
	defer logger.Sync()

	logger.Log.Sugar().Infof("Agent config: %+v", cfg)

	agent, err := agent.New(cfg)
	if err != nil {
		logger.Log.Sugar().Fatalln("failed initializing agent:", err)
	}

	agent.Start()

	waitExit(agent)
}

// waitExit stops the application from exiting instantly.
// Implements graceful shutdown.
func waitExit(agent *agent.Agent) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	s := <-quit

	log.Printf("Agent caught os signal (%v). Starting shutdown...\n", s)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := agent.Shutdown(ctx); err != nil {
		log.Fatalln("Agent was stopped, but error occurred:", err)
	}

	log.Println("Agent was stopped")
}
