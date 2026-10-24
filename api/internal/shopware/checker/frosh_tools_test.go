package checker

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockHTTPClient implements HTTPClient for testing checkFroshTools.
type mockHTTPClient struct {
	responses map[string][]byte
	errs      map[string]error
}

func (m *mockHTTPClient) Get(_ context.Context, path string) ([]byte, error) {
	if err, ok := m.errs[path]; ok {
		return nil, err
	}
	if body, ok := m.responses[path]; ok {
		return body, nil
	}
	return nil, errors.New("unexpected path: " + path)
}

func findCheck(checks []Check, id string) (Check, bool) {
	for _, c := range checks {
		if c.ID == id {
			return c, true
		}
	}
	return Check{}, false
}

func TestMapFroshChecks(t *testing.T) {
	tests := []struct {
		name            string
		check           froshToolsCheck
		wantEmitted     bool
		wantID          string
		wantLevel       Status
		wantKey         string
		wantCurrent     string
		wantRecommended string
	}{
		{
			name:        "STATE_OK known snippet emits success with mapped key",
			check:       froshToolsCheck{Snippet: "phpGood", State: "STATE_OK", URL: "http://example.com"},
			wantEmitted: true,
			wantID:      "frosh.phpGood",
			wantLevel:   StatusGreen,
			wantKey:     "check.frosh.phpGood",
		},
		{
			name:        "STATE_WARNING emits warning",
			check:       froshToolsCheck{Snippet: "phpOutdated", State: "STATE_WARNING"},
			wantEmitted: true,
			wantID:      "frosh.phpOutdated",
			wantLevel:   StatusYellow,
			wantKey:     "check.frosh.phpOutdated",
		},
		{
			name:        "STATE_ERROR emits error",
			check:       froshToolsCheck{Snippet: "cacheWarning", State: "STATE_ERROR"},
			wantEmitted: true,
			wantID:      "frosh.cacheWarning",
			wantLevel:   StatusRed,
			wantKey:     "check.frosh.cacheWarning",
		},
		{
			name:        "ignored snippet always emits success even on error state",
			check:       froshToolsCheck{Snippet: "scheduledTaskWarning", State: "STATE_ERROR"},
			wantEmitted: true,
			wantID:      "frosh.scheduledTaskWarning",
			wantLevel:   StatusGreen,
			wantKey:     "check.frosh.scheduledTaskWarning",
		},
		{
			name:        "composer security advisories stays green despite duplicating the Security source",
			check:       froshToolsCheck{Snippet: "Composer security advisories", State: "STATE_ERROR"},
			wantEmitted: true,
			wantID:      "frosh.Composer security advisories",
			wantLevel:   StatusGreen,
			wantKey:     "check.frosh.Composer security advisories",
		},
		{
			name:        "unknown snippet still keyed, snippet param carries fallback",
			check:       froshToolsCheck{Snippet: "somethingCustom", State: "STATE_OK"},
			wantEmitted: true,
			wantID:      "frosh.somethingCustom",
			wantLevel:   StatusGreen,
			wantKey:     "check.frosh.somethingCustom",
		},
		{
			name:            "current and recommended carried as params",
			check:           froshToolsCheck{Snippet: "phpOutdated", State: "STATE_WARNING", Current: new("7.4"), Recommended: new("8.2")},
			wantEmitted:     true,
			wantID:          "frosh.phpOutdated",
			wantLevel:       StatusYellow,
			wantKey:         "check.frosh.phpOutdated",
			wantCurrent:     "7.4",
			wantRecommended: "8.2",
		},
		{
			name:        "empty snippet uses id for check id and key",
			check:       froshToolsCheck{ID: "custom-id", Snippet: "", State: "STATE_OK"},
			wantEmitted: true,
			wantID:      "frosh.custom-id",
			wantLevel:   StatusGreen,
			wantKey:     "check.frosh.custom-id",
		},
		{
			name:        "empty snippet and id is skipped",
			check:       froshToolsCheck{State: "STATE_OK"},
			wantEmitted: false,
		},
		{
			name:        "unknown state is not emitted",
			check:       froshToolsCheck{Snippet: "phpGood", State: "STATE_UNKNOWN"},
			wantEmitted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewOutput(nil)
			mapFroshChecks([]froshToolsCheck{tt.check}, output)

			checks := output.Result().Checks
			if !tt.wantEmitted {
				assert.Empty(t, checks, "expected no checks")
				return
			}

			require.Len(t, checks, 1)
			check := checks[0]
			assert.Equal(t, tt.wantID, check.ID)
			assert.Equal(t, tt.wantLevel, check.Level)
			assert.Equal(t, tt.wantKey, check.MessageKey)
			assert.Equal(t, "FroshTools", check.Source)

			if tt.wantCurrent != "" {
				assert.Equal(t, tt.wantCurrent, check.MessageParams["current"])
				assert.Equal(t, tt.wantRecommended, check.MessageParams["recommended"])
			}
		})
	}
}

