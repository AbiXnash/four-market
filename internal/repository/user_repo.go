package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/AbiXnash/4-market/internal/model"
)

type userRepo struct {
	mu    sync.RWMutex
	users map[string]*model.User
	byID  map[string]*model.User
}

func NewUserRepo() UserRepository {
	return &userRepo{
		users: make(map[string]*model.User),
		byID:  make(map[string]*model.User),
	}
}

func (r *userRepo) Create(_ context.Context, user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.Email]; exists {
		return fmt.Errorf("user with email %s already exists", user.Email)
	}

	r.users[user.Email] = user
	r.byID[user.ID] = user
	return nil
}

func (r *userRepo) FindByEmail(_ context.Context, email string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[email]
	if !exists {
		return nil, nil
	}
	return user, nil
}

func (r *userRepo) FindByID(_ context.Context, id string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.byID[id]
	if !exists {
		return nil, nil
	}
	return user, nil
}
