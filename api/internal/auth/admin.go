package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/authapi"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/identity"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func (h *AuthHandler) AdminListUsers(w http.ResponseWriter, r *http.Request, params authapi.AdminListUsersParams) {
	if h.requireAdmin(w, r) == nil {
		return
	}

	filter := identity.AdminUserFilter{Limit: 100, SortBy: "createdAt", SortDirection: "desc"}
	if params.Limit != nil && *params.Limit > 0 && *params.Limit <= 500 {
		filter.Limit = int32(*params.Limit)
	}
	if params.Offset != nil && *params.Offset >= 0 {
		filter.Offset = int32(*params.Offset)
	}
	if params.Search != nil {
		if value := strings.TrimSpace(*params.Search); value != "" {
			filter.Search = &value
		}
	}
	if params.Role != nil {
		value := string(*params.Role)
		filter.Role = &value
	}
	if params.Status != nil {
		value := string(*params.Status)
		filter.Status = &value
	}
	if params.SortBy != nil {
		filter.SortBy = string(*params.SortBy)
	}
	if params.SortDirection != nil && string(*params.SortDirection) == "asc" {
		filter.SortDirection = "asc"
	}

	page, err := h.adminUsers.ListUsers(r.Context(), filter)
	if err != nil {
		h.writeIdentityAdminError(w, r, "list users", err)
		return
	}
	users := make([]adminUserResponse, 0, len(page.Users))
	for _, user := range page.Users {
		users = append(users, adminUserResponse{
			ID: user.ID, Name: user.Name, Email: user.Email, EmailVerified: user.EmailVerified,
			Image: user.Image, Role: user.Role, Banned: user.Banned, BanReason: user.BanReason,
			CreatedAt: user.CreatedAt,
		})
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"users": users, "total": page.Total})
}

func (h *AuthHandler) AdminGetUserDetail(w http.ResponseWriter, r *http.Request, userID string) {
	if h.requireAdmin(w, r) == nil {
		return
	}

	user, err := h.adminUsers.UserDetail(r.Context(), userID)
	if err != nil {
		h.writeIdentityAdminError(w, r, "get user detail", err)
		return
	}
	detail := authapi.AdminUserDetail{
		Id: user.ID, Name: user.Name, Email: user.Email, EmailVerified: user.EmailVerified,
		Image: user.Image, Role: user.Role, Banned: user.Banned != nil && *user.Banned,
		BanReason: user.BanReason, BanExpires: user.BanExpires, CreatedAt: user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
		AuthProviders: make([]authapi.AdminUserAuthProvider, 0, len(user.AuthProviders)),
		Sessions:      make([]authapi.AdminUserSession, 0, len(user.Sessions)),
		Memberships:   make([]authapi.AdminUserMembership, 0, len(user.Memberships)),
		AuditLog:      make([]authapi.AdminAuditLogEntry, 0, len(user.AuditLog)),
	}
	for _, account := range user.AuthProviders {
		accountID := account.AccountID
		detail.AuthProviders = append(detail.AuthProviders, authapi.AdminUserAuthProvider{
			Id: account.ID, ProviderId: account.Provider, AccountId: &accountID, CreatedAt: account.CreatedAt,
		})
	}
	for _, session := range user.Sessions {
		detail.Sessions = append(detail.Sessions, authapi.AdminUserSession{
			Id: session.ID, IpAddress: session.IPAddress, UserAgent: session.UserAgent,
			Impersonated: session.ImpersonatedBy != nil && *session.ImpersonatedBy != "",
			CreatedAt:    session.CreatedAt, ExpiresAt: session.ExpiresAt,
		})
	}
	for _, membership := range user.Memberships {
		detail.Memberships = append(detail.Memberships, authapi.AdminUserMembership{
			OrganizationId: membership.OrganizationID, OrganizationName: membership.OrganizationName,
			OrganizationSlug: membership.OrganizationSlug, Role: membership.Role, CreatedAt: membership.CreatedAt,
		})
	}
	for _, entry := range user.AuditLog {
		detail.AuditLog = append(detail.AuditLog, mapIdentityAuditEntry(entry))
	}
	httputil.WriteJSON(w, http.StatusOK, detail)
}

