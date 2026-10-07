package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/kristopher1027/devflow-ai/internal/integration/anthropic"
	"github.com/kristopher1027/devflow-ai/internal/repository"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

type RepositoryExplainHandler struct {
	service service.RepositoryExplainService
}

func NewRepositoryExplainHandler(
	explainService service.RepositoryExplainService,
) *RepositoryExplainHandler {
	return &RepositoryExplainHandler{service: explainService}
}

type repositoryExplainResponse struct {
	SnapshotID  string `json:"snapshot_id"`
	CommitSHA   string `json:"commit_sha"`
	Explanation string `json:"explanation"`
}

// Explain handles POST /repositories/{id}/explain.
func (h *RepositoryExplainHandler) Explain(
	w http.ResponseWriter,
	r *http.Request,
) {
	repositoryID := repositoryIDFromExplainPath(r.URL.Path)
	if repositoryID == "" {
		http.Error(w, "repository ID is required", http.StatusBadRequest)
		return
	}

	requesterID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	explanation, err := h.service.Explain(
		r.Context(),
		requesterID,
		repositoryID,
	)
	if err != nil {
		writeExplainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(repositoryExplainResponse{
		SnapshotID:  explanation.SnapshotID,
		CommitSHA:   explanation.CommitSHA,
		Explanation: explanation.Text,
	})
}

func writeExplainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrRepositoryExplainUnauthorized):
		http.Error(w, "forbidden", http.StatusForbidden)

	case errors.Is(err, repository.ErrRepositoryNotFound):
		http.Error(w, "repository not found", http.StatusNotFound)

	case errors.Is(err, repository.ErrProjectNotFound):
		http.Error(w, "project not found", http.StatusNotFound)

	case errors.Is(err, service.ErrRepositoryExplainNoSnapshot):
		http.Error(w, "repository has not been synced yet", http.StatusConflict)

	case errors.Is(err, service.ErrRepositoryExplainNoFiles):
		http.Error(w, "repository has no ingested files yet", http.StatusConflict)

	case errors.Is(err, anthropic.ErrAPIKeyRequired):
		http.Error(
			w,
			"AI explanations are not configured",
			http.StatusServiceUnavailable,
		)

	case errors.Is(err, anthropic.ErrUnexpectedStatus),
		errors.Is(err, anthropic.ErrEmptyResponse):
		http.Error(w, "AI provider error", http.StatusBadGateway)

	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

// authRequirer is satisfied by *AuthMiddleware.
type authRequirer interface {
	RequireAuth(next http.Handler) http.Handler
}

// WithRepositoryExplain adds POST /repositories/{id}/explain (authenticated)
// in front of an existing router. Every other request is passed through
// unchanged.
func WithRepositoryExplain(
	next http.Handler,
	auth authRequirer,
	handler *RepositoryExplainHandler,
) http.Handler {
	protected := auth.RequireAuth(http.HandlerFunc(handler.Explain))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost &&
			repositoryIDFromExplainPath(r.URL.Path) != "" {
			protected.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func repositoryIDFromExplainPath(path string) string {
	const prefix = "/repositories/"
	const suffix = "/explain"

	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
		return ""
	}

	id := strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix)
	if id == "" || strings.Contains(id, "/") {
		return ""
	}

	return id
}
