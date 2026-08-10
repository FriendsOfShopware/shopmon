package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeploymentScrapeDelay(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		want    time.Duration
		wantErr bool
	}{
		{name: "default when unset", env: "", want: 5 * time.Minute},
		{name: "custom duration", env: "2m", want: 2 * time.Minute},
		{name: "zero disables delay", env: "0s", want: 0},
		{name: "invalid is rejected", env: "not-a-duration", wantErr: true},
		{name: "negative is rejected", env: "-1m", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DEPLOYMENT_SCRAPE_DELAY", tt.env)

			cfg, err := Load()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, cfg.DeploymentScrapeDelay)
		})
	}
}

func TestQueueTransportDefaultsToPostgres(t *testing.T) {
	for _, key := range []string{
		"QUEUE_TRANSPORT",
		"QUEUE_AMQP_DSN",
		"QUEUE_AMQP_EXCHANGE",
		"QUEUE_AMQP_QUEUE",
		"QUEUE_AMQP_PREFETCH",
		"QUEUE_AMQP_DELAYED_EXCHANGE",
	} {
		t.Setenv(key, "")
	}

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "postgres", cfg.QueueTransport)
	assert.Equal(t, "amqp://guest:guest@localhost:5672/", cfg.QueueAMQP.DSN)
	assert.Equal(t, "shopmon", cfg.QueueAMQP.Exchange)
	assert.Equal(t, "shopmon", cfg.QueueAMQP.Queue)
	assert.Equal(t, 10, cfg.QueueAMQP.PrefetchCount)
	assert.True(t, cfg.QueueAMQP.DelayedExchange)
}

func TestQueueTransportIsNormalized(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want string
	}{
		{name: "lowercased", env: "AMQP", want: "amqp"},
		{name: "trimmed", env: "  amqp  ", want: "amqp"},
		{name: "unknown value kept for the bus to reject", env: "kafka", want: "kafka"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("QUEUE_TRANSPORT", tt.env)

			cfg, err := Load()
			require.NoError(t, err)
			assert.Equal(t, tt.want, cfg.QueueTransport)
		})
	}
}

func TestQueueAMQPPrefetch(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		want    int
		wantErr bool
	}{
		{name: "default when unset", env: "", want: 10},
		{name: "custom count", env: "50", want: 50},
		{name: "invalid is rejected", env: "many", wantErr: true},
		{name: "zero is rejected", env: "0", wantErr: true},
		{name: "negative is rejected", env: "-1", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("QUEUE_AMQP_PREFETCH", tt.env)

			cfg, err := Load()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, cfg.QueueAMQP.PrefetchCount)
		})
	}
}

func TestQueueAMQPDelayedExchange(t *testing.T) {
	// Delayed delivery is load-bearing (post-deployment scrapes, sitespeed
	// reruns), so it may only be turned off by an explicit, parseable false.
	// Anything else fails the load instead of silently picking a side.
	tests := []struct {
		name    string
		env     string
		want    bool
		wantErr bool
	}{
		{name: "default when unset", env: "", want: true},
		{name: "explicit false", env: "false", want: false},
		{name: "uppercase FALSE", env: "FALSE", want: false},
		{name: "zero", env: "0", want: false},
		{name: "uppercase TRUE", env: "TRUE", want: true},
		{name: "titlecase True", env: "True", want: true},
		{name: "one", env: "1", want: true},
		{name: "unparseable is rejected", env: "yes", wantErr: true},
		{name: "typo is rejected", env: "flase", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("QUEUE_AMQP_DELAYED_EXCHANGE", tt.env)

			cfg, err := Load()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, cfg.QueueAMQP.DelayedExchange)
		})
	}
}

func TestAuthRateLimitMax(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		want    int
		wantErr bool
	}{
		{name: "default when unset", env: "", want: 20},
		{name: "custom budget", env: "500", want: 500},
		{name: "invalid is rejected", env: "lots", wantErr: true},
		{name: "zero is rejected", env: "0", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("AUTH_RATE_LIMIT_MAX", tt.env)

			cfg, err := Load()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, cfg.AuthRateLimitMax)
		})
	}
}

