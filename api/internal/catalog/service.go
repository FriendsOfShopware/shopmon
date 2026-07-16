package catalog

import (
	"context"
	"fmt"
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

type VersionRepository interface {
	ListShopwareVersions(ctx context.Context) ([]string, error)
}

type CompatibilityGateway interface {
	CheckCompatibility(ctx context.Context, command CompatibilityCommand) ([]byte, error)
}

type Service struct {
	versions      VersionRepository
	compatibility CompatibilityGateway
}

func NewService(versions VersionRepository, compatibility CompatibilityGateway) *Service {
	return &Service{versions: versions, compatibility: compatibility}
}

func (s *Service) CheckCompatibility(ctx context.Context, command CompatibilityCommand) ([]byte, error) {
	result, err := s.compatibility.CheckCompatibility(ctx, command)
	if err != nil {
		return nil, fmt.Errorf("check extension compatibility: %w", err)
	}
	return result, nil
}

func (s *Service) ListShopwareVersions(ctx context.Context) ([]string, error) {
	versions, err := s.versions.ListShopwareVersions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list shopware versions: %w", err)
	}
	return versions, nil
}
