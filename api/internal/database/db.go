package database

import (
	"context"
	"fmt"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	config.MinConns = 2
	config.MaxConns = 20
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.ConnConfig.ConnectTimeout = 10 * time.Second
	// SQLSpanName prefers the sqlc query name, else the first SQL keyword.
	// WithDisableAcquireTracer drops pool.acquire / client.request spans: they
	// are high-volume and low-signal in Datadog compared to query/exec/prepare.
	config.ConnConfig.Tracer = otelpgx.NewTracer(
		otelpgx.WithSpanNameFunc(SQLSpanName),
		otelpgx.WithDisableAcquireTracer(),
	)

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}
