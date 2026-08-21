package handler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/organization"
	organizationsso "github.com/friendsofshopware/shopmon/api/internal/organization/sso"
)

type getSsoProvidersInput struct {
	OrgID string `path:"orgId"`
}

type getSsoProvidersOutput struct {
	Body []api.SsoProvider
}

// GetSsoProviders lists SSO providers for an organization.
func (h *Handler) GetSsoProviders(ctx context.Context, input *getSsoProvidersInput) (*getSsoProvidersOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	providers, err := h.sso.List(ctx, user.ID, input.OrgID)
	if err != nil {
		return nil, h.writeSSOError(ctx, "list SSO providers", err)
	}

	response := make([]api.SsoProvider, 0, len(providers))
	for _, provider := range providers {
		response = append(response, api.SsoProvider{
			Id:                    provider.ID,
			Issuer:                provider.Issuer,
			Domain:                provider.Domain,
			AuthorizationEndpoint: provider.Configuration.AuthorizationEndpoint,
			TokenEndpoint:         provider.Configuration.TokenEndpoint,
			JwksEndpoint:          provider.Configuration.JWKSEndpoint,
			ClientId:              provider.Configuration.ClientID,
		})
	}

	return &getSsoProvidersOutput{Body: response}, nil
}

type ssoProviderPathInput struct {
	OrgID      string `path:"orgId"`
	ProviderID string `path:"providerId"`
}

type updateSsoProviderInput struct {
	OrgID      string `path:"orgId"`
	ProviderID string `path:"providerId"`
	Body       api.UpdateSsoProviderRequest
}

// UpdateSsoProvider updates an SSO provider.
func (h *Handler) UpdateSsoProvider(ctx context.Context, input *updateSsoProviderInput) (*struct{}, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	err = h.sso.Update(ctx, organizationsso.UpdateCommand{
		UserID:         user.ID,
		OrganizationID: input.OrgID,
		ProviderID:     input.ProviderID,
		Issuer:         input.Body.Issuer,
		Domain:         input.Body.Domain,
		Configuration: organizationsso.ProviderConfiguration{
			AuthorizationEndpoint: input.Body.AuthorizationEndpoint,
			TokenEndpoint:         input.Body.TokenEndpoint,
			JWKSEndpoint:          input.Body.JwksEndpoint,
			ClientID:              input.Body.ClientId,
			ClientSecret:          input.Body.ClientSecret,
		},
	})
	if err != nil {
		return nil, h.writeSSOError(ctx, "update SSO provider", err)
	}

	return nil, nil
}

// DeleteSsoProvider deletes an SSO provider.
func (h *Handler) DeleteSsoProvider(ctx context.Context, input *ssoProviderPathInput) (*struct{}, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.sso.Delete(ctx, user.ID, input.OrgID, input.ProviderID); err != nil {
		return nil, h.writeSSOError(ctx, "delete SSO provider", err)
	}

	return nil, nil
}

type discoverSsoInput struct {
	Issuer string `query:"issuer"`
}

type discoverSsoOutput struct {
	Body api.SsoDiscovery
}

// DiscoverSso discovers OIDC configuration from an issuer URL.
func (h *Handler) DiscoverSso(ctx context.Context, input *discoverSsoInput) (*discoverSsoOutput, error) {
	// Authentication limits the tenant-controlled outbound request surface to
	// signed-in accounts.
	if _, err := h.requireUser(ctx); err != nil {
		return nil, err
	}

	discovery, err := h.sso.Discover(ctx, input.Issuer)
	if err != nil {
		return nil, h.writeSSOError(ctx, "discover SSO provider", err)
	}

	return &discoverSsoOutput{Body: api.SsoDiscovery{
		Issuer:                discovery.Issuer,
		AuthorizationEndpoint: discovery.AuthorizationEndpoint,
		TokenEndpoint:         discovery.TokenEndpoint,
		JwksEndpoint:          discovery.JWKSEndpoint,
		UserInfoEndpoint:      discovery.UserInfoEndpoint,
		Scopes:                discovery.Scopes,
	}}, nil
}

func (h *Handler) writeSSOError(ctx context.Context, operation string, err error) error {
	var validationError *organizationsso.ValidationError
	switch {
	case errors.Is(err, organization.ErrMembershipNotFound), errors.Is(err, organization.ErrRoleNotAllowed):
		return huma.Error403Forbidden(httputil.MsgForbidden)
	case errors.As(err, &validationError):
		return huma.Error400BadRequest(validationError.Error())
	case errors.Is(err, organizationsso.ErrIssuerRequired):
		return huma.Error400BadRequest("issuer is required")
	case errors.Is(err, organizationsso.ErrInvalidIssuer):
		return huma.Error400BadRequest("invalid issuer URL")
	case errors.Is(err, organizationsso.ErrDiscoveryFailed):
		slog.ErrorContext(ctx, "SSO discovery failed", "error", err)
		return huma.Error502BadGateway("failed to fetch OIDC discovery document")
	default:
		slog.ErrorContext(ctx, "SSO operation failed", "operation", operation, "error", err)
		return huma.Error500InternalServerError("SSO operation failed")
	}
}
