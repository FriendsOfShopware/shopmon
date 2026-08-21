package auth

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/identity"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type adminListUsersInput struct {
	Limit         int    `query:"limit"`
	Offset        int    `query:"offset"`
	Search        string `query:"search"`
	Role          string `query:"role"`
	Status        string `query:"status"`
	SortBy        string `query:"sortBy"`
	SortDirection string `query:"sortDirection"`
}

type adminListUsersOutput struct {
	Body struct {
		Users []adminUserResponse `json:"users"`
		Total int32               `json:"total"`
	}
}

func (h *AuthHandler) AdminListUsers(ctx context.Context, input *adminListUsersInput) (*adminListUsersOutput, error) {
	if _, err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}

	filter := identity.AdminUserFilter{Limit: 100, SortBy: "createdAt", SortDirection: "desc"}
	if input.Limit > 0 && input.Limit <= 500 {
		filter.Limit = int32(input.Limit)
	}
	if input.Offset > 0 {
		filter.Offset = int32(input.Offset)
	}
	if value := strings.TrimSpace(input.Search); value != "" {
		filter.Search = &value
	}
	if input.Role != "" {
		filter.Role = &input.Role
	}
	if input.Status != "" {
		filter.Status = &input.Status
	}
	if input.SortBy != "" {
		filter.SortBy = input.SortBy
	}
	if input.SortDirection == "asc" {
		filter.SortDirection = "asc"
	}

	page, err := h.adminUsers.ListUsers(ctx, filter)
	if err != nil {
		return nil, h.identityAdminError(ctx, "list users", err)
	}
	users := make([]adminUserResponse, 0, len(page.Users))
	for _, user := range page.Users {
		users = append(users, adminUserResponse{
			ID: user.ID, Name: user.Name, Email: user.Email, EmailVerified: user.EmailVerified,
			Image: user.Image, Role: user.Role, Banned: user.Banned, BanReason: user.BanReason,
			CreatedAt: user.CreatedAt,
		})
	}
	out := &adminListUsersOutput{}
	out.Body.Users = users
	out.Body.Total = page.Total
	return out, nil
}

type adminUserIDInput struct {
	UserID string `path:"userId"`
}

