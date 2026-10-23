package advisory

import (
	"context"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/shopware/sbom"
)

// AdminAdvisories maps a page in batches. Impact is evaluated against collected
// inventory with the same matcher as background jobs, including hidden drafts.
// It is read-only: previewing impact never publishes matches or sends alerts.
func (s *Service) AdminAdvisories(ctx context.Context, views []AdvisoryView) ([]api.AdminAdvisory, error) {
	items := make([]api.AdminAdvisory, 0, len(views))
	ids := make([]string, 0, len(views))
	var packages []sbom.AdvisoryPackage
	for _, view := range views {
		item, err := ToAdminAdvisory(view)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
		ids = append(ids, view.Advisory.AdvisoryID)
		for _, pkg := range view.Packages {
			packages = append(packages, sbom.AdvisoryPackage{
				AdvisoryID: view.Advisory.AdvisoryID, PackageName: pkg.PackageName, AffectedVersions: pkg.AffectedVersions,
			})
		}
	}
	if len(items) == 0 {
		return items, nil
	}

	index := sbom.NewIndex(packages)
	inventory, err := s.queries.ListAdminAdvisoryInventory(ctx, index.PackageNames())
	if err != nil {
		return nil, fmt.Errorf("load admin advisory inventory: %w", err)
	}
	suppressions, err := s.queries.ListActiveSuppressionsForAdvisories(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("load admin advisory suppressions: %w", err)
	}
	type suppressionKey struct {
		advisoryID    string
		shopID        int32
		environmentID int32 // zero denotes a shop-wide suppression
	}
	suppressed := make(map[suppressionKey]bool, len(suppressions))
	for _, row := range suppressions {
		var environmentID int32
		if row.EnvironmentID != nil {
			environmentID = *row.EnvironmentID
		}
		suppressed[suppressionKey{row.AdvisoryID, row.ShopID, environmentID}] = true
	}

	type impactSets struct {
		shops, environments, activeShops, suppressedEnvironments map[int32]bool
	}
	counts := make(map[string]*impactSets, len(views))
	for _, id := range ids {
		counts[id] = &impactSets{
			shops: make(map[int32]bool), environments: make(map[int32]bool),
			activeShops: make(map[int32]bool), suppressedEnvironments: make(map[int32]bool),
		}
	}
	// Environments often share the same package versions. Match each distinct
	// package/version once per page, then deduplicate shops and environments.
	type packageVersion struct{ name, version string }
	cache := make(map[packageVersion][]sbom.Match)
	for _, row := range inventory {
		key := packageVersion{row.PackageName, row.Version}
		matches, ok := cache[key]
		if !ok {
			matches = index.Match(map[string]string{row.PackageName: row.Version})
			cache[key] = matches
		}
		for _, match := range matches {
			count := counts[match.AdvisoryID]
			count.shops[row.ShopID] = true
			count.environments[row.EnvironmentID] = true
			if suppressed[suppressionKey{match.AdvisoryID, row.ShopID, 0}] || suppressed[suppressionKey{match.AdvisoryID, row.ShopID, row.EnvironmentID}] {
				count.suppressedEnvironments[row.EnvironmentID] = true
			} else {
				count.activeShops[row.ShopID] = true
			}
		}
	}
	for i := range items {
		count := counts[items[i].AdvisoryId]
		items[i].Impact = api.AdminAdvisoryImpact{
			ShopCount: len(count.shops), EnvironmentCount: len(count.environments),
			ActiveShopCount: len(count.activeShops), SuppressedEnvironmentCount: len(count.suppressedEnvironments),
		}
	}
	return items, nil
}

func (s *Service) AdminAdvisory(ctx context.Context, view AdvisoryView) (api.AdminAdvisory, error) {
	items, err := s.AdminAdvisories(ctx, []AdvisoryView{view})
	if err != nil {
		return api.AdminAdvisory{}, err
	}
	return items[0], nil
}
