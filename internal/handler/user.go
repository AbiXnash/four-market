package handler

import (
	"log/slog"
	"net/http"

	"github.com/AbiXnash/4-market/internal/middleware"
	"github.com/AbiXnash/4-market/internal/repository"
	"github.com/AbiXnash/4-market/internal/response"
	"github.com/AbiXnash/4-market/internal/security"
)

type UserHandler struct {
	userRepo repository.UserRepository
}

func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserKey).(*security.Claims)
	if !ok || claims == nil {
		response.Unauthorized(w, r)
		return
	}

	user, err := h.userRepo.FindByID(r.Context(), claims.UserID)
	if err != nil {
		slog.Error("user lookup failed", "user_id", claims.UserID, "error", err)
		response.InternalError(w, r)
		return
	}
	if user == nil {
		slog.Warn("user not found for id in token", "user_id", claims.UserID)
		response.NotFound(w, r)
		return
	}

	slog.Debug("user data accessed", "user_id", user.ID, "email", user.Email)
	response.JSON(w, http.StatusOK, map[string]any{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
	})
}
