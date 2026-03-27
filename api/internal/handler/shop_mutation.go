package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/jobs"
	goqueue "github.com/shyim/go-queue"
)

type createShopCommand struct {
	Name         string
	ShopURL      string
	ClientID     string
	ClientSecret string
	ProjectID    int32
	ShopToken    string
}

func (h *Handler) runCreateShop(ctx context.Context, organizationID string, cmd createShopCommand) (int32, error) {
	encryptedSecret, err := h.encryptShopSecret(cmd.ClientSecret)
	if err != nil {
		return 0, err
	}

	shopInfo, err := h.validateShopConnection(ctx, cmd.ShopURL, cmd.ClientID, cmd.ClientSecret, cmd.ShopToken)
	if err != nil {
		return 0, err
	}

	shopID, err := h.queries.CreateShop(ctx, queries.CreateShopParams{
		OrganizationID:  organizationID,
		ProjectID:       cmd.ProjectID,
		Name:            cmd.Name,
		Url:             cmd.ShopURL,
		ClientID:        cmd.ClientID,
		ClientSecret:    encryptedSecret,
		ShopwareVersion: shopInfo.Version,
		ShopToken:       cmd.ShopToken,
	})
	if err != nil {
		return 0, fmt.Errorf("create shop: %w", err)
	}

	if err := goqueue.Dispatch(ctx, h.bus, jobs.ShopScrape{ShopID: shopID}); err != nil {
		return shopID, fmt.Errorf("enqueue initial scrape: %w", err)
	}

	return shopID, nil
}

type updateShopCommand struct {
	Name         string
	URL          string
	ClientID     string
	ClientSecret string
	Ignores      json.RawMessage
	ProjectID    int32
}

func (h *Handler) buildUpdateShopCommand(ctx context.Context, shop *queries.GetShopByIDRow, req api.UpdateShopRequest) (*updateShopCommand, error) {
	cmd := &updateShopCommand{
		Name:         shop.Name,
		URL:          shop.Url,
		ClientID:     shop.ClientID,
		ClientSecret: shop.ClientSecret,
		Ignores:      shop.Ignores,
		ProjectID:    int32(req.ProjectId),
	}

	if req.Name != nil {
		cmd.Name = *req.Name
	}
	if req.ShopUrl != nil {
		cmd.URL = *req.ShopUrl
	}
	if req.ClientId != nil {
		cmd.ClientID = *req.ClientId
	}
	if req.ClientSecret != nil {
		if _, err := h.validateShopConnection(ctx, cmd.URL, cmd.ClientID, *req.ClientSecret, shop.ShopToken); err != nil {
			return nil, err
		}

		encryptedSecret, err := h.encryptShopSecret(*req.ClientSecret)
		if err != nil {
			return nil, err
		}
		cmd.ClientSecret = encryptedSecret
	}
	if req.Ignores != nil {
		ignoresJSON, err := json.Marshal(req.Ignores)
		if err != nil {
			return nil, fmt.Errorf("marshal ignores: %w", err)
		}
		cmd.Ignores = ignoresJSON
	}

	return cmd, nil
}

func (h *Handler) runUpdateShop(ctx context.Context, shopID int32, cmd *updateShopCommand) error {
	if err := h.queries.UpdateShop(ctx, queries.UpdateShopParams{
		Name:         cmd.Name,
		Url:          cmd.URL,
		ClientID:     cmd.ClientID,
		ClientSecret: cmd.ClientSecret,
		Ignores:      cmd.Ignores,
		ProjectID:    cmd.ProjectID,
		ID:           shopID,
	}); err != nil {
		return fmt.Errorf("update shop: %w", err)
	}

	return nil
}
