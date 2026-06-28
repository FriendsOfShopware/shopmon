package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/shopwareaccount"
)

// GetInstanceConfig returns the feature configuration for this instance.
func (h *Handler) GetInstanceConfig(w http.ResponseWriter, r *http.Request) {
	httputil.WriteJSON(w, http.StatusOK, api.InstanceConfig{
		RegistrationEnabled:  !h.cfg.DisableRegistration,
		GithubAuthEnabled:    h.cfg.GithubClientID != "" && h.cfg.GithubClientSecret != "",
		SitespeedEnabled:     h.cfg.SitespeedEndpoint != "",
		PackageMirrorEnabled: h.cfg.PackagesAPIURL != "" && h.cfg.PackagesAPIToken != "",
	})
}

// GetEcosystemStats returns aggregate ecosystem statistics (user and environment
// growth over time plus Shopware version distribution) for any authenticated user.
// The underlying data is aggregate-only and contains no per-user information.
func (h *Handler) GetEcosystemStats(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	environmentRows, err := h.queries.AdminGetGrowthEnvironments(r.Context())
	if err != nil {
		slog.Error("failed to get environment growth", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get ecosystem stats")
		return
	}

	userRows, err := h.queries.AdminGetGrowthUsers(r.Context())
	if err != nil {
		slog.Error("failed to get user growth", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get ecosystem stats")
		return
	}

	versionRows, err := h.queries.AdminGetShopwareVersions(r.Context())
	if err != nil {
		slog.Error("failed to get shopware versions", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get ecosystem stats")
		return
	}

	environmentGrowth := make([]api.GrowthDataPoint, 0, len(environmentRows))
	for _, row := range environmentRows {
		environmentGrowth = append(environmentGrowth, api.GrowthDataPoint{
			Month: row.Month,
			Count: int(row.Count),
		})
	}

	userGrowth := make([]api.GrowthDataPoint, 0, len(userRows))
	for _, row := range userRows {
		userGrowth = append(userGrowth, api.GrowthDataPoint{
			Month: row.Month,
			Count: int(row.Count),
		})
	}

	versions := make([]api.ShopwareVersionCount, 0, len(versionRows))
	for _, row := range versionRows {
		versions = append(versions, api.ShopwareVersionCount{
			Version: row.Version,
			Count:   int(row.Count),
		})
	}

	httputil.WriteJSON(w, http.StatusOK, api.EcosystemStats{
		Growth: api.AdminGrowth{
			Environments: environmentGrowth,
			Users:        userGrowth,
		},
		ShopwareVersions: versions,
	})
}

// CheckExtensionCompatibility checks extension compatibility between Shopware versions.
func (h *Handler) CheckExtensionCompatibility(w http.ResponseWriter, r *http.Request) {
	var req api.ExtensionCompatibilityRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	extensions := make([]shopwareaccount.CompatibilityExtension, 0, len(req.Extensions))
	for _, ext := range req.Extensions {
		extensions = append(extensions, shopwareaccount.CompatibilityExtension{
			Name:    ext.Name,
			Version: ext.Version,
		})
	}

	client := shopwareaccount.NewClient(h.cfg.ShopwareAPIURL, httputil.NewHTTPClient())
	resp, err := client.CheckExtensionCompatibility(r.Context(), req.CurrentVersion, req.FutureVersion, extensions)
	if err != nil {
		slog.Error("failed to check extension compatibility", "error", err)
		httputil.WriteError(w, http.StatusBadGateway, "failed to check extension compatibility")
		return
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("shopware account API returned non-200", "status", resp.StatusCode, "body", string(resp.Body))
		httputil.WriteError(w, http.StatusBadGateway, "shopware account API returned an error")
		return
	}

	var results []api.ExtensionCompatibilityResult
	if err := json.Unmarshal(resp.Body, &results); err != nil {
		// Pass through raw response if it doesn't match our type
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(resp.Body)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, results)
}

// GetShopwareVersions returns all known Shopware versions, newest first. Each
// version is wrapped in an object so the response can grow extra fields (release
// date, EOL, …) without breaking clients. The data is served from the
// shopware_version table, which the worker keeps up to date by crawling the
// Shopware release changelog hourly, so no external call is made at request time.
func (h *Handler) GetShopwareVersions(w http.ResponseWriter, r *http.Request) {
	names, err := h.queries.ListShopwareVersions(r.Context())
	if err != nil {
		slog.Error("failed to list shopware versions", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load shopware versions")
		return
	}

	versions := make([]api.ShopwareVersion, 0, len(names))
	for _, name := range names {
		versions = append(versions, api.ShopwareVersion{Name: name})
	}

	httputil.WriteJSON(w, http.StatusOK, versions)
}
