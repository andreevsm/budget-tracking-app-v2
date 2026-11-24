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
	secret string
}

func New(serv *service.Service, logger *slog.Logger, secret string) *Handler {
	return &Handler{serv: serv, logger: logger, secret: secret}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/currencies", requireAuth(h.secret)(http.HandlerFunc(h.listCurrencies)))
	mux.Handle("POST /api/v1/accounts", requireAuth(h.secret)(http.HandlerFunc(h.createAccount)))
	mux.Handle("GET /api/v1/accounts", requireAuth(h.secret)(http.HandlerFunc(h.listAccounts)))
	mux.Handle("POST /api/v1/categories", requireAuth(h.secret)(http.HandlerFunc(h.createCategory)))
	mux.Handle("GET /api/v1/categories", requireAuth(h.secret)(http.HandlerFunc(h.listCategories)))
	mux.Handle("POST /api/v1/transactions", requireAuth(h.secret)(http.HandlerFunc(h.createOperation)))
	mux.Handle("GET /api/v1/accounts/{accountID}/transactions", requireAuth(h.secret)(http.HandlerFunc(h.listOperationsByAccount)))
	mux.Handle("POST /api/v1/transfers", requireAuth(h.secret)(http.HandlerFunc(h.createTransfer)))

	mux.HandleFunc("POST /api/v1/auth/register", h.register)
	mux.HandleFunc("POST /api/v1/auth/login", h.login)
	mux.HandleFunc("POST /api/v1/auth/refresh", h.refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", h.logout)
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.logger.Error("write json failed", "error", err)
	}
}
