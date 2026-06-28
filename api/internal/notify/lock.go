package notify

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// lockAcquirer atomically acquires a dedup lock for a key and TTL, returning
// true only when this caller obtained it.
type lockAcquirer interface {
	acquire(ctx context.Context, key string, ttl time.Duration) bool
}

// queryLocker implements lockAcquirer via the shared `lock` table.
type queryLocker struct {
	q *queries.Queries
}

func (l queryLocker) acquire(ctx context.Context, key string, ttl time.Duration) bool {
	_, err := l.q.AcquireLockIfFree(ctx, queries.AcquireLockIfFreeParams{
		Key:     key,
		Expires: pgtype.Timestamp{Time: time.Now().Add(ttl), Valid: true},
	})
	if err == nil {
		return true
	}
	if errors.Is(err, pgx.ErrNoRows) {
		// Lock already held — deduped.
		return false
	}
	// On an unknown error, do not send: avoids spamming when lock state is
	// unclear.
	slog.Error("notify: failed to acquire dedup lock", "key", key, "error", err)
	return false
}
