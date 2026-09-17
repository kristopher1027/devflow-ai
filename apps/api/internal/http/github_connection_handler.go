package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

type GitHubConnectionHandler struct {
	service service.GitHubConnectionService
}

func NewGitHubConnectionHandler(
	service service.GitHubConnectionService,
) *GitHubConnectionHandler {
	return &GitHubConnectionHandler{service: service}
}

type CreateGitHubConnectionRequest struct {
	InstallationID string `json:"installation_id"`
	AccountLogin   string `json:"account_login"`
}

type UpdateGitHubConnectionStatusRequest struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func (h *GitHubConnectionHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	requesterID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	workspaceID := githubWorkspaceIDFromPath(r)
	if workspaceID == "" {
		http.Error(w, "workspace ID is required", http.StatusBadRequest)
		return
	}

	var request CreateGitHubConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	connection, err := h.service.Create(
		r.Context(),
		requesterID,
		workspaceID,
		request.InstallationID,
		request.AccountLogin,
	)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, connection)
}

func (h *GitHubConnectionHandler) GetByWorkspaceID(
	w http.ResponseWriter,
	r *http.Request,
) {
	requesterID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	workspaceID := githubWorkspaceIDFromPath(r)
	if workspaceID == "" {
		http.Error(w, "workspace ID is required", http.StatusBadRequest)
		return
	}

	connection, err := h.service.FindByWorkspaceID(
		r.Context(),
		requesterID,
		workspaceID,
	)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, connection)
}

func (h *GitHubConnectionHandler) UpdateStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	requesterID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	if githubWorkspaceIDFromPath(r) == "" {
		http.Error(w, "workspace ID is required", http.StatusBadRequest)
		return
	}

	var request UpdateGitHubConnectionStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	request.ID = strings.TrimSpace(request.ID)
	request.Status = strings.TrimSpace(request.Status)
	if request.ID == "" {
		http.Error(w, "connection ID is required", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateStatus(
		r.Context(),
		requesterID,
		request.ID,
		request.Status,
	); err != nil {
		h.writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func githubWorkspaceIDFromPath(r *http.Request) string {
	path := strings.TrimPrefix(r.URL.Path, "/workspaces/")
	path = strings.TrimSuffix(path, "/github")
	return strings.TrimSpace(path)
}

func (h *GitHubConnectionHandler) writeServiceError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, service.ErrGitHubConnectionUnauthorized):
		http.Error(w, "github connection access denied", http.StatusForbidden)
	case errors.Is(err, service.ErrGitHubConnectionAlreadyExists):
		http.Error(w, "github connection already exists", http.StatusConflict)
	case errors.Is(err, repository.ErrGitHubConnectionNotFound),
		errors.Is(err, repository.ErrWorkspaceNotFound):
		http.Error(w, "github connection not found", http.StatusNotFound)
	case errors.Is(err, domain.ErrGitHubConnectionWorkspaceRequired),
		errors.Is(err, domain.ErrGitHubConnectionInstallationRequired),
		errors.Is(err, domain.ErrGitHubConnectionAccountRequired),
		errors.Is(err, domain.ErrGitHubConnectionStatusInvalid),
		errors.Is(err, service.ErrGitHubConnectionStatusTransitionInvalid):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
