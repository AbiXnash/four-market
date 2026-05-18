package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/AbiXnash/4-market/internal/dto"
	"github.com/AbiXnash/4-market/internal/model"
	"github.com/AbiXnash/4-market/internal/repository"
	"github.com/AbiXnash/4-market/internal/security"
	"github.com/AbiXnash/4-market/internal/validator"
)

func generateID() string {
	b := make([]byte, 12)
	rand.Read(b)
	return hex.EncodeToString(b)
}

type AuthService struct {
	userRepo   repository.UserRepository
	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string, accessTTL, refreshTTL int) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtSecret:  []byte(jwtSecret),
		accessTTL:  time.Duration(accessTTL) * time.Minute,
		refreshTTL: time.Duration(refreshTTL) * time.Minute,
	}
}

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrUnauthorized       = errors.New("unauthorized")
)

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	if err := validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := security.CheckPassword(req.Password, user.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := security.SignJWT(user.ID, "user", s.jwtSecret, s.accessTTL)
	if err != nil {
		return nil, fmt.Errorf("sign access: %w", err)
	}

	refreshToken, err := security.SignJWT(user.ID, "refresh", s.jwtSecret, s.refreshTTL)
	if err != nil {
		return nil, fmt.Errorf("sign refresh: %w", err)
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) error {
	if err := validator.Struct(req); err != nil {
		return fmt.Errorf("validation: %w", err)
	}

	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return fmt.Errorf("check existing: %w", err)
	}
	if existing != nil {
		return ErrEmailTaken
	}

	hash, err := security.HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		ID:           generateID(),
		Email:        req.Email,
		PasswordHash: hash,
		Name:         req.Name,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (s *AuthService) Refresh(ctx context.Context, req dto.RefreshRequest) (*dto.AuthResponse, error) {
	if err := validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	claims, err := security.ValidateJWT(req.RefreshToken, s.jwtSecret)
	if err != nil {
		return nil, ErrInvalidToken
	}
	if claims.Role != "refresh" {
		return nil, ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		return nil, ErrUnauthorized
	}

	accessToken, err := security.SignJWT(user.ID, "user", s.jwtSecret, s.accessTTL)
	if err != nil {
		return nil, fmt.Errorf("sign access: %w", err)
	}

	return &dto.AuthResponse{
		AccessToken: accessToken,
	}, nil
}

