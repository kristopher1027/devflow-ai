package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/kristopher1027/devflow-ai/internal/repository"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

type GitHubRepositoryImportJobHandler struct {
	service service.GitHubRepositoryImportJobService
}

func NewGitHubRepositoryImportJobHandler(
	service service.GitHubRepositoryImportJobService,
) *GitHubRepositoryImportJobHandler {
	return &GitHubRepositoryImportJobHandler{
		service: service,
	}
}
func (h *GitHubRepositoryImportJobHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	requesterID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	id := strings.TrimPrefix(
		r.URL.Path,
		"/github/repository-import-jobs/",
	)

	if id == "" {
		http.Error(
			w,
			"job ID is required",
			http.StatusBadRequest,
		)
		return
	}

	job, err := h.service.FindByID(
		r.Context(),
		requesterID,
		id,
	)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(job)
}

func (h *GitHubRepositoryImportJobHandler) writeServiceError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(
		err,
		service.ErrGitHubRepositoryImportJobUnauthorized,
	):
		http.Error(
			w,
			"job access denied",
			http.StatusForbidden,
		)

	case errors.Is(
		err,
		repository.ErrGitHubRepositoryImportJobNotFound,
	):
		http.Error(
			w,
			"job not found",
			http.StatusNotFound,
		)

	default:
		http.Error(
			w,
			"internal server error",
			http.StatusInternalServerError,
		)
	}
}
