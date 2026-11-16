package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"budget/internal/domain"
)

func (h *Handler) createAccount(w http.ResponseWriter, r *http.Request) {
	var in domain.CreateAccountInput

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&in)

	if err != nil {
		h.logger.Warn("failed to decode account", "error", err)
		http.Error(w, "failed to decode account", http.StatusBadRequest)
		return
	}

	account, err := h.serv.CreateAccount(r.Context(), in)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDuplicateName):
			h.logger.Warn("duplicate name rejected", "error", err)
			http.Error(w, "name already exists", http.StatusConflict)

		case errors.Is(err, domain.ErrInvalidInput):
			h.logger.Warn("wrong input", "error", err)
			http.Error(w, "wrong input", http.StatusBadRequest)

		case errors.Is(err, domain.ErrInvalidReference):
			h.logger.Warn("invalid reference", "error", err)
			http.Error(w, "invalid reference", http.StatusBadRequest)

		case errors.Is(err, domain.ErrUserNotFound):
			h.logger.Warn("user not found", "error", err)
			http.Error(w, "user not found", http.StatusBadRequest)

		default:
			h.logger.Error("create account failed", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}

		return
	}

	h.logger.Info("account created", "id", account.ID)
	h.writeJSON(w, http.StatusCreated, account)
}

func (h *Handler) listAccounts(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")

	accounts, err := h.serv.ListAccounts(r.Context(), userID)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			h.logger.Warn("wrong input", "error", err)
			http.Error(w, "wrong input", http.StatusBadRequest)

		case errors.Is(err, domain.ErrUserNotFound):
			h.logger.Warn("user not found", "error", err)
			http.Error(w, "user not found", http.StatusNotFound)

		default:
			h.logger.Error("get accounts failed", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}

		return
	}

	h.writeJSON(w, http.StatusOK, accounts)
}