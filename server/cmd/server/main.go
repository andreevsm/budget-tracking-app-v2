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
	"budget/internal/domain"
	"budget/internal/repository"
	"budget/internal/service"
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
	serv := service.NewService(repo)

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
		currencies, err := repo.ListCurrencies(r.Context())

		if err != nil {
			logger.Error("failed to get currencies", "error", err)
			http.Error(w, "failed to get currencies", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(currencies)
	})

	mux.HandleFunc("POST /api/v1/accounts", func(w http.ResponseWriter, r *http.Request) {
		var in domain.CreateAccountInput

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		err := decoder.Decode(&in)

		if err != nil {
			logger.Warn("failed to decode account", "error", err)
			http.Error(w, "failed to decode account", http.StatusBadRequest)
			return
		}

		account, err := serv.CreateAccount(r.Context(), in)

		if err != nil {
			switch {
			case errors.Is(err, domain.ErrDuplicateName):
				logger.Warn("duplicate name rejected", "error", err)
				http.Error(w, "name already exists", http.StatusConflict)

			case errors.Is(err, domain.ErrInvalidInput):
				logger.Warn("wrong input", "error", err)
				http.Error(w, "wrong input", http.StatusBadRequest)

			case errors.Is(err, domain.ErrInvalidReference):
				logger.Warn("invalid reference", "error", err)
				http.Error(w, "invalid reference", http.StatusBadRequest)

			case errors.Is(err, domain.ErrUserNotFound):
				logger.Warn("user not found", "error", err)
				http.Error(w, "user not found", http.StatusBadRequest)

			default:
				logger.Error("create account failed", "error", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
			}

			return
		}

		logger.Info("account created", "id", account.ID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(account)
	})

	mux.HandleFunc("GET /api/v1/users/{id}/accounts", func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")

		accounts, err := serv.ListAccounts(r.Context(), userID)

		if err != nil {
			switch {
			case errors.Is(err, domain.ErrInvalidInput):
				logger.Warn("wrong input", "error", err)
				http.Error(w, "wrong input", http.StatusBadRequest)

			case errors.Is(err, domain.ErrUserNotFound):
				logger.Warn("user not found", "error", err)
				http.Error(w, "user not found", http.StatusNotFound)

			default:
				logger.Error("get accounts failed", "error", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
			}

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(accounts)
	})

	mux.HandleFunc("POST /api/v1/categories", func(w http.ResponseWriter, r *http.Request) {
		var in domain.CreateCategoryInput

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		err := decoder.Decode(&in)

		if err != nil {
			logger.Warn("failed to decode category", "error", err)
			http.Error(w, "failed to decode category", http.StatusBadRequest)
			return
		}

		category, err := serv.CreateCategory(r.Context(), in)

		if err != nil {
			switch {
			case errors.Is(err, domain.ErrDuplicateName):
				logger.Warn("duplicate name rejected", "error", err)
				http.Error(w, "name already exists", http.StatusConflict)

			case errors.Is(err, domain.ErrInvalidInput):
				logger.Warn("wrong input", "error", err)
				http.Error(w, "wrong input", http.StatusBadRequest)

			case errors.Is(err, domain.ErrInvalidReference):
				logger.Warn("invalid reference", "error", err)
				http.Error(w, "invalid reference", http.StatusBadRequest)

			case errors.Is(err, domain.ErrUserNotFound):
				logger.Warn("user not found", "error", err)
				http.Error(w, "user not found", http.StatusBadRequest)

			default:
				logger.Error("create category failed", "error", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
			}

			return
		}

		logger.Info("category created", "id", category.ID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(category)
	})

	mux.HandleFunc("GET /api/v1/users/{id}/categories", func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")

		categories, err := serv.ListCategories(r.Context(), userID)

		if err != nil {
			switch {
			case errors.Is(err, domain.ErrInvalidInput):
				logger.Warn("wrong input", "error", err)
				http.Error(w, "wrong input", http.StatusBadRequest)

			case errors.Is(err, domain.ErrUserNotFound):
				logger.Warn("user not found", "error", err)
				http.Error(w, "user not found", http.StatusNotFound)

			default:
				logger.Error("get categories failed", "error", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
			}

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(categories)
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
