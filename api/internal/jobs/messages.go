package jobs

const (
	// TransportName is the legacy async transport (`queue_messages` / AMQP
	// `QUEUE_AMQP_QUEUE`). StoreExtensionSync stays here so a catalog backlog
	// does not share claim-order with shop-facing jobs.
	TransportName = "async"
	// ScrapeTransportName is the shop-facing job transport
	// (`queue_messages_scrape` / AMQP `<QUEUE_AMQP_QUEUE>-scrape`).
	// Environment scrapes, sitespeed, and the other scheduled jobs publish here.
	//
	// go-queue's postgres consumer claims `ORDER BY created_at` and retries
	// keep the original created_at, so StoreExtensionSync 429 retries on the
	// shared table stay ahead of newer EnvironmentScrape rows. Production
	// (thousands of shops) starves shop refreshes; staging (two shops) does not.
	ScrapeTransportName = "scrape"
)

// Message types are deliberately kept in this package so their reflected type
// names remain stable for envelopes already persisted in PostgreSQL.

type EnvironmentScrape struct {
	EnvironmentID int32 `json:"environment_id"`
}

type SitespeedScrape struct {
	EnvironmentID int32 `json:"environment_id"`
}

type LockCleanup struct{}

type InvitationCleanup struct{}

type OldDataCleanup struct{}

type ShopwareChangelogSync struct{}

// ComposerAdvisorySync refreshes Packagist security advisories for official
// shopware/* Composer packages into the local composer_advisory catalog.
type ComposerAdvisorySync struct{}

// SecurityPluginSync refreshes the map of which SwagPlatformSecurity release
// backports which advisory, derived from the plugin's store changelog.
type SecurityPluginSync struct{}

// StoreExtensionSync refreshes the shared store extension catalog for a set of
// extension names. ShopwareVersion is the version of the environment that
// requested the sync; it is included in the compatibility probing.
type StoreExtensionSync struct {
	Names           []string `json:"names"`
	ShopwareVersion string   `json:"shopware_version"`
}
