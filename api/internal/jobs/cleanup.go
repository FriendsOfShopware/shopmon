package jobs

import (
	"context"
	"fmt"
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

func (h *CleanupHandler) HandleOldDataCleanup(ctx context.Context, _ OldDataCleanup) error {
	if err := h.queries.CleanupExpiredSessions(ctx); err != nil {
		return fmt.Errorf("cleanup expired sessions: %w", err)
	}
	if err := h.queries.CleanupOldNotifications(ctx); err != nil {
		return fmt.Errorf("cleanup old notifications: %w", err)
	}
	if err := h.queries.CleanupOldSitespeedData(ctx); err != nil {
		return fmt.Errorf("cleanup old sitespeed data: %w", err)
	}
	if err := h.queries.CleanupOldChangelogData(ctx); err != nil {
		return fmt.Errorf("cleanup old changelog data: %w", err)
	}
	slog.Info("cleaned up old data: sessions, notifications, sitespeed, changelog")
	return nil
}
