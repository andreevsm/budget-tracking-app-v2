package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"budget/internal/domain"
)

func (h *Handler) createOperation(w http.ResponseWriter, r *http.Request) {
	var in domain.CreateOperationInput

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&in)

	if err != nil {
		h.logger.Warn("failed to decode operation", "error", err)
		http.Error(w, "failed to decode operation", http.StatusBadRequest)
		return
	}

	operation, err := h.serv.CreateOperation(r.Context(), in)

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

		case errors.Is(err, domain.ErrAccountNotFound):
			h.logger.Warn("account not found", "error", err)
			http.Error(w, "account not found", http.StatusBadRequest)

		default:
			h.logger.Error("create operation failed", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}

		return
	}

	h.logger.Info("operation created", "id", operation.ID)
	h.writeJSON(w, http.StatusCreated, operation)
}

func (h *Handler) listOperationsByAccount(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	accountID := r.PathValue("accountID")

	operations, err := h.serv.ListOperationsByAccount(r.Context(), userID, accountID)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			h.logger.Warn("wrong input", "error", err)
			http.Error(w, "wrong input", http.StatusBadRequest)

		case errors.Is(err, domain.ErrAccountNotFound):
			h.logger.Warn("account not found", "error", err)
			http.Error(w, "account not found", http.StatusNotFound)

		default:
			h.logger.Error("get operations failed", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}

		return
	}

	h.writeJSON(w, http.StatusOK, operations)
}