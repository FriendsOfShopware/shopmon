package auth

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func registerAccount(api huma.API, h *AuthHandler) {
	huma.Register(api, authOperation("getFullOrganization", http.MethodGet, "/auth/get-full-organization", "Get full organization details by ID", false), h.GetFullOrganization)
	huma.Register(api, authOperation("listSessions", http.MethodGet, "/auth/list-sessions", "List active sessions", false), h.ListSessions)
	huma.Register(api, authOperation("revokeSession", http.MethodPost, "/auth/revoke-session", "Revoke a session", false), h.RevokeSession)
	huma.Register(api, authOperation("listAccounts", http.MethodGet, "/auth/list-accounts", "List linked auth providers", false), h.ListAccounts)
	huma.Register(api, authOperation("unlinkAccount", http.MethodPost, "/auth/unlink-account", "Unlink an auth provider", false), h.UnlinkAccount)
	huma.Register(api, authOperation("changeEmail", http.MethodPost, "/auth/change-email", "Change email address", false), h.ChangeEmail)
	huma.Register(api, authOperation("updateUser", http.MethodPost, "/auth/update-user", "Update user profile", false), h.UpdateUser)

	op := authOperation("changePassword", http.MethodPost, "/auth/change-password", "Change password", false)
	op.Errors = []int{http.StatusUnauthorized}
	huma.Register(api, op, h.ChangePassword)

	huma.Register(api, authOperation("deleteUser", http.MethodPost, "/auth/delete-user", "Delete user account", false), h.DeleteUser)
	huma.Register(api, authOperation("linkSocial", http.MethodPost, "/auth/link-social", "Link a social provider", false), h.LinkSocial)
	huma.Register(api, authOperation("listUserOrganizations", http.MethodGet, "/auth/list-organizations", "List user's organizations", false), h.ListUserOrganizations)
	huma.Register(api, authOperation("hasPermission", http.MethodPost, "/auth/has-permission", "Check organization permission", false), h.HasPermission)
	huma.Register(api, authOperation("cancelInvitation", http.MethodPost, "/auth/cancel-invitation", "Cancel a pending invitation", false), h.CancelInvitation)
	huma.Register(api, authOperation("listUserPasskeys", http.MethodGet, "/auth/passkey/list-user-passkeys", "List user's passkeys", false), h.ListUserPasskeys)
	huma.Register(api, authOperation("deletePasskey", http.MethodPost, "/auth/passkey/delete-passkey", "Delete a passkey", false), h.DeletePasskey)
	huma.Register(api, authOperation("registerSSOProvider", http.MethodPost, "/auth/sso/register", "Register an SSO provider", false), h.RegisterSSOProvider)
}
