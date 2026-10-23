package handler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/access"
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
	var orgIDs []string
	user := access.UserFromContext(ctx)
	if user != nil {
		var err error
		orgIDs, err = h.advisories.OrganizationIDsForUser(ctx, user.ID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to load user organizations", "userId", user.ID, "error", err)
			return nil, huma.Error500InternalServerError("failed to list advisories")
		}
	}

	// Default to the affected scope for authenticated callers: the catalog is
	// large and mostly irrelevant to any single tenant, so "what hits my
	// shops" is the useful landing view. Anonymous callers have no shops, so
	// they always see the full catalog.
	scope := api.ListAdvisoriesParamsScopeAll
	if user != nil {
		scope = api.ListAdvisoriesParamsScopeAffected
		if parsed := api.ListAdvisoriesParamsScope(input.Scope); parsed.Valid() {
			scope = parsed
		}
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
	view, err := h.advisories.Get(ctx, input.AdvisoryID, true)
	if err != nil {
		if errors.Is(err, advisoryread.ErrNotFound) {
			return nil, huma.Error404NotFound("advisory not found")
		}
		slog.ErrorContext(ctx, "failed to get advisory", "advisoryId", input.AdvisoryID, "error", err)
		return nil, huma.Error500InternalServerError("failed to get advisory")
	}

	views := []advisoryread.AdvisoryView{view}
	if user := access.UserFromContext(ctx); user != nil {
		h.annotateAffectedCounts(ctx, user.ID, views)
	}

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
	Origin   string            `query:"origin" doc:"When set, filter by origin (shopmon or upstream)"`
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
		Origin:      originParamString(input.Origin),
		IsVisible:   input.Visible.ptr(),
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to list admin advisories", "error", err)
		return nil, huma.Error500InternalServerError("failed to list advisories")
	}

	items, err := h.advisories.AdminAdvisories(ctx, rows)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load admin advisory impact", "error", err)
		return nil, huma.Error500InternalServerError("failed to list advisories")
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

	item, err := h.advisories.AdminAdvisory(ctx, view)
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
		CoreFields:            coreFieldsFromRequest(input.Body),
	})
	if err != nil {
		if errors.Is(err, advisoryread.ErrNotFound) {
			return nil, huma.Error404NotFound("advisory not found")
		}
		if errors.Is(err, advisoryread.ErrConflict) {
			return nil, huma.Error409Conflict(err.Error())
		}
		if errors.Is(err, advisoryread.ErrInvalidInput) {
			return nil, huma.Error400BadRequest(err.Error())
		}
		slog.ErrorContext(ctx, "failed to update advisory", "advisoryId", input.AdvisoryID, "error", err)
		return nil, huma.Error500InternalServerError("failed to update advisory")
	}

	// Visibility flips and package-range edits change which shops match, so
	// re-evaluate inventories now instead of at the next hourly sync.
	if input.Body.IsVisible != nil || input.Body.Packages != nil {
		h.enqueueAdvisorySync(ctx, view.Advisory.AdvisoryID, "advisory updated")
	}

	item, err := h.advisories.AdminAdvisory(ctx, view)
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

type adminCreateAdvisoryInput struct {
	Body api.CreateAdvisoryRequest
}

type adminCreateAdvisoryOutput struct {
	Body api.AdminAdvisory
}

func (h *Handler) AdminCreateAdvisory(ctx context.Context, input *adminCreateAdvisoryInput) (*adminCreateAdvisoryOutput, error) {
	user, err := h.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}

	view, err := h.advisories.Create(ctx, createInputFromRequest(input.Body, user.ID))
	if err != nil {
		if errors.Is(err, advisoryread.ErrConflict) {
			return nil, huma.Error409Conflict(err.Error())
		}
		if errors.Is(err, advisoryread.ErrInvalidInput) {
			return nil, huma.Error400BadRequest(err.Error())
		}
		slog.ErrorContext(ctx, "failed to create advisory", "error", err)
		return nil, huma.Error500InternalServerError("failed to create advisory")
	}

	// A draft matches nobody; only a published advisory needs the rematch and
	// alert pass the advisory sync ends with.
	if view.Advisory.IsVisible {
		h.enqueueAdvisorySync(ctx, view.Advisory.AdvisoryID, "advisory published on create")
	}

	item, err := h.advisories.AdminAdvisory(ctx, view)
	if err != nil {
		slog.ErrorContext(ctx, "failed to map admin advisory", "advisoryId", view.Advisory.AdvisoryID, "error", err)
		return nil, huma.Error500InternalServerError("failed to create advisory")
	}
	return &adminCreateAdvisoryOutput{Body: item}, nil
}

