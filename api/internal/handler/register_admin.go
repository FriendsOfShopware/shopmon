package handler

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func registerAdmin(api huma.API, h *Handler) {
	huma.Register(api, huma.Operation{
		OperationID: "adminGetOrganizations",
		Method:      http.MethodGet,
		Path:        "/admin/organizations",
		Summary:     "List all organizations (admin only)",
		Tags:        []string{"Admin"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden},
	}, h.AdminGetOrganizations)

	huma.Register(api, huma.Operation{
		OperationID: "adminGetEnvironments",
		Method:      http.MethodGet,
		Path:        "/admin/environments",
		Summary:     "List all environments (admin only)",
		Tags:        []string{"Admin"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden},
	}, h.AdminGetEnvironments)

	huma.Register(api, huma.Operation{
		OperationID: "adminGetStats",
		Method:      http.MethodGet,
		Path:        "/admin/stats",
		Summary:     "Get admin dashboard statistics",
		Tags:        []string{"Admin"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden},
	}, h.AdminGetStats)

	huma.Register(api, huma.Operation{
		OperationID: "adminGetGrowth",
		Method:      http.MethodGet,
		Path:        "/admin/growth",
		Summary:     "Get growth data over time",
		Tags:        []string{"Admin"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden},
	}, h.AdminGetGrowth)

	huma.Register(api, huma.Operation{
		OperationID: "adminGetOrganizationDetail",
		Method:      http.MethodGet,
		Path:        "/admin/organizations/{orgId}",
		Summary:     "Get an organization with detail (admin only)",
		Tags:        []string{"Admin"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
	}, h.AdminGetOrganizationDetail)

	huma.Register(api, huma.Operation{
		OperationID: "adminGetEnvironmentDetail",
		Method:      http.MethodGet,
		Path:        "/admin/environments/{envId}",
		Summary:     "Get an environment with detail (admin only)",
		Tags:        []string{"Admin"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
	}, h.AdminGetEnvironmentDetail)

	huma.Register(api, huma.Operation{
		OperationID: "adminGetAuditLog",
		Method:      http.MethodGet,
		Path:        "/admin/audit-log",
		Summary:     "List audit log entries (admin only)",
		Tags:        []string{"Admin"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden},
	}, h.AdminGetAuditLog)

	huma.Register(api, huma.Operation{
		OperationID: "adminGetRecentActivity",
		Method:      http.MethodGet,
		Path:        "/admin/recent-activity",
		Summary:     "Get recent user and environment activity",
		Tags:        []string{"Admin"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden},
	}, h.AdminGetRecentActivity)

	huma.Register(api, huma.Operation{
		OperationID: "adminGetShopwareVersions",
		Method:      http.MethodGet,
		Path:        "/admin/shopware-versions",
		Summary:     "Get Shopware version distribution across environments",
		Tags:        []string{"Admin"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden},
	}, h.AdminGetShopwareVersions)
}
