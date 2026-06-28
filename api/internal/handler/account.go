package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// GetAccountMe returns the current user's profile.
func (h *Handler) GetAccountMe(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	dbUser, err := h.queries.GetUserByID(r.Context(), user.ID)
	if err != nil {
		slog.Error("failed to get user", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	profile := api.UserProfile{
		Id:          dbUser.ID,
		DisplayName: dbUser.Name,
		Email:       openapi_types.Email(dbUser.Email),
		CreatedAt:   pgtimeToTime(dbUser.CreatedAt),
	}

	httputil.WriteJSON(w, http.StatusOK, profile)
}

// GetAccountExtensions returns aggregated extensions for the user's active organization.
func (h *Handler) GetAccountExtensions(w http.ResponseWriter, r *http.Request, params api.GetAccountExtensionsParams) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	var langStr *string
	if params.Language != nil {
		s := string(*params.Language)
		langStr = &s
	}
	language := resolveLanguage(langStr)

	activeOrgID := h.getActiveOrganizationID(r)
	if activeOrgID == nil {
		httputil.WriteJSON(w, http.StatusOK, []api.AccountExtension{})
		return
	}

	storeRows, err := h.queries.GetUserStoreExtensionsByOrg(r.Context(), queries.GetUserStoreExtensionsByOrgParams{
		UserID:         user.ID,
		OrganizationID: *activeOrgID,
		Language:       language,
	})
	if err != nil {
		slog.Error("failed to get user store extensions", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get extensions")
		return
	}
	unknownRows, err := h.queries.GetUserUnknownExtensionsByOrg(r.Context(), queries.GetUserUnknownExtensionsByOrgParams{
		UserID:         user.ID,
		OrganizationID: *activeOrgID,
	})
	if err != nil {
		slog.Error("failed to get user unknown extensions", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get extensions")
		return
	}
	changelogRows, err := h.queries.GetUserStoreExtensionChangelogsByOrg(r.Context(), queries.GetUserStoreExtensionChangelogsByOrgParams{
		UserID:         user.ID,
		OrganizationID: *activeOrgID,
		Language:       language,
	})
	if err != nil {
		slog.Error("failed to get user store extension changelogs", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get extensions")
		return
	}

	changelogByExt := make(map[string][]changelogVersion)
	for _, r := range changelogRows {
		changelogByExt[r.ExtensionName] = append(changelogByExt[r.ExtensionName], changelogVersion{
			Version:    r.Version,
			Text:       deref(r.Changelog),
			ReleasedAt: deref(r.ReleasedAt),
		})
	}

	imageRows, err := h.queries.GetUserStoreExtensionImagesByOrg(r.Context(), queries.GetUserStoreExtensionImagesByOrgParams{
		UserID:         user.ID,
		OrganizationID: *activeOrgID,
	})
	if err != nil {
		slog.Error("failed to get user store extension images", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get extensions")
		return
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
	result := aggregateAccountExtensions(rows, changelogByExt, screenshotsByExt)

	httputil.WriteJSON(w, http.StatusOK, result)
}

// GetAccountExtension returns a single aggregated extension by technical name,
// scoped to the user's active organization. Unlike GetAccountExtensions it only
// queries rows for the requested extension, so the detail view does not transfer
// the whole catalog.
func (h *Handler) GetAccountExtension(w http.ResponseWriter, r *http.Request, name string, params api.GetAccountExtensionParams) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	var langStr *string
	if params.Language != nil {
		s := string(*params.Language)
		langStr = &s
	}
	language := resolveLanguage(langStr)

	activeOrgID := h.getActiveOrganizationID(r)
	if activeOrgID == nil {
		httputil.WriteError(w, http.StatusNotFound, "extension not found")
		return
	}

	storeRows, err := h.queries.GetUserStoreExtensionByOrgAndName(r.Context(), queries.GetUserStoreExtensionByOrgAndNameParams{
		UserID:         user.ID,
		OrganizationID: *activeOrgID,
		Language:       language,
		ExtensionName:  name,
	})
	if err != nil {
		slog.Error("failed to get user store extension", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get extension")
		return
	}
	unknownRows, err := h.queries.GetUserUnknownExtensionByOrgAndName(r.Context(), queries.GetUserUnknownExtensionByOrgAndNameParams{
		UserID:         user.ID,
		OrganizationID: *activeOrgID,
		Name:           name,
	})
	if err != nil {
		slog.Error("failed to get user unknown extension", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get extension")
		return
	}
	changelogRows, err := h.queries.GetUserStoreExtensionChangelogsByOrgAndName(r.Context(), queries.GetUserStoreExtensionChangelogsByOrgAndNameParams{
		UserID:         user.ID,
		OrganizationID: *activeOrgID,
		Language:       language,
		ExtensionName:  name,
	})
	if err != nil {
		slog.Error("failed to get user store extension changelogs", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get extension")
		return
	}

	changelogByExt := make(map[string][]changelogVersion)
	for _, r := range changelogRows {
		changelogByExt[r.ExtensionName] = append(changelogByExt[r.ExtensionName], changelogVersion{
			Version:    r.Version,
			Text:       deref(r.Changelog),
			ReleasedAt: deref(r.ReleasedAt),
		})
	}

	imageRows, err := h.queries.GetUserStoreExtensionImagesByOrgAndName(r.Context(), queries.GetUserStoreExtensionImagesByOrgAndNameParams{
		UserID:         user.ID,
		OrganizationID: *activeOrgID,
		ExtensionName:  name,
	})
	if err != nil {
		slog.Error("failed to get user store extension images", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get extension")
		return
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
		httputil.WriteError(w, http.StatusNotFound, "extension not found")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, result[0])
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
				LatestVersion: deref(row.LatestVersion),
				Active:        row.Active,
				Installed:     row.Installed,
				StoreLink:     row.StoreLink,
				RatingAverage: ratingAvg,
				Changelog:     changelog,
				InstalledAt:   parseInstalledAt(row.InstalledAt),
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
			EnvironmentUrl:              row.EnvironmentUrl,
			EnvironmentOrganizationName: row.EnvironmentOrganizationName,
			EnvironmentOrganizationId:   row.EnvironmentOrganizationID,
			ShopId:                      int(row.EnvironmentShopID),
			ShopName:                    row.EnvironmentShopName,
			Version:                     row.Version,
			LatestVersion:               deref(row.LatestVersion),
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

// GetAccountOrganizations returns the organizations the user belongs to.
func (h *Handler) GetAccountOrganizations(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	rows, err := h.queries.GetUserOrganizations(r.Context(), user.ID)
	if err != nil {
		slog.Error("failed to get user organizations", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get organizations")
		return
	}

	result := make([]api.AccountOrganization, 0, len(rows))
	for _, row := range rows {
		result = append(result, api.AccountOrganization{
			Id:               row.ID,
			Name:             row.Name,
			Logo:             row.Logo,
			CreatedAt:        pgtimeToTime(row.CreatedAt),
			EnvironmentCount: int(row.EnvironmentCount),
			MemberCount:      int(row.MemberCount),
		})
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// GetAccountEnvironments returns environments for the user's active organization.
func (h *Handler) GetAccountEnvironments(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	activeOrgID := h.getActiveOrganizationID(r)
	if activeOrgID == nil {
		httputil.WriteJSON(w, http.StatusOK, []api.AccountEnvironment{})
		return
	}

	rows, err := h.queries.GetUserEnvironmentsByOrg(r.Context(), queries.GetUserEnvironmentsByOrgParams{
		UserID:         user.ID,
		OrganizationID: *activeOrgID,
	})
	if err != nil {
		slog.Error("failed to get user environments", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get environments")
		return
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
			LastScrapedAt: pgtimeToTimePtr(row.LastScrapedAt), LastScrapedError: row.LastScrapedError,
			OrganizationId: row.OrganizationID, OrganizationName: row.OrganizationName,
			ShopId: shopId, ShopName: row.ShopName,
		})
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// GetAccountShops returns shops for the user's active organization.
func (h *Handler) GetAccountShops(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	activeOrgID := h.getActiveOrganizationID(r)
	if activeOrgID == nil {
		httputil.WriteJSON(w, http.StatusOK, []api.AccountShop{})
		return
	}

	rows, err := h.queries.GetUserShopsByOrg(r.Context(), queries.GetUserShopsByOrgParams{
		UserID:         user.ID,
		OrganizationID: *activeOrgID,
	})
	if err != nil {
		slog.Error("failed to get user shops", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get shops")
		return
	}

	result := make([]api.AccountShop, 0, len(rows))
	for _, row := range rows {
		result = append(result, api.AccountShop{
			Id: int(row.ID), Name: row.Name, Description: row.Description,
			GitUrl: row.GitUrl, OrganizationId: row.OrganizationID, OrganizationName: row.OrganizationName,
			DefaultEnvironmentId: int32PtrToIntPtr(row.DefaultEnvironmentID),
		})
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// GetAccountChangelogs returns changelogs for the user's active organization.
func (h *Handler) GetAccountChangelogs(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	activeOrgID := h.getActiveOrganizationID(r)
	if activeOrgID == nil {
		httputil.WriteJSON(w, http.StatusOK, []api.AccountChangelog{})
		return
	}

	rows, err := h.queries.GetUserChangelogsByOrg(r.Context(), queries.GetUserChangelogsByOrgParams{
		UserID:         user.ID,
		OrganizationID: *activeOrgID,
	})
	if err != nil {
		slog.Error("failed to get user changelogs", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get changelogs")
		return
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
				slog.Error("failed to unmarshal changelog extensions", "error", err)
			}
		}
		if extensions == nil {
			extensions = []api.ExtensionDiff{}
		}

		result = append(result, api.AccountChangelog{
			Id:                          int(row.ID),
			EnvironmentId:               environmentID,
			EnvironmentName:             row.EnvironmentName,
			EnvironmentOrganizationName: row.EnvironmentOrganizationName,
			EnvironmentOrganizationId:   row.EnvironmentOrganizationID,
			Extensions:                  extensions,
			OldShopwareVersion:          row.OldShopwareVersion,
			NewShopwareVersion:          row.NewShopwareVersion,
			Date:                        pgtimeToTime(row.Date),
		})
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// GetAccountSubscribedEnvironments returns environments the user is subscribed to for notifications.
func (h *Handler) GetAccountSubscribedEnvironments(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	// User notifications is a list of strings like "environment-123"
	subscribedEnvironments := []api.SubscribedEnvironment{}

	for _, notification := range user.Notifications {
		if strings.HasPrefix(notification, "environment-") {
			idStr := strings.TrimPrefix(notification, "environment-")
			environmentID, err := strconv.Atoi(idStr)
			if err != nil {
				continue
			}

			// Fetch environment name
			environment, err := h.queries.GetEnvironmentByID(r.Context(), int32(environmentID))
			if err != nil {
				slog.Warn("failed to get subscribed environment", "environmentId", environmentID, "error", err)
				continue
			}

			subscribedEnvironments = append(subscribedEnvironments, api.SubscribedEnvironment{
				Id:   environmentID,
				Name: environment.Name,
			})
		}
	}

	httputil.WriteJSON(w, http.StatusOK, subscribedEnvironments)
}

// helper to check for "environment-{id}" pattern match
func environmentNotificationKey(environmentID int) string {
	return fmt.Sprintf("environment-%d", environmentID)
}
