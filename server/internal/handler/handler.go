package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"budget/internal/service"
)

type Handler struct {
	serv   *service.Service
	logger *slog.Logger
}

func New(serv *service.Service, logger *slog.Logger) *Handler {
	return &Handler{serv: serv, logger: logger}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/currencies", h.listCurrencies)
	mux.HandleFunc("POST /api/v1/accounts", h.createAccount)
	mux.HandleFunc("GET /api/v1/users/{id}/accounts", h.listAccounts)
	mux.HandleFunc("POST /api/v1/categories", h.createCategory)
	mux.HandleFunc("GET /api/v1/users/{id}/categories", h.listCategories)
	mux.HandleFunc("POST /api/v1/transactions", h.createOperation)
	mux.HandleFunc("GET /api/v1/users/{userID}/accounts/{accountID}/transactions", h.listOperationsByAccount)
	mux.HandleFunc("POST /api/v1/transfers", h.createTransfer)
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.logger.Error("write json failed", "error", err)
	}
}