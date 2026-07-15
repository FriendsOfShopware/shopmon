package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/organization"
)

func (h *AuthHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	principal := h.requireAuth(w, r)
	if principal == nil {
		return
	}

	var request createOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if request.Name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}

	result, err := h.organizations.Create(r.Context(), organization.CreateCommand{
		UserID:       principal.User.ID,
		SessionToken: httputil.ExtractToken(r),
		Name:         request.Name,
	})
	if err != nil {
		h.writeOrganizationError(w, r, "create organization", err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, createOrganizationResponse{ID: result.ID, Name: result.Name})
}

func (h *AuthHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request, organizationID string) {
	principal := h.requireAuth(w, r)
	if principal == nil {
		return
	}

	var request updateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.organizations.Update(r.Context(), organization.UpdateCommand{
		UserID:         principal.User.ID,
		OrganizationID: organizationID,
		Name:           request.Name,
		Logo:           request.Logo,
	}); err != nil {
		h.writeOrganizationError(w, r, "update organization", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

func (h *AuthHandler) DeleteOrganization(w http.ResponseWriter, r *http.Request, organizationID string) {
	principal := h.requireAuth(w, r)
	if principal == nil {
		return
	}

	if err := h.organizations.Delete(r.Context(), principal.User.ID, organizationID); err != nil {
		h.writeOrganizationError(w, r, "delete organization", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

func (h *AuthHandler) InviteMember(w http.ResponseWriter, r *http.Request, organizationID string) {
	principal := h.requireAuth(w, r)
	if principal == nil {
		return
	}

	var request inviteMemberRequest
	if err := httputil.DecodeBody(r, &request); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if request.Email == "" {
		httputil.WriteError(w, http.StatusBadRequest, "email is required")
		return
	}
	if request.Role == "" {
		request.Role = string(organization.RoleMember)
	}

	invitationID, err := h.organizations.Invite(r.Context(), organization.InviteCommand{
		OrganizationID: organizationID,
		InviterID:      principal.User.ID,
		InviterName:    principal.User.Name,
		Email:          request.Email,
		Role:           request.Role,
	})
	if err != nil {
		h.writeOrganizationError(w, r, "invite organization member", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, idResponse{ID: invitationID})
}

func (h *AuthHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request, invitationID string) {
	principal := h.requireAuth(w, r)
	if principal == nil {
		return
	}

	if err := h.organizations.AcceptInvitation(r.Context(), invitationID, principal.User.ID, principal.User.Email); err != nil {
		h.writeOrganizationError(w, r, "accept organization invitation", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

func (h *AuthHandler) RejectInvitation(w http.ResponseWriter, r *http.Request, invitationID string) {
	principal := h.requireAuth(w, r)
	if principal == nil {
		return
	}

	if err := h.organizations.RejectInvitation(r.Context(), invitationID, principal.User.Email); err != nil {
		h.writeOrganizationError(w, r, "reject organization invitation", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

func (h *AuthHandler) RemoveMember(w http.ResponseWriter, r *http.Request, organizationID, userID string) {
	principal := h.requireAuth(w, r)
	if principal == nil {
		return
	}

	if err := h.organizations.RemoveMember(r.Context(), principal.User.ID, organizationID, userID); err != nil {
		h.writeOrganizationError(w, r, "remove organization member", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

func (h *AuthHandler) LeaveOrganization(w http.ResponseWriter, r *http.Request, organizationID string) {
	principal := h.requireAuth(w, r)
	if principal == nil {
		return
	}

	if err := h.organizations.Leave(r.Context(), principal.User.ID, organizationID); err != nil {
		h.writeOrganizationError(w, r, "leave organization", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

func (h *AuthHandler) SetMemberRole(w http.ResponseWriter, r *http.Request, organizationID, userID string) {
	principal := h.requireAuth(w, r)
	if principal == nil {
		return
	}

	var request setMemberRoleRequest
	if err := httputil.DecodeBody(r, &request); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.organizations.SetMemberRole(r.Context(), principal.User.ID, organizationID, userID, request.Role); err != nil {
		h.writeOrganizationError(w, r, "set organization member role", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

func (h *AuthHandler) ListOrganizationMembers(w http.ResponseWriter, r *http.Request, organizationID string) {
	principal := h.requireAuth(w, r)
	if principal == nil {
		return
	}

	members, err := h.organizations.ListMembers(r.Context(), principal.User.ID, organizationID)
	if err != nil {
		h.writeOrganizationError(w, r, "list organization members", err)
		return
	}
	response := make([]organizationMemberResponse, 0, len(members))
	for _, member := range members {
		response = append(response, organizationMemberResponse{
			ID: member.ID, UserID: member.UserID, Role: string(member.Role),
			Name: member.Name, Email: member.Email, Image: member.Image,
		})
	}
	httputil.WriteJSON(w, http.StatusOK, response)
}

func (h *AuthHandler) ListOrganizationInvitations(w http.ResponseWriter, r *http.Request, organizationID string) {
	principal := h.requireAuth(w, r)
	if principal == nil {
		return
	}

	invitations, err := h.organizations.ListInvitations(r.Context(), principal.User.ID, organizationID)
	if err != nil {
		h.writeOrganizationError(w, r, "list organization invitations", err)
		return
	}
	response := make([]organizationInvitationResponse, 0, len(invitations))
	for _, invitation := range invitations {
		var role *string
		if invitation.Role != nil {
			value := string(*invitation.Role)
			role = &value
		}
		response = append(response, organizationInvitationResponse{
			ID: invitation.ID, Email: invitation.Email, Role: role, Status: invitation.Status,
			ExpiresAt: invitation.ExpiresAt, InviterName: invitation.InviterName,
		})
	}
	httputil.WriteJSON(w, http.StatusOK, response)
}

func (h *AuthHandler) SetActiveOrganization(w http.ResponseWriter, r *http.Request) {
	principal := h.requireAuth(w, r)
	if principal == nil {
		return
	}

	var request struct {
		OrganizationID string `json:"organizationId"`
	}
	if err := httputil.DecodeBody(r, &request); err != nil || request.OrganizationID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "organizationId is required")
		return
	}
	if err := h.organizations.SetActive(r.Context(), principal.User.ID, httputil.ExtractToken(r), request.OrganizationID); err != nil {
		h.writeOrganizationError(w, r, "set active organization", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

func (h *AuthHandler) writeOrganizationError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	switch {
	case errors.Is(err, organization.ErrOrganizationNotFound):
		httputil.WriteError(w, http.StatusNotFound, "organization not found")
	case errors.Is(err, organization.ErrInvitationNotFound):
		httputil.WriteError(w, http.StatusNotFound, "invitation not found or expired")
	case errors.Is(err, organization.ErrInvitationEmailMismatch):
		httputil.WriteError(w, http.StatusForbidden, "this invitation is not for your email address")
	case errors.Is(err, organization.ErrInvalidRole):
		httputil.WriteError(w, http.StatusBadRequest, "role must be 'owner', 'admin', or 'member'")
	case errors.Is(err, organization.ErrCannotGrantRole):
		httputil.WriteError(w, http.StatusForbidden, "cannot invite a member with equal or higher role")
	case errors.Is(err, organization.ErrCannotRemoveRole):
		httputil.WriteError(w, http.StatusForbidden, "cannot remove a member with equal or higher role")
	case errors.Is(err, organization.ErrOwnerRequired):
		httputil.WriteError(w, http.StatusForbidden, "organization owner role required")
	case errors.Is(err, organization.ErrLastOwner):
		httputil.WriteError(w, http.StatusBadRequest, "cannot leave as the only owner. Transfer ownership first.")
	case errors.Is(err, organization.ErrMembershipNotFound), errors.Is(err, organization.ErrRoleNotAllowed):
		httputil.WriteForbidden(w)
	default:
		slog.ErrorContext(r.Context(), "organization operation failed", "operation", operation, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "organization operation failed")
	}
}
