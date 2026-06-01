package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
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
	if !h.requireOrgRole(w, r, user, orgId, "owner", "admin") {
		return
	}

	var req api.UpdateSsoProviderRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate the tenant-supplied OIDC endpoints so the login flow never issues
	// server-side requests to internal/private targets.
	for _, endpoint := range []string{req.Issuer, req.AuthorizationEndpoint, req.TokenEndpoint, req.JwksEndpoint} {
		if err := httputil.ValidateHTTPSEndpoint(endpoint); err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid OIDC endpoint: "+err.Error())
			return
		}
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
	if !h.requireOrgRole(w, r, user, orgId, "owner", "admin") {
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
	// Only authenticated users may trigger server-side discovery requests; this
	// limits the SSRF surface to logged-in accounts.
	if h.requireUser(w, r) == nil {
		return
	}

	issuer := params.Issuer
	if issuer == "" {
		httputil.WriteError(w, http.StatusBadRequest, "issuer is required")
		return
	}

	parsed, err := url.Parse(issuer)
	if err != nil || parsed.Host == "" {
		httputil.WriteError(w, http.StatusBadRequest, "invalid issuer URL")
		return
	}

	// A loopback issuer (local-dev IdP) may use plain HTTP. Match the hostname
	// exactly to avoid bypasses like "localhost.attacker.com".
	isLoopback := isLoopbackHost(parsed.Hostname())

	// Every other issuer must use HTTPS.
	if parsed.Scheme != "https" && !isLoopback {
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

	// Use an SSRF-safe client that refuses private/loopback/link-local targets
	// and bounds the request with a timeout. The plain client is only used for
	// the explicitly-allowed loopback dev case.
	client := httputil.NewSSRFSafeHTTPClient(15 * time.Second)
	if isLoopback {
		client = httputil.NewHTTPClient(httputil.WithTimeout(15 * time.Second))
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("failed to fetch OIDC discovery", "url", discoveryURL, "error", err)
		httputil.WriteError(w, http.StatusBadGateway, "failed to fetch OIDC discovery document")
		return
	}
	defer func() { _ = resp.Body.Close() }()

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

// isLoopbackHost reports whether host is exactly "localhost" or a loopback IP
// literal. It deliberately does NOT match suffixes like "localhost.evil.com".
func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}
