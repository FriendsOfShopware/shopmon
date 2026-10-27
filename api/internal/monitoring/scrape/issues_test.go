package scrape

import (
	"context"
	"fmt"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/shopware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldNotifyIssue(t *testing.T) {
	// Deterministic problems (missing ACL privilege, bad credentials, …)
	// notify on first sight but never twice for the same occurrence.
	assert.True(t, shouldNotifyIssue(true, 1, false))
	assert.False(t, shouldNotifyIssue(true, 2, true))

	// Transient problems (timeout, shop 5xx) stay quiet until they persist
	// across transientNotifyThreshold consecutive scrapes.
	assert.False(t, shouldNotifyIssue(false, 1, false))
	assert.False(t, shouldNotifyIssue(false, transientNotifyThreshold-1, false))
	assert.True(t, shouldNotifyIssue(false, transientNotifyThreshold, false))
	assert.False(t, shouldNotifyIssue(false, transientNotifyThreshold+5, true))
}

func TestIssueFromErrorMissingPrivilege(t *testing.T) {
	body := `{"errors":[{"status":"403","code":"FRAMEWORK__MISSING_PRIVILEGE_ERROR","title":"Forbidden","detail":"{\"message\":\"Missing privilege\",\"missingPrivileges\":[\"message_queue_stats:read\"]}"}]}`
	in := issueFromError(datasetQueue, fmt.Errorf("fetch: %w", &shopware.ApiError{StatusCode: 403, Body: body}))

	assert.Equal(t, "queue.missing_privilege", in.key)
	assert.Equal(t, string(shopware.FailureMissingPrivilege), in.kind)
	assert.Equal(t, datasetQueue, in.dataset)
	assert.True(t, in.deterministic)
	assert.Equal(t, "message_queue_stats:read", in.params["privileges"])
	// The raw upstream body must never reach stored params.
	assert.NotContains(t, fmt.Sprintf("%v", in.params), "Forbidden")
}

func TestIssueFromErrorTimeoutIsTransient(t *testing.T) {
	in := issueFromError(datasetAuth, fmt.Errorf("token request: %w", errTimeout{}))
	assert.Equal(t, string(shopware.FailureTimeout), in.kind)
	assert.False(t, in.deterministic)
	assert.NotContains(t, in.params, "privileges")
}

// TestRecordScrapeIssuesLifecycle exercises the persisted issue state across
// scrapes: a deterministic issue notifies once, a transient one only at the
// threshold, unevaluated datasets never resolve, and a fixed shop reports a
// recovery exactly for the issues owners were told about.
func TestRecordScrapeIssuesLifecycle(t *testing.T) {
	h, _, env := setupPersistTest(t)
	ctx := context.Background()

	privBody := `{"errors":[{"status":"403","code":"FRAMEWORK__MISSING_PRIVILEGE_ERROR","detail":"{\"missingPrivileges\":[\"message_queue_stats:read\"]}"}]}`
	queueIssue := issueFromError(datasetQueue, &shopware.ApiError{StatusCode: 403, Body: privBody})
	timeoutIssue := issueFromError(datasetPlugins, errTimeout{})
	evaluated := map[string]bool{datasetQueue: true, datasetPlugins: true}

	// First scrape: the ACL problem is notifiable immediately, the timeout not.
	res := h.recordScrapeIssues(ctx, env, []scrapeIssueInput{queueIssue, timeoutIssue}, evaluated)
	require.Len(t, res.toNotify, 1)
	assert.Equal(t, "queue.missing_privilege", res.toNotify[0].key)
	assert.Equal(t, "scrapeIssue.queue.missing_privilege", res.toNotify[0].messageKey)
	h.markIssuesNotified(ctx, env.ID, res.toNotify)

	// Same problems again: nothing re-notifies (timeout still below threshold).
	res = h.recordScrapeIssues(ctx, env, []scrapeIssueInput{queueIssue, timeoutIssue}, evaluated)
	assert.Empty(t, res.toNotify)

	// Third consecutive timeout crosses the threshold.
	res = h.recordScrapeIssues(ctx, env, []scrapeIssueInput{queueIssue, timeoutIssue}, evaluated)
	require.Len(t, res.toNotify, 1)
	assert.Equal(t, "plugins.timeout", res.toNotify[0].key)
	assert.Equal(t, "scrapeIssue.timeout", res.toNotify[0].messageKey)
	h.markIssuesNotified(ctx, env.ID, res.toNotify)

	// A scrape that never evaluated the datasets (e.g. auth failed first)
	// must not resolve them.
	res = h.recordScrapeIssues(ctx, env, nil, map[string]bool{})
	assert.False(t, res.recoveredNotified)
	stored, err := h.queries.GetEnvironmentScrapeIssues(ctx, env.ID)
	require.NoError(t, err)
	assert.Len(t, stored, 2)

	// Clean scrape with both datasets evaluated: both resolve, and since both
	// were notified, a recovery is due.
	res = h.recordScrapeIssues(ctx, env, nil, evaluated)
	assert.True(t, res.recoveredNotified)
	stored, err = h.queries.GetEnvironmentScrapeIssues(ctx, env.ID)
	require.NoError(t, err)
	assert.Empty(t, stored)

	// A problem that resolves before ever being notified stays silent.
	res = h.recordScrapeIssues(ctx, env, []scrapeIssueInput{timeoutIssue}, evaluated)
	assert.Empty(t, res.toNotify)
	res = h.recordScrapeIssues(ctx, env, nil, evaluated)
	assert.False(t, res.recoveredNotified)

	// Stored params are structured and owner-safe (no upstream bodies).
	res = h.recordScrapeIssues(ctx, env, []scrapeIssueInput{queueIssue}, evaluated)
	require.Len(t, res.toNotify, 1)
	stored, err = h.queries.GetEnvironmentScrapeIssues(ctx, env.ID)
	require.NoError(t, err)
	require.Len(t, stored, 1)
	assert.JSONEq(t, `{"dataset":"queue","privileges":"message_queue_stats:read"}`, string(stored[0].Params))
	assert.Equal(t, int32(1), stored[0].ConsecutiveCount)
}

type errTimeout struct{}

func (errTimeout) Error() string   { return "i/o timeout" }
func (errTimeout) Timeout() bool   { return true }
func (errTimeout) Temporary() bool { return true }
