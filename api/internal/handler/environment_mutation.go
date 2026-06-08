package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/jobs"
	"github.com/jackc/pgx/v5"
	goqueue "github.com/shyim/go-queue"
)

// withTx runs fn inside a database transaction, committing on success and rolling
// back on error.
func (h *Handler) withTx(ctx context.Context, fn func(*queries.Queries) error) error {
	tx, err := h.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(h.queries.WithTx(tx)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

type createEnvironmentCommand struct {
	Name             string
	ShopURL          string
	ClientID         string
	ClientSecret     string
	ShopID           int32
	EnvironmentToken string
	// ShopwareVersion, when set, is the version discovered by a prior
	// validateShopConnection call. Supplying it lets insertEnvironment skip the
	// network round-trip so the validation never happens inside a transaction.
	ShopwareVersion string
}

func (h *Handler) runCreateEnvironment(ctx context.Context, organizationID string, cmd createEnvironmentCommand) (int32, error) {
	if cmd.ShopwareVersion == "" {
		shopInfo, err := h.validateShopConnection(ctx, cmd.ShopURL, cmd.ClientID, cmd.ClientSecret, cmd.EnvironmentToken)
		if err != nil {
			return 0, err
		}
		cmd.ShopwareVersion = shopInfo.Version
	}

	environmentID, err := h.insertEnvironment(ctx, h.queries, organizationID, cmd)
	if err != nil {
		return 0, err
	}

	h.enqueueInitialScrape(ctx, environmentID)

	return environmentID, nil
}

// insertEnvironment inserts the environment row using the provided queries handle
// (which may be transaction-scoped). It performs NO network I/O: the caller must
// have already validated the shop connection and populated cmd.ShopwareVersion,
// so this is safe to call inside a transaction. It does not enqueue the initial
// scrape; callers must do that after any surrounding transaction has committed
// via enqueueInitialScrape.
func (h *Handler) insertEnvironment(ctx context.Context, q *queries.Queries, organizationID string, cmd createEnvironmentCommand) (int32, error) {
	encryptedSecret, err := h.encryptShopSecret(cmd.ClientSecret)
	if err != nil {
		return 0, err
	}

	environmentID, err := q.CreateEnvironment(ctx, queries.CreateEnvironmentParams{
		OrganizationID:   organizationID,
		ShopID:           cmd.ShopID,
		Name:             cmd.Name,
		Url:              cmd.ShopURL,
		ClientID:         cmd.ClientID,
		ClientSecret:     encryptedSecret,
		ShopwareVersion:  cmd.ShopwareVersion,
		EnvironmentToken: cmd.EnvironmentToken,
	})
	if err != nil {
		return 0, fmt.Errorf("create environment: %w", err)
	}

	return environmentID, nil
}

func (h *Handler) enqueueInitialScrape(ctx context.Context, environmentID int32) {
	if err := goqueue.Dispatch(ctx, h.bus, jobs.EnvironmentScrape{EnvironmentID: environmentID}); err != nil {
		slog.Error("failed to enqueue initial scrape", "environmentId", environmentID, "error", err)
	}
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

	if cmd.ShopID != environment.ShopID {
		shopOrgID, err := h.queries.GetShopOrganizationID(ctx, cmd.ShopID)
		if err != nil {
			return nil, fmt.Errorf("target shop not found: %w", err)
		}
		if shopOrgID != environment.OrganizationID {
			return nil, fmt.Errorf("target shop belongs to a different organization")
		}
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
