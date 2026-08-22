// Package ptr provides generic helpers for optional pointer values.
package ptr

import "time"

// Of returns a pointer to v.
//
//go:fix inline
func Of[T any](v T) *T {
	return new(v)
}

// Deref returns the value pointed to by p, or the zero value of T when p is nil.
func Deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// DerefOr returns the value pointed to by p, or fallback when p is nil.
func DerefOr[T any](p *T, fallback T) T {
	if p == nil {
		return fallback
	}
	return *p
}

// Map converts *T to *U with fn, preserving nil.
func Map[T, U any](p *T, fn func(T) U) *U {
	if p == nil {
		return nil
	}
	v := fn(*p)
	return &v
}

// Int converts a pointer to any integer type into *int, preserving nil.
func Int[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](p *T) *int {
	if p == nil {
		return nil
	}
	v := int(*p)
	return &v
}

// ParseTime parses *string with the given layout, returning nil when absent or invalid.
func ParseTime(value *string, layout string) *time.Time {
	if value == nil {
		return nil
	}
	parsed, err := time.Parse(layout, *value)
	if err != nil {
		return nil
	}
	return &parsed
}
