package account

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/ptr"
	"github.com/friendsofshopware/shopmon/api/internal/shopware"
	"github.com/friendsofshopware/shopmon/api/internal/version"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

var ErrExtensionNotFound = errors.New("extension not found")

type Service struct {
	queries *queries.Queries
}

func NewService(q *queries.Queries) *Service {
	return &Service{queries: q}
}

func (s *Service) Profile(ctx context.Context, userID string) (api.UserProfile, error) {
	dbUser, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return api.UserProfile{}, fmt.Errorf("get user: %w", err)
	}
	return api.UserProfile{
		Id:          dbUser.ID,
		DisplayName: dbUser.Name,
		Email:       openapi_types.Email(dbUser.Email),
		CreatedAt:   database.Time(dbUser.CreatedAt),
		Locale:      dbUser.Locale,
	}, nil
}

// UpdateLocale persists the user's preferred UI/email locale.
func (s *Service) UpdateLocale(ctx context.Context, userID, locale string) error {
	if err := s.queries.UpdateUserLocale(ctx, queries.UpdateUserLocaleParams{
		Locale: locale,
		ID:     userID,
	}); err != nil {
		return fmt.Errorf("update user locale: %w", err)
	}
	return nil
}

func (s *Service) Extensions(ctx context.Context, userID string, activeOrgID, requestedLanguage *string) ([]api.AccountExtension, error) {
	language := shopware.ContentLanguage(requestedLanguage)
	if activeOrgID == nil {
		return []api.AccountExtension{}, nil
	}

	storeRows, err := s.queries.GetUserStoreExtensionsByOrg(ctx, queries.GetUserStoreExtensionsByOrgParams{
		UserID:         userID,
		OrganizationID: *activeOrgID,
		Language:       language,
	})
	if err != nil {
		return nil, fmt.Errorf("get user store extensions: %w", err)
	}
	unknownRows, err := s.queries.GetUserUnknownExtensionsByOrg(ctx, queries.GetUserUnknownExtensionsByOrgParams{
		UserID:         userID,
		OrganizationID: *activeOrgID,
	})
	if err != nil {
		return nil, fmt.Errorf("get user unknown extensions: %w", err)
	}
	changelogRows, err := s.queries.GetUserStoreExtensionChangelogsByOrg(ctx, queries.GetUserStoreExtensionChangelogsByOrgParams{
		UserID:         userID,
		OrganizationID: *activeOrgID,
		Language:       language,
	})
	if err != nil {
		return nil, fmt.Errorf("get user store extension changelogs: %w", err)
	}

	changelogByExt := make(map[string][]changelogVersion)
	for _, r := range changelogRows {
		changelogByExt[r.ExtensionName] = append(changelogByExt[r.ExtensionName], changelogVersion{
			Version:    r.Version,
			Text:       ptr.Deref(r.Changelog),
			ReleasedAt: ptr.Deref(r.ReleasedAt),
		})
	}

	imageRows, err := s.queries.GetUserStoreExtensionImagesByOrg(ctx, queries.GetUserStoreExtensionImagesByOrgParams{
		UserID:         userID,
		OrganizationID: *activeOrgID,
	})
	if err != nil {
		return nil, fmt.Errorf("get user store extension images: %w", err)
	}

	screenshotsByExt := make(map[string][]api.ExtensionScreenshot)
	for _, r := range imageRows {
		screenshotsByExt[r.ExtensionName] = append(screenshotsByExt[r.ExtensionName], api.ExtensionScreenshot{
			Url:     r.Url,
			Preview: r.Preview,
		})
	}

	// Normalize both sources into a single iteration order (by name, then
	// environment) so grouping is stable.
	rows := make([]accountExtRow, 0, len(storeRows)+len(unknownRows))
	for _, row := range storeRows {
		rows = append(rows, accountExtRow{
			Name: row.Name, Label: row.Label, Version: row.Version, LatestVersion: row.LatestVersion,
			Active: row.Active, Installed: row.Installed, StoreLink: row.StoreLink, RatingAverage: row.RatingAverage,
			InstalledAt: row.InstalledAt, IconUrl: row.IconUrl, ProducerName: row.ProducerName,
			ProducerWebsite: row.ProducerWebsite, ReleaseDate: row.ReleaseDate,
			ShortDescription: row.ShortDescription, Description: row.Description,
			InstallationManual: row.InstallationManual,
			EnvironmentID:      row.EnvironmentID, EnvironmentName: row.EnvironmentName,
			EnvironmentUrl: row.EnvironmentUrl, EnvironmentOrganizationName: row.EnvironmentOrganizationName,
			EnvironmentOrganizationID: row.EnvironmentOrganizationID,
			EnvironmentShopID:         row.EnvironmentShopID, EnvironmentShopName: row.EnvironmentShopName,
			isStore: true,
		})
	}
	for _, row := range unknownRows {
		rows = append(rows, accountExtRow{
			Name: row.Name, Label: row.Label, Version: row.Version, LatestVersion: row.LatestVersion,
			Active: row.Active, Installed: row.Installed, StoreLink: row.StoreLink, RatingAverage: row.RatingAverage,
			InstalledAt: row.InstalledAt, EnvironmentID: row.EnvironmentID, EnvironmentName: row.EnvironmentName,
			EnvironmentUrl: row.EnvironmentUrl, EnvironmentOrganizationName: row.EnvironmentOrganizationName,
			EnvironmentOrganizationID: row.EnvironmentOrganizationID,
			EnvironmentShopID:         row.EnvironmentShopID, EnvironmentShopName: row.EnvironmentShopName,
			isStore: false,
		})
	}
	return aggregateAccountExtensions(rows, changelogByExt, screenshotsByExt), nil
}

