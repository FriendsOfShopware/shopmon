package jobs

// TransportName is the stable queue transport used by every Shopmon job.
const TransportName = "async"

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

// UptimeDailyRollup aggregates the previous day's uptime probes into daily
// availability rollups and enforces uptime data retention.
type UptimeDailyRollup struct{}

// UptimeResume re-arms a paused uptime monitor after a deployment's settle
// window, so probes do not fire while post-deploy tasks are still running.
type UptimeResume struct {
	EnvironmentID int32 `json:"environment_id"`
}

// StoreExtensionSync refreshes the shared store extension catalog for a set of
// extension names. ShopwareVersion is the version of the environment that
// requested the sync; it is included in the compatibility probing.
type StoreExtensionSync struct {
	Names           []string `json:"names"`
	ShopwareVersion string   `json:"shopware_version"`
}
