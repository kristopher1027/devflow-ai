package http

import "net/http"

// WithLogout adds POST /auth/logout in front of an existing handler.
// Every other request is passed through unchanged.
func WithLogout(next http.Handler, handler *LogoutHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/auth/logout" {
			handler.Logout(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