// Extension returns a single aggregated extension by technical name. Unlike Extensions it only
// queries rows for the requested extension, so the detail view does not transfer
// the whole catalog.
func (s *Service) Extension(ctx context.Context, userID string, activeOrgID *string, name string, requestedLanguage *string) (api.AccountExtension, error) {
	language := shopware.ContentLanguage(requestedLanguage)
	if activeOrgID == nil {
		return api.AccountExtension{}, ErrExtensionNotFound
	}

	storeRows, err := s.queries.GetUserStoreExtensionByOrgAndName(ctx, queries.GetUserStoreExtensionByOrgAndNameParams{
		UserID:         userID,
		OrganizationID: *activeOrgID,
		Language:       language,
		ExtensionName:  name,
	})
	if err != nil {
		return api.AccountExtension{}, fmt.Errorf("get user store extension: %w", err)
	}
	unknownRows, err := s.queries.GetUserUnknownExtensionByOrgAndName(ctx, queries.GetUserUnknownExtensionByOrgAndNameParams{
		UserID:         userID,
		OrganizationID: *activeOrgID,
		Name:           name,
	})
	if err != nil {
		return api.AccountExtension{}, fmt.Errorf("get user unknown extension: %w", err)
	}
	changelogRows, err := s.queries.GetUserStoreExtensionChangelogsByOrgAndName(ctx, queries.GetUserStoreExtensionChangelogsByOrgAndNameParams{
		UserID:         userID,
		OrganizationID: *activeOrgID,
		Language:       language,
		ExtensionName:  name,
	})
	if err != nil {
		return api.AccountExtension{}, fmt.Errorf("get user store extension changelogs: %w", err)
	}

	changelogByExt := make(map[string][]changelogVersion)
	for _, r := range changelogRows {
		changelogByExt[r.ExtensionName] = append(changelogByExt[r.ExtensionName], changelogVersion{
			Version:    r.Version,
			Text:       ptr.Deref(r.Changelog),
			ReleasedAt: ptr.Deref(r.ReleasedAt),
		})
	}

	imageRows, err := s.queries.GetUserStoreExtensionImagesByOrgAndName(ctx, queries.GetUserStoreExtensionImagesByOrgAndNameParams{
		UserID:         userID,
		OrganizationID: *activeOrgID,
		ExtensionName:  name,
	})
	if err != nil {
		return api.AccountExtension{}, fmt.Errorf("get user store extension images: %w", err)
	}

	screenshotsByExt := make(map[string][]api.ExtensionScreenshot)
	for _, r := range imageRows {
		screenshotsByExt[r.ExtensionName] = append(screenshotsByExt[r.ExtensionName], api.ExtensionScreenshot{
			Url:     r.Url,
			Preview: r.Preview,
		})
	}

	rows := make([]accountExtRow, 0, len(storeRows)+len(unknownRows))
	for _, row := range storeRows {
		rows = append(rows, accountExtRow{
			Name: row.Name, Label: row.Label, Version: row.Version, LatestVersion: row.LatestVersion,
			Active: row.Active, Installed: row.Installed, StoreLink: row.StoreLink, RatingAverage: row.RatingAverage,
			InstalledAt: row.InstalledAt, IconUrl: row.IconUrl, ProducerName: row.ProducerName,
			ProducerWebsite: row.ProducerWebsite, ReleaseDate: row.ReleaseDate,
			ShortDescription: row.ShortDescription, Description: row.Description,
			InstallationManual: row.InstallationManual,
			EnvironmentID:      row.EnvironmentID, EnvironmentName: row.EnvironmentName,
			EnvironmentUrl: row.EnvironmentUrl, EnvironmentOrganizationName: row.EnvironmentOrganizationName,
			EnvironmentOrganizationID: row.EnvironmentOrganizationID,
			EnvironmentShopID:         row.EnvironmentShopID, EnvironmentShopName: row.EnvironmentShopName,
			isStore: true,
		})
	}
	for _, row := range unknownRows {
		rows = append(rows, accountExtRow{
			Name: row.Name, Label: row.Label, Version: row.Version, LatestVersion: row.LatestVersion,
			Active: row.Active, Installed: row.Installed, StoreLink: row.StoreLink, RatingAverage: row.RatingAverage,
			InstalledAt: row.InstalledAt, EnvironmentID: row.EnvironmentID, EnvironmentName: row.EnvironmentName,
			EnvironmentUrl: row.EnvironmentUrl, EnvironmentOrganizationName: row.EnvironmentOrganizationName,
			EnvironmentOrganizationID: row.EnvironmentOrganizationID,
			EnvironmentShopID:         row.EnvironmentShopID, EnvironmentShopName: row.EnvironmentShopName,
			isStore: false,
		})
	}

	result := aggregateAccountExtensions(rows, changelogByExt, screenshotsByExt)
	if len(result) == 0 {
		return api.AccountExtension{}, ErrExtensionNotFound
	}
	return result[0], nil
}