func TestCheckFroshTools_ExtensionNotPresent(t *testing.T) {
	output := NewOutput(nil)
	// Client would error if called; verify it's not called.
	client := &mockHTTPClient{errs: map[string]error{}}
	checkFroshTools(context.Background(), Input{
		Extensions: []Extension{{Name: "OtherPlugin", Active: true, Installed: true}},
		Client:     client,
	}, output)

	result := output.Result()
	assert.Empty(t, result.Checks, "expected no checks when FroshTools not installed")
	assert.Empty(t, result.Unavailable, "an absent extension is a known state, not an unavailable source")
}

func TestCheckFroshTools_ExtensionInactive(t *testing.T) {
	output := NewOutput(nil)
	checkFroshTools(context.Background(), Input{
		Extensions: []Extension{{Name: "FroshTools", Active: false, Installed: true}},
		Client:     &mockHTTPClient{},
	}, output)

	assert.Empty(t, output.Result().Checks, "expected no checks when FroshTools inactive")
}

func TestCheckFroshTools_NilClient(t *testing.T) {
	output := NewOutput(nil)
	checkFroshTools(context.Background(), Input{
		Extensions: []Extension{{Name: "FroshTools", Active: true, Installed: true}},
		Client:     nil,
	}, output)

	assert.Empty(t, output.Result().Checks, "expected no checks when client is nil")
}

func TestCheckFroshTools_HealthAndPerformance(t *testing.T) {
	client := &mockHTTPClient{
		responses: map[string][]byte{
			"/_action/frosh-tools/health/status":      []byte(`[{"snippet":"phpGood","state":"STATE_OK"}]`),
			"/_action/frosh-tools/performance/status": []byte(`[{"snippet":"cacheWarning","state":"STATE_WARNING"}]`),
		},
	}

	output := NewOutput(nil)
	checkFroshTools(context.Background(), Input{
		Extensions: []Extension{{Name: "FroshTools", Active: true, Installed: true}},
		Client:     client,
	}, output)

	checks := output.Result().Checks
	require.Len(t, checks, 2)

	php, ok := findCheck(checks, "frosh.phpGood")
	require.True(t, ok, "expected to find frosh.phpGood")
	assert.Equal(t, StatusGreen, php.Level)
	cache, ok := findCheck(checks, "frosh.cacheWarning")
	require.True(t, ok, "expected to find frosh.cacheWarning")
	assert.Equal(t, StatusYellow, cache.Level)
}

func TestCheckFroshTools_HealthError(t *testing.T) {
	client := &mockHTTPClient{
		errs: map[string]error{
			"/_action/frosh-tools/health/status": errors.New("boom"),
		},
	}

	output := NewOutput(nil)
	checkFroshTools(context.Background(), Input{
		Extensions: []Extension{{Name: "FroshTools", Active: true, Installed: true}},
		Client:     client,
	}, output)

	result := output.Result()
	assert.Empty(t, result.Checks, "expected no checks when health request errors")
	assert.Equal(t, []UnavailableSource{{Source: SourceFroshTools, IDPrefix: prefixFroshTools}}, result.Unavailable,
		"a failed health request leaves the source unevaluated, not resolved")
}

