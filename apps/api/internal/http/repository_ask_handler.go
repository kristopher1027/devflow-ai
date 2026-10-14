package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/kristopher1027/devflow-ai/internal/service"
)

const maxAskRequestBytes = 8 * 1024

type RepositoryAskHandler struct {
	service service.RepositoryAskService
}

func NewRepositoryAskHandler(
	askService service.RepositoryAskService,
) *RepositoryAskHandler {
	return &RepositoryAskHandler{service: askService}
}

type repositoryAskRequest struct {
	Question string `json:"question"`
}

type repositoryAskSource struct {
	FilePath  string `json:"file_path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

type repositoryAskResponse struct {
	SnapshotID string                `json:"snapshot_id"`
	CommitSHA  string                `json:"commit_sha"`
	Answer     string                `json:"answer"`
	Sources    []repositoryAskSource `json:"sources"`
}

// Ask handles POST /repositories/{id}/ask with {"question": "..."}.
func (h *RepositoryAskHandler) Ask(
	w http.ResponseWriter,
	r *http.Request,
) {
	repositoryID := repositoryIDFromActionPath(r.URL.Path, "/ask")
	if repositoryID == "" {
		http.Error(w, "repository ID is required", http.StatusBadRequest)
		return
	}

	requesterID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxAskRequestBytes)

	var request repositoryAskRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	answer, err := h.service.Ask(
		r.Context(),
		requesterID,
		repositoryID,
		request.Question,
	)
	if err != nil {
		writeAskError(w, err)
		return
	}

	sources := make([]repositoryAskSource, 0, len(answer.Sources))
	for _, source := range answer.Sources {
		if source == nil {
			continue
		}
		sources = append(sources, repositoryAskSource{
			FilePath:  source.FilePath,
			StartLine: source.StartLine,
			EndLine:   source.EndLine,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(repositoryAskResponse{
		SnapshotID: answer.SnapshotID,
		CommitSHA:  answer.CommitSHA,
		Answer:     answer.Answer,
		Sources:    sources,
	})
}

func writeAskError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrRepositoryAskQuestionRequired):
		http.Error(w, "question is required", http.StatusBadRequest)

	case errors.Is(err, service.ErrRepositoryAskQuestionTooLong):
		http.Error(w, "question is too long", http.StatusBadRequest)

	case errors.Is(err, service.ErrRepositoryAskUnauthorized):
		http.Error(w, "forbidden", http.StatusForbidden)

	case errors.Is(err, service.ErrRepositoryAskNoSnapshot):
		http.Error(w, "repository has not been synced yet", http.StatusConflict)

	case errors.Is(err, service.ErrRepositoryAskNoMatches):
		http.Error(
			w,
			"no code matched that question; try different keywords",
			http.StatusNotFound,
		)

	default:
		// Repository/project not found, AI provider errors and unexpected
		// failures share the explain endpoint's mapping.
		writeExplainError(w, err)
	}
}

// WithRepositoryAsk adds POST /repositories/{id}/ask (authenticated) in
// front of an existing handler. Every other request passes through.
func WithRepositoryAsk(
	next http.Handler,
	auth authRequirer,
	handler *RepositoryAskHandler,
) http.Handler {
	protected := auth.RequireAuth(http.HandlerFunc(handler.Ask))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost &&
			repositoryIDFromActionPath(r.URL.Path, "/ask") != "" {
			protected.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func repositoryIDFromActionPath(path string, action string) string {
	const prefix = "/repositories/"

	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, action) {
		return ""
	}

	id := strings.TrimSuffix(strings.TrimPrefix(path, prefix), action)
	if id == "" || strings.Contains(id, "/") {
		return ""
	}

	return id
}