type authAdminAuditLogEntry struct {
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

type authAdminUserAuthProvider struct {
	AccountId  *string   `json:"accountId,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	Id         string    `json:"id"`
	ProviderId string    `json:"providerId"`
}

type authAdminUserMembership struct {
	CreatedAt        time.Time `json:"createdAt"`
	OrganizationId   string    `json:"organizationId"`
	OrganizationName string    `json:"organizationName"`
	OrganizationSlug string    `json:"organizationSlug"`
	Role             string    `json:"role"`
}

type authAdminUserSession struct {
	CreatedAt    time.Time `json:"createdAt"`
	ExpiresAt    time.Time `json:"expiresAt"`
	Id           string    `json:"id"`
	Impersonated bool      `json:"impersonated"`
	IpAddress    *string   `json:"ipAddress,omitempty"`
	UserAgent    *string   `json:"userAgent,omitempty"`
}

type authAdminUserDetail struct {
	AuditLog      []authAdminAuditLogEntry    `json:"auditLog"`
	AuthProviders []authAdminUserAuthProvider `json:"authProviders"`
	BanExpires    *time.Time                  `json:"banExpires,omitempty"`
	BanReason     *string                     `json:"banReason,omitempty"`
	Banned        bool                        `json:"banned"`
	CreatedAt     time.Time                   `json:"createdAt"`
	Email         string                      `json:"email"`
	EmailVerified bool                        `json:"emailVerified"`
	Id            string                      `json:"id"`
	Image         *string                     `json:"image,omitempty"`
	Memberships   []authAdminUserMembership   `json:"memberships"`
	Name          string                      `json:"name"`
	Role          string                      `json:"role"`
	Sessions      []authAdminUserSession      `json:"sessions"`
	UpdatedAt     time.Time                   `json:"updatedAt"`
}

type adminGetUserDetailOutput struct {
	Body authAdminUserDetail
}

func (h *AuthHandler) AdminGetUserDetail(ctx context.Context, input *adminUserIDInput) (*adminGetUserDetailOutput, error) {
	if _, err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}

	user, err := h.adminUsers.UserDetail(ctx, input.UserID)
	if err != nil {
		return nil, h.identityAdminError(ctx, "get user detail", err)
	}
	detail := authAdminUserDetail{
		Id: user.ID, Name: user.Name, Email: user.Email, EmailVerified: user.EmailVerified,
		Image: user.Image, Role: user.Role, Banned: user.Banned != nil && *user.Banned,
		BanReason: user.BanReason, BanExpires: user.BanExpires, CreatedAt: user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
		AuthProviders: make([]authAdminUserAuthProvider, 0, len(user.AuthProviders)),
		Sessions:      make([]authAdminUserSession, 0, len(user.Sessions)),
		Memberships:   make([]authAdminUserMembership, 0, len(user.Memberships)),
		AuditLog:      make([]authAdminAuditLogEntry, 0, len(user.AuditLog)),
	}
	for _, account := range user.AuthProviders {
		accountID := account.AccountID
		detail.AuthProviders = append(detail.AuthProviders, authAdminUserAuthProvider{
			Id: account.ID, ProviderId: account.Provider, AccountId: &accountID, CreatedAt: account.CreatedAt,
		})
	}
	for _, session := range user.Sessions {
		detail.Sessions = append(detail.Sessions, authAdminUserSession{
			Id: session.ID, IpAddress: session.IPAddress, UserAgent: session.UserAgent,
			Impersonated: session.ImpersonatedBy != nil && *session.ImpersonatedBy != "",
			CreatedAt:    session.CreatedAt, ExpiresAt: session.ExpiresAt,
		})
	}
	for _, membership := range user.Memberships {
		detail.Memberships = append(detail.Memberships, authAdminUserMembership{
			OrganizationId: membership.OrganizationID, OrganizationName: membership.OrganizationName,
			OrganizationSlug: membership.OrganizationSlug, Role: membership.Role, CreatedAt: membership.CreatedAt,
		})
	}
	for _, entry := range user.AuditLog {
		detail.AuditLog = append(detail.AuditLog, mapIdentityAuditEntry(entry))
	}
	return &adminGetUserDetailOutput{Body: detail}, nil
}

func mapIdentityAuditEntry(entry identity.AdminAuditEntry) authAdminAuditLogEntry {
	return authAdminAuditLogEntry{
		Id: entry.ID, ActorUserId: entry.ActorUserID, ActorName: entry.ActorName,
		ActorEmail: entry.ActorEmail, Action: entry.Action, TargetUserId: entry.TargetUserID,
		TargetName: entry.TargetName, TargetEmail: entry.TargetEmail, Detail: entry.Detail,
		IpAddress: entry.IPAddress, CreatedAt: entry.CreatedAt,
	}
}

type adminSetUserRoleInput struct {
	UserID string `path:"userId"`
	Body   adminSetUserRoleRequest
}

func (h *AuthHandler) AdminSetUserRole(ctx context.Context, input *adminSetUserRoleInput) (*statusOutput, error) {
	principal, err := h.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.adminUsers.SetUserRole(ctx, principal.User.ID, input.UserID, input.Body.Role); err != nil {
		return nil, h.identityAdminError(ctx, "set user role", err)
	}
	h.recordAudit(requestFromContext(ctx), principal.User.ID, AuditActionSetRole, input.UserID, "role="+input.Body.Role)
	return statusOK(), nil
}

type adminBanUserInput struct {
	UserID string `path:"userId"`
	Body   adminBanUserRequest
}

func (h *AuthHandler) AdminBanUser(ctx context.Context, input *adminBanUserInput) (*statusOutput, error) {
	principal, err := h.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.adminUsers.BanUser(ctx, input.UserID, input.Body.BanReason); err != nil {
		return nil, h.identityAdminError(ctx, "ban user", err)
	}
	reason := ""
	if input.Body.BanReason != nil {
		reason = *input.Body.BanReason
	}
	h.recordAudit(requestFromContext(ctx), principal.User.ID, AuditActionBanUser, input.UserID, reason)
	return statusOK(), nil
}

func (h *AuthHandler) AdminUnbanUser(ctx context.Context, input *adminUserIDInput) (*statusOutput, error) {
	principal, err := h.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.adminUsers.UnbanUser(ctx, input.UserID); err != nil {
		return nil, h.identityAdminError(ctx, "unban user", err)
	}
	h.recordAudit(requestFromContext(ctx), principal.User.ID, AuditActionUnbanUser, input.UserID, "")
	return statusOK(), nil
}

type adminImpersonateOutput struct {
	Body impersonationResponse
}

func (h *AuthHandler) AdminImpersonate(ctx context.Context, input *adminUserIDInput) (*adminImpersonateOutput, error) {
	principal, err := h.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	r := requestFromContext(ctx)
	ipAddress := chimiddleware.GetClientIP(ctx)
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}
	token, err := h.adminUsers.Impersonate(ctx, identity.ImpersonateCommand{
		ActorUserID: principal.User.ID, TargetUserID: input.UserID,
		IPAddress: ipAddress, UserAgent: r.UserAgent(),
	})
	if err != nil {
		return nil, h.identityAdminError(ctx, "impersonate user", err)
	}
	h.recordAudit(r, principal.User.ID, AuditActionImpersonate, input.UserID, "")
	return &adminImpersonateOutput{Body: impersonationResponse{
		Token:   token,
		Session: impersonationSessionResponse{Token: token, ImpersonatedBy: principal.User.ID},
	}}, nil
}

func (h *AuthHandler) identityAdminError(ctx context.Context, operation string, err error) error {
	switch {
	case errors.Is(err, identity.ErrUserNotFound):
		return huma.Error404NotFound("user not found")
	case errors.Is(err, identity.ErrCannotChangeSelf):
		return huma.Error400BadRequest("cannot change your own role")
	case errors.Is(err, identity.ErrInvalidSystemRole):
		return huma.Error400BadRequest("role must be 'user' or 'admin'")
	default:
		slog.ErrorContext(ctx, "identity admin operation failed", "operation", operation, "error", err)
		return huma.Error500InternalServerError("admin operation failed")
	}
}
