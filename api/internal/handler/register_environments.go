package handler

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func registerEnvironments(api huma.API, h *Handler) {
	huma.Register(api, huma.Operation{
		OperationID: "getOrganizationEnvironments",
		Method:      http.MethodGet,
		Path:        "/organizations/{orgId}/environments",
		Summary:     "Get all environments in an organization",
		Tags:        []string{"Environments"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden},
	}, h.GetOrganizationEnvironments)

	huma.Register(api, huma.Operation{
		OperationID:   "createEnvironment",
		Method:        http.MethodPost,
		Path:          "/environments",
		Summary:       "Create a new environment",
		Tags:          []string{"Environments"},
		DefaultStatus: http.StatusCreated,
		Errors:        []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusUnprocessableEntity},
	}, h.CreateEnvironment)

	huma.Register(api, huma.Operation{
		OperationID: "getEnvironment",
		Method:      http.MethodGet,
		Path:        "/environments/{environmentId}",
		Summary:     "Get full environment details",
		Tags:        []string{"Environments"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
	}, h.GetEnvironment)

	huma.Register(api, huma.Operation{
		OperationID:   "updateEnvironment",
		Method:        http.MethodPatch,
		Path:          "/environments/{environmentId}",
		Summary:       "Update an environment",
		Tags:          []string{"Environments"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusUnprocessableEntity},
	}, h.UpdateEnvironment)

	huma.Register(api, huma.Operation{
		OperationID:   "deleteEnvironment",
		Method:        http.MethodDelete,
		Path:          "/environments/{environmentId}",
		Summary:       "Delete an environment",
		Tags:          []string{"Environments"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
	}, h.DeleteEnvironment)

	huma.Register(api, huma.Operation{
		OperationID:   "refreshEnvironment",
		Method:        http.MethodPost,
		Path:          "/environments/{environmentId}/refresh",
		Summary:       "Refresh environment data from Shopware API",
		Tags:          []string{"Environments"},
		DefaultStatus: http.StatusAccepted,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
	}, h.RefreshEnvironment)

	huma.Register(api, huma.Operation{
		OperationID:   "clearEnvironmentCache",
		Method:        http.MethodPost,
		Path:          "/environments/{environmentId}/clear-cache",
		Summary:       "Clear Shopware cache for the environment",
		Tags:          []string{"Environments"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusBadGateway},
	}, h.ClearEnvironmentCache)

	huma.Register(api, huma.Operation{
		OperationID:   "rescheduleTask",
		Method:        http.MethodPost,
		Path:          "/environments/{environmentId}/tasks/{taskId}/reschedule",
		Summary:       "Reschedule a scheduled task",
		Tags:          []string{"Environments"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusBadGateway},
	}, h.RescheduleTask)

	huma.Register(api, huma.Operation{
		OperationID:   "subscribeToEnvironment",
		Method:        http.MethodPost,
		Path:          "/environments/{environmentId}/subscribe",
		Summary:       "Subscribe to environment notifications",
		Tags:          []string{"Environments"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
	}, h.SubscribeToEnvironment)

	huma.Register(api, huma.Operation{
		OperationID:   "unsubscribeFromEnvironment",
		Method:        http.MethodDelete,
		Path:          "/environments/{environmentId}/subscribe",
		Summary:       "Unsubscribe from environment notifications",
		Tags:          []string{"Environments"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
	}, h.UnsubscribeFromEnvironment)

	huma.Register(api, huma.Operation{
		OperationID: "getEnvironmentStatusEvents",
		Method:      http.MethodGet,
		Path:        "/environments/{environmentId}/status-events",
		Summary:     "Get the status change history for an environment",
		Tags:        []string{"Environments"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
	}, h.GetEnvironmentStatusEvents)

	huma.Register(api, huma.Operation{
		OperationID: "getEnvironmentChangelogs",
		Method:      http.MethodGet,
		Path:        "/environments/{environmentId}/changelogs",
		Summary:     "List recorded changelog entries for an environment",
		Tags:        []string{"Environments"},
		Errors:      []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
	}, h.GetEnvironmentChangelogs)

	huma.Register(api, huma.Operation{
		OperationID:   "updateSitespeedSettings",
		Method:        http.MethodPut,
		Path:          "/environments/{environmentId}/sitespeed-settings",
		Summary:       "Update sitespeed settings for an environment",
		Tags:          []string{"Environments"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusUnprocessableEntity},
	}, h.UpdateSitespeedSettings)
}
