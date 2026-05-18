package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/AbiXnash/4-market/internal/response"
	"github.com/AbiXnash/4-market/internal/security"
	"github.com/AbiXnash/4-market/internal/validator"
)

type authDeps struct {
	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

var Auth *authDeps

func InitAuth(secret string, accessTTL, refreshTTL int) {
	Auth = &authDeps{
		jwtSecret:  []byte(secret),
		accessTTL:  time.Duration(accessTTL) * time.Minute,
		refreshTTL: time.Duration(refreshTTL) * time.Minute,
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		response.ValidationError(w, "invalid request body")
		return
	}

	if err := validator.Struct(req); err != nil {
		response.ValidationError(w, err.Error())
		return
	}

	if Auth == nil {
		response.InternalError(w)
		return
	}

	// TODO: verify credentials against DB
	accessToken, err := security.SignJWT("user-id", "user", Auth.jwtSecret, Auth.accessTTL)
	if err != nil {
		response.InternalError(w)
		return
	}

	refreshToken, err := security.SignJWT("user-id", "refresh", Auth.jwtSecret, Auth.refreshTTL)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

type registerRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		response.ValidationError(w, "invalid request body")
		return
	}

	if err := validator.Struct(req); err != nil {
		response.ValidationError(w, err.Error())
		return
	}

	if _, err := security.HashPassword(req.Password); err != nil {
		response.InternalError(w)
		return
	}

	// TODO: store user in DB
	response.JSON(w, http.StatusCreated, map[string]string{"message": "user created"})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := decodeJSON(r, &req); err != nil {
		response.ValidationError(w, "invalid request body")
		return
	}

	if Auth == nil {
		response.InternalError(w)
		return
	}

	claims, err := security.ValidateJWT(req.RefreshToken, Auth.jwtSecret)
	if err != nil || claims.Role != "refresh" {
		response.Unauthorized(w)
		return
	}

	accessToken, err := security.SignJWT(claims.UserID, "user", Auth.jwtSecret, Auth.accessTTL)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"access_token": accessToken,
	})
}

func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
