// Package main represents entry point for http server service.
//
// Server service stores runtime metrics gathered by the agent service.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Dmitrevicz/gometrics/internal/logger"
	"github.com/Dmitrevicz/gometrics/internal/server"
	"github.com/Dmitrevicz/gometrics/internal/server/config"
	"github.com/Dmitrevicz/gometrics/internal/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	grpcServer "github.com/Dmitrevicz/gometrics/internal/server/grpc"
	pb "github.com/Dmitrevicz/gometrics/internal/server/grpc/proto"
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

	// print config in purpose to debug autotests
	logger.Log.Sugar().Infof("Server config: %+v", cfg)

	// TODO: decide whether to run both grpc and http servers in parallel or only one of...
	grpcStub(cfg)

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
			logger.Log.Fatal("HTTP [Server.ListenAndServe] failed", zap.Error(err))
		}
	}()

	waitShutdown(s, srv.Dumper, srv.Storage)
}

// waitShutdown implements graceful shutdown.
func waitShutdown(s *http.Server, dumper *server.Dumper, storage storage.Storage) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	sig := <-quit

	const maxShutdownTimeout = 10 * time.Second

	logger.Log.Info("Server caught os signal. Starting shutdown...",
		zap.String("signal", sig.String()),
		zap.Duration("max_timeout", maxShutdownTimeout),
	)

	// XXX: should I disable keep-alive like this on shutdown or server.Shutdown()
	// will handle it by itself?
	// s.SetKeepAlivesEnabled(false)

	ctx, cancel := context.WithTimeout(context.Background(), maxShutdownTimeout/3)
	defer cancel()

	var errs []error

	// 1. Shutdown server
	if err := s.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("HTTP [Server.Shutdown] failed: %v", err))
	}

	// XXX: can I use same context from before or should create new like this?
	ctxDumper, cancelCtxDumper := context.WithTimeout(context.Background(), maxShutdownTimeout/3)
	defer cancelCtxDumper()

	// 2. Stop dumper
	if err := dumper.Quit(ctxDumper); err != nil {
		errs = append(errs, fmt.Errorf("failed to Quit the Dumper: %v", err))
	}

	// XXX: can I use same context from before or should create new like this?
	ctxStorage, cancelCtxStorage := context.WithTimeout(context.Background(), maxShutdownTimeout/3)
	defer cancelCtxStorage()

	// 3. Close storage
	if err := storage.Close(ctxStorage); err != nil {
		errs = append(errs, fmt.Errorf("failed to Close the Storage: %v", err))
	}

	if len(errs) > 0 {
		logger.Log.Fatal("Server was stopped, but errors occurred", zap.Errors("errors", errs))
	}

	logger.Log.Info("Server was stopped")
}

func grpcStub(cfg *config.Config) {
	if cfg.ServerAddressGRPC == "" {
		return
	}

	logger.Log.Info("gRPC port found in config, trying to start gRPC server...")

	listen, err := net.Listen("tcp", cfg.ServerAddressGRPC)
	if err != nil {
		logger.Log.Sugar().Fatalf("Failed to listen on port '%s', err: %v", cfg.ServerAddressGRPC, err)
	}

	metricsServer := grpcServer.NewMetricsServer(cfg)

	// TODO: interceptors
	s := grpc.NewServer()
	pb.RegisterMetricsServer(s, metricsServer)
	reflection.Register(s)

	logger.Log.Sugar().Infof("gRPC server started on %s", cfg.ServerAddressGRPC)
	if err := s.Serve(listen); err != nil {
		logger.Log.Sugar().Fatalf("gRPC server failed to Serve, err: %v", err)
	}

	// TODO: graceful shutdown for grpc server
	// s.GracefulStop()
}
