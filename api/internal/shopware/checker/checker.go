package checker

import (
	"context"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"
)

type Status string

const (
	StatusGreen  Status = "green"
	StatusYellow Status = "yellow"
	StatusRed    Status = "red"
)

// Check sources. Every check carries the source it originates from, which is
// what the UI groups by.
const (
	SourceShopware   = "Shopware"
	SourceFroshTools = "FroshTools"
	SourceSecurity   = "Security"
)

// Check ID prefixes. Each checker names its checks after the data set it reads,
// which is what makes a prefix a precise handle on the checks a single checker
// owns — see UnavailableSource.
const (
	prefixEnv           = "shopware.env"
	prefixScheduledTask = "task."
	prefixFroshTools    = "frosh."
	prefixSecurity      = "security."
)

// UnavailableSource marks one checker's checks as unevaluated for this run.
//
// The source alone is too coarse to act on: several checkers report under
// SourceShopware, so a failed scheduled-task fetch must not imply that the
// environment and worker checks are unknown too — those were evaluated, and
// their findings disappearing is a real resolution. IDPrefix narrows an entry
// to the checks the failing checker owns.
type UnavailableSource struct {
	Source   string
	IDPrefix string
}

// Owns reports whether a persisted check falls under this entry.
func (u UnavailableSource) Owns(source, checkID string) bool {
	return source == u.Source && strings.HasPrefix(checkID, u.IDPrefix)
}

type Check struct {
	ID    string `json:"id"`
	Level Status `json:"level"`
	// MessageKey is a translation catalog key; MessageParams interpolate into it.
	// Rendering happens at the edges (UI per viewer, server for the English
	// fallback) so check text is never stored pre-translated.
	MessageKey    string         `json:"messageKey"`
	MessageParams map[string]any `json:"messageParams,omitempty"`
	Source        string         `json:"source"`
	Link          string         `json:"link,omitempty"`
}

type Extension struct {
	Name          string  `json:"name"`
	Label         string  `json:"label"`
	Active        bool    `json:"active"`
	Version       string  `json:"version"`
	LatestVersion *string `json:"latestVersion"`
	Installed     bool    `json:"installed"`
}

type ScheduledTask struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Status            string  `json:"status"`
	RunInterval       int     `json:"runInterval"`
	NextExecutionTime *string `json:"nextExecutionTime"`
	LastExecutionTime *string `json:"lastExecutionTime"`
}

type QueueInfo struct {
	Name string `json:"name"`
	Size int    `json:"size"`
}

type CacheInfo struct {
	Environment  string `json:"environment"`
	HttpCache    bool   `json:"httpCache"`
	CacheAdapter string `json:"cacheAdapter"`
}

type ShopConfig struct {
	Version     string `json:"version"`
	AdminWorker struct {
		EnableAdminWorker bool `json:"enableAdminWorker"`
	} `json:"adminWorker"`
}

// HTTPClient interface for making requests to the shop
type HTTPClient interface {
	Get(ctx context.Context, path string) ([]byte, error)
}

// MissingData marks upstream data sets the caller failed to collect for this
// run. A checker that depends on a missing data set must not derive checks from
// it — absent data is not a passing check — and marks its source unavailable
// instead, so the caller can carry the last known checks forward.
type MissingData struct {
	Extensions     bool
	ScheduledTasks bool
	CacheInfo      bool
}

type Input struct {
	Extensions     []Extension
	Config         ShopConfig
	ScheduledTasks []ScheduledTask
	QueueInfo      []QueueInfo
	CacheInfo      CacheInfo
	Client         HTTPClient
	Ignores        []string
	Missing        MissingData
}

type Result struct {
	Status Status  `json:"status"`
	Checks []Check `json:"checks"`
	// Unavailable lists the check groups that could not be evaluated in this
	// run. Their checks are missing from Checks because the data was
	// unreachable, not because the underlying problems went away.
	Unavailable []UnavailableSource `json:"unavailable,omitempty"`
}

