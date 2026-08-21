package auth

import (
	"context"
	"errors"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/identity"
	"github.com/friendsofshopware/shopmon/api/internal/organization"
	organizationsso "github.com/friendsofshopware/shopmon/api/internal/organization/sso"
)

type getFullOrganizationInput struct {
	OrganizationID string `query:"organizationId"`
}

type getFullOrganizationOutput struct {
	Body fullOrganizationResponse
}

func (h *AuthHandler) GetFullOrganization(ctx context.Context, input *getFullOrganizationInput) (*getFullOrganizationOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	if input.OrganizationID == "" {
		return nil, huma.Error400BadRequest("organizationId is required")
	}

	full, err := h.organizations.Full(ctx, principal.User.ID, input.OrganizationID)
	if err != nil {
		return nil, h.organizationError(ctx, "get full organization", err)
	}

	return &getFullOrganizationOutput{Body: fullOrganizationResponse{
		ID:          full.Organization.ID,
		Name:        full.Organization.Name,
		Logo:        full.Organization.Logo,
		Members:     mapOrganizationMembers(full.Members),
		Invitations: mapOrganizationInvitations(full.Invitations),
	}}, nil
}

type listSessionsOutput struct {
	Body []sessionResponse
}

func (h *AuthHandler) ListSessions(ctx context.Context, _ *struct{}) (*listSessionsOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	sessions, err := h.accounts.ListSessions(ctx, principal.User.ID)
	if err != nil {
		return nil, h.identityAccountError(ctx, "list sessions", err)
	}
	response := make([]sessionResponse, 0, len(sessions))
	for _, session := range sessions {
		response = append(response, sessionResponse{
			ID: session.ID, ExpiresAt: session.ExpiresAt, CreatedAt: session.CreatedAt,
			IPAddress: session.IPAddress, UserAgent: session.UserAgent, ImpersonatedBy: session.ImpersonatedBy,
		})
	}
	return &listSessionsOutput{Body: response}, nil
}

type revokeSessionInput struct {
	Body revokeSessionRequest
}

func (h *AuthHandler) RevokeSession(ctx context.Context, input *revokeSessionInput) (*statusOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	if input.Body.SessionID == "" {
		return nil, huma.Error400BadRequest("sessionId is required")
	}

	if err := h.accounts.RevokeSession(ctx, principal.User.ID, input.Body.SessionID); err != nil {
		return nil, h.identityAccountError(ctx, "revoke session", err)
	}
	return statusOK(), nil
}

type listAccountsOutput struct {
	Body []linkedAccountResponse
}

func (h *AuthHandler) ListAccounts(ctx context.Context, _ *struct{}) (*listAccountsOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	accounts, err := h.accounts.ListLinkedAccounts(ctx, principal.User.ID)
	if err != nil {
		return nil, h.identityAccountError(ctx, "list linked accounts", err)
	}
	response := make([]linkedAccountResponse, 0, len(accounts))
	for _, account := range accounts {
		response = append(response, linkedAccountResponse{
			ID: account.ID, Provider: account.Provider, AccountID: account.AccountID, CreatedAt: account.CreatedAt,
		})
	}
	return &listAccountsOutput{Body: response}, nil
}

type unlinkAccountInput struct {
	Body unlinkAccountRequest
}

func (h *AuthHandler) UnlinkAccount(ctx context.Context, input *unlinkAccountInput) (*statusOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	if input.Body.ProviderID == "" {
		return nil, huma.Error400BadRequest("providerId is required")
	}

	if err := h.accounts.UnlinkAccount(ctx, principal.User.ID, input.Body.ProviderID); err != nil {
		return nil, h.identityAccountError(ctx, "unlink account", err)
	}
	return statusOK(), nil
}

type changeEmailInput struct {
	Body changeEmailRequest
}

