package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"agentsight/internal/config"
	"agentsight/internal/handlers"
	"agentsight/internal/middleware"
	"agentsight/internal/repository"
	"agentsight/internal/scraper"
	"agentsight/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg := config.Load()
	slog.Info("starting AgentSight server", "port", cfg.AppPort, "env", cfg.Env)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := repository.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "err", err, "db_url", cfg.DatabaseURL)
		os.Exit(1)
	}
	defer pool.Close()

	if err := repository.RunMigrations(ctx, pool); err != nil {
		slog.Error("database migration error", "err", err)
		os.Exit(1)
	}

	repos := repository.NewPostgresRepositories(pool)
	services := service.NewServices(repos, cfg)

	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	scraperScheduler := scraper.NewScheduler(repos, cfg)
	scraperScheduler.Start(appCtx)
	defer scraperScheduler.Stop()

	rateLimiter := middleware.NewRateLimiter(60, time.Minute)
	defer rateLimiter.Close()

	h := handlers.NewHandlers(services, repos, pool)
	router := chi.NewRouter()
	handlers.RegisterRoutes(router, h, services.Auth, rateLimiter)

	server := &http.Server{
		Addr:         cfg.AppPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("server listening", "addr", cfg.AppPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server listen error", "err", err)
			os.Exit(1)
		}
	}()

	<-shutdownCh
	slog.Info("shutdown signal received, shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server graceful shutdown failed", "err", err)
	}

	slog.Info("server stopped")
}
