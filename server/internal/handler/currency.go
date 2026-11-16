package handler

import "net/http"

func (h *Handler) listCurrencies(w http.ResponseWriter, r *http.Request) {
	currencies, err := h.serv.ListCurrencies(r.Context())

	if err != nil {
		h.logger.Error("failed to get currencies", "error", err)
		http.Error(w, "failed to get currencies", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusOK, currencies)
}