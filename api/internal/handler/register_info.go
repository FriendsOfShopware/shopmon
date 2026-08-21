package handler

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func registerInfo(api huma.API, h *Handler) {
	huma.Register(api, huma.Operation{
		OperationID: "getInstanceConfig",
		Method:      http.MethodGet,
		Path:        "/info/config",
		Summary:     "Get instance feature configuration",
		Description: "Returns which features are enabled on this instance, based on server configuration.",
		Tags:        []string{"Info"},
		Security:    []map[string][]string{},
	}, h.GetInstanceConfig)

	huma.Register(api, huma.Operation{
		OperationID: "getEcosystemStats",
		Method:      http.MethodGet,
		Path:        "/info/ecosystem",
		Summary:     "Get public ecosystem statistics",
		Description: "Aggregate, non-identifying ecosystem statistics (user growth, environment growth and Shopware version distribution) visible to any authenticated user.",
		Tags:        []string{"Info"},
		Errors:      []int{http.StatusUnauthorized},
	}, h.GetEcosystemStats)

	huma.Register(api, huma.Operation{
		OperationID: "checkExtensionCompatibility",
		Method:      http.MethodPost,
		Path:        "/info/extension-compatibility",
		Summary:     "Check extension compatibility between Shopware versions",
		Tags:        []string{"Info"},
		Security:    []map[string][]string{},
		Errors:      []int{http.StatusUnprocessableEntity, http.StatusBadGateway},
	}, h.CheckExtensionCompatibility)

	huma.Register(api, huma.Operation{
		OperationID: "getShopwareVersions",
		Method:      http.MethodGet,
		Path:        "/info/shopware-versions",
		Summary:     "Get known Shopware versions",
		Description: "Returns all known Shopware versions, newest first. Served from a local cache that the worker refreshes hourly from the Shopware release changelog.",
		Tags:        []string{"Info"},
		Security:    []map[string][]string{},
	}, h.GetShopwareVersions)
}
