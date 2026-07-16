package postgres

import (
	"context"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/jobs"
)

type ScheduleRepository struct {
	queries *queries.Queries
}

var _ jobs.ScheduleRepository = (*ScheduleRepository)(nil)

func NewScheduleRepository(q *queries.Queries) *ScheduleRepository {
	return &ScheduleRepository{queries: q}
}

func (r *ScheduleRepository) ListEnvironmentScrapeTargets(ctx context.Context) ([]jobs.ScheduledEnvironment, error) {
	rows, err := r.queries.GetAllEnvironments(ctx)
	if err != nil {
		return nil, fmt.Errorf("list scheduled environment scrape targets: %w", err)
	}
	targets := make([]jobs.ScheduledEnvironment, 0, len(rows))
	for _, row := range rows {
		targets = append(targets, jobs.ScheduledEnvironment{
			ID:                   row.ID,
			ConnectionIssueCount: row.ConnectionIssueCount,
		})
	}
	return targets, nil
}

func (r *ScheduleRepository) ListSitespeedScrapeTargets(ctx context.Context) ([]int32, error) {
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
