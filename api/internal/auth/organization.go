package auth

import (
	"context"
	"errors"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/organization"
)

type statusOutput struct {
	Body statusResponse
}

type urlOutput struct {
	Body urlResponse
}

func statusOK() *statusOutput {
	return &statusOutput{Body: newStatusResponse()}
}

type createOrganizationInput struct {
	Body createOrganizationRequest
}

type createOrganizationOutput struct {
	Body createOrganizationResponse
}

func (h *AuthHandler) CreateOrganization(ctx context.Context, input *createOrganizationInput) (*createOrganizationOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	if input.Body.Name == "" {
		return nil, huma.Error400BadRequest("name is required")
	}

	result, err := h.organizations.Create(ctx, organization.CreateCommand{
		UserID:       principal.User.ID,
		SessionToken: httputil.ExtractToken(requestFromContext(ctx)),
		Name:         input.Body.Name,
	})
	if err != nil {
		return nil, h.organizationError(ctx, "create organization", err)
	}

	return &createOrganizationOutput{Body: createOrganizationResponse{ID: result.ID, Name: result.Name}}, nil
}

type organizationIDInput struct {
	OrganizationID string `path:"organizationId"`
}

type updateOrganizationInput struct {
	OrganizationID string `path:"organizationId"`
	Body           updateOrganizationRequest
}

func (h *AuthHandler) UpdateOrganization(ctx context.Context, input *updateOrganizationInput) (*statusOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.organizations.Update(ctx, organization.UpdateCommand{
		UserID:         principal.User.ID,
		OrganizationID: input.OrganizationID,
		Name:           input.Body.Name,
		Logo:           input.Body.Logo,
	}); err != nil {
		return nil, h.organizationError(ctx, "update organization", err)
	}
	return statusOK(), nil
}

func (h *AuthHandler) DeleteOrganization(ctx context.Context, input *organizationIDInput) (*statusOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.organizations.Delete(ctx, principal.User.ID, input.OrganizationID); err != nil {
		return nil, h.organizationError(ctx, "delete organization", err)
	}
	return statusOK(), nil
}

type inviteMemberInput struct {
	OrganizationID string `path:"organizationId"`
	Body           inviteMemberRequest
}

type idOutput struct {
	Body idResponse
}

func (h *AuthHandler) InviteMember(ctx context.Context, input *inviteMemberInput) (*idOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	if input.Body.Email == "" {
		return nil, huma.Error400BadRequest("email is required")
	}
	role := input.Body.Role
	if role == "" {
		role = string(organization.RoleMember)
	}

	invitationID, err := h.organizations.Invite(ctx, organization.InviteCommand{
		OrganizationID: input.OrganizationID,
		InviterID:      principal.User.ID,
		InviterName:    principal.User.Name,
		Email:          input.Body.Email,
		Role:           role,
	})
	if err != nil {
		return nil, h.organizationError(ctx, "invite organization member", err)
	}
	return &idOutput{Body: idResponse{ID: invitationID}}, nil
}

type invitationIDInput struct {
	InvitationID string `path:"invitationId"`
}

func (h *AuthHandler) AcceptInvitation(ctx context.Context, input *invitationIDInput) (*statusOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.organizations.AcceptInvitation(ctx, input.InvitationID, principal.User.ID, principal.User.Email); err != nil {
		return nil, h.organizationError(ctx, "accept organization invitation", err)
	}
	return statusOK(), nil
}

func (h *AuthHandler) RejectInvitation(ctx context.Context, input *invitationIDInput) (*statusOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.organizations.RejectInvitation(ctx, input.InvitationID, principal.User.Email); err != nil {
		return nil, h.organizationError(ctx, "reject organization invitation", err)
	}
	return statusOK(), nil
}

type organizationMemberInput struct {
	OrganizationID string `path:"organizationId"`
	UserID         string `path:"userId"`
}

func (h *AuthHandler) RemoveMember(ctx context.Context, input *organizationMemberInput) (*statusOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.organizations.RemoveMember(ctx, principal.User.ID, input.OrganizationID, input.UserID); err != nil {
		return nil, h.organizationError(ctx, "remove organization member", err)
	}
	return statusOK(), nil
}

func (h *AuthHandler) LeaveOrganization(ctx context.Context, input *organizationIDInput) (*statusOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.organizations.Leave(ctx, principal.User.ID, input.OrganizationID); err != nil {
		return nil, h.organizationError(ctx, "leave organization", err)
	}
	return statusOK(), nil
}

