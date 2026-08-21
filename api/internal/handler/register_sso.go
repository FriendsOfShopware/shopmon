package handler

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func registerSSO(api huma.API, h *Handler) {
	huma.Register(api, huma.Operation{
		OperationID: "getSsoProviders",
		Method:      http.MethodGet,
		Path:        "/organizations/{orgId}/sso-providers",
		Summary:     "List SSO providers for an organization",
		Tags:        []string{"SSO"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden},
	}, h.GetSsoProviders)

	huma.Register(api, huma.Operation{
		OperationID: "discoverSso",
		Method:      http.MethodGet,
		Path:        "/sso/discover",
		Summary:     "Discover OIDC configuration from an issuer URL",
		Tags:        []string{"SSO"},
		Errors:      []int{http.StatusUnauthorized, http.StatusUnprocessableEntity},
	}, h.DiscoverSso)

	huma.Register(api, huma.Operation{
		OperationID:   "updateSsoProvider",
		Method:        http.MethodPut,
		Path:          "/organizations/{orgId}/sso-providers/{providerId}",
		Summary:       "Update an SSO provider",
		Tags:          []string{"SSO"},
		DefaultStatus: http.StatusNoContent,
		Errors: []int{
			http.StatusUnauthorized,
			http.StatusForbidden,
			http.StatusNotFound,
			http.StatusUnprocessableEntity,
		},
	}, h.UpdateSsoProvider)

	huma.Register(api, huma.Operation{
		OperationID:   "deleteSsoProvider",
		Method:        http.MethodDelete,
		Path:          "/organizations/{orgId}/sso-providers/{providerId}",
		Summary:       "Delete an SSO provider",
		Tags:          []string{"SSO"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
	}, h.DeleteSsoProvider)
}
