package monitoring

import (
	"context"
	"errors"
	"strings"
	"testing"
	"unicode/utf8"
)

func accountService(repository *stubRepository, jobs *stubJobDispatcher, member bool) *Service {
	return NewService(repository, &stubAuthorizer{member: member}, &stubEnvironmentGateway{}, jobs, nil)
}

// Acknowledging has to refresh the environment: the status is only ever
// written by a scrape, so without one the shop stays red until the next hour.
func TestAcknowledgeAccountEventRefreshesEnvironment(t *testing.T) {
	repository := &stubRepository{environment: Environment{ID: 7, OrganizationID: "org-1"}}
	jobs := &stubJobDispatcher{}
	service := accountService(repository, jobs, true)

	note := "checked with the team, this is our new colleague"
	err := service.AcknowledgeAccountEvent(context.Background(), "user-1", 7, 42, 3, &note)

	if err != nil {
		t.Fatalf("AcknowledgeAccountEvent: %v", err)
	}
	if repository.acknowledgedEvent != 42 || repository.acknowledgedEnvironment != 7 {
		t.Errorf("acknowledged event %d on environment %d", repository.acknowledgedEvent, repository.acknowledgedEnvironment)
	}
	if repository.acknowledgedBy != "user-1" {
		t.Errorf("acknowledgedBy = %q, the reviewer must be recorded", repository.acknowledgedBy)
	}
	if len(jobs.scrapeIDs) != 1 || jobs.scrapeIDs[0] != 7 {
		t.Errorf("scrapeIDs = %v, want one refresh of environment 7", jobs.scrapeIDs)
	}
}

// An event that is already reviewed must not look like a fresh acknowledgement.
func TestAcknowledgeAccountEventRejectsUnknownOrAcknowledged(t *testing.T) {
	repository := &stubRepository{
		environment:    Environment{ID: 7, OrganizationID: "org-1"},
		acknowledgeErr: ErrAccountEventNotFound,
	}
	jobs := &stubJobDispatcher{}
	service := accountService(repository, jobs, true)

	err := service.AcknowledgeAccountEvent(context.Background(), "user-1", 7, 42, 3, nil)

	if !errors.Is(err, ErrAccountEventNotFound) {
		t.Fatalf("err = %v, want ErrAccountEventNotFound", err)
	}
	if len(jobs.scrapeIDs) != 0 {
		t.Error("a rejected acknowledgement must not refresh the environment")
	}
}

func TestAcknowledgeAccountEventRequiresMembership(t *testing.T) {
	repository := &stubRepository{environment: Environment{ID: 7, OrganizationID: "org-1"}}
	jobs := &stubJobDispatcher{}
	service := accountService(repository, jobs, false)

	err := service.AcknowledgeAccountEvent(context.Background(), "outsider", 7, 42, 3, nil)

	if !errors.Is(err, ErrNotAuthorized) {
		t.Fatalf("err = %v, want ErrNotAuthorized", err)
	}
	if repository.acknowledgedEvent != 0 {
		t.Error("a non-member must not be able to clear a finding")
	}
}

// A throttle would drop later confirmations rather than coalescing them: a
// confirmation arriving after a queued scrape has already read the open
// findings would queue nothing, and the status would stay red until the next
// hourly run. Every confirmation therefore queues its own refresh.
func TestAcknowledgeAccountEventRefreshesForEveryConfirmation(t *testing.T) {
	repository := &stubRepository{
		environment: Environment{ID: 7, OrganizationID: "org-1"},
	}
	jobs := &stubJobDispatcher{}
	service := accountService(repository, jobs, true)

	for _, id := range []int64{1, 2, 3} {
		if err := service.AcknowledgeAccountEvent(context.Background(), "user-1", 7, id, 3, nil); err != nil {
			t.Fatalf("AcknowledgeAccountEvent(%d): %v", id, err)
		}
	}

	if len(jobs.scrapeIDs) != 3 {
		t.Errorf("scrapeIDs = %v, want one refresh per confirmation", jobs.scrapeIDs)
	}
}