func TestAppSecretLength(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		wantErr bool
	}{
		{name: "empty is allowed", env: ""},
		{name: "16 bytes", env: "0123456789abcdef"},
		{name: "24 bytes", env: "0123456789abcdef01234567"},
		{name: "32 bytes", env: "0123456789abcdef0123456789abcdef"},
		{name: "too short is rejected", env: "short", wantErr: true},
		{name: "odd length is rejected", env: "0123456789abcdef0", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("APP_SECRET", tt.env)

			_, err := Load()
			if tt.wantErr {
				assert.ErrorContains(t, err, "APP_SECRET")
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestMailDSN(t *testing.T) {
	t.Run("MAIL_DSN wins", func(t *testing.T) {
		t.Setenv("MAIL_DSN", "smtp://custom:2525")
		t.Setenv("SMTP_HOST", "ignored")

		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, "smtp://custom:2525", cfg.MailDSN)
	})

	tests := []struct {
		name   string
		host   string
		port   string
		user   string
		pass   string
		secure string
		want   string
	}{
		{name: "defaults to local mailpit", want: "smtp://localhost:1025"},
		{name: "host and port", host: "mail.example.com", port: "587", want: "smtp://mail.example.com:587"},
		{name: "credentials", host: "mail.example.com", port: "587", user: "u", pass: "p", want: "smtp://u:p@mail.example.com:587"},
		{name: "secure selects smtps", host: "mail.example.com", port: "465", secure: "true", want: "smtps://mail.example.com:465"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MAIL_DSN", "")
			t.Setenv("SMTP_HOST", tt.host)
			t.Setenv("SMTP_PORT", tt.port)
			t.Setenv("SMTP_USER", tt.user)
			t.Setenv("SMTP_PASS", tt.pass)
			t.Setenv("SMTP_SECURE", tt.secure)

			cfg, err := Load()
			require.NoError(t, err)
			assert.Equal(t, tt.want, cfg.MailDSN)
		})
	}
}

func TestOtelEndpointsFallBackToGenericEndpoint(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://collector:4318")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", "")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "http://collector:4318", cfg.OtelTraceEndpoint)
	assert.Equal(t, "http://collector:4318", cfg.OtelLogEndpoint)
	assert.True(t, cfg.OtelEnabled)

	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "http://traces:4318")

	cfg, err = Load()
	require.NoError(t, err)

	assert.Equal(t, "http://traces:4318", cfg.OtelTraceEndpoint)
	assert.Equal(t, "http://collector:4318", cfg.OtelLogEndpoint)
}

func TestOtelDisabledWithoutEndpoint(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Empty(t, cfg.OtelTraceEndpoint)
	assert.False(t, cfg.OtelEnabled)
}

func TestOtelSamplerRatioIsClamped(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		want    float64
		wantErr bool
	}{
		{name: "default samples everything", env: "", want: 1},
		{name: "half", env: "0.5", want: 0.5},
		{name: "above one clamps", env: "2.5", want: 1},
		{name: "below zero clamps", env: "-1", want: 0},
		{name: "invalid is rejected", env: "not-a-number", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("OTEL_TRACES_SAMPLER_RATIO", tt.env)

			cfg, err := Load()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, cfg.OtelSamplerRatio)
		})
	}
}

func TestTrustedProxiesAreTrimmed(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want []string
	}{
		{name: "unset", env: "", want: nil},
		{name: "single", env: "10.0.0.1", want: []string{"10.0.0.1"}},
		{name: "trimmed and empties dropped", env: " 10.0.0.1 , ,10.0.0.2,", want: []string{"10.0.0.1", "10.0.0.2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TRUSTED_PROXIES", tt.env)

			cfg, err := Load()
			require.NoError(t, err)
			assert.Equal(t, tt.want, cfg.TrustedProxies)
		})
	}
}

func TestWebAuthnDerivedFromFrontendURL(t *testing.T) {
	t.Setenv("FRONTEND_URL", "https://shopmon.example.com")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "shopmon.example.com", cfg.WebAuthnRPID)
	assert.Equal(t, "Shopmon", cfg.WebAuthnRPName)
	assert.Equal(t, []string{"https://shopmon.example.com"}, cfg.WebAuthnRPOrigins)
}
