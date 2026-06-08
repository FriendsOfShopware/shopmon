package auth

import (
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
)

type statusResponse struct {
	Status string `json:"status"`
}

type urlResponse struct {
	URL string `json:"url"`
}

type idResponse struct {
	ID string `json:"id"`
}

type permissionResponse struct {
	Success bool `json:"success"`
}

type organizationMemberResponse struct {
	ID     string  `json:"id"`
	UserID string  `json:"userId"`
	Role   string  `json:"role"`
	Name   string  `json:"name"`
	Email  string  `json:"email"`
	Image  *string `json:"image"`
}

type organizationInvitationResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Role        *string   `json:"role"`
	Status      string    `json:"status"`
	ExpiresAt   time.Time `json:"expiresAt"`
	InviterName string    `json:"inviterName"`
}

type fullOrganizationResponse struct {
	ID          string                           `json:"id"`
	Name        string                           `json:"name"`
	Logo        *string                          `json:"logo"`
	Members     []organizationMemberResponse     `json:"members"`
	Invitations []organizationInvitationResponse `json:"invitations"`
}

type sessionResponse struct {
	ID             string    `json:"id"`
	ExpiresAt      time.Time `json:"expiresAt"`
	CreatedAt      time.Time `json:"createdAt"`
	IPAddress      *string   `json:"ipAddress"`
	UserAgent      *string   `json:"userAgent"`
	ImpersonatedBy *string   `json:"impersonatedBy"`
}

type linkedAccountResponse struct {
	ID        string    `json:"id"`
	Provider  string    `json:"provider"`
	AccountID string    `json:"accountId"`
	CreatedAt time.Time `json:"createdAt"`
}

type passkeyResponse struct {
	ID         string    `json:"id"`
	Name       *string   `json:"name"`
	DeviceType string    `json:"deviceType"`
	BackedUp   bool      `json:"backedUp"`
	CreatedAt  time.Time `json:"createdAt"`
}

type userOrganizationResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Logo      *string   `json:"logo"`
	CreatedAt time.Time `json:"createdAt"`
	Role      string    `json:"role"`
}

type adminUserResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"emailVerified"`
	Image         *string   `json:"image"`
	Role          string    `json:"role"`
	Banned        *bool     `json:"banned"`
	BanReason     *string   `json:"banReason"`
	CreatedAt     time.Time `json:"createdAt"`
}

type impersonationSessionResponse struct {
	Token          string `json:"token"`
	ImpersonatedBy string `json:"impersonatedBy"`
}

type impersonationResponse struct {
	Token   string                       `json:"token"`
	Session impersonationSessionResponse `json:"session"`
}

type passkeyOptionsResponse struct {
	Options      any    `json:"options"`
	ChallengeKey string `json:"challengeKey"`
}

type passkeyRegisterResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type authenticatedUserResponse struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Email         string   `json:"email"`
	EmailVerified bool     `json:"emailVerified,omitempty"`
	Image         *string  `json:"image,omitempty"`
	Role          string   `json:"role,omitempty"`
	Notifications []string `json:"notifications,omitempty"`
}

type currentSessionResponse struct {
	ID                   string    `json:"id"`
	UserID               string    `json:"userId"`
	ExpiresAt            time.Time `json:"expiresAt"`
	ActiveOrganizationID *string   `json:"activeOrganizationId"`
}

type tokenUserResponse struct {
	Token string                    `json:"token"`
	User  authenticatedUserResponse `json:"user"`
}

type currentSessionEnvelope struct {
	User    authenticatedUserResponse `json:"user"`
	Session currentSessionResponse    `json:"session"`
}

type exchangeCodeResponse struct {
	Token string `json:"token"`
}

type createOrganizationResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func newStatusResponse() statusResponse {
	return statusResponse{Status: "ok"}
}

func mapOrganizationMembers(rows []queries.ListMembersRow) []organizationMemberResponse {
	result := make([]organizationMemberResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, organizationMemberResponse{
			ID:     row.ID,
			UserID: row.UserID,
			Role:   row.Role,
			Name:   row.UserName,
			Email:  row.UserEmail,
			Image:  row.UserImage,
		})
	}

	return result
}

func mapOrganizationInvitations(rows []queries.ListInvitationsRow) []organizationInvitationResponse {
	result := make([]organizationInvitationResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, organizationInvitationResponse{
			ID:          row.ID,
			Email:       row.Email,
			Role:        row.Role,
			Status:      row.Status,
			ExpiresAt:   row.ExpiresAt.Time,
			InviterName: row.InviterName,
		})
	}

	return result
}

func mapUserSessions(rows []queries.ListUserSessionsRow) []sessionResponse {
	result := make([]sessionResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, sessionResponse{
			ID:             row.ID,
			ExpiresAt:      row.ExpiresAt.Time,
			CreatedAt:      row.CreatedAt.Time,
			IPAddress:      row.IpAddress,
			UserAgent:      row.UserAgent,
			ImpersonatedBy: row.ImpersonatedBy,
		})
	}

	return result
}

func mapLinkedAccounts(rows []queries.ListUserAccountsRow) []linkedAccountResponse {
	result := make([]linkedAccountResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, linkedAccountResponse{
			ID:        row.ID,
			Provider:  row.ProviderID,
			AccountID: row.AccountID,
			CreatedAt: row.CreatedAt.Time,
		})
	}

	return result
}

func mapUserPasskeys(rows []queries.ListUserPasskeysRow) []passkeyResponse {
	result := make([]passkeyResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, passkeyResponse{
			ID:         row.ID,
			Name:       row.Name,
			DeviceType: row.DeviceType,
			BackedUp:   row.BackedUp,
			CreatedAt:  row.CreatedAt.Time,
		})
	}

	return result
}

func mapUserOrganizations(rows []queries.ListUserOrganizationsRow) []userOrganizationResponse {
	result := make([]userOrganizationResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, userOrganizationResponse{
			ID:        row.ID,
			Name:      row.Name,
			Logo:      row.Logo,
			CreatedAt: row.CreatedAt.Time,
			Role:      row.Role,
		})
	}

	return result
}

func mapAdminUsers(rows []queries.AdminListUsersRow) []adminUserResponse {
	result := make([]adminUserResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, adminUserResponse{
			ID:            row.ID,
			Name:          row.Name,
			Email:         row.Email,
			EmailVerified: row.EmailVerified,
			Image:         row.Image,
			Role:          row.Role,
			Banned:        row.Banned,
			BanReason:     row.BanReason,
			CreatedAt:     row.CreatedAt.Time,
		})
	}

	return result
}
