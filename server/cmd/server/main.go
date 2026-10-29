package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"budget/internal/config"
	"budget/internal/db"
	"budget/internal/repository"
)

func main() {
	conf, err := config.Load()

	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	defer stop()

	pool, poolErr := db.New(ctx, conf.DSN)

	if poolErr != nil {
		log.Fatalf("connection pool failed: %v", poolErr)
	}

	defer pool.Close()

	repo := repository.NewRepository(pool)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		err := pool.Ping(r.Context())

		if err != nil {
			log.Printf("health check failed: %v", err)
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
			log.Printf("failed to get currencies: %v", err)
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

	go func() {
		err := srv.ListenAndServe()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}()

	<-ctx.Done()
	log.Printf("Shutting down")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)

	defer cancel()

	errShutdown := srv.Shutdown(shutdownCtx)

	if errShutdown != nil {
		log.Printf("shutdown failed: %v", errShutdown)

		_ = srv.Close()
	}

	log.Printf("Server stopped")
}
