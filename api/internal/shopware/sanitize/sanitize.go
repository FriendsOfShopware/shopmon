// Package sanitize bounds shop-supplied strings before they reach the database
// or a user-visible column.
//
// Everything a shop returns is untrusted input: a NUL byte aborts the whole
// Postgres transaction, invalid UTF-8 breaks JSON encoding, and an unbounded
// response body would be stored verbatim. The helpers here are shared by every
// package that ingests shop responses so the rules stay identical across them.
package sanitize

import (
	"strings"
	"unicode/utf8"
)

const (
	// MaxMetaLen caps a single metadata value. Real values are short; anything
	// longer is a malformed or hostile response.
	MaxMetaLen = 512
	// MaxErrorLen caps error text that may embed a shop's response body.
	MaxErrorLen = 500
)

// Meta bounds optional shop-supplied metadata (type, purl, spec version, serial
// number, labels); anything unprintable or oversized is dropped rather than
// stored.
func Meta(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > MaxMetaLen || !utf8.ValidString(s) {
		return ""
	}
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			return ""
		}
	}
	return s
}

// Error bounds error text that may embed a shop's response body before it lands
// in a user-visible column: invalid UTF-8 and control characters are stripped
// (a NUL would abort the whole Postgres transaction), and the result is
// truncated on a rune boundary.
func Error(s string) string {
	s = strings.ToValidUTF8(s, "�")
	s = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return ' '
		}
		return r
	}, s)
	if len(s) > MaxErrorLen {
		cut := MaxErrorLen
		for cut > 0 && !utf8.RuneStart(s[cut]) {
			cut--
		}
		s = s[:cut] + "…"
	}
	return strings.TrimSpace(s)
}
