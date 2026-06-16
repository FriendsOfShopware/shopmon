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
	"github.com/jackc/pgx/v5/pgxpool"
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

// seeder bundles the dependencies the per-domain fixture helpers share, so each
// helper reads as a focused unit of seeding rather than threading the same
// handful of parameters everywhere.
type seeder struct {
	ctx  context.Context
	pool *pgxpool.Pool
	q    *queries.Queries
	cfg  *config.Config
	now  time.Time
}

// orgFixture holds the IDs produced while seeding users and the organization,
// consumed by later seeding steps.
type orgFixture struct {
	orgID   string
	user1ID string
	user2ID string
	user3ID string
}

func runFixtures(ctx context.Context, skipShop bool) error {
	cfg := config.Load()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	s := &seeder{
		ctx:  ctx,
		pool: pool,
		q:    queries.New(pool),
		cfg:  cfg,
		now:  time.Now(),
	}

	org, err := s.seedUsersAndOrganization()
	if err != nil {
		return err
	}

	s.seedNotifications(org)

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

	shopID, environmentIDs, err := s.seedShopAndEnvironments(org)
	if err != nil {
		return err
	}

	if err := s.dispatchEnvironmentScrapes(environmentIDs); err != nil {
		return err
	}

	for _, environmentID := range environmentIDs {
		s.seedEnvironmentActivity(environmentID)
	}

	fmt.Println("Fixtures applied successfully")
	fmt.Println("  Owner:   owner@fos.gg (admin role)")
	fmt.Println("  Admin:   admin@fos.gg")
	fmt.Println("  Member:  member@fos.gg")
	fmt.Println("  Regular: regular@fos.gg (no org)")
	fmt.Printf("  Org:     Acme Corp (%s)\n", org.orgID)
	fmt.Printf("  Shop:    Acme Shop (id: %d)\n", shopID)
	fmt.Printf("  Environments: Production (id: %d), Staging (id: %d)\n", environmentIDs[0], environmentIDs[1])
	fmt.Println("  All users have password: password")

	return nil
}

// createUser creates a credential-backed user with the password "password" and
// returns its ID. Admins are promoted after creation.
func (s *seeder) createUser(email, name, role string) (string, error) {
	userID := uuid.New().String()
	_, err := s.q.CreateUser(s.ctx, queries.CreateUserParams{
		ID: userID, Name: name, Email: email,
	})
	if err != nil {
		return "", fmt.Errorf("create user %s: %w", email, err)
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte("password"), 10)
	pw := string(hashed)
	if err := s.q.CreateAccount(s.ctx, queries.CreateAccountParams{
		ID: uuid.New().String(), AccountID: userID,
		ProviderID: "credential", UserID: userID, Password: &pw,
	}); err != nil {
		slog.Error("failed to create account", "email", email, "error", err)
	}

	if err := s.q.UpdateUserEmailVerified(s.ctx, userID); err != nil {
		slog.Error("failed to verify user email", "email", email, "error", err)
	}

	if role == "admin" {
		if _, err := s.pool.Exec(s.ctx, `UPDATE "user" SET role = 'admin' WHERE id = $1`, userID); err != nil {
			slog.Error("failed to set admin role", "email", email, "error", err)
		}
	}

	return userID, nil
}

// seedUsersAndOrganization creates the four demo users plus the Acme Corp
// organization with its three members.
func (s *seeder) seedUsersAndOrganization() (orgFixture, error) {
	var org orgFixture
	var err error

	if org.user1ID, err = s.createUser("owner@fos.gg", "Owner", "admin"); err != nil {
		return org, err
	}
	if org.user2ID, err = s.createUser("admin@fos.gg", "Admin", "user"); err != nil {
		return org, err
	}
	if org.user3ID, err = s.createUser("member@fos.gg", "Member", "user"); err != nil {
		return org, err
	}
	if _, err = s.createUser("regular@fos.gg", "Regular", "user"); err != nil {
		return org, err
	}

	org.orgID = uuid.New().String()
	if _, err := s.q.CreateOrganization(s.ctx, queries.CreateOrganizationParams{
		ID: org.orgID, Name: "Acme Corp", Slug: "acme-corp",
	}); err != nil {
		slog.Error("failed to create organization", "name", "Acme Corp", "error", err)
	}

	members := []struct {
		userID string
		role   string
	}{
		{org.user1ID, "owner"},
		{org.user2ID, "admin"},
		{org.user3ID, "member"},
	}
	for _, m := range members {
		if err := s.q.CreateMember(s.ctx, queries.CreateMemberParams{
			ID: uuid.New().String(), OrganizationID: org.orgID, UserID: m.userID, Role: m.role,
		}); err != nil {
			slog.Error("failed to create member", "role", m.role, "error", err)
		}
	}

	return org, nil
}

