package monitoring

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"
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

// accountRescrapeLock throttles the refresh triggered by acknowledging. The
// environment status is only ever written by a scrape, so clearing a finding
// means running one — but acknowledging five findings in a row must not hit
// the shop five times.
const accountRescrapeLock = time.Minute

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
// recomputed now rather than at the next hourly run. A failure here is not
// fatal: the scheduled scrape corrects the status either way, it just takes
// longer.
func (s *Service) refreshAfterAcknowledge(ctx context.Context, environmentID int32) {
	if s.jobs == nil {
		return
	}
	if !s.repository.TryLock(ctx, "account_ack_rescrape_"+strconv.Itoa(int(environmentID)), accountRescrapeLock) {
		return
	}
	if err := s.jobs.DispatchEnvironmentScrape(ctx, environmentID); err != nil {
		slog.WarnContext(ctx, "failed to queue refresh after account acknowledgement",
			"environmentId", environmentID, "error", err)
	}
}

func boundNote(note *string) *string {
	if note == nil || *note == "" {
		return nil
	}
	s := *note
	if len(s) > maxAcknowledgeNoteLen {
		s = s[:maxAcknowledgeNoteLen]
	}
	return &s
}