// aggregateAccountExtensions groups normalized rows by technical name into
// AccountExtension entries, attaching the full changelog history and screenshots
// for store extensions. Rows are ordered by name then environment so the output
// is stable.
func aggregateAccountExtensions(
	rows []accountExtRow,
	changelogByExt map[string][]changelogVersion,
	screenshotsByExt map[string][]api.ExtensionScreenshot,
) []api.AccountExtension {
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Name != rows[j].Name {
			return rows[i].Name < rows[j].Name
		}
		return rows[i].EnvironmentName < rows[j].EnvironmentName
	})

	extMap := make(map[string]*api.AccountExtension)
	extOrder := []string{}

	for _, row := range rows {
		key := row.Name

		ext, exists := extMap[key]
		if !exists {
			var ratingAvg *float32
			if row.RatingAverage != nil {
				v := float32(*row.RatingAverage)
				ratingAvg = &v
			}

			// The catalog/detail view shows the full version history, not just the
			// pending-update delta the per-environment view uses.
			var changelog *[]api.ExtensionChangelogEntry
			if row.isStore {
				changelog = buildFullChangelog(changelogByExt[row.Name])
			}

			ext = &api.AccountExtension{
				Name:          row.Name,
				Label:         row.Label,
				Version:       row.Version,
				LatestVersion: ptr.Deref(row.LatestVersion),
				Active:        row.Active,
				Installed:     row.Installed,
				StoreLink:     row.StoreLink,
				RatingAverage: ratingAvg,
				Changelog:     changelog,
				InstalledAt:   ptr.ParseTime(row.InstalledAt, time.RFC3339),
				Environments:  []api.AccountExtensionEnvironment{},
			}
			if row.isStore {
				ext.IconUrl = row.IconUrl
				ext.ProducerName = row.ProducerName
				ext.ProducerWebsite = row.ProducerWebsite
				ext.ReleaseDate = row.ReleaseDate
				ext.ShortDescription = row.ShortDescription
				ext.Description = row.Description
				ext.InstallationManual = row.InstallationManual
				if shots := screenshotsByExt[row.Name]; len(shots) > 0 {
					ext.Screenshots = &shots
				}
			}
			extMap[key] = ext
			extOrder = append(extOrder, key)
		}

		env := api.AccountExtensionEnvironment{
			EnvironmentId:               int(row.EnvironmentID),
			EnvironmentName:             row.EnvironmentName,
			EnvironmentShopName:         row.EnvironmentShopName,
			EnvironmentUrl:              row.EnvironmentUrl,
			EnvironmentOrganizationName: row.EnvironmentOrganizationName,
			EnvironmentOrganizationId:   row.EnvironmentOrganizationID,
			ShopId:                      int(row.EnvironmentShopID),
			ShopName:                    row.EnvironmentShopName,
			Version:                     row.Version,
			LatestVersion:               ptr.Deref(row.LatestVersion),
			Active:                      row.Active,
			Installed:                   row.Installed,
		}
		ext.Environments = append(ext.Environments, env)
	}

	result := make([]api.AccountExtension, 0, len(extOrder))
	for _, key := range extOrder {
		result = append(result, *extMap[key])
	}
	return result
}

