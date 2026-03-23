package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/spf13/cobra"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func newMigrate(databaseURL string) (*migrate.Migrate, error) {
	fsys, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("sub fs: %w", err)
	}

	source, err := iofs.New(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("create source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create migrate: %w", err)
	}

	return m, nil
}

func migrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration commands",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "up",
		Short: "Run all pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			m, err := newMigrate(cfg.DatabaseURL)
			if err != nil {
				return err
			}
			defer m.Close()

			if err := m.Up(); err != nil && err != migrate.ErrNoChange {
				return err
			}

			slog.Info("migrations applied")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "down",
		Short: "Rollback the last migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			m, err := newMigrate(cfg.DatabaseURL)
			if err != nil {
				return err
			}
			defer m.Close()

			if err := m.Steps(-1); err != nil {
				return err
			}

			slog.Info("migration rolled back")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show current migration version",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			m, err := newMigrate(cfg.DatabaseURL)
			if err != nil {
				return err
			}
			defer m.Close()

			version, dirty, err := m.Version()
			if err != nil {
				if err == migrate.ErrNilVersion {
					fmt.Println("No migrations applied yet")
					return nil
				}
				return err
			}

			fmt.Printf("Version: %d, Dirty: %v\n", version, dirty)
			return nil
		},
	})

	return cmd
}
