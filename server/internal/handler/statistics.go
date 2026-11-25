package handler

import (
	"budget/internal/domain"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

func (h *Handler) getAccountStatistics(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	accountID := r.PathValue("accountID")
	accountUUID, err := uuid.Parse(accountID)

	if err != nil {
		http.Error(w, "wrong input", http.StatusBadRequest)
		return
	}

	from := domain.Date(r.URL.Query().Get("from"))
	to := domain.Date(r.URL.Query().Get("to"))

	bucket := domain.StatBucket(r.URL.Query().Get("bucket"))
	if bucket == "" {
		bucket = domain.BucketMonth
	}

	stats, err := h.serv.AccountStatistics(r.Context(), userID, accountUUID, from, to, bucket)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			h.logger.Warn("wrong input", "error", err)
			http.Error(w, "wrong input", http.StatusBadRequest)

		case errors.Is(err, domain.ErrAccountNotFound):
			h.logger.Warn("account not found", "error", err)
			http.Error(w, "account not found", http.StatusNotFound)

		default:
			h.logger.Error("get account statistics failed", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	h.writeJSON(w, http.StatusOK, stats)
}