// seedNotifications creates a spread of read/unread notifications across the
// three org members to populate the notification feed.
func (s *seeder) seedNotifications(org orgFixture) {
	link := func(label, url string) json.RawMessage {
		return mustJSON(map[string]interface{}{"label": label, "url": url})
	}

	notifications := []struct {
		userID  string
		key     string
		level   string
		title   string
		message string
		link    json.RawMessage
		read    bool
	}{
		{org.user1ID, "environment.change-status.1", "warning", "Environment: Production status changed", "Status changed from green to yellow", link("View Environment", "/environments/1"), false},
		{org.user1ID, "environment.update-auth-error.2", "error", "Environment: Staging could not be updated", "Could not connect to environment. Please check your credentials and try again.", link("View Environment", "/environments/2"), false},
		{org.user1ID, "deployment.completed.1", "info", "Deployment completed", "Theme Update deployment finished successfully.", link("View Deployment", "/environments/1"), true},
		{org.user2ID, "environment.change-status.1", "warning", "Environment: Production status changed", "Status changed from green to yellow", link("View Environment", "/environments/1"), false},
		{org.user2ID, "deployment.failed.1", "error", "Deployment failed", "Hotfix Login deployment failed with exit code 1.", link("View Deployment", "/environments/1"), false},
		{org.user3ID, "environment.not.updated_2", "error", "Environment: Staging could not be updated", "Could not connect to environment. Please check your credentials and try again.", link("View Environment", "/environments/2"), false},
		{org.user1ID, "environment.cache-cleared.1", "info", "Cache cleared", "Production environment cache has been cleared successfully.", link("View Environment", "/environments/1"), true},
		{org.user1ID, "environment.extension-update.1", "info", "Extension update available", "SwagPayPal has a new version 6.1.0 available for Production.", link("View Extensions", "/environments/1"), true},
		{org.user1ID, "deployment.completed.2", "info", "Deployment completed", "Cache Warmup deployment finished successfully.", link("View Deployment", "/environments/1"), true},
		{org.user1ID, "environment.shopware-update.2", "info", "Shopware version updated", "Staging environment updated from Shopware 6.5.1.0 to 6.5.2.0.", link("View Environment", "/environments/2"), true},
		{org.user2ID, "environment.extension-update.2", "info", "Extension update available", "FroshTools has a new version 1.2.0 available for Staging.", link("View Extensions", "/environments/2"), true},
		{org.user2ID, "deployment.completed.3", "info", "Deployment completed", "v6.5.8.0 Release deployment finished successfully.", link("View Deployment", "/environments/1"), true},
		{org.user2ID, "environment.change-status.2", "warning", "Environment: Staging status changed", "Status changed from green to yellow", link("View Environment", "/environments/2"), false},
		{org.user3ID, "deployment.completed.4", "info", "Deployment completed", "Index Rebuild deployment finished successfully.", link("View Deployment", "/environments/1"), true},
		{org.user3ID, "environment.extension-deactivated.2", "warning", "Extension deactivated", "SwagCmsExtensions was deactivated on Staging.", link("View Extensions", "/environments/2"), false},
	}

	for i, n := range notifications {
		if err := s.q.UpsertNotification(s.ctx, queries.UpsertNotificationParams{
			UserID:  n.userID,
			Key:     n.key,
			Level:   n.level,
			Title:   n.title,
			Message: n.message,
			Link:    n.link,
		}); err != nil {
			slog.Warn("failed to create notification", "key", n.key, "error", err)
			continue
		}
		if n.read {
			if _, err := s.pool.Exec(s.ctx,
				`UPDATE user_notification SET read = true WHERE user_id = $1 AND key = $2`,
				n.userID, n.key,
			); err != nil {
				slog.Warn("failed to mark notification as read", "key", n.key, "error", err)
			}
		}
		slog.Info("created notification", "key", n.key, "read", n.read, "userIndex", i)
	}
}

