package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ashira-C/ku-acoustic-pwa/api/internal/config"
	"github.com/Ashira-C/ku-acoustic-pwa/api/internal/db"
	"github.com/Ashira-C/ku-acoustic-pwa/api/internal/health"
	"github.com/Ashira-C/ku-acoustic-pwa/api/internal/part"
	"github.com/Ashira-C/ku-acoustic-pwa/api/internal/platform"
	"github.com/Ashira-C/ku-acoustic-pwa/api/internal/server"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server exited with error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	logger := platform.NewLogger(cfg.IsProduction())
	slog.SetDefault(logger)

	// Cancelled on Ctrl+C (local) or SIGTERM (sent by Render/Docker on deploy/stop).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := platform.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	rdb, err := platform.NewRedisClient(ctx, cfg.RedisURL)
	if err != nil {
		return err
	}
	defer rdb.Close()

	// Wiring: repository -> service -> handler.
	queries := db.New(pool)
	partHandler := part.NewHandler(part.NewService(part.NewRepository(queries)), logger)

	healthHandler := health.NewHandler(map[string]health.CheckFunc{
		"postgres": pool.Ping,
		"redis": func(ctx context.Context) error {
			return rdb.Ping(ctx).Err()
		},
	})

	router := server.NewRouter(server.Deps{
		Logger: logger,
		Health: healthHandler,
		Parts:  partHandler,
	})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("starting API", "env", cfg.AppEnv, "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
		logger.Info("shutdown signal received, draining requests")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	logger.Info("server stopped")
	return nil
}