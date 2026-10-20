package jobs

import (
	"context"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
)

type scheduleStore struct {
	queries *queries.Queries
}

var _ ScheduleRepository = (*scheduleStore)(nil)

func (r *scheduleStore) ListEnvironmentScrapeTargets(ctx context.Context) ([]ScheduledEnvironment, error) {
	rows, err := r.queries.GetAllEnvironments(ctx)
	if err != nil {
		return nil, fmt.Errorf("list scheduled environment scrape targets: %w", err)
	}
	targets := make([]ScheduledEnvironment, 0, len(rows))
	for _, row := range rows {
		targets = append(targets, ScheduledEnvironment{
			ID:                   row.ID,
			ConnectionIssueCount: row.ConnectionIssueCount,
		})
	}
	return targets, nil
}

func (r *scheduleStore) ListSitespeedScrapeTargets(ctx context.Context) ([]int32, error) {
	rows, err := r.queries.GetEnvironmentsWithSitespeedEnabled(ctx)
	if err != nil {
		return nil, fmt.Errorf("list scheduled sitespeed scrape targets: %w", err)
	}
	targets := make([]int32, 0, len(rows))
	for _, row := range rows {
		targets = append(targets, row.ID)
	}
	return targets, nil
}
