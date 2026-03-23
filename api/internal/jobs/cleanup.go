package jobs

import (
	"context"
	"log/slog"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
)

type CleanupHandler struct {
	queries *queries.Queries
}

func NewCleanupHandler(q *queries.Queries) *CleanupHandler {
	return &CleanupHandler{queries: q}
}

func (h *CleanupHandler) HandleLockCleanup(ctx context.Context, _ LockCleanup) error {
	err := h.queries.CleanupExpiredLocks(ctx)
	if err != nil {
		return err
	}
	slog.Info("cleaned up expired locks")
	return nil
}

func (h *CleanupHandler) HandleInvitationCleanup(ctx context.Context, _ InvitationCleanup) error {
	err := h.queries.CleanupExpiredInvitations(ctx)
	if err != nil {
		return err
	}
	slog.Info("cleaned up expired invitations")
	return nil
}
