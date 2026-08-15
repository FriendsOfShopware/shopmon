package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	advisoryread "github.com/friendsofshopware/shopmon/api/internal/readmodel/advisory"
)

func (h *Handler) ListAdvisories(w http.ResponseWriter, r *http.Request, params api.ListAdvisoriesParams) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	orgIDs, err := h.advisories.OrganizationIDsForUser(r.Context(), user.ID)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to load user organizations", "userId", user.ID, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list advisories")
		return
	}

	// Default to the affected scope: the catalog is large and mostly irrelevant
	// to any single tenant, so "what hits my shops" is the useful landing view.
	// An unknown scope value falls back to the same default rather than
	// silently widening to the whole catalog.
	scope := api.ListAdvisoriesParamsScopeAffected
	if params.Scope != nil && params.Scope.Valid() {
		scope = *params.Scope
	}
	onlyAffected := scope == api.ListAdvisoriesParamsScopeAffected
	onlySuppressed := scope == api.ListAdvisoriesParamsScopeSuppressed

	sort := "reported"
	if params.Sort != nil {
		sort = string(*params.Sort)
	}

	rows, counts, err := h.advisories.ListScoped(r.Context(), advisoryread.ScopedListParams{
		Limit:           int32(ptrIntOr(params.Limit, 25)),
		Offset:          int32(ptrIntOr(params.Offset, 0)),
		OrganizationIDs: orgIDs,
		OnlyAffected:    onlyAffected,
		OnlySuppressed:  onlySuppressed,
		Sort:            sort,
		PackageName:     params.Package,
		Severity:        severityParamString(params.Severity),
		Tag:             params.Tag,
		Search:          params.Q,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to list advisories", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list advisories")
		return
	}

	items := make([]api.Advisory, 0, len(rows))
	for _, view := range rows {
		item, err := advisoryread.ToUserAdvisory(view)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to map advisory", "advisoryId", view.Advisory.AdvisoryID, "error", err)
			httputil.WriteError(w, http.StatusInternalServerError, "failed to list advisories")
			return
		}
		items = append(items, item)
	}

	total := counts.All
	switch {
	case onlyAffected:
		total = counts.Affected
	case onlySuppressed:
		total = counts.Suppressed
	}

	httputil.WriteJSON(w, http.StatusOK, api.AdvisoryListResponse{
		Advisories: items,
		Total:      total,
		ScopeCounts: &api.AdvisoryScopeCounts{
			All:        counts.All,
			Affected:   counts.Affected,
			Suppressed: counts.Suppressed,
		},
	})
}

// annotateAffectedCounts fills in how many of the user's own environments each
// advisory affects. A failure here degrades the response to counts-unknown
// rather than failing the whole listing.
func (h *Handler) annotateAffectedCounts(r *http.Request, userID string, views []advisoryread.AdvisoryView) {
	if len(views) == 0 {
		return
	}

	orgIDs, err := h.advisories.OrganizationIDsForUser(r.Context(), userID)
	if err != nil {
		slog.WarnContext(r.Context(), "failed to load organizations for advisory counts", "userId", userID, "error", err)
		return
	}
	if len(orgIDs) == 0 {
		return
	}

	ids := make([]string, 0, len(views))
	for _, v := range views {
		ids = append(ids, v.Advisory.AdvisoryID)
	}

	counts, err := h.advisories.AffectedCounts(r.Context(), ids, orgIDs)
	if err != nil {
		slog.WarnContext(r.Context(), "failed to count affected environments", "error", err)
		return
	}

	for i := range views {
		count := counts[views[i].Advisory.AdvisoryID]
		views[i].AffectedEnvironments = &count
	}
}

func (h *Handler) GetAdvisory(w http.ResponseWriter, r *http.Request, advisoryID api.AdvisoryId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	view, err := h.advisories.Get(r.Context(), advisoryID, true)
	if err != nil {
		if errors.Is(err, advisoryread.ErrNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "advisory not found")
			return
		}
		slog.ErrorContext(r.Context(), "failed to get advisory", "advisoryId", advisoryID, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get advisory")
		return
	}

	views := []advisoryread.AdvisoryView{view}
	h.annotateAffectedCounts(r, user.ID, views)

	item, err := advisoryread.ToUserAdvisory(views[0])
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to map advisory", "advisoryId", advisoryID, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get advisory")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) ListAdvisoryPackages(w http.ResponseWriter, r *http.Request) {
	if h.requireUser(w, r) == nil {
		return
	}

	names, err := h.advisories.ListPackages(r.Context(), true)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to list advisory packages", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list advisory packages")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, names)
}

// ListAdvisoryAffectedEnvironments reports which of the caller's environments
// ship a package version covered by the advisory. Admins additionally get a
// fleet-wide count.
func (h *Handler) ListAdvisoryAffectedEnvironments(w http.ResponseWriter, r *http.Request, advisoryID api.AdvisoryId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	// Resolve the advisory first so an unknown id is a 404 rather than an empty
	// list, and so hidden advisories stay hidden from non-admins.
	if _, err := h.advisories.Get(r.Context(), advisoryID, user.Role != "admin"); err != nil {
		if errors.Is(err, advisoryread.ErrNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "advisory not found")
			return
		}
		slog.ErrorContext(r.Context(), "failed to get advisory", "advisoryId", advisoryID, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load advisory")
		return
	}

	orgIDs, err := h.advisories.OrganizationIDsForUser(r.Context(), user.ID)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to load user organizations", "userId", user.ID, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load affected environments")
		return
	}

	result, err := h.advisories.ListAffected(r.Context(), advisoryID, orgIDs, user.Role == "admin")
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to list affected environments", "advisoryId", advisoryID, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load affected environments")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, api.AdvisoryAffectedResponse{
		Environments: result.Environments,
		Total:        result.Total,
		GlobalTotal:  result.GlobalTotal,
	})
}

