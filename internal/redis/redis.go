package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	refreshKeyPrefix = "refresh:"
	sessionKeyPrefix = "session:"
)

type Store struct {
	client *redis.Client
}

func Connect(ctx context.Context, addr, password string, db int) (*Store, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}

	return &Store{client: client}, nil
}

func (s *Store) Close() error {
	if s.client == nil {
		return nil
	}
	return s.client.Close()
}

func (s *Store) IsAvailable() bool {
	return s.client != nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (s *Store) CreateRefreshToken(ctx context.Context, userID, refreshToken string, ttl time.Duration) error {
	if s.client == nil {
		return fmt.Errorf("redis not available")
	}
	key := refreshKeyPrefix + hashToken(refreshToken)
	return s.client.Set(ctx, key, userID, ttl).Err()
}

func (s *Store) ValidateRefreshToken(ctx context.Context, refreshToken string) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("redis not available")
	}
	key := refreshKeyPrefix + hashToken(refreshToken)
	userID, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("refresh token not found or expired")
		}
		return "", fmt.Errorf("redis error: %w", err)
	}
	return userID, nil
}

func (s *Store) DeleteRefreshToken(ctx context.Context, refreshToken string) error {
	if s.client == nil {
		return nil
	}
	key := refreshKeyPrefix + hashToken(refreshToken)
	return s.client.Del(ctx, key).Err()
}

func (s *Store) DeleteUserSessions(ctx context.Context, userID string) error {
	if s.client == nil {
		return nil
	}
	pattern := refreshKeyPrefix + "*"
	iter := s.client.Scan(ctx, 0, pattern, 0).Iterator()

	var keys []string
	for iter.Next(ctx) {
		val, err := s.client.Get(ctx, iter.Val()).Result()
		if err != nil {
			continue
		}
		if val == userID {
			keys = append(keys, iter.Val())
		}
	}

	if len(keys) > 0 {
		return s.client.Del(ctx, keys...).Err()
	}
	return iter.Err()
}

func (s *Store) CreateSession(ctx context.Context, sessionID, userID string, ttl time.Duration) error {
	if s.client == nil {
		return fmt.Errorf("redis not available")
	}
	key := sessionKeyPrefix + sessionID
	return s.client.Set(ctx, key, userID, ttl).Err()
}

func (s *Store) GetSession(ctx context.Context, sessionID string) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("redis not available")
	}
	key := sessionKeyPrefix + sessionID
	userID, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("session not found")
		}
		return "", fmt.Errorf("redis error: %w", err)
	}
	return userID, nil
}

func (s *Store) DeleteSession(ctx context.Context, sessionID string) error {
	if s.client == nil {
		return nil
	}
	key := sessionKeyPrefix + sessionID
	return s.client.Del(ctx, key).Err()
}
