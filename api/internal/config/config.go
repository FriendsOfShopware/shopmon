// Package config loads the application configuration from the environment.
//
// Every setting is declared as a struct field with an `env` tag: the tag names
// the variable, `envDefault` carries the fallback, and `expand` lets a default
// reference another variable (used for the OTLP/Datadog aliases). Anything that
// cannot be expressed as a tag — values derived from other settings, or checks
// that span fields — lives in normalize and validate below.
package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"runtime/debug"
	"strings"
	"time"

	env "github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	AppSecret   string `env:"APP_SECRET"`
	DatabaseURL string `env:"DATABASE_URL" envDefault:"postgres://shopmon:shopmon@localhost:5432/shopmon"`
	RedisURL    string `env:"REDIS_URL" envDefault:"redis://localhost:6379"`
	FrontendURL string `env:"FRONTEND_URL" envDefault:"http://localhost:3000"`

	// QueueTransport selects the background job backend: "postgres" (default,
	// jobs live in the app database) or "amqp" (RabbitMQ/LavinMQ broker).
	QueueTransport string `env:"QUEUE_TRANSPORT" envDefault:"postgres"`
	// QueueAMQP is only read when QueueTransport is "amqp".
	QueueAMQP QueueAMQPConfig `envPrefix:"QUEUE_AMQP_"`

	// MailDSN is the go-mailer SMTP DSN. When unset it is assembled from the
	// legacy SMTP* fields below so existing deployments keep working.
	MailDSN  string `env:"MAIL_DSN"`
	MailFrom string `env:"MAIL_FROM" envDefault:"noreply@shopmon.io"`

	SMTPHost    string `env:"SMTP_HOST" envDefault:"localhost"`
	SMTPPort    string `env:"SMTP_PORT" envDefault:"1025"`
	SMTPUser    string `env:"SMTP_USER"`
	SMTPPass    string `env:"SMTP_PASS"`
	SMTPSecure  bool   `env:"SMTP_SECURE"`
	SMTPReplyTo string `env:"SMTP_REPLY_TO"`

	SitespeedEndpoint string `env:"APP_SITESPEED_ENDPOINT"`
	SitespeedPrefix   string `env:"APP_SITESPEED_PREFIX" envDefault:"local-"`
	SitespeedAPIKey   string `env:"APP_SITESPEED_API_KEY"`

	S3Endpoint  string `env:"APP_S3_ENDPOINT"`
	S3AccessKey string `env:"APP_S3_ACCESS_KEY_ID"`
	S3SecretKey string `env:"APP_S3_SECRET_ACCESS_KEY"`
	S3Bucket    string `env:"APP_S3_BUCKET" envDefault:"shopmon"`
	S3Region    string `env:"APP_S3_REGION" envDefault:"auto"`

	GithubClientID     string `env:"APP_OAUTH_GITHUB_CLIENT_ID"`
	GithubClientSecret string `env:"APP_OAUTH_GITHUB_CLIENT_SECRET"`
	// GithubToken is an optional personal access token used for GitHub API
	// calls that enrich security advisories without a CVE (higher rate limits).
	GithubToken string `env:"GITHUB_TOKEN,expand" envDefault:"${APP_GITHUB_TOKEN}"`
	// NVDAPIKey is an optional NIST NVD API key for higher rate limits when
	// enriching CVE descriptions from services.nvd.nist.gov.
	NVDAPIKey string `env:"NVD_API_KEY,expand" envDefault:"${APP_NVD_API_KEY}"`

	PackagesAPIURL   string `env:"PACKAGES_API_URL"`
	PackagesAPIToken string `env:"PACKAGES_API_TOKEN"`

	DisableRegistration bool `env:"DISABLE_REGISTRATION"`

	// DeploymentScrapeDelay is how long to wait after a CLI deployment is
	// recorded before re-scraping the environment, giving post-deploy tasks
	// (theme compile, indexing, cache warming) time to settle.
	DeploymentScrapeDelay time.Duration `env:"DEPLOYMENT_SCRAPE_DELAY" envDefault:"5m"`

	ShopwareAPIURL string `env:"SHOPWARE_API_URL" envDefault:"https://api.shopware.com"`
	// ShopwareChangelogURL is the base URL of the Shopware release changelog API
	// (index.json + per-version JSON) crawled hourly by the worker.
	ShopwareChangelogURL string `env:"SHOPWARE_CHANGELOG_URL" envDefault:"https://releases.shopware.com/changelog"`

	// The OTLP signal endpoints fall back to the generic
	// OTEL_EXPORTER_OTLP_ENDPOINT, and service env/version to the Datadog
	// unified service tagging variables.
	OtelTraceEndpoint string `env:"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT,expand" envDefault:"${OTEL_EXPORTER_OTLP_ENDPOINT}"`
	OtelLogEndpoint   string `env:"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT,expand" envDefault:"${OTEL_EXPORTER_OTLP_ENDPOINT}"`
	// OTEL_EXPORTER_OTLP_METRICS_ENDPOINT selects the OTLP HTTP metrics
	// exporter. When empty (and the generic OTEL_EXPORTER_OTLP_ENDPOINT is
	// also empty), metrics stay disabled and MeterProvider remains a no-op.
	OtelMetricEndpoint string `env:"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT,expand" envDefault:"${OTEL_EXPORTER_OTLP_ENDPOINT}"`
	// OtelServiceName is the shared base APM service name. The HTTP server
	// reports as ${OTEL_SERVICE_NAME}-api and the worker as
	// ${OTEL_SERVICE_NAME}-worker (see api/server.go and api/worker.go).
	OtelServiceName    string  `env:"OTEL_SERVICE_NAME" envDefault:"shopmon"`
	OtelDeploymentEnv  string  `env:"OTEL_DEPLOYMENT_ENVIRONMENT,expand" envDefault:"${DD_ENV}"`
	OtelServiceVersion string  `env:"OTEL_SERVICE_VERSION,expand" envDefault:"${DD_VERSION}"`
	OtelSamplerRatio   float64 `env:"OTEL_TRACES_SAMPLER_RATIO" envDefault:"1"`
	// OtelEnabled is derived from OtelTraceEndpoint.
	OtelEnabled bool

	// The WebAuthn relying party is derived from FrontendURL.
	WebAuthnRPID      string
	WebAuthnRPName    string
	WebAuthnRPOrigins []string

	ListenAddr     string   `env:"LISTEN_ADDR" envDefault:":8080"`
	TrustedProxies []string `env:"TRUSTED_PROXIES"`

	// AuthRateLimitMax is the number of auth requests allowed per IP per minute.
	// Raise it (e.g. for E2E tests) via AUTH_RATE_LIMIT_MAX.
	AuthRateLimitMax int `env:"AUTH_RATE_LIMIT_MAX" envDefault:"20"`
}

