package uptime

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// pgTimestamp converts a wall time to a UTC-normalized pgx timestamp for the
// timestamp-without-time-zone columns used by the uptime tables.
func pgTimestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: t.UTC(), Valid: true}
}

// pgDate converts a wall time to a pgx date (time portion discarded).
func pgDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t.UTC(), Valid: true}
}

// timePtr unwraps a nullable pgx timestamp.
func timePtr(ts pgtype.Timestamp) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}
