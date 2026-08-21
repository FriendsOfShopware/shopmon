package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/jackc/pgx/v5"
)

var (
	ErrOrganizationNotFound = errors.New("organization not found")
	ErrEnvironmentNotFound  = errors.New("environment not found")
)

type Service struct {
	queries *queries.Queries
}

func NewService(q *queries.Queries) *Service {
	return &Service{queries: q}
}

func (s *Service) Stats(ctx context.Context) (api.AdminStats, error) {
	stats, err := s.queries.AdminGetStats(ctx)
	if err != nil {
		return api.AdminStats{}, fmt.Errorf("get admin stats: %w", err)
	}
	return api.AdminStats{
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
	}, nil
}

func (s *Service) Organizations(ctx context.Context, params api.AdminGetOrganizationsParams) (api.AdminOrganizationsResponse, error) {
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

	rows, err := s.queries.AdminListOrganizations(ctx, queries.AdminListOrganizationsParams{
		Limit:   limit,
		Offset:  offset,
		Search:  search,
		SortBy:  sortBy,
		SortDir: sortDir,
	})
	if err != nil {
		return api.AdminOrganizationsResponse{}, fmt.Errorf("list organizations: %w", err)
	}

	total, err := s.queries.AdminCountOrganizations(ctx, search)
	if err != nil {
		return api.AdminOrganizationsResponse{}, fmt.Errorf("count organizations: %w", err)
	}

	orgs := make([]api.AccountOrganization, 0, len(rows))
	for _, row := range rows {
		orgs = append(orgs, api.AccountOrganization{
			Id:               row.ID,
			Name:             row.Name,
			Logo:             row.Logo,
			CreatedAt:        database.Time(row.CreatedAt),
			EnvironmentCount: int(row.EnvironmentCount),
			MemberCount:      int(row.MemberCount),
		})
	}

	return api.AdminOrganizationsResponse{
		Organizations: orgs,
		Total:         int(total),
	}, nil
}

func (s *Service) Environments(ctx context.Context, params api.AdminGetEnvironmentsParams) (api.AdminEnvironmentsResponse, error) {
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

	rows, err := s.queries.AdminListEnvironments(ctx, queries.AdminListEnvironmentsParams{
		Limit:   limit,
		Offset:  offset,
		Search:  search,
		SortBy:  sortBy,
		SortDir: sortDir,
	})
	if err != nil {
		return api.AdminEnvironmentsResponse{}, fmt.Errorf("list environments: %w", err)
	}

	total, err := s.queries.AdminCountEnvironments(ctx, search)
	if err != nil {
		return api.AdminEnvironmentsResponse{}, fmt.Errorf("count environments: %w", err)
	}

	environments := make([]api.AccountEnvironment, 0, len(rows))
	for _, row := range rows {
		shopID := int(row.ShopID)
		shopName := row.ShopName
		environments = append(environments, api.AccountEnvironment{
			Id:               int(row.ID),
			Name:             row.Name,
			Url:              row.Url,
			Status:           row.Status,
			ShopwareVersion:  row.ShopwareVersion,
			LastScrapedAt:    database.TimePtr(row.LastScrapedAt),
			OrganizationId:   row.OrganizationID,
			OrganizationName: row.OrganizationName,
			ShopId:           &shopID,
			ShopName:         &shopName,
		})
	}

	return api.AdminEnvironmentsResponse{
		Environments: environments,
		Total:        int(total),
	}, nil
}

