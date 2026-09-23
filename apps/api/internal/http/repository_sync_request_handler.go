package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/kristopher1027/devflow-ai/internal/repository"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

type RepositorySyncRequestHandler struct {
	service service.RepositorySyncRequestService
}

func NewRepositorySyncRequestHandler(
	syncRequestService service.RepositorySyncRequestService,
) *RepositorySyncRequestHandler {
	return &RepositorySyncRequestHandler{
		service: syncRequestService,
	}
}

func (h *RepositorySyncRequestHandler) RequestSync(
	w http.ResponseWriter,
	r *http.Request,
) {
	repositoryID := repositoryIDFromSyncPath(r.URL.Path)

	if repositoryID == "" {
		http.Error(
			w,
			"repository ID is required",
			http.StatusBadRequest,
		)
		return
	}

	requesterID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(
			w,
			"unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	job, err := h.service.RequestSync(
		r.Context(),
		requesterID,
		repositoryID,
	)
	if err != nil {
		switch {
		case errors.Is(
			err,
			service.ErrRepositorySyncRequestUnauthorized,
		):
			http.Error(
				w,
				"forbidden",
				http.StatusForbidden,
			)

		case errors.Is(
			err,
			repository.ErrRepositoryNotFound,
		):
			http.Error(
				w,
				"repository not found",
				http.StatusNotFound,
			)

		case errors.Is(
			err,
			repository.ErrProjectNotFound,
		):
			http.Error(
				w,
				"project not found",
				http.StatusNotFound,
			)

		default:
			http.Error(
				w,
				"internal server error",
				http.StatusInternalServerError,
			)
		}

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	if err := json.NewEncoder(w).Encode(job); err != nil {
		return
	}
}

func repositoryIDFromSyncPath(path string) string {
	const prefix = "/repositories/"

	if !strings.HasPrefix(path, prefix) {
		return ""
	}

	id := strings.TrimPrefix(path, prefix)

	const suffix = "/sync"

	if !strings.HasSuffix(id, suffix) {
		return ""
	}

	id = strings.TrimSuffix(id, suffix)
	id = strings.Trim(id, "/")

	return id
}
