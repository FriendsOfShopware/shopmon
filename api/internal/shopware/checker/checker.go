package checker

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type Status string

const (
	StatusGreen  Status = "green"
	StatusYellow Status = "yellow"
	StatusRed    Status = "red"
)

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

type Input struct {
	Extensions     []Extension
	Config         ShopConfig
	ScheduledTasks []ScheduledTask
	QueueInfo      []QueueInfo
	CacheInfo      CacheInfo
	Client         HTTPClient
	Ignores        []string
}

type Result struct {
	Status Status  `json:"status"`
	Checks []Check `json:"checks"`
}

type Output struct {
	mu      sync.Mutex
	status  Status
	checks  []Check
	ignores map[string]bool
}

func NewOutput(ignores []string) *Output {
	ignoreMap := make(map[string]bool, len(ignores))
	for _, id := range ignores {
		ignoreMap[id] = true
	}
	return &Output{
		status:  StatusGreen,
		checks:  make([]Check, 0),
		ignores: ignoreMap,
	}
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
	if !o.ignores[id] && o.status != StatusRed {
		o.status = StatusYellow
	}
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
	if !o.ignores[id] {
		o.status = StatusRed
	}
}

func (o *Output) Result() Result {
	return Result{
		Status: o.status,
		Checks: o.checks,
	}
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
