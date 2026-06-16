package checker

import (
	"context"
	"errors"
	"testing"
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

func TestGetFroshMessage(t *testing.T) {
	tests := []struct {
		snippet string
		want    string
	}{
		{"phpGood", "PHP version is up to date"},
		{"phpOutdated", "PHP version is outdated"},
		{"opcacheNoJit", "OPcache JIT is not enabled"},
		{"cacheWarning", "Cache is using filesystem adapter"},
		{"unknownSnippet", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.snippet, func(t *testing.T) {
			if got := getFroshMessage(tt.snippet); got != tt.want {
				t.Errorf("getFroshMessage(%q) = %q, want %q", tt.snippet, got, tt.want)
			}
		})
	}
}

func TestMapFroshChecks(t *testing.T) {
	tests := []struct {
		name        string
		check       froshToolsCheck
		wantEmitted bool
		wantID      string
		wantLevel   Status
		wantMessage string
	}{
		{
			name:        "STATE_OK known snippet emits success with mapped message",
			check:       froshToolsCheck{Snippet: "phpGood", State: "STATE_OK", URL: "http://example.com"},
			wantEmitted: true,
			wantID:      "frosh.phpGood",
			wantLevel:   StatusGreen,
			wantMessage: "PHP version is up to date",
		},
		{
			name:        "STATE_WARNING emits warning",
			check:       froshToolsCheck{Snippet: "phpOutdated", State: "STATE_WARNING"},
			wantEmitted: true,
			wantID:      "frosh.phpOutdated",
			wantLevel:   StatusYellow,
			wantMessage: "PHP version is outdated",
		},
		{
			name:        "STATE_ERROR emits error",
			check:       froshToolsCheck{Snippet: "cacheWarning", State: "STATE_ERROR"},
			wantEmitted: true,
			wantID:      "frosh.cacheWarning",
			wantLevel:   StatusRed,
			wantMessage: "Cache is using filesystem adapter",
		},
		{
			name:        "ignored snippet always emits success even on error state",
			check:       froshToolsCheck{Snippet: "scheduledTaskWarning", State: "STATE_ERROR"},
			wantEmitted: true,
			wantID:      "frosh.scheduledTaskWarning",
			wantLevel:   StatusGreen,
			wantMessage: "Some scheduled tasks are overdue",
		},
		{
			name:        "unknown snippet falls back to snippet as message",
			check:       froshToolsCheck{Snippet: "somethingCustom", State: "STATE_OK"},
			wantEmitted: true,
			wantID:      "frosh.somethingCustom",
			wantLevel:   StatusGreen,
			wantMessage: "somethingCustom",
		},
		{
			name:        "current and recommended appended to message",
			check:       froshToolsCheck{Snippet: "phpOutdated", State: "STATE_WARNING", Current: strPtr("7.4"), Recommended: strPtr("8.2")},
			wantEmitted: true,
			wantID:      "frosh.phpOutdated",
			wantLevel:   StatusYellow,
			wantMessage: "PHP version is outdated (Current: 7.4, Recommended: 8.2)",
		},
		{
			name:        "empty snippet uses id for check id",
			check:       froshToolsCheck{ID: "custom-id", Snippet: "", State: "STATE_OK"},
			wantEmitted: true,
			wantID:      "frosh.custom-id",
			wantLevel:   StatusGreen,
			wantMessage: "",
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
				if len(checks) != 0 {
					t.Fatalf("expected no checks, got %d (%+v)", len(checks), checks)
				}
				return
			}

			if len(checks) != 1 {
				t.Fatalf("expected 1 check, got %d (%+v)", len(checks), checks)
			}
			check := checks[0]
			if check.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", check.ID, tt.wantID)
			}
			if check.Level != tt.wantLevel {
				t.Errorf("Level = %q, want %q", check.Level, tt.wantLevel)
			}
			if check.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q", check.Message, tt.wantMessage)
			}
			if check.Source != "FroshTools" {
				t.Errorf("Source = %q, want %q", check.Source, "FroshTools")
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

	if len(output.Result().Checks) != 0 {
		t.Errorf("expected no checks when FroshTools not installed, got %d", len(output.Result().Checks))
	}
}

func TestCheckFroshTools_ExtensionInactive(t *testing.T) {
	output := NewOutput(nil)
	checkFroshTools(context.Background(), Input{
		Extensions: []Extension{{Name: "FroshTools", Active: false, Installed: true}},
		Client:     &mockHTTPClient{},
	}, output)

	if len(output.Result().Checks) != 0 {
		t.Errorf("expected no checks when FroshTools inactive, got %d", len(output.Result().Checks))
	}
}

func TestCheckFroshTools_NilClient(t *testing.T) {
	output := NewOutput(nil)
	checkFroshTools(context.Background(), Input{
		Extensions: []Extension{{Name: "FroshTools", Active: true, Installed: true}},
		Client:     nil,
	}, output)

	if len(output.Result().Checks) != 0 {
		t.Errorf("expected no checks when client is nil, got %d", len(output.Result().Checks))
	}
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
	if len(checks) != 2 {
		t.Fatalf("expected 2 checks, got %d (%+v)", len(checks), checks)
	}

	php, ok := findCheck(checks, "frosh.phpGood")
	if !ok || php.Level != StatusGreen {
		t.Errorf("expected frosh.phpGood success, got %+v (found=%v)", php, ok)
	}
	cache, ok := findCheck(checks, "frosh.cacheWarning")
	if !ok || cache.Level != StatusYellow {
		t.Errorf("expected frosh.cacheWarning warning, got %+v (found=%v)", cache, ok)
	}
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

	if len(output.Result().Checks) != 0 {
		t.Errorf("expected no checks when health request errors, got %d", len(output.Result().Checks))
	}
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

	if len(output.Result().Checks) != 0 {
		t.Errorf("expected no checks when health JSON invalid, got %d", len(output.Result().Checks))
	}
}