func (s *Service) Growth(ctx context.Context) (api.AdminGrowth, error) {
	environmentRows, err := s.queries.AdminGetGrowthEnvironments(ctx)
	if err != nil {
		return api.AdminGrowth{}, fmt.Errorf("get environment growth: %w", err)
	}

	userRows, err := s.queries.AdminGetGrowthUsers(ctx)
	if err != nil {
		return api.AdminGrowth{}, fmt.Errorf("get user growth: %w", err)
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

	return api.AdminGrowth{
		Environments: environmentGrowth,
		Users:        userGrowth,
	}, nil
}

func (s *Service) EcosystemStats(ctx context.Context) (api.EcosystemStats, error) {
	growth, err := s.Growth(ctx)
	if err != nil {
		return api.EcosystemStats{}, err
	}
	versions, err := s.ShopwareVersions(ctx)
	if err != nil {
		return api.EcosystemStats{}, err
	}
	return api.EcosystemStats{Growth: growth, ShopwareVersions: versions}, nil
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

func (s *Service) RecentActivity(ctx context.Context) (api.AdminRecentActivity, error) {
	recentEnvironmentRows, err := s.queries.AdminGetRecentEnvironments(ctx)
	if err != nil {
		return api.AdminRecentActivity{}, fmt.Errorf("get recent environments: %w", err)
	}

	recentUserRows, err := s.queries.AdminGetRecentUsers(ctx)
	if err != nil {
		return api.AdminRecentActivity{}, fmt.Errorf("get recent users: %w", err)
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
			Email:       row.Email,
			CreatedAt:   database.Time(row.CreatedAt),
		})
	}

	return api.AdminRecentActivity{
		RecentEnvironments: recentEnvironments,
		RecentUsers:        recentUsers,
	}, nil
}

func (s *Service) ShopwareVersions(ctx context.Context) ([]api.ShopwareVersionCount, error) {
	rows, err := s.queries.AdminGetShopwareVersions(ctx)
	if err != nil {
		return nil, fmt.Errorf("get shopware versions: %w", err)
	}

	result := make([]api.ShopwareVersionCount, 0, len(rows))
	for _, row := range rows {
		result = append(result, api.ShopwareVersionCount{
			Version: row.Version,
			Count:   int(row.Count),
		})
	}

	return result, nil
}

func (s *Service) OrganizationDetail(ctx context.Context, orgID string) (api.AdminOrganizationDetail, error) {
	org, err := s.queries.AdminGetOrganizationDetail(ctx, orgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.AdminOrganizationDetail{}, ErrOrganizationNotFound
		}
		return api.AdminOrganizationDetail{}, fmt.Errorf("get organization detail: %w", err)
	}

	members, err := s.queries.AdminGetOrganizationMembers(ctx, orgID)
	if err != nil {
		return api.AdminOrganizationDetail{}, fmt.Errorf("get organization members: %w", err)
	}
	environments, err := s.queries.AdminGetOrganizationEnvironments(ctx, orgID)
	if err != nil {
		return api.AdminOrganizationDetail{}, fmt.Errorf("get organization environments: %w", err)
	}
	invitations, err := s.queries.AdminGetOrganizationInvitations(ctx, orgID)
	if err != nil {
		return api.AdminOrganizationDetail{}, fmt.Errorf("get organization invitations: %w", err)
	}
	ssoProviders, err := s.queries.AdminGetOrganizationSSO(ctx, &orgID)
	if err != nil {
		return api.AdminOrganizationDetail{}, fmt.Errorf("get organization sso: %w", err)
	}
	shops, err := s.queries.AdminGetOrganizationShops(ctx, orgID)
	if err != nil {
		return api.AdminOrganizationDetail{}, fmt.Errorf("get organization shops: %w", err)
	}

	detail := api.AdminOrganizationDetail{
		Id:               org.ID,
		Name:             org.Name,
		Slug:             org.Slug,
		Logo:             org.Logo,
		CreatedAt:        database.Time(org.CreatedAt),
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
			CreatedAt: database.Time(m.CreatedAt),
		})
	}
	for _, e := range environments {
		detail.Environments = append(detail.Environments, api.AdminOrganizationEnvironment{
			Id:              int(e.ID),
			Name:            e.Name,
			Url:             e.Url,
			Status:          e.Status,
			ShopwareVersion: e.ShopwareVersion,
			LastScrapedAt:   database.TimePtr(e.LastScrapedAt),
			ShopId:          int(e.ShopID),
			ShopName:        e.ShopName,
		})
	}
	for _, i := range invitations {
		detail.Invitations = append(detail.Invitations, api.AdminOrganizationInvitation{
			Id:           i.ID,
			Email:        i.Email,
			Role:         i.Role,
			Status:       i.Status,
			ExpiresAt:    database.Time(i.ExpiresAt),
			CreatedAt:    database.Time(i.CreatedAt),
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
			CreatedAt:   database.Time(s.CreatedAt),
		}
		if s.DefaultEnvironmentID != nil {
			id := int(*s.DefaultEnvironmentID)
			shop.DefaultEnvironmentId = &id
		}
		detail.Shops = append(detail.Shops, shop)
	}

	return detail, nil
}

