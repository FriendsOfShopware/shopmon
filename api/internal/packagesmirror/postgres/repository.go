package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/packagesmirror"
	"github.com/jackc/pgx/v5"
)

type ShopRepository struct {
	queries *queries.Queries
}

var _ packagesmirror.ShopRepository = (*ShopRepository)(nil)

func NewShopRepository(q *queries.Queries) *ShopRepository {
	return &ShopRepository{queries: q}
}

func (r *ShopRepository) OrganizationIDForShop(ctx context.Context, shopID int32) (string, error) {
	shop, err := r.queries.GetShopByID(ctx, shopID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", packagesmirror.ErrShopNotFound
		}
		return "", fmt.Errorf("get shop: %w", err)
	}
	return shop.OrganizationID, nil
}
