package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// AdminGetStats returns admin dashboard statistics.
func (h *Handler) AdminGetStats(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if user.Role != "admin" {
		httputil.WriteError(w, http.StatusForbidden, httputil.MsgAdminRequired)
		return
	}

	stats, err := h.queries.AdminGetStats(r.Context())
	if err != nil {
		slog.Error("failed to get admin stats", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get stats")
		return
	}

	result := api.AdminStats{
		TotalUsers:         int(stats.TotalUsers),
		TotalOrganizations: int(stats.TotalOrganizations),
		TotalEnvironments:  int(stats.TotalEnvironments),
		EnvironmentsByStatus: struct {
			Green  int `json:"green"`
			Red    int `json:"red"`
			Yellow int `json:"yellow"`
		}{
			Green:  int(stats.EnvironmentsGreen),
			Yellow: int(stats.EnvironmentsYellow),
			Red:    int(stats.EnvironmentsRed),
		},
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// AdminGetOrganizations lists all organizations (admin only).
func (h *Handler) AdminGetOrganizations(w http.ResponseWriter, r *http.Request, params api.AdminGetOrganizationsParams) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if user.Role != "admin" {
		httputil.WriteError(w, http.StatusForbidden, httputil.MsgAdminRequired)
		return
	}

	limit := int32(25)
	offset := int32(0)
	if params.Limit != nil {
		limit = int32(*params.Limit)
	}
	if params.Offset != nil {
		offset = int32(*params.Offset)
	}

	search := searchValue(params.SearchValue)
	sortBy, sortDir := adminSort((*string)(params.SortBy), (*string)(params.SortDirection), "createdAt")

	rows, err := h.queries.AdminListOrganizations(r.Context(), queries.AdminListOrganizationsParams{
		Limit:   limit,
		Offset:  offset,
		Search:  search,
		SortBy:  sortBy,
		SortDir: sortDir,
	})
	if err != nil {
		slog.Error("failed to list organizations", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get organizations")
		return
	}

	total, err := h.queries.AdminCountOrganizations(r.Context(), search)
	if err != nil {
		slog.Error("failed to count organizations", "error", err)
		total = 0
	}

	orgs := make([]api.AccountOrganization, 0, len(rows))
	for _, row := range rows {
		orgs = append(orgs, api.AccountOrganization{
			Id:               row.ID,
			Name:             row.Name,
			Logo:             row.Logo,
			CreatedAt:        pgtimeToTime(row.CreatedAt),
			EnvironmentCount: int(row.EnvironmentCount),
			MemberCount:      int(row.MemberCount),
		})
	}

	httputil.WriteJSON(w, http.StatusOK, api.AdminOrganizationsResponse{
		Organizations: orgs,
		Total:         int(total),
	})
}

// AdminGetEnvironments lists all environments (admin only).
func (h *Handler) AdminGetEnvironments(w http.ResponseWriter, r *http.Request, params api.AdminGetEnvironmentsParams) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if user.Role != "admin" {
		httputil.WriteError(w, http.StatusForbidden, httputil.MsgAdminRequired)
		return
	}

	limit := int32(25)
	offset := int32(0)
	if params.Limit != nil {
		limit = int32(*params.Limit)
	}
	if params.Offset != nil {
		offset = int32(*params.Offset)
	}

	search := searchValue(params.SearchValue)
	sortBy, sortDir := adminSort(params.SortBy, (*string)(params.SortDirection), "createdAt")

	rows, err := h.queries.AdminListEnvironments(r.Context(), queries.AdminListEnvironmentsParams{
		Limit:   limit,
		Offset:  offset,
		Search:  search,
		SortBy:  sortBy,
		SortDir: sortDir,
	})
	if err != nil {
		slog.Error("failed to list environments", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get environments")
		return
	}

	total, err := h.queries.AdminCountEnvironments(r.Context(), search)
	if err != nil {
		slog.Error("failed to count environments", "error", err)
		total = 0
	}

	environments := make([]api.AccountEnvironment, 0, len(rows))
	for _, row := range rows {
		environments = append(environments, api.AccountEnvironment{
			Id:               int(row.ID),
			Name:             row.Name,
			Url:              row.Url,
			Status:           row.Status,
			ShopwareVersion:  row.ShopwareVersion,
			LastScrapedAt:    pgtimeToTimePtr(row.LastScrapedAt),
			OrganizationId:   row.OrganizationID,
			OrganizationName: row.OrganizationName,
		})
	}

	httputil.WriteJSON(w, http.StatusOK, api.AdminEnvironmentsResponse{
		Environments: environments,
		Total:        int(total),
	})
}

// AdminGetGrowth returns growth data over time.
func (h *Handler) AdminGetGrowth(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if user.Role != "admin" {
		httputil.WriteError(w, http.StatusForbidden, httputil.MsgAdminRequired)
		return
	}

	environmentRows, err := h.queries.AdminGetGrowthEnvironments(r.Context())
	if err != nil {
		slog.Error("failed to get environment growth", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get growth data")
		return
	}

	userRows, err := h.queries.AdminGetGrowthUsers(r.Context())
	if err != nil {
		slog.Error("failed to get user growth", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get growth data")
		return
	}

	environmentGrowth := make([]api.GrowthDataPoint, 0, len(environmentRows))
	for _, row := range environmentRows {
		environmentGrowth = append(environmentGrowth, api.GrowthDataPoint{
			Month: row.Month,
			Count: int(row.Count),
		})
	}

	userGrowth := make([]api.GrowthDataPoint, 0, len(userRows))
	for _, row := range userRows {
		userGrowth = append(userGrowth, api.GrowthDataPoint{
			Month: row.Month,
			Count: int(row.Count),
		})
	}

	httputil.WriteJSON(w, http.StatusOK, api.AdminGrowth{
		Environments: environmentGrowth,
		Users:        userGrowth,
	})
}

// searchValue trims the optional search parameter and returns nil when it is
// empty so the list/count queries skip filtering.
func searchValue(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// adminSort normalises the sort column and direction. The column is passed
// through to the query's whitelisted CASE expressions (an unknown value falls
// back to the default ordering), and the direction defaults to descending.
func adminSort(sortBy, sortDir *string, fallback string) (string, string) {
	by := fallback
	if sortBy != nil && *sortBy != "" {
		by = *sortBy
	}
	dir := "desc"
	if sortDir != nil && strings.ToLower(*sortDir) == "asc" {
		dir = "asc"
	}
	return by, dir
}

// AdminGetRecentActivity returns recent user and environment activity.
func (h *Handler) AdminGetRecentActivity(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if user.Role != "admin" {
		httputil.WriteError(w, http.StatusForbidden, httputil.MsgAdminRequired)
		return
	}

	recentEnvironmentRows, err := h.queries.AdminGetRecentEnvironments(r.Context())
	if err != nil {
		slog.Error("failed to get recent environments", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get recent activity")
		return
	}

	recentUserRows, err := h.queries.AdminGetRecentUsers(r.Context())
	if err != nil {
		slog.Error("failed to get recent users", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get recent activity")
		return
	}

	recentEnvironments := make([]api.AccountEnvironment, 0, len(recentEnvironmentRows))
	for _, row := range recentEnvironmentRows {
		recentEnvironments = append(recentEnvironments, api.AccountEnvironment{
			Id:               int(row.ID),
			Name:             row.Name,
			Url:              row.Url,
			OrganizationName: row.OrganizationName,
		})
	}

	recentUsers := make([]api.UserProfile, 0, len(recentUserRows))
	for _, row := range recentUserRows {
		recentUsers = append(recentUsers, api.UserProfile{
			Id:          row.ID,
			DisplayName: row.Name,
			Email:       openapi_types.Email(row.Email),
			CreatedAt:   pgtimeToTime(row.CreatedAt),
		})
	}

	httputil.WriteJSON(w, http.StatusOK, api.AdminRecentActivity{
		RecentEnvironments: recentEnvironments,
		RecentUsers:        recentUsers,
	})
}

// AdminGetShopwareVersions returns Shopware version distribution across environments.
func (h *Handler) AdminGetShopwareVersions(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if user.Role != "admin" {
		httputil.WriteError(w, http.StatusForbidden, httputil.MsgAdminRequired)
		return
	}

	rows, err := h.queries.AdminGetShopwareVersions(r.Context())
	if err != nil {
		slog.Error("failed to get shopware versions", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get shopware versions")
		return
	}

	result := make([]api.ShopwareVersionCount, 0, len(rows))
	for _, row := range rows {
		result = append(result, api.ShopwareVersionCount{
			Version: row.Version,
			Count:   int(row.Count),
		})
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}
