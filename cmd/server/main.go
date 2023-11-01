package main

import (
	"context"
	"errors"
	"flag"
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

func main() {
	cfg := config.New()
	if err := parseFlags(cfg); err != nil {
		log.Fatalln("failed parsing flags:", err)
	}

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalln("failed initializing logger:", err)
	}
	defer logger.Sync()

	srv := server.New()
	s := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: srv,
	}

	dumper := server.NewDumper(srv.Storage, cfg)
	if err := dumper.Start(); err != nil {
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

	waitShutdown(s, dumper)
}

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

func parseFlags(cfg *config.Config) error {
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "TCP address for the server to listen on")
	flag.StringVar(&cfg.LogLevel, "loglvl", cfg.LogLevel, "logger level")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "file path for metrics data to be dumped in")
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
