package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/authapi"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"golang.org/x/crypto/bcrypt"
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

	org, err := h.queries.GetOrganizationByID(r.Context(), orgID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "organization not found")
		return
	}

	isMember, _ := h.queries.IsOrgMember(r.Context(), queries.IsOrgMemberParams{
		OrganizationID: org.ID,
		UserID:         su.User.ID,
	})
	if !isMember {
		httputil.WriteError(w, http.StatusForbidden, "not a member of this organization")
		return
	}

	members, _ := h.queries.ListMembers(r.Context(), org.ID)
	invitations, _ := h.queries.ListInvitations(r.Context(), org.ID)

	memberList := make([]map[string]interface{}, 0, len(members))
	for _, m := range members {
		memberList = append(memberList, map[string]interface{}{
			"id": m.ID, "userId": m.UserID, "role": m.Role,
			"name": m.UserName, "email": m.UserEmail, "image": m.UserImage,
		})
	}

	inviteList := make([]map[string]interface{}, 0, len(invitations))
	for _, i := range invitations {
		inviteList = append(inviteList, map[string]interface{}{
			"id": i.ID, "email": i.Email, "role": i.Role,
			"status": i.Status, "expiresAt": i.ExpiresAt.Time, "inviterName": i.InviterName,
		})
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id": org.ID, "name": org.Name, "logo": org.Logo,
		"members": memberList, "invitations": inviteList,
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

	result := make([]map[string]interface{}, 0, len(sessions))
	for _, s := range sessions {
		result = append(result, map[string]interface{}{
			"id":             s.ID,
			"expiresAt":      s.ExpiresAt.Time,
			"createdAt":      s.CreatedAt.Time,
			"ipAddress":      s.IpAddress,
			"userAgent":      s.UserAgent,
			"impersonatedBy": s.ImpersonatedBy,
		})
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

// RevokeSession revokes a specific session.
func (h *AuthHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var req struct {
		SessionID string `json:"sessionId"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	h.queries.DeleteSessionByID(r.Context(), queries.DeleteSessionByIDParams{
		ID:     req.SessionID,
		UserID: su.User.ID,
	})
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
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

	result := make([]map[string]interface{}, 0, len(accounts))
	for _, a := range accounts {
		result = append(result, map[string]interface{}{
			"id":        a.ID,
			"provider":  a.ProviderID,
			"accountId": a.AccountID,
			"createdAt": a.CreatedAt.Time,
		})
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

// UnlinkAccount removes a linked auth provider.
func (h *AuthHandler) UnlinkAccount(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var req struct {
		ProviderId string `json:"providerId"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	accounts, _ := h.queries.ListUserAccounts(r.Context(), su.User.ID)
	passkeys, _ := h.queries.ListUserPasskeys(r.Context(), su.User.ID)
	if len(accounts)+len(passkeys) <= 1 {
		httputil.WriteError(w, http.StatusBadRequest, "cannot remove your last authentication method")
		return
	}

	h.queries.DeleteAccountByProviderAndUser(r.Context(), queries.DeleteAccountByProviderAndUserParams{
		ProviderID: req.ProviderId,
		UserID:     su.User.ID,
	})
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ChangeEmail updates the user's email.
func (h *AuthHandler) ChangeEmail(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var req struct {
		NewEmail        string `json:"newEmail"`
		CurrentPassword string `json:"currentPassword"`
	}
	json.NewDecoder(r.Body).Decode(&req)

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
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// UpdateUser updates the user's name.
func (h *AuthHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Name != "" {
		h.queries.UpdateUserName(r.Context(), queries.UpdateUserNameParams{
			Name: req.Name,
			ID:   su.User.ID,
		})
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ChangePassword changes the user's password.
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	json.NewDecoder(r.Body).Decode(&req)

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

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
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
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
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

	var req struct {
		Provider    string `json:"provider"`
		CallbackURL string `json:"callbackURL"`
	}
	json.Unmarshal(bodyBytes, &req)

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

	result := make([]map[string]interface{}, 0, len(passkeys))
	for _, p := range passkeys {
		result = append(result, map[string]interface{}{
			"id":         p.ID,
			"name":       p.Name,
			"deviceType": p.DeviceType,
			"backedUp":   p.BackedUp,
			"createdAt":  p.CreatedAt.Time,
		})
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

// DeletePasskey deletes a passkey.
func (h *AuthHandler) DeletePasskey(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var req struct {
		ID string `json:"id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	h.queries.DeletePasskey(r.Context(), queries.DeletePasskeyParams{
		ID:     req.ID,
		UserID: su.User.ID,
	})
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
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

	result := make([]map[string]interface{}, 0, len(orgs))
	for _, o := range orgs {
		result = append(result, map[string]interface{}{
			"id":        o.ID,
			"name":      o.Name,
			"logo":      o.Logo,
			"createdAt": o.CreatedAt.Time,
			"role":      o.Role,
		})
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

// HasPermission checks if the user has a specific permission in an organization.
func (h *AuthHandler) HasPermission(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var req struct {
		OrganizationID string `json:"organizationId"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	role, err := h.queries.GetMemberRole(r.Context(), queries.GetMemberRoleParams{
		OrganizationID: req.OrganizationID,
		UserID:         su.User.ID,
	})

	hasPermission := err == nil && (role == "owner" || role == "admin")
	httputil.WriteJSON(w, http.StatusOK, map[string]bool{"success": hasPermission})
}

// CancelInvitation cancels an invitation (org admin action).
func (h *AuthHandler) CancelInvitation(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var req struct {
		InvitationID string `json:"invitationId"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	// Get invitation to check org
	invitation, err := h.queries.GetInvitationByID(r.Context(), req.InvitationID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "invitation not found")
		return
	}

	role, err := h.queries.GetMemberRole(r.Context(), queries.GetMemberRoleParams{
		OrganizationID: invitation.OrganizationID,
		UserID:         su.User.ID,
	})
	if err != nil || (role != "owner" && role != "admin") {
		httputil.WriteError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	h.queries.DeleteInvitationByID(r.Context(), req.InvitationID)
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
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
		h.queries.DeleteSession(r.Context(), token)
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// RegisterSSOProvider registers a new SSO provider for an organization.
func (h *AuthHandler) RegisterSSOProvider(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var req struct {
		OrganizationID        string `json:"organizationId"`
		Domain                string `json:"domain"`
		Issuer                string `json:"issuer"`
		ClientID              string `json:"clientId"`
		ClientSecret          string `json:"clientSecret"`
		AuthorizationEndpoint string `json:"authorizationEndpoint"`
		TokenEndpoint         string `json:"tokenEndpoint"`
		JwksEndpoint          string `json:"jwksEndpoint"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	role, err := h.queries.GetMemberRole(r.Context(), queries.GetMemberRoleParams{
		OrganizationID: req.OrganizationID,
		UserID:         su.User.ID,
	})
	if err != nil || (role != "owner" && role != "admin") {
		httputil.WriteError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	providerID := "sso-" + req.Domain

	oidcConfig, _ := json.Marshal(map[string]string{
		"authorizationEndpoint": req.AuthorizationEndpoint,
		"tokenEndpoint":         req.TokenEndpoint,
		"jwksEndpoint":          req.JwksEndpoint,
		"clientId":              req.ClientID,
		"clientSecret":          req.ClientSecret,
	})
	oidcConfigStr := string(oidcConfig)

	h.queries.CreateSSOProvider(r.Context(), queries.CreateSSOProviderParams{
		ID:             "sso-" + generateToken()[:16],
		Issuer:         req.Issuer,
		OidcConfig:     &oidcConfigStr,
		ProviderID:     providerID,
		OrganizationID: &req.OrganizationID,
		Domain:         req.Domain,
	})

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
