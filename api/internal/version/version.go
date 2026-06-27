// Package version provides comparison of dot-separated numeric version strings
// (e.g. Shopware extension versions like "6.6.1.0"). It is shared by the scrape
// write path and the read/handler path so both filter changelogs identically.
package version

import (
	"strconv"
	"strings"
)

// Compare compares two dot-separated numeric version strings.
// It returns 1 if a > b, -1 if a < b, and 0 if they are equal. Non-numeric or
// missing segments are treated as 0.
func Compare(a, b string) int {
	ap := strings.Split(a, ".")
	bp := strings.Split(b, ".")

	maxLen := len(ap)
	if len(bp) > maxLen {
		maxLen = len(bp)
	}

	for i := 0; i < maxLen; i++ {
		var an, bn int
		if i < len(ap) {
			an, _ = strconv.Atoi(ap[i])
		}
		if i < len(bp) {
			bn, _ = strconv.Atoi(bp[i])
		}
		if an > bn {
			return 1
		}
		if bn > an {
			return -1
		}
	}
	return 0
}
