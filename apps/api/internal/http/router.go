package http

import "net/http"

func NewRouter(
	userHandler *UserHandler,
	registrationHandler *RegistrationHandler,
	loginHandler *LoginHandler,
	workspaceHandler *WorkspaceHandler,
	workspaceMemberHandler *WorkspaceMemberHandler,
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
		case hasSuffix(path, "/members") &&
			r.Method == http.MethodPost:
			protectedAddWorkspaceMember.ServeHTTP(w, r)

		case hasSuffix(path, "/members") &&
			r.Method == http.MethodGet:
			protectedListWorkspaceMembers.ServeHTTP(w, r)

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