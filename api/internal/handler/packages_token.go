package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	neturl "net/url"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

// GetPackagesTokenConfiguration returns the packages token configuration.
func (h *Handler) GetPackagesTokenConfiguration(w http.ResponseWriter, r *http.Request) {
	configured := h.cfg.PackagesAPIURL != "" && h.cfg.PackagesAPIToken != ""

	var composerURL *string
	if configured {
		url := h.cfg.PackagesAPIURL + "/composer"
		composerURL = &url
	}

	httputil.WriteJSON(w, http.StatusOK, api.PackagesTokenConfiguration{
		Configured:  configured,
		ComposerUrl: composerURL,
	})
}

// packagesSource builds the "source" identifier used to scope tokens to a shop
// on the packages API. The "shopmon-project-" prefix is kept for compatibility
// with tokens registered by the previous implementation.
func packagesSource(shopId api.ShopId) string {
	return fmt.Sprintf("shopmon-project-%d", shopId)
}

// GetPackagesTokens lists packages tokens for a shop.
func (h *Handler) GetPackagesTokens(w http.ResponseWriter, r *http.Request, orgId api.OrgId, shopId api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}
	if !h.requireShopInOrganization(w, r, int32(shopId), orgId) {
		return
	}

	if h.cfg.PackagesAPIURL == "" {
		httputil.WriteError(w, http.StatusNotFound, "packages API not configured")
		return
	}

	// Proxy to packages API. Tokens are scoped per shop via the "source" field.
	url := fmt.Sprintf("%s/api/tokens?source=%s", h.cfg.PackagesAPIURL, neturl.QueryEscape(packagesSource(shopId)))
	resp, err := h.packagesRequest(r.Context(), "GET", url, nil)
	if err != nil {
		slog.Error("failed to fetch packages tokens", "error", err)
		httputil.WriteError(w, http.StatusBadGateway, "failed to get packages tokens")
		return
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to read response")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)
}

// CreatePackagesToken creates a new packages token.
func (h *Handler) CreatePackagesToken(w http.ResponseWriter, r *http.Request, orgId api.OrgId, shopId api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}
	if !h.requireShopInOrganization(w, r, int32(shopId), orgId) {
		return
	}

	if h.cfg.PackagesAPIURL == "" {
		httputil.WriteError(w, http.StatusNotFound, "packages API not configured")
		return
	}

	// Decode the incoming token so we can scope it to this shop via "source".
	var reqBody struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil || reqBody.Token == "" {
		httputil.WriteError(w, http.StatusBadRequest, "token is required")
		return
	}

	payload, err := json.Marshal(map[string]string{
		"token":  reqBody.Token,
		"source": packagesSource(shopId),
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to encode request")
		return
	}

	// Proxy to packages API.
	url := fmt.Sprintf("%s/api/tokens", h.cfg.PackagesAPIURL)
	resp, err := h.packagesRequest(r.Context(), "POST", url, bytes.NewReader(payload))
	if err != nil {
		slog.Error("failed to create packages token", "error", err)
		httputil.WriteError(w, http.StatusBadGateway, "failed to create packages token")
		return
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to read response")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)
}

// DeletePackagesToken deletes a packages token.
func (h *Handler) DeletePackagesToken(w http.ResponseWriter, r *http.Request, orgId api.OrgId, shopId api.ShopId, tokenId api.TokenId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}
	if !h.requireShopInOrganization(w, r, int32(shopId), orgId) {
		return
	}

	if h.cfg.PackagesAPIURL == "" {
		httputil.WriteError(w, http.StatusNotFound, "packages API not configured")
		return
	}

	owned, err := h.tokenBelongsToShop(r.Context(), shopId, tokenId)
	if err != nil {
		slog.Error("failed to verify packages token ownership", "error", err)
		httputil.WriteError(w, http.StatusBadGateway, "failed to delete packages token")
		return
	}
	if !owned {
		httputil.WriteError(w, http.StatusNotFound, "token not found")
		return
	}

	url := fmt.Sprintf("%s/api/tokens/%d", h.cfg.PackagesAPIURL, tokenId)
	resp, err := h.packagesRequest(r.Context(), "DELETE", url, nil)
	if err != nil {
		slog.Error("failed to delete packages token", "error", err)
		httputil.WriteError(w, http.StatusBadGateway, "failed to delete packages token")
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to read response")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)
}

// SyncPackagesToken syncs a packages token.
func (h *Handler) SyncPackagesToken(w http.ResponseWriter, r *http.Request, orgId api.OrgId, shopId api.ShopId, tokenId api.TokenId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}
	if !h.requireShopInOrganization(w, r, int32(shopId), orgId) {
		return
	}

	if h.cfg.PackagesAPIURL == "" {
		httputil.WriteError(w, http.StatusNotFound, "packages API not configured")
		return
	}

	owned, err := h.tokenBelongsToShop(r.Context(), shopId, tokenId)
	if err != nil {
		slog.Error("failed to verify packages token ownership", "error", err)
		httputil.WriteError(w, http.StatusBadGateway, "failed to sync packages token")
		return
	}
	if !owned {
		httputil.WriteError(w, http.StatusNotFound, "token not found")
		return
	}

	url := fmt.Sprintf("%s/api/tokens/%d/sync", h.cfg.PackagesAPIURL, tokenId)
	resp, err := h.packagesRequest(r.Context(), "POST", url, nil)
	if err != nil {
		slog.Error("failed to sync packages token", "error", err)
		httputil.WriteError(w, http.StatusBadGateway, "failed to sync packages token")
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to read response")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)
}

// tokenBelongsToShop checks that tokenId is scoped to shopId on the packages
// API. Tokens are isolated per shop via the "source" field, but the upstream
// delete/sync routes operate on the global token id, so we must verify ownership
// before proxying those mutations to prevent cross-shop access (IDOR).
func (h *Handler) tokenBelongsToShop(ctx context.Context, shopId api.ShopId, tokenId api.TokenId) (bool, error) {
	url := fmt.Sprintf("%s/api/tokens?source=%s", h.cfg.PackagesAPIURL, neturl.QueryEscape(packagesSource(shopId)))
	resp, err := h.packagesRequest(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("packages API returned status %d", resp.StatusCode)
	}

	var tokens []struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return false, err
	}

	for _, t := range tokens {
		if t.ID == int(tokenId) {
			return true, nil
		}
	}
	return false, nil
}

// packagesRequest creates an authenticated request to the packages API.
func (h *Handler) packagesRequest(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+h.cfg.PackagesAPIToken)
	if body != nil && (strings.ToUpper(method) == "POST" || strings.ToUpper(method) == "PUT" || strings.ToUpper(method) == "PATCH") {
		req.Header.Set("Content-Type", "application/json")
	}

	client := httputil.NewHTTPClient()
	return client.Do(req)
}
