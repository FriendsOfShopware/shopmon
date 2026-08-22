package redisstate

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Store persists consume-once identity flow state such as OAuth, OIDC and
// WebAuthn challenges.
type Store struct {
	client *redis.Client
	prefix string
}

func New(redisURL, prefix string) (*Store, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse identity state Redis URL: %w", err)
	}
	return &Store{client: redis.NewClient(options), prefix: prefix}, nil
}

func (s *Store) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal identity state: %w", err)
	}
	if err := s.client.Set(ctx, s.prefix+key, data, ttl).Err(); err != nil {
		return fmt.Errorf("set identity state: %w", err)
	}
	return nil
}

func (s *Store) Get(ctx context.Context, key string, destination any) error {
	data, err := s.client.GetDel(ctx, s.prefix+key).Bytes()
	if err != nil {
		return fmt.Errorf("consume identity state: %w", err)
	}
	if err := json.Unmarshal(data, destination); err != nil {
		return fmt.Errorf("decode identity state: %w", err)
	}
	return nil
}

func (s *Store) Close() error {
	if err := s.client.Close(); err != nil {
		return fmt.Errorf("close identity state store: %w", err)
	}
	return nil
}