func createInputFromRequest(body api.CreateAdvisoryRequest, userID string) advisoryread.CreateInput {
	return advisoryread.CreateInput{
		Title:                 body.Title,
		Packages:              packageInputsFromRequest(body.Packages),
		Severity:              body.Severity,
		Summary:               body.Summary,
		Description:           body.Description,
		CvssScore:             body.CvssScore,
		CvssVector:            body.CvssVector,
		Cwes:                  body.Cwes,
		Link:                  body.Link,
		ExternalReferences:    body.ExternalReferences,
		Cve:                   body.Cve,
		GhsaID:                body.GhsaId,
		ReportedAt:            body.ReportedAt,
		IsVisible:             body.IsVisible,
		NotesPublic:           body.NotesPublic,
		NotesInternal:         body.NotesInternal,
		RemediationSummary:    body.RemediationSummary,
		RemediationUrl:        body.RemediationUrl,
		RecommendedUpgrade:    body.RecommendedUpgrade,
		ShopwareImpactSummary: body.ShopwareImpactSummary,
		AffectedComponents:    body.AffectedComponents,
		Tags:                  body.Tags,
		CreatedBy:             userID,
	}
}

type adminPreviewGhsaImportInput struct {
	Body api.ImportGhsaPreviewRequest
}

type adminPreviewGhsaImportOutput struct {
	Body api.GhsaAdvisoryPreview
}

func (h *Handler) AdminPreviewGhsaImport(ctx context.Context, input *adminPreviewGhsaImportInput) (*adminPreviewGhsaImportOutput, error) {
	if _, err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}

	preview, err := h.advisories.PreviewGHSAImport(ctx, input.Body.Input)
	if err != nil {
		switch {
		case errors.Is(err, advisoryread.ErrGHSAImportUnavailable):
			return nil, huma.Error503ServiceUnavailable("ghsa import is not available")
		case errors.Is(err, advisoryread.ErrNotFound):
			return nil, huma.Error404NotFound(err.Error())
		case errors.Is(err, advisoryread.ErrGHSAFetchFailed):
			slog.WarnContext(ctx, "failed to fetch github advisory for import preview", "error", err)
			return nil, huma.Error502BadGateway("failed to fetch github advisory")
		case errors.Is(err, advisoryread.ErrInvalidInput):
			return nil, huma.Error400BadRequest(err.Error())
		default:
			slog.ErrorContext(ctx, "failed to preview ghsa import", "error", err)
			return nil, huma.Error500InternalServerError("failed to preview ghsa import")
		}
	}

	packages := make([]api.GhsaPreviewPackage, 0, len(preview.Packages))
	for _, pkg := range preview.Packages {
		var patched *string
		if pkg.PatchedVersion != "" {
			patched = &pkg.PatchedVersion
		}
		packages = append(packages, api.GhsaPreviewPackage{
			PackageName:      pkg.PackageName,
			AffectedVersions: pkg.AffectedVersions,
			PatchedVersion:   patched,
		})
	}
	var severity *api.SeverityLevel
	if preview.Severity != nil {
		level := api.SeverityLevel(*preview.Severity)
		severity = &level
	}
	cwes := preview.Cwes
	if cwes == nil {
		cwes = []api.AdvisoryCWE{}
	}
	refs := preview.ExternalReferences
	if refs == nil {
		refs = []string{}
	}
	var existing *string
	if preview.ExistingAdvisoryID != "" {
		existing = &preview.ExistingAdvisoryID
	}

	return &adminPreviewGhsaImportOutput{Body: api.GhsaAdvisoryPreview{
		GhsaId:             preview.GhsaID,
		Cve:                preview.Cve,
		Title:              preview.Title,
		Description:        preview.Description,
		Severity:           severity,
		CvssScore:          preview.CvssScore,
		CvssVector:         preview.CvssVector,
		Cwes:               cwes,
		Link:               preview.Link,
		ExternalReferences: refs,
		ReportedAt:         preview.ReportedAt,
		Packages:           packages,
		ExistingAdvisoryId: existing,
	}}, nil
}

type adminImportGhsaAdvisoryInput struct {
	Body api.CreateAdvisoryRequest
}

type adminImportGhsaAdvisoryOutput struct {
	Body api.AdminAdvisory
}

func (h *Handler) AdminImportGhsaAdvisory(ctx context.Context, input *adminImportGhsaAdvisoryInput) (*adminImportGhsaAdvisoryOutput, error) {
	user, err := h.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}

	view, err := h.advisories.ImportGHSA(ctx, createInputFromRequest(input.Body, user.ID))
	if err != nil {
		if errors.Is(err, advisoryread.ErrConflict) {
			return nil, huma.Error409Conflict(err.Error())
		}
		if errors.Is(err, advisoryread.ErrInvalidInput) {
			return nil, huma.Error400BadRequest(err.Error())
		}
		slog.ErrorContext(ctx, "failed to import ghsa advisory", "error", err)
		return nil, huma.Error500InternalServerError("failed to import advisory")
	}

	if view.Advisory.IsVisible {
		h.enqueueAdvisorySync(ctx, view.Advisory.AdvisoryID, "ghsa advisory published on import")
	}

	item, err := h.advisories.AdminAdvisory(ctx, view)
	if err != nil {
		slog.ErrorContext(ctx, "failed to map admin advisory", "advisoryId", view.Advisory.AdvisoryID, "error", err)
		return nil, huma.Error500InternalServerError("failed to import advisory")
	}
	return &adminImportGhsaAdvisoryOutput{Body: item}, nil
}

