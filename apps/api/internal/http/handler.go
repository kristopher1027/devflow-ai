package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	"github.com/kristopher1027/devflow-ai/internal/repository"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

func HealthHandler(w nethttp.ResponseWriter, r *nethttp.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]string{
		"service": "devflow-api",
		"status":  "ok",
	}

	json.NewEncoder(w).Encode(response)
}

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

func (h *UserHandler) GetUser(
	w nethttp.ResponseWriter,
	r *nethttp.Request,
) {
	email := r.URL.Query().Get("email")

	if email == "" {
		nethttp.Error(
			w,
			"email is required",
			nethttp.StatusBadRequest,
		)
		return
	}

	user, err := h.service.FindUserByEmail(
		r.Context(),
		email,
	)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			nethttp.Error(
				w,
				"user not found",
				nethttp.StatusNotFound,
			)
			return
		}

		nethttp.Error(
			w,
			"internal server error",
			nethttp.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(user)
}
