package http

import (
	"net/http"

	"github.com/egoisthemain/pr-reviewer/internal/service"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	Router      chi.Router
	TeamService *service.TeamService
}

func NewServer(teamService *service.TeamService) *Server {
	r := chi.NewRouter()

	s := &Server{
		Router:      r,
		TeamService: teamService,
	}

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Post("/team/add", s.handleCreateTeam)
	r.Get("/team/get", s.handleGetTeam)

	return s
}
