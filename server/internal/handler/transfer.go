package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"budget/internal/domain"

	"github.com/google/uuid"
)

func (h *Handler) createTransfer(w http.ResponseWriter, r *http.Request) {
	var in domain.CreateTransferInput

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&in)

	if err != nil {
		h.logger.Warn("failed to decode transfer", "error", err)
		http.Error(w, "failed to decode transfer", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	transferGroupID, err := h.serv.CreateTransfer(r.Context(), userID, in)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			h.logger.Warn("wrong input", "error", err)
			http.Error(w, "wrong input", http.StatusBadRequest)

		case errors.Is(err, domain.ErrAccountNotFound):
			h.logger.Warn("account not found", "error", err)
			http.Error(w, "account not found", http.StatusNotFound)

		default:
			h.logger.Error("create transfer failed", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}

		return
	}

	h.logger.Info("transfer created", "id", transferGroupID)
	h.writeJSON(w, http.StatusCreated, transferGroupID)
}
