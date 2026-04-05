package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/crypto"
	"github.com/friendsofshopware/shopmon/api/internal/database"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/jobs"
	"github.com/friendsofshopware/shopmon/api/internal/mail"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	goqueue "github.com/shyim/go-queue"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

func fixturesCmd() *cobra.Command {
	var skipShop bool

	cmd := &cobra.Command{
		Use:   "fixtures",
		Short: "Apply test fixtures (users, org, shop, deployments)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFixtures(cmd.Context(), skipShop)
		},
	}

	cmd.Flags().BoolVar(&skipShop, "skip-shop", false, "Skip shop creation")
	return cmd
}

func runFixtures(ctx context.Context, skipShop bool) error {
	cfg := config.Load()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	q := queries.New(pool)
	now := time.Now()
	pgNow := pgtype.Timestamp{Time: now, Valid: true}

	// Helper to create a user with password
	createUser := func(email, name, role string) (string, error) {
		userID := uuid.New().String()
		_, err := q.CreateUser(ctx, queries.CreateUserParams{
			ID: userID, Name: name, Email: email,
		})
		if err != nil {
			return "", fmt.Errorf("create user %s: %w", email, err)
		}

		hashed, _ := bcrypt.GenerateFromPassword([]byte("password"), 10)
		pw := string(hashed)
		if err := q.CreateAccount(ctx, queries.CreateAccountParams{
			ID: uuid.New().String(), AccountID: userID,
			ProviderID: "credential", UserID: userID, Password: &pw,
		}); err != nil {
			slog.Error("failed to create account", "email", email, "error", err)
		}

		// Mark email verified
		if err := q.UpdateUserEmailVerified(ctx, userID); err != nil {
			slog.Error("failed to verify user email", "email", email, "error", err)
		}

		if role == "admin" {
			if _, err := pool.Exec(ctx, `UPDATE "user" SET role = 'admin' WHERE id = $1`, userID); err != nil {
				slog.Error("failed to set admin role", "email", email, "error", err)
			}
		}

		return userID, nil
	}

	// Create users
	user1ID, err := createUser("owner@fos.gg", "Owner", "admin")
	if err != nil {
		return err
	}
	user2ID, err := createUser("admin@fos.gg", "Admin", "user")
	if err != nil {
		return err
	}
	user3ID, err := createUser("member@fos.gg", "Member", "user")
	if err != nil {
		return err
	}
	_, err = createUser("regular@fos.gg", "Regular", "user")
	if err != nil {
		return err
	}

	// Create organization
	orgID := uuid.New().String()
	if _, err := q.CreateOrganization(ctx, queries.CreateOrganizationParams{
		ID: orgID, Name: "Acme Corp", Slug: "acme-corp",
	}); err != nil {
		slog.Error("failed to create organization", "name", "Acme Corp", "error", err)
	}

	if err := q.CreateMember(ctx, queries.CreateMemberParams{
		ID: uuid.New().String(), OrganizationID: orgID, UserID: user1ID, Role: "owner",
	}); err != nil {
		slog.Error("failed to create member", "user", "owner", "error", err)
	}
	if err := q.CreateMember(ctx, queries.CreateMemberParams{
		ID: uuid.New().String(), OrganizationID: orgID, UserID: user2ID, Role: "admin",
	}); err != nil {
		slog.Error("failed to create member", "user", "admin", "error", err)
	}
	if err := q.CreateMember(ctx, queries.CreateMemberParams{
		ID: uuid.New().String(), OrganizationID: orgID, UserID: user3ID, Role: "member",
	}); err != nil {
		slog.Error("failed to create member", "user", "member", "error", err)
	}

	slog.Info("created users and organization",
		"org", "Acme Corp",
		"owner", "owner@fos.gg",
		"admin", "admin@fos.gg",
		"member", "member@fos.gg",
		"regular", "regular@fos.gg",
	)

	if skipShop {
		slog.Info("skipping shop creation (--skip-shop)")
		fmt.Println("Fixtures applied. All users have password: password")
		return nil
	}

	// Create shop
	shopID, err := q.CreateShop(ctx, queries.CreateShopParams{
		OrganizationID: orgID, Name: "Acme Shop",
	})
	if err != nil {
		return fmt.Errorf("create shop: %w", err)
	}

	// Create environments (Production + Staging on the same shop)
	envURL := "http://localhost:3889"
	clientID := "SWIAUZL4OXRKEG1RR3PMCEVNMG"
	clientSecret := "aXhNQ3NoRHZONmxPYktHT0c2c09rNkR0UHI0elZHOFIycjBzWks"

	encryptedSecret, err := crypto.Encrypt(clientSecret, cfg.AppSecret)
	if err != nil {
		return fmt.Errorf("encrypt secret: %w", err)
	}

	type envDef struct {
		name            string
		shopwareVersion string
	}
	envDefs := []envDef{
		{name: "Production", shopwareVersion: "6.6.0.0"},
		{name: "Staging", shopwareVersion: "6.5.8.0"},
	}

	var environmentIDs []int32

	for _, ed := range envDefs {
		envID, err := q.CreateEnvironment(ctx, queries.CreateEnvironmentParams{
			OrganizationID:   orgID,
			ShopID:           shopID,
			Name:             ed.name,
			Url:              envURL,
			ClientID:         clientID,
			ClientSecret:     encryptedSecret,
			ShopwareVersion:  ed.shopwareVersion,
			EnvironmentToken: uuid.New().String(),
		})
		if err != nil {
			return fmt.Errorf("create environment %s: %w", ed.name, err)
		}
		environmentIDs = append(environmentIDs, envID)
		slog.Info("created environment", "name", ed.name, "id", envID, "url", envURL)
	}

	// Set the first environment as the shop's default
	if len(environmentIDs) > 0 {
		if err := q.SetShopDefaultEnvironment(ctx, queries.SetShopDefaultEnvironmentParams{
			DefaultEnvironmentID: &environmentIDs[0],
			ID:                   shopID,
			OrganizationID:       orgID,
		}); err != nil {
			slog.Error("failed to set default environment", "shopId", shopID, "error", err)
		}
	}

	// Dispatch environment scrape tasks so the worker picks them up immediately
	mailSvc := mail.NewService(mail.SMTPConfig{
		Host: cfg.SMTPHost, Port: cfg.SMTPPort, Secure: cfg.SMTPSecure,
		User: cfg.SMTPUser, Pass: cfg.SMTPPass, From: cfg.MailFrom, ReplyTo: cfg.SMTPReplyTo,
	})
	bus, err := jobs.NewBus(ctx, pool, q, cfg, mailSvc)
	if err != nil {
		return fmt.Errorf("create queue bus: %w", err)
	}
	for _, envID := range environmentIDs {
		if err := goqueue.Dispatch(ctx, bus, jobs.EnvironmentScrape{EnvironmentID: envID}); err != nil {
			slog.Warn("failed to dispatch environment scrape", "environmentId", envID, "error", err)
		} else {
			slog.Info("dispatched environment scrape task", "environmentId", envID)
		}
	}

	// Create deployments, sitespeed, and changelogs for each environment
	commands := []string{
		"bin/console app:install", "bin/console theme:compile",
		"bin/console cache:clear", "bin/console dal:refresh:index",
		"bin/console plugin:update", "bin/console scheduled-task:run",
		"bin/console messenger:consume", "bin/console assets:install",
	}
	deploymentNames := []string{
		"v6.5.8.0 Release", "Hotfix Login", "Theme Update", "Plugin Refresh",
		"Index Rebuild", "Cache Warmup", "Messenger Fix", "Asset Deploy",
	}

	type deploymentRef struct {
		id        int32
		createdAt time.Time
	}

	type changelogEntry struct {
		extensions         json.RawMessage
		oldShopwareVersion *string
		newShopwareVersion *string
		date               time.Time
	}

	v650 := "6.5.0.0"
	v651 := "6.5.1.0"
	v652 := "6.5.2.0"
	v660 := "6.6.0.0"

	for _, environmentID := range environmentIDs {
		// Deployments over the last 7 days
		var deployments []deploymentRef

		for i := 0; i < 8; i++ {
			hoursAgo := (7 * 24 * i) / 8
			startDate := now.Add(-time.Duration(hoursAgo) * time.Hour)
			execTime := rand.Float32()*120 + 5
			endDate := startDate.Add(time.Duration(execTime) * time.Second)
			returnCode := int32(0)
			if i == 3 {
				returnCode = 1
			}

			ref := fmt.Sprintf("abc%d", 1000+i)
			depID, err := q.CreateDeployment(ctx, queries.CreateDeploymentParams{
				EnvironmentID: environmentID,
				Name:          deploymentNames[i],
				Command:       commands[i],
				ReturnCode:    returnCode,
				StartDate:     pgtype.Timestamp{Time: startDate, Valid: true},
				EndDate:       pgtype.Timestamp{Time: endDate, Valid: true},
				ExecutionTime: execTime,
				Reference:     &ref,
			})
			if err != nil {
				slog.Warn("failed to create deployment", "environmentId", environmentID, "error", err)
				continue
			}
			deployments = append(deployments, deploymentRef{id: depID, createdAt: startDate})
		}

		slog.Info("created deployments", "environmentId", environmentID, "count", len(deployments))

		// Sitespeed entries over the last 7 days (every ~6 hours)
		baseTTFB := 120.0
		baseFullyLoaded := 2800.0
		baseLCP := 1800.0
		baseFCP := 900.0
		baseCLS := 0.05
		baseTransferSize := 1200000.0
		sitespeedCount := 0

		for hoursAgo := 7 * 24; hoursAgo >= 0; hoursAgo -= 6 {
			createdAt := now.Add(-time.Duration(hoursAgo) * time.Hour)

			var deploymentID *int32
			deploymentPenalty := 1.0
			for _, d := range deployments {
				diff := createdAt.Sub(d.createdAt)
				if diff >= 0 && diff < 2*time.Hour {
					id := d.id
					deploymentID = &id
					deploymentPenalty = 1.3
					break
				}
			}

			jitter := func() float64 { return 0.85 + rand.Float64()*0.3 }

			ttfb := int32(math.Round(baseTTFB * jitter() * deploymentPenalty))
			fullyLoaded := int32(math.Round(baseFullyLoaded * jitter() * deploymentPenalty))
			lcp := int32(math.Round(baseLCP * jitter() * deploymentPenalty))
			fcp := int32(math.Round(baseFCP * jitter() * deploymentPenalty))
			cls := float32(math.Round(baseCLS*jitter()*deploymentPenalty*1000) / 1000)
			transferSize := int32(math.Round(baseTransferSize * jitter()))

			if err := q.InsertEnvironmentSitespeed(ctx, queries.InsertEnvironmentSitespeedParams{
				EnvironmentID:          &environmentID,
				DeploymentID:           deploymentID,
				Ttfb:                   &ttfb,
				FullyLoaded:            &fullyLoaded,
				LargestContentfulPaint: &lcp,
				FirstContentfulPaint:   &fcp,
				CumulativeLayoutShift:  &cls,
				TransferSize:           &transferSize,
			}); err != nil {
				slog.Error("failed to insert sitespeed entry", "environmentId", environmentID, "error", err)
			}
			sitespeedCount++
		}

		slog.Info("created sitespeed entries", "environmentId", environmentID, "count", sitespeedCount)

		// Changelog entries over the last 7 days
		changelogs := []changelogEntry{
			{
				extensions: mustJSON([]map[string]interface{}{
					{"name": "SwagPayPal", "label": "PayPal", "state": "updated", "old_version": "5.0.0", "new_version": "5.1.0", "active": true},
					{"name": "SwagCmsExtensions", "label": "CMS Extensions", "state": "activated", "active": true},
				}),
				oldShopwareVersion: &v650,
				newShopwareVersion: &v651,
				date:               now.Add(-6 * 24 * time.Hour),
			},
			{
				extensions: mustJSON([]map[string]interface{}{
					{"name": "SwagPayPal", "label": "PayPal", "state": "updated", "old_version": "5.1.0", "new_version": "5.2.0", "active": true},
					{"name": "FroshTools", "label": "FroshTools", "state": "installed", "new_version": "1.0.0", "active": true},
				}),
				date: now.Add(-4 * 24 * time.Hour),
			},
			{
				extensions: mustJSON([]map[string]interface{}{
					{"name": "SwagCmsExtensions", "label": "CMS Extensions", "state": "deactivated", "active": false},
				}),
				oldShopwareVersion: &v651,
				newShopwareVersion: &v652,
				date:               now.Add(-2 * 24 * time.Hour),
			},
			{
				extensions: mustJSON([]map[string]interface{}{
					{"name": "FroshTools", "label": "FroshTools", "state": "updated", "old_version": "1.0.0", "new_version": "1.1.0", "active": true},
					{"name": "SwagPayPal", "label": "PayPal", "state": "updated", "old_version": "5.2.0", "new_version": "6.0.0", "active": true},
				}),
				oldShopwareVersion: &v652,
				newShopwareVersion: &v660,
				date:               now.Add(-12 * time.Hour),
			},
		}

		changelogCount := 0
		for _, cl := range changelogs {
			if _, err := pool.Exec(ctx,
				`INSERT INTO environment_changelog (environment_id, extensions, old_shopware_version, new_shopware_version, date) VALUES ($1, $2, $3, $4, $5)`,
				environmentID, cl.extensions, cl.oldShopwareVersion, cl.newShopwareVersion, cl.date,
			); err != nil {
				slog.Warn("failed to create changelog entry", "environmentId", environmentID, "error", err)
				continue
			}
			changelogCount++
		}

		slog.Info("created changelog entries", "environmentId", environmentID, "count", changelogCount)

		if len(deployments) > 0 {
			if err := q.UpdateEnvironmentActiveDeployment(ctx, queries.UpdateEnvironmentActiveDeploymentParams{
				ActiveDeploymentID: &deployments[0].id,
				ID:                 environmentID,
			}); err != nil {
				slog.Error("failed to update environment active deployment", "environmentId", environmentID, "error", err)
			}
		}
	}

	fmt.Println("Fixtures applied successfully")
	fmt.Println("  Owner:   owner@fos.gg (admin role)")
	fmt.Println("  Admin:   admin@fos.gg")
	fmt.Println("  Member:  member@fos.gg")
	fmt.Println("  Regular: regular@fos.gg (no org)")
	fmt.Printf("  Org:     Acme Corp (%s)\n", orgID)
	fmt.Printf("  Shop:    Acme Shop (id: %d)\n", shopID)
	fmt.Printf("  Environments: Production (id: %d), Staging (id: %d)\n", environmentIDs[0], environmentIDs[1])
	fmt.Println("  All users have password: password")

	_ = pgNow
	return nil
}

func mustJSON(v interface{}) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