func TestCheckFroshTools_PerformanceError(t *testing.T) {
	client := &mockHTTPClient{
		responses: map[string][]byte{
			"/_action/frosh-tools/health/status": []byte(`[{"snippet":"phpGood","state":"STATE_OK"}]`),
		},
		errs: map[string]error{
			"/_action/frosh-tools/performance/status": errors.New("upstream request timeout"),
		},
	}

	output := NewOutput(nil)
	checkFroshTools(context.Background(), Input{
		Extensions: []Extension{{Name: "FroshTools", Active: true, Installed: true}},
		Client:     client,
	}, output)

	result := output.Result()
	assert.Len(t, result.Checks, 1, "health checks that did arrive are kept")
	assert.Equal(t, []UnavailableSource{{Source: SourceFroshTools, IDPrefix: prefixFroshTools}}, result.Unavailable,
		"the performance checks are missing, so the source stays unevaluated")
}

func TestCheckFroshTools_MissingExtensions(t *testing.T) {
	output := NewOutput(nil)
	checkFroshTools(context.Background(), Input{
		Client:  &mockHTTPClient{errs: map[string]error{}},
		Missing: MissingData{Extensions: true},
	}, output)

	result := output.Result()
	assert.Empty(t, result.Checks)
	assert.Equal(t, []UnavailableSource{{Source: SourceFroshTools, IDPrefix: prefixFroshTools}}, result.Unavailable,
		"without the extension list FroshTools cannot be ruled in or out")
}

func TestCheckFroshTools_InvalidHealthJSON(t *testing.T) {
	client := &mockHTTPClient{
		responses: map[string][]byte{
			"/_action/frosh-tools/health/status": []byte(`not json`),
		},
	}

	output := NewOutput(nil)
	checkFroshTools(context.Background(), Input{
		Extensions: []Extension{{Name: "FroshTools", Active: true, Installed: true}},
		Client:     client,
	}, output)

	result := output.Result()
	assert.Empty(t, result.Checks, "expected no checks when health JSON invalid")
	assert.Equal(t, []UnavailableSource{{Source: SourceFroshTools, IDPrefix: prefixFroshTools}}, result.Unavailable,
		"unparseable health data leaves the source unevaluated, not resolved")
}

func TestCheckFroshTools_ExtractsRuntimeVersionsAndMatchesAdvisories(t *testing.T) {
	client := &mockHTTPClient{
		responses: map[string][]byte{
			"/_action/frosh-tools/health/status": []byte(`[
				{"id":"php-version","snippet":"PHP Version","state":"STATE_OK","current":"8.3.14","recommended":"min 8.2.0"},
				{"id":"mysql","snippet":"MySQL Version","state":"STATE_OK","current":"8.0.36","recommended":"min 8.0"}
			]`),
			"/_action/frosh-tools/performance/status": []byte(`[]`),
		},
	}

	result := RunAll(context.Background(), Input{
		Config:     ShopConfig{Version: "6.6.10.0"},
		Extensions: []Extension{{Name: "FroshTools", Active: true, Installed: true}},
		Client:     client,
		SecurityAdvisories: []SecurityAdvisory{
			{
				ID:       "SHOPMON-2026-0001",
				Title:    "PHP PDO parser vulnerability",
				Severity: "high",
				Origin:   "shopmon",
				Packages: []SecurityAdvisoryPackage{
					{PackageName: "php", AffectedVersions: "<8.4.0"},
				},
			},
			{
				ID:       "SHOPMON-2026-0002",
				Title:    "MySQL server vulnerability",
				Severity: "medium",
				Origin:   "shopmon",
				Packages: []SecurityAdvisoryPackage{
					{PackageName: "mysql", AffectedVersions: ">=8.0.0,<8.0.40"},
				},
			},
		},
	})

	require.NotNil(t, result.PhpVersion)
	assert.Equal(t, "8.3.14", *result.PhpVersion)
	require.NotNil(t, result.DatabaseVersion)
	assert.Equal(t, "MySQL 8.0.36", *result.DatabaseVersion)

	phpAdv, ok := findCheck(result.Checks, "security.SHOPMON-2026-0001")
	require.True(t, ok, "expected PHP advisory check to be emitted")
	assert.Equal(t, StatusRed, phpAdv.Level)
	assert.Equal(t, "php", phpAdv.MessageParams["package"])
	assert.Equal(t, "8.3.14", phpAdv.MessageParams["installedVersion"])

	mysqlAdv, ok := findCheck(result.Checks, "security.SHOPMON-2026-0002")
	require.True(t, ok, "expected MySQL advisory check to be emitted")
	assert.Equal(t, StatusRed, mysqlAdv.Level)
	assert.Equal(t, "mysql", mysqlAdv.MessageParams["package"])
	assert.Equal(t, "8.0.36", mysqlAdv.MessageParams["installedVersion"])
}

