package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"golang.org/x/crypto/bcrypt"
)

// GetFullOrganization returns the active organization with members and invitations.
func (h *AuthHandler) GetFullOrganization(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	orgID := h.requireActiveOrganization(w, su)
	if orgID == "" {
		return
	}

	if !h.requireOrgMembership(w, r, su.User.ID, orgID) {
		return
	}

	org, err := h.queries.GetOrganizationByID(r.Context(), orgID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "organization not found")
		return
	}

	members, err := h.queries.ListMembers(r.Context(), org.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list members")
		return
	}
	invitations, err := h.queries.ListInvitations(r.Context(), org.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list invitations")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, fullOrganizationResponse{
		ID:          org.ID,
		Name:        org.Name,
		Logo:        org.Logo,
		Members:     mapOrganizationMembers(members),
		Invitations: mapOrganizationInvitations(invitations),
	})
}

// ListSessions returns all active sessions for the authenticated user.
func (h *AuthHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	sessions, err := h.queries.ListUserSessions(r.Context(), su.User.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list sessions")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, mapUserSessions(sessions))
}

// RevokeSession revokes a specific session.
func (h *AuthHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var req revokeSessionRequest
	if err := httputil.DecodeBody(r, &req); err != nil || req.SessionID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "sessionId is required")
		return
	}

	if err := h.queries.DeleteSessionByID(r.Context(), queries.DeleteSessionByIDParams{
		ID:     req.SessionID,
		UserID: su.User.ID,
	}); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to revoke session")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

// ListAccounts returns linked auth providers for the user.
func (h *AuthHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	accounts, err := h.queries.ListUserAccounts(r.Context(), su.User.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list accounts")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, mapLinkedAccounts(accounts))
}

// UnlinkAccount removes a linked auth provider.
func (h *AuthHandler) UnlinkAccount(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var req unlinkAccountRequest
	if err := httputil.DecodeBody(r, &req); err != nil || req.ProviderID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "providerId is required")
		return
	}

	accounts, err := h.queries.ListUserAccounts(r.Context(), su.User.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list accounts")
		return
	}
	passkeys, err := h.queries.ListUserPasskeys(r.Context(), su.User.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list passkeys")
		return
	}
	if len(accounts)+len(passkeys) <= 1 {
		httputil.WriteError(w, http.StatusBadRequest, "cannot remove your last authentication method")
		return
	}

	if err := h.queries.DeleteAccountByProviderAndUser(r.Context(), queries.DeleteAccountByProviderAndUserParams{
		ProviderID: req.ProviderID,
		UserID:     su.User.ID,
	}); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to unlink account")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

