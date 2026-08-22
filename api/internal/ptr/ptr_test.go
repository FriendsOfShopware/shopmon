package ptr_test

import (
	"testing"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/ptr"
)

func TestOf(t *testing.T) {
	p := new("hello")
	if p == nil || *p != "hello" {
		t.Fatalf("Of() = %v, want pointer to hello", p)
	}
}

func TestDeref(t *testing.T) {
	if got := ptr.Deref((*string)(nil)); got != "" {
		t.Fatalf("Deref(nil) = %q, want empty", got)
	}
	if got := ptr.Deref(new(42)); got != 42 {
		t.Fatalf("Deref(42) = %d, want 42", got)
	}
}

func TestDerefOr(t *testing.T) {
	if got := ptr.DerefOr((*string)(nil), "fallback"); got != "fallback" {
		t.Fatalf("DerefOr(nil) = %q, want fallback", got)
	}
	if got := ptr.DerefOr(new("value"), "fallback"); got != "value" {
		t.Fatalf("DerefOr(value) = %q, want value", got)
	}
}

func TestMap(t *testing.T) {
	if got := ptr.Map((*int32)(nil), func(v int32) int { return int(v) }); got != nil {
		t.Fatalf("Map(nil) = %v, want nil", got)
	}
	got := ptr.Map(new(int32(7)), func(v int32) int { return int(v) })
	if got == nil || *got != 7 {
		t.Fatalf("Map(7) = %v, want pointer to 7", got)
	}
}

func TestInt(t *testing.T) {
	if got := ptr.Int((*int32)(nil)); got != nil {
		t.Fatalf("Int(nil) = %v, want nil", got)
	}
	got := ptr.Int(new(int32(9)))
	if got == nil || *got != 9 {
		t.Fatalf("Int(9) = %v, want pointer to 9", got)
	}
}

func TestParseTime(t *testing.T) {
	if got := ptr.ParseTime(nil, time.RFC3339); got != nil {
		t.Fatalf("ParseTime(nil) = %v, want nil", got)
	}
	if got := ptr.ParseTime(new("not-a-time"), time.RFC3339); got != nil {
		t.Fatalf("ParseTime(invalid) = %v, want nil", got)
	}
	got := ptr.ParseTime(new("2024-01-02T03:04:05Z"), time.RFC3339)
	if got == nil || !got.Equal(time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)) {
		t.Fatalf("ParseTime(valid) = %v, want 2024-01-02T03:04:05Z", got)
	}
}
