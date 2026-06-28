package version

import "testing"

func TestCompare(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"2.0.0", "1.9.9", 1},
		{"1.9.9", "2.0.0", -1},
		{"10.0.0", "9.0.0", 1}, // numeric, not lexicographic
		{"1.2", "1.2.0", 0},    // missing segments treated as 0
		{"1.2.1", "1.2", 1},
		{"6.6.1.0", "6.6.0.0", 1},
	}
	for _, c := range cases {
		if got := Compare(c.a, c.b); got != c.want {
			t.Errorf("Compare(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}
