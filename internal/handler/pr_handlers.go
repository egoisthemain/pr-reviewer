package handler

import (
	"encoding/json"
	"net/http"

	"github.com/egoisthemain/pr-reviewer/internal/service"
)

type Handler struct {
	pr *service.PRService
	tm *service.TeamService
}

func New(pr *service.PRService, tm *service.TeamService) *Handler {
	return &Handler{pr: pr, tm: tm}
}

func (h *Handler) CreatePR(w http.ResponseWriter, r *http.Request) {
	type Req struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}

	type Resp struct {
		PR interface{} `json:"pull_request"`
	}

	var req Req
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pr, err := h.pr.CreatePR(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(Resp{PR: pr})
}

func (h *Handler) MergePR(w http.ResponseWriter, r *http.Request) {
	type Req struct {
		PullRequestID string `json:"pull_request_id"`
	}

	type Resp struct {
		PR interface{} `json:"pull_request"`
	}

	var req Req
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pr, err := h.pr.MergePR(r.Context(), req.PullRequestID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(Resp{PR: pr})
}

func (h *Handler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	type Req struct {
		PullRequestID string `json:"pull_request_id"`
		OldReviewerID string `json:"old_reviewer_id"`
	}

	type OKResp struct {
		NewReviewerID string `json:"new_reviewer_id"`
	}

	type ErrResp struct {
		Error string `json:"error"`
	}

	var req Req
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newID, err := h.pr.ReassignReviewer(r.Context(), req.PullRequestID, req.OldReviewerID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrResp{Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(OKResp{NewReviewerID: newID})
}

func (h *Handler) ListReviews(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	prs, err := h.pr.ListPRByReviewer(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	resp := map[string]interface{}{
		"pull_requests": prs,
	}

	json.NewEncoder(w).Encode(resp)
}
