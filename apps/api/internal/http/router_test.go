package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

type routerFakeUserService struct{}

func (f *routerFakeUserService) FindUserByEmail(
	ctx context.Context,
	email string,
) (*domain.User, error) {
	return nil, repository.ErrUserNotFound
}

type routerFakeRegistrationService struct{}

func (f *routerFakeRegistrationService) Register(
	ctx context.Context,
	email string,
	password string,
) (*domain.User, error) {
	return &domain.User{
		ID:    "user-123",
		Email: email,
	}, nil
}

type routerFakeLoginService struct{}

func (f *routerFakeLoginService) Login(
	ctx context.Context,
	email string,
	password string,
) (string, error) {
	return "test-token", nil
}

type routerFakeWorkspaceService struct{}

func (f *routerFakeWorkspaceService) Create(
	ctx context.Context,
	ownerID string,
	name string,
) (*domain.Workspace, error) {
	return &domain.Workspace{
		ID:      "workspace-123",
		OwnerID: ownerID,
		Name:    name,
	}, nil
}

func (f *routerFakeWorkspaceService) FindByID(
	ctx context.Context,
	ownerID string,
	id string,
) (*domain.Workspace, error) {
	return &domain.Workspace{
		ID:      id,
		OwnerID: ownerID,
	}, nil
}

func (f *routerFakeWorkspaceService) ListByOwnerID(
	ctx context.Context,
	ownerID string,
) ([]*domain.Workspace, error) {
	return []*domain.Workspace{}, nil
}

func (f *routerFakeWorkspaceService) Delete(
	ctx context.Context,
	ownerID string,
	id string,
) error {
	return nil
}

type routerFakeWorkspaceMemberService struct{}

func (f *routerFakeWorkspaceMemberService) Add(
	ctx context.Context,
	requesterID string,
	workspaceID string,
	userID string,
	role string,
) (*domain.WorkspaceMember, error) {
	return &domain.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        role,
	}, nil
}

func (f *routerFakeWorkspaceMemberService) Find(
	ctx context.Context,
	requesterID string,
	workspaceID string,
	userID string,
) (*domain.WorkspaceMember, error) {
	return &domain.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        service.WorkspaceMemberRoleMember,
	}, nil
}

func (f *routerFakeWorkspaceMemberService) ListByWorkspaceID(
	ctx context.Context,
	requesterID string,
	workspaceID string,
) ([]*domain.WorkspaceMember, error) {
	return []*domain.WorkspaceMember{}, nil
}

func (f *routerFakeWorkspaceMemberService) UpdateRole(
	ctx context.Context,
	requesterID string,
	workspaceID string,
	userID string,
	role string,
) error {
	return nil
}

func (f *routerFakeWorkspaceMemberService) Remove(
	ctx context.Context,
	requesterID string,
	workspaceID string,
	userID string,
) error {
	return nil
}

type routerFakeAuthService struct{}

func (f *routerFakeAuthService) AuthenticateSession(
	ctx context.Context,
	tokenHash string,
) (*domain.Session, error) {
	return &domain.Session{
		UserID: "user-123",
	}, nil
}

func (f *routerFakeAuthService) Logout(
	ctx context.Context,
	tokenHash string,
) error {
	return nil
}

func newTestRouter() http.Handler {
	userHandler := NewUserHandler(
		&routerFakeUserService{},
	)

	registrationHandler := NewRegistrationHandler(
		&routerFakeRegistrationService{},
	)

	loginHandler := NewLoginHandler(
		&routerFakeLoginService{},
		false,
	)

	workspaceHandler := NewWorkspaceHandler(
		&routerFakeWorkspaceService{},
	)

	workspaceMemberHandler := NewWorkspaceMemberHandler(
		&routerFakeWorkspaceMemberService{},
	)

	projectHandler := NewProjectHandler(
		&fakeProjectService{},
	)

	repositoryHandler := NewRepositoryHandler(
		&fakeRepositoryService{},
	)

	githubRepositoryImportHandler := NewGitHubRepositoryImportHandler(
		&fakeGitHubRepositoryImportQueue{},
	)

	githubRepositoryImportJobHandler :=
		NewGitHubRepositoryImportJobHandler(
			&fakeGitHubRepositoryImportJobService{},
		)

	githubConnectionHandler := NewGitHubConnectionHandler(
		&fakeGitHubConnectionService{},
	)
	repositorySyncJobHandler := NewRepositorySyncJobHandler(
	&fakeRepositorySyncJobService{},
)

	authMiddleware := NewAuthMiddleware(
		&routerFakeAuthService{},
	)

	return NewRouter(
		userHandler,
		registrationHandler,
		loginHandler,
		workspaceHandler,
		workspaceMemberHandler,
		projectHandler,
		repositoryHandler,
		githubRepositoryImportHandler,
		githubRepositoryImportJobHandler,
		githubConnectionHandler,
		repositorySyncJobHandler,
		authMiddleware,
	)
}

func TestRouterHealth(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestRouterPublicAuthRoutes(t *testing.T) {
	router := newTestRouter()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "register",
			method: http.MethodPost,
			path:   "/auth/register",
		},
		{
			name:   "login",
			method: http.MethodPost,
			path:   "/auth/login",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(
				tt.method,
				tt.path,
				nil,
			)

			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code == http.StatusUnauthorized {
				t.Fatalf(
					"expected public route, got unauthorized",
				)
			}
		})
	}
}

