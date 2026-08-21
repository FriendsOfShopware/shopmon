package handler

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func registerShops(api huma.API, h *Handler) {
	huma.Register(api, huma.Operation{
		OperationID: "getOrganizationShops",
		Method:      http.MethodGet,
		Path:        "/organizations/{orgId}/shops",
		Summary:     "Get all shops in an organization",
		Tags:        []string{"Shops"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden},
	}, h.GetOrganizationShops)

	huma.Register(api, huma.Operation{
		OperationID:   "createShop",
		Method:        http.MethodPost,
		Path:          "/organizations/{orgId}/shops",
		Summary:       "Create a new shop",
		Tags:          []string{"Shops"},
		DefaultStatus: http.StatusCreated,
		Errors:        []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusUnprocessableEntity},
	}, h.CreateShop)

	huma.Register(api, huma.Operation{
		OperationID:   "updateShop",
		Method:        http.MethodPatch,
		Path:          "/organizations/{orgId}/shops/{shopId}",
		Summary:       "Update a shop",
		Tags:          []string{"Shops"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusUnprocessableEntity},
	}, h.UpdateShop)

	huma.Register(api, huma.Operation{
		OperationID:   "deleteShop",
		Method:        http.MethodDelete,
		Path:          "/organizations/{orgId}/shops/{shopId}",
		Summary:       "Delete a shop",
		Tags:          []string{"Shops"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusConflict},
	}, h.DeleteShop)
}
