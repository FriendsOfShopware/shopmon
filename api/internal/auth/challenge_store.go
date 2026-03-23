package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ChallengeStore stores short-lived challenge data in Redis for OAuth and WebAuthn flows.
type ChallengeStore struct {
	client *redis.Client
	prefix string
}

// NewChallengeStore creates a new Redis-backed challenge store.
func NewChallengeStore(redisURL, prefix string) *ChallengeStore {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		opts = &redis.Options{Addr: "localhost:6379"}
	}

	return &ChallengeStore{
		client: redis.NewClient(opts),
		prefix: prefix,
	}
}

// Set stores a value with a TTL.
func (s *ChallengeStore) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal challenge: %w", err)
	}
	return s.client.Set(ctx, s.prefix+key, data, ttl).Err()
}

// Get retrieves and deletes a value (consume-once).
func (s *ChallengeStore) Get(ctx context.Context, key string, dest interface{}) error {
	fullKey := s.prefix + key

	data, err := s.client.GetDel(ctx, fullKey).Bytes()
	if err != nil {
		return fmt.Errorf("get challenge: %w", err)
	}

	return json.Unmarshal(data, dest)
}

// Peek retrieves a value without deleting it.
func (s *ChallengeStore) Peek(ctx context.Context, key string, dest interface{}) error {
	data, err := s.client.Get(ctx, s.prefix+key).Bytes()
	if err != nil {
		return fmt.Errorf("peek challenge: %w", err)
	}
	return json.Unmarshal(data, dest)
}
