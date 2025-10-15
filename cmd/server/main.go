package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpAdapter "gosper/internal/adapter/inbound/http"
	"gosper/internal/adapter/outbound/model"
	"gosper/internal/adapter/outbound/storage"
	"gosper/internal/adapter/outbound/whispercpp"
	"gosper/internal/config"
	"gosper/internal/usecase"
)

func main() {
	// Load configuration
	cfg := config.FromEnv()
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Initialize use case with dependencies
	transcribeUC := &usecase.TranscribeFile{
		Repo:  &model.FSRepo{BaseURL: cfg.ModelBaseURL},
		Trans: &whispercpp.Transcriber{},
		Store: storage.FS{},
	}

	// Create HTTP server
	httpServer := httpAdapter.NewServer(
		transcribeUC,
		logger,
		httpAdapter.Config{
			Addr:            cfg.Addr,
			ModelDefault:    cfg.Model,
			LanguageDefault: cfg.Language,
		},
	)

	// Start HTTP server in goroutine
	go func() {
		if err := httpServer.Start(); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Println("Shutting down servers...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}

	logger.Println("Server stopped gracefully")
}
