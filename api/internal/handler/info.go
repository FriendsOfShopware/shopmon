package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/shopwareaccount"
)

// shopwareVersionsCacheKey is the Redis key for the cached versions response.
const shopwareVersionsCacheKey = "shopmon:shopware-versions"

// shopwareVersionsCacheTTL is how long the versions response is cached.
const shopwareVersionsCacheTTL = time.Hour

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

// ListSecurityAdvisories returns the catalog of imported security advisories for
// any authenticated user.
func (h *Handler) ListSecurityAdvisories(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	rows, err := h.queries.ListSecurityAdvisories(r.Context())
	if err != nil {
		slog.Error("failed to list security advisories", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list security advisories")
		return
	}

	advisories := make([]api.SecurityAdvisory, 0, len(rows))
	for _, row := range rows {
		advisories = append(advisories, api.SecurityAdvisory{
			AdvisoryId:       row.AdvisoryID,
			Origin:           row.Origin,
			PackageName:      row.PackageName,
			Title:            row.Title,
			Link:             row.Link,
			Cve:              row.Cve,
			AffectedVersions: row.AffectedVersions,
			Severity:         row.Severity,
			Source:           row.SourceName,
			ReportedAt:       pgtimeToTimePtr(row.ReportedAt),
		})
	}

	httputil.WriteJSON(w, http.StatusOK, advisories)
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

// GetShopwareVersions returns the latest Shopware version information.
// The upstream response is cached in Redis to avoid hitting it on every request.
func (h *Handler) GetShopwareVersions(w http.ResponseWriter, r *http.Request) {
	if h.redis != nil {
		if cached, err := h.redis.Get(r.Context(), shopwareVersionsCacheKey).Bytes(); err == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(cached)
			return
		}
	}

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

	if h.redis != nil && resp.StatusCode == http.StatusOK {
		if err := h.redis.Set(r.Context(), shopwareVersionsCacheKey, body, shopwareVersionsCacheTTL).Err(); err != nil {
			slog.Warn("failed to cache shopware versions", "error", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
