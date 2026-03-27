package auth

import (
	"net/http"
	"strconv"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/jackc/pgx/v5/pgtype"
)

func (h *AuthHandler) AdminListUsers(w http.ResponseWriter, r *http.Request) {
	su := h.requireAdmin(w, r)
	if su == nil {
		return
	}

	limit := int32(100)
	offset := int32(0)
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 500 {
			limit = int32(v)
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = int32(v)
		}
	}

	rows, err := h.queries.AdminListUsers(r.Context(), queries.AdminListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, mapAdminUsers(rows))
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

	ipAddress := r.RemoteAddr
	userAgent := r.UserAgent()
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

	httputil.WriteJSON(w, http.StatusOK, impersonationResponse{
		Token: token,
		Session: impersonationSessionResponse{
			Token:          token,
			ImpersonatedBy: su.User.ID,
		},
	})
}
