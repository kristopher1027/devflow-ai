package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/kristopher1027/devflow-ai/internal/repository"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

type WorkspaceHandler struct {
	service service.WorkspaceService
}

func NewWorkspaceHandler(
	service service.WorkspaceService,
) *WorkspaceHandler {
	return &WorkspaceHandler{
		service: service,
	}
}

type CreateWorkspaceRequest struct {
	Name string `json:"name"`
}

func (h *WorkspaceHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(
			w,
			"authentication required",
			http.StatusUnauthorized,
		)
		return
	}

	var request CreateWorkspaceRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	workspace, err := h.service.Create(
		r.Context(),
		userID,
		request.Name,
	)
	if err != nil {
		if errors.Is(err, service.ErrWorkspaceNameRequired) {
			http.Error(
				w,
				"workspace name is required",
				http.StatusBadRequest,
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
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(workspace)
}

func (h *WorkspaceHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(
			w,
			"authentication required",
			http.StatusUnauthorized,
		)
		return
	}

	workspaces, err := h.service.ListByOwnerID(
		r.Context(),
		userID,
	)
	if err != nil {
		http.Error(
			w,
			"internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(workspaces)
}

func (h *WorkspaceHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := UserIDFromContext(r.Context())
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
		"/workspaces/",
	)

	if id == "" {
		http.Error(
			w,
			"workspace ID is required",
			http.StatusBadRequest,
		)
		return
	}

	workspace, err := h.service.FindByID(
		r.Context(),
		userID,
		id,
	)
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceNotFound) {
			http.Error(
				w,
				"workspace not found",
				http.StatusNotFound,
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

	json.NewEncoder(w).Encode(workspace)
}

func (h *WorkspaceHandler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := UserIDFromContext(r.Context())
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
		"/workspaces/",
	)

	if id == "" {
		http.Error(
			w,
			"workspace ID is required",
			http.StatusBadRequest,
		)
		return
	}

	err := h.service.Delete(
		r.Context(),
		userID,
		id,
	)
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceNotFound) {
			http.Error(
				w,
				"workspace not found",
				http.StatusNotFound,
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

	w.WriteHeader(http.StatusNoContent)
}