// QueueAMQPConfig holds the broker settings for the AMQP job transport.
type QueueAMQPConfig struct {
	DSN      string `env:"DSN" envDefault:"amqp://guest:guest@localhost:5672/"`
	Exchange string `env:"EXCHANGE" envDefault:"shopmon"`
	Queue    string `env:"QUEUE" envDefault:"shopmon"`
	// PrefetchCount bounds unacknowledged deliveries per consumer. Keep it at or
	// above the worker concurrency so workers never idle waiting for messages.
	PrefetchCount int `env:"PREFETCH" envDefault:"10"`
	// DelayedExchange declares the exchange as x-delayed-message so delayed jobs
	// (post-deployment scrapes, sitespeed reruns) are held by the broker instead
	// of being delivered immediately. Needs LavinMQ (native) or the RabbitMQ
	// delayed-message plugin.
	DelayedExchange bool `env:"DELAYED_EXCHANGE" envDefault:"true"`
}

// Load reads the configuration from .env (when present) and the environment.
// An unparseable or out-of-range value is an error rather than a silent
// fallback, so a typo cannot start the process with a setting the operator
// never asked for.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("read environment: %w", err)
	}

	cfg.normalize()

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// normalize fills in the settings that are derived from other settings and
// cleans up the ones the struct tags cannot express on their own.
func (c *Config) normalize() {
	c.QueueTransport = strings.ToLower(strings.TrimSpace(c.QueueTransport))
	c.TrustedProxies = trimList(c.TrustedProxies)

	if c.MailDSN == "" {
		c.MailDSN = c.smtpDSN()
	}

	if c.OtelServiceVersion == "" {
		c.OtelServiceVersion = buildVersion()
	}
	c.OtelEnabled = c.OtelTraceEndpoint != ""
	c.OtelSamplerRatio = min(max(c.OtelSamplerRatio, 0), 1)

	if parsed, err := url.Parse(c.FrontendURL); err == nil {
		c.WebAuthnRPID = parsed.Hostname()
		c.WebAuthnRPName = "Shopmon"
		c.WebAuthnRPOrigins = []string{c.FrontendURL}
	}
}

