package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	httphandler "pubmed-api/internal/http"
	"pubmed-api/internal/platform"
	"pubmed-api/internal/repo"
	"pubmed-api/internal/service"
	"syscall"
	"time"
)

func main() {
	// Load configuration
	cfg, err := platform.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := platform.NewLogger(cfg.LogLevel)
	logger.Info("starting pubmed-api", "version", "1.0.0", "port", cfg.Port)

	// Initialize repository
	repository, err := repo.NewSQLiteRepository(cfg.DBPath, logger)
	if err != nil {
		logger.Error("failed to create repository", "error", err)
		os.Exit(1)
	}
	defer repository.Close()

	// Load data
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := platform.LoadArticles(ctx, repository, cfg, logger); err != nil {
		logger.Error("failed to load articles", "error", err)
		os.Exit(1)
	}

	// Initialize service
	articleService := service.NewArticleService(repository)

	// Initialize HTTP router
	router := httphandler.NewRouter(articleService, logger)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")

	// Graceful shutdown with timeout
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("server exited")
}

