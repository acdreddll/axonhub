package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/axonhub/axonhub/internal/config"
	"github.com/axonhub/axonhub/internal/server"
)

// @title AxonHub API
// @version 1.0
// @description AxonHub — a modern hub for managing and routing event-driven workflows.
// @contact.name AxonHub Contributors
// @license.name MIT
// @BasePath /api/v1
func main() {
	// Load application configuration from environment / config file
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Build the HTTP server (router, middleware, handlers are wired inside)
	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("failed to initialise server: %v", err)
	}

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      srv,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start listening in a separate goroutine so we can handle shutdown signals.
	go func() {
		log.Printf("AxonHub listening on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Block until we receive an interrupt or termination signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")

	// Allow up to 30 seconds for in-flight requests to complete.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}

	log.Println("server stopped cleanly")
}
