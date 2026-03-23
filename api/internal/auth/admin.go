package auth

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
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

	users := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		users = append(users, map[string]interface{}{
			"id":            row.ID,
			"name":          row.Name,
			"email":         row.Email,
			"emailVerified": row.EmailVerified,
			"image":         row.Image,
			"role":          row.Role,
			"banned":        row.Banned,
			"banReason":     row.BanReason,
			"createdAt":     row.CreatedAt.Time,
		})
	}

	httputil.WriteJSON(w, http.StatusOK, users)
}

func (h *AuthHandler) AdminSetUserRole(w http.ResponseWriter, r *http.Request, userId string) {
	su := h.requireAdmin(w, r)
	if su == nil {
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	json.NewDecoder(r.Body).Decode(&req)

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

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandler) AdminBanUser(w http.ResponseWriter, r *http.Request, userId string) {
	su := h.requireAdmin(w, r)
	if su == nil {
		return
	}

	var req struct {
		BanReason *string `json:"banReason"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	err := h.queries.AdminBanUser(r.Context(), queries.AdminBanUserParams{
		BanReason: req.BanReason,
		ID:        userId,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to ban user")
		return
	}

	// Invalidate all sessions
	h.queries.DeleteUserSessions(r.Context(), userId)

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
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

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
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

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"session": map[string]string{
			"token":          token,
			"impersonatedBy": su.User.ID,
		},
	})
}
