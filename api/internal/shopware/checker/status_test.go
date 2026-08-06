package checker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAggregateStatus(t *testing.T) {
	tests := []struct {
		name    string
		checks  []Check
		ignores []string
		want    Status
	}{
		{
			name: "no checks is green",
			want: StatusGreen,
		},
		{
			name:   "only successes is green",
			checks: []Check{{ID: "a", Level: StatusGreen}, {ID: "b", Level: StatusGreen}},
			want:   StatusGreen,
		},
		{
			name:   "a warning escalates to yellow",
			checks: []Check{{ID: "a", Level: StatusGreen}, {ID: "b", Level: StatusYellow}},
			want:   StatusYellow,
		},
		{
			name:   "an error wins over a warning regardless of order",
			checks: []Check{{ID: "a", Level: StatusRed}, {ID: "b", Level: StatusYellow}},
			want:   StatusRed,
		},
		{
			name:    "ignored checks never escalate",
			checks:  []Check{{ID: "a", Level: StatusRed}, {ID: "b", Level: StatusYellow}},
			ignores: []string{"a", "b"},
			want:    StatusGreen,
		},
		{
			name:    "ignoring the error leaves the warning",
			checks:  []Check{{ID: "a", Level: StatusRed}, {ID: "b", Level: StatusYellow}},
			ignores: []string{"a"},
			want:    StatusYellow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, AggregateStatus(tt.checks, tt.ignores))
		})
	}
}

// The status an Output reports must be the one AggregateStatus derives from the
// same checks, so carried-over checks can be folded in without changing the
// rules.
func TestOutputResultStatusMatchesAggregate(t *testing.T) {
	output := NewOutput([]string{"ignored.error"})
	output.Success("ok", "check.ok", nil, SourceShopware, "")
	output.Warning("warn", "check.warn", nil, SourceShopware, "")
	output.Error("ignored.error", "check.err", nil, SourceShopware, "")

	result := output.Result()
	assert.Equal(t, StatusYellow, result.Status)
	assert.Equal(t, AggregateStatus(result.Checks, []string{"ignored.error"}), result.Status)
}