// seedShopAndEnvironments creates the Acme shop with Production and Staging
// environments and sets the first as the shop default.
func (s *seeder) seedShopAndEnvironments(org orgFixture) (int32, []int32, error) {
	shopID, err := s.q.CreateShop(s.ctx, queries.CreateShopParams{
		OrganizationID: org.orgID, Name: "Acme Shop",
	})
	if err != nil {
		return 0, nil, fmt.Errorf("create shop: %w", err)
	}

	const (
		envURL       = "http://localhost:3889"
		clientID     = "SWIAUZL4OXRKEG1RR3PMCEVNMG"
		clientSecret = "aXhNQ3NoRHZONmxPYktHT0c2c09rNkR0UHI0elZHOFIycjBzWks"
	)

	encryptedSecret, err := crypto.Encrypt(clientSecret, s.cfg.AppSecret)
	if err != nil {
		return 0, nil, fmt.Errorf("encrypt secret: %w", err)
	}

	envDefs := []struct {
		name            string
		shopwareVersion string
	}{
		{name: "Production", shopwareVersion: "6.6.0.0"},
		{name: "Staging", shopwareVersion: "6.5.8.0"},
	}

	var environmentIDs []int32
	for _, ed := range envDefs {
		envID, err := s.q.CreateEnvironment(s.ctx, queries.CreateEnvironmentParams{
			OrganizationID:   org.orgID,
			ShopID:           shopID,
			Name:             ed.name,
			Url:              envURL,
			ClientID:         clientID,
			ClientSecret:     encryptedSecret,
			ShopwareVersion:  ed.shopwareVersion,
			EnvironmentToken: uuid.New().String(),
		})
		if err != nil {
			return 0, nil, fmt.Errorf("create environment %s: %w", ed.name, err)
		}
		environmentIDs = append(environmentIDs, envID)
		slog.Info("created environment", "name", ed.name, "id", envID, "url", envURL)
	}

	if len(environmentIDs) > 0 {
		if err := s.q.SetShopDefaultEnvironment(s.ctx, queries.SetShopDefaultEnvironmentParams{
			DefaultEnvironmentID: &environmentIDs[0],
			ID:                   shopID,
			OrganizationID:       org.orgID,
		}); err != nil {
			slog.Error("failed to set default environment", "shopId", shopID, "error", err)
		}
	}

	return shopID, environmentIDs, nil
}

// dispatchEnvironmentScrapes enqueues an initial scrape for each environment so
// the worker populates them immediately.
func (s *seeder) dispatchEnvironmentScrapes(environmentIDs []int32) error {
	mailSvc := mail.NewService(mail.SMTPConfig{
		Host: s.cfg.SMTPHost, Port: s.cfg.SMTPPort, Secure: s.cfg.SMTPSecure,
		User: s.cfg.SMTPUser, Pass: s.cfg.SMTPPass, From: s.cfg.MailFrom, ReplyTo: s.cfg.SMTPReplyTo,
	})
	bus, err := jobs.NewBus(s.ctx, s.pool, s.q, s.cfg, mailSvc)
	if err != nil {
		return fmt.Errorf("create queue bus: %w", err)
	}
	for _, envID := range environmentIDs {
		if err := goqueue.Dispatch(s.ctx, bus, jobs.EnvironmentScrape{EnvironmentID: envID}); err != nil {
			slog.Warn("failed to dispatch environment scrape", "environmentId", envID, "error", err)
		} else {
			slog.Info("dispatched environment scrape task", "environmentId", envID)
		}
	}
	return nil
}

// seedEnvironmentActivity populates one environment with a week of deployments,
// sitespeed samples, and changelog entries, then marks the latest deployment active.
func (s *seeder) seedEnvironmentActivity(environmentID int32) {
	deployments := s.seedDeployments(environmentID)
	s.seedSitespeed(environmentID, deployments)
	s.seedChangelogs(environmentID)

	if len(deployments) > 0 {
		if err := s.q.UpdateEnvironmentActiveDeployment(s.ctx, queries.UpdateEnvironmentActiveDeploymentParams{
			ActiveDeploymentID: &deployments[0].id,
			ID:                 environmentID,
		}); err != nil {
			slog.Error("failed to update environment active deployment", "environmentId", environmentID, "error", err)
		}
	}
}

type deploymentRef struct {
	id        int32
	createdAt time.Time
}

