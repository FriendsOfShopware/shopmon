package handler

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func registerDeployments(humaAPI huma.API, h *Handler) {
	huma.Register(humaAPI, huma.Operation{
		OperationID: "getDeployments",
		Method:      http.MethodGet,
		Path:        "/environments/{environmentId}/deployments",
		Summary:     "List deployments for an environment",
		Tags:        []string{"Deployments"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden},
	}, h.GetDeployments)

	huma.Register(humaAPI, huma.Operation{
		OperationID: "getDeployment",
		Method:      http.MethodGet,
		Path:        "/environments/{environmentId}/deployments/{deploymentId}",
		Summary:     "Get deployment details with output",
		Tags:        []string{"Deployments"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
	}, h.GetDeployment)

	huma.Register(humaAPI, huma.Operation{
		OperationID:   "deleteDeployment",
		Method:        http.MethodDelete,
		Path:          "/environments/{environmentId}/deployments/{deploymentId}",
		Summary:       "Delete a deployment",
		Tags:          []string{"Deployments"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
	}, h.DeleteDeployment)

	huma.Register(humaAPI, huma.Operation{
		OperationID:   "createCliDeployment",
		Method:        http.MethodPost,
		Path:          "/cli/deployments",
		Summary:       "Create a deployment via CLI",
		Tags:          []string{"CLI"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
		Errors:        []int{http.StatusUnauthorized, http.StatusUnprocessableEntity},
	}, h.CreateCliDeployment)
}
