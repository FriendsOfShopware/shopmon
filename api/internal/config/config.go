package config

import (
	"fmt"
	"net/url"
	"os"

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

	ShopwareAPIURL      string
	ShopwareVersionsURL string

	OtelEnabled       bool
	OtelTraceEndpoint string
	OtelLogEndpoint   string
	OtelServiceName   string

	WebAuthnRPID      string
	WebAuthnRPName    string
	WebAuthnRPOrigins []string

	ListenAddr string
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

		SitespeedEndpoint: getEnv("APP_SITESPEED_ENDPOINT", "http://localhost:3001"),
		SitespeedPrefix:   getEnv("APP_SITESPEED_PREFIX", "local-"),
		SitespeedAPIKey:   getEnv("APP_SITESPEED_API_KEY", "secret"),

		S3Endpoint:  getEnv("APP_S3_ENDPOINT", ""),
		S3AccessKey: getEnv("APP_S3_ACCESS_KEY_ID", ""),
		S3SecretKey: getEnv("APP_S3_SECRET_ACCESS_KEY", ""),
		S3Bucket:    getEnv("APP_S3_BUCKET", "shopmon"),
		S3Region:    getEnv("APP_S3_REGION", "auto"),

		GithubClientID:     getEnv("APP_OAUTH_GITHUB_CLIENT_ID", ""),
		GithubClientSecret: getEnv("APP_OAUTH_GITHUB_CLIENT_SECRET", ""),

		PackagesAPIURL:   getEnv("PACKAGES_API_URL", ""),
		PackagesAPIToken: getEnv("PACKAGES_API_TOKEN", ""),

		ShopwareAPIURL:      getEnv("SHOPWARE_API_URL", "https://api.shopware.com"),
		ShopwareVersionsURL: getEnv("SHOPWARE_VERSIONS_URL", "https://raw.githubusercontent.com/FriendsOfShopware/shopware-static-data/main/data/all-supported-php-versions-by-shopware-version.json"),

		OtelEnabled:       getEnv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "")) != "",
		OtelTraceEndpoint: getEnv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "")),
		OtelLogEndpoint:   getEnv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "")),
		OtelServiceName:   getEnv("OTEL_SERVICE_NAME", "shopmon"),

		ListenAddr: getEnv("LISTEN_ADDR", ":8080"),
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
			panic(fmt.Sprintf("APP_SECRET must be exactly 16, 24, or 32 bytes for AES encryption (got %d)", keyLen))
		}
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
