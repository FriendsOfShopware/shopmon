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

func strPtr(s string) *string { return &s }

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
			name:        "unknown snippet still keyed, snippet param carries fallback",
			check:       froshToolsCheck{Snippet: "somethingCustom", State: "STATE_OK"},
			wantEmitted: true,
			wantID:      "frosh.somethingCustom",
			wantLevel:   StatusGreen,
			wantKey:     "check.frosh.somethingCustom",
		},
		{
			name:            "current and recommended carried as params",
			check:           froshToolsCheck{Snippet: "phpOutdated", State: "STATE_WARNING", Current: strPtr("7.4"), Recommended: strPtr("8.2")},
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

	assert.Empty(t, output.Result().Checks, "expected no checks when FroshTools not installed")
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

	assert.Empty(t, output.Result().Checks, "expected no checks when health request errors")
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

	assert.Empty(t, output.Result().Checks, "expected no checks when health JSON invalid")
}
