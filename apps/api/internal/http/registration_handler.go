package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	"github.com/kristopher1027/devflow-ai/internal/repository"
	"github.com/kristopher1027/devflow-ai/internal/service"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegistrationHandler struct {
	service service.RegistrationService
}

func NewRegistrationHandler(
	service service.RegistrationService,
) *RegistrationHandler {
	return &RegistrationHandler{
		service: service,
	}
}

func (h *RegistrationHandler) Register(
	w nethttp.ResponseWriter,
	r *nethttp.Request,
) {
	var request RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		nethttp.Error(
			w,
			"invalid request body",
			nethttp.StatusBadRequest,
		)
		return
	}

	user, err := h.service.Register(
		r.Context(),
		request.Email,
		request.Password,
	)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailRequired):
			nethttp.Error(
				w,
				"email is required",
				nethttp.StatusBadRequest,
			)

		case errors.Is(err, service.ErrPasswordRequired):
			nethttp.Error(
				w,
				"password is required",
				nethttp.StatusBadRequest,
			)

		case errors.Is(err, repository.ErrEmailAlreadyExists):
			nethttp.Error(
				w,
				"email already exists",
				nethttp.StatusConflict,
			)

		default:
			nethttp.Error(
				w,
				"internal server error",
				nethttp.StatusInternalServerError,
			)
		}

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(nethttp.StatusCreated)

	if err := json.NewEncoder(w).Encode(user); err != nil {
		return
	}
}