func (h *Handler) AdminListAdvisories(w http.ResponseWriter, r *http.Request, params api.AdminListAdvisoriesParams) {
	if !h.requireAdmin(w, r) {
		return
	}

	rows, total, err := h.advisories.List(r.Context(), advisoryread.ListParams{
		Limit:       int32(ptrIntOr(params.Limit, 25)),
		Offset:      int32(ptrIntOr(params.Offset, 0)),
		PackageName: params.Package,
		Severity:    adminSeverityParamString(params.Severity),
		Tag:         params.Tag,
		Search:      params.Q,
		IsVisible:   params.Visible,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to list admin advisories", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list advisories")
		return
	}

	items := make([]api.AdminAdvisory, 0, len(rows))
	for _, view := range rows {
		item, err := advisoryread.ToAdminAdvisory(view)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to map admin advisory", "advisoryId", view.Advisory.AdvisoryID, "error", err)
			httputil.WriteError(w, http.StatusInternalServerError, "failed to list advisories")
			return
		}
		items = append(items, item)
	}

	httputil.WriteJSON(w, http.StatusOK, api.AdminAdvisoryListResponse{
		Advisories: items,
		Total:      int(total),
	})
}

func (h *Handler) AdminGetAdvisory(w http.ResponseWriter, r *http.Request, advisoryID api.AdvisoryId) {
	if !h.requireAdmin(w, r) {
		return
	}

	view, err := h.advisories.Get(r.Context(), advisoryID, false)
	if err != nil {
		if errors.Is(err, advisoryread.ErrNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "advisory not found")
			return
		}
		slog.ErrorContext(r.Context(), "failed to get admin advisory", "advisoryId", advisoryID, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get advisory")
		return
	}

	item, err := advisoryread.ToAdminAdvisory(view)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to map admin advisory", "advisoryId", advisoryID, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get advisory")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) AdminUpdateAdvisory(w http.ResponseWriter, r *http.Request, advisoryID api.AdvisoryId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if user.Role != "admin" {
		httputil.WriteError(w, http.StatusForbidden, httputil.MsgAdminRequired)
		return
	}

	var body api.UpdateAdvisoryEnrichmentRequest
	if err := httputil.DecodeBody(r, &body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var severityOverride *string
	if body.SeverityOverride != nil {
		s := string(*body.SeverityOverride)
		severityOverride = &s
	}

	view, err := h.advisories.UpdateEnrichment(r.Context(), advisoryID, advisoryread.EnrichmentUpdate{
		SeverityOverride:      severityOverride,
		IsVisible:             body.IsVisible,
		NotesPublic:           body.NotesPublic,
		NotesInternal:         body.NotesInternal,
		RemediationSummary:    body.RemediationSummary,
		RemediationUrl:        body.RemediationUrl,
		RecommendedUpgrade:    body.RecommendedUpgrade,
		ShopwareImpactSummary: body.ShopwareImpactSummary,
		AffectedComponents:    body.AffectedComponents,
		Tags:                  body.Tags,
		EnrichedBy:            user.ID,
	})
	if err != nil {
		if errors.Is(err, advisoryread.ErrNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "advisory not found")
			return
		}
		if errors.Is(err, advisoryread.ErrInvalidInput) {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		slog.ErrorContext(r.Context(), "failed to update advisory", "advisoryId", advisoryID, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update advisory")
		return
	}

	item, err := advisoryread.ToAdminAdvisory(view)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to map admin advisory", "advisoryId", advisoryID, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update advisory")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) AdminSyncAdvisories(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	if h.advisorySync == nil {
		httputil.WriteError(w, http.StatusServiceUnavailable, "advisory sync is not available")
		return
	}
	if err := h.advisorySync.EnqueueComposerAdvisorySync(r.Context()); err != nil {
		slog.ErrorContext(r.Context(), "failed to enqueue advisory sync", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to enqueue advisory sync")
		return
	}
	httputil.WriteJSON(w, http.StatusAccepted, api.AdminAdvisorySyncResponse{Enqueued: true})
}

func ptrIntOr(v *int, fallback int) int {
	if v == nil {
		return fallback
	}
	return *v
}

// A severity outside the enum is ignored rather than forwarded: the query
// would match nothing, turning a typo into a misleading empty page.
func severityParamString(s *api.ListAdvisoriesParamsSeverity) *string {
	if s == nil || !s.Valid() {
		return nil
	}
	v := string(*s)
	return &v
}

func adminSeverityParamString(s *api.AdminListAdvisoriesParamsSeverity) *string {
	if s == nil || !s.Valid() {
		return nil
	}
	v := string(*s)
	return &v
}