func (h *AuthHandler) ChangeEmail(ctx context.Context, input *changeEmailInput) (*statusOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	if input.Body.NewEmail == "" {
		return nil, huma.Error400BadRequest("email is required")
	}

	if err := h.accounts.ChangeEmail(ctx, identity.ChangeEmailCommand{
		UserID: principal.User.ID, Email: input.Body.NewEmail, CurrentPassword: input.Body.CurrentPassword,
	}); err != nil {
		return nil, h.identityAccountError(ctx, "change email", err)
	}
	return statusOK(), nil
}

type updateUserInput struct {
	Body updateUserRequest
}

func (h *AuthHandler) UpdateUser(ctx context.Context, input *updateUserInput) (*statusOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.accounts.UpdateName(ctx, principal.User.ID, input.Body.Name); err != nil {
		return nil, h.identityAccountError(ctx, "update user", err)
	}
	return statusOK(), nil
}

type changePasswordInput struct {
	Body changePasswordRequest
}

func (h *AuthHandler) ChangePassword(ctx context.Context, input *changePasswordInput) (*statusOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.accounts.ChangePassword(ctx, identity.ChangePasswordCommand{
		UserID: principal.User.ID, CurrentPassword: input.Body.CurrentPassword, NewPassword: input.Body.NewPassword,
	}); err != nil {
		return nil, h.identityAccountError(ctx, "change password", err)
	}

	h.recordAudit(requestFromContext(ctx), principal.User.ID, AuditActionPasswordChange, principal.User.ID, "")

	return statusOK(), nil
}

func (h *AuthHandler) DeleteUser(ctx context.Context, _ *struct{}) (*statusOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.accounts.DeleteUser(ctx, principal.User.ID); err != nil {
		return nil, h.identityAccountError(ctx, "delete user", err)
	}
	return statusOK(), nil
}

func (h *AuthHandler) LinkSocial(ctx context.Context, input *socialSignInInput) (*urlOutput, error) {
	if _, err := h.requireAuth(ctx); err != nil {
		return nil, err
	}
	return h.SignInSocial(ctx, input)
}

type listUserPasskeysOutput struct {
	Body []passkeyResponse
}

func (h *AuthHandler) ListUserPasskeys(ctx context.Context, _ *struct{}) (*listUserPasskeysOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	passkeys, err := h.accounts.ListPasskeys(ctx, principal.User.ID)
	if err != nil {
		return nil, h.identityAccountError(ctx, "list passkeys", err)
	}
	response := make([]passkeyResponse, 0, len(passkeys))
	for _, passkey := range passkeys {
		response = append(response, passkeyResponse{
			ID: passkey.ID, Name: passkey.Name, DeviceType: passkey.DeviceType,
			BackedUp: passkey.BackedUp, CreatedAt: passkey.CreatedAt,
		})
	}
	return &listUserPasskeysOutput{Body: response}, nil
}

type deletePasskeyInput struct {
	Body deletePasskeyRequest
}

func (h *AuthHandler) DeletePasskey(ctx context.Context, input *deletePasskeyInput) (*statusOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	if input.Body.ID == "" {
		return nil, huma.Error400BadRequest("id is required")
	}

	if err := h.accounts.DeletePasskey(ctx, principal.User.ID, input.Body.ID); err != nil {
		return nil, h.identityAccountError(ctx, "delete passkey", err)
	}
	return statusOK(), nil
}

type listUserOrganizationsOutput struct {
	Body []userOrganizationResponse
}

func (h *AuthHandler) ListUserOrganizations(ctx context.Context, _ *struct{}) (*listUserOrganizationsOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	memberships, err := h.organizations.ListForUser(ctx, principal.User.ID)
	if err != nil {
		return nil, h.organizationError(ctx, "list user organizations", err)
	}
	response := make([]userOrganizationResponse, 0, len(memberships))
	for _, membership := range memberships {
		response = append(response, userOrganizationResponse{
			ID: membership.Organization.ID, Name: membership.Organization.Name,
			Logo: membership.Organization.Logo, CreatedAt: membership.Organization.CreatedAt,
			Role: string(membership.Role),
		})
	}
	return &listUserOrganizationsOutput{Body: response}, nil
}

type hasPermissionInput struct {
	Body hasPermissionRequest
}

type permissionOutput struct {
	Body permissionResponse
}

