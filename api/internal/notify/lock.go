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
// true only when this caller obtained it, and releases one again.
type lockAcquirer interface {
	acquire(ctx context.Context, key string, ttl time.Duration) bool
	release(ctx context.Context, key string)
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

// release drops a dedup lock this caller is holding. Used when the delivery it
// was guarding failed: keeping the lock for its full TTL would make the next
// attempt skip the channel, report nothing failed, and let the caller mark the
// alert as sent -- losing it.
func (l queryLocker) release(ctx context.Context, key string) {
	if err := l.q.ReleaseLock(ctx, key); err != nil {
		slog.Error("notify: failed to release dedup lock", "key", key, "error", err)
	}
}
