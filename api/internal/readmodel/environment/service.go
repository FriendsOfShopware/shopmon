package environment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/jackc/pgx/v5"
	"golang.org/x/sync/errgroup"
)

var (
	ErrNotFound      = errors.New("environment not found")
	ErrNotAuthorized = errors.New("not authorized for organization")
)

type MembershipAuthorizer interface {
	IsMember(ctx context.Context, userID, organizationID string) (bool, error)
}

type Service struct {
	queries    *queries.Queries
	authorizer MembershipAuthorizer
	config     *config.Config
}

func NewService(q *queries.Queries, authorizer MembershipAuthorizer, cfg *config.Config) *Service {
	return &Service{queries: q, authorizer: authorizer, config: cfg}
}

func (s *Service) OrganizationEnvironments(ctx context.Context, userID, organizationID string) ([]api.AccountEnvironment, error) {
	member, err := s.authorizer.IsMember(ctx, userID, organizationID)
	if err != nil {
		return nil, fmt.Errorf("authorize organization membership: %w", err)
	}
	if !member {
		return nil, ErrNotAuthorized
	}
	rows, err := s.queries.ListEnvironmentsByOrganization(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("list organization environments: %w", err)
	}
	return toAccountEnvironments(rows), nil
}

func (s *Service) OrganizationShops(ctx context.Context, userID, organizationID string) ([]api.Shop, error) {
	member, err := s.authorizer.IsMember(ctx, userID, organizationID)
	if err != nil {
		return nil, fmt.Errorf("authorize organization membership: %w", err)
	}
	if !member {
		return nil, ErrNotAuthorized
	}
	rows, err := s.queries.ListShopsByOrganization(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("list organization shops: %w", err)
	}
	result := make([]api.Shop, 0, len(rows))
	for _, row := range rows {
		result = append(result, api.Shop{
			Id:                   int(row.ID),
			Name:                 row.Name,
			Description:          row.Description,
			GitUrl:               row.GitUrl,
			OrganizationId:       row.OrganizationID,
			DefaultEnvironmentId: int32PointerToInt(row.DefaultEnvironmentID),
		})
	}
	return result, nil
}

func (s *Service) Detail(ctx context.Context, userID string, environmentID int32, requestedLanguage *string, subscribed bool) (api.EnvironmentDetail, error) {
	environment, err := s.queries.GetEnvironmentByID(ctx, environmentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.EnvironmentDetail{}, ErrNotFound
		}
		return api.EnvironmentDetail{}, fmt.Errorf("get environment: %w", err)
	}
	member, err := s.authorizer.IsMember(ctx, userID, environment.OrganizationID)
	if err != nil {
		return api.EnvironmentDetail{}, fmt.Errorf("authorize environment organization: %w", err)
	}
	if !member {
		return api.EnvironmentDetail{}, ErrNotAuthorized
	}

	aggregate, err := s.loadAggregate(ctx, environmentID, resolveLanguage(requestedLanguage))
	if err != nil {
		return api.EnvironmentDetail{}, err
	}
	aggregate.subscribed = subscribed
	return s.buildDetail(&environment, aggregate)
}

type detailAggregate struct {
	extensions  []api.EnvironmentExtension
	tasks       []queries.EnvironmentScheduledTask
	queues      []queries.EnvironmentQueue
	cache       *queries.EnvironmentCache
	checks      []queries.EnvironmentCheck
	sitespeeds  []queries.EnvironmentSitespeed
	changelogs  []queries.EnvironmentChangelog
	deployCount int32
	subscribed  bool
}

