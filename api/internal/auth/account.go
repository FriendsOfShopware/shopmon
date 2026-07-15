package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/authapi"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/identity"
	"github.com/friendsofshopware/shopmon/api/internal/organization"
	organizationsso "github.com/friendsofshopware/shopmon/api/internal/organization/sso"
)

// GetFullOrganization returns an organization with members and invitations.
func (h *AuthHandler) GetFullOrganization(w http.ResponseWriter, r *http.Request, params authapi.GetFullOrganizationParams) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	orgID := params.OrganizationId
	if orgID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "organizationId is required")
		return
	}

	full, err := h.organizations.Full(r.Context(), su.User.ID, orgID)
	if err != nil {
		h.writeOrganizationError(w, r, "get full organization", err)
		return
	}
	members := make([]organizationMemberResponse, 0, len(full.Members))
	for _, member := range full.Members {
		members = append(members, organizationMemberResponse{
			ID: member.ID, UserID: member.UserID, Role: string(member.Role),
			Name: member.Name, Email: member.Email, Image: member.Image,
		})
	}
	invitations := make([]organizationInvitationResponse, 0, len(full.Invitations))
	for _, invitation := range full.Invitations {
		var role *string
		if invitation.Role != nil {
			value := string(*invitation.Role)
			role = &value
		}
		invitations = append(invitations, organizationInvitationResponse{
			ID: invitation.ID, Email: invitation.Email, Role: role, Status: invitation.Status,
			ExpiresAt: invitation.ExpiresAt, InviterName: invitation.InviterName,
		})
	}

	httputil.WriteJSON(w, http.StatusOK, fullOrganizationResponse{
		ID:          full.Organization.ID,
		Name:        full.Organization.Name,
		Logo:        full.Organization.Logo,
		Members:     members,
		Invitations: invitations,
	})
}

