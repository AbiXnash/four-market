package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/AbiXnash/4-market/internal/dto"
	"github.com/AbiXnash/4-market/internal/response"
	"github.com/AbiXnash/4-market/internal/service"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{service: svc}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest

	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "application/x-www-form-urlencoded") {
		if err := r.ParseForm(); err != nil {
			response.ValidationError(w, r, "invalid form data")
			return
		}
		req.Email = r.FormValue("email")
		req.Password = r.FormValue("password")
	} else {
		if err := decodeJSON(r, &req); err != nil {
			response.ValidationError(w, r, "invalid request body")
			return
		}
	}

	slog.Debug("Login Request: ", "request", req)

	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Unauthorized(w, r)
			return
		}
		response.InternalError(w, r)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		response.ValidationError(w, r, "invalid request body")
		return
	}

	if err := h.service.Register(r.Context(), req); err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			response.ValidationError(w, r, err.Error())
			return
		}
		response.InternalError(w, r)
		return
	}

	response.JSON(w, http.StatusCreated, dto.MessageResponse{Message: "user created"})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := decodeJSON(r, &req); err != nil {
		response.ValidationError(w, r, "invalid request body")
		return
	}

	resp, err := h.service.Refresh(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			response.Unauthorized(w, r)
			return
		}
		response.InternalError(w, r)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
