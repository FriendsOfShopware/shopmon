package auth

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func adminOperation(id, method, path, summary string) huma.Operation {
	op := authOperation(id, method, path, summary, false)
	op.Tags = []string{"AuthAdmin"}
	return op
}

func registerAdmin(api huma.API, h *AuthHandler) {
	huma.Register(api, adminOperation("adminListUsers", http.MethodGet, "/auth/admin/users", "List all users (admin only)"), h.AdminListUsers)

	op := adminOperation("adminGetUserDetail", http.MethodGet, "/auth/admin/users/{userId}", "Get a single user with detail (admin only)")
	op.Errors = []int{http.StatusNotFound}
	huma.Register(api, op, h.AdminGetUserDetail)

	huma.Register(api, adminOperation("adminSetUserRole", http.MethodPatch, "/auth/admin/users/{userId}/role", "Set a user's role"), h.AdminSetUserRole)
	huma.Register(api, adminOperation("adminBanUser", http.MethodPost, "/auth/admin/users/{userId}/ban", "Ban a user"), h.AdminBanUser)
	huma.Register(api, adminOperation("adminUnbanUser", http.MethodPost, "/auth/admin/users/{userId}/unban", "Unban a user"), h.AdminUnbanUser)
	huma.Register(api, adminOperation("adminImpersonate", http.MethodPost, "/auth/admin/users/{userId}/impersonate", "Impersonate a user"), h.AdminImpersonate)
	huma.Register(api, adminOperation("adminStopImpersonating", http.MethodPost, "/auth/admin/stop-impersonating", "Stop impersonating a user"), h.AdminStopImpersonating)
}
