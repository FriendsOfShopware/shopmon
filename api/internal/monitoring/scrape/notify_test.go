package scrape

import (
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/shopware/checker"
)

func oldCheck(id, level string) queries.EnvironmentCheck {
	key := "check." + id
	return queries.EnvironmentCheck{CheckID: id, Level: level, MessageKey: &key}
}

func newCheck(id, level string) checker.Check {
	return checker.Check{ID: id, Level: checker.Status(level), MessageKey: "check." + id}
}

func TestComputeStatusReasonsDegraded(t *testing.T) {
	old := []queries.EnvironmentCheck{
		oldCheck("a", "green"),
		oldCheck("b", "green"),
	}
	// b worsened to red, c appeared as yellow, a stayed green.
	now := []checker.Check{
		newCheck("a", "green"),
		newCheck("b", "red"),
		newCheck("c", "yellow"),
	}

	reasons := computeStatusReasons(old, now, true)
	if len(reasons) != 2 {
		t.Fatalf("expected 2 degradation reasons, got %d: %+v", len(reasons), reasons)
	}
	got := map[string]string{}
	for _, r := range reasons {
		got[r.Key] = r.Level
	}
	if got["check.b"] != "red" || got["check.c"] != "yellow" {
		t.Fatalf("unexpected reasons: %+v", got)
	}
}

func TestComputeStatusReasonsRecovered(t *testing.T) {
	old := []queries.EnvironmentCheck{
		oldCheck("a", "red"),
		oldCheck("b", "yellow"),
		oldCheck("c", "green"),
	}
	// a recovered to green, b disappeared (implicitly green), c stayed green.
	now := []checker.Check{
		newCheck("a", "green"),
		newCheck("c", "green"),
	}

	reasons := computeStatusReasons(old, now, false)
	if len(reasons) != 2 {
		t.Fatalf("expected 2 recovery reasons, got %d: %+v", len(reasons), reasons)
	}
	got := map[string]bool{}
	for _, r := range reasons {
		got[r.Key] = true
	}
	if !got["check.a"] || !got["check.b"] {
		t.Fatalf("expected a and b as recovered reasons, got %+v", got)
	}
}

func TestComputeStatusReasonsNoChange(t *testing.T) {
	old := []queries.EnvironmentCheck{oldCheck("a", "red")}
	now := []checker.Check{newCheck("a", "red")}

	if reasons := computeStatusReasons(old, now, true); len(reasons) != 0 {
		t.Fatalf("expected no degradation reasons when nothing worsened, got %+v", reasons)
	}
	if reasons := computeStatusReasons(old, now, false); len(reasons) != 0 {
		t.Fatalf("expected no recovery reasons when nothing improved, got %+v", reasons)
	}
}

func sourcedCheck(id, level, source string) queries.EnvironmentCheck {
	c := oldCheck(id, level)
	c.Source = source
	c.Params = []byte(`{"snippet":"` + id + `"}`)
	return c
}

func TestCarryOverChecksKeepsUnavailableSourceFindings(t *testing.T) {
	old := []queries.EnvironmentCheck{
		sourcedCheck("frosh.elasticsearch", "red", checker.SourceFroshTools),
		sourcedCheck("frosh.phpGood", "green", checker.SourceFroshTools),
		sourcedCheck("shopware.env", "yellow", checker.SourceShopware),
	}
	// The Shopware checks were re-evaluated; FroshTools timed out.
	now := []checker.Check{newCheck("shopware.env", "green")}

	carried := carryOverChecks(old, now, []string{checker.SourceFroshTools})
	if len(carried) != 2 {
		t.Fatalf("expected both FroshTools checks carried over, got %d: %+v", len(carried), carried)
	}

	byID := map[string]checker.Check{}
	for _, c := range carried {
		byID[c.ID] = c
	}
	es, ok := byID["frosh.elasticsearch"]
	if !ok {
		t.Fatalf("expected frosh.elasticsearch to be carried over, got %+v", byID)
	}
	if es.Level != checker.StatusRed {
		t.Fatalf("expected the carried check to keep its red level, got %q", es.Level)
	}
	if es.MessageKey != "check.frosh.elasticsearch" || es.Source != checker.SourceFroshTools {
		t.Fatalf("expected key/source preserved, got %+v", es)
	}
	if es.MessageParams["snippet"] != "frosh.elasticsearch" {
		t.Fatalf("expected params preserved, got %+v", es.MessageParams)
	}
	if _, ok := byID["shopware.env"]; ok {
		t.Fatalf("checks of available sources must not be carried over: %+v", byID)
	}

	// The environment must stay red instead of reporting a false recovery.
	if status := checker.AggregateStatus(append(now, carried...), nil); status != checker.StatusRed {
		t.Fatalf("expected status to stay red while FroshTools is unreachable, got %q", status)
	}
}

func TestCarryOverChecksSkipsRereportedChecks(t *testing.T) {
	old := []queries.EnvironmentCheck{
		sourcedCheck("frosh.phpGood", "red", checker.SourceFroshTools),
		sourcedCheck("frosh.cacheWarning", "yellow", checker.SourceFroshTools),
	}
	// Health responded (phpGood now green), performance did not.
	now := []checker.Check{newCheck("frosh.phpGood", "green")}

	carried := carryOverChecks(old, now, []string{checker.SourceFroshTools})
	if len(carried) != 1 || carried[0].ID != "frosh.cacheWarning" {
		t.Fatalf("expected only the unreported check carried over, got %+v", carried)
	}
}

func TestCarryOverChecksWithoutUnavailableSources(t *testing.T) {
	old := []queries.EnvironmentCheck{sourcedCheck("frosh.phpGood", "red", checker.SourceFroshTools)}

	if carried := carryOverChecks(old, nil, nil); len(carried) != 0 {
		t.Fatalf("expected nothing carried over when every source was evaluated, got %+v", carried)
	}
}
