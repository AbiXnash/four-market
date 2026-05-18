package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/AbiXnash/4-market/internal/dto"
	redisStore "github.com/AbiXnash/4-market/internal/redis"
	"github.com/AbiXnash/4-market/internal/response"
	"github.com/AbiXnash/4-market/internal/security"
	"github.com/AbiXnash/4-market/internal/service"
	"github.com/AbiXnash/4-market/web"
	"github.com/a-h/templ"
)

type AuthHandler struct {
	service    *service.AuthService
	redis      *redisStore.Store
	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	secure     bool
}

func NewAuthHandler(svc *service.AuthService, rstore *redisStore.Store, jwtSecret string, accessTTL, refreshTTL int, secure bool) *AuthHandler {
	return &AuthHandler{
		service:    svc,
		redis:      rstore,
		jwtSecret:  []byte(jwtSecret),
		accessTTL:  time.Duration(accessTTL) * time.Minute,
		refreshTTL: time.Duration(refreshTTL) * time.Minute,
		secure:     secure,
	}
}

func setTokenCookies(w http.ResponseWriter, accessToken, refreshToken string, accessTTL, refreshTTL time.Duration, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(accessTTL.Seconds()),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(refreshTTL.Seconds()),
	})
}

func clearTokenCookies(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// Login is the JSON API handler for POST /api/v1/auth/login.
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

// Register is the JSON API handler for POST /api/v1/auth/register.
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

// Refresh handles token refresh. Checks refresh_token cookie first, then JSON body.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var refreshToken string

	cookie, err := r.Cookie("refresh_token")
	if err == nil && cookie.Value != "" {
		refreshToken = cookie.Value
	} else {
		var req dto.RefreshRequest
		if err := decodeJSON(r, &req); err != nil {
			response.ValidationError(w, r, "invalid request body")
			return
		}
		refreshToken = req.RefreshToken
	}

	if refreshToken == "" {
		response.Unauthorized(w, r)
		return
	}

	userID, err := h.redis.ValidateRefreshToken(r.Context(), refreshToken)
	if err != nil {
		slog.Debug("refresh token validation failed", "error", err)
		response.Unauthorized(w, r)
		return
	}

	accessToken, err := security.SignJWT(userID, "user", h.jwtSecret, h.accessTTL)
	if err != nil {
		response.InternalError(w, r)
		return
	}

	newRefreshToken, err := security.SignJWT(userID, "refresh", h.jwtSecret, h.refreshTTL)
	if err != nil {
		response.InternalError(w, r)
		return
	}

	h.redis.DeleteRefreshToken(r.Context(), refreshToken)
	if err := h.redis.CreateRefreshToken(r.Context(), userID, newRefreshToken, h.refreshTTL); err != nil {
		slog.Warn("failed to store new refresh token in redis", "error", err)
	}

	setTokenCookies(w, accessToken, newRefreshToken, h.accessTTL, h.refreshTTL, h.secure)

	response.JSON(w, http.StatusOK, dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	})
}

// FormLogin handles POST /login with form-urlencoded data.
func (h *AuthHandler) FormLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		templ.Handler(web.LoginForm("invalid form data")).ServeHTTP(w, r)
		return
	}

	req := dto.LoginRequest{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		slog.Debug("form login failed", "email", req.Email, "error", err)
		templ.Handler(web.LoginForm("invalid email or password")).ServeHTTP(w, r)
		return
	}

	claims, err := security.ValidateJWT(resp.AccessToken, h.jwtSecret)
	if err != nil {
		response.InternalError(w, r)
		return
	}

	if err := h.redis.CreateRefreshToken(r.Context(), claims.UserID, resp.RefreshToken, h.refreshTTL); err != nil {
		slog.Warn("failed to store refresh token in redis, continuing", "error", err)
	}

	setTokenCookies(w, resp.AccessToken, resp.RefreshToken, h.accessTTL, h.refreshTTL, h.secure)
	http.Redirect(w, r, "/app", http.StatusFound)
}

// FormRegister handles GET /register and POST /register.
func (h *AuthHandler) FormRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		templ.Handler(web.RegisterForm()).ServeHTTP(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		templ.Handler(web.RegisterForm("invalid form data")).ServeHTTP(w, r)
		return
	}

	req := dto.RegisterRequest{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
		Name:     r.FormValue("name"),
	}

	if err := h.service.Register(r.Context(), req); err != nil {
		slog.Debug("form register failed", "email", req.Email, "error", err)
		templ.Handler(web.RegisterForm(err.Error())).ServeHTTP(w, r)
		return
	}

	http.Redirect(w, r, "/?registered=true", http.StatusFound)
}

// FormLogout handles POST /logout.
func (h *AuthHandler) FormLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err == nil && cookie.Value != "" {
		if err := h.redis.DeleteRefreshToken(r.Context(), cookie.Value); err != nil {
			slog.Error("failed to delete refresh token from redis", "error", err)
		}
	}

	clearTokenCookies(w, h.secure)
	http.Redirect(w, r, "/", http.StatusFound)
}

func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
