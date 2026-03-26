package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/api"
)

// GetSsoProviders lists SSO providers for an organization.
func (h *Handler) GetSsoProviders(w http.ResponseWriter, r *http.Request, orgId api.OrgId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}

	orgIDPtr := orgId
	rows, err := h.queries.ListSSOProviders(r.Context(), &orgIDPtr)
	if err != nil {
		slog.Error("failed to list SSO providers", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get SSO providers")
		return
	}

	result := make([]api.SsoProvider, 0, len(rows))
	for _, row := range rows {
		provider := api.SsoProvider{
			Id:     row.ProviderID,
			Issuer: row.Issuer,
			Domain: row.Domain,
		}

		// Parse OIDC config if present
		if row.OidcConfig != nil {
			var oidcCfg struct {
				AuthorizationEndpoint string `json:"authorizationEndpoint"`
				TokenEndpoint         string `json:"tokenEndpoint"`
				JwksEndpoint          string `json:"jwksEndpoint"`
				ClientID              string `json:"clientId"`
			}
			if err := json.Unmarshal([]byte(*row.OidcConfig), &oidcCfg); err == nil {
				provider.AuthorizationEndpoint = oidcCfg.AuthorizationEndpoint
				provider.TokenEndpoint = oidcCfg.TokenEndpoint
				provider.JwksEndpoint = oidcCfg.JwksEndpoint
				provider.ClientId = oidcCfg.ClientID
			}
		}

		result = append(result, provider)
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// UpdateSsoProvider updates an SSO provider.
func (h *Handler) UpdateSsoProvider(w http.ResponseWriter, r *http.Request, orgId api.OrgId, providerId api.ProviderId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}

	var req api.UpdateSsoProviderRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Build the OIDC config JSON
	oidcConfig := map[string]interface{}{
		"authorizationEndpoint": req.AuthorizationEndpoint,
		"tokenEndpoint":         req.TokenEndpoint,
		"jwksEndpoint":          req.JwksEndpoint,
		"clientId":              req.ClientId,
	}
	if req.ClientSecret != nil {
		oidcConfig["clientSecret"] = *req.ClientSecret
	}

	oidcConfigJSON, err := json.Marshal(oidcConfig)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to serialize config")
		return
	}

	oidcConfigStr := string(oidcConfigJSON)
	orgIDPtr := orgId

	if err := h.queries.UpdateSSOProvider(r.Context(), queries.UpdateSSOProviderParams{
		Issuer:         req.Issuer,
		OidcConfig:     &oidcConfigStr,
		Domain:         req.Domain,
		ProviderID:     providerId,
		OrganizationID: &orgIDPtr,
	}); err != nil {
		slog.Error("failed to update SSO provider", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update SSO provider")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteSsoProvider deletes an SSO provider.
func (h *Handler) DeleteSsoProvider(w http.ResponseWriter, r *http.Request, orgId api.OrgId, providerId api.ProviderId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}

	orgIDPtr := orgId
	if err := h.queries.DeleteSSOProvider(r.Context(), queries.DeleteSSOProviderParams{
		ProviderID:     providerId,
		OrganizationID: &orgIDPtr,
	}); err != nil {
		slog.Error("failed to delete SSO provider", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete SSO provider")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DiscoverSso discovers OIDC configuration from an issuer URL.
func (h *Handler) DiscoverSso(w http.ResponseWriter, r *http.Request, params api.DiscoverSsoParams) {
	issuer := params.Issuer
	if issuer == "" {
		httputil.WriteError(w, http.StatusBadRequest, "issuer is required")
		return
	}

	// Validate URL scheme - only allow HTTPS (except localhost/127.0.0.1 for dev)
	parsed, err := url.Parse(issuer)
	isLocal := parsed != nil && (strings.HasPrefix(parsed.Host, "localhost") || strings.HasPrefix(parsed.Host, "127.0.0.1"))
	if err != nil || (parsed.Scheme != "https" && !isLocal) {
		httputil.WriteError(w, http.StatusBadRequest, "issuer must use HTTPS")
		return
	}

	// Fetch the OpenID Connect discovery document
	discoveryURL := issuer + "/.well-known/openid-configuration"
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, discoveryURL, nil)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create OIDC discovery request")
		return
	}

	resp, err := httputil.NewHTTPClient().Do(req)
	if err != nil {
		slog.Error("failed to fetch OIDC discovery", "url", discoveryURL, "error", err)
		httputil.WriteError(w, http.StatusBadGateway, "failed to fetch OIDC discovery document")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		httputil.WriteError(w, http.StatusBadGateway, "OIDC discovery endpoint returned non-200 status")
		return
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB limit
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to read discovery response")
		return
	}

	var discovery struct {
		Issuer                string   `json:"issuer"`
		AuthorizationEndpoint string   `json:"authorization_endpoint"`
		TokenEndpoint         string   `json:"token_endpoint"`
		JwksURI               string   `json:"jwks_uri"`
		UserinfoEndpoint      string   `json:"userinfo_endpoint"`
		ScopesSupported       []string `json:"scopes_supported"`
	}
	if err := json.Unmarshal(body, &discovery); err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "failed to parse OIDC discovery document")
		return
	}

	result := api.SsoDiscovery{
		Issuer:                discovery.Issuer,
		AuthorizationEndpoint: discovery.AuthorizationEndpoint,
		TokenEndpoint:         discovery.TokenEndpoint,
		JwksEndpoint:          discovery.JwksURI,
		UserInfoEndpoint:      discovery.UserinfoEndpoint,
		Scopes:                discovery.ScopesSupported,
	}

	if result.Scopes == nil {
		result.Scopes = []string{}
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}
