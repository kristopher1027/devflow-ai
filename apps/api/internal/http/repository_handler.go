package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/kristopher1027/devflow-ai/internal/repository"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

type RepositoryHandler struct {
	service service.RepositoryService
}

func NewRepositoryHandler(
	service service.RepositoryService,
) *RepositoryHandler {
	return &RepositoryHandler{
		service: service,
	}
}

type CreateRepositoryRequest struct {
	Provider      string `json:"provider"`
	ExternalID    string `json:"external_id"`
	Owner         string `json:"owner"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	DefaultBranch string `json:"default_branch"`
	HTMLURL       string `json:"html_url"`
	CloneURL      string `json:"clone_url"`
	IsPrivate     bool   `json:"is_private"`
}

func (h *RepositoryHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	requesterID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	projectID := strings.TrimPrefix(r.URL.Path, "/projects/")
	projectID = strings.TrimSuffix(projectID, "/repositories")
	if projectID == "" {
		http.Error(w, "project ID is required", http.StatusBadRequest)
		return
	}

	var request CreateRepositoryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	newRepository, err := h.service.Create(
		r.Context(),
		requesterID,
		projectID,
		request.Provider,
		request.ExternalID,
		request.Owner,
		request.Name,
		request.FullName,
		request.DefaultBranch,
		request.HTMLURL,
		request.CloneURL,
		request.IsPrivate,
	)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(newRepository)
}

func (h *RepositoryHandler) ListByProjectID(
	w http.ResponseWriter,
	r *http.Request,
) {
	requesterID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	projectID := strings.TrimPrefix(r.URL.Path, "/projects/")
	projectID = strings.TrimSuffix(projectID, "/repositories")
	if projectID == "" {
		http.Error(w, "project ID is required", http.StatusBadRequest)
		return
	}

	repositories, err := h.service.ListByProjectID(
		r.Context(),
		requesterID,
		projectID,
	)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(repositories)
}

func (h *RepositoryHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	requesterID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/repositories/")
	if id == "" {
		http.Error(w, "repository ID is required", http.StatusBadRequest)
		return
	}

	repositoryModel, err := h.service.FindByID(
		r.Context(),
		requesterID,
		id,
	)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(repositoryModel)
}

func (h *RepositoryHandler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	requesterID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/repositories/")
	if id == "" {
		http.Error(w, "repository ID is required", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), requesterID, id); err != nil {
		h.writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RepositoryHandler) writeServiceError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, service.ErrRepositoryUnauthorized):
		http.Error(w, "repository access denied", http.StatusForbidden)
	case errors.Is(err, repository.ErrRepositoryNotFound):
		http.Error(w, "repository not found", http.StatusNotFound)
	case errors.Is(err, service.ErrRepositoryAlreadyExists):
		http.Error(w, "repository already exists", http.StatusConflict)
	case errors.Is(err, service.ErrRepositoryProviderUnsupported),
		errors.Is(err, service.ErrRepositoryExternalIDRequired),
		errors.Is(err, service.ErrRepositoryOwnerRequired),
		errors.Is(err, service.ErrRepositoryNameRequired),
		errors.Is(err, service.ErrRepositoryFullNameRequired),
		errors.Is(err, service.ErrRepositoryDefaultBranchRequired),
		errors.Is(err, service.ErrRepositoryHTMLURLRequired),
		errors.Is(err, service.ErrRepositoryCloneURLRequired):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
