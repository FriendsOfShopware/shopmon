package handler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/api"
	advisoryread "github.com/friendsofshopware/shopmon/api/internal/readmodel/advisory"
)

type listAdvisoriesInput struct {
	Limit    int    `query:"limit" default:"25"`
	Offset   int    `query:"offset" default:"0"`
	Package  string `query:"package"`
	Severity string `query:"severity"`
	Tag      string `query:"tag"`
	Q        string `query:"q"`
	Scope    string `query:"scope"`
	Sort     string `query:"sort"`
}

type listAdvisoriesOutput struct {
	Body api.AdvisoryListResponse
}

func (h *Handler) ListAdvisories(ctx context.Context, input *listAdvisoriesInput) (*listAdvisoriesOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	orgIDs, err := h.advisories.OrganizationIDsForUser(ctx, user.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load user organizations", "userId", user.ID, "error", err)
		return nil, huma.Error500InternalServerError("failed to list advisories")
	}

	// Default to the affected scope: the catalog is large and mostly irrelevant
	// to any single tenant, so "what hits my shops" is the useful landing view.
	// An unknown scope value falls back to the same default rather than
	// silently widening to the whole catalog.
	scope := api.ListAdvisoriesParamsScopeAffected
	if parsed := api.ListAdvisoriesParamsScope(input.Scope); parsed.Valid() {
		scope = parsed
	}
	onlyAffected := scope == api.ListAdvisoriesParamsScopeAffected
	onlySuppressed := scope == api.ListAdvisoriesParamsScopeSuppressed

	sort := "reported"
	if input.Sort != "" {
		sort = input.Sort
	}

	rows, counts, err := h.advisories.ListScoped(ctx, advisoryread.ScopedListParams{
		Limit:           int32(input.Limit),
		Offset:          int32(input.Offset),
		OrganizationIDs: orgIDs,
		OnlyAffected:    onlyAffected,
		OnlySuppressed:  onlySuppressed,
		Sort:            sort,
		PackageName:     optionalString(input.Package),
		Severity:        severityParamString(input.Severity),
		Tag:             optionalString(input.Tag),
		Search:          optionalString(input.Q),
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to list advisories", "error", err)
		return nil, huma.Error500InternalServerError("failed to list advisories")
	}

	items := make([]api.Advisory, 0, len(rows))
	for _, view := range rows {
		item, err := advisoryread.ToUserAdvisory(view)
		if err != nil {
			slog.ErrorContext(ctx, "failed to map advisory", "advisoryId", view.Advisory.AdvisoryID, "error", err)
			return nil, huma.Error500InternalServerError("failed to list advisories")
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

	return &listAdvisoriesOutput{Body: api.AdvisoryListResponse{
		Advisories: items,
		Total:      total,
		ScopeCounts: &api.AdvisoryScopeCounts{
			All:        counts.All,
			Affected:   counts.Affected,
			Suppressed: counts.Suppressed,
		},
	}}, nil
}

// annotateAffectedCounts fills in how many of the user's own environments each
// advisory affects. A failure here degrades the response to counts-unknown
// rather than failing the whole listing.
func (h *Handler) annotateAffectedCounts(ctx context.Context, userID string, views []advisoryread.AdvisoryView) {
	if len(views) == 0 {
		return
	}

	orgIDs, err := h.advisories.OrganizationIDsForUser(ctx, userID)
	if err != nil {
		slog.WarnContext(ctx, "failed to load organizations for advisory counts", "userId", userID, "error", err)
		return
	}
	if len(orgIDs) == 0 {
		return
	}

	ids := make([]string, 0, len(views))
	for _, v := range views {
		ids = append(ids, v.Advisory.AdvisoryID)
	}

	counts, err := h.advisories.AffectedCounts(ctx, ids, orgIDs)
	if err != nil {
		slog.WarnContext(ctx, "failed to count affected environments", "error", err)
		return
	}

	for i := range views {
		count := counts[views[i].Advisory.AdvisoryID]
		views[i].AffectedEnvironments = &count
	}
}

type advisoryIDInput struct {
	AdvisoryID string `path:"advisoryId"`
}

type getAdvisoryOutput struct {
	Body api.Advisory
}

func (h *Handler) GetAdvisory(ctx context.Context, input *advisoryIDInput) (*getAdvisoryOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	view, err := h.advisories.Get(ctx, input.AdvisoryID, true)
	if err != nil {
		if errors.Is(err, advisoryread.ErrNotFound) {
			return nil, huma.Error404NotFound("advisory not found")
		}
		slog.ErrorContext(ctx, "failed to get advisory", "advisoryId", input.AdvisoryID, "error", err)
		return nil, huma.Error500InternalServerError("failed to get advisory")
	}

	views := []advisoryread.AdvisoryView{view}
	h.annotateAffectedCounts(ctx, user.ID, views)

	item, err := advisoryread.ToUserAdvisory(views[0])
	if err != nil {
		slog.ErrorContext(ctx, "failed to map advisory", "advisoryId", input.AdvisoryID, "error", err)
		return nil, huma.Error500InternalServerError("failed to get advisory")
	}
	return &getAdvisoryOutput{Body: item}, nil
}

type listAdvisoryPackagesOutput struct {
	Body []string
}

func (h *Handler) ListAdvisoryPackages(ctx context.Context, _ *struct{}) (*listAdvisoryPackagesOutput, error) {
	if _, err := h.requireUser(ctx); err != nil {
		return nil, err
	}

	names, err := h.advisories.ListPackages(ctx, true)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list advisory packages", "error", err)
		return nil, huma.Error500InternalServerError("failed to list advisory packages")
	}
	return &listAdvisoryPackagesOutput{Body: names}, nil
}

type listAdvisoryAffectedEnvironmentsOutput struct {
	Body api.AdvisoryAffectedResponse
}

// ListAdvisoryAffectedEnvironments reports which of the caller's environments
// ship a package version covered by the advisory. Admins additionally get a
// fleet-wide count.
func (h *Handler) ListAdvisoryAffectedEnvironments(ctx context.Context, input *advisoryIDInput) (*listAdvisoryAffectedEnvironmentsOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	// Resolve the advisory first so an unknown id is a 404 rather than an empty
	// list, and so hidden advisories stay hidden from non-admins.
	if _, err := h.advisories.Get(ctx, input.AdvisoryID, user.Role != "admin"); err != nil {
		if errors.Is(err, advisoryread.ErrNotFound) {
			return nil, huma.Error404NotFound("advisory not found")
		}
		slog.ErrorContext(ctx, "failed to get advisory", "advisoryId", input.AdvisoryID, "error", err)
		return nil, huma.Error500InternalServerError("failed to load advisory")
	}

	orgIDs, err := h.advisories.OrganizationIDsForUser(ctx, user.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load user organizations", "userId", user.ID, "error", err)
		return nil, huma.Error500InternalServerError("failed to load affected environments")
	}

	result, err := h.advisories.ListAffected(ctx, input.AdvisoryID, orgIDs, user.Role == "admin")
	if err != nil {
		slog.ErrorContext(ctx, "failed to list affected environments", "advisoryId", input.AdvisoryID, "error", err)
		return nil, huma.Error500InternalServerError("failed to load affected environments")
	}

	return &listAdvisoryAffectedEnvironmentsOutput{Body: api.AdvisoryAffectedResponse{
		Environments: result.Environments,
		Total:        result.Total,
		GlobalTotal:  result.GlobalTotal,
	}}, nil
}

type adminListAdvisoriesInput struct {
	Limit    int               `query:"limit" default:"25"`
	Offset   int               `query:"offset" default:"0"`
	Package  string            `query:"package"`
	Severity string            `query:"severity"`
	Tag      string            `query:"tag"`
	Q        string            `query:"q"`
	Visible  optionalQueryBool `query:"visible" doc:"When set, filter by visibility"`
}

type adminListAdvisoriesOutput struct {
	Body api.AdminAdvisoryListResponse
}

func (h *Handler) AdminListAdvisories(ctx context.Context, input *adminListAdvisoriesInput) (*adminListAdvisoriesOutput, error) {
	if _, err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}

	rows, total, err := h.advisories.List(ctx, advisoryread.ListParams{
		Limit:       int32(input.Limit),
		Offset:      int32(input.Offset),
		PackageName: optionalString(input.Package),
		Severity:    adminSeverityParamString(input.Severity),
		Tag:         optionalString(input.Tag),
		Search:      optionalString(input.Q),
		IsVisible:   input.Visible.ptr(),
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to list admin advisories", "error", err)
		return nil, huma.Error500InternalServerError("failed to list advisories")
	}

	items := make([]api.AdminAdvisory, 0, len(rows))
	for _, view := range rows {
		item, err := advisoryread.ToAdminAdvisory(view)
		if err != nil {
			slog.ErrorContext(ctx, "failed to map admin advisory", "advisoryId", view.Advisory.AdvisoryID, "error", err)
			return nil, huma.Error500InternalServerError("failed to list advisories")
		}
		items = append(items, item)
	}

	return &adminListAdvisoriesOutput{Body: api.AdminAdvisoryListResponse{
		Advisories: items,
		Total:      int(total),
	}}, nil
}

type adminGetAdvisoryOutput struct {
	Body api.AdminAdvisory
}

func (h *Handler) AdminGetAdvisory(ctx context.Context, input *advisoryIDInput) (*adminGetAdvisoryOutput, error) {
	if _, err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}

	view, err := h.advisories.Get(ctx, input.AdvisoryID, false)
	if err != nil {
		if errors.Is(err, advisoryread.ErrNotFound) {
			return nil, huma.Error404NotFound("advisory not found")
		}
		slog.ErrorContext(ctx, "failed to get admin advisory", "advisoryId", input.AdvisoryID, "error", err)
		return nil, huma.Error500InternalServerError("failed to get advisory")
	}

	item, err := advisoryread.ToAdminAdvisory(view)
	if err != nil {
		slog.ErrorContext(ctx, "failed to map admin advisory", "advisoryId", input.AdvisoryID, "error", err)
		return nil, huma.Error500InternalServerError("failed to get advisory")
	}
	return &adminGetAdvisoryOutput{Body: item}, nil
}

type adminUpdateAdvisoryInput struct {
	AdvisoryID string `path:"advisoryId"`
	Body       api.UpdateAdvisoryEnrichmentRequest
}

type adminUpdateAdvisoryOutput struct {
	Body api.AdminAdvisory
}

func (h *Handler) AdminUpdateAdvisory(ctx context.Context, input *adminUpdateAdvisoryInput) (*adminUpdateAdvisoryOutput, error) {
	user, err := h.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}

	view, err := h.advisories.UpdateEnrichment(ctx, input.AdvisoryID, advisoryread.EnrichmentUpdate{
		SeverityOverride:      input.Body.SeverityOverride,
		IsVisible:             input.Body.IsVisible,
		NotesPublic:           input.Body.NotesPublic,
		NotesInternal:         input.Body.NotesInternal,
		RemediationSummary:    input.Body.RemediationSummary,
		RemediationUrl:        input.Body.RemediationUrl,
		RecommendedUpgrade:    input.Body.RecommendedUpgrade,
		ShopwareImpactSummary: input.Body.ShopwareImpactSummary,
		AffectedComponents:    input.Body.AffectedComponents,
		Tags:                  input.Body.Tags,
		EnrichedBy:            user.ID,
	})
	if err != nil {
		if errors.Is(err, advisoryread.ErrNotFound) {
			return nil, huma.Error404NotFound("advisory not found")
		}
		if errors.Is(err, advisoryread.ErrInvalidInput) {
			return nil, huma.Error400BadRequest(err.Error())
		}
		slog.ErrorContext(ctx, "failed to update advisory", "advisoryId", input.AdvisoryID, "error", err)
		return nil, huma.Error500InternalServerError("failed to update advisory")
	}

	item, err := advisoryread.ToAdminAdvisory(view)
	if err != nil {
		slog.ErrorContext(ctx, "failed to map admin advisory", "advisoryId", input.AdvisoryID, "error", err)
		return nil, huma.Error500InternalServerError("failed to update advisory")
	}
	return &adminUpdateAdvisoryOutput{Body: item}, nil
}

