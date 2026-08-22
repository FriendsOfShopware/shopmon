package database

import (
	"strings"
	"unicode"
)

const sqlOperationUnknown = "UNKNOWN"

// SQLSpanName returns a low-cardinality operation name for an OpenTelemetry DB
// span. sqlc embeds a leading `-- name: <Name> ...` comment in generated SQL;
// otelpgx's default first-word trim treats that `--` as the operation, which
// collapses Datadog resources to `query --`. Prefer the sqlc query name when
// present; otherwise fall back to the first SQL keyword.
func SQLSpanName(stmt string) string {
	for line := range strings.SplitSeq(stmt, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if name, ok := sqlcQueryName(line); ok {
			return name
		}
		if strings.HasPrefix(line, "--") {
			continue
		}
		return firstSQLKeyword(line)
	}
	return sqlOperationUnknown
}

// sqlcQueryName extracts the query name from a single sqlc annotation line
// (`-- name: Foo :exec`). The line must already be trimmed.
func sqlcQueryName(line string) (string, bool) {
	const prefix = "-- name:"
	if !strings.HasPrefix(line, prefix) {
		return "", false
	}
	rest := strings.TrimSpace(strings.TrimPrefix(line, prefix))
	if rest == "" {
		return "", false
	}
	name := rest
	if i := strings.IndexFunc(rest, unicode.IsSpace); i >= 0 {
		name = rest[:i]
	}
	if name == "" || !isSQLCQueryName(name) {
		return "", false
	}
	return name, true
}

func firstSQLKeyword(line string) string {
	for field := range strings.FieldsSeq(line) {
		return strings.ToUpper(field)
	}
	return sqlOperationUnknown
}

// isSQLCQueryName keeps span names low-cardinality by accepting only identifier-
// like sqlc names (letters, digits, underscore). Rejects anything that looks
// like SQL body text.
func isSQLCQueryName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		return false
	}
	return true
}
