package shopwarepackages

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func feedServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/feeds/package/store.shopware.com/swagplatformsecurity.json", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestPackageFeedParsesVersions(t *testing.T) {
	srv := feedServer(t, http.StatusOK, `{
		"name": "store.shopware.com/swagplatformsecurity",
		"link": "https://store.shopware.com/en/detail/67457acbc21beb4318cf782a7ec39a5c",
		"date": "2026-08-26T08:58:12+00:00",
		"versions": [
			{
				"version": "4.0.14",
				"date": "2026-08-26T08:58:12+00:00",
				"changelog": "<ul><li>Added fix for GHSA-6qhw-38wm-7g7h</li></ul>",
				"require": {"shopware/core": "~6.7.0"}
			},
			{
				"version": "3.0.18",
				"date": "2026-08-25T15:53:34+00:00",
				"changelog": "<ul><li>Added fix for GHSA-rrc3-p9vx-5373</li></ul>",
				"require": {"shopware/core": "~6.6.0"}
			}
		]
	}`)

	pkg, err := NewClient(srv.URL, nil).PackageFeed(context.Background(), "store.shopware.com/swagplatformsecurity")
	require.NoError(t, err)

	assert.Equal(t, "store.shopware.com/swagplatformsecurity", pkg.Name)
	assert.Equal(t, "https://store.shopware.com/en/detail/67457acbc21beb4318cf782a7ec39a5c", pkg.Link)
	require.Len(t, pkg.Versions, 2)

	assert.Equal(t, "4.0.14", pkg.Versions[0].Version)
	assert.Contains(t, pkg.Versions[0].Changelog, "GHSA-6qhw-38wm-7g7h")
	assert.Equal(t, "~6.7.0", pkg.Versions[0].Require["shopware/core"])
	require.NotNil(t, pkg.Versions[0].ReleasedAt)
	assert.Equal(t, time.Date(2026, 8, 26, 8, 58, 12, 0, time.UTC), *pkg.Versions[0].ReleasedAt)

	assert.Equal(t, "3.0.18", pkg.Versions[1].Version)
	assert.Equal(t, "~6.6.0", pkg.Versions[1].Require["shopware/core"])
}

func TestPackageFeedToleratesMissingDate(t *testing.T) {
	srv := feedServer(t, http.StatusOK, `{
		"name": "store.shopware.com/swagplatformsecurity",
		"versions": [{"version": "1.0.0", "changelog": "", "require": {"shopware/core": "*"}}]
	}`)

	pkg, err := NewClient(srv.URL, nil).PackageFeed(context.Background(), "store.shopware.com/swagplatformsecurity")
	require.NoError(t, err)
	require.Len(t, pkg.Versions, 1)
	assert.Nil(t, pkg.Versions[0].ReleasedAt)
}

func TestPackageFeedErrorsOnNon200(t *testing.T) {
	srv := feedServer(t, http.StatusTeapot, `{}`)

	_, err := NewClient(srv.URL, nil).PackageFeed(context.Background(), "store.shopware.com/swagplatformsecurity")
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrNotFound)
	assert.Contains(t, err.Error(), "418")
}

func TestPackageFeedNotFound(t *testing.T) {
	srv := feedServer(t, http.StatusNotFound, `{}`)

	_, err := NewClient(srv.URL, nil).PackageFeed(context.Background(), "store.shopware.com/swagplatformsecurity")
	require.ErrorIs(t, err, ErrNotFound)
}

func TestPackageFeedErrorsOnInvalidJSON(t *testing.T) {
	srv := feedServer(t, http.StatusOK, `not json`)

	_, err := NewClient(srv.URL, nil).PackageFeed(context.Background(), "store.shopware.com/swagplatformsecurity")
	require.Error(t, err)
}

func TestNewClientDefaults(t *testing.T) {
	client := NewClient("", nil)
	assert.Equal(t, DefaultBaseURL, client.baseURL)
	assert.NotNil(t, client.httpClient)
}
