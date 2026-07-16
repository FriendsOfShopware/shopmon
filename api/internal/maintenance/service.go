package maintenance

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
)

type Service struct {
	queries *queries.Queries
}

func NewService(q *queries.Queries) *Service {
	return &Service{queries: q}
}

func (s *Service) CleanupLocks(ctx context.Context) error {
	if err := s.queries.CleanupExpiredLocks(ctx); err != nil {
		return fmt.Errorf("cleanup expired locks: %w", err)
	}
	slog.Info("cleaned up expired locks")
	return nil
}

func (s *Service) CleanupInvitations(ctx context.Context) error {
	if err := s.queries.CleanupExpiredInvitations(ctx); err != nil {
		return fmt.Errorf("cleanup expired invitations: %w", err)
	}
	slog.Info("cleaned up expired invitations")
	return nil
}

func (s *Service) CleanupOldData(ctx context.Context) error {
	if err := s.queries.CleanupExpiredSessions(ctx); err != nil {
		return fmt.Errorf("cleanup expired sessions: %w", err)
	}
	if err := s.queries.CleanupOldNotifications(ctx); err != nil {
		return fmt.Errorf("cleanup old notifications: %w", err)
	}
	if err := s.queries.CleanupOldSitespeedData(ctx); err != nil {
		return fmt.Errorf("cleanup old sitespeed data: %w", err)
	}
	if err := s.queries.CleanupExpiredBans(ctx); err != nil {
		return fmt.Errorf("cleanup expired bans: %w", err)
	}
	// The store catalog itself (store_extension and its versions, changelogs and
	// images) is retained even once no environment links it, so historical
	// changelog data survives. Only the internal bookkeeping is pruned: sync
	// state for names no environment has anymore, and compatibility rows for
	// Shopware versions no environment runs. Both rebuild on the next sync.
	if err := s.queries.CleanupOrphanedStoreExtensionSyncStates(ctx); err != nil {
		return fmt.Errorf("cleanup orphaned store extension sync states: %w", err)
	}
	if err := s.queries.CleanupUnusedStoreExtensionCompatibility(ctx); err != nil {
		return fmt.Errorf("cleanup unused store extension compatibility: %w", err)
	}
	slog.Info("cleaned up old data: sessions, notifications, sitespeed, bans, store bookkeeping")
	return nil
}
