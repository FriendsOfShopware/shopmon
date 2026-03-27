package auth

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
	Name *string `json:"name"`
	Logo *string `json:"logo"`
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
