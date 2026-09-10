package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/kristopher1027/devflow-ai/internal/service"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginHandler struct {
	service      service.LoginService
	cookieSecure bool
}

func NewLoginHandler(
	service service.LoginService,
	cookieSecure bool,
) *LoginHandler {
	return &LoginHandler{
		service:      service,
		cookieSecure: cookieSecure,
	}
}

func (h *LoginHandler) Login(
	w http.ResponseWriter,
	r *http.Request,
) {
	var request LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	token, err := h.service.Login(
		r.Context(),
		request.Email,
		request.Password,
	)

	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			http.Error(
				w,
				"invalid credentials",
				http.StatusUnauthorized,
			)
			return
		}

		http.Error(
			w,
			"internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "devflow_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusOK)
}