type setMemberRoleInput struct {
	OrganizationID string `path:"organizationId"`
	UserID         string `path:"userId"`
	Body           setMemberRoleRequest
}

func (h *AuthHandler) SetMemberRole(ctx context.Context, input *setMemberRoleInput) (*statusOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.organizations.SetMemberRole(ctx, principal.User.ID, input.OrganizationID, input.UserID, input.Body.Role); err != nil {
		return nil, h.organizationError(ctx, "set organization member role", err)
	}
	return statusOK(), nil
}

type listOrganizationMembersOutput struct {
	Body []organizationMemberResponse
}

func (h *AuthHandler) ListOrganizationMembers(ctx context.Context, input *organizationIDInput) (*listOrganizationMembersOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	members, err := h.organizations.ListMembers(ctx, principal.User.ID, input.OrganizationID)
	if err != nil {
		return nil, h.organizationError(ctx, "list organization members", err)
	}
	return &listOrganizationMembersOutput{Body: mapOrganizationMembers(members)}, nil
}

type listOrganizationInvitationsOutput struct {
	Body []organizationInvitationResponse
}

func (h *AuthHandler) ListOrganizationInvitations(ctx context.Context, input *organizationIDInput) (*listOrganizationInvitationsOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	invitations, err := h.organizations.ListInvitations(ctx, principal.User.ID, input.OrganizationID)
	if err != nil {
		return nil, h.organizationError(ctx, "list organization invitations", err)
	}
	return &listOrganizationInvitationsOutput{Body: mapOrganizationInvitations(invitations)}, nil
}

type setActiveOrganizationInput struct {
	Body struct {
		OrganizationID string `json:"organizationId"`
	}
}

func (h *AuthHandler) SetActiveOrganization(ctx context.Context, input *setActiveOrganizationInput) (*statusOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	if input.Body.OrganizationID == "" {
		return nil, huma.Error400BadRequest("organizationId is required")
	}
	if err := h.organizations.SetActive(ctx, principal.User.ID, httputil.ExtractToken(requestFromContext(ctx)), input.Body.OrganizationID); err != nil {
		return nil, h.organizationError(ctx, "set active organization", err)
	}
	return statusOK(), nil
}

func mapOrganizationMembers(members []organization.Member) []organizationMemberResponse {
	response := make([]organizationMemberResponse, 0, len(members))
	for _, member := range members {
		response = append(response, organizationMemberResponse{
			ID: member.ID, UserID: member.UserID, Role: string(member.Role),
			Name: member.Name, Email: member.Email, Image: member.Image,
		})
	}
	return response
}

func mapOrganizationInvitations(invitations []organization.Invitation) []organizationInvitationResponse {
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
	return response
}

func (h *AuthHandler) organizationError(ctx context.Context, operation string, err error) error {
	switch {
	case errors.Is(err, organization.ErrOrganizationNotFound):
		return huma.Error404NotFound("organization not found")
	case errors.Is(err, organization.ErrInvitationNotFound):
		return huma.Error404NotFound("invitation not found or expired")
	case errors.Is(err, organization.ErrInvitationEmailMismatch):
		return huma.Error403Forbidden("this invitation is not for your email address")
	case errors.Is(err, organization.ErrInvalidRole):
		return huma.Error400BadRequest("role must be 'owner', 'admin', or 'member'")
	case errors.Is(err, organization.ErrCannotGrantRole):
		return huma.Error403Forbidden("cannot invite a member with equal or higher role")
	case errors.Is(err, organization.ErrCannotRemoveRole):
		return huma.Error403Forbidden("cannot remove a member with equal or higher role")
	case errors.Is(err, organization.ErrOwnerRequired):
		return huma.Error403Forbidden("organization owner role required")
	case errors.Is(err, organization.ErrLastOwner):
		return huma.Error400BadRequest("cannot leave as the only owner. Transfer ownership first.")
	case errors.Is(err, organization.ErrMembershipNotFound), errors.Is(err, organization.ErrRoleNotAllowed):
		return huma.Error403Forbidden(httputil.MsgForbidden)
	default:
		slog.ErrorContext(ctx, "organization operation failed", "operation", operation, "error", err)
		return huma.Error500InternalServerError("organization operation failed")
	}
}
