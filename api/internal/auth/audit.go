package auth

import (
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// Audit action names recorded in the audit_log table.
const (
	AuditActionSetRole        = "admin.set_role"
	AuditActionBanUser        = "admin.ban_user"
	AuditActionUnbanUser      = "admin.unban_user"
	AuditActionImpersonate    = "admin.impersonate"
	AuditActionPasswordChange = "user.password_change"
	AuditActionPasswordReset  = "user.password_reset"
)

// recordAudit writes an audit log entry for a sensitive action. It is
// best-effort: a failure is logged but never propagated, so auditing can never
// break the action being audited. targetUserID may be empty for self-actions.
func (h *AuthHandler) recordAudit(r *http.Request, actorUserID, action, targetUserID, detail string) {
	ip := chimiddleware.GetClientIP(r.Context())
	if ip == "" {
		ip = r.RemoteAddr
	}

	var actor, target, det, ipPtr *string
	if actorUserID != "" {
		actor = &actorUserID
	}
	if targetUserID != "" {
		target = &targetUserID
	}
	if detail != "" {
		det = &detail
	}
	if ip != "" {
		ipPtr = &ip
	}

	if err := h.queries.CreateAuditLog(r.Context(), queries.CreateAuditLogParams{
		ActorUserID:  actor,
		Action:       action,
		TargetUserID: target,
		Detail:       det,
		IpAddress:    ipPtr,
	}); err != nil {
		slog.Error("failed to write audit log", "error", err, "action", action)
	}

	slog.Info("audit", "action", action, "actor", actorUserID, "target", targetUserID, "ip", ip)
}
