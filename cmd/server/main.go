package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Dmitrevicz/gometrics/internal/logger"
	"github.com/Dmitrevicz/gometrics/internal/server"
	"go.uber.org/zap"
)

var (
	// address for the server to listen on
	serverAddress string

	// logger level
	logLevel string
)

func main() {
	parseFlags()

	if err := logger.Initialize(logLevel); err != nil {
		log.Fatalln(err)
	}
	defer logger.Sync()

	s := &http.Server{
		Addr:    serverAddress,
		Handler: server.New(),
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

	waitShutdown(s)
}

func waitShutdown(s *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	logger.Log.Info("Server caught os signal. Starting shutdown...\n",
		zap.String("signal", sig.String()),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		logger.Log.Fatal("Shutdown got error",
			zap.Error(err),
		)
	}

	logger.Log.Info("Server was stopped")
}

func parseFlags() {
	flag.StringVar(&serverAddress, "a", "localhost:8080", "TCP address for the server to listen on")
	flag.StringVar(&logLevel, "loglvl", "info", "logger level")
	flag.Parse()

	// get from env if exist
	if e, ok := os.LookupEnv("ADDRESS"); ok {
		serverAddress = e
	}

	if e, ok := os.LookupEnv("LOG_LVL"); ok {
		logLevel = e
	}
}
