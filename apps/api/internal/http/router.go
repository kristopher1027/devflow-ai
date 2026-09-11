package http

import "net/http"

func NewRouter(
	userHandler *UserHandler,
	registrationHandler *RegistrationHandler,
	loginHandler *LoginHandler,
	workspaceHandler *WorkspaceHandler,
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
		switch r.Method {
		case http.MethodGet:
			protectedGetWorkspace.ServeHTTP(w, r)

		case http.MethodDelete:
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