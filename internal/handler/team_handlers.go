package handler

import (
	"encoding/json"
	"net/http"

	"github.com/egoisthemain/pr-reviewer/internal/domain"
)

func (h *Handler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var t domain.Team

	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.tm.CreateTeam(r.Context(), t); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"team": t,
	})
}
