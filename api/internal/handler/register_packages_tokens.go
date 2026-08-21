package handler

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func registerPackagesTokens(humaAPI huma.API, h *Handler) {
	huma.Register(humaAPI, huma.Operation{
		OperationID: "getPackagesTokenConfiguration",
		Method:      http.MethodGet,
		Path:        "/packages-token/configuration",
		Summary:     "Get packages token configuration",
		Tags:        []string{"PackagesToken"},
		Errors:      []int{http.StatusUnauthorized},
	}, h.GetPackagesTokenConfiguration)

	huma.Register(humaAPI, huma.Operation{
		OperationID: "getPackagesTokens",
		Method:      http.MethodGet,
		Path:        "/organizations/{orgId}/shops/{shopId}/packages-tokens",
		Summary:     "List packages tokens for a shop",
		Tags:        []string{"PackagesToken"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden},
	}, h.GetPackagesTokens)

	huma.Register(humaAPI, huma.Operation{
		OperationID:   "createPackagesToken",
		Method:        http.MethodPost,
		Path:          "/organizations/{orgId}/shops/{shopId}/packages-tokens",
		Summary:       "Create a new packages token",
		Tags:          []string{"PackagesToken"},
		DefaultStatus: http.StatusCreated,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusUnprocessableEntity},
	}, h.CreatePackagesToken)

	huma.Register(humaAPI, huma.Operation{
		OperationID:   "deletePackagesToken",
		Method:        http.MethodDelete,
		Path:          "/organizations/{orgId}/shops/{shopId}/packages-tokens/{tokenId}",
		Summary:       "Delete a packages token",
		Tags:          []string{"PackagesToken"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
	}, h.DeletePackagesToken)

	huma.Register(humaAPI, huma.Operation{
		OperationID:   "syncPackagesToken",
		Method:        http.MethodPost,
		Path:          "/organizations/{orgId}/shops/{shopId}/packages-tokens/{tokenId}/sync",
		Summary:       "Sync a packages token",
		Tags:          []string{"PackagesToken"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
	}, h.SyncPackagesToken)
}
