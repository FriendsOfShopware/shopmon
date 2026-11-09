package handler

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func registerDestinations(api huma.API, h *Handler) {
	const preferencesPath = "/organizations/{orgId}/notification-destinations/{destinationId}/preferences"
	huma.Register(api, huma.Operation{OperationID: "getDestinationNotificationPreferences", Method: http.MethodGet, Path: preferencesPath, Summary: "List shared destination event preferences", Tags: []string{"Notification destinations"}, Errors: []int{401, 403, 404}}, h.GetDestinationNotificationPreferences)
	huma.Register(api, huma.Operation{OperationID: "setDestinationNotificationPreference", Method: http.MethodPut, Path: preferencesPath, Summary: "Set a shared destination event preference", Tags: []string{"Notification destinations"}, DefaultStatus: http.StatusNoContent, Errors: []int{400, 401, 403, 404}}, h.SetDestinationNotificationPreference)
	huma.Register(api, huma.Operation{OperationID: "deleteDestinationNotificationPreference", Method: http.MethodDelete, Path: preferencesPath, Summary: "Inherit a shared destination event preference", Tags: []string{"Notification destinations"}, DefaultStatus: http.StatusNoContent, Errors: []int{400, 401, 403, 404}}, h.DeleteDestinationNotificationPreference)

	huma.Register(api, huma.Operation{
		OperationID: "listNotificationDestinations",
		Method:      http.MethodGet,
		Path:        "/organizations/{orgId}/notification-destinations",
		Summary:     "List organization notification destinations",
		Tags:        []string{"Notification destinations"},
		Errors:      []int{400, 401, 403, 404, 502},
	}, h.ListNotificationDestinations)
	huma.Register(api, huma.Operation{
		OperationID:   "createNotificationDestination",
		Method:        http.MethodPost,
		Path:          "/organizations/{orgId}/notification-destinations",
		Summary:       "Create organization notification destinations",
		Tags:          []string{"Notification destinations"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{400, 401, 403, 404, 502},
	}, h.CreateNotificationDestination)
	huma.Register(api, huma.Operation{
		OperationID:   "updateNotificationDestination",
		Method:        http.MethodPut,
		Path:          "/organizations/{orgId}/notification-destinations/{destinationId}",
		Summary:       "Update organization notification destinations",
		Tags:          []string{"Notification destinations"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{400, 401, 403, 404, 502},
	}, h.UpdateNotificationDestination)
	huma.Register(api, huma.Operation{
		OperationID:   "deleteNotificationDestination",
		Method:        http.MethodDelete,
		Path:          "/organizations/{orgId}/notification-destinations/{destinationId}",
		Summary:       "Delete organization notification destinations",
		Tags:          []string{"Notification destinations"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{400, 401, 403, 404, 502},
	}, h.DeleteNotificationDestination)
	huma.Register(api, huma.Operation{
		OperationID:   "testNotificationDestination",
		Method:        http.MethodPost,
		Path:          "/organizations/{orgId}/notification-destinations/{destinationId}/test",
		Summary:       "Test organization notification destinations",
		Tags:          []string{"Notification destinations"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{400, 401, 403, 404, 502},
	}, h.TestNotificationDestination)
	huma.Register(api, huma.Operation{
		OperationID: "listShopNotificationDestinations",
		Method:      http.MethodGet,
		Path:        "/organizations/{orgId}/shops/{shopId}/notification-destinations",
		Summary:     "List shop notification destinations",
		Tags:        []string{"Notification destinations"},
		Errors:      []int{400, 401, 403, 404, 502},
	}, h.ListShopNotificationDestinations)
	huma.Register(api, huma.Operation{
		OperationID:   "createShopNotificationDestination",
		Method:        http.MethodPost,
		Path:          "/organizations/{orgId}/shops/{shopId}/notification-destinations",
		Summary:       "Create shop notification destinations",
		Tags:          []string{"Notification destinations"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{400, 401, 403, 404, 502},
	}, h.CreateShopNotificationDestination)
	huma.Register(api, huma.Operation{
		OperationID:   "updateShopNotificationDestination",
		Method:        http.MethodPut,
		Path:          "/organizations/{orgId}/shops/{shopId}/notification-destinations/{destinationId}",
		Summary:       "Update shop notification destinations",
		Tags:          []string{"Notification destinations"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{400, 401, 403, 404, 502},
	}, h.UpdateShopNotificationDestination)
	huma.Register(api, huma.Operation{
		OperationID:   "deleteShopNotificationDestination",
		Method:        http.MethodDelete,
		Path:          "/organizations/{orgId}/shops/{shopId}/notification-destinations/{destinationId}",
		Summary:       "Delete shop notification destinations",
		Tags:          []string{"Notification destinations"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{400, 401, 403, 404, 502},
	}, h.DeleteShopNotificationDestination)
	huma.Register(api, huma.Operation{
		OperationID:   "testShopNotificationDestination",
		Method:        http.MethodPost,
		Path:          "/organizations/{orgId}/shops/{shopId}/notification-destinations/{destinationId}/test",
		Summary:       "Test shop notification destinations",
		Tags:          []string{"Notification destinations"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{400, 401, 403, 404, 502},
	}, h.TestShopNotificationDestination)
}
