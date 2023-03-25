package main

import (
	"log"
	"os"
	"syscall"

	"os/signal"

	"mutating-webhook/internal/config"
)

// global configuration
var cfg config.Config

func forever() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	sig := <-c
	log.Printf("[INFO] Received %s signal, shutting down...", sig)
}

func main() {
	defer func() {
		log.Println("[DEBUG] shutdown sequence complete")
	}()

	// initialize application configuration
	cfg = config.Init()

	go httpServer(&cfg)

	forever()
}
