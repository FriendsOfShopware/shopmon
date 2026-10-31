package monitoring

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"unicode/utf8"
)

var (
	// ErrAccountEventNotFound is returned when the event does not belong to
	// the environment, or no longer exists.
	ErrAccountEventNotFound = errors.New("account event not found")
	// ErrAccountEventAlreadyAcknowledged keeps a double submit from looking
	// like it recorded a second review.
	ErrAccountEventAlreadyAcknowledged = errors.New("account event already acknowledged")
)

// maxAcknowledgeNoteLen bounds the reviewer's note before it is stored.
const maxAcknowledgeNoteLen = 500

// AcknowledgeAccountEvent records that a reviewer accepted an account change
// and refreshes the environment so the finding stops colouring its status.
func (s *Service) AcknowledgeAccountEvent(ctx context.Context, userID string, environmentID int32, eventID int64, note *string) error {
	if _, err := s.authorizedEnvironment(ctx, userID, environmentID); err != nil {
		return err
	}

	acknowledged, err := s.repository.AcknowledgeAccountEvent(ctx, environmentID, eventID, userID, boundNote(note))
	if err != nil {
		return fmt.Errorf("acknowledge account event: %w", err)
	}
	if !acknowledged {
		// The row either belongs to another environment, does not exist, or
		// was already reviewed. The handler maps both cases; the caller gets
		// no signal about rows outside their environment either way.
		return ErrAccountEventNotFound
	}

	s.refreshAfterAcknowledge(ctx, environmentID)
	return nil
}

// SetAccountMonitoring turns account monitoring on or off for one environment.
func (s *Service) SetAccountMonitoring(ctx context.Context, userID string, environmentID int32, enabled bool) error {
	if _, err := s.authorizedEnvironment(ctx, userID, environmentID); err != nil {
		return err
	}

	if err := s.repository.SetAccountMonitoring(ctx, environmentID, enabled, userID); err != nil {
		return fmt.Errorf("set account monitoring: %w", err)
	}

	s.refreshAfterAcknowledge(ctx, environmentID)
	return nil
}

// refreshAfterAcknowledge queues a scrape so the environment status is
// recomputed now rather than at the next hourly run.
//
// Deliberately unthrottled. A throttle looks attractive -- acknowledging
// several findings in a row then hits the shop once instead of five times --
// but it drops the later requests rather than coalescing them: a confirmation
// arriving after a scrape has already read the open findings would queue
// nothing, and the status would stay red until the next hourly run. Correct
// status beats saving a handful of requests on a rare human action.
//
// A failure here is not fatal: the scheduled scrape corrects the status either
// way, it just takes longer.
func (s *Service) refreshAfterAcknowledge(ctx context.Context, environmentID int32) {
	if s.jobs == nil {
		return
	}
	if err := s.jobs.DispatchEnvironmentScrape(ctx, environmentID); err != nil {
		slog.WarnContext(ctx, "failed to queue refresh after account acknowledgement",
			"environmentId", environmentID, "error", err)
	}
}

// boundNote caps the reviewer's note.
//
// Counted and cut in runes, not bytes: Huma validates maxLength by character
// count, so a note of 500 accented characters is well over 500 bytes. Slicing
// bytes would split the final character, and Postgres rejects the invalid
// UTF-8 that results -- losing the acknowledgement over a note.
func boundNote(note *string) *string {
	if note == nil || *note == "" {
		return nil
	}
	s := *note
	if utf8.RuneCountInString(s) > maxAcknowledgeNoteLen {
		s = string([]rune(s)[:maxAcknowledgeNoteLen])
	}
	return &s
}
