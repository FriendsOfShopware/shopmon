package handler

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func registerAccount(api huma.API, h *Handler) {
	huma.Register(api, huma.Operation{
		OperationID: "getAccountMe",
		Method:      http.MethodGet,
		Path:        "/account/me",
		Summary:     "Get current user profile",
		Tags:        []string{"Account"},
		Errors:      []int{http.StatusUnauthorized},
	}, h.GetAccountMe)

	huma.Register(api, huma.Operation{
		OperationID: "updateAccountMe",
		Method:      http.MethodPatch,
		Path:        "/account/me",
		Summary:     "Update current user preferences",
		Tags:        []string{"Account"},
		Errors:      []int{http.StatusBadRequest, http.StatusUnauthorized},
	}, h.UpdateAccountMe)

	huma.Register(api, huma.Operation{
		OperationID: "getAccountExtensions",
		Method:      http.MethodGet,
		Path:        "/account/extensions",
		Summary:     "Get aggregated extensions across all environments",
		Tags:        []string{"Account"},
		Errors:      []int{http.StatusUnauthorized},
	}, h.GetAccountExtensions)

	huma.Register(api, huma.Operation{
		OperationID: "getAccountExtension",
		Method:      http.MethodGet,
		Path:        "/account/extensions/{name}",
		Summary:     "Get a single aggregated extension by technical name",
		Tags:        []string{"Account"},
		Errors:      []int{http.StatusUnauthorized, http.StatusNotFound},
	}, h.GetAccountExtension)

	huma.Register(api, huma.Operation{
		OperationID: "getAccountOrganizations",
		Method:      http.MethodGet,
		Path:        "/account/organizations",
		Summary:     "Get organizations the user belongs to",
		Tags:        []string{"Account"},
		Errors:      []int{http.StatusUnauthorized},
	}, h.GetAccountOrganizations)

	huma.Register(api, huma.Operation{
		OperationID: "getAccountEnvironments",
		Method:      http.MethodGet,
		Path:        "/account/environments",
		Summary:     "Get all environments accessible to the user",
		Tags:        []string{"Account"},
		Errors:      []int{http.StatusUnauthorized},
	}, h.GetAccountEnvironments)

	huma.Register(api, huma.Operation{
		OperationID: "getAccountShops",
		Method:      http.MethodGet,
		Path:        "/account/shops",
		Summary:     "Get all shops accessible to the user",
		Tags:        []string{"Account"},
		Errors:      []int{http.StatusUnauthorized},
	}, h.GetAccountShops)

	huma.Register(api, huma.Operation{
		OperationID: "getAccountChangelogs",
		Method:      http.MethodGet,
		Path:        "/account/changelogs",
		Summary:     "Get changelogs across all environments",
		Tags:        []string{"Account"},
		Errors:      []int{http.StatusUnauthorized},
	}, h.GetAccountChangelogs)

	huma.Register(api, huma.Operation{
		OperationID: "getAccountSubscribedEnvironments",
		Method:      http.MethodGet,
		Path:        "/account/subscribed-environments",
		Summary:     "Get environments the user is subscribed to for notifications",
		Tags:        []string{"Account"},
		Errors:      []int{http.StatusUnauthorized},
	}, h.GetAccountSubscribedEnvironments)
}
