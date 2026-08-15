package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/suppression"
)

// CreateAdvisorySuppression records that an advisory is accepted or mitigated
// for a shop, removing it from the "affecting my shops" list and silencing its
// alerts.
func (h *Handler) CreateAdvisorySuppression(w http.ResponseWriter, r *http.Request, advisoryID api.AdvisoryId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	var req api.CreateAdvisorySuppressionRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var environmentID *int32
	if req.EnvironmentId != nil {
		id := int32(*req.EnvironmentId)
		environmentID = &id
	}

	created, err := h.suppressions.Create(r.Context(), suppression.CreateCommand{
		UserID:        user.ID,
		ShopID:        int32(req.ShopId),
		EnvironmentID: environmentID,
		AdvisoryID:    advisoryID,
		Reason:        req.Reason,
		ExpiresAt:     req.ExpiresAt,
	})
	if err != nil {
		writeSuppressionError(w, r, err)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, toAPISuppression(created))
}

// RevokeAdvisorySuppression lifts a suppression so the advisory resurfaces. The
// row is soft-deleted, preserving who accepted the risk and why.
func (h *Handler) RevokeAdvisorySuppression(w http.ResponseWriter, r *http.Request, suppressionID int64) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	orgIDs, err := h.advisories.OrganizationIDsForUser(r.Context(), user.ID)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to load user organizations", "userId", user.ID, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to revoke suppression")
		return
	}

	if _, err := h.suppressions.Revoke(r.Context(), suppressionID, user.ID, orgIDs); err != nil {
		writeSuppressionError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListSuppressions returns the suppressions across the caller's organizations.
func (h *Handler) ListSuppressions(w http.ResponseWriter, r *http.Request, params api.ListSuppressionsParams) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	orgIDs, err := h.advisories.OrganizationIDsForUser(r.Context(), user.ID)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to load user organizations", "userId", user.ID, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list suppressions")
		return
	}

	includeInactive := params.IncludeInactive != nil && *params.IncludeInactive
	records, err := h.suppressions.List(r.Context(), orgIDs, includeInactive)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to list suppressions", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list suppressions")
		return
	}

	items := make([]api.AdvisorySuppression, 0, len(records))
	for _, record := range records {
		items = append(items, toAPISuppression(record))
	}
	httputil.WriteJSON(w, http.StatusOK, api.AdvisorySuppressionListResponse{Suppressions: items})
}

func toAPISuppression(record suppression.Suppression) api.AdvisorySuppression {
	item := api.AdvisorySuppression{
		Id:             record.ID,
		AdvisoryId:     record.AdvisoryID,
		OrganizationId: record.OrganizationID,
		ShopId:         int(record.ShopID),
		ShopName:       record.ShopName,
		Reason:         record.Reason,
		ExpiresAt:      record.ExpiresAt,
		CreatedBy:      record.CreatedBy,
		CreatedAt:      record.CreatedAt,
		RevokedAt:      record.RevokedAt,
	}
	if record.EnvironmentID != nil {
		id := int(*record.EnvironmentID)
		item.EnvironmentId = &id
	}
	if record.EnvironmentName != nil && *record.EnvironmentName != "" {
		item.EnvironmentName = record.EnvironmentName
	}
	if record.CreatedByName != nil && *record.CreatedByName != "" {
		item.CreatedByName = record.CreatedByName
	}
	return item
}

// writeSuppressionError maps capability errors onto transport status codes.
// A duplicate is a 409 rather than a silent overwrite, which is the behaviour
// the legacy whole-array ignores PATCH could not offer.
func writeSuppressionError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, suppression.ErrReasonRequired):
		httputil.WriteError(w, http.StatusBadRequest, "a reason is required")
	case errors.Is(err, suppression.ErrExpiryInPast):
		httputil.WriteError(w, http.StatusBadRequest, "expiry must be in the future")
	case errors.Is(err, suppression.ErrEnvironmentScope):
		httputil.WriteError(w, http.StatusBadRequest, "environment does not belong to shop")
	case errors.Is(err, suppression.ErrNotAuthorized):
		httputil.WriteError(w, http.StatusForbidden, "not authorized to suppress for this shop")
	case errors.Is(err, suppression.ErrShopNotFound):
		httputil.WriteError(w, http.StatusNotFound, "shop not found")
	case errors.Is(err, suppression.ErrAdvisoryNotFound):
		httputil.WriteError(w, http.StatusNotFound, "advisory not found")
	case errors.Is(err, suppression.ErrNotFound):
		httputil.WriteError(w, http.StatusNotFound, "suppression not found")
	case errors.Is(err, suppression.ErrAlreadySuppressed):
		httputil.WriteError(w, http.StatusConflict, "an active suppression already covers this scope")
	default:
		slog.ErrorContext(r.Context(), "suppression request failed", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to process suppression")
	}
}
