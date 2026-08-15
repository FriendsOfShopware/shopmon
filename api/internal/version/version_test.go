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

func TestSatisfies(t *testing.T) {
	cases := []struct {
		version, constraint string
		want                bool
	}{
		{"6.7.5.0", ">=6.7.0.0,<6.7.10.1", true},
		{"6.7.10.1", ">=6.7.0.0,<6.7.10.1", false},
		{"6.6.10.0", "<6.6.10.18|>=6.7.0.0,<6.7.10.1", true},
		{"6.6.10.18", "<6.6.10.18|>=6.7.0.0,<6.7.10.1", false},
		{"6.5.0.0", ">=6.7.0.0,<6.7.10.1", false},
	}
	for _, c := range cases {
		got, err := Satisfies(c.version, c.constraint)
		if err != nil {
			t.Errorf("Satisfies(%q, %q) unexpected error: %v", c.version, c.constraint, err)
			continue
		}
		if got != c.want {
			t.Errorf("Satisfies(%q, %q) = %v, want %v", c.version, c.constraint, got, c.want)
		}
	}
}
