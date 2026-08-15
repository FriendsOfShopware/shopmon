package advisory

import (
	"context"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
)

// AffectedResult is the org-scoped list of environments matching an advisory,
// optionally accompanied by a fleet-wide count for admins.
type AffectedResult struct {
	Environments []api.AdvisoryAffectedEnvironment
	Total        int
	GlobalTotal  *int
}

// OrganizationIDsForUser returns the organizations the user belongs to, used to
// scope advisory exposure to the caller's own shops.
func (s *Service) OrganizationIDsForUser(ctx context.Context, userID string) ([]string, error) {
	orgs, err := s.queries.GetUserOrganizations(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user organizations: %w", err)
	}
	out := make([]string, 0, len(orgs))
	for _, org := range orgs {
		out = append(out, org.ID)
	}
	return out, nil
}

// ListAffected returns the environments affected by an advisory.
//
// organizationIDs restricts the result to the caller's tenants: advisories are
// global data, but which shops are exposed is not. Passing includeGlobalTotal
// (admins only) additionally reports the fleet-wide affected count without
// leaking any other tenant's shop names.
func (s *Service) ListAffected(ctx context.Context, advisoryID string, organizationIDs []string, includeGlobalTotal bool) (AffectedResult, error) {
	var result AffectedResult

	if len(organizationIDs) > 0 {
		rows, err := s.queries.ListAffectedEnvironmentsForAdvisory(ctx, queries.ListAffectedEnvironmentsForAdvisoryParams{
			AdvisoryID: advisoryID,
			Column2:    organizationIDs,
		})
		if err != nil {
			return result, fmt.Errorf("list affected environments: %w", err)
		}

		envIDs := make(map[int32]bool, len(rows))
		result.Environments = make([]api.AdvisoryAffectedEnvironment, 0, len(rows))
		for _, row := range rows {
			envIDs[row.EnvironmentID] = true
			result.Environments = append(result.Environments, api.AdvisoryAffectedEnvironment{
				EnvironmentId:     int(row.EnvironmentID),
				EnvironmentName:   row.EnvironmentName,
				EnvironmentStatus: row.EnvironmentStatus,
				ShopId:            int(row.ShopID),
				ShopName:          row.ShopName,
				OrganizationId:    row.OrganizationID,
				OrganizationName:  row.OrganizationName,
				ShopwareVersion:   nilIfEmpty(row.ShopwareVersion),
				PackageName:       row.PackageName,
				InstalledVersion:  row.InstalledVersion,
				AffectedVersions:  row.AffectedVersions,
				Suppressed:        row.Suppressed,
				MatchedAt:         row.MatchedAt.Time,
			})
		}
		// One environment can match through several packages; the headline
		// count is distinct environments, not match rows.
		result.Total = len(envIDs)
	} else {
		result.Environments = []api.AdvisoryAffectedEnvironment{}
	}

	if includeGlobalTotal {
		counts, err := s.queries.CountAffectedEnvironmentsForAdvisories(ctx, []string{advisoryID})
		if err != nil {
			return result, fmt.Errorf("count affected environments: %w", err)
		}
		total := 0
		for _, c := range counts {
			if c.AdvisoryID == advisoryID {
				total = int(c.EnvironmentCount)
			}
		}
		result.GlobalTotal = &total
	}

	return result, nil
}

// ScopedListParams filters the user-facing advisory list.
type ScopedListParams struct {
	Limit           int32
	Offset          int32
	OrganizationIDs []string
	// OnlyAffected restricts results to advisories matching the caller's own
	// environment inventory, excluding suppressed ones.
	OnlyAffected bool
	// OnlySuppressed restricts results to advisories the caller has
	// acknowledged. Mutually exclusive with OnlyAffected.
	OnlySuppressed bool
	// Sort is one of reported (default), severity, affected, cvss.
	Sort        string
	PackageName *string
	Severity    *string
	Tag         *string
	Search      *string
}

// ScopeCounts are the per-scope totals behind the list's scope tabs.
type ScopeCounts struct {
	All        int
	Affected   int
	Suppressed int
}

var allowedSorts = map[string]bool{
	"reported": true,
	"severity": true,
	"affected": true,
	"cvss":     true,
}

