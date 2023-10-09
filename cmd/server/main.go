package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/Dmitrevicz/gometrics/internal/server"
)

// address for the server to listen on
var serverAddress string

func main() {
	parseFlags()

	s := server.New()

	log.Printf("Starting Server on %s", serverAddress)
	if err := http.ListenAndServe(serverAddress, s); err != nil {
		log.Fatalln(err)
	}
}

func parseFlags() {
	flag.StringVar(&serverAddress, "a", "localhost:8080", "TCP address for the server to listen on")
	flag.Parse()

	// get from env if exist
	if e, ok := os.LookupEnv("ADDRESS"); ok {
		serverAddress = e
	}
}
