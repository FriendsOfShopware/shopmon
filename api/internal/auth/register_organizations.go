package auth

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func organizationOperation(id, method, path, summary string) huma.Operation {
	op := authOperation(id, method, path, summary, false)
	op.Tags = []string{"Organizations"}
	return op
}

func registerOrganizations(api huma.API, h *AuthHandler) {
	huma.Register(api, organizationOperation("createOrganization", http.MethodPost, "/auth/organizations", "Create an organization"), h.CreateOrganization)

	huma.Register(api, organizationOperation("updateOrganization", http.MethodPatch, "/auth/organizations/{organizationId}", "Update an organization"), h.UpdateOrganization)

	op := organizationOperation("deleteOrganization", http.MethodDelete, "/auth/organizations/{organizationId}", "Delete an organization")
	op.Errors = []int{http.StatusForbidden}
	huma.Register(api, op, h.DeleteOrganization)

	huma.Register(api, organizationOperation("listOrganizationMembers", http.MethodGet, "/auth/organizations/{organizationId}/members", "List organization members"), h.ListOrganizationMembers)

	huma.Register(api, organizationOperation("setMemberRole", http.MethodPatch, "/auth/organizations/{organizationId}/members/{userId}", "Change a member's role"), h.SetMemberRole)

	huma.Register(api, organizationOperation("removeMember", http.MethodDelete, "/auth/organizations/{organizationId}/members/{userId}", "Remove a member from the organization"), h.RemoveMember)

	op = organizationOperation("leaveOrganization", http.MethodPost, "/auth/organizations/{organizationId}/leave", "Leave an organization")
	op.Errors = []int{http.StatusBadRequest}
	huma.Register(api, op, h.LeaveOrganization)

	huma.Register(api, organizationOperation("listOrganizationInvitations", http.MethodGet, "/auth/organizations/{organizationId}/invitations", "List pending invitations"), h.ListOrganizationInvitations)

	huma.Register(api, organizationOperation("inviteMember", http.MethodPost, "/auth/organizations/{organizationId}/invitations", "Invite a user to the organization"), h.InviteMember)

	huma.Register(api, organizationOperation("acceptInvitation", http.MethodPost, "/auth/invitations/{invitationId}/accept", "Accept an invitation"), h.AcceptInvitation)

	huma.Register(api, organizationOperation("rejectInvitation", http.MethodPost, "/auth/invitations/{invitationId}/reject", "Reject an invitation"), h.RejectInvitation)

	op = authOperation("setActiveOrganization", http.MethodPost, "/auth/set-active-organization", "Set the active organization for the current session", false)
	op.Errors = []int{http.StatusForbidden}
	huma.Register(api, op, h.SetActiveOrganization)
}
