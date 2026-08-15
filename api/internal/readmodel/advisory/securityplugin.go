package advisory

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/api"
)

// securityPluginFixesByGhsa loads the backport map for a set of advisories in
// one query, following the same batching shape as packagesByAdvisoryIDs so a
// list page never issues one query per row.
func (s *Service) securityPluginFixesByGhsa(ctx context.Context, ghsaIDs []string) (map[string][]api.AdvisorySecurityPluginFix, error) {
	if len(ghsaIDs) == 0 {
		return map[string][]api.AdvisorySecurityPluginFix{}, nil
	}

	rows, err := s.queries.ListSecurityPluginFixesForGhsas(ctx, ghsaIDs)
	if err != nil {
		return nil, fmt.Errorf("list security plugin fixes: %w", err)
	}

	out := make(map[string][]api.AdvisorySecurityPluginFix, len(rows))
	for _, row := range rows {
		out[row.GhsaID] = append(out[row.GhsaID], api.AdvisorySecurityPluginFix{
			PluginBranch:   row.PluginBranch,
			PluginVersion:  row.PluginVersion,
			ShopwareBranch: row.ShopwareBranch,
		})
	}
	return out, nil
}

// parseFirstPatchedVersions decodes the machine-derived {line: version} map.
// A malformed value yields nil rather than an error: the remediation hint is
// advisory, and losing it must never fail an advisory listing.
func parseFirstPatchedVersions(raw json.RawMessage) *map[string]string {
	if len(raw) == 0 {
		return nil
	}
	var parsed map[string]string
	if err := json.Unmarshal(raw, &parsed); err != nil || len(parsed) == 0 {
		return nil
	}
	return &parsed
}

// ghsaIDsOf collects the non-empty GHSA ids across advisory views.
func ghsaIDsOf(views []AdvisoryView) []string {
	out := make([]string, 0, len(views))
	seen := make(map[string]bool, len(views))
	for _, view := range views {
		if view.Advisory.GhsaID == nil || *view.Advisory.GhsaID == "" {
			continue
		}
		if seen[*view.Advisory.GhsaID] {
			continue
		}
		seen[*view.Advisory.GhsaID] = true
		out = append(out, *view.Advisory.GhsaID)
	}
	return out
}
