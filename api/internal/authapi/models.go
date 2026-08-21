// Package authapi contains hand-written HTTP DTO types shared by handlers and read models.
package authapi

import (
	"time"
)

// Defines values for AdminListUsersParamsRole.
const (
	AdminListUsersParamsRoleAdmin AdminListUsersParamsRole = "admin"
	AdminListUsersParamsRoleUser  AdminListUsersParamsRole = "user"
)

// Valid indicates whether the value is a known member of the AdminListUsersParamsRole enum.
func (e AdminListUsersParamsRole) Valid() bool {
	switch e {
	case AdminListUsersParamsRoleAdmin:
		return true
	case AdminListUsersParamsRoleUser:
		return true
	default:
		return false
	}
}

// Defines values for AdminListUsersParamsStatus.
const (
	Active     AdminListUsersParamsStatus = "active"
	Banned     AdminListUsersParamsStatus = "banned"
	Unverified AdminListUsersParamsStatus = "unverified"
)

// Valid indicates whether the value is a known member of the AdminListUsersParamsStatus enum.
func (e AdminListUsersParamsStatus) Valid() bool {
	switch e {
	case Active:
		return true
	case Banned:
		return true
	case Unverified:
		return true
	default:
		return false
	}
}

// Defines values for AdminListUsersParamsSortBy.
const (
	CreatedAt AdminListUsersParamsSortBy = "createdAt"
	Email     AdminListUsersParamsSortBy = "email"
	Name      AdminListUsersParamsSortBy = "name"
)

// Valid indicates whether the value is a known member of the AdminListUsersParamsSortBy enum.
func (e AdminListUsersParamsSortBy) Valid() bool {
	switch e {
	case CreatedAt:
		return true
	case Email:
		return true
	case Name:
		return true
	default:
		return false
	}
}

// Defines values for AdminListUsersParamsSortDirection.
const (
	Asc  AdminListUsersParamsSortDirection = "asc"
	Desc AdminListUsersParamsSortDirection = "desc"
)

// Valid indicates whether the value is a known member of the AdminListUsersParamsSortDirection enum.
func (e AdminListUsersParamsSortDirection) Valid() bool {
	switch e {
	case Asc:
		return true
	case Desc:
		return true
	default:
		return false
	}
}

// Defines values for AdminSetUserRoleJSONBodyRole.
const (
	AdminSetUserRoleJSONBodyRoleAdmin AdminSetUserRoleJSONBodyRole = "admin"
	AdminSetUserRoleJSONBodyRoleUser  AdminSetUserRoleJSONBodyRole = "user"
)

// Valid indicates whether the value is a known member of the AdminSetUserRoleJSONBodyRole enum.
func (e AdminSetUserRoleJSONBodyRole) Valid() bool {
	switch e {
	case AdminSetUserRoleJSONBodyRoleAdmin:
		return true
	case AdminSetUserRoleJSONBodyRoleUser:
		return true
	default:
		return false
	}
}

// Defines values for SignInSocialJSONBodyProvider.
const (
	Github SignInSocialJSONBodyProvider = "github"
)

// Valid indicates whether the value is a known member of the SignInSocialJSONBodyProvider enum.
func (e SignInSocialJSONBodyProvider) Valid() bool {
	switch e {
	case Github:
		return true
	default:
		return false
	}
}

// AdminAuditLogEntry defines model for AdminAuditLogEntry.
type AdminAuditLogEntry struct {
	Action       string    `json:"action"`
	ActorEmail   *string   `json:"actorEmail,omitempty"`
	ActorName    *string   `json:"actorName,omitempty"`
	ActorUserId  *string   `json:"actorUserId,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	Detail       *string   `json:"detail,omitempty"`
	Id           int64     `json:"id"`
	IpAddress    *string   `json:"ipAddress,omitempty"`
	TargetEmail  *string   `json:"targetEmail,omitempty"`
	TargetName   *string   `json:"targetName,omitempty"`
	TargetUserId *string   `json:"targetUserId,omitempty"`
}

// AdminUser defines model for AdminUser.
type AdminUser struct {
	BanReason     *string   `json:"banReason,omitempty"`
	Banned        *bool     `json:"banned,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"emailVerified"`
	Id            string    `json:"id"`
	Image         *string   `json:"image,omitempty"`
	Name          string    `json:"name"`
	Role          string    `json:"role"`
}

