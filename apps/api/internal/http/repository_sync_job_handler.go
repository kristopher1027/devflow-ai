package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/kristopher1027/devflow-ai/internal/repository"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

type RepositorySyncJobHandler struct {
	service service.RepositorySyncJobService
}

func NewRepositorySyncJobHandler(
	syncJobService service.RepositorySyncJobService,
) *RepositorySyncJobHandler {
	return &RepositorySyncJobHandler{
		service: syncJobService,
	}
}

func (h *RepositorySyncJobHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	jobID := repositorySyncJobIDFromPath(r.URL.Path)
	if jobID == "" {
		http.Error(w, "repository sync job ID is required", http.StatusBadRequest)
		return
	}

	requesterID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	job, err := h.service.FindByID(
		r.Context(),
		requesterID,
		jobID,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRepositorySyncJobUnauthorized):
			http.Error(w, "forbidden", http.StatusForbidden)

		case errors.Is(err, repository.ErrRepositorySyncJobNotFound):
			http.Error(w, "repository sync job not found", http.StatusNotFound)

		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(job); err != nil {
		return
	}
}

func repositorySyncJobIDFromPath(path string) string {
	const prefix = "/repositories/sync-jobs/"

	if !strings.HasPrefix(path, prefix) {
		return ""
	}

	id := strings.TrimPrefix(path, prefix)
	id = strings.Trim(id, "/")

	return id
}
