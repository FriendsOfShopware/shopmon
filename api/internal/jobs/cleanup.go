package jobs

import (
	"context"
	"log/slog"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"go.opentelemetry.io/otel/codes"
)

type CleanupHandler struct {
	queries *queries.Queries
}

func NewCleanupHandler(q *queries.Queries) *CleanupHandler {
	return &CleanupHandler{queries: q}
}

func (h *CleanupHandler) HandleLockCleanup(ctx context.Context, _ LockCleanup) error {
	ctx, span := tracer.Start(ctx, "cleanup.locks")
	defer span.End()

	err := h.queries.CleanupExpiredLocks(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	slog.Info("cleaned up expired locks")
	return nil
}

func (h *CleanupHandler) HandleInvitationCleanup(ctx context.Context, _ InvitationCleanup) error {
	ctx, span := tracer.Start(ctx, "cleanup.invitations")
	defer span.End()

	err := h.queries.CleanupExpiredInvitations(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	slog.Info("cleaned up expired invitations")
	return nil
}
