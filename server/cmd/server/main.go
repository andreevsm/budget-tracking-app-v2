package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"budget/internal/config"
	"budget/internal/db"
	"budget/internal/repository"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	conf, err := config.Load()

	if err != nil {
		logger.Error("load config failed", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	defer stop()

	pool, poolErr := db.New(ctx, conf.DSN)

	if poolErr != nil {
		logger.Error("connection pool failed", "error", poolErr)
		os.Exit(1)
	}

	defer pool.Close()

	repo := repository.NewRepository(pool)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		err := pool.Ping(r.Context())

		if err != nil {
			logger.Error("health check failed", "error", err)
			http.Error(w, "health check failed", http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	})

	mux.HandleFunc("GET /slow", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		time.Sleep(5 * time.Second)

		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	})

	mux.HandleFunc("GET /api/v1/currencies", func(w http.ResponseWriter, r *http.Request) {
		currencies, err := repo.List(r.Context())

		if err != nil {
			logger.Error("failed to get currencies", "error", err)
			http.Error(w, "failed to get currencies", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(currencies)
	})

	srv := &http.Server{
		Addr:              conf.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	logger.Info("server started", "addr", conf.Addr)

	go func() {
		err := srv.ListenAndServe()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)

	defer cancel()

	errShutdown := srv.Shutdown(shutdownCtx)

	if errShutdown != nil {
		logger.Error("shutdown failed", "error", errShutdown)

		_ = srv.Close()
	}

	logger.Info("server stopped")
}
