package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

func (h *AuthHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var req createOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}

	orgID, err := h.createOrganizationFlow(r.Context(), su.User.ID, httputil.ExtractToken(r), req.Name)
	if err != nil {
		slog.Error("failed to create organization", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create organization")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, createOrganizationResponse{
		ID:   orgID,
		Name: req.Name,
	})
}

func (h *AuthHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	organizationId := h.requireActiveOrganization(w, su)
	if organizationId == "" {
		return
	}

	var req updateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if h.requireOrgRole(w, r, su.User.ID, organizationId, "owner", "admin") == "" {
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

	if err := h.queries.UpdateOrganization(r.Context(), queries.UpdateOrganizationParams{
		Name: name,
		Slug: org.Slug,
		Logo: logo,
		ID:   organizationId,
	}); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update organization")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

func (h *AuthHandler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	organizationId := h.requireActiveOrganization(w, su)
	if organizationId == "" {
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

	if err := h.queries.DeleteOrganization(r.Context(), organizationId); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete organization")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

func (h *AuthHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	organizationId := h.requireActiveOrganization(w, su)
	if organizationId == "" {
		return
	}

	var req inviteMemberRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" {
		httputil.WriteError(w, http.StatusBadRequest, "email is required")
		return
	}
	if req.Role == "" {
		req.Role = "member"
	}

	if h.requireOrgRole(w, r, su.User.ID, organizationId, "owner", "admin") == "" {
		return
	}

	invitationID, err := h.inviteMemberFlow(r.Context(), inviteMemberCommand{
		OrganizationID: organizationId,
		InviterID:      su.User.ID,
		InviterName:    su.User.Name,
		Email:          req.Email,
		Role:           req.Role,
	})
	if err != nil {
		slog.Error("failed to create invitation", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create invitation")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, idResponse{ID: invitationID})
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

	if err := h.acceptInvitationFlow(r.Context(), invitationId, invitation.OrganizationID, su.User.ID, invitation.Role); err != nil {
		slog.Error("failed to accept invitation", "error", err, "userID", su.User.ID, "orgID", invitation.OrganizationID)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to join organization")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
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

	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

func (h *AuthHandler) RemoveMember(w http.ResponseWriter, r *http.Request, userId string) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	organizationId := h.requireActiveOrganization(w, su)
	if organizationId == "" {
		return
	}

	callerRole := h.requireOrgRole(w, r, su.User.ID, organizationId, "owner", "admin")
	if callerRole == "" {
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

	if err := h.queries.DeleteMember(r.Context(), queries.DeleteMemberParams{
		OrganizationID: organizationId,
		UserID:         userId,
	}); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to remove member")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

func (h *AuthHandler) LeaveOrganization(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	organizationId := h.requireActiveOrganization(w, su)
	if organizationId == "" {
		return
	}

	role, err := h.queries.GetMemberRole(r.Context(), queries.GetMemberRoleParams{
		OrganizationID: organizationId,
		UserID:         su.User.ID,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusForbidden, "not a member")
		return
	}
	if role == "owner" {
		ownerCount, err := h.queries.CountOrgOwners(r.Context(), organizationId)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to leave organization")
			return
		}
		if ownerCount <= 1 {
			httputil.WriteError(w, http.StatusBadRequest, "cannot leave as the only owner. Transfer ownership first.")
			return
		}
	}

	if err := h.queries.DeleteMember(r.Context(), queries.DeleteMemberParams{
		OrganizationID: organizationId,
		UserID:         su.User.ID,
	}); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to leave organization")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

func (h *AuthHandler) SetMemberRole(w http.ResponseWriter, r *http.Request, userId string) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	organizationId := h.requireActiveOrganization(w, su)
	if organizationId == "" {
		return
	}

	var req setMemberRoleRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

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

	if err := h.queries.UpdateMemberRole(r.Context(), queries.UpdateMemberRoleParams{
		Role:           req.Role,
		OrganizationID: organizationId,
		UserID:         userId,
	}); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update member role")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

func (h *AuthHandler) ListOrganizationMembers(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	organizationId := h.requireActiveOrganization(w, su)
	if organizationId == "" {
		return
	}

	if !h.requireOrgMembership(w, r, su.User.ID, organizationId) {
		return
	}

	members, err := h.queries.ListMembers(r.Context(), organizationId)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list members")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, mapOrganizationMembers(members))
}

func (h *AuthHandler) ListOrganizationInvitations(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	organizationId := h.requireActiveOrganization(w, su)
	if organizationId == "" {
		return
	}

	if h.requireOrgRole(w, r, su.User.ID, organizationId, "owner", "admin") == "" {
		return
	}

	invitations, err := h.queries.ListInvitations(r.Context(), organizationId)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list invitations")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, mapOrganizationInvitations(invitations))
}

func (h *AuthHandler) SetActiveOrganization(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var req struct {
		OrganizationID string `json:"organizationId"`
	}
	if err := httputil.DecodeBody(r, &req); err != nil || req.OrganizationID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "organizationId is required")
		return
	}

	if !h.requireOrgMembership(w, r, su.User.ID, req.OrganizationID) {
		return
	}

	token := httputil.ExtractToken(r)
	if err := h.queries.SetActiveOrganization(r.Context(), queries.SetActiveOrganizationParams{
		ActiveOrganizationID: &req.OrganizationID,
		Token:                token,
	}); err != nil {
		slog.Error("failed to set active organization", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to set active organization")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}
