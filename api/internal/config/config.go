package config

import (
	"log/slog"
	"net/url"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppSecret   string
	DatabaseURL string
	RedisURL    string
	FrontendURL string

	SMTPHost    string
	SMTPPort    string
	SMTPSecure  bool
	SMTPUser    string
	SMTPPass    string
	MailFrom    string
	SMTPReplyTo string

	SitespeedEndpoint string
	SitespeedPrefix   string
	SitespeedAPIKey   string

	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string
	S3Bucket    string
	S3Region    string

	GithubClientID     string
	GithubClientSecret string

	PackagesAPIURL   string
	PackagesAPIToken string

	DisableRegistration bool

	// DeploymentScrapeDelay is how long to wait after a CLI deployment is
	// recorded before re-scraping the environment, giving post-deploy tasks
	// (theme compile, indexing, cache warming) time to settle.
	DeploymentScrapeDelay time.Duration

	ShopwareAPIURL      string
	ShopwareVersionsURL string

	OtelEnabled        bool
	OtelTraceEndpoint  string
	OtelLogEndpoint    string
	OtelServiceName    string
	OtelDeploymentEnv  string
	OtelServiceVersion string

	WebAuthnRPID      string
	WebAuthnRPName    string
	WebAuthnRPOrigins []string

	ListenAddr     string
	TrustedProxies []string
}

func loadDotEnv() {
	_ = godotenv.Load()
}

func Load() *Config {
	loadDotEnv()

	cfg := &Config{
		AppSecret:   getEnv("APP_SECRET", ""),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://shopmon:shopmon@localhost:5432/shopmon"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),

		SMTPHost:    getEnv("SMTP_HOST", "localhost"),
		SMTPPort:    getEnv("SMTP_PORT", "1025"),
		SMTPSecure:  getEnv("SMTP_SECURE", "false") == "true",
		SMTPUser:    getEnv("SMTP_USER", ""),
		SMTPPass:    getEnv("SMTP_PASS", ""),
		MailFrom:    getEnv("MAIL_FROM", "noreply@shopmon.io"),
		SMTPReplyTo: getEnv("SMTP_REPLY_TO", ""),

		SitespeedEndpoint: getEnv("APP_SITESPEED_ENDPOINT", ""),
		SitespeedPrefix:   getEnv("APP_SITESPEED_PREFIX", "local-"),
		SitespeedAPIKey:   getEnv("APP_SITESPEED_API_KEY", ""),

		S3Endpoint:  getEnv("APP_S3_ENDPOINT", ""),
		S3AccessKey: getEnv("APP_S3_ACCESS_KEY_ID", ""),
		S3SecretKey: getEnv("APP_S3_SECRET_ACCESS_KEY", ""),
		S3Bucket:    getEnv("APP_S3_BUCKET", "shopmon"),
		S3Region:    getEnv("APP_S3_REGION", "auto"),

		GithubClientID:     getEnv("APP_OAUTH_GITHUB_CLIENT_ID", ""),
		GithubClientSecret: getEnv("APP_OAUTH_GITHUB_CLIENT_SECRET", ""),

		PackagesAPIURL:   getEnv("PACKAGES_API_URL", ""),
		PackagesAPIToken: getEnv("PACKAGES_API_TOKEN", ""),

		DisableRegistration: getEnv("DISABLE_REGISTRATION", "false") == "true",

		ShopwareAPIURL:      getEnv("SHOPWARE_API_URL", "https://api.shopware.com"),
		ShopwareVersionsURL: getEnv("SHOPWARE_VERSIONS_URL", "https://raw.githubusercontent.com/FriendsOfShopware/shopware-static-data/main/data/all-supported-php-versions-by-shopware-version.json"),

		OtelEnabled:        getEnv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "")) != "",
		OtelTraceEndpoint:  getEnv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "")),
		OtelLogEndpoint:    getEnv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "")),
		OtelServiceName:    getEnv("OTEL_SERVICE_NAME", "shopmon"),
		OtelDeploymentEnv:  getEnv("OTEL_DEPLOYMENT_ENVIRONMENT", getEnv("DD_ENV", "")),
		OtelServiceVersion: getEnv("OTEL_SERVICE_VERSION", getEnv("DD_VERSION", buildVersion())),

		ListenAddr:     getEnv("LISTEN_ADDR", ":8080"),
		TrustedProxies: parseCommaList(getEnv("TRUSTED_PROXIES", "")),
	}

	// Parse the post-deployment scrape delay, falling back to 5m on an
	// empty or invalid value.
	cfg.DeploymentScrapeDelay = 5 * time.Minute
	if raw := getEnv("DEPLOYMENT_SCRAPE_DELAY", ""); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil && d >= 0 {
			cfg.DeploymentScrapeDelay = d
		} else {
			slog.Warn("invalid DEPLOYMENT_SCRAPE_DELAY, using default", "value", raw, "default", cfg.DeploymentScrapeDelay)
		}
	}

	// Derive WebAuthn config from FrontendURL
	if parsed, err := url.Parse(cfg.FrontendURL); err == nil {
		cfg.WebAuthnRPID = parsed.Hostname()
		cfg.WebAuthnRPName = "Shopmon"
		cfg.WebAuthnRPOrigins = []string{cfg.FrontendURL}
	}

	// Validate APP_SECRET length for AES encryption
	if cfg.AppSecret != "" {
		keyLen := len(cfg.AppSecret)
		if keyLen != 16 && keyLen != 24 && keyLen != 32 {
			slog.Error("invalid APP_SECRET length: must be exactly 16, 24, or 32 bytes for AES encryption", "length", keyLen)
			os.Exit(1)
		}
	}

	return cfg
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

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseCommaList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