// ListSessions returns all active sessions for the authenticated user.
func (h *AuthHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	sessions, err := h.accounts.ListSessions(r.Context(), su.User.ID)
	if err != nil {
		h.writeIdentityAccountError(w, r, "list sessions", err)
		return
	}
	response := make([]sessionResponse, 0, len(sessions))
	for _, session := range sessions {
		response = append(response, sessionResponse{
			ID: session.ID, ExpiresAt: session.ExpiresAt, CreatedAt: session.CreatedAt,
			IPAddress: session.IPAddress, UserAgent: session.UserAgent, ImpersonatedBy: session.ImpersonatedBy,
		})
	}
	httputil.WriteJSON(w, http.StatusOK, response)
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

	if err := h.accounts.RevokeSession(r.Context(), su.User.ID, req.SessionID); err != nil {
		h.writeIdentityAccountError(w, r, "revoke session", err)
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

	accounts, err := h.accounts.ListLinkedAccounts(r.Context(), su.User.ID)
	if err != nil {
		h.writeIdentityAccountError(w, r, "list linked accounts", err)
		return
	}
	response := make([]linkedAccountResponse, 0, len(accounts))
	for _, account := range accounts {
		response = append(response, linkedAccountResponse{
			ID: account.ID, Provider: account.Provider, AccountID: account.AccountID, CreatedAt: account.CreatedAt,
		})
	}
	httputil.WriteJSON(w, http.StatusOK, response)
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

	if err := h.accounts.UnlinkAccount(r.Context(), su.User.ID, req.ProviderID); err != nil {
		h.writeIdentityAccountError(w, r, "unlink account", err)
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

	if err := h.accounts.ChangeEmail(r.Context(), identity.ChangeEmailCommand{
		UserID: su.User.ID, Email: req.NewEmail, CurrentPassword: req.CurrentPassword,
	}); err != nil {
		h.writeIdentityAccountError(w, r, "change email", err)
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

	if err := h.accounts.UpdateName(r.Context(), su.User.ID, req.Name); err != nil {
		h.writeIdentityAccountError(w, r, "update user", err)
		return
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

	if err := h.accounts.ChangePassword(r.Context(), identity.ChangePasswordCommand{
		UserID: su.User.ID, CurrentPassword: req.CurrentPassword, NewPassword: req.NewPassword,
	}); err != nil {
		h.writeIdentityAccountError(w, r, "change password", err)
		return
	}

	h.recordAudit(r, su.User.ID, AuditActionPasswordChange, su.User.ID, "")

	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

// DeleteUser deletes the authenticated user and all their data.
func (h *AuthHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	if err := h.accounts.DeleteUser(r.Context(), su.User.ID); err != nil {
		h.writeIdentityAccountError(w, r, "delete user", err)
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

	passkeys, err := h.accounts.ListPasskeys(r.Context(), su.User.ID)
	if err != nil {
		h.writeIdentityAccountError(w, r, "list passkeys", err)
		return
	}
	response := make([]passkeyResponse, 0, len(passkeys))
	for _, passkey := range passkeys {
		response = append(response, passkeyResponse{
			ID: passkey.ID, Name: passkey.Name, DeviceType: passkey.DeviceType,
			BackedUp: passkey.BackedUp, CreatedAt: passkey.CreatedAt,
		})
	}
	httputil.WriteJSON(w, http.StatusOK, response)
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

	if err := h.accounts.DeletePasskey(r.Context(), su.User.ID, req.ID); err != nil {
		h.writeIdentityAccountError(w, r, "delete passkey", err)
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

	memberships, err := h.organizations.ListForUser(r.Context(), su.User.ID)
	if err != nil {
		h.writeOrganizationError(w, r, "list user organizations", err)
		return
	}
	response := make([]userOrganizationResponse, 0, len(memberships))
	for _, membership := range memberships {
		response = append(response, userOrganizationResponse{
			ID: membership.Organization.ID, Name: membership.Organization.Name,
			Logo: membership.Organization.Logo, CreatedAt: membership.Organization.CreatedAt,
			Role: string(membership.Role),
		})
	}
	httputil.WriteJSON(w, http.StatusOK, response)
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

	hasPermission, err := h.organizations.HasAdministrativePermission(r.Context(), su.User.ID, req.OrganizationID)
	if err != nil {
		h.writeOrganizationError(w, r, "check organization permission", err)
		return
	}
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

	if err := h.organizations.CancelInvitation(r.Context(), su.User.ID, req.InvitationID); err != nil {
		h.writeOrganizationError(w, r, "cancel organization invitation", err)
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

	if err := h.accounts.StopImpersonating(r.Context(), httputil.ExtractToken(r)); err != nil {
		h.writeIdentityAccountError(w, r, "stop impersonating", err)
		return
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

	clientSecret := req.ClientSecret
	if err := h.sso.Create(r.Context(), organizationsso.CreateCommand{
		UserID:         su.User.ID,
		OrganizationID: req.OrganizationID,
		Issuer:         req.Issuer,
		Domain:         req.Domain,
		Configuration: organizationsso.ProviderConfiguration{
			AuthorizationEndpoint: req.AuthorizationEndpoint,
			TokenEndpoint:         req.TokenEndpoint,
			JWKSEndpoint:          req.JwksEndpoint,
			ClientID:              req.ClientID,
			ClientSecret:          &clientSecret,
		},
	}); err != nil {
		h.writeAuthSSOError(w, r, "register SSO provider", err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

func (h *AuthHandler) writeAuthSSOError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	var validationError *organizationsso.ValidationError
	switch {
	case errors.Is(err, organization.ErrMembershipNotFound), errors.Is(err, organization.ErrRoleNotAllowed):
		httputil.WriteForbidden(w)
	case errors.As(err, &validationError):
		httputil.WriteError(w, http.StatusBadRequest, validationError.Error())
	default:
		slog.ErrorContext(r.Context(), "SSO operation failed", "operation", operation, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "SSO operation failed")
	}
}

func (h *AuthHandler) writeIdentityAccountError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	switch {
	case errors.Is(err, identity.ErrLastAuthenticationMethod):
		httputil.WriteError(w, http.StatusBadRequest, "cannot remove your last authentication method")
	case errors.Is(err, identity.ErrPasswordNotSet):
		httputil.WriteError(w, http.StatusBadRequest, "no password set for this account")
	case errors.Is(err, identity.ErrIncorrectPassword):
		httputil.WriteError(w, http.StatusUnauthorized, "current password is incorrect")
	case errors.Is(err, identity.ErrEmailInUse):
		httputil.WriteError(w, http.StatusConflict, "email already in use")
	case errors.Is(err, identity.ErrPasswordTooShort):
		httputil.WriteError(w, http.StatusBadRequest, "password must be at least 8 characters")
	default:
		slog.ErrorContext(r.Context(), "identity account operation failed", "operation", operation, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "account operation failed")
	}
}
