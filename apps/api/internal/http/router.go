package http

import "net/http"

func NewRouter(
	userHandler *UserHandler,
	registrationHandler *RegistrationHandler,
	loginHandler *LoginHandler,
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

	return mux
}