// accountExtRow is a unified row for aggregating account extensions from the
// store and unknown sources.
type accountExtRow struct {
	Name                        string
	Label                       string
	Version                     string
	LatestVersion               *string
	Active                      bool
	Installed                   bool
	StoreLink                   *string
	RatingAverage               *int32
	InstalledAt                 *string
	IconUrl                     *string
	ProducerName                *string
	ProducerWebsite             *string
	ReleaseDate                 *string
	ShortDescription            *string
	Description                 *string
	InstallationManual          *string
	EnvironmentID               int32
	EnvironmentName             string
	EnvironmentUrl              string
	EnvironmentOrganizationName string
	EnvironmentOrganizationID   string
	EnvironmentShopID           int32
	EnvironmentShopName         string
	isStore                     bool
}

func (s *Service) Organizations(ctx context.Context, userID string) ([]api.AccountOrganization, error) {
	rows, err := s.queries.GetUserOrganizations(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user organizations: %w", err)
	}

	result := make([]api.AccountOrganization, 0, len(rows))
	for _, row := range rows {
		result = append(result, api.AccountOrganization{
			Id:               row.ID,
			Name:             row.Name,
			Logo:             row.Logo,
			CreatedAt:        database.Time(row.CreatedAt),
			EnvironmentCount: int(row.EnvironmentCount),
			MemberCount:      int(row.MemberCount),
		})
	}

	return result, nil
}

func (s *Service) Environments(ctx context.Context, userID string, activeOrgID *string) ([]api.AccountEnvironment, error) {
	if activeOrgID == nil {
		return []api.AccountEnvironment{}, nil
	}

	rows, err := s.queries.GetUserEnvironmentsByOrg(ctx, queries.GetUserEnvironmentsByOrgParams{
		UserID:         userID,
		OrganizationID: *activeOrgID,
	})
	if err != nil {
		return nil, fmt.Errorf("get user environments: %w", err)
	}

	result := make([]api.AccountEnvironment, 0, len(rows))
	for _, row := range rows {
		var shopId *int
		if row.ShopID != nil {
			v := int(*row.ShopID)
			shopId = &v
		}
		result = append(result, api.AccountEnvironment{
			Id: int(row.ID), Name: row.Name, Url: row.Url, Favicon: row.Favicon,
			Status: row.Status, ShopwareVersion: row.ShopwareVersion,
			LastScrapedAt: database.TimePtr(row.LastScrapedAt), LastScrapedError: row.LastScrapedError,
			OrganizationId: row.OrganizationID, OrganizationName: row.OrganizationName,
			ShopId: shopId, ShopName: row.ShopName,
		})
	}

	return result, nil
}