func (s *Service) loadAggregate(ctx context.Context, environmentID int32, language string) (detailAggregate, error) {
	var aggregate detailAggregate
	group, groupCtx := errgroup.WithContext(ctx)

	group.Go(func() error {
		extensions, err := s.loadExtensions(groupCtx, environmentID, language)
		aggregate.extensions = extensions
		return err
	})
	group.Go(func() error {
		rows, err := s.queries.GetEnvironmentScheduledTasks(groupCtx, environmentID)
		if err != nil {
			return fmt.Errorf("get environment scheduled tasks: %w", err)
		}
		aggregate.tasks = rows
		return nil
	})
	group.Go(func() error {
		rows, err := s.queries.GetEnvironmentQueues(groupCtx, environmentID)
		if err != nil {
			return fmt.Errorf("get environment queues: %w", err)
		}
		aggregate.queues = rows
		return nil
	})
	group.Go(func() error {
		row, err := s.queries.GetEnvironmentCache(groupCtx, environmentID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("get environment cache: %w", err)
		}
		aggregate.cache = &row
		return nil
	})
	group.Go(func() error {
		rows, err := s.queries.GetEnvironmentChecks(groupCtx, environmentID)
		if err != nil {
			return fmt.Errorf("get environment checks: %w", err)
		}
		aggregate.checks = rows
		return nil
	})
	group.Go(func() error {
		rows, err := s.queries.GetEnvironmentSitespeeds(groupCtx, &environmentID)
		if err != nil {
			return fmt.Errorf("get environment sitespeeds: %w", err)
		}
		aggregate.sitespeeds = rows
		return nil
	})
	group.Go(func() error {
		rows, err := s.queries.GetEnvironmentChangelogs(groupCtx, &environmentID)
		if err != nil {
			return fmt.Errorf("get environment changelogs: %w", err)
		}
		aggregate.changelogs = rows
		return nil
	})
	group.Go(func() error {
		count, err := s.queries.CountEnvironmentDeployments(groupCtx, environmentID)
		if err != nil {
			return fmt.Errorf("count environment deployments: %w", err)
		}
		aggregate.deployCount = count
		return nil
	})

	if err := group.Wait(); err != nil {
		return detailAggregate{}, err
	}
	return aggregate, nil
}

func (s *Service) loadExtensions(ctx context.Context, environmentID int32, language string) ([]api.EnvironmentExtension, error) {
	storeRows, err := s.queries.GetEnvironmentStoreExtensions(ctx, queries.GetEnvironmentStoreExtensionsParams{
		EnvironmentID: environmentID,
		Language:      language,
	})
	if err != nil {
		return nil, fmt.Errorf("get environment store extensions: %w", err)
	}
	unknownRows, err := s.queries.GetEnvironmentExtensions(ctx, environmentID)
	if err != nil {
		return nil, fmt.Errorf("get environment extensions: %w", err)
	}
	changelogRows, err := s.queries.GetEnvironmentStoreExtensionChangelogs(ctx, queries.GetEnvironmentStoreExtensionChangelogsParams{
		EnvironmentID: environmentID,
		Language:      language,
	})
	if err != nil {
		return nil, fmt.Errorf("get environment store extension changelogs: %w", err)
	}
	return mergeEnvironmentExtensions(storeRows, unknownRows, changelogRows), nil
}

