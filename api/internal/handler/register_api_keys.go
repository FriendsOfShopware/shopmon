package handler

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func registerAPIKeys(humaAPI huma.API, h *Handler) {
	huma.Register(humaAPI, huma.Operation{
		OperationID: "getApiKeyScopes",
		Method:      http.MethodGet,
		Path:        "/api-key-scopes",
		Summary:     "Get available API key scopes",
		Tags:        []string{"ApiKeys"},
		Security:    []map[string][]string{},
	}, h.GetApiKeyScopes)

	huma.Register(humaAPI, huma.Operation{
		OperationID: "getApiKeys",
		Method:      http.MethodGet,
		Path:        "/organizations/{orgId}/shops/{shopId}/api-keys",
		Summary:     "List API keys for a shop",
		Tags:        []string{"ApiKeys"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden},
	}, h.GetApiKeys)

	huma.Register(humaAPI, huma.Operation{
		OperationID:   "createApiKey",
		Method:        http.MethodPost,
		Path:          "/organizations/{orgId}/shops/{shopId}/api-keys",
		Summary:       "Create a new API key",
		Tags:          []string{"ApiKeys"},
		DefaultStatus: http.StatusCreated,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusUnprocessableEntity},
	}, h.CreateApiKey)

	huma.Register(humaAPI, huma.Operation{
		OperationID:   "deleteApiKey",
		Method:        http.MethodDelete,
		Path:          "/organizations/{orgId}/shops/{shopId}/api-keys/{keyId}",
		Summary:       "Delete an API key",
		Tags:          []string{"ApiKeys"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
	}, h.DeleteApiKey)
}
