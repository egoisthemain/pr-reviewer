package main

import (
	"log"
	"net/http"
	"os"

	"github.com/egoisthemain/pr-reviewer/internal/handler"
	"github.com/egoisthemain/pr-reviewer/internal/repository"
	"github.com/egoisthemain/pr-reviewer/internal/repository/pg"
	"github.com/egoisthemain/pr-reviewer/internal/service"

	"github.com/gorilla/mux"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:15432/prsvc?sslmode=disable"
	}

	db, err := repository.NewPostgres(dsn)
	if err != nil {
		log.Fatalf("db init: %v", err)
	}

	if err := repository.ApplyMigrations(db); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	teamRepo := pg.NewTeamRepository(db)
	prRepo := pg.NewPRRepository(db)

	teamService := service.NewTeamService(teamRepo)
	prService := service.NewPRService(prRepo, teamRepo)

	h := handler.New(prService, teamService)

	router := mux.NewRouter()

	router.HandleFunc("/pullRequest/create", h.CreatePR).Methods("POST")
	router.HandleFunc("/pullRequest/merge", h.MergePR).Methods("POST")
	router.HandleFunc("/pullRequest/reassign", h.ReassignReviewer).Methods("POST")
	router.HandleFunc("/users/getReview", h.ListReviews).Methods("GET")
	router.HandleFunc("/team/add", h.CreateTeam).Methods("POST")

	router.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	addr := ":8080"
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
