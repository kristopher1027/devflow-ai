package http

import "net/http"

func NewRouter(
	userHandler *UserHandler,
	registrationHandler *RegistrationHandler,
	loginHandler *LoginHandler,
	workspaceHandler *WorkspaceHandler,
	workspaceMemberHandler *WorkspaceMemberHandler,
	projectHandler *ProjectHandler,
	repositoryHandler *RepositoryHandler,
	githubRepositoryImportHandler *GitHubRepositoryImportHandler,
	githubRepositoryImportJobHandler *GitHubRepositoryImportJobHandler,
	githubConnectionHandler *GitHubConnectionHandler,
	authMiddleware *AuthMiddleware,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", HealthHandler)
	mux.HandleFunc("/auth/register", registrationHandler.Register)
	mux.HandleFunc("/auth/login", loginHandler.Login)

	protectedUsers := authMiddleware.RequireAuth(
		http.HandlerFunc(userHandler.GetUser),
	)

	mux.Handle("/users", protectedUsers)

	protectedCreateWorkspace := authMiddleware.RequireAuth(
		http.HandlerFunc(workspaceHandler.Create),
	)

	protectedListWorkspaces := authMiddleware.RequireAuth(
		http.HandlerFunc(workspaceHandler.List),
	)

	protectedGetWorkspace := authMiddleware.RequireAuth(
		http.HandlerFunc(workspaceHandler.GetByID),
	)
	protectedGetImportJob := authMiddleware.RequireAuth(
		http.HandlerFunc(githubRepositoryImportJobHandler.GetByID),
	)

	protectedDeleteWorkspace := authMiddleware.RequireAuth(
		http.HandlerFunc(workspaceHandler.Delete),
	)

	protectedAddWorkspaceMember := authMiddleware.RequireAuth(
		http.HandlerFunc(workspaceMemberHandler.Add),
	)

	protectedListWorkspaceMembers := authMiddleware.RequireAuth(
		http.HandlerFunc(workspaceMemberHandler.List),
	)

	protectedGetWorkspaceMember := authMiddleware.RequireAuth(
		http.HandlerFunc(workspaceMemberHandler.Get),
	)

	protectedRemoveWorkspaceMember := authMiddleware.RequireAuth(
		http.HandlerFunc(workspaceMemberHandler.Remove),
	)
	protectedUpdateWorkspaceMemberRole := authMiddleware.RequireAuth(
		http.HandlerFunc(workspaceMemberHandler.UpdateRole),
	)
	protectedCreateProject := authMiddleware.RequireAuth(
		http.HandlerFunc(projectHandler.Create),
	)

	protectedListProjects := authMiddleware.RequireAuth(
		http.HandlerFunc(projectHandler.ListByWorkspaceID),
	)

	protectedGetProject := authMiddleware.RequireAuth(
		http.HandlerFunc(projectHandler.GetByID),
	)

	protectedDeleteProject := authMiddleware.RequireAuth(
		http.HandlerFunc(projectHandler.Delete),
	)

	protectedCreateRepository := authMiddleware.RequireAuth(
		http.HandlerFunc(repositoryHandler.Create),
	)

	protectedListRepositories := authMiddleware.RequireAuth(
		http.HandlerFunc(repositoryHandler.ListByProjectID),
	)

	protectedGetRepository := authMiddleware.RequireAuth(
		http.HandlerFunc(repositoryHandler.GetByID),
	)

	protectedDeleteRepository := authMiddleware.RequireAuth(
		http.HandlerFunc(repositoryHandler.Delete),
	)

	protectedImportRepositories := authMiddleware.RequireAuth(
		http.HandlerFunc(githubRepositoryImportHandler.Import),
	)

	protectedCreateGitHubConnection := authMiddleware.RequireAuth(
		http.HandlerFunc(githubConnectionHandler.Create),
	)

	protectedGetGitHubConnection := authMiddleware.RequireAuth(
		http.HandlerFunc(githubConnectionHandler.GetByWorkspaceID),
	)

	protectedUpdateGitHubConnectionStatus := authMiddleware.RequireAuth(
		http.HandlerFunc(githubConnectionHandler.UpdateStatus),
	)

	mux.Handle("/workspaces", http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		switch r.Method {
		case http.MethodPost:
			protectedCreateWorkspace.ServeHTTP(w, r)

		case http.MethodGet:
			protectedListWorkspaces.ServeHTTP(w, r)

		default:
			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
		}
	}))

	mux.Handle("/workspaces/", http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		path := r.URL.Path

		switch {
		case containsMemberPath(path) &&
			r.Method == http.MethodPatch:
			protectedUpdateWorkspaceMemberRole.ServeHTTP(w, r)
		case hasSuffix(path, "/members") &&
			r.Method == http.MethodPost:
			protectedAddWorkspaceMember.ServeHTTP(w, r)

		case hasSuffix(path, "/members") &&
			r.Method == http.MethodGet:
			protectedListWorkspaceMembers.ServeHTTP(w, r)

		case hasSuffix(path, "/projects") &&
			r.Method == http.MethodPost:
			protectedCreateProject.ServeHTTP(w, r)

		case hasSuffix(path, "/projects") &&
			r.Method == http.MethodGet:
			protectedListProjects.ServeHTTP(w, r)

		case hasSuffix(path, "/github") &&
			r.Method == http.MethodPost:
			protectedCreateGitHubConnection.ServeHTTP(w, r)

		case hasSuffix(path, "/github") &&
			r.Method == http.MethodGet:
			protectedGetGitHubConnection.ServeHTTP(w, r)

		case hasSuffix(path, "/github") &&
			r.Method == http.MethodPatch:
			protectedUpdateGitHubConnectionStatus.ServeHTTP(w, r)

		case containsMemberPath(path) &&
			r.Method == http.MethodGet:
			protectedGetWorkspaceMember.ServeHTTP(w, r)

		case containsMemberPath(path) &&
			r.Method == http.MethodDelete:
			protectedRemoveWorkspaceMember.ServeHTTP(w, r)

		case r.Method == http.MethodGet:
			protectedGetWorkspace.ServeHTTP(w, r)

		case r.Method == http.MethodDelete:
			protectedDeleteWorkspace.ServeHTTP(w, r)

		default:
			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
		}
	}))
	mux.Handle("/repositories/", http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		switch r.Method {
		case http.MethodGet:
			protectedGetRepository.ServeHTTP(w, r)

		case http.MethodDelete:
			protectedDeleteRepository.ServeHTTP(w, r)

		default:
			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
		}
	}))
	mux.Handle("/github/repository-import-jobs/", http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		if r.Method != http.MethodGet {
			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
			return
		}

		protectedGetImportJob.ServeHTTP(w, r)
	}))
	mux.Handle("/projects/", http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		path := r.URL.Path

		switch {
		case hasSuffix(path, "/repositories/import") &&
			r.Method == http.MethodPost:
			protectedImportRepositories.ServeHTTP(w, r)

		case hasSuffix(path, "/repositories") &&
			r.Method == http.MethodPost:
			protectedCreateRepository.ServeHTTP(w, r)

		case hasSuffix(path, "/repositories") &&
			r.Method == http.MethodGet:
			protectedListRepositories.ServeHTTP(w, r)

		case r.Method == http.MethodGet:
			protectedGetProject.ServeHTTP(w, r)

		case r.Method == http.MethodDelete:
			protectedDeleteProject.ServeHTTP(w, r)

		default:
			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
		}
	}))

	return mux
}

func hasSuffix(path string, suffix string) bool {
	return len(path) >= len(suffix) &&
		path[len(path)-len(suffix):] == suffix
}

func containsMemberPath(path string) bool {
	const prefix = "/workspaces/"

	if len(path) <= len(prefix) {
		return false
	}

	remaining := path[len(prefix):]

	const marker = "/members/"

	for i := 0; i+len(marker) <= len(remaining); i++ {
		if remaining[i:i+len(marker)] == marker {
			return i > 0 &&
				i+len(marker) < len(remaining)
		}
	}

	return false
}