func TestRouterProtectedRoutesRequireAuth(t *testing.T) {
	router := newTestRouter()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "users",
			method: http.MethodGet,
			path:   "/users?email=test@example.com",
		},
		{
			name:   "create workspace",
			method: http.MethodPost,
			path:   "/workspaces",
		},
		{
			name:   "list workspaces",
			method: http.MethodGet,
			path:   "/workspaces",
		},
		{
			name:   "get workspace",
			method: http.MethodGet,
			path:   "/workspaces/workspace-123",
		},
		{
			name:   "delete workspace",
			method: http.MethodDelete,
			path:   "/workspaces/workspace-123",
		},
		{
			name:   "add member",
			method: http.MethodPost,
			path:   "/workspaces/workspace-123/members",
		},
		{
			name:   "list members",
			method: http.MethodGet,
			path:   "/workspaces/workspace-123/members",
		},
		{
			name:   "get member",
			method: http.MethodGet,
			path:   "/workspaces/workspace-123/members/user-456",
		},
		{
			name:   "update member role",
			method: http.MethodPatch,
			path:   "/workspaces/workspace-123/members/user-456",
		},
		{
			name:   "remove member",
			method: http.MethodDelete,
			path:   "/workspaces/workspace-123/members/user-456",
		},
		{
			name:   "create repository",
			method: http.MethodPost,
			path:   "/projects/project-123/repositories",
		},
		{
			name:   "list repositories",
			method: http.MethodGet,
			path:   "/projects/project-123/repositories",
		},
		{
			name:   "get repository",
			method: http.MethodGet,
			path:   "/repositories/repository-123",
		},
		{
			name:   "delete repository",
			method: http.MethodDelete,
			path:   "/repositories/repository-123",
		},
		{
			name:   "import repositories",
			method: http.MethodPost,
			path:   "/projects/project-123/repositories/import",
		},
		{
			name:   "create github connection",
			method: http.MethodPost,
			path:   "/workspaces/workspace-123/github",
		},
		{
			name:   "get github connection",
			method: http.MethodGet,
			path:   "/workspaces/workspace-123/github",
		},
		{
			name:   "update github connection status",
			method: http.MethodPatch,
			path:   "/workspaces/workspace-123/github",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(
				tt.method,
				tt.path,
				nil,
			)

			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf(
					"expected status %d, got %d",
					http.StatusUnauthorized,
					rec.Code,
				)
			}
		})
	}
}

func TestRouterProtectedRoutesWithAuth(t *testing.T) {
	router := newTestRouter()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "list workspaces",
			method: http.MethodGet,
			path:   "/workspaces",
		},
		{
			name:   "get workspace",
			method: http.MethodGet,
			path:   "/workspaces/workspace-123",
		},
		{
			name:   "list members",
			method: http.MethodGet,
			path:   "/workspaces/workspace-123/members",
		},
		{
			name:   "get member",
			method: http.MethodGet,
			path:   "/workspaces/workspace-123/members/user-456",
		},
		{
			name:   "update member role",
			method: http.MethodPatch,
			path:   "/workspaces/workspace-123/members/user-456",
		},
		{
			name:   "remove member",
			method: http.MethodDelete,
			path:   "/workspaces/workspace-123/members/user-456",
		},
		{
			name:   "create repository",
			method: http.MethodPost,
			path:   "/projects/project-123/repositories",
		},
		{
			name:   "list repositories",
			method: http.MethodGet,
			path:   "/projects/project-123/repositories",
		},
		{
			name:   "get repository",
			method: http.MethodGet,
			path:   "/repositories/repository-123",
		},
		{
			name:   "delete repository",
			method: http.MethodDelete,
			path:   "/repositories/repository-123",
		},
		{
			name:   "import repositories",
			method: http.MethodPost,
			path:   "/projects/project-123/repositories/import",
		},
		{
			name:   "create github connection",
			method: http.MethodPost,
			path:   "/workspaces/workspace-123/github",
		},
		{
			name:   "get github connection",
			method: http.MethodGet,
			path:   "/workspaces/workspace-123/github",
		},
		{
			name:   "update github connection status",
			method: http.MethodPatch,
			path:   "/workspaces/workspace-123/github",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(
				tt.method,
				tt.path,
				nil,
			)

			req.AddCookie(&http.Cookie{
				Name:  "devflow_session",
				Value: "test-session-token",
			})

			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code == http.StatusUnauthorized {
				t.Fatalf(
					"expected authenticated request to pass middleware",
				)
			}
		})
	}
}

func TestRouterMethodNotAllowed(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(
		http.MethodPut,
		"/workspaces",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusMethodNotAllowed,
			rec.Code,
		)
	}
}

func TestRouterMemberPatchRoute(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(
		http.MethodPatch,
		"/workspaces/workspace-123/members/user-456",
		nil,
	)

	req.AddCookie(&http.Cookie{
		Name:  "devflow_session",
		Value: "test-session-token",
	})

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code == http.StatusMethodNotAllowed {
		t.Fatalf(
			"PATCH member route was not dispatched correctly",
		)
	}

	if rec.Code == http.StatusUnauthorized {
		t.Fatalf(
			"PATCH member route unexpectedly rejected authentication",
		)
	}
}
