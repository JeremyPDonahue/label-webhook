package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
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

	initialize()

	go httpServer(config.WebSrvIP, config.WebSrvPort)

	forever()
}