// EnvironmentDetail returns an environment with checks, extensions, scheduled tasks and its last deployment. Secrets
// (client_secret, environment_token) are never included.
func (s *Service) EnvironmentDetail(ctx context.Context, environmentID int) (api.AdminEnvironmentDetail, error) {
	env, err := s.queries.AdminGetEnvironmentDetail(ctx, int32(environmentID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.AdminEnvironmentDetail{}, ErrEnvironmentNotFound
		}
		return api.AdminEnvironmentDetail{}, fmt.Errorf("get environment detail: %w", err)
	}

	checks, err := s.queries.AdminGetEnvironmentChecks(ctx, env.ID)
	if err != nil {
		return api.AdminEnvironmentDetail{}, fmt.Errorf("get environment checks: %w", err)
	}
	extensions, err := s.queries.AdminGetEnvironmentExtensions(ctx, env.ID)
	if err != nil {
		return api.AdminEnvironmentDetail{}, fmt.Errorf("get environment extensions: %w", err)
	}
	tasks, err := s.queries.AdminGetEnvironmentTasks(ctx, env.ID)
	if err != nil {
		return api.AdminEnvironmentDetail{}, fmt.Errorf("get environment tasks: %w", err)
	}

	detail := api.AdminEnvironmentDetail{
		Id:                   int(env.ID),
		Name:                 env.Name,
		Url:                  env.Url,
		Status:               env.Status,
		ShopwareVersion:      env.ShopwareVersion,
		CreatedAt:            database.Time(env.CreatedAt),
		LastScrapedAt:        database.TimePtr(env.LastScrapedAt),
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

	if dep, err := s.queries.AdminGetEnvironmentLastDeployment(ctx, env.ID); err == nil {
		detail.LastDeployment = &api.AdminEnvironmentDeployment{
			Id:            int(dep.ID),
			Name:          dep.Name,
			Command:       dep.Command,
			ReturnCode:    int(dep.ReturnCode),
			ExecutionTime: dep.ExecutionTime,
			Reference:     dep.Reference,
			CreatedAt:     database.Time(dep.CreatedAt),
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return api.AdminEnvironmentDetail{}, fmt.Errorf("get environment last deployment: %w", err)
	}

	return detail, nil
}

func (s *Service) AuditLog(ctx context.Context, params api.AdminGetAuditLogParams) (api.AdminAuditLogResponse, error) {
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

	rows, err := s.queries.AdminListAuditLog(ctx, queries.AdminListAuditLogParams{
		Limit:        limit,
		Offset:       offset,
		Action:       action,
		ActorUserID:  actor,
		TargetUserID: target,
	})
	if err != nil {
		return api.AdminAuditLogResponse{}, fmt.Errorf("list audit log: %w", err)
	}

	total, err := s.queries.AdminCountAuditLog(ctx, queries.AdminCountAuditLogParams{
		Action:       action,
		ActorUserID:  actor,
		TargetUserID: target,
	})
	if err != nil {
		return api.AdminAuditLogResponse{}, fmt.Errorf("count audit log: %w", err)
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
			CreatedAt:    database.Time(a.CreatedAt),
		})
	}

	return api.AdminAuditLogResponse{
		Entries: entries,
		Total:   int(total),
	}, nil
}
