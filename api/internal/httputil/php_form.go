package httputil

import (
	"net/url"
	"strconv"
	"strings"
)

// PHPFormArray returns the values of a flat PHP-style form array, accepting
// field[] and field[0], field[1], etc. Scalar fields, associative keys and nested
// arrays are ignored. Values retain their casing, duplicates and empty strings;
// order across different form keys is unspecified.
func PHPFormArray(form url.Values, field string) []string {
	var values []string
	prefix := field + "["
	for key, entries := range form {
		if !strings.HasPrefix(key, prefix) || !strings.HasSuffix(key, "]") {
			continue
		}
		index := strings.TrimSuffix(strings.TrimPrefix(key, prefix), "]")
		if index != "" {
			if n, err := strconv.Atoi(index); err != nil || n < 0 {
				continue
			}
		}
		values = append(values, entries...)
	}
	return values
}
