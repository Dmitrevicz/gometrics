package main

import (
	"log"
	"net/http"

	"github.com/Dmitrevicz/gometrics/internal/server"
)

// default TCP address for the server to listen on
const serverAddress = ":8080"

func main() {
	s := server.New()

	log.Printf("Starting Server on %s", serverAddress)
	if err := http.ListenAndServe(serverAddress, s); err != nil {
		log.Fatalln(err)
	}
}
