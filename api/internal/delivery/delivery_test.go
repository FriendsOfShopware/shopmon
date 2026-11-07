package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestValidateProvidersAndRedactSecrets(t *testing.T) {
	for _, tc := range []struct{ dsn, provider, summary string }{
		{"slack://xoxb-secret@default?channel=C123", "slack", "C123"},
		{"slack://default?webhook_url=https%3A%2F%2Fhooks.slack.com%2Fservices%2FT%2FB%2Fsecret", "slack", "Incoming webhook"},
		{"telegram://123:secret@default?channel=456", "telegram", "456"},
		{"discord://secret@default?webhook_id=123", "discord", "123"},
		{"gotify://secret@notify.example.com", "gotify", "notify.example.com"},
		{"microsoftteams://id@default?token=secret", "microsoftteams", "Incoming webhook"},
	} {
		t.Run(tc.provider+tc.summary, func(t *testing.T) {
			details, err := Validate(tc.dsn)
			require.NoError(t, err)
			assert.Equal(t, Details{Provider: tc.provider, Summary: tc.summary}, details)
			assert.NotContains(t, details.Summary, "secret")
		})
	}
	for _, dsn := range []string{"slack://secret@default?channel=C123", "discord://secret@default", "ftp://secret@example.com", "telegram://123:secret@default", "gotify://secret@default", "microsoftteams://default?webhook_url=secret", "slack://default?webhook_url=https://evil.example/secret", "slack://default?webhook_url=http://hooks.slack.com/services/secret"} {
		_, err := Validate(dsn)
		require.ErrorIs(t, err, ErrInvalidDSN)
		assert.NotContains(t, err.Error(), "secret")
	}
}

func TestProviderDeliveryPayloads(t *testing.T) {
	for _, tc := range []struct{ dsn, host, response, field string }{
		{"slack://default?webhook_url=https%3A%2F%2Fhooks.slack.com%2Fservices%2FT%2FB%2Fsecret", "hooks.slack.com", "ok", "text"},
		{"slack://xoxb-secret@default?channel=C123", "slack.com", `{"ok":true,"ts":"123"}`, "text"},
		{"telegram://123:secret@default?channel=456", "api.telegram.org", `{"ok":true,"result":{"message_id":123}}`, "text"},
		{"discord://secret@default?webhook_id=123", "discord.com", `{"id":"123"}`, "content"},
		{"gotify://secret@notify.example.com", "notify.example.com", `{"id":123}`, "message"},
		{"microsoftteams://id@default?token=secret", "outlook.office.com", "1", "text"},
	} {
		t.Run(tc.dsn[:strings.Index(tc.dsn, ":")]+tc.host, func(t *testing.T) {
			called := false
			client := newClient(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				called = true
				assert.Equal(t, tc.host, r.URL.Host)
				assert.Equal(t, http.MethodPost, r.Method)
				var body map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				text, ok := body[tc.field].(string)
				require.True(t, ok)
				assert.Contains(t, text, "Shop")
				if tc.host == "hooks.slack.com" {
					assert.Contains(t, text, "&lt;!channel&gt;")
				}
				if tc.host == "api.telegram.org" {
					assert.Equal(t, "MarkdownV2", body["parse_mode"])
					assert.Contains(t, text, `\!channel\>`)
					assert.Contains(t, r.URL.Path, "123:secret")
				}
				if tc.host == "discord.com" {
					assert.NotContains(t, text, "@everyone")
				}
				status := http.StatusOK
				if tc.host == "discord.com" {
					status = http.StatusNoContent
				}
				return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(tc.response)), Header: make(http.Header)}, nil
			})})
			require.NoError(t, client.Post(t.Context(), tc.dsn, "⚠️ Shop <!channel> @everyone\nhttps://shopmon.example/app"))
			assert.True(t, called)
		})
	}
}

func TestDeliveryErrorsNeverExposeCredentials(t *testing.T) {
	client := newClient(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, errors.New("secret-token in provider URL") })})
	err := client.Post(t.Context(), "telegram://123:secret-token@default?channel=456", "Test")
	require.ErrorIs(t, err, ErrDelivery)
	assert.NotContains(t, err.Error(), "secret-token")
}

func TestPrivateEndpointsBlocked(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	client := NewClient(time.Second)
	err := client.Post(ctx, "gotify://secret@127.0.0.1", "Test")
	require.ErrorIs(t, err, ErrDelivery)
}
