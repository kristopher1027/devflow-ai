package http

import (
	"net/http"
	"strings"

	"github.com/kristopher1027/devflow-ai/internal/service"
)

type GitHubRepositoryImportHandler struct {
	queue service.GitHubRepositoryImportQueue
}

func NewGitHubRepositoryImportHandler(
	queue service.GitHubRepositoryImportQueue,
) *GitHubRepositoryImportHandler {
	return &GitHubRepositoryImportHandler{queue: queue}
}

func (h *GitHubRepositoryImportHandler) Import(
	w http.ResponseWriter,
	r *http.Request,
) {
	requesterID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	projectID := strings.TrimPrefix(r.URL.Path, "/projects/")
	projectID = strings.TrimSuffix(projectID, "/repositories/import")
	if projectID == "" {
		http.Error(w, "project ID is required", http.StatusBadRequest)
		return
	}

	job, err := h.queue.Enqueue(
		r.Context(),
		service.GitHubRepositoryImportRequest{
			RequesterID: requesterID,
			ProjectID:   projectID,
		},
	)
	if err != nil {
		http.Error(w, "repository import could not be queued", http.StatusServiceUnavailable)
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"job_id": job.ID,
		"status": "queued",
	})
}