func mapIdentityAuditEntry(entry identity.AdminAuditEntry) authapi.AdminAuditLogEntry {
	return authapi.AdminAuditLogEntry{
		Id: entry.ID, ActorUserId: entry.ActorUserID, ActorName: entry.ActorName,
		ActorEmail: entry.ActorEmail, Action: entry.Action, TargetUserId: entry.TargetUserID,
		TargetName: entry.TargetName, TargetEmail: entry.TargetEmail, Detail: entry.Detail,
		IpAddress: entry.IPAddress, CreatedAt: entry.CreatedAt,
	}
}

func (h *AuthHandler) AdminSetUserRole(w http.ResponseWriter, r *http.Request, userID string) {
	principal := h.requireAdmin(w, r)
	if principal == nil {
		return
	}

	var request adminSetUserRoleRequest
	if err := httputil.DecodeBody(r, &request); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.adminUsers.SetUserRole(r.Context(), principal.User.ID, userID, request.Role); err != nil {
		h.writeIdentityAdminError(w, r, "set user role", err)
		return
	}
	h.recordAudit(r, principal.User.ID, AuditActionSetRole, userID, "role="+request.Role)
	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

func (h *AuthHandler) AdminBanUser(w http.ResponseWriter, r *http.Request, userID string) {
	principal := h.requireAdmin(w, r)
	if principal == nil {
		return
	}

	var request adminBanUserRequest
	if err := httputil.DecodeBody(r, &request); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.adminUsers.BanUser(r.Context(), userID, request.BanReason); err != nil {
		h.writeIdentityAdminError(w, r, "ban user", err)
		return
	}
	reason := ""
	if request.BanReason != nil {
		reason = *request.BanReason
	}
	h.recordAudit(r, principal.User.ID, AuditActionBanUser, userID, reason)
	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

func (h *AuthHandler) AdminUnbanUser(w http.ResponseWriter, r *http.Request, userID string) {
	principal := h.requireAdmin(w, r)
	if principal == nil {
		return
	}
	if err := h.adminUsers.UnbanUser(r.Context(), userID); err != nil {
		h.writeIdentityAdminError(w, r, "unban user", err)
		return
	}
	h.recordAudit(r, principal.User.ID, AuditActionUnbanUser, userID, "")
	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

func (h *AuthHandler) AdminImpersonate(w http.ResponseWriter, r *http.Request, userID string) {
	principal := h.requireAdmin(w, r)
	if principal == nil {
		return
	}
	ipAddress := chimiddleware.GetClientIP(r.Context())
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}
	token, err := h.adminUsers.Impersonate(r.Context(), identity.ImpersonateCommand{
		ActorUserID: principal.User.ID, TargetUserID: userID,
		IPAddress: ipAddress, UserAgent: r.UserAgent(),
	})
	if err != nil {
		h.writeIdentityAdminError(w, r, "impersonate user", err)
		return
	}
	h.recordAudit(r, principal.User.ID, AuditActionImpersonate, userID, "")
	httputil.WriteJSON(w, http.StatusOK, impersonationResponse{
		Token:   token,
		Session: impersonationSessionResponse{Token: token, ImpersonatedBy: principal.User.ID},
	})
}

func (h *AuthHandler) writeIdentityAdminError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	switch {
	case errors.Is(err, identity.ErrUserNotFound):
		httputil.WriteError(w, http.StatusNotFound, "user not found")
	case errors.Is(err, identity.ErrCannotChangeSelf):
		httputil.WriteError(w, http.StatusBadRequest, "cannot change your own role")
	case errors.Is(err, identity.ErrInvalidSystemRole):
		httputil.WriteError(w, http.StatusBadRequest, "role must be 'user' or 'admin'")
	default:
		slog.ErrorContext(r.Context(), "identity admin operation failed", "operation", operation, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "admin operation failed")
	}
}
