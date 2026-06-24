package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/authapi"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (h *AuthHandler) AdminListUsers(w http.ResponseWriter, r *http.Request, params authapi.AdminListUsersParams) {
	su := h.requireAdmin(w, r)
	if su == nil {
		return
	}

	limit := int32(100)
	offset := int32(0)
	if params.Limit != nil && *params.Limit > 0 && *params.Limit <= 500 {
		limit = int32(*params.Limit)
	}
	if params.Offset != nil && *params.Offset >= 0 {
		offset = int32(*params.Offset)
	}

	var search *string
	if params.Search != nil {
		if s := strings.TrimSpace(*params.Search); s != "" {
			search = &s
		}
	}

	var role *string
	if params.Role != nil {
		rl := string(*params.Role)
		role = &rl
	}

	var status *string
	if params.Status != nil {
		st := string(*params.Status)
		status = &st
	}

	sortBy := "createdAt"
	if params.SortBy != nil {
		sortBy = string(*params.SortBy)
	}
	sortDir := "desc"
	if params.SortDirection != nil && string(*params.SortDirection) == "asc" {
		sortDir = "asc"
	}

	rows, err := h.queries.AdminListUsers(r.Context(), queries.AdminListUsersParams{
		Limit:   limit,
		Offset:  offset,
		Search:  search,
		Role:    role,
		Status:  status,
		SortBy:  sortBy,
		SortDir: sortDir,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	total, err := h.queries.AdminCountUsers(r.Context(), queries.AdminCountUsersParams{
		Search: search,
		Role:   role,
		Status: status,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to count users")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"users": mapAdminUsers(rows),
		"total": total,
	})
}

// AdminGetUserDetail returns a single user with sessions, auth providers,
// organization memberships and recent audit-log entries (admin only). Secrets
// (passwords, tokens) are never included.
func (h *AuthHandler) AdminGetUserDetail(w http.ResponseWriter, r *http.Request, userID string) {
	su := h.requireAdmin(w, r)
	if su == nil {
		return
	}

	user, err := h.queries.AdminGetUserDetail(r.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "user not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	accounts, err := h.queries.AdminListUserAccounts(r.Context(), userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get user accounts")
		return
	}
	sessions, err := h.queries.AdminListUserSessions(r.Context(), userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get user sessions")
		return
	}
	memberships, err := h.queries.AdminListUserMemberships(r.Context(), userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get user memberships")
		return
	}
	auditRows, err := h.queries.AdminListUserAuditLog(r.Context(), queries.AdminListUserAuditLogParams{
		ActorUserID: &userID,
		Limit:       50,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get user audit log")
		return
	}

	detail := authapi.AdminUserDetail{
		Id:            user.ID,
		Name:          user.Name,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		Image:         user.Image,
		Role:          user.Role,
		Banned:        user.Banned != nil && *user.Banned,
		BanReason:     user.BanReason,
		CreatedAt:     user.CreatedAt.Time,
		UpdatedAt:     user.UpdatedAt.Time,
		AuthProviders: make([]authapi.AdminUserAuthProvider, 0, len(accounts)),
		Sessions:      make([]authapi.AdminUserSession, 0, len(sessions)),
		Memberships:   make([]authapi.AdminUserMembership, 0, len(memberships)),
		AuditLog:      make([]authapi.AdminAuditLogEntry, 0, len(auditRows)),
	}
	if user.BanExpires.Valid {
		t := user.BanExpires.Time
		detail.BanExpires = &t
	}

	for _, a := range accounts {
		accID := a.AccountID
		detail.AuthProviders = append(detail.AuthProviders, authapi.AdminUserAuthProvider{
			Id:         a.ID,
			ProviderId: a.ProviderID,
			AccountId:  &accID,
			CreatedAt:  a.CreatedAt.Time,
		})
	}
	for _, s := range sessions {
		detail.Sessions = append(detail.Sessions, authapi.AdminUserSession{
			Id:           s.ID,
			IpAddress:    s.IpAddress,
			UserAgent:    s.UserAgent,
			Impersonated: s.ImpersonatedBy != nil && *s.ImpersonatedBy != "",
			CreatedAt:    s.CreatedAt.Time,
			ExpiresAt:    s.ExpiresAt.Time,
		})
	}
	for _, m := range memberships {
		detail.Memberships = append(detail.Memberships, authapi.AdminUserMembership{
			OrganizationId:   m.OrganizationID,
			OrganizationName: m.OrganizationName,
			OrganizationSlug: m.OrganizationSlug,
			Role:             m.Role,
			CreatedAt:        m.CreatedAt.Time,
		})
	}
	for _, a := range auditRows {
		detail.AuditLog = append(detail.AuditLog, mapAuthAuditEntry(a))
	}

	httputil.WriteJSON(w, http.StatusOK, detail)
}

// mapAuthAuditEntry converts an audit row (auth package shape) to the API entry.
func mapAuthAuditEntry(a queries.AdminListUserAuditLogRow) authapi.AdminAuditLogEntry {
	return authapi.AdminAuditLogEntry{
		Id:           a.ID,
		ActorUserId:  a.ActorUserID,
		ActorName:    a.ActorName,
		ActorEmail:   a.ActorEmail,
		Action:       a.Action,
		TargetUserId: a.TargetUserID,
		TargetName:   a.TargetName,
		TargetEmail:  a.TargetEmail,
		Detail:       a.Detail,
		IpAddress:    a.IpAddress,
		CreatedAt:    a.CreatedAt.Time,
	}
}

func (h *AuthHandler) AdminSetUserRole(w http.ResponseWriter, r *http.Request, userId string) {
	su := h.requireAdmin(w, r)
	if su == nil {
		return
	}

	var req adminSetUserRoleRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if userId == su.User.ID {
		httputil.WriteError(w, http.StatusBadRequest, "cannot change your own role")
		return
	}

	if req.Role != "user" && req.Role != "admin" {
		httputil.WriteError(w, http.StatusBadRequest, "role must be 'user' or 'admin'")
		return
	}

	err := h.queries.AdminSetUserRole(r.Context(), queries.AdminSetUserRoleParams{
		Role: req.Role,
		ID:   userId,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update role")
		return
	}

	h.recordAudit(r, su.User.ID, AuditActionSetRole, userId, "role="+req.Role)

	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

func (h *AuthHandler) AdminBanUser(w http.ResponseWriter, r *http.Request, userId string) {
	su := h.requireAdmin(w, r)
	if su == nil {
		return
	}

	var req adminBanUserRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.withTx(r.Context(), func(txq *queries.Queries) error {
		if err := txq.AdminBanUser(r.Context(), queries.AdminBanUserParams{
			BanReason: req.BanReason,
			ID:        userId,
		}); err != nil {
			return err
		}

		return txq.DeleteUserSessions(r.Context(), userId)
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to invalidate user sessions")
		return
	}

	banReason := ""
	if req.BanReason != nil {
		banReason = *req.BanReason
	}
	h.recordAudit(r, su.User.ID, AuditActionBanUser, userId, banReason)

	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

func (h *AuthHandler) AdminUnbanUser(w http.ResponseWriter, r *http.Request, userId string) {
	su := h.requireAdmin(w, r)
	if su == nil {
		return
	}

	err := h.queries.AdminUnbanUser(r.Context(), userId)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to unban user")
		return
	}

	h.recordAudit(r, su.User.ID, AuditActionUnbanUser, userId, "")

	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

func (h *AuthHandler) AdminImpersonate(w http.ResponseWriter, r *http.Request, userId string) {
	su := h.requireAdmin(w, r)
	if su == nil {
		return
	}

	// Verify target user exists
	_, err := h.queries.GetUserByID(r.Context(), userId)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "user not found")
		return
	}

	// Create session for the target user, marked as impersonated
	token := generateToken()
	sessionID := "imp-" + generateToken()[:16]

	userAgent := r.UserAgent()
	ipAddress := chimiddleware.GetClientIP(r.Context())
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}
	adminID := su.User.ID

	_, err = h.queries.AdminCreateImpersonationSession(r.Context(), queries.AdminCreateImpersonationSessionParams{
		ID:             sessionID,
		ExpiresAt:      pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour), Valid: true},
		Token:          token,
		IpAddress:      &ipAddress,
		UserAgent:      &userAgent,
		UserID:         userId,
		ImpersonatedBy: &adminID,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create impersonation session")
		return
	}

	h.recordAudit(r, su.User.ID, AuditActionImpersonate, userId, "")

	httputil.WriteJSON(w, http.StatusOK, impersonationResponse{
		Token: token,
		Session: impersonationSessionResponse{
			Token:          token,
			ImpersonatedBy: su.User.ID,
		},
	})
}
