package http

import "net/http"

func NewRouter(userHandler *UserHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", HealthHandler)
	mux.HandleFunc("/users", userHandler.GetUser)

	return mux
}
