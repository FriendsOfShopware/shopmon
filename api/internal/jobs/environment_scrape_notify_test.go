package jobs

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