func TestCheckFroshTools_DistinguishesMariaDBFromMySQL(t *testing.T) {
	client := &mockHTTPClient{
		responses: map[string][]byte{
			"/_action/frosh-tools/health/status": []byte(`[
				{"id":"mysql","snippet":"MariaDB Version","state":"STATE_OK","current":"10.11.8-MariaDB-0ubuntu0.24.04.1","recommended":"min 10.11"}
			]`),
			"/_action/frosh-tools/performance/status": []byte(`[]`),
		},
	}

	result := RunAll(context.Background(), Input{
		Config:     ShopConfig{Version: "6.6.10.0"},
		Extensions: []Extension{{Name: "FroshTools", Active: true, Installed: true}},
		Client:     client,
		SecurityAdvisories: []SecurityAdvisory{
			{
				ID:       "SHOPMON-2026-0003",
				Title:    "MySQL-only vulnerability",
				Severity: "high",
				Origin:   "shopmon",
				Packages: []SecurityAdvisoryPackage{
					{PackageName: "mysql", AffectedVersions: "<11.0.0"},
				},
			},
			{
				ID:       "SHOPMON-2026-0004",
				Title:    "MariaDB vulnerability",
				Severity: "high",
				Origin:   "shopmon",
				Packages: []SecurityAdvisoryPackage{
					{PackageName: "mariadb", AffectedVersions: ">=10.11.0,<10.11.10"},
				},
			},
		},
	})

	require.NotNil(t, result.DatabaseVersion)
	assert.Equal(t, "MariaDB 10.11.8-MariaDB-0ubuntu0.24.04.1", *result.DatabaseVersion)

	_, mysqlMatched := findCheck(result.Checks, "security.SHOPMON-2026-0003")
	assert.False(t, mysqlMatched, "MariaDB environment must not match a mysql advisory")

	mariadbAdv, ok := findCheck(result.Checks, "security.SHOPMON-2026-0004")
	require.True(t, ok, "expected MariaDB advisory check to be emitted")
	assert.Equal(t, "mariadb", mariadbAdv.MessageParams["package"])
	assert.Equal(t, "10.11.8-MariaDB-0ubuntu0.24.04.1", mariadbAdv.MessageParams["installedVersion"])
}

func TestExtractRuntimeVersion_IgnoresUnknownAndEmpty(t *testing.T) {
	output := NewOutput(nil)
	mapFroshChecks([]froshToolsCheck{
		{ID: "php-version", Snippet: "PHP Version", State: "STATE_ERROR", Current: new("")},
		{ID: "mysql", Snippet: "MySQL Version", State: "STATE_ERROR", Current: new("unknown")},
	}, output)

	res := output.Result()
	assert.Nil(t, res.PhpVersion)
	assert.Nil(t, res.DatabaseVersion)
}
