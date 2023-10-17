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

	"github.com/Dmitrevicz/gometrics/internal/server"
)

// address for the server to listen on
var serverAddress string

func main() {
	parseFlags()

	s := &http.Server{
		Addr:    serverAddress,
		Handler: server.New(),
	}

	go func() {
		log.Printf("Starting Server on %s", serverAddress)
		if err := s.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	waitShutdown(s)
}

func waitShutdown(s *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.Printf("Server caught signal: %v. Starting shutdown...\n", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Fatalln("Shutdown got error:", err)
	}

	log.Println("Server was stopped")
}

func parseFlags() {
	flag.StringVar(&serverAddress, "a", "localhost:8080", "TCP address for the server to listen on")
	flag.Parse()

	// get from env if exist
	if e, ok := os.LookupEnv("ADDRESS"); ok {
		serverAddress = e
	}
}
