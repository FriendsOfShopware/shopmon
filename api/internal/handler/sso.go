package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/organization"
	organizationsso "github.com/friendsofshopware/shopmon/api/internal/organization/sso"
)

// GetSsoProviders lists SSO providers for an organization.
func (h *Handler) GetSsoProviders(w http.ResponseWriter, r *http.Request, orgID api.OrgId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	providers, err := h.sso.List(r.Context(), user.ID, orgID)
	if err != nil {
		h.writeSSOError(w, r, "list SSO providers", err)
		return
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

	httputil.WriteJSON(w, http.StatusOK, response)
}

// UpdateSsoProvider updates an SSO provider.
func (h *Handler) UpdateSsoProvider(w http.ResponseWriter, r *http.Request, orgID api.OrgId, providerID api.ProviderId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	var request api.UpdateSsoProviderRequest
	if err := httputil.DecodeBody(r, &request); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.sso.Update(r.Context(), organizationsso.UpdateCommand{
		UserID:         user.ID,
		OrganizationID: orgID,
		ProviderID:     providerID,
		Issuer:         request.Issuer,
		Domain:         request.Domain,
		Configuration: organizationsso.ProviderConfiguration{
			AuthorizationEndpoint: request.AuthorizationEndpoint,
			TokenEndpoint:         request.TokenEndpoint,
			JWKSEndpoint:          request.JwksEndpoint,
			ClientID:              request.ClientId,
			ClientSecret:          request.ClientSecret,
		},
	})
	if err != nil {
		h.writeSSOError(w, r, "update SSO provider", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteSsoProvider deletes an SSO provider.
func (h *Handler) DeleteSsoProvider(w http.ResponseWriter, r *http.Request, orgID api.OrgId, providerID api.ProviderId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	if err := h.sso.Delete(r.Context(), user.ID, orgID, providerID); err != nil {
		h.writeSSOError(w, r, "delete SSO provider", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DiscoverSso discovers OIDC configuration from an issuer URL.
func (h *Handler) DiscoverSso(w http.ResponseWriter, r *http.Request, params api.DiscoverSsoParams) {
	// Authentication limits the tenant-controlled outbound request surface to
	// signed-in accounts.
	if h.requireUser(w, r) == nil {
		return
	}

	discovery, err := h.sso.Discover(r.Context(), params.Issuer)
	if err != nil {
		h.writeSSOError(w, r, "discover SSO provider", err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, api.SsoDiscovery{
		Issuer:                discovery.Issuer,
		AuthorizationEndpoint: discovery.AuthorizationEndpoint,
		TokenEndpoint:         discovery.TokenEndpoint,
		JwksEndpoint:          discovery.JWKSEndpoint,
		UserInfoEndpoint:      discovery.UserInfoEndpoint,
		Scopes:                discovery.Scopes,
	})
}

func (h *Handler) writeSSOError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	var validationError *organizationsso.ValidationError
	switch {
	case errors.Is(err, organization.ErrMembershipNotFound), errors.Is(err, organization.ErrRoleNotAllowed):
		httputil.WriteForbidden(w)
	case errors.As(err, &validationError):
		httputil.WriteError(w, http.StatusBadRequest, validationError.Error())
	case errors.Is(err, organizationsso.ErrIssuerRequired):
		httputil.WriteError(w, http.StatusBadRequest, "issuer is required")
	case errors.Is(err, organizationsso.ErrInvalidIssuer):
		httputil.WriteError(w, http.StatusBadRequest, "invalid issuer URL")
	case errors.Is(err, organizationsso.ErrDiscoveryFailed):
		slog.ErrorContext(r.Context(), "SSO discovery failed", "error", err)
		httputil.WriteError(w, http.StatusBadGateway, "failed to fetch OIDC discovery document")
	default:
		slog.ErrorContext(r.Context(), "SSO operation failed", "operation", operation, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "SSO operation failed")
	}
}
