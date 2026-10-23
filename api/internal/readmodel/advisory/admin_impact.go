package advisory

import (
	"context"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/shopware/sbom"
)

// PreviewImpact checks submitted constraints without changing advisories,
// publishing matches, or sending alerts. It also works for unsaved edits.
func (s *Service) PreviewImpact(ctx context.Context, input []PackageInput) (api.AdminAdvisoryImpact, error) {
	result := api.AdminAdvisoryImpact{Shops: []api.AdminAdvisoryImpactShop{}}
	packages, err := normalizePackages(input)
	if err != nil {
		return result, err
	}
	advisoryPackages := make([]sbom.AdvisoryPackage, 0, len(packages))
	for _, pkg := range packages {
		advisoryPackages = append(advisoryPackages, sbom.AdvisoryPackage{
			PackageName: pkg.PackageName, AffectedVersions: pkg.AffectedVersions,
		})
	}
	index := sbom.NewIndex(advisoryPackages)
	inventory, err := s.queries.ListAdminAdvisoryInventory(ctx, index.PackageNames())
	if err != nil {
		return result, fmt.Errorf("load advisory impact inventory: %w", err)
	}
	shops := make(map[int32]int)
	environments := make(map[int32]bool)
	// Environments often share versions. Evaluate each package/version once.
	type packageVersion struct{ name, version string }
	cache := make(map[packageVersion][]sbom.Match)
	for _, row := range inventory {
		key := packageVersion{row.PackageName, row.Version}
		matches, ok := cache[key]
		if !ok {
			matches = index.Match(map[string]string{row.PackageName: row.Version})
			cache[key] = matches
		}
		if len(matches) == 0 {
			continue
		}
		shopIndex, ok := shops[row.ShopID]
		if !ok {
			shopIndex = len(result.Shops)
			shops[row.ShopID] = shopIndex
			result.Shops = append(result.Shops, api.AdminAdvisoryImpactShop{
				ShopId: int(row.ShopID), ShopName: row.ShopName,
				OrganizationId: row.OrganizationID, OrganizationName: row.OrganizationName,
				Matches: []api.AdminAdvisoryImpactMatch{},
			})
		}
		environments[row.EnvironmentID] = true
		for _, match := range matches {
			result.Shops[shopIndex].Matches = append(result.Shops[shopIndex].Matches, api.AdminAdvisoryImpactMatch{
				EnvironmentId: int(row.EnvironmentID), EnvironmentName: row.EnvironmentName,
				PackageName: match.PackageName, InstalledVersion: match.InstalledVersion,
				AffectedVersions: match.AffectedVersions,
			})
		}
	}
	result.ShopCount = len(result.Shops)
	result.EnvironmentCount = len(environments)
	return result, nil
}
