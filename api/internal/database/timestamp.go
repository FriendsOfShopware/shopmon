package database

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// Time returns the wall time stored in ts. Invalid timestamps yield the zero time.
func Time(ts pgtype.Timestamp) time.Time {
	return ts.Time
}

// TimePtr returns a pointer to the wall time stored in ts, or nil when invalid.
func TimePtr(ts pgtype.Timestamp) *time.Time {
	if !ts.Valid {
		return nil
	}
	return &ts.Time
}

// TimestampFromPtr is the inverse of TimePtr: a nil time yields an invalid
// (SQL NULL) timestamp. Values are normalized to UTC so a caller's location
// cannot shift what is stored in a timestamp-without-time-zone column.
func TimestampFromPtr(t *time.Time) pgtype.Timestamp {
	if t == nil {
		return pgtype.Timestamp{}
	}
	return pgtype.Timestamp{Time: t.UTC(), Valid: true}
}
