package http

import (
	"errors"
	"net/http"

	"github.com/kristopher1027/devflow-ai/internal/auth"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

type LogoutHandler struct {
	service      service.AuthService
	cookieSecure bool
}

func NewLogoutHandler(
	service service.AuthService,
	cookieSecure bool,
) *LogoutHandler {
	return &LogoutHandler{
		service:      service,
		cookieSecure: cookieSecure,
	}
}

func (h *LogoutHandler) Logout(
	w http.ResponseWriter,
	r *http.Request,
) {
	cookie, err := r.Cookie("devflow_session")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		http.Error(
			w,
			"authentication required",
			http.StatusUnauthorized,
		)
		return
	}

	tokenHash := auth.HashSessionToken(cookie.Value)

	if err := h.service.Logout(
		r.Context(),
		tokenHash,
	); err != nil {
		http.Error(
			w,
			"internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "devflow_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusNoContent)
}
