package http

import (
	"encoding/json"
	"net/http"

	"github.com/egoisthemain/pr-reviewer/internal/domain"
)

func (s *Server) handleCreateTeam(w http.ResponseWriter, r *http.Request) {
	var req CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	team := domain.Team{
		TeamName: req.TeamName,
	}
	for _, m := range req.Members {
		team.Members = append(team.Members, domain.User{
			UserID:   m.UserID,
			Username: m.Username,
			TeamName: req.TeamName,
			IsActive: m.IsActive,
		})
	}

	if err := s.TeamService.CreateTeam(r.Context(), team); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // потом поменяем под ErrorResponse
		return
	}

	resp := TeamDTO{
		TeamName: req.TeamName,
		Members:  req.Members,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"team": resp,
	})
}

func (s *Server) handleGetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		http.Error(w, "team_name is required", http.StatusBadRequest)
		return
	}

	team, err := s.TeamService.GetTeam(r.Context(), teamName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound) // потом заменим на нормальный формат
		return
	}

	resp := TeamDTO{
		TeamName: team.TeamName,
		Members:  make([]TeamMemberDTO, 0, len(team.Members)),
	}
	for _, u := range team.Members {
		resp.Members = append(resp.Members, TeamMemberDTO{
			UserID:   u.UserID,
			Username: u.Username,
			IsActive: u.IsActive,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"team": resp,
	})
}
