package auth

import (
	"context"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/jackc/pgx/v5"
)

func (h *AuthHandler) withTx(ctx context.Context, fn func(*queries.Queries) error) error {
	tx, err := h.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(h.queries.WithTx(tx)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
