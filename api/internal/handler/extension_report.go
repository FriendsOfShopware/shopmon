package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
)

// Audit action names recorded when an admin moderates an extension report.
const (
	auditActionExtensionReportApproved = "extension_report.approved"
	auditActionExtensionReportRejected = "extension_report.rejected"
)

// ReportExtension lets an authenticated user flag a store extension (e.g. for a
// performance or security issue). Reports start in the 'pending' state and only
// become community-visible once an admin approves them.
func (h *Handler) ReportExtension(w http.ResponseWriter, r *http.Request, extensionName api.ExtensionName) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	var req api.ExtensionReportRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusUnprocessableEntity, "invalid request body")
		return
	}

	if !req.Category.Valid() {
		httputil.WriteError(w, http.StatusUnprocessableEntity, "invalid category")
		return
	}
	comment := strings.TrimSpace(req.Comment)
	if comment == "" {
		httputil.WriteError(w, http.StatusUnprocessableEntity, "comment is required")
		return
	}
	if len(comment) > 2000 {
		comment = comment[:2000]
	}

	exists, err := h.queries.StoreExtensionExists(r.Context(), extensionName)
	if err != nil {
		slog.Error("failed to check store extension", "extension", extensionName, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to submit report")
		return
	}
	if !exists {
		httputil.WriteError(w, http.StatusNotFound, "extension not found")
		return
	}

	// Avoid letting the same user pile up duplicate pending reports for an
	// extension; treat a repeat submission as a no-op success.
	pending, err := h.queries.CountUserPendingReportsForExtension(r.Context(), queries.CountUserPendingReportsForExtensionParams{
		ExtensionName: extensionName,
		UserID:        &user.ID,
	})
	if err != nil {
		slog.Error("failed to count pending reports", "extension", extensionName, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to submit report")
		return
	}
	if pending > 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if _, err := h.queries.CreateStoreExtensionReport(r.Context(), queries.CreateStoreExtensionReportParams{
		ExtensionName: extensionName,
		UserID:        &user.ID,
		Category:      string(req.Category),
		Comment:       comment,
	}); err != nil {
		slog.Error("failed to create extension report", "extension", extensionName, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to submit report")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AdminGetExtensionReports lists submitted extension reports for moderation.
func (h *Handler) AdminGetExtensionReports(w http.ResponseWriter, r *http.Request, params api.AdminGetExtensionReportsParams) {
	if !h.requireAdmin(w, r) {
		return
	}

	limit := int32(50)
	offset := int32(0)
	if params.Limit != nil && *params.Limit > 0 && *params.Limit <= 200 {
		limit = int32(*params.Limit)
	}
	if params.Offset != nil && *params.Offset >= 0 {
		offset = int32(*params.Offset)
	}

	var status *string
	if params.Status != nil {
		s := string(*params.Status)
		status = &s
	}

	rows, err := h.queries.AdminListExtensionReports(r.Context(), queries.AdminListExtensionReportsParams{
		Limit:  limit,
		Offset: offset,
		Status: status,
	})
	if err != nil {
		slog.Error("failed to list extension reports", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get reports")
		return
	}

	total, err := h.queries.AdminCountExtensionReports(r.Context(), status)
	if err != nil {
		slog.Error("failed to count extension reports", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get reports")
		return
	}

	reports := make([]api.AdminExtensionReport, 0, len(rows))
	for _, row := range rows {
		reports = append(reports, api.AdminExtensionReport{
			Id:             int(row.ID),
			ExtensionName:  row.ExtensionName,
			ExtensionLabel: row.ExtensionLabel,
			Category:       row.Category,
			Comment:        row.Comment,
			Status:         row.Status,
			ReporterId:     row.ReporterID,
			ReporterName:   row.ReporterName,
			ReporterEmail:  row.ReporterEmail,
			ReviewerName:   row.ReviewerName,
			ReviewedAt:     pgtimeToTimePtr(row.ReviewedAt),
			CreatedAt:      pgtimeToTime(row.CreatedAt),
		})
	}

	httputil.WriteJSON(w, http.StatusOK, api.AdminExtensionReportsResponse{
		Reports: reports,
		Total:   int(total),
	})
}

// AdminApproveExtensionReport marks a report approved so it becomes a
// community-visible warning on the extension.
func (h *Handler) AdminApproveExtensionReport(w http.ResponseWriter, r *http.Request, reportId api.ReportId) {
	h.moderateExtensionReport(w, r, int32(reportId), "approved", auditActionExtensionReportApproved)
}

// AdminRejectExtensionReport marks a report rejected; it stays in the audit
// trail but is never surfaced to other users.
func (h *Handler) AdminRejectExtensionReport(w http.ResponseWriter, r *http.Request, reportId api.ReportId) {
	h.moderateExtensionReport(w, r, int32(reportId), "rejected", auditActionExtensionReportRejected)
}

func (h *Handler) moderateExtensionReport(w http.ResponseWriter, r *http.Request, reportID int32, status, auditAction string) {
	if !h.requireAdmin(w, r) {
		return
	}
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	report, err := h.queries.GetStoreExtensionReportByID(r.Context(), reportID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "report not found")
			return
		}
		slog.Error("failed to get extension report", "reportId", reportID, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update report")
		return
	}

	if err := h.queries.SetStoreExtensionReportStatus(r.Context(), queries.SetStoreExtensionReportStatusParams{
		ID:         reportID,
		Status:     status,
		ReviewedBy: &user.ID,
	}); err != nil {
		slog.Error("failed to update extension report", "reportId", reportID, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update report")
		return
	}

	h.recordReportAudit(r, user.ID, auditAction, report.ExtensionName)

	w.WriteHeader(http.StatusNoContent)
}

// recordReportAudit writes a best-effort audit log entry for report moderation.
func (h *Handler) recordReportAudit(r *http.Request, actorUserID, action, extensionName string) {
	ip := chimiddleware.GetClientIP(r.Context())
	if ip == "" {
		ip = r.RemoteAddr
	}

	var actor, detail, ipPtr *string
	if actorUserID != "" {
		actor = &actorUserID
	}
	if extensionName != "" {
		detail = &extensionName
	}
	if ip != "" {
		ipPtr = &ip
	}

	if err := h.queries.CreateAuditLog(r.Context(), queries.CreateAuditLogParams{
		ActorUserID: actor,
		Action:      action,
		Detail:      detail,
		IpAddress:   ipPtr,
	}); err != nil {
		slog.Error("failed to write audit log", "error", err, "action", action)
	}
}

// loadApprovedReportSummaries returns approved community report counts grouped
// by extension name, used to surface warnings on the extension lists. On error
// it returns nil so callers degrade gracefully (warnings are supplementary).
func (h *Handler) loadApprovedReportSummaries(ctx context.Context) map[string][]api.ExtensionReportSummary {
	rows, err := h.queries.GetApprovedExtensionReportSummary(ctx)
	if err != nil {
		slog.Error("failed to get approved extension report summaries", "error", err)
		return nil
	}
	summaries := make(map[string][]api.ExtensionReportSummary)
	for _, row := range rows {
		summaries[row.ExtensionName] = append(summaries[row.ExtensionName], api.ExtensionReportSummary{
			Category: row.Category,
			Count:    int(row.Count),
		})
	}
	return summaries
}