func (c *Config) validate() error {
	var errs []error

	// AES-128/192/256 need a key of exactly 16, 24 or 32 bytes. An empty secret
	// stays allowed: commands that never touch encrypted data (migrate) run
	// without one.
	if n := len(c.AppSecret); n != 0 && n != 16 && n != 24 && n != 32 {
		errs = append(errs, fmt.Errorf("APP_SECRET must be exactly 16, 24 or 32 bytes for AES encryption, got %d", n))
	}
	if c.DeploymentScrapeDelay < 0 {
		errs = append(errs, fmt.Errorf("DEPLOYMENT_SCRAPE_DELAY must not be negative, got %s", c.DeploymentScrapeDelay))
	}
	if c.AuthRateLimitMax <= 0 {
		errs = append(errs, fmt.Errorf("AUTH_RATE_LIMIT_MAX must be greater than 0, got %d", c.AuthRateLimitMax))
	}
	if c.QueueAMQP.PrefetchCount <= 0 {
		errs = append(errs, fmt.Errorf("QUEUE_AMQP_PREFETCH must be greater than 0, got %d", c.QueueAMQP.PrefetchCount))
	}

	return errors.Join(errs...)
}

// smtpDSN assembles a go-mailer SMTP DSN from the legacy SMTP* settings.
// SMTPSecure selects the smtps scheme (implicit TLS).
func (c *Config) smtpDSN() string {
	scheme := "smtp"
	if c.SMTPSecure {
		scheme = "smtps"
	}

	u := url.URL{
		Scheme: scheme,
		Host:   net.JoinHostPort(c.SMTPHost, c.SMTPPort),
	}
	if c.SMTPUser != "" {
		u.User = url.UserPassword(c.SMTPUser, c.SMTPPass)
	}
	return u.String()
}

// buildVersion returns the VCS revision the binary was built from, embedded by
// the Go toolchain at build time (no -ldflags needed). It returns the short
// commit hash, suffixed with "-dirty" when the working tree had uncommitted
// changes, or an empty string when build info is unavailable (e.g. go run).
func buildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}

	var revision string
	var modified bool
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			modified = s.Value == "true"
		}
	}
	if revision == "" {
		return ""
	}
	if len(revision) > 12 {
		revision = revision[:12]
	}
	if modified {
		revision += "-dirty"
	}
	return revision
}

// trimList trims each entry of a comma-separated list and drops empty ones, so
// "a, b," yields ["a", "b"].
func trimList(values []string) []string {
	result := make([]string, 0, len(values))
	for _, v := range values {
		if v = strings.TrimSpace(v); v != "" {
			result = append(result, v)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