// A reviewer note is user input and ends up in a stored column.
func TestAcknowledgeAccountEventBoundsNote(t *testing.T) {
	repository := &stubRepository{environment: Environment{ID: 7, OrganizationID: "org-1"}}
	service := accountService(repository, &stubJobDispatcher{}, true)

	long := strings.Repeat("x", maxAcknowledgeNoteLen+100)
	if err := service.AcknowledgeAccountEvent(context.Background(), "user-1", 7, 42, 3, &long); err != nil {
		t.Fatalf("AcknowledgeAccountEvent: %v", err)
	}
	if repository.acknowledgedNote == nil || utf8.RuneCountInString(*repository.acknowledgedNote) != maxAcknowledgeNoteLen {
		t.Errorf("note was not bounded to %d characters", maxAcknowledgeNoteLen)
	}

	// Multibyte input: Huma counts characters, so this note is valid at 500
	// characters but far longer in bytes. Cutting bytes would split the last
	// character and Postgres would reject the write.
	multibyte := strings.Repeat("ä", maxAcknowledgeNoteLen+100)
	if err := service.AcknowledgeAccountEvent(context.Background(), "user-1", 7, 44, 3, &multibyte); err != nil {
		t.Fatalf("AcknowledgeAccountEvent: %v", err)
	}
	got := repository.acknowledgedNote
	if got == nil {
		t.Fatal("expected a note")
	}
	if !utf8.ValidString(*got) {
		t.Error("the truncated note is not valid UTF-8")
	}
	if utf8.RuneCountInString(*got) != maxAcknowledgeNoteLen {
		t.Errorf("note has %d characters, want %d", utf8.RuneCountInString(*got), maxAcknowledgeNoteLen)
	}

	empty := ""
	if err := service.AcknowledgeAccountEvent(context.Background(), "user-1", 7, 43, 3, &empty); err != nil {
		t.Fatalf("AcknowledgeAccountEvent: %v", err)
	}
	if repository.acknowledgedNote != nil {
		t.Error("an empty note must be stored as NULL, not as an empty string")
	}
}

func TestSetAccountMonitoring(t *testing.T) {
	repository := &stubRepository{environment: Environment{ID: 7, OrganizationID: "org-1"}}
	jobs := &stubJobDispatcher{}
	service := accountService(repository, jobs, true)

	if err := service.SetAccountMonitoring(context.Background(), "user-1", 7, false); err != nil {
		t.Fatalf("SetAccountMonitoring: %v", err)
	}
	if repository.accountMonitoringCalls != 1 || repository.accountMonitoringEnabled {
		t.Errorf("calls=%d enabled=%v", repository.accountMonitoringCalls, repository.accountMonitoringEnabled)
	}
	// Disabling has to refresh too, or the checks it removes keep colouring
	// the status until the next scheduled scrape.
	if len(jobs.scrapeIDs) != 1 {
		t.Errorf("scrapeIDs = %v, want a refresh", jobs.scrapeIDs)
	}
}

func TestSetAccountMonitoringRequiresMembership(t *testing.T) {
	repository := &stubRepository{environment: Environment{ID: 7, OrganizationID: "org-1"}}
	service := accountService(repository, &stubJobDispatcher{}, false)

	if err := service.SetAccountMonitoring(context.Background(), "outsider", 7, false); !errors.Is(err, ErrNotAuthorized) {
		t.Fatalf("err = %v, want ErrNotAuthorized", err)
	}
	if repository.accountMonitoringCalls != 0 {
		t.Error("a non-member must not be able to switch monitoring off")
	}
}

// The revision pins what the reviewer actually saw. A scrape can refresh an
// open finding with a worse state between the page rendering and the click, and
// closing it on the older view would bury a change nobody has read.
func TestAcknowledgeAccountEventPassesTheReviewedRevision(t *testing.T) {
	repository := &stubRepository{environment: Environment{ID: 7, OrganizationID: "org-1"}}
	service := accountService(repository, &stubJobDispatcher{}, true)

	if err := service.AcknowledgeAccountEvent(context.Background(), "user-1", 7, 42, 5, nil); err != nil {
		t.Fatalf("AcknowledgeAccountEvent: %v", err)
	}
	if repository.acknowledgedRevision != 5 {
		t.Errorf("acknowledgedRevision = %d, want the revision the reviewer saw", repository.acknowledgedRevision)
	}
}

func TestAcknowledgeAccountEventSurfacesAChangedFinding(t *testing.T) {
	repository := &stubRepository{
		environment:    Environment{ID: 7, OrganizationID: "org-1"},
		acknowledgeErr: ErrAccountEventChanged,
	}
	jobs := &stubJobDispatcher{}
	service := accountService(repository, jobs, true)

	err := service.AcknowledgeAccountEvent(context.Background(), "user-1", 7, 42, 1, nil)

	if !errors.Is(err, ErrAccountEventChanged) {
		t.Fatalf("err = %v, want ErrAccountEventChanged", err)
	}
	if len(jobs.scrapeIDs) != 0 {
		t.Error("a rejected acknowledgement must not refresh the environment")
	}
}