type adminSyncAdvisoriesOutput struct {
	Body api.AdminAdvisorySyncResponse
}

func (h *Handler) AdminSyncAdvisories(ctx context.Context, _ *struct{}) (*adminSyncAdvisoriesOutput, error) {
	if _, err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if h.advisorySync == nil {
		return nil, huma.Error503ServiceUnavailable("advisory sync is not available")
	}
	if err := h.advisorySync.EnqueueComposerAdvisorySync(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to enqueue advisory sync", "error", err)
		return nil, huma.Error500InternalServerError("failed to enqueue advisory sync")
	}
	return &adminSyncAdvisoriesOutput{Body: api.AdminAdvisorySyncResponse{Enqueued: true}}, nil
}

// optionalQueryBool is an optional boolean query parameter. Huma panics on
// pointer query fields, so omitted / true / false are represented here instead.
type optionalQueryBool struct {
	set   bool
	value bool
}

func (optionalQueryBool) Schema(huma.Registry) *huma.Schema {
	return &huma.Schema{Type: huma.TypeBoolean}
}

func (o *optionalQueryBool) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}
	switch string(text) {
	case "true", "1":
		o.set = true
		o.value = true
		return nil
	case "false", "0":
		o.set = true
		o.value = false
		return nil
	default:
		return errors.New("invalid boolean")
	}
}

func (o optionalQueryBool) ptr() *bool {
	if !o.set {
		return nil
	}
	v := o.value
	return &v
}

func optionalString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

// A severity outside the enum is ignored rather than forwarded: the query
// would match nothing, turning a typo into a misleading empty page.
func severityParamString(s string) *string {
	e := api.ListAdvisoriesParamsSeverity(s)
	if !e.Valid() {
		return nil
	}
	v := string(e)
	return &v
}

func adminSeverityParamString(s string) *string {
	e := api.AdminListAdvisoriesParamsSeverity(s)
	if !e.Valid() {
		return nil
	}
	v := string(e)
	return &v
}
