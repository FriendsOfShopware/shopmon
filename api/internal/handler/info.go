package handler

import (
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/catalog"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

// GetInstanceConfig returns the feature configuration for this instance.
func (h *Handler) GetInstanceConfig(w http.ResponseWriter, r *http.Request) {
	httputil.WriteJSON(w, http.StatusOK, api.InstanceConfig{
		RegistrationEnabled:  h.features.RegistrationEnabled,
		GithubAuthEnabled:    h.features.GithubAuthEnabled,
		SitespeedEnabled:     h.features.SitespeedEnabled,
		PackageMirrorEnabled: h.features.PackageMirrorEnabled,
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

	result, err := h.admin.EcosystemStats(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to get ecosystem stats", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get ecosystem stats")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

// CheckExtensionCompatibility checks extension compatibility between Shopware versions.
func (h *Handler) CheckExtensionCompatibility(w http.ResponseWriter, r *http.Request) {
	var req api.ExtensionCompatibilityRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	extensions := make([]catalog.Extension, 0, len(req.Extensions))
	for _, ext := range req.Extensions {
		extensions = append(extensions, catalog.Extension{
			Name:    ext.Name,
			Version: ext.Version,
		})
	}

	result, err := h.catalog.CheckCompatibility(r.Context(), catalog.CompatibilityCommand{
		CurrentVersion: req.CurrentVersion,
		FutureVersion:  req.FutureVersion,
		Extensions:     extensions,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to check extension compatibility", "error", err)
		httputil.WriteError(w, http.StatusBadGateway, "failed to check extension compatibility")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result)
}

// GetShopwareVersions returns all known Shopware versions, newest first. Each
// version is wrapped in an object so the response can gain fields later without
// breaking clients. The data is served from the shopware_version table, which
// the worker refreshes hourly from the Shopware release changelog, so no
// external call is made at request time.
func (h *Handler) GetShopwareVersions(w http.ResponseWriter, r *http.Request) {
	names, err := h.catalog.ListShopwareVersions(r.Context())
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
