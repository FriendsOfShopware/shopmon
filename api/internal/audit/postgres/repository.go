package postgres

import (
	"context"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/audit"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
)

type Repository struct {
	queries *queries.Queries
}

func NewRepository(q *queries.Queries) *Repository {
	return &Repository{queries: q}
}

func (r *Repository) Create(ctx context.Context, event audit.Event) error {
	if err := r.queries.CreateAuditLog(ctx, queries.CreateAuditLogParams{
		ActorUserID:  event.ActorUserID,
		Action:       event.Action,
		TargetUserID: event.TargetUserID,
		Detail:       event.Detail,
		IpAddress:    event.IPAddress,
	}); err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}
