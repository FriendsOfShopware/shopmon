package advisory

import (
	"context"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/api"
)

// ComposerFeed returns the complete public store advisory catalog without pagination.
func (s *Service) ComposerFeed(ctx context.Context) (api.ComposerAdvisoryResponse, error) {
	result := api.ComposerAdvisoryResponse{Advisories: make(map[string][]api.ComposerAdvisory)}
	rows, err := s.queries.ListPublishedStoreComposerAdvisories(ctx)
	if err != nil {
		return result, fmt.Errorf("list published store advisories: %w", err)
	}
	for _, row := range rows {
		sources, err := parseSources(row.Sources)
		if err != nil {
			return result, fmt.Errorf("advisory %s: %w", row.AdvisoryID, err)
		}
		if len(sources) == 0 {
			sources = []api.AdvisorySource{{Name: "Shopmon", RemoteId: row.AdvisoryID}}
		}
		severity := row.Severity
		if row.SeverityOverride != nil {
			severity = row.SeverityOverride
		}
		id := row.AdvisoryID
		if row.PackagistAdvisoryID != nil && *row.PackagistAdvisoryID != "" {
			id = *row.PackagistAdvisoryID
		}
		result.Advisories[row.PackageName] = append(result.Advisories[row.PackageName], api.ComposerAdvisory{
			AdvisoryID: id, PackageName: row.PackageName, AffectedVersions: row.AffectedVersions,
			Title: row.Title, CVE: row.Cve, Link: row.Link, ReportedAt: row.ReportedAt.Time.UTC(),
			Sources: sources, Severity: severity,
		})
	}
	return result, nil
}