// UnavailableNames returns the distinct source names of the unevaluated groups,
// for logging and span attributes.
func (r Result) UnavailableNames() []string {
	seen := make(map[string]bool, len(r.Unavailable))
	names := make([]string, 0, len(r.Unavailable))
	for _, u := range r.Unavailable {
		if seen[u.Source] {
			continue
		}
		seen[u.Source] = true
		names = append(names, u.Source)
	}
	return names
}

type Output struct {
	mu          sync.Mutex
	checks      []Check
	ignores     map[string]bool
	unavailable map[UnavailableSource]bool
}

func NewOutput(ignores []string) *Output {
	ignoreMap := make(map[string]bool, len(ignores))
	for _, id := range ignores {
		ignoreMap[id] = true
	}
	return &Output{
		checks:      make([]Check, 0),
		ignores:     ignoreMap,
		unavailable: make(map[UnavailableSource]bool),
	}
}

// MarkUnavailable records that the checks named with idPrefix could not be
// evaluated in this run.
func (o *Output) MarkUnavailable(source, idPrefix string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.unavailable[UnavailableSource{Source: source, IDPrefix: idPrefix}] = true
}

func (o *Output) Success(id, messageKey string, params map[string]any, source, link string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.checks = append(o.checks, Check{
		ID:            id,
		Level:         StatusGreen,
		MessageKey:    messageKey,
		MessageParams: params,
		Source:        source,
		Link:          link,
	})
}

func (o *Output) Warning(id, messageKey string, params map[string]any, source, link string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.checks = append(o.checks, Check{
		ID:            id,
		Level:         StatusYellow,
		MessageKey:    messageKey,
		MessageParams: params,
		Source:        source,
		Link:          link,
	})
}

func (o *Output) Error(id, messageKey string, params map[string]any, source, link string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.checks = append(o.checks, Check{
		ID:            id,
		Level:         StatusRed,
		MessageKey:    messageKey,
		MessageParams: params,
		Source:        source,
		Link:          link,
	})
}

func (o *Output) Result() Result {
	o.mu.Lock()
	defer o.mu.Unlock()

	unavailable := make([]UnavailableSource, 0, len(o.unavailable))
	for u := range o.unavailable {
		unavailable = append(unavailable, u)
	}
	sort.Slice(unavailable, func(i, j int) bool {
		if unavailable[i].Source != unavailable[j].Source {
			return unavailable[i].Source < unavailable[j].Source
		}
		return unavailable[i].IDPrefix < unavailable[j].IDPrefix
	})

	return Result{
		Status:      aggregateStatus(o.checks, o.ignores),
		Checks:      o.checks,
		Unavailable: unavailable,
	}
}

// AggregateStatus derives the environment status from a set of checks. Checks
// whose ID is ignored never escalate the status. It is exported so callers that
// combine a run's checks with carried-over ones can recompute the status the
// same way the run itself does.
func AggregateStatus(checks []Check, ignores []string) Status {
	ignoreMap := make(map[string]bool, len(ignores))
	for _, id := range ignores {
		ignoreMap[id] = true
	}
	return aggregateStatus(checks, ignoreMap)
}

func aggregateStatus(checks []Check, ignores map[string]bool) Status {
	status := StatusGreen
	for _, c := range checks {
		if ignores[c.ID] {
			continue
		}
		switch c.Level {
		case StatusRed:
			return StatusRed
		case StatusYellow:
			status = StatusYellow
		}
	}
	return status
}

// RunAll runs all checkers concurrently and returns the aggregated result.
func RunAll(ctx context.Context, input Input) Result {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	output := NewOutput(input.Ignores)

	var wg sync.WaitGroup
	checkers := []func(context.Context, Input, *Output){
		checkEnv,
		checkWorker,
		checkTasks,
		checkSecurity,
		checkFroshTools,
	}

	wg.Add(len(checkers))
	for _, fn := range checkers {
		go func(check func(context.Context, Input, *Output)) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					slog.Error("checker panicked", "panic", r)
				}
			}()
			check(ctx, input, output)
		}(fn)
	}
	wg.Wait()

	return output.Result()
}