// ListScoped returns advisories with their affected-environment counts, and the
// per-scope totals used for the scope tabs.
//
// A user with no organizations can have no affected advisories, so the affected
// scope is empty for them; the counts still report the full catalog size so the
// UI can steer them to the "all" tab.
func (s *Service) ListScoped(ctx context.Context, params ScopedListParams) ([]AdvisoryView, ScopeCounts, error) {
	if params.Limit <= 0 {
		params.Limit = 25
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Offset < 0 {
		params.Offset = 0
	}
	if !allowedSorts[params.Sort] {
		params.Sort = "reported"
	}
	// ANY($1::text[]) needs a non-nil array; an empty one matches nothing,
	// which is the correct result for a user without organizations.
	orgIDs := params.OrganizationIDs
	if orgIDs == nil {
		orgIDs = []string{}
	}

	rows, err := s.queries.ListComposerAdvisoriesScoped(ctx, queries.ListComposerAdvisoriesScopedParams{
		OrganizationIds: orgIDs,
		OnlyAffected:    params.OnlyAffected,
		OnlySuppressed:  params.OnlySuppressed,
		PackageName:     emptyToNil(params.PackageName),
		Severity:        emptyToNil(params.Severity),
		Tag:             emptyToNil(params.Tag),
		Search:          emptyToNil(params.Search),
		Sort:            params.Sort,
		Limit:           params.Limit,
		Offset:          params.Offset,
	})
	if err != nil {
		return nil, ScopeCounts{}, fmt.Errorf("list scoped advisories: %w", err)
	}

	counts, err := s.queries.CountComposerAdvisoriesScoped(ctx, queries.CountComposerAdvisoriesScopedParams{
		OrganizationIds: orgIDs,
		PackageName:     emptyToNil(params.PackageName),
		Severity:        emptyToNil(params.Severity),
		Tag:             emptyToNil(params.Tag),
		Search:          emptyToNil(params.Search),
	})
	if err != nil {
		return nil, ScopeCounts{}, fmt.Errorf("count scoped advisories: %w", err)
	}

	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.AdvisoryID)
	}
	pkgsByID, err := s.packagesByAdvisoryIDs(ctx, ids)
	if err != nil {
		return nil, ScopeCounts{}, err
	}

	views := make([]AdvisoryView, 0, len(rows))
	for _, row := range rows {
		count := int(row.AffectedCount)
		suppressed := int(row.SuppressedCount)
		views = append(views, AdvisoryView{
			Advisory:               scopedRowToAdvisory(row),
			Packages:               pkgsByID[row.AdvisoryID],
			AffectedEnvironments:   &count,
			SuppressedEnvironments: &suppressed,
		})
	}

	// One batched lookup for the whole page, keyed by GHSA.
	if fixes, err := s.securityPluginFixesByGhsa(ctx, ghsaIDsOf(views)); err == nil {
		for i := range views {
			if ghsa := views[i].Advisory.GhsaID; ghsa != nil {
				views[i].SecurityPluginFixes = fixes[*ghsa]
			}
		}
	}

	return views, ScopeCounts{
		All:        int(counts.AllCount),
		Affected:   int(counts.AffectedCount),
		Suppressed: int(counts.SuppressedCount),
	}, nil
}

// scopedRowToAdvisory drops the joined affected_count so the row can reuse the
// shared advisory mapper.
func scopedRowToAdvisory(row queries.ListComposerAdvisoriesScopedRow) queries.ComposerAdvisory {
	return queries.ComposerAdvisory{
		AdvisoryID:            row.AdvisoryID,
		Title:                 row.Title,
		Link:                  row.Link,
		Cve:                   row.Cve,
		GhsaID:                row.GhsaID,
		Severity:              row.Severity,
		Sources:               row.Sources,
		ReportedAt:            row.ReportedAt,
		ComposerRepository:    row.ComposerRepository,
		SyncedAt:              row.SyncedAt,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
		Summary:               row.Summary,
		Description:           row.Description,
		CvssScore:             row.CvssScore,
		CvssVector:            row.CvssVector,
		Cwes:                  row.Cwes,
		ExternalReferences:    row.ExternalReferences,
		DetailsSource:         row.DetailsSource,
		DetailsSyncedAt:       row.DetailsSyncedAt,
		SeverityOverride:      row.SeverityOverride,
		IsVisible:             row.IsVisible,
		NotesPublic:           row.NotesPublic,
		NotesInternal:         row.NotesInternal,
		RemediationSummary:    row.RemediationSummary,
		RemediationUrl:        row.RemediationUrl,
		RecommendedUpgrade:    row.RecommendedUpgrade,
		ShopwareImpactSummary: row.ShopwareImpactSummary,
		AffectedComponents:    row.AffectedComponents,
		Tags:                  row.Tags,
		EnrichedAt:            row.EnrichedAt,
		EnrichedBy:            row.EnrichedBy,
		FirstPatchedVersions:  row.FirstPatchedVersions,
	}
}

// AffectedCounts returns, per advisory id, how many of the caller's
// environments are affected. Used to annotate advisory lists.
func (s *Service) AffectedCounts(ctx context.Context, advisoryIDs, organizationIDs []string) (map[string]int, error) {
	if len(advisoryIDs) == 0 || len(organizationIDs) == 0 {
		return map[string]int{}, nil
	}

	rows, err := s.queries.CountAffectedEnvironmentsForAdvisoriesInOrgs(ctx, queries.CountAffectedEnvironmentsForAdvisoriesInOrgsParams{
		Column1: advisoryIDs,
		Column2: organizationIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("count affected environments in orgs: %w", err)
	}

	out := make(map[string]int, len(rows))
	for _, row := range rows {
		out[row.AdvisoryID] = int(row.EnvironmentCount)
	}
	return out, nil
}
