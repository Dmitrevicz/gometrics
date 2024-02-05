package main

import "os"

func main() {
	_ = etc()
	os.Exit(1) // want "os.Exit must not be called from main"
}

func Other() {
	os.Exit(1)
}

func etc() int {
	return 2 + 3
}