type adminLinkAdvisoryOfficialInput struct {
	AdvisoryID string `path:"advisoryId"`
	Body       api.LinkAdvisoryOfficialRequest
}

type adminLinkAdvisoryOfficialOutput struct {
	Body api.AdminAdvisory
}

func (h *Handler) AdminLinkAdvisoryOfficial(ctx context.Context, input *adminLinkAdvisoryOfficialInput) (*adminLinkAdvisoryOfficialOutput, error) {
	if _, err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}

	view, err := h.advisories.LinkOfficial(ctx, input.AdvisoryID, advisoryread.LinkOfficialInput{
		Cve:    input.Body.Cve,
		GhsaID: input.Body.GhsaId,
		Link:   input.Body.Link,
	})
	if err != nil {
		if errors.Is(err, advisoryread.ErrNotFound) {
			return nil, huma.Error404NotFound("advisory not found")
		}
		if errors.Is(err, advisoryread.ErrConflict) {
			return nil, huma.Error409Conflict(err.Error())
		}
		if errors.Is(err, advisoryread.ErrInvalidInput) {
			return nil, huma.Error400BadRequest(err.Error())
		}
		slog.ErrorContext(ctx, "failed to link advisory", "advisoryId", input.AdvisoryID, "error", err)
		return nil, huma.Error500InternalServerError("failed to link advisory")
	}

	// Linking can replace the package set with the official one, and it arms
	// the NVD/GitHub detail enrichment; both happen in the advisory sync.
	h.enqueueAdvisorySync(ctx, view.Advisory.AdvisoryID, "advisory linked to official")

	item, err := h.advisories.AdminAdvisory(ctx, view)
	if err != nil {
		slog.ErrorContext(ctx, "failed to map admin advisory", "advisoryId", input.AdvisoryID, "error", err)
		return nil, huma.Error500InternalServerError("failed to link advisory")
	}
	return &adminLinkAdvisoryOfficialOutput{Body: item}, nil
}

// enqueueAdvisorySync best-effort triggers the sync whose rematch pass
// re-evaluates shop inventories and alerts newly affected environments. A
// failure only delays that to the hourly scheduled sync, so it is logged
// rather than failing the admin write that already committed.
func (h *Handler) enqueueAdvisorySync(ctx context.Context, advisoryID, reason string) {
	if h.advisorySync == nil {
		return
	}
	if err := h.advisorySync.EnqueueComposerAdvisorySync(ctx); err != nil {
		slog.WarnContext(ctx, "failed to enqueue advisory sync after advisory change",
			"advisoryId", advisoryID, "reason", reason, "error", err)
	}
}

func packageInputsFromRequest(packages []api.AdvisoryPackageInput) []advisoryread.PackageInput {
	out := make([]advisoryread.PackageInput, 0, len(packages))
	for _, pkg := range packages {
		out = append(out, advisoryread.PackageInput{
			PackageName:      pkg.PackageName,
			AffectedVersions: pkg.AffectedVersions,
		})
	}
	return out
}

// coreFieldsFromRequest collects the content fields of an admin update, or
// returns nil when the request only carries enrichment — the distinction the
// service uses to apply the Shopmon-only editing guard.
func coreFieldsFromRequest(body api.UpdateAdvisoryEnrichmentRequest) *advisoryread.CoreFieldsUpdate {
	if body.Title == nil && body.Packages == nil && body.Severity == nil &&
		body.Summary == nil && body.Description == nil && !body.CvssScore.Sent &&
		body.CvssVector == nil && body.Cwes == nil && body.Link == nil &&
		body.ExternalReferences == nil && body.ReportedAt == nil {
		return nil
	}

	core := &advisoryread.CoreFieldsUpdate{
		Title:              body.Title,
		Severity:           body.Severity,
		Summary:            body.Summary,
		Description:        body.Description,
		CvssVector:         body.CvssVector,
		Cwes:               body.Cwes,
		Link:               body.Link,
		ExternalReferences: body.ExternalReferences,
		ReportedAt:         body.ReportedAt,
	}
	if body.CvssScore.Sent {
		core.CvssScore = &body.CvssScore.Value
	}
	if body.Packages != nil {
		packages := packageInputsFromRequest(*body.Packages)
		core.Packages = &packages
	}
	return core
}

// Like the severity filters, an unknown origin is ignored rather than
// forwarded: the query would match nothing and turn a typo into an empty page.
func originParamString(s string) *string {
	e := api.AdvisoryOrigin(s)
	if !e.Valid() {
		return nil
	}
	v := string(e)
	return &v
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
