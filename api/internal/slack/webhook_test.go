package slack

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateWebhookURL(t *testing.T) {
	valid := []string{
		"https://hooks.slack.com/services/T000/B000/XXXX",
		"  https://hooks.slack.com/triggers/T000/123/abc  ",
	}
	for _, u := range valid {
		assert.NoError(t, ValidateWebhookURL(u), u)
	}

	invalid := []string{
		"",
		"http://hooks.slack.com/services/T000/B000/XXXX",
		"https://hooks.slack.com/",
		"https://hooks.slack.com.evil.test/services/x",
		"https://user@hooks.slack.com/services/x",
		"https://169.254.169.254/latest/meta-data",
		"not a url",
	}
	for _, u := range invalid {
		assert.ErrorIs(t, ValidateWebhookURL(u), ErrInvalidWebhookURL, u)
	}
}

func testClient() *Client {
	c := NewClient(2 * time.Second)
	c.validate = func(string) error { return nil }
	return c
}

func TestPostSendsTextPayload(t *testing.T) {
	var got map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	require.NoError(t, testClient().Post(context.Background(), srv.URL+"/services/x", "*hello*"))
	assert.Equal(t, map[string]string{"text": "*hello*"}, got)
}

func TestPostReportsSlackError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("no_service"))
	}))
	defer srv.Close()

	err := testClient().Post(context.Background(), srv.URL+"/services/x", "hi")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "no_service")
}

func TestPostDoesNotLeakURLOnTransportError(t *testing.T) {
	err := testClient().Post(context.Background(), "http://127.0.0.1:1/services/T000/SECRET", "hi")
	require.Error(t, err)
	assert.False(t, strings.Contains(err.Error(), "SECRET"), err.Error())
}

func TestPostRejectsForeignHost(t *testing.T) {
	err := NewClient(time.Second).Post(context.Background(), "https://example.com/hook", "hi")
	assert.ErrorIs(t, err, ErrInvalidWebhookURL)
}

func TestEscape(t *testing.T) {
	assert.Equal(t, "a &amp; &lt;b&gt;", Escape("a & <b>"))
}
