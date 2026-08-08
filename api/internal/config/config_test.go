package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeploymentScrapeDelay(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want time.Duration
	}{
		{name: "default when unset", env: "", want: 5 * time.Minute},
		{name: "custom duration", env: "2m", want: 2 * time.Minute},
		{name: "zero disables delay", env: "0s", want: 0},
		{name: "invalid falls back to default", env: "not-a-duration", want: 5 * time.Minute},
		{name: "negative falls back to default", env: "-1m", want: 5 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.env == "" {
				t.Setenv("DEPLOYMENT_SCRAPE_DELAY", "")
			} else {
				t.Setenv("DEPLOYMENT_SCRAPE_DELAY", tt.env)
			}

			cfg := Load()
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

	cfg := Load()

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

			assert.Equal(t, tt.want, Load().QueueTransport)
		})
	}
}

func TestQueueAMQPPrefetch(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want int
	}{
		{name: "default when unset", env: "", want: 10},
		{name: "custom count", env: "50", want: 50},
		{name: "invalid falls back to default", env: "many", want: 10},
		{name: "zero falls back to default", env: "0", want: 10},
		{name: "negative falls back to default", env: "-1", want: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("QUEUE_AMQP_PREFETCH", tt.env)

			assert.Equal(t, tt.want, Load().QueueAMQP.PrefetchCount)
		})
	}
}

func TestQueueAMQPDelayedExchangeOptOut(t *testing.T) {
	t.Setenv("QUEUE_AMQP_DELAYED_EXCHANGE", "false")

	assert.False(t, Load().QueueAMQP.DelayedExchange)
}
