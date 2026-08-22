package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/packagesmirror"
)

type getPackagesTokenConfigurationOutput struct {
	Body api.PackagesTokenConfiguration
}

type getPackagesTokensInput struct {
	OrgID  api.OrgId  `path:"orgId"`
	ShopID api.ShopId `path:"shopId"`
}

type getPackagesTokensOutput struct {
	Body []api.PackagesToken
}

type createPackagesTokenInput struct {
	OrgID  api.OrgId  `path:"orgId"`
	ShopID api.ShopId `path:"shopId"`
	Body   api.CreatePackagesTokenRequest
}

type createPackagesTokenOutput struct {
	Status int
}

type deletePackagesTokenInput struct {
	OrgID   api.OrgId   `path:"orgId"`
	ShopID  api.ShopId  `path:"shopId"`
	TokenID api.TokenId `path:"tokenId"`
}

type deletePackagesTokenOutput struct {
	Status int
}

type syncPackagesTokenInput struct {
	OrgID   api.OrgId   `path:"orgId"`
	ShopID  api.ShopId  `path:"shopId"`
	TokenID api.TokenId `path:"tokenId"`
}

type syncPackagesTokenOutput struct {
	Status int
}

type packagesRemoteError struct {
	status int
	body   json.RawMessage
}

func (e *packagesRemoteError) Error() string {
	return string(e.body)
}

func (e *packagesRemoteError) GetStatus() int {
	return e.status
}

func (e *packagesRemoteError) ContentType(string) string {
	return "application/json"
}

func (e *packagesRemoteError) MarshalJSON() ([]byte, error) {
	return e.body, nil
}

func newPackagesRemoteError(err *packagesmirror.RemoteError) error {
	if len(err.Body) == 0 || !json.Valid(err.Body) {
		return huma.NewError(err.StatusCode, "packages API request failed")
	}
	return &packagesRemoteError{status: err.StatusCode, body: json.RawMessage(err.Body)}
}

// GetPackagesTokenConfiguration returns the packages token configuration.
func (h *Handler) GetPackagesTokenConfiguration(_ context.Context, _ *struct{}) (*getPackagesTokenConfigurationOutput, error) {
	configuration := h.packages.Configuration()
	return &getPackagesTokenConfigurationOutput{Body: api.PackagesTokenConfiguration{
		Configured:  configuration.Configured,
		ComposerUrl: configuration.ComposerURL,
	}}, nil
}

// GetPackagesTokens lists packages tokens for a shop.
func (h *Handler) GetPackagesTokens(ctx context.Context, input *getPackagesTokensInput) (*getPackagesTokensOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	tokens, err := h.packages.List(ctx, packagesmirror.ShopCommand{
		UserID:         user.ID,
		OrganizationID: input.OrgID,
		ShopID:         int32(input.ShopID),
	})
	if err != nil {
		return nil, h.writePackagesError(ctx, "list packages tokens", err)
	}

	response := make([]api.PackagesToken, 0, len(tokens))
	for _, token := range tokens {
		response = append(response, api.PackagesToken{
			Id:           token.ID,
			Source:       token.Source,
			LastSyncedAt: token.LastSyncedAt,
		})
	}
	return &getPackagesTokensOutput{Body: response}, nil
}

// CreatePackagesToken creates a new packages token.
func (h *Handler) CreatePackagesToken(ctx context.Context, input *createPackagesTokenInput) (*createPackagesTokenOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.packages.Create(ctx, packagesmirror.CreateCommand{
		UserID:         user.ID,
		OrganizationID: input.OrgID,
		ShopID:         int32(input.ShopID),
		Token:          input.Body.Token,
	}); err != nil {
		return nil, h.writePackagesError(ctx, "create packages token", err)
	}

	return &createPackagesTokenOutput{Status: http.StatusCreated}, nil
}

// DeletePackagesToken deletes a packages token.
func (h *Handler) DeletePackagesToken(ctx context.Context, input *deletePackagesTokenInput) (*deletePackagesTokenOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.packages.Delete(ctx, packagesmirror.TokenCommand{
		UserID:         user.ID,
		OrganizationID: input.OrgID,
		ShopID:         int32(input.ShopID),
		TokenID:        int(input.TokenID),
	}); err != nil {
		return nil, h.writePackagesError(ctx, "delete packages token", err)
	}

	return &deletePackagesTokenOutput{Status: http.StatusNoContent}, nil
}

// SyncPackagesToken syncs a packages token.
func (h *Handler) SyncPackagesToken(ctx context.Context, input *syncPackagesTokenInput) (*syncPackagesTokenOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.packages.Sync(ctx, packagesmirror.TokenCommand{
		UserID:         user.ID,
		OrganizationID: input.OrgID,
		ShopID:         int32(input.ShopID),
		TokenID:        int(input.TokenID),
	}); err != nil {
		return nil, h.writePackagesError(ctx, "sync packages token", err)
	}

	return &syncPackagesTokenOutput{Status: http.StatusNoContent}, nil
}

func (h *Handler) writePackagesError(ctx context.Context, operation string, err error) error {
	switch {
	case errors.Is(err, packagesmirror.ErrNotConfigured):
		return huma.Error404NotFound("packages API not configured")
	case errors.Is(err, packagesmirror.ErrNotAuthorized):
		return huma.Error403Forbidden("not a member of this organization")
	case errors.Is(err, packagesmirror.ErrShopNotFound):
		return huma.Error404NotFound("shop not found")
	case errors.Is(err, packagesmirror.ErrShopOrganizationMismatch):
		return huma.Error403Forbidden("shop does not belong to this organization")
	case errors.Is(err, packagesmirror.ErrTokenNotFound):
		return huma.Error404NotFound("token not found")
	case errors.Is(err, packagesmirror.ErrTokenRequired):
		return huma.Error400BadRequest("token is required")
	default:
		if remoteError, ok := errors.AsType[*packagesmirror.RemoteError](err); ok {
			return newPackagesRemoteError(remoteError)
		}
		slog.ErrorContext(ctx, "packages operation failed", "operation", operation, "error", err)
		if errors.Is(err, packagesmirror.ErrRemote) {
			return huma.Error502BadGateway("packages API request failed")
		}
		return huma.Error500InternalServerError("packages operation failed")
	}
}
