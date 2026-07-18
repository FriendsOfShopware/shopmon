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
