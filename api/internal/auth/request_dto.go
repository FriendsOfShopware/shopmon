package auth

import "encoding/json"

type signUpEmailInput struct {
	Body signUpEmailRequest
}

type signInEmailInput struct {
	Body signInEmailRequest
}

type forgetPasswordInput struct {
	Body forgetPasswordRequest
}

type resetPasswordInput struct {
	Body resetPasswordRequest
}

type verifyEmailInput struct {
	Token string `query:"token"`
}

type socialSignInInput struct {
	Body socialSignInRequest
}

type ssoSignInInput struct {
	Body ssoSignInRequest
}

type githubCallbackInput struct {
	Code  string `query:"code"`
	State string `query:"state"`
}

type ssoCallbackInput struct {
	ProviderID string `path:"providerId"`
	Code       string `query:"code"`
	State      string `query:"state"`
}

type exchangeCodeInput struct {
	Body exchangeCodeRequest
}

type passkeyRegisterInput struct {
	// Body must be JSON (not RawBody). RawBody is advertised as
	// application/octet-stream, and the generated client then skips JSON
	// serialization of the WebAuthn payload.
	Body json.RawMessage
}

type passkeyLoginInput struct {
	Body json.RawMessage
}

type tokenUserOutput struct {
	Body tokenUserResponse
}

type sessionOutput struct {
	Body currentSessionEnvelope
}

type authCodeOutput struct {
	Body authCodeResponse
}

type exchangeCodeOutput struct {
	Body exchangeCodeResponse
}

type passkeyOptionsOutput struct {
	Body passkeyOptionsResponse
}

type passkeyRegisterOutput struct {
	Body passkeyRegisterResponse
}

type authCodeResponse struct {
	Code string `json:"code"`
}

type exchangeCodeRequest struct {
	Code string `json:"code"`
}

type socialSignInRequest struct {
	Provider    string `json:"provider"`
	CallbackURL string `json:"callbackURL"`
}

type ssoSignInRequest struct {
	Email       string `json:"email"`
	CallbackURL string `json:"callbackURL"`
}

type signUpEmailRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type signInEmailRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type forgetPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

type revokeSessionRequest struct {
	SessionID string `json:"sessionId"`
}

type unlinkAccountRequest struct {
	ProviderID string `json:"providerId"`
}

type changeEmailRequest struct {
	NewEmail        string `json:"newEmail"`
	CurrentPassword string `json:"currentPassword"`
}

type updateUserRequest struct {
	Name string `json:"name"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type deletePasskeyRequest struct {
	ID string `json:"id"`
}

type hasPermissionRequest struct {
	OrganizationID string `json:"organizationId"`
}

type cancelInvitationRequest struct {
	InvitationID string `json:"invitationId"`
}

type registerSSOProviderRequest struct {
	OrganizationID        string `json:"organizationId"`
	Domain                string `json:"domain"`
	Issuer                string `json:"issuer"`
	ClientID              string `json:"clientId"`
	ClientSecret          string `json:"clientSecret"`
	AuthorizationEndpoint string `json:"authorizationEndpoint"`
	TokenEndpoint         string `json:"tokenEndpoint"`
	JwksEndpoint          string `json:"jwksEndpoint"`
}

type createOrganizationRequest struct {
	Name string `json:"name"`
}

type updateOrganizationRequest struct {
	Name *string `json:"name,omitempty"`
	Logo *string `json:"logo,omitempty"`
}

type inviteMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type setMemberRoleRequest struct {
	Role string `json:"role"`
}

type adminSetUserRoleRequest struct {
	Role string `json:"role"`
}

type adminBanUserRequest struct {
	BanReason *string `json:"banReason"`
}
