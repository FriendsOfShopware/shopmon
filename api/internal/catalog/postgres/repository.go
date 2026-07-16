package postgres

import (
	"context"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/catalog"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
)

type VersionRepository struct {
	queries *queries.Queries
}

var _ catalog.VersionRepository = (*VersionRepository)(nil)

func NewVersionRepository(q *queries.Queries) *VersionRepository {
	return &VersionRepository{queries: q}
}

func (r *VersionRepository) ListShopwareVersions(ctx context.Context) ([]string, error) {
	versions, err := r.queries.ListShopwareVersions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list shopware versions: %w", err)
	}
	return versions, nil
}