func (h *AuthHandler) HasPermission(ctx context.Context, input *hasPermissionInput) (*permissionOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	if input.Body.OrganizationID == "" {
		return nil, huma.Error400BadRequest("organizationId is required")
	}

	hasPermission, err := h.organizations.HasAdministrativePermission(ctx, principal.User.ID, input.Body.OrganizationID)
	if err != nil {
		return nil, h.organizationError(ctx, "check organization permission", err)
	}
	return &permissionOutput{Body: permissionResponse{Success: hasPermission}}, nil
}

type cancelInvitationInput struct {
	Body cancelInvitationRequest
}

func (h *AuthHandler) CancelInvitation(ctx context.Context, input *cancelInvitationInput) (*statusOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	if input.Body.InvitationID == "" {
		return nil, huma.Error400BadRequest("invitationId is required")
	}

	if err := h.organizations.CancelInvitation(ctx, principal.User.ID, input.Body.InvitationID); err != nil {
		return nil, h.organizationError(ctx, "cancel organization invitation", err)
	}
	return statusOK(), nil
}

func (h *AuthHandler) AdminStopImpersonating(ctx context.Context, _ *struct{}) (*statusOutput, error) {
	if _, err := h.requireAuth(ctx); err != nil {
		return nil, err
	}

	if err := h.accounts.StopImpersonating(ctx, httputil.ExtractToken(requestFromContext(ctx))); err != nil {
		return nil, h.identityAccountError(ctx, "stop impersonating", err)
	}
	return statusOK(), nil
}

type registerSSOProviderInput struct {
	Body registerSSOProviderRequest
}

func (h *AuthHandler) RegisterSSOProvider(ctx context.Context, input *registerSSOProviderInput) (*statusOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	clientSecret := input.Body.ClientSecret
	if err := h.sso.Create(ctx, organizationsso.CreateCommand{
		UserID:         principal.User.ID,
		OrganizationID: input.Body.OrganizationID,
		Issuer:         input.Body.Issuer,
		Domain:         input.Body.Domain,
		Configuration: organizationsso.ProviderConfiguration{
			AuthorizationEndpoint: input.Body.AuthorizationEndpoint,
			TokenEndpoint:         input.Body.TokenEndpoint,
			JWKSEndpoint:          input.Body.JwksEndpoint,
			ClientID:              input.Body.ClientID,
			ClientSecret:          &clientSecret,
		},
	}); err != nil {
		return nil, h.authSSOError(ctx, "register SSO provider", err)
	}

	return statusOK(), nil
}

func (h *AuthHandler) authSSOError(ctx context.Context, operation string, err error) error {
	var validationError *organizationsso.ValidationError
	switch {
	case errors.Is(err, organization.ErrMembershipNotFound), errors.Is(err, organization.ErrRoleNotAllowed):
		return huma.Error403Forbidden(httputil.MsgForbidden)
	case errors.As(err, &validationError):
		return huma.Error400BadRequest(validationError.Error())
	default:
		slog.ErrorContext(ctx, "SSO operation failed", "operation", operation, "error", err)
		return huma.Error500InternalServerError("SSO operation failed")
	}
}

func (h *AuthHandler) identityAccountError(ctx context.Context, operation string, err error) error {
	switch {
	case errors.Is(err, identity.ErrLastAuthenticationMethod):
		return huma.Error400BadRequest("cannot remove your last authentication method")
	case errors.Is(err, identity.ErrPasswordNotSet):
		return huma.Error400BadRequest("no password set for this account")
	case errors.Is(err, identity.ErrIncorrectPassword):
		return huma.Error401Unauthorized("current password is incorrect")
	case errors.Is(err, identity.ErrEmailInUse):
		return huma.Error409Conflict("email already in use")
	case errors.Is(err, identity.ErrPasswordTooShort):
		return huma.Error400BadRequest("password must be at least 8 characters")
	default:
		slog.ErrorContext(ctx, "identity account operation failed", "operation", operation, "error", err)
		return huma.Error500InternalServerError("account operation failed")
	}
}
