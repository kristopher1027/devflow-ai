package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/kristopher1027/devflow-ai/internal/repository"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

type WorkspaceMemberHandler struct {
	service service.WorkspaceMemberService
}

func NewWorkspaceMemberHandler(
	service service.WorkspaceMemberService,
) *WorkspaceMemberHandler {
	return &WorkspaceMemberHandler{
		service: service,
	}
}

type AddWorkspaceMemberRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

func (h *WorkspaceMemberHandler) Add(
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

	workspaceID := workspaceIDFromPath(r)
	if workspaceID == "" {
		http.Error(
			w,
			"workspace ID is required",
			http.StatusBadRequest,
		)
		return
	}

	var request AddWorkspaceMemberRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	request.UserID = strings.TrimSpace(request.UserID)

	if request.UserID == "" {
		http.Error(
			w,
			"user ID is required",
			http.StatusBadRequest,
		)
		return
	}

	member, err := h.service.Add(
		r.Context(),
		requesterID,
		workspaceID,
		request.UserID,
		request.Role,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrWorkspaceMemberRoleRequired):
			http.Error(
				w,
				"workspace member role is required",
				http.StatusBadRequest,
			)
		case errors.Is(err, service.ErrInvalidWorkspaceMemberRole):
			http.Error(
				w,
				"invalid workspace member role",
				http.StatusBadRequest,
			)
		case errors.Is(
			err,
			service.ErrWorkspaceMemberAdminCannotAssignRole,
		):
			http.Error(
				w,
				"admins can only add members",
				http.StatusForbidden,
			)
		case errors.Is(
			err,
			service.ErrWorkspaceMemberUnauthorized,
		):
			http.Error(
				w,
				"unauthorized to manage workspace members",
				http.StatusForbidden,
			)
		case errors.Is(err, repository.ErrWorkspaceNotFound):
			http.Error(
				w,
				"workspace not found",
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

	writeJSON(
		w,
		http.StatusCreated,
		member,
	)
}

func (h *WorkspaceMemberHandler) List(
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

	workspaceID := workspaceIDFromPath(r)
	if workspaceID == "" {
		http.Error(
			w,
			"workspace ID is required",
			http.StatusBadRequest,
		)
		return
	}

	members, err := h.service.ListByWorkspaceID(
		r.Context(),
		requesterID,
		workspaceID,
	)
	if err != nil {
		switch {
		case errors.Is(
			err,
			service.ErrWorkspaceMemberUnauthorized,
		):
			http.Error(
				w,
				"unauthorized to manage workspace members",
				http.StatusForbidden,
			)
		case errors.Is(err, repository.ErrWorkspaceNotFound):
			http.Error(
				w,
				"workspace not found",
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

	writeJSON(
		w,
		http.StatusOK,
		members,
	)
}

func (h *WorkspaceMemberHandler) Get(
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

	workspaceID, userID := workspaceMemberPathIDs(r)

	if workspaceID == "" || userID == "" {
		http.Error(
			w,
			"workspace ID and user ID are required",
			http.StatusBadRequest,
		)
		return
	}

	member, err := h.service.Find(
		r.Context(),
		requesterID,
		workspaceID,
		userID,
	)
	if err != nil {
		switch {
		case errors.Is(
			err,
			service.ErrWorkspaceMemberUnauthorized,
		):
			http.Error(
				w,
				"unauthorized to manage workspace members",
				http.StatusForbidden,
			)
		case errors.Is(
			err,
			repository.ErrWorkspaceNotFound,
		):
			http.Error(
				w,
				"workspace not found",
				http.StatusNotFound,
			)
		case errors.Is(
			err,
			repository.ErrWorkspaceMemberNotFound,
		):
			http.Error(
				w,
				"workspace member not found",
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

	writeJSON(
		w,
		http.StatusOK,
		member,
	)
}

func (h *WorkspaceMemberHandler) Remove(
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

	workspaceID, userID := workspaceMemberPathIDs(r)

	if workspaceID == "" || userID == "" {
		http.Error(
			w,
			"workspace ID and user ID are required",
			http.StatusBadRequest,
		)
		return
	}

	err := h.service.Remove(
		r.Context(),
		requesterID,
		workspaceID,
		userID,
	)
	if err != nil {
		switch {
		case errors.Is(
			err,
			service.ErrWorkspaceMemberUnauthorized,
		):
			http.Error(
				w,
				"unauthorized to manage workspace members",
				http.StatusForbidden,
			)
		case errors.Is(
			err,
			service.ErrWorkspaceOwnerCannotBeRemoved,
		):
			http.Error(
				w,
				"workspace owner cannot be removed",
				http.StatusForbidden,
			)
		case errors.Is(
			err,
			repository.ErrWorkspaceNotFound,
		):
			http.Error(
				w,
				"workspace not found",
				http.StatusNotFound,
			)
		case errors.Is(
			err,
			repository.ErrWorkspaceMemberNotFound,
		):
			http.Error(
				w,
				"workspace member not found",
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

	w.WriteHeader(http.StatusNoContent)
}

func workspaceIDFromPath(r *http.Request) string {
	path := strings.TrimPrefix(
		r.URL.Path,
		"/workspaces/",
	)

	path = strings.TrimSuffix(path, "/members")

	return strings.TrimSpace(path)
}

func workspaceMemberPathIDs(
	r *http.Request,
) (string, string) {
	path := strings.TrimPrefix(
		r.URL.Path,
		"/workspaces/",
	)

	path = strings.TrimSuffix(path, "/")

	parts := strings.Split(
		path,
		"/members/",
	)

	if len(parts) != 2 {
		return "", ""
	}

	workspaceID := strings.TrimSpace(parts[0])
	userID := strings.TrimSpace(parts[1])

	return workspaceID, userID
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	value any,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}
func (h *WorkspaceMemberHandler) UpdateRole(
	w http.ResponseWriter,
	r *http.Request,
) {

	ctx := r.Context()

	requesterID, ok := UserIDFromContext(ctx)
	if !ok {
		http.Error(
			w,
			"unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	workspaceID, userID := workspaceMemberPathIDs(r)

	if workspaceID == "" || userID == "" {
		http.Error(
			w,
			"workspace ID and user ID are required",
			http.StatusBadRequest,
		)
		return
	}

	var request struct {
		Role string `json:"role"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)

	if err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	err = h.service.UpdateRole(
		ctx,
		requesterID,
		workspaceID,
		userID,
		request.Role,
	)

	if err != nil {

		switch {

		case errors.Is(
			err,
			service.ErrWorkspaceMemberUnauthorized,
		):
			http.Error(
				w,
				err.Error(),
				http.StatusForbidden,
			)

		case errors.Is(
			err,
			service.ErrInvalidWorkspaceMemberRole,
		),
			errors.Is(
				err,
				service.ErrWorkspaceMemberRoleRequired,
			):
			http.Error(
				w,
				err.Error(),
				http.StatusBadRequest,
			)

		case errors.Is(
			err,
			repository.ErrWorkspaceNotFound,
		),
			errors.Is(
				err,
				repository.ErrWorkspaceMemberNotFound,
			):
			http.Error(
				w,
				err.Error(),
				http.StatusNotFound,
			)

		case errors.Is(
			err,
			service.ErrWorkspaceMemberAdminCannotChangeRole,
		),
			errors.Is(
				err,
				service.ErrWorkspaceOwnerCannotBeRemoved,
			):
			http.Error(
				w,
				err.Error(),
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
