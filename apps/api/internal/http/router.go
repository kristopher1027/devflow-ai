package http

import "net/http"

func NewRouter(
	userHandler *UserHandler,
	registrationHandler *RegistrationHandler,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", HealthHandler)
	mux.HandleFunc("/users", userHandler.GetUser)
	mux.HandleFunc("/auth/register", registrationHandler.Register)

	return mux
}
