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
	if err := h.queries.CleanupExpiredBans(ctx); err != nil {
		return fmt.Errorf("cleanup expired bans: %w", err)
	}
	// The store catalog itself (store_extension and its versions, changelogs and
	// images) is retained even once no environment links it, so historical
	// changelog data survives. Only the internal bookkeeping is pruned: sync
	// state for names no environment has anymore, and compatibility rows for
	// Shopware versions no environment runs. Both rebuild on the next sync.
	if err := h.queries.CleanupOrphanedStoreExtensionSyncStates(ctx); err != nil {
		return fmt.Errorf("cleanup orphaned store extension sync states: %w", err)
	}
	if err := h.queries.CleanupUnusedStoreExtensionCompatibility(ctx); err != nil {
		return fmt.Errorf("cleanup unused store extension compatibility: %w", err)
	}
	slog.Info("cleaned up old data: sessions, notifications, sitespeed, bans, store bookkeeping")
	return nil
}
