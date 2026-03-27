package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

// CheckExtensionCompatibility checks extension compatibility between Shopware versions.
func (h *Handler) CheckExtensionCompatibility(w http.ResponseWriter, r *http.Request) {
	var req api.ExtensionCompatibilityRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Build the request to Shopware's auto-update API
	type swExtension struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}

	type swRequest struct {
		CurrentVersion string        `json:"shopwareVersion"`
		FutureVersion  string        `json:"futureShopwareVersion"`
		Plugins        []swExtension `json:"plugins"`
	}

	plugins := make([]swExtension, 0, len(req.Extensions))
	for _, ext := range req.Extensions {
		plugins = append(plugins, swExtension{
			Name:    ext.Name,
			Version: ext.Version,
		})
	}

	swReq := swRequest{
		CurrentVersion: req.CurrentVersion,
		FutureVersion:  req.FutureVersion,
		Plugins:        plugins,
	}

	reqBody, err := json.Marshal(swReq)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to build request")
		return
	}

	// Proxy to Shopware API
	httpReq, err := http.NewRequestWithContext(r.Context(), "POST",
		h.cfg.ShopwareAPIURL+"/swplatform/autoupdate", bytes.NewReader(reqBody))
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create request")
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := httputil.NewHTTPClient()
	resp, err := client.Do(httpReq)
	if err != nil {
		slog.Error("failed to check extension compatibility", "error", err)
		httputil.WriteError(w, http.StatusBadGateway, "failed to check extension compatibility")
		return
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to read response")
		return
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("shopware API returned non-200", "status", resp.StatusCode, "body", string(body))
		httputil.WriteError(w, http.StatusBadGateway, "shopware API returned an error")
		return
	}

	var results []api.ExtensionCompatibilityResult
	if err := json.Unmarshal(body, &results); err != nil {
		// Pass through raw response if it doesn't match our type
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, results)
}

// GetShopwareVersions returns the latest Shopware version information.
func (h *Handler) GetShopwareVersions(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, h.cfg.ShopwareVersionsURL, nil)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create versions request")
		return
	}

	resp, err := httputil.NewHTTPClient().Do(req)
	if err != nil {
		slog.Error("failed to fetch shopware versions", "error", err)
		httputil.WriteError(w, http.StatusBadGateway, "failed to fetch shopware versions")
		return
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to read versions response")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
