package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/kristopher1027/devflow-ai/internal/repository"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

type ProjectHandler struct {
	service service.ProjectService
}

func NewProjectHandler(
	service service.ProjectService,
) *ProjectHandler {
	return &ProjectHandler{
		service: service,
	}
}

type CreateProjectRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func (h *ProjectHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	requesterID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(
			w,
			"authentication required",
			http.StatusUnauthorized,
		)
		return
	}

	workspaceID := strings.TrimPrefix(
		r.URL.Path,
		"/workspaces/",
	)

	workspaceID = strings.TrimSuffix(
		workspaceID,
		"/projects",
	)

	if workspaceID == "" {
		http.Error(
			w,
			"workspace ID is required",
			http.StatusBadRequest,
		)
		return
	}

	var request CreateProjectRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	project, err := h.service.Create(
		r.Context(),
		requesterID,
		workspaceID,
		request.Name,
		request.Description,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProjectNameRequired):
			http.Error(
				w,
				"project name is required",
				http.StatusBadRequest,
			)

		case errors.Is(err, service.ErrProjectUnauthorized):
			http.Error(
				w,
				"project access denied",
				http.StatusForbidden,
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
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(project)
}

func (h *ProjectHandler) ListByWorkspaceID(
	w http.ResponseWriter,
	r *http.Request,
) {
	requesterID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(
			w,
			"authentication required",
			http.StatusUnauthorized,
		)
		return
	}

	workspaceID := strings.TrimPrefix(
		r.URL.Path,
		"/workspaces/",
	)

	workspaceID = strings.TrimSuffix(
		workspaceID,
		"/projects",
	)

	if workspaceID == "" {
		http.Error(
			w,
			"workspace ID is required",
			http.StatusBadRequest,
		)
		return
	}

	projects, err := h.service.ListByWorkspaceID(
		r.Context(),
		requesterID,
		workspaceID,
	)
	if err != nil {
		if errors.Is(err, service.ErrProjectUnauthorized) {
			http.Error(
				w,
				"project access denied",
				http.StatusForbidden,
			)
			return
		}

		http.Error(
			w,
			"internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(projects)
}

func (h *ProjectHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	requesterID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(
			w,
			"authentication required",
			http.StatusUnauthorized,
		)
		return
	}

	id := strings.TrimPrefix(
		r.URL.Path,
		"/projects/",
	)

	if id == "" {
		http.Error(
			w,
			"project ID is required",
			http.StatusBadRequest,
		)
		return
	}

	project, err := h.service.FindByID(
		r.Context(),
		requesterID,
		id,
	)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrProjectNotFound):
			http.Error(
				w,
				"project not found",
				http.StatusNotFound,
			)

		case errors.Is(err, service.ErrProjectUnauthorized):
			http.Error(
				w,
				"project access denied",
				http.StatusForbidden,
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

	_ = json.NewEncoder(w).Encode(project)
}

func (h *ProjectHandler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	requesterID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(
			w,
			"authentication required",
			http.StatusUnauthorized,
		)
		return
	}

	id := strings.TrimPrefix(
		r.URL.Path,
		"/projects/",
	)

	if id == "" {
		http.Error(
			w,
			"project ID is required",
			http.StatusBadRequest,
		)
		return
	}

	err := h.service.Delete(
		r.Context(),
		requesterID,
		id,
	)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrProjectNotFound):
			http.Error(
				w,
				"project not found",
				http.StatusNotFound,
			)

		case errors.Is(err, service.ErrProjectUnauthorized):
			http.Error(
				w,
				"project access denied",
				http.StatusForbidden,
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

	w.WriteHeader(http.StatusNoContent)
}
