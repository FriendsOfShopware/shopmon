package handler

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
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

// GetPackagesTokens lists packages tokens for a shop.
func (h *Handler) GetPackagesTokens(w http.ResponseWriter, r *http.Request, shopId api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	orgId := h.requireActiveOrganization(w, r)
	if orgId == "" {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}

	if h.cfg.PackagesAPIURL == "" {
		httputil.WriteError(w, http.StatusNotFound, "packages API not configured")
		return
	}

	// Proxy to packages API
	url := fmt.Sprintf("%s/api/shops/%d/tokens", h.cfg.PackagesAPIURL, shopId)
	resp, err := h.packagesRequest("GET", url, nil)
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
func (h *Handler) CreatePackagesToken(w http.ResponseWriter, r *http.Request, shopId api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	orgId := h.requireActiveOrganization(w, r)
	if orgId == "" {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}

	if h.cfg.PackagesAPIURL == "" {
		httputil.WriteError(w, http.StatusNotFound, "packages API not configured")
		return
	}

	// Proxy the request body to packages API
	url := fmt.Sprintf("%s/api/shops/%d/tokens", h.cfg.PackagesAPIURL, shopId)
	resp, err := h.packagesRequest("POST", url, r.Body)
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
func (h *Handler) DeletePackagesToken(w http.ResponseWriter, r *http.Request, shopId api.ShopId, tokenId api.TokenId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	orgId := h.requireActiveOrganization(w, r)
	if orgId == "" {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}

	if h.cfg.PackagesAPIURL == "" {
		httputil.WriteError(w, http.StatusNotFound, "packages API not configured")
		return
	}

	url := fmt.Sprintf("%s/api/shops/%d/tokens/%d", h.cfg.PackagesAPIURL, shopId, tokenId)
	resp, err := h.packagesRequest("DELETE", url, nil)
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
func (h *Handler) SyncPackagesToken(w http.ResponseWriter, r *http.Request, shopId api.ShopId, tokenId api.TokenId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	orgId := h.requireActiveOrganization(w, r)
	if orgId == "" {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}

	if h.cfg.PackagesAPIURL == "" {
		httputil.WriteError(w, http.StatusNotFound, "packages API not configured")
		return
	}

	url := fmt.Sprintf("%s/api/shops/%d/tokens/%d/sync", h.cfg.PackagesAPIURL, shopId, tokenId)
	resp, err := h.packagesRequest("POST", url, nil)
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

// packagesRequest creates an authenticated request to the packages API.
func (h *Handler) packagesRequest(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
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
