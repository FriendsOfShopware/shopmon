package catalog

import (
	"context"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
)

type Extension struct {
	Name    string
	Version string
}

type CompatibilityCommand struct {
	CurrentVersion string
	FutureVersion  string
	Extensions     []Extension
}

type CompatibilityGateway interface {
	CheckCompatibility(ctx context.Context, command CompatibilityCommand) ([]byte, error)
}

type Service struct {
	queries       *queries.Queries
	compatibility CompatibilityGateway
}

func NewService(q *queries.Queries, compatibility CompatibilityGateway) *Service {
	return &Service{queries: q, compatibility: compatibility}
}

func (s *Service) CheckCompatibility(ctx context.Context, command CompatibilityCommand) ([]byte, error) {
	result, err := s.compatibility.CheckCompatibility(ctx, command)
	if err != nil {
		return nil, fmt.Errorf("check extension compatibility: %w", err)
	}
	return result, nil
}

func (s *Service) ListShopwareVersions(ctx context.Context) ([]string, error) {
	versions, err := s.queries.ListShopwareVersions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list shopware versions: %w", err)
	}
	return versions, nil
}