// seedDeployments creates 8 deployments spread over the last 7 days (one failed).
func (s *seeder) seedDeployments(environmentID int32) []deploymentRef {
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

	var deployments []deploymentRef
	for i := 0; i < 8; i++ {
		hoursAgo := (7 * 24 * i) / 8
		startDate := s.now.Add(-time.Duration(hoursAgo) * time.Hour)
		execTime := rand.Float32()*120 + 5
		endDate := startDate.Add(time.Duration(execTime) * time.Second)
		returnCode := int32(0)
		if i == 3 {
			returnCode = 1
		}

		ref := fmt.Sprintf("abc%d", 1000+i)
		depID, err := s.q.CreateDeployment(s.ctx, queries.CreateDeploymentParams{
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
	return deployments
}

// seedSitespeed creates sitespeed samples every ~6h over the last 7 days, with a
// performance penalty applied to samples taken shortly after a deployment.
func (s *seeder) seedSitespeed(environmentID int32, deployments []deploymentRef) {
	const (
		baseTTFB         = 120.0
		baseFullyLoaded  = 2800.0
		baseLCP          = 1800.0
		baseFCP          = 900.0
		baseCLS          = 0.05
		baseTransferSize = 1200000.0
	)
	count := 0

	for hoursAgo := 7 * 24; hoursAgo >= 0; hoursAgo -= 6 {
		createdAt := s.now.Add(-time.Duration(hoursAgo) * time.Hour)

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

		if err := s.q.InsertEnvironmentSitespeed(s.ctx, queries.InsertEnvironmentSitespeedParams{
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
		count++
	}

	slog.Info("created sitespeed entries", "environmentId", environmentID, "count", count)
}

// seedChangelogs creates a series of extension/Shopware-version changelog entries
// over the last week.
func (s *seeder) seedChangelogs(environmentID int32) {
	v650 := "6.5.0.0"
	v651 := "6.5.1.0"
	v652 := "6.5.2.0"
	v660 := "6.6.0.0"

	changelogs := []struct {
		extensions         json.RawMessage
		oldShopwareVersion *string
		newShopwareVersion *string
		date               time.Time
	}{
		{
			extensions: mustJSON([]map[string]interface{}{
				{"name": "SwagPayPal", "label": "PayPal", "state": "updated", "oldVersion": "5.0.0", "newVersion": "5.1.0", "active": true},
				{"name": "SwagCmsExtensions", "label": "CMS Extensions", "state": "activated", "active": true},
			}),
			oldShopwareVersion: &v650,
			newShopwareVersion: &v651,
			date:               s.now.Add(-6 * 24 * time.Hour),
		},
		{
			extensions: mustJSON([]map[string]interface{}{
				{"name": "SwagPayPal", "label": "PayPal", "state": "updated", "oldVersion": "5.1.0", "newVersion": "5.2.0", "active": true},
				{"name": "FroshTools", "label": "FroshTools", "state": "installed", "newVersion": "1.0.0", "active": true},
			}),
			date: s.now.Add(-4 * 24 * time.Hour),
		},
		{
			extensions: mustJSON([]map[string]interface{}{
				{"name": "SwagCmsExtensions", "label": "CMS Extensions", "state": "deactivated", "active": false},
			}),
			oldShopwareVersion: &v651,
			newShopwareVersion: &v652,
			date:               s.now.Add(-2 * 24 * time.Hour),
		},
		{
			extensions: mustJSON([]map[string]interface{}{
				{"name": "FroshTools", "label": "FroshTools", "state": "updated", "oldVersion": "1.0.0", "newVersion": "1.1.0", "active": true},
				{"name": "SwagPayPal", "label": "PayPal", "state": "updated", "oldVersion": "5.2.0", "newVersion": "6.0.0", "active": true},
			}),
			oldShopwareVersion: &v652,
			newShopwareVersion: &v660,
			date:               s.now.Add(-12 * time.Hour),
		},
	}

	count := 0
	for _, cl := range changelogs {
		if _, err := s.pool.Exec(s.ctx,
			`INSERT INTO environment_changelog (environment_id, extensions, old_shopware_version, new_shopware_version, date) VALUES ($1, $2, $3, $4, $5)`,
			environmentID, cl.extensions, cl.oldShopwareVersion, cl.newShopwareVersion, cl.date,
		); err != nil {
			slog.Warn("failed to create changelog entry", "environmentId", environmentID, "error", err)
			continue
		}
		count++
	}

	slog.Info("created changelog entries", "environmentId", environmentID, "count", count)
}

func mustJSON(v interface{}) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
