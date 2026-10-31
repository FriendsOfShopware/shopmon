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

// lockAcquirer atomically acquires a dedup lock for a key and TTL, and releases
// one again.
//
// acquire separates "somebody else holds it" from "the lock store could not be
// reached": the first means the channel was already served and skipping it is
// correct, the second means nothing is known, and a caller that records
// delivery state must treat it as a failure rather than as a successful skip.
type lockAcquirer interface {
	acquire(ctx context.Context, key string, ttl time.Duration) (acquired bool, err error)
	release(ctx context.Context, key string)
}

// queryLocker implements lockAcquirer via the shared `lock` table.
type queryLocker struct {
	q *queries.Queries
}

func (l queryLocker) acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	_, err := l.q.AcquireLockIfFree(ctx, queries.AcquireLockIfFreeParams{
		Key:     key,
		Expires: pgtype.Timestamp{Time: time.Now().Add(ttl), Valid: true},
	})
	if err == nil {
		return true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		// Lock already held — deduped.
		return false, nil
	}
	// Lock state unknown. Do not send, but say why: a caller recording "these
	// recipients were told" must not read this as a successful skip.
	slog.Error("notify: failed to acquire dedup lock", "key", key, "error", err)
	return false, err
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
