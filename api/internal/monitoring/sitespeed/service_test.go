package sitespeed

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	goqueue "github.com/shyim/go-queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type repositoryStub struct {
	environments []queries.GetEnvironmentsWithSitespeedEnabledRow
	err          error
	inserted     int
}

func (r *repositoryStub) GetEnvironmentsWithSitespeedEnabled(context.Context) ([]queries.GetEnvironmentsWithSitespeedEnabledRow, error) {
	return r.environments, r.err
}
func (r *repositoryStub) InsertEnvironmentSitespeed(context.Context, queries.InsertEnvironmentSitespeedParams) error {
	r.inserted++
	return nil
}
func (r *repositoryStub) UpdateEnvironmentImage(context.Context, queries.UpdateEnvironmentImageParams) error {
	return nil
}

func TestNormalizeURLs(t *testing.T) {
	for _, value := range []string{"", "  ", "ftp://example.com", "file:///etc/passwd", "https:///path", "https://user:secret@example.com", "://bad"} {
		t.Run(value, func(t *testing.T) {
			_, err := NormalizeURLs([]string{value})
			require.ErrorIs(t, err, ErrInvalidURL)
			assert.Contains(t, err.Error(), "entry 1")
			assert.NotContains(t, err.Error(), "user:secret")
		})
	}
	values, err := NormalizeURLs([]string{" https://example.com/path ", "http://localhost:8000"})
	require.NoError(t, err)
	assert.Equal(t, []string{"https://example.com/path", "http://localhost:8000"}, values)
	_, err = NormalizeURLs(nil)
	require.NoError(t, err)
}

func TestObsoleteJobsAreSkipped(t *testing.T) {
	for _, environments := range [][]queries.GetEnvironmentsWithSitespeedEnabledRow{nil, {{ID: 2}}} {
		repo := &repositoryStub{environments: environments}
		require.NoError(t, NewService(repo, &config.Config{}).Scrape(context.Background(), 1))
		assert.Zero(t, repo.inserted)
	}
	repo := &repositoryStub{err: errors.New("database unavailable")}
	require.Error(t, NewService(repo, &config.Config{}).Scrape(context.Background(), 1))
}

func TestSitespeedFailureClassification(t *testing.T) {
	for _, status := range []int{200, 301, 400, 401, 403, 404, 408, 422, 429, 500, 502, 504} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			calls := 0
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				assert.Equal(t, http.MethodPost, r.Method)
				w.WriteHeader(status)
				if status == 200 {
					_, _ = w.Write([]byte(`{"ttfb":123}`))
				} else {
					_, _ = w.Write([]byte("secret upstream body"))
				}
			}))
			defer upstream.Close()
			repo := &repositoryStub{environments: []queries.GetEnvironmentsWithSitespeedEnabledRow{{ID: 1, Url: "https://example.com"}}}
			svc := NewService(repo, &config.Config{SitespeedEndpoint: upstream.URL, SitespeedAPIKey: "secret"})
			err := svc.Scrape(context.Background(), 1)
			if status == 200 {
				require.NoError(t, err)
				assert.Equal(t, 1, repo.inserted)
			} else {
				require.Error(t, err)
				assert.NotContains(t, err.Error(), "secret")
				assert.Equal(t, status < 500 && status != 408 && status != 429, goqueue.IsUnrecoverable(err))
				assert.Zero(t, repo.inserted)
			}
			assert.Equal(t, 1, calls)
		})
	}
}

func TestInvalidStoredConfigurationNeverCallsUpstream(t *testing.T) {
	for _, raw := range []json.RawMessage{json.RawMessage(`[""]`), json.RawMessage(`["ftp://example.com"]`), json.RawMessage(`broken`)} {
		repo := &repositoryStub{environments: []queries.GetEnvironmentsWithSitespeedEnabledRow{{ID: 1, Url: "https://example.com", SitespeedUrls: raw}}}
		err := NewService(repo, &config.Config{SitespeedEndpoint: "https://example.invalid", SitespeedAPIKey: "secret"}).Scrape(context.Background(), 1)
		require.Error(t, err)
		assert.True(t, goqueue.IsUnrecoverable(err))
		assert.Zero(t, repo.inserted)
	}
}

func TestSitespeedConnectionFailuresRemainRetryable(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	endpoint := upstream.URL
	upstream.Close()
	for _, cancelled := range []bool{false, true} {
		ctx, cancel := context.WithCancel(t.Context())
		if cancelled {
			cancel()
		}
		repo := &repositoryStub{environments: []queries.GetEnvironmentsWithSitespeedEnabledRow{{ID: 1, Url: "https://example.com"}}}
		err := NewService(repo, &config.Config{SitespeedEndpoint: endpoint, SitespeedAPIKey: "secret"}).Scrape(ctx, 1)
		cancel()
		require.Error(t, err)
		assert.False(t, goqueue.IsUnrecoverable(err))
		assert.Zero(t, repo.inserted)
	}
}