func (s *Service) Shops(ctx context.Context, userID string, activeOrgID *string) ([]api.AccountShop, error) {
	if activeOrgID == nil {
		return []api.AccountShop{}, nil
	}

	rows, err := s.queries.GetUserShopsByOrg(ctx, queries.GetUserShopsByOrgParams{
		UserID:         userID,
		OrganizationID: *activeOrgID,
	})
	if err != nil {
		return nil, fmt.Errorf("get user shops: %w", err)
	}

	result := make([]api.AccountShop, 0, len(rows))
	for _, row := range rows {
		result = append(result, api.AccountShop{
			Id: int(row.ID), Name: row.Name, Description: row.Description,
			GitUrl: row.GitUrl, OrganizationId: row.OrganizationID, OrganizationName: row.OrganizationName,
			DefaultEnvironmentId: ptr.Int(row.DefaultEnvironmentID),
		})
	}

	return result, nil
}

func (s *Service) Changelogs(ctx context.Context, userID string, activeOrgID *string) ([]api.AccountChangelog, error) {
	if activeOrgID == nil {
		return []api.AccountChangelog{}, nil
	}

	rows, err := s.queries.GetUserChangelogsByOrg(ctx, queries.GetUserChangelogsByOrgParams{
		UserID:         userID,
		OrganizationID: *activeOrgID,
	})
	if err != nil {
		return nil, fmt.Errorf("get user changelogs: %w", err)
	}

	result := make([]api.AccountChangelog, 0, len(rows))
	for _, row := range rows {
		environmentID := 0
		if row.EnvironmentID != nil {
			environmentID = int(*row.EnvironmentID)
		}

		var extensions []api.ExtensionDiff
		if row.Extensions != nil {
			if err := json.Unmarshal(row.Extensions, &extensions); err != nil {
				return nil, fmt.Errorf("decode changelog %d extensions: %w", row.ID, err)
			}
		}
		if extensions == nil {
			extensions = []api.ExtensionDiff{}
		}

		result = append(result, api.AccountChangelog{
			Id:                          int(row.ID),
			EnvironmentId:               environmentID,
			EnvironmentName:             row.EnvironmentName,
			EnvironmentShopName:         row.EnvironmentShopName,
			EnvironmentOrganizationName: row.EnvironmentOrganizationName,
			EnvironmentOrganizationId:   row.EnvironmentOrganizationID,
			Extensions:                  extensions,
			OldShopwareVersion:          row.OldShopwareVersion,
			NewShopwareVersion:          row.NewShopwareVersion,
			Date:                        database.Time(row.Date),
		})
	}

	return result, nil
}

// SubscribedEnvironments resolves the user's environment subscription markers
// (notification_preference rows) to display names.
func (s *Service) SubscribedEnvironments(ctx context.Context, userID string) []api.SubscribedEnvironment {
	scopeIDs, err := s.queries.ListSubscribedEnvironmentIDs(ctx, userID)
	if err != nil {
		slog.WarnContext(ctx, "failed to list subscribed environments", "userId", userID, "error", err)
		return []api.SubscribedEnvironment{}
	}

	subscribedEnvironments := make([]api.SubscribedEnvironment, 0, len(scopeIDs))
	for _, idStr := range scopeIDs {
		environmentID, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}

		environment, err := s.queries.GetEnvironmentByID(ctx, int32(environmentID))
		if err != nil {
			slog.WarnContext(ctx, "failed to get subscribed environment", "environmentId", environmentID, "error", err)
			continue
		}

		sub := api.SubscribedEnvironment{
			Id:       environmentID,
			Name:     environment.Name,
			ShopName: environment.ShopName,
		}
		if environment.ShopID > 0 {
			shopID := int(environment.ShopID)
			sub.ShopId = &shopID
		}
		subscribedEnvironments = append(subscribedEnvironments, sub)
	}

	return subscribedEnvironments
}

type changelogVersion struct {
	Version    string
	Text       string
	ReleasedAt string
}

func buildFullChangelog(versions []changelogVersion) *[]api.ExtensionChangelogEntry {
	entries := make([]api.ExtensionChangelogEntry, 0, len(versions))
	for _, item := range versions {
		entry := api.ExtensionChangelogEntry{Version: item.Version, Text: item.Text}
		if item.ReleasedAt != "" {
			if parsed, err := time.Parse(time.RFC3339, item.ReleasedAt); err == nil {
				entry.CreationDate = parsed
			}
		}
		entries = append(entries, entry)
	}
	if len(entries) == 0 {
		return nil
	}
	sort.Slice(entries, func(i, j int) bool {
		return version.Compare(entries[i].Version, entries[j].Version) > 0
	})
	return &entries
}
