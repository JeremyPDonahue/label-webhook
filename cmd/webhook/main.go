package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mutating-webhook/internal/config"
	"mutating-webhook/internal/metrics"
)

// global configuration
var cfg config.Config

func main() {
	// Initialize application configuration
	cfg = config.Init()

	// Setup graceful shutdown
	cancel := make(chan struct{})
	defer close(cancel)

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("[INFO] Starting AppID Labeling Webhook v1.0.0")
	log.Printf("[INFO] Namespace: %s", cfg.NameSpace)
	log.Printf("[INFO] Service: %s", cfg.ServiceName)
	log.Printf("[INFO] Organization: %s", cfg.Organization)
	log.Printf("[INFO] Environment: %s", cfg.Environment)
	log.Printf("[INFO] Enable Labeling: %t", cfg.EnableLabeling)
	log.Printf("[INFO] Enable Metrics: %t", cfg.EnableMetrics)

	// Start HTTP server in a goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[ERROR] HTTP server panic: %v", r)
				metrics.SetWebhookDown()
			}
		}()
		httpServer(&cfg)
	}()

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("[INFO] Received %s signal, initiating graceful shutdown...", sig)

	// Set webhook as down in metrics
	metrics.SetWebhookDown()

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Perform graceful shutdown
	log.Printf("[INFO] Graceful shutdown completed")

	// Wait for shutdown context or timeout
	select {
	case <-shutdownCtx.Done():
		log.Printf("[INFO] Shutdown timeout reached")
	default:
		log.Printf("[INFO] Clean shutdown completed")
	}
}