// ChangeEmail updates the user's email.
func (h *AuthHandler) ChangeEmail(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var req changeEmailRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.NewEmail == "" {
		httputil.WriteError(w, http.StatusBadRequest, "email is required")
		return
	}

	account, err := h.queries.GetAccountByProviderAndUser(r.Context(), queries.GetAccountByProviderAndUserParams{
		ProviderID: "credential",
		UserID:     su.User.ID,
	})
	if err != nil || account.Password == nil {
		httputil.WriteError(w, http.StatusBadRequest, "no password set for this account")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*account.Password), []byte(req.CurrentPassword)); err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "current password is incorrect")
		return
	}

	_, err = h.queries.GetUserByEmail(r.Context(), req.NewEmail)
	if err == nil {
		httputil.WriteError(w, http.StatusConflict, "email already in use")
		return
	}

	if err := h.queries.UpdateUserEmail(r.Context(), queries.UpdateUserEmailParams{
		Email: req.NewEmail,
		ID:    su.User.ID,
	}); err != nil {
		slog.Error("failed to update user email", "error", err, "userID", su.User.ID)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update email")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

// UpdateUser updates the user's name.
func (h *AuthHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var req updateUserRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name != "" {
		if err := h.queries.UpdateUserName(r.Context(), queries.UpdateUserNameParams{
			Name: req.Name,
			ID:   su.User.ID,
		}); err != nil {
			slog.Error("failed to update user name", "error", err, "userID", su.User.ID)
			httputil.WriteError(w, http.StatusInternalServerError, "failed to update user")
			return
		}
	}
	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

// ChangePassword changes the user's password.
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var req changePasswordRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.NewPassword) < 8 {
		httputil.WriteError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	account, err := h.queries.GetAccountByProviderAndUser(r.Context(), queries.GetAccountByProviderAndUserParams{
		ProviderID: "credential",
		UserID:     su.User.ID,
	})
	if err != nil || account.Password == nil {
		httputil.WriteError(w, http.StatusBadRequest, "no password set for this account")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*account.Password), []byte(req.CurrentPassword)); err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "current password is incorrect")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}
	pw := string(hashed)
	if err := h.queries.UpdateUserPassword(r.Context(), queries.UpdateUserPasswordParams{
		Password: &pw,
		UserID:   su.User.ID,
	}); err != nil {
		slog.Error("failed to update user password", "error", err, "userID", su.User.ID)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update password")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

// DeleteUser deletes the authenticated user and all their data.
func (h *AuthHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	if err := h.queries.DeleteUser(r.Context(), su.User.ID); err != nil {
		slog.Error("failed to delete user", "error", err, "userID", su.User.ID)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete account")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

// LinkSocial initiates linking a social provider (returns redirect URL).
func (h *AuthHandler) LinkSocial(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	// Read and buffer the body so SignInSocial can read it too
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var req socialSignInRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// For now only GitHub
	if req.Provider != "github" {
		httputil.WriteError(w, http.StatusBadRequest, "unsupported provider")
		return
	}

	// Reconstruct body for SignInSocial
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	// Reuse the social sign-in flow -- it handles account linking automatically
	h.SignInSocial(w, r)
}

// ListUserPasskeys returns all passkeys for the authenticated user.
func (h *AuthHandler) ListUserPasskeys(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	passkeys, err := h.queries.ListUserPasskeys(r.Context(), su.User.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list passkeys")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, mapUserPasskeys(passkeys))
}

// DeletePasskey deletes a passkey.
func (h *AuthHandler) DeletePasskey(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var req deletePasskeyRequest
	if err := httputil.DecodeBody(r, &req); err != nil || req.ID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "id is required")
		return
	}

	if err := h.queries.DeletePasskey(r.Context(), queries.DeletePasskeyParams{
		ID:     req.ID,
		UserID: su.User.ID,
	}); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete passkey")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

// ListUserOrganizations returns organizations the user belongs to.
func (h *AuthHandler) ListUserOrganizations(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	orgs, err := h.queries.ListUserOrganizations(r.Context(), su.User.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list organizations")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, mapUserOrganizations(orgs))
}

// HasPermission checks if the user has a specific permission in an organization.
func (h *AuthHandler) HasPermission(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var req hasPermissionRequest
	if err := httputil.DecodeBody(r, &req); err != nil || req.OrganizationID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "organizationId is required")
		return
	}

	role, err := h.queries.GetMemberRole(r.Context(), queries.GetMemberRoleParams{
		OrganizationID: req.OrganizationID,
		UserID:         su.User.ID,
	})

	hasPermission := err == nil && (role == "owner" || role == "admin")
	httputil.WriteJSON(w, http.StatusOK, permissionResponse{Success: hasPermission})
}

// CancelInvitation cancels an invitation (org admin action).
func (h *AuthHandler) CancelInvitation(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var req cancelInvitationRequest
	if err := httputil.DecodeBody(r, &req); err != nil || req.InvitationID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "invitationId is required")
		return
	}

	// Get invitation to check org
	invitation, err := h.queries.GetInvitationByID(r.Context(), req.InvitationID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "invitation not found")
		return
	}

	if h.requireOrgRole(w, r, su.User.ID, invitation.OrganizationID, "owner", "admin") == "" {
		return
	}

	if err := h.queries.DeleteInvitationByID(r.Context(), req.InvitationID); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to cancel invitation")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

// AdminStopImpersonating ends an impersonation session.
func (h *AuthHandler) AdminStopImpersonating(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	// Delete the current impersonation session
	token := httputil.ExtractToken(r)
	if token != "" {
		if err := h.queries.DeleteSession(r.Context(), token); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to stop impersonation")
			return
		}
	}
	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

// RegisterSSOProvider registers a new SSO provider for an organization.
func (h *AuthHandler) RegisterSSOProvider(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var req registerSSOProviderRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if h.requireOrgRole(w, r, su.User.ID, req.OrganizationID, "owner", "admin") == "" {
		return
	}

	providerID := "sso-" + req.Domain

	oidcConfig, err := json.Marshal(map[string]string{
		"authorizationEndpoint": req.AuthorizationEndpoint,
		"tokenEndpoint":         req.TokenEndpoint,
		"jwksEndpoint":          req.JwksEndpoint,
		"clientId":              req.ClientID,
		"clientSecret":          req.ClientSecret,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to serialize OIDC config")
		return
	}
	oidcConfigStr := string(oidcConfig)

	if err := h.queries.CreateSSOProvider(r.Context(), queries.CreateSSOProviderParams{
		ID:             "sso-" + generateToken()[:16],
		Issuer:         req.Issuer,
		OidcConfig:     &oidcConfigStr,
		ProviderID:     providerID,
		OrganizationID: &req.OrganizationID,
		Domain:         req.Domain,
	}); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create SSO provider")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}