func (s *Service) buildDetail(environment *queries.GetEnvironmentByIDRow, aggregate detailAggregate) (api.EnvironmentDetail, error) {
	var ignores *[]string
	if len(environment.Ignores) > 0 {
		var values []string
		if err := json.Unmarshal(environment.Ignores, &values); err != nil {
			return api.EnvironmentDetail{}, fmt.Errorf("decode environment ignores: %w", err)
		}
		ignores = &values
	}

	var sitespeedURLs *[]string
	if len(environment.SitespeedUrls) > 0 {
		var values []string
		if err := json.Unmarshal(environment.SitespeedUrls, &values); err != nil {
			return api.EnvironmentDetail{}, fmt.Errorf("decode sitespeed URLs: %w", err)
		}
		sitespeedURLs = &values
	}

	var lastChangelog *api.AccountChangelog
	if len(environment.LastChangelog) > 0 {
		var changelog api.AccountChangelog
		if err := json.Unmarshal(environment.LastChangelog, &changelog); err != nil {
			return api.EnvironmentDetail{}, fmt.Errorf("decode last changelog: %w", err)
		}
		if !changelog.Date.IsZero() {
			lastChangelog = &changelog
		}
	}

	changelogs, err := mapEnvironmentChangelogs(environment, aggregate.changelogs)
	if err != nil {
		return api.EnvironmentDetail{}, err
	}
	shopID := int(environment.ShopID)
	var shopIDPointer *int
	if environment.ShopID > 0 {
		shopIDPointer = &shopID
	}

	return api.EnvironmentDetail{
		Id:                 int(environment.ID),
		Name:               environment.Name,
		Url:                environment.Url,
		Favicon:            environment.Favicon,
		Status:             environment.Status,
		ShopwareVersion:    environment.ShopwareVersion,
		LastScrapedAt:      pgtimeToTimePtr(environment.LastScrapedAt),
		LastScrapedError:   environment.LastScrapedError,
		OrganizationId:     environment.OrganizationID,
		OrganizationName:   environment.OrganizationName,
		ShopId:             shopIDPointer,
		ShopName:           environment.ShopName,
		ShopDescription:    environment.ShopDescription,
		EnvironmentImage:   environment.EnvironmentImage,
		EnvironmentToken:   environment.EnvironmentToken,
		Ignores:            ignores,
		CreatedAt:          pgtimeToTime(environment.CreatedAt),
		SitespeedEnabled:   environment.SitespeedEnabled,
		SitespeedDetailUrl: sitespeedDetailURL(s.config, environment.ID, environment.SitespeedEnabled),
		SitespeedUrls:      sitespeedURLs,
		Extensions:         aggregate.extensions,
		ScheduledTasks:     mapEnvironmentScheduledTasks(aggregate.tasks),
		Queues:             mapEnvironmentQueues(aggregate.queues),
		Cache:              mapEnvironmentCache(aggregate.cache),
		Checks:             mapEnvironmentChecks(aggregate.checks),
		Sitespeeds:         mapEnvironmentSitespeeds(aggregate.sitespeeds),
		Changelogs:         changelogs,
		DeploymentsCount:   int(aggregate.deployCount),
		LastChangelog:      lastChangelog,
		Subscribed:         aggregate.subscribed,
	}, nil
}

func sitespeedDetailURL(cfg *config.Config, environmentID int32, enabled bool) *string {
	if !enabled || cfg.SitespeedEndpoint == "" {
		return nil
	}
	value := fmt.Sprintf("%s/result/%senvironment-%d/index.html", cfg.SitespeedEndpoint, cfg.SitespeedPrefix, environmentID)
	return &value
}

func int32PointerToInt(value *int32) *int {
	if value == nil {
		return nil
	}
	converted := int(*value)
	return &converted
}

// StatusEvents returns the recent status change timeline for an environment.
func (s *Service) StatusEvents(ctx context.Context, userID string, environmentID int32) ([]api.StatusEvent, error) {
	environment, err := s.queries.GetEnvironmentByID(ctx, environmentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get environment: %w", err)
	}
	member, err := s.authorizer.IsMember(ctx, userID, environment.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("authorize organization membership: %w", err)
	}
	if !member {
		return nil, ErrNotAuthorized
	}

	rows, err := s.queries.ListEnvironmentStatusEvents(ctx, environmentID)
	if err != nil {
		return nil, fmt.Errorf("list environment status events: %w", err)
	}

	events := make([]api.StatusEvent, 0, len(rows))
	for _, row := range rows {
		reasons := []api.StatusReason{}
		if len(row.Reasons) > 0 {
			if err := json.Unmarshal(row.Reasons, &reasons); err != nil {
				reasons = []api.StatusReason{}
			}
		}
		events = append(events, api.StatusEvent{
			Id:        int(row.ID),
			OldStatus: row.OldStatus,
			NewStatus: row.NewStatus,
			Reasons:   reasons,
			CreatedAt: pgtimeToTime(row.CreatedAt),
		})
	}
	return events, nil
}
