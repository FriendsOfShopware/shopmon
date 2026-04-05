package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/jobs"
	goqueue "github.com/shyim/go-queue"
)

type createEnvironmentCommand struct {
	Name             string
	ShopURL          string
	ClientID         string
	ClientSecret     string
	ShopID           int32
	EnvironmentToken string
}

func (h *Handler) runCreateEnvironment(ctx context.Context, organizationID string, cmd createEnvironmentCommand) (int32, error) {
	encryptedSecret, err := h.encryptShopSecret(cmd.ClientSecret)
	if err != nil {
		return 0, err
	}

	shopInfo, err := h.validateShopConnection(ctx, cmd.ShopURL, cmd.ClientID, cmd.ClientSecret, cmd.EnvironmentToken)
	if err != nil {
		return 0, err
	}

	environmentID, err := h.queries.CreateEnvironment(ctx, queries.CreateEnvironmentParams{
		OrganizationID:   organizationID,
		ShopID:           cmd.ShopID,
		Name:             cmd.Name,
		Url:              cmd.ShopURL,
		ClientID:         cmd.ClientID,
		ClientSecret:     encryptedSecret,
		ShopwareVersion:  shopInfo.Version,
		EnvironmentToken: cmd.EnvironmentToken,
	})
	if err != nil {
		return 0, fmt.Errorf("create environment: %w", err)
	}

	if err := goqueue.Dispatch(ctx, h.bus, jobs.EnvironmentScrape{EnvironmentID: environmentID}); err != nil {
		slog.Error("failed to enqueue initial scrape", "environmentId", environmentID, "error", err)
	}

	return environmentID, nil
}

type updateEnvironmentCommand struct {
	Name         string
	URL          string
	ClientID     string
	ClientSecret string
	Ignores      json.RawMessage
	ShopID       int32
}

func (h *Handler) buildUpdateEnvironmentCommand(ctx context.Context, environment *queries.GetEnvironmentByIDRow, req api.UpdateEnvironmentRequest) (*updateEnvironmentCommand, error) {
	cmd := &updateEnvironmentCommand{
		Name:         environment.Name,
		URL:          environment.Url,
		ClientID:     environment.ClientID,
		ClientSecret: environment.ClientSecret,
		Ignores:      environment.Ignores,
		ShopID:       int32(req.ShopId),
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
		if _, err := h.validateShopConnection(ctx, cmd.URL, cmd.ClientID, *req.ClientSecret, environment.EnvironmentToken); err != nil {
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

func (h *Handler) runUpdateEnvironment(ctx context.Context, environmentID int32, cmd *updateEnvironmentCommand) error {
	if err := h.queries.UpdateEnvironment(ctx, queries.UpdateEnvironmentParams{
		Name:         cmd.Name,
		Url:          cmd.URL,
		ClientID:     cmd.ClientID,
		ClientSecret: cmd.ClientSecret,
		Ignores:      cmd.Ignores,
		ShopID:       cmd.ShopID,
		ID:           environmentID,
	}); err != nil {
		return fmt.Errorf("update environment: %w", err)
	}

	return nil
}
