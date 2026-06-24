package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/jackc/pgx/v5"
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

// requireAdmin checks the request is from an authenticated admin user. It
// returns false (and writes the response) when the caller is not an admin.
func (h *Handler) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	user := h.requireUser(w, r)
	if user == nil {
		return false
	}
	if user.Role != "admin" {
		httputil.WriteError(w, http.StatusForbidden, httputil.MsgAdminRequired)
		return false
	}
	return true
}

// AdminGetOrganizationDetail returns a single organization with its members,
// environments, invitations, SSO providers and shops (admin only).
func (h *Handler) AdminGetOrganizationDetail(w http.ResponseWriter, r *http.Request, orgID string) {
	if !h.requireAdmin(w, r) {
		return
	}

	org, err := h.queries.AdminGetOrganizationDetail(r.Context(), orgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "organization not found")
			return
		}
		slog.Error("failed to get organization detail", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get organization")
		return
	}

	members, err := h.queries.AdminGetOrganizationMembers(r.Context(), orgID)
	if err != nil {
		slog.Error("failed to get organization members", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get organization")
		return
	}
	environments, err := h.queries.AdminGetOrganizationEnvironments(r.Context(), orgID)
	if err != nil {
		slog.Error("failed to get organization environments", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get organization")
		return
	}
	invitations, err := h.queries.AdminGetOrganizationInvitations(r.Context(), orgID)
	if err != nil {
		slog.Error("failed to get organization invitations", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get organization")
		return
	}
	ssoProviders, err := h.queries.AdminGetOrganizationSSO(r.Context(), &orgID)
	if err != nil {
		slog.Error("failed to get organization sso", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get organization")
		return
	}
	shops, err := h.queries.AdminGetOrganizationShops(r.Context(), orgID)
	if err != nil {
		slog.Error("failed to get organization shops", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get organization")
		return
	}

	detail := api.AdminOrganizationDetail{
		Id:               org.ID,
		Name:             org.Name,
		Slug:             org.Slug,
		Logo:             org.Logo,
		CreatedAt:        pgtimeToTime(org.CreatedAt),
		MemberCount:      int(org.MemberCount),
		EnvironmentCount: int(org.EnvironmentCount),
		Members:          make([]api.AdminOrganizationMember, 0, len(members)),
		Environments:     make([]api.AdminOrganizationEnvironment, 0, len(environments)),
		Invitations:      make([]api.AdminOrganizationInvitation, 0, len(invitations)),
		SsoProviders:     make([]api.AdminOrganizationSSO, 0, len(ssoProviders)),
		Shops:            make([]api.AdminOrganizationShop, 0, len(shops)),
	}

	for _, m := range members {
		detail.Members = append(detail.Members, api.AdminOrganizationMember{
			UserId:    m.UserID,
			Name:      m.Name,
			Email:     m.Email,
			Image:     m.Image,
			Role:      m.Role,
			CreatedAt: pgtimeToTime(m.CreatedAt),
		})
	}
	for _, e := range environments {
		detail.Environments = append(detail.Environments, api.AdminOrganizationEnvironment{
			Id:              int(e.ID),
			Name:            e.Name,
			Url:             e.Url,
			Status:          e.Status,
			ShopwareVersion: e.ShopwareVersion,
			LastScrapedAt:   pgtimeToTimePtr(e.LastScrapedAt),
		})
	}
	for _, i := range invitations {
		detail.Invitations = append(detail.Invitations, api.AdminOrganizationInvitation{
			Id:           i.ID,
			Email:        i.Email,
			Role:         i.Role,
			Status:       i.Status,
			ExpiresAt:    pgtimeToTime(i.ExpiresAt),
			CreatedAt:    pgtimeToTime(i.CreatedAt),
			InviterName:  i.InviterName,
			InviterEmail: i.InviterEmail,
		})
	}
	for _, s := range ssoProviders {
		detail.SsoProviders = append(detail.SsoProviders, api.AdminOrganizationSSO{
			Id:         s.ID,
			Issuer:     s.Issuer,
			ProviderId: s.ProviderID,
			Domain:     s.Domain,
		})
	}
	for _, s := range shops {
		shop := api.AdminOrganizationShop{
			Id:          int(s.ID),
			Name:        s.Name,
			Description: s.Description,
			CreatedAt:   pgtimeToTime(s.CreatedAt),
		}
		if s.DefaultEnvironmentID != nil {
			id := int(*s.DefaultEnvironmentID)
			shop.DefaultEnvironmentId = &id
		}
		detail.Shops = append(detail.Shops, shop)
	}

	httputil.WriteJSON(w, http.StatusOK, detail)
}

// AdminGetEnvironmentDetail returns a single environment with checks,
// extensions, scheduled tasks and last deployment (admin only). Secrets
// (client_secret, environment_token) are never included.
func (h *Handler) AdminGetEnvironmentDetail(w http.ResponseWriter, r *http.Request, envID int) {
	if !h.requireAdmin(w, r) {
		return
	}

	env, err := h.queries.AdminGetEnvironmentDetail(r.Context(), int32(envID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "environment not found")
			return
		}
		slog.Error("failed to get environment detail", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get environment")
		return
	}

	checks, err := h.queries.AdminGetEnvironmentChecks(r.Context(), env.ID)
	if err != nil {
		slog.Error("failed to get environment checks", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get environment")
		return
	}
	extensions, err := h.queries.AdminGetEnvironmentExtensions(r.Context(), env.ID)
	if err != nil {
		slog.Error("failed to get environment extensions", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get environment")
		return
	}
	tasks, err := h.queries.AdminGetEnvironmentTasks(r.Context(), env.ID)
	if err != nil {
		slog.Error("failed to get environment tasks", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get environment")
		return
	}

	detail := api.AdminEnvironmentDetail{
		Id:                   int(env.ID),
		Name:                 env.Name,
		Url:                  env.Url,
		Status:               env.Status,
		ShopwareVersion:      env.ShopwareVersion,
		CreatedAt:            pgtimeToTime(env.CreatedAt),
		LastScrapedAt:        pgtimeToTimePtr(env.LastScrapedAt),
		LastScrapedError:     env.LastScrapedError,
		ConnectionIssueCount: int(env.ConnectionIssueCount),
		OrganizationId:       env.OrganizationID,
		OrganizationName:     env.OrganizationName,
		ShopId:               int(env.ShopID),
		ShopName:             env.ShopName,
		Checks:               make([]api.AdminEnvironmentCheck, 0, len(checks)),
		Extensions:           make([]api.AdminEnvironmentExtension, 0, len(extensions)),
		ScheduledTasks:       make([]api.AdminEnvironmentTask, 0, len(tasks)),
	}

	for _, c := range checks {
		detail.Checks = append(detail.Checks, api.AdminEnvironmentCheck{
			Id:      int(c.ID),
			CheckId: c.CheckID,
			Level:   c.Level,
			Message: c.Message,
			Source:  c.Source,
			Link:    c.Link,
		})
	}
	for _, e := range extensions {
		detail.Extensions = append(detail.Extensions, api.AdminEnvironmentExtension{
			Id:            int(e.ID),
			Name:          e.Name,
			Label:         e.Label,
			Active:        e.Active,
			Version:       e.Version,
			LatestVersion: e.LatestVersion,
			Installed:     e.Installed,
			StoreLink:     e.StoreLink,
		})
	}
	for _, t := range tasks {
		detail.ScheduledTasks = append(detail.ScheduledTasks, api.AdminEnvironmentTask{
			Id:                int(t.ID),
			TaskId:            t.TaskID,
			Name:              t.Name,
			Status:            t.Status,
			Interval:          int(t.Interval),
			Overdue:           t.Overdue,
			LastExecutionTime: t.LastExecutionTime,
			NextExecutionTime: t.NextExecutionTime,
		})
	}

	if dep, err := h.queries.AdminGetEnvironmentLastDeployment(r.Context(), env.ID); err == nil {
		detail.LastDeployment = &api.AdminEnvironmentDeployment{
			Id:            int(dep.ID),
			Name:          dep.Name,
			Command:       dep.Command,
			ReturnCode:    int(dep.ReturnCode),
			ExecutionTime: dep.ExecutionTime,
			Reference:     dep.Reference,
			CreatedAt:     pgtimeToTime(dep.CreatedAt),
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		slog.Error("failed to get environment last deployment", "error", err)
	}

	httputil.WriteJSON(w, http.StatusOK, detail)
}

// AdminGetAuditLog returns a paginated, filterable list of audit-log entries
// with actor/target names resolved (admin only).
func (h *Handler) AdminGetAuditLog(w http.ResponseWriter, r *http.Request, params api.AdminGetAuditLogParams) {
	if !h.requireAdmin(w, r) {
		return
	}

	limit := int32(50)
	offset := int32(0)
	if params.Limit != nil && *params.Limit > 0 && *params.Limit <= 200 {
		limit = int32(*params.Limit)
	}
	if params.Offset != nil && *params.Offset >= 0 {
		offset = int32(*params.Offset)
	}

	action := searchValue(params.Action)
	actor := searchValue(params.ActorUserId)
	target := searchValue(params.TargetUserId)

	rows, err := h.queries.AdminListAuditLog(r.Context(), queries.AdminListAuditLogParams{
		Limit:        limit,
		Offset:       offset,
		Action:       action,
		ActorUserID:  actor,
		TargetUserID: target,
	})
	if err != nil {
		slog.Error("failed to list audit log", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get audit log")
		return
	}

	total, err := h.queries.AdminCountAuditLog(r.Context(), queries.AdminCountAuditLogParams{
		Action:       action,
		ActorUserID:  actor,
		TargetUserID: target,
	})
	if err != nil {
		slog.Error("failed to count audit log", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get audit log")
		return
	}

	entries := make([]api.AdminAuditLogEntry, 0, len(rows))
	for _, a := range rows {
		entries = append(entries, api.AdminAuditLogEntry{
			Id:           a.ID,
			ActorUserId:  a.ActorUserID,
			ActorName:    a.ActorName,
			ActorEmail:   a.ActorEmail,
			Action:       a.Action,
			TargetUserId: a.TargetUserID,
			TargetName:   a.TargetName,
			TargetEmail:  a.TargetEmail,
			Detail:       a.Detail,
			IpAddress:    a.IpAddress,
			CreatedAt:    pgtimeToTime(a.CreatedAt),
		})
	}

	httputil.WriteJSON(w, http.StatusOK, api.AdminAuditLogResponse{
		Entries: entries,
		Total:   int(total),
	})
}
