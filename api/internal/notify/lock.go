package notify

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// lockAcquirer atomically acquires a dedup lock for a key and TTL, returning
// true only when this caller obtained it. A held lock is (false, nil); an error
// means the lock state is unknown. release drops a lock whose send failed, so
// the retry is not mistaken for a duplicate.
type lockAcquirer interface {
	acquire(ctx context.Context, key string, ttl time.Duration) (bool, error)
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
	slog.Error("notify: failed to acquire dedup lock", "key", key, "error", err)
	return false, fmt.Errorf("acquire dedup lock: %w", err)
}

// releaseTimeout bounds lock cleanup, which runs detached from the caller's
// context: a send that failed because the job was cancelled must still drop
// its lock, or the retry would be deduped away.
const releaseTimeout = 5 * time.Second

func (l queryLocker) release(ctx context.Context, key string) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), releaseTimeout)
	defer cancel()
	if err := l.q.ReleaseLock(ctx, key); err != nil {
		slog.Error("notify: failed to release dedup lock", "key", key, "error", err)
	}
}
