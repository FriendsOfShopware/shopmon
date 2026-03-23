package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/mail"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (h *AuthHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}

	orgID := uuid.New().String()
	// Auto-generate slug from org ID for backwards compatibility with the DB column
	slug := orgID

	_, err := h.queries.CreateOrganization(r.Context(), queries.CreateOrganizationParams{
		ID:   orgID,
		Name: req.Name,
		Slug: slug,
	})
	if err != nil {
		slog.Error("failed to create organization", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create organization")
		return
	}

	err = h.queries.CreateMember(r.Context(), queries.CreateMemberParams{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		UserID:         su.User.ID,
		Role:           "owner",
	})
	if err != nil {
		slog.Error("failed to add owner to organization", "error", err)
	}

	h.queries.SetActiveOrganization(r.Context(), queries.SetActiveOrganizationParams{
		ActiveOrganizationID: &orgID,
		Token:                httputil.ExtractToken(r),
	})

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":   orgID,
		"name": req.Name,
	})
}

func (h *AuthHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request, organizationId string) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var req struct {
		Name *string `json:"name"`
		Logo *string `json:"logo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	role, err := h.queries.GetMemberRole(r.Context(), queries.GetMemberRoleParams{
		OrganizationID: organizationId,
		UserID:         su.User.ID,
	})
	if err != nil || (role != "owner" && role != "admin") {
		httputil.WriteError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	org, err := h.queries.GetOrganizationByID(r.Context(), organizationId)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "organization not found")
		return
	}

	name := org.Name
	logo := org.Logo
	if req.Name != nil {
		name = *req.Name
	}
	if req.Logo != nil {
		logo = req.Logo
	}

	h.queries.UpdateOrganization(r.Context(), queries.UpdateOrganizationParams{
		Name: name,
		Slug: org.Slug,
		Logo: logo,
		ID:   organizationId,
	})

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandler) DeleteOrganization(w http.ResponseWriter, r *http.Request, organizationId string) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	role, err := h.queries.GetMemberRole(r.Context(), queries.GetMemberRoleParams{
		OrganizationID: organizationId,
		UserID:         su.User.ID,
	})
	if err != nil || role != "owner" {
		httputil.WriteError(w, http.StatusForbidden, "only owners can delete organizations")
		return
	}

	h.queries.DeleteOrganization(r.Context(), organizationId)
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandler) InviteMember(w http.ResponseWriter, r *http.Request, organizationId string) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Email == "" {
		httputil.WriteError(w, http.StatusBadRequest, "email is required")
		return
	}
	if req.Role == "" {
		req.Role = "member"
	}

	role, err := h.queries.GetMemberRole(r.Context(), queries.GetMemberRoleParams{
		OrganizationID: organizationId,
		UserID:         su.User.ID,
	})
	if err != nil || (role != "owner" && role != "admin") {
		httputil.WriteError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	invitationID := uuid.New().String()
	reqRole := req.Role
	_, err = h.queries.CreateInvitation(r.Context(), queries.CreateInvitationParams{
		ID:             invitationID,
		OrganizationID: organizationId,
		Email:          req.Email,
		Role:           &reqRole,
		ExpiresAt:      pgtype.Timestamp{Time: time.Now().Add(7 * 24 * time.Hour), Valid: true},
		InviterID:      su.User.ID,
	})
	if err != nil {
		slog.Error("failed to create invitation", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create invitation")
		return
	}

	// Send invitation email
	org, _ := h.queries.GetOrganizationByID(r.Context(), organizationId)
	acceptURL := h.cfg.FrontendURL + "/app/organizations/accept/" + invitationID
	rejectURL := h.cfg.FrontendURL + "/app/organizations/reject/" + invitationID

	orgName := org.Name

	h.mail.Send(req.Email,
		"You have been invited to join "+orgName+" at Shopmon",
		mail.BuildOrgInviteEmail(su.User.Name, orgName, acceptURL, rejectURL))

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"id": invitationID})
}

func (h *AuthHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request, invitationId string) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	invitation, err := h.queries.GetInvitationByID(r.Context(), invitationId)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "invitation not found or expired")
		return
	}

	// Verify the invitation is for this user
	if invitation.Email != su.User.Email {
		httputil.WriteError(w, http.StatusForbidden, "this invitation is not for your email address")
		return
	}

	memberRole := "member"
	if invitation.Role != nil {
		memberRole = *invitation.Role
	}

	if err := h.queries.CreateMember(r.Context(), queries.CreateMemberParams{
		ID:             uuid.New().String(),
		OrganizationID: invitation.OrganizationID,
		UserID:         su.User.ID,
		Role:           memberRole,
	}); err != nil {
		slog.Error("failed to create member on invitation accept", "error", err, "userID", su.User.ID, "orgID", invitation.OrganizationID)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to join organization")
		return
	}

	if err := h.queries.UpdateInvitationStatus(r.Context(), queries.UpdateInvitationStatusParams{
		Status: "accepted",
		ID:     invitationId,
	}); err != nil {
		slog.Warn("failed to update invitation status", "error", err, "invitationID", invitationId)
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandler) RejectInvitation(w http.ResponseWriter, r *http.Request, invitationId string) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	// Verify the invitation exists and is for this user
	invitation, err := h.queries.GetInvitationByID(r.Context(), invitationId)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "invitation not found or expired")
		return
	}

	if invitation.Email != su.User.Email {
		httputil.WriteError(w, http.StatusForbidden, "this invitation is not for your email address")
		return
	}

	if err = h.queries.UpdateInvitationStatus(r.Context(), queries.UpdateInvitationStatusParams{
		Status: "rejected",
		ID:     invitationId,
	}); err != nil {
		slog.Warn("failed to update invitation status to rejected", "error", err, "invitationID", invitationId)
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandler) RemoveMember(w http.ResponseWriter, r *http.Request, organizationId string, userId string) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	callerRole, err := h.queries.GetMemberRole(r.Context(), queries.GetMemberRoleParams{
		OrganizationID: organizationId,
		UserID:         su.User.ID,
	})
	if err != nil || (callerRole != "owner" && callerRole != "admin") {
		httputil.WriteError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	// Prevent removing members with equal or higher role
	targetRole, err := h.queries.GetMemberRole(r.Context(), queries.GetMemberRoleParams{
		OrganizationID: organizationId,
		UserID:         userId,
	})
	if err == nil {
		roleWeight := map[string]int{"member": 1, "admin": 2, "owner": 3}
		if roleWeight[targetRole] >= roleWeight[callerRole] {
			httputil.WriteError(w, http.StatusForbidden, "cannot remove a member with equal or higher role")
			return
		}
	}

	h.queries.DeleteMember(r.Context(), queries.DeleteMemberParams{
		OrganizationID: organizationId,
		UserID:         userId,
	})

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandler) LeaveOrganization(w http.ResponseWriter, r *http.Request, organizationId string) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	role, _ := h.queries.GetMemberRole(r.Context(), queries.GetMemberRoleParams{
		OrganizationID: organizationId,
		UserID:         su.User.ID,
	})
	if role == "owner" {
		ownerCount, _ := h.queries.CountOrgOwners(r.Context(), organizationId)
		if ownerCount <= 1 {
			httputil.WriteError(w, http.StatusBadRequest, "cannot leave as the only owner. Transfer ownership first.")
			return
		}
	}

	h.queries.DeleteMember(r.Context(), queries.DeleteMemberParams{
		OrganizationID: organizationId,
		UserID:         su.User.ID,
	})

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandler) SetMemberRole(w http.ResponseWriter, r *http.Request, organizationId string, userId string) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	// Validate role
	if req.Role != "owner" && req.Role != "admin" && req.Role != "member" {
		httputil.WriteError(w, http.StatusBadRequest, "role must be 'owner', 'admin', or 'member'")
		return
	}

	role, err := h.queries.GetMemberRole(r.Context(), queries.GetMemberRoleParams{
		OrganizationID: organizationId,
		UserID:         su.User.ID,
	})
	if err != nil || role != "owner" {
		httputil.WriteError(w, http.StatusForbidden, "only owners can change roles")
		return
	}

	h.queries.UpdateMemberRole(r.Context(), queries.UpdateMemberRoleParams{
		Role:           req.Role,
		OrganizationID: organizationId,
		UserID:         userId,
	})

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandler) ListOrganizationMembers(w http.ResponseWriter, r *http.Request, organizationId string) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	isMember, _ := h.queries.IsOrganizationMember(r.Context(), queries.IsOrganizationMemberParams{
		OrganizationID: organizationId,
		UserID:         su.User.ID,
	})
	if !isMember {
		httputil.WriteError(w, http.StatusForbidden, "not a member")
		return
	}

	members, err := h.queries.ListMembers(r.Context(), organizationId)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list members")
		return
	}

	result := make([]map[string]interface{}, 0, len(members))
	for _, m := range members {
		result = append(result, map[string]interface{}{
			"id":     m.ID,
			"userId": m.UserID,
			"role":   m.Role,
			"name":   m.UserName,
			"email":  m.UserEmail,
			"image":  m.UserImage,
		})
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *AuthHandler) ListOrganizationInvitations(w http.ResponseWriter, r *http.Request, organizationId string) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	role, err := h.queries.GetMemberRole(r.Context(), queries.GetMemberRoleParams{
		OrganizationID: organizationId,
		UserID:         su.User.ID,
	})
	if err != nil || (role != "owner" && role != "admin") {
		httputil.WriteError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	invitations, err := h.queries.ListInvitations(r.Context(), organizationId)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list invitations")
		return
	}

	result := make([]map[string]interface{}, 0, len(invitations))
	for _, i := range invitations {
		result = append(result, map[string]interface{}{
			"id":          i.ID,
			"email":       i.Email,
			"role":        i.Role,
			"status":      i.Status,
			"expiresAt":   i.ExpiresAt.Time,
			"inviterName": i.InviterName,
		})
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}
