package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"mutating-webhook/internal/initialize"
)

func forever() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	sig := <-c
	log.Printf("[INFO] shutting down, detected signal: %s", sig)
}

func main() {
	defer func() {
		log.Println("[DEBUG] shutdown sequence complete")
	}()

	// initialize application configuration
	initialize.Init()

	//go httpServer(cfg.WebSrvIP, cfg.WebSrvPort)

	forever()
}
