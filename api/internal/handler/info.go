package handler

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/catalog"
)

type getInstanceConfigOutput struct {
	Body api.InstanceConfig
}

// GetInstanceConfig returns the feature configuration for this instance.
func (h *Handler) GetInstanceConfig(_ context.Context, _ *struct{}) (*getInstanceConfigOutput, error) {
	return &getInstanceConfigOutput{Body: api.InstanceConfig{
		RegistrationEnabled:  h.features.RegistrationEnabled,
		GithubAuthEnabled:    h.features.GithubAuthEnabled,
		SitespeedEnabled:     h.features.SitespeedEnabled,
		PackageMirrorEnabled: h.features.PackageMirrorEnabled,
	}}, nil
}

type getEcosystemStatsOutput struct {
	Body api.EcosystemStats
}

// GetEcosystemStats returns aggregate ecosystem statistics (user and environment
// growth over time plus Shopware version distribution) for any authenticated user.
// The underlying data is aggregate-only and contains no per-user information.
func (h *Handler) GetEcosystemStats(ctx context.Context, _ *struct{}) (*getEcosystemStatsOutput, error) {
	if _, err := h.requireUser(ctx); err != nil {
		return nil, err
	}

	result, err := h.admin.EcosystemStats(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get ecosystem stats", "error", err)
		return nil, huma.Error500InternalServerError("failed to get ecosystem stats")
	}
	return &getEcosystemStatsOutput{Body: result}, nil
}

type checkExtensionCompatibilityInput struct {
	Body api.ExtensionCompatibilityRequest
}

type checkExtensionCompatibilityOutput struct {
	Body []api.ExtensionCompatibilityResult
}

// CheckExtensionCompatibility checks extension compatibility between Shopware versions.
func (h *Handler) CheckExtensionCompatibility(ctx context.Context, input *checkExtensionCompatibilityInput) (*checkExtensionCompatibilityOutput, error) {
	extensions := make([]catalog.Extension, 0, len(input.Body.Extensions))
	for _, ext := range input.Body.Extensions {
		extensions = append(extensions, catalog.Extension{
			Name:    ext.Name,
			Version: ext.Version,
		})
	}

	result, err := h.catalog.CheckCompatibility(ctx, catalog.CompatibilityCommand{
		CurrentVersion: input.Body.CurrentVersion,
		FutureVersion:  input.Body.FutureVersion,
		Extensions:     extensions,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to check extension compatibility", "error", err)
		return nil, huma.Error502BadGateway("failed to check extension compatibility")
	}

	var body []api.ExtensionCompatibilityResult
	if err := json.Unmarshal(result, &body); err != nil {
		slog.ErrorContext(ctx, "failed to decode extension compatibility response", "error", err)
		return nil, huma.Error502BadGateway("failed to check extension compatibility")
	}
	if body == nil {
		body = []api.ExtensionCompatibilityResult{}
	}
	return &checkExtensionCompatibilityOutput{Body: body}, nil
}

type getShopwareVersionsOutput struct {
	Body []api.ShopwareVersion
}

// GetShopwareVersions returns all known Shopware versions, newest first. Each
// version is wrapped in an object so the response can gain fields later without
// breaking clients. The data is served from the shopware_version table, which
// the worker refreshes hourly from the Shopware release changelog, so no
// external call is made at request time.
func (h *Handler) GetShopwareVersions(ctx context.Context, _ *struct{}) (*getShopwareVersionsOutput, error) {
	names, err := h.catalog.ListShopwareVersions(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list shopware versions", "error", err)
		return nil, huma.Error500InternalServerError("failed to load shopware versions")
	}

	versions := make([]api.ShopwareVersion, 0, len(names))
	for _, name := range names {
		versions = append(versions, api.ShopwareVersion{Name: name})
	}
	return &getShopwareVersionsOutput{Body: versions}, nil
}