// AdminUserAuthProvider defines model for AdminUserAuthProvider.
type AdminUserAuthProvider struct {
	AccountId  *string   `json:"accountId,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	Id         string    `json:"id"`
	ProviderId string    `json:"providerId"`
}

// AdminUserDetail defines model for AdminUserDetail.
type AdminUserDetail struct {
	AuditLog      []AdminAuditLogEntry    `json:"auditLog"`
	AuthProviders []AdminUserAuthProvider `json:"authProviders"`
	BanExpires    *time.Time              `json:"banExpires,omitempty"`
	BanReason     *string                 `json:"banReason,omitempty"`
	Banned        bool                    `json:"banned"`
	CreatedAt     time.Time               `json:"createdAt"`
	Email         string                  `json:"email"`
	EmailVerified bool                    `json:"emailVerified"`
	Id            string                  `json:"id"`
	Image         *string                 `json:"image,omitempty"`
	Memberships   []AdminUserMembership   `json:"memberships"`
	Name          string                  `json:"name"`
	Role          string                  `json:"role"`
	Sessions      []AdminUserSession      `json:"sessions"`
	UpdatedAt     time.Time               `json:"updatedAt"`
}

// AdminUserMembership defines model for AdminUserMembership.
type AdminUserMembership struct {
	CreatedAt        time.Time `json:"createdAt"`
	OrganizationId   string    `json:"organizationId"`
	OrganizationName string    `json:"organizationName"`
	OrganizationSlug string    `json:"organizationSlug"`
	Role             string    `json:"role"`
}

// AdminUserSession defines model for AdminUserSession.
type AdminUserSession struct {
	CreatedAt    time.Time `json:"createdAt"`
	ExpiresAt    time.Time `json:"expiresAt"`
	Id           string    `json:"id"`
	Impersonated bool      `json:"impersonated"`
	IpAddress    *string   `json:"ipAddress,omitempty"`
	UserAgent    *string   `json:"userAgent,omitempty"`
}

// AdminUsersResponse defines model for AdminUsersResponse.
type AdminUsersResponse struct {
	Total int         `json:"total"`
	Users []AdminUser `json:"users"`
}

// AuthUser defines model for AuthUser.
type AuthUser struct {
	Email         *string   `json:"email,omitempty"`
	EmailVerified *bool     `json:"emailVerified,omitempty"`
	Id            *string   `json:"id,omitempty"`
	Image         *string   `json:"image,omitempty"`
	Name          *string   `json:"name,omitempty"`
	Notifications *[]string `json:"notifications,omitempty"`
	Role          *string   `json:"role,omitempty"`
}

// ErrorResponse defines model for ErrorResponse.
type ErrorResponse struct {
	Message string `json:"message"`
}

// SessionInfo defines model for SessionInfo.
type SessionInfo struct {
	ActiveOrganizationId *string    `json:"activeOrganizationId,omitempty"`
	ExpiresAt            *time.Time `json:"expiresAt,omitempty"`
	Id                   *string    `json:"id,omitempty"`

	// ImpersonatedBy Admin user ID when this session is an impersonation
	ImpersonatedBy *string `json:"impersonatedBy,omitempty"`
	UserId         *string `json:"userId,omitempty"`
}

// NotFound defines model for NotFound.
type NotFound = ErrorResponse

// Unauthorized defines model for Unauthorized.
type Unauthorized = ErrorResponse

// ValidationError defines model for ValidationError.
type ValidationError = ErrorResponse

// AdminListUsersParams defines parameters for AdminListUsers.
type AdminListUsersParams struct {
	Limit  *int `form:"limit,omitempty" json:"limit,omitempty"`
	Offset *int `form:"offset,omitempty" json:"offset,omitempty"`

	// Search Filter users by email or name (case-insensitive substring)
	Search *string `form:"search,omitempty" json:"search,omitempty"`

	// Role Filter users by role
	Role *AdminListUsersParamsRole `form:"role,omitempty" json:"role,omitempty"`

	// Status Filter users by account status
	Status        *AdminListUsersParamsStatus        `form:"status,omitempty" json:"status,omitempty"`
	SortBy        *AdminListUsersParamsSortBy        `form:"sortBy,omitempty" json:"sortBy,omitempty"`
	SortDirection *AdminListUsersParamsSortDirection `form:"sortDirection,omitempty" json:"sortDirection,omitempty"`
}

// AdminListUsersParamsRole defines parameters for AdminListUsers.
type AdminListUsersParamsRole string

// AdminListUsersParamsStatus defines parameters for AdminListUsers.
type AdminListUsersParamsStatus string

// AdminListUsersParamsSortBy defines parameters for AdminListUsers.
type AdminListUsersParamsSortBy string

// AdminListUsersParamsSortDirection defines parameters for AdminListUsers.
type AdminListUsersParamsSortDirection string

// AdminBanUserJSONBody defines parameters for AdminBanUser.
type AdminBanUserJSONBody struct {
	BanReason *string `json:"banReason,omitempty"`
}

// AdminSetUserRoleJSONBody defines parameters for AdminSetUserRole.
type AdminSetUserRoleJSONBody struct {
	Role AdminSetUserRoleJSONBodyRole `json:"role"`
}

// AdminSetUserRoleJSONBodyRole defines parameters for AdminSetUserRole.
type AdminSetUserRoleJSONBodyRole string

// GithubCallbackParams defines parameters for GithubCallback.
type GithubCallbackParams struct {
	Code  string `form:"code" json:"code"`
	State string `form:"state" json:"state"`
}

// CancelInvitationJSONBody defines parameters for CancelInvitation.
type CancelInvitationJSONBody struct {
	InvitationId string `json:"invitationId"`
}

// ChangeEmailJSONBody defines parameters for ChangeEmail.
type ChangeEmailJSONBody struct {
	CurrentPassword string `json:"currentPassword"`
	NewEmail        string `json:"newEmail"`
}

// ChangePasswordJSONBody defines parameters for ChangePassword.
type ChangePasswordJSONBody struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// ExchangeCodeJSONBody defines parameters for ExchangeCode.
type ExchangeCodeJSONBody struct {
	Code string `json:"code"`
}

// ForgetPasswordJSONBody defines parameters for ForgetPassword.
type ForgetPasswordJSONBody struct {
	Email string `json:"email"`
}

// GetFullOrganizationParams defines parameters for GetFullOrganization.
type GetFullOrganizationParams struct {
	OrganizationId string `form:"organizationId" json:"organizationId"`
}

// HasPermissionJSONBody defines parameters for HasPermission.
type HasPermissionJSONBody struct {
	OrganizationId string `json:"organizationId"`
}

// LinkSocialJSONBody defines parameters for LinkSocial.
type LinkSocialJSONBody struct {
	CallbackURL *string `json:"callbackURL,omitempty"`
	Provider    string  `json:"provider"`
}

// CreateOrganizationJSONBody defines parameters for CreateOrganization.
type CreateOrganizationJSONBody struct {
	Name string `json:"name"`
}

// UpdateOrganizationJSONBody defines parameters for UpdateOrganization.
type UpdateOrganizationJSONBody struct {
	Logo *string `json:"logo,omitempty"`
	Name *string `json:"name,omitempty"`
}

// InviteMemberJSONBody defines parameters for InviteMember.
type InviteMemberJSONBody struct {
	Email string  `json:"email"`
	Role  *string `json:"role,omitempty"`
}

// SetMemberRoleJSONBody defines parameters for SetMemberRole.
type SetMemberRoleJSONBody struct {
	Role string `json:"role"`
}

// DeletePasskeyJSONBody defines parameters for DeletePasskey.
type DeletePasskeyJSONBody struct {
	Id string `json:"id"`
}

// PasskeyLoginJSONBody defines parameters for PasskeyLogin.
type PasskeyLoginJSONBody struct {
	ChallengeKey string `json:"challengeKey"`
}

// PasskeyRegisterJSONBody defines parameters for PasskeyRegister.
type PasskeyRegisterJSONBody struct {
	ChallengeKey string  `json:"challengeKey"`
	Name         *string `json:"name,omitempty"`
}

// ResetPasswordJSONBody defines parameters for ResetPassword.
type ResetPasswordJSONBody struct {
	NewPassword string `json:"newPassword"`
	Token       string `json:"token"`
}

// RevokeSessionJSONBody defines parameters for RevokeSession.
type RevokeSessionJSONBody struct {
	SessionId string `json:"sessionId"`
}

// SetActiveOrganizationJSONBody defines parameters for SetActiveOrganization.
type SetActiveOrganizationJSONBody struct {
	OrganizationId string `json:"organizationId"`
}

// SignInEmailJSONBody defines parameters for SignInEmail.
type SignInEmailJSONBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SignInSocialJSONBody defines parameters for SignInSocial.
type SignInSocialJSONBody struct {
	CallbackURL *string                      `json:"callbackURL,omitempty"`
	Provider    SignInSocialJSONBodyProvider `json:"provider"`
}

// SignInSocialJSONBodyProvider defines parameters for SignInSocial.
type SignInSocialJSONBodyProvider string

// SignInSSOJSONBody defines parameters for SignInSSO.
type SignInSSOJSONBody struct {
	CallbackURL *string `json:"callbackURL,omitempty"`
	Email       string  `json:"email"`
}

// SignUpEmailJSONBody defines parameters for SignUpEmail.
type SignUpEmailJSONBody struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// SsoCallbackParams defines parameters for SsoCallback.
type SsoCallbackParams struct {
	Code  string `form:"code" json:"code"`
	State string `form:"state" json:"state"`
}

// RegisterSSOProviderJSONBody defines parameters for RegisterSSOProvider.
type RegisterSSOProviderJSONBody struct {
	AuthorizationEndpoint string `json:"authorizationEndpoint"`
	ClientId              string `json:"clientId"`
	ClientSecret          string `json:"clientSecret"`
	Domain                string `json:"domain"`
	Issuer                string `json:"issuer"`
	JwksEndpoint          string `json:"jwksEndpoint"`
	OrganizationId        string `json:"organizationId"`
	TokenEndpoint         string `json:"tokenEndpoint"`
}

// UnlinkAccountJSONBody defines parameters for UnlinkAccount.
type UnlinkAccountJSONBody struct {
	ProviderId string `json:"providerId"`
}

// UpdateUserJSONBody defines parameters for UpdateUser.
type UpdateUserJSONBody struct {
	Name *string `json:"name,omitempty"`
}

// VerifyEmailParams defines parameters for VerifyEmail.
type VerifyEmailParams struct {
	Token string `form:"token" json:"token"`
}

// AdminBanUserJSONRequestBody defines body for AdminBanUser for application/json ContentType.
type AdminBanUserJSONRequestBody AdminBanUserJSONBody

// AdminSetUserRoleJSONRequestBody defines body for AdminSetUserRole for application/json ContentType.
type AdminSetUserRoleJSONRequestBody AdminSetUserRoleJSONBody

// CancelInvitationJSONRequestBody defines body for CancelInvitation for application/json ContentType.
type CancelInvitationJSONRequestBody CancelInvitationJSONBody

// ChangeEmailJSONRequestBody defines body for ChangeEmail for application/json ContentType.
type ChangeEmailJSONRequestBody ChangeEmailJSONBody

// ChangePasswordJSONRequestBody defines body for ChangePassword for application/json ContentType.
type ChangePasswordJSONRequestBody ChangePasswordJSONBody

// ExchangeCodeJSONRequestBody defines body for ExchangeCode for application/json ContentType.
type ExchangeCodeJSONRequestBody ExchangeCodeJSONBody

// ForgetPasswordJSONRequestBody defines body for ForgetPassword for application/json ContentType.
type ForgetPasswordJSONRequestBody ForgetPasswordJSONBody

// HasPermissionJSONRequestBody defines body for HasPermission for application/json ContentType.
type HasPermissionJSONRequestBody HasPermissionJSONBody

// LinkSocialJSONRequestBody defines body for LinkSocial for application/json ContentType.
type LinkSocialJSONRequestBody LinkSocialJSONBody

// CreateOrganizationJSONRequestBody defines body for CreateOrganization for application/json ContentType.
type CreateOrganizationJSONRequestBody CreateOrganizationJSONBody

// UpdateOrganizationJSONRequestBody defines body for UpdateOrganization for application/json ContentType.
type UpdateOrganizationJSONRequestBody UpdateOrganizationJSONBody

// InviteMemberJSONRequestBody defines body for InviteMember for application/json ContentType.
type InviteMemberJSONRequestBody InviteMemberJSONBody

// SetMemberRoleJSONRequestBody defines body for SetMemberRole for application/json ContentType.
type SetMemberRoleJSONRequestBody SetMemberRoleJSONBody

// DeletePasskeyJSONRequestBody defines body for DeletePasskey for application/json ContentType.
type DeletePasskeyJSONRequestBody DeletePasskeyJSONBody

// PasskeyLoginJSONRequestBody defines body for PasskeyLogin for application/json ContentType.
type PasskeyLoginJSONRequestBody PasskeyLoginJSONBody

// PasskeyRegisterJSONRequestBody defines body for PasskeyRegister for application/json ContentType.
type PasskeyRegisterJSONRequestBody PasskeyRegisterJSONBody

// ResetPasswordJSONRequestBody defines body for ResetPassword for application/json ContentType.
type ResetPasswordJSONRequestBody ResetPasswordJSONBody

// RevokeSessionJSONRequestBody defines body for RevokeSession for application/json ContentType.
type RevokeSessionJSONRequestBody RevokeSessionJSONBody

// SetActiveOrganizationJSONRequestBody defines body for SetActiveOrganization for application/json ContentType.
type SetActiveOrganizationJSONRequestBody SetActiveOrganizationJSONBody

// SignInEmailJSONRequestBody defines body for SignInEmail for application/json ContentType.
type SignInEmailJSONRequestBody SignInEmailJSONBody

// SignInSocialJSONRequestBody defines body for SignInSocial for application/json ContentType.
type SignInSocialJSONRequestBody SignInSocialJSONBody

// SignInSSOJSONRequestBody defines body for SignInSSO for application/json ContentType.
type SignInSSOJSONRequestBody SignInSSOJSONBody

// SignUpEmailJSONRequestBody defines body for SignUpEmail for application/json ContentType.
type SignUpEmailJSONRequestBody SignUpEmailJSONBody

// RegisterSSOProviderJSONRequestBody defines body for RegisterSSOProvider for application/json ContentType.
type RegisterSSOProviderJSONRequestBody RegisterSSOProviderJSONBody

// UnlinkAccountJSONRequestBody defines body for UnlinkAccount for application/json ContentType.
type UnlinkAccountJSONRequestBody UnlinkAccountJSONBody

// UpdateUserJSONRequestBody defines body for UpdateUser for application/json ContentType.
type UpdateUserJSONRequestBody UpdateUserJSONBody
