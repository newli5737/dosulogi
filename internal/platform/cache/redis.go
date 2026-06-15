package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store struct {
	client *redis.Client
	ok     bool
}

func New(client *redis.Client, connected bool) *Store {
	return &Store{client: client, ok: connected}
}

func (s *Store) Available() bool { return s != nil && s.ok && s.client != nil }

func (s *Store) AllowRate(ctx context.Context, userID, route string, limit int, window time.Duration) (bool, error) {
	if !s.Available() {
		return true, nil
	}
	if window <= 0 {
		window = time.Minute
	}
	key := fmt.Sprintf("rate_limit:%s:%s", userID, route)
	n, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return true, err
	}
	if n == 1 {
		_ = s.client.Expire(ctx, key, window).Err()
	}
	return n <= int64(limit), nil
}

func (s *Store) MarkSLANotified(ctx context.Context, ticketID string) error {
	if !s.Available() {
		return nil
	}
	key := fmt.Sprintf("sla_notified:%s", ticketID)
	return s.client.Set(ctx, key, "1", time.Hour).Err()
}

func (s *Store) SLAWasNotified(ctx context.Context, ticketID string) (bool, error) {
	if !s.Available() {
		return false, nil
	}
	n, err := s.client.Exists(ctx, fmt.Sprintf("sla_notified:%s", ticketID)).Result()
	return n > 0, err
}
