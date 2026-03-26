package checker

import (
	"context"
	"sync"
)

type Status string

const (
	StatusGreen  Status = "green"
	StatusYellow Status = "yellow"
	StatusRed    Status = "red"
)

type Check struct {
	ID      string `json:"id"`
	Level   Status `json:"level"`
	Message string `json:"message"`
	Source  string `json:"source"`
	Link    string `json:"link,omitempty"`
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

func (o *Output) Success(id, message, source string, link string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.checks = append(o.checks, Check{
		ID:      id,
		Level:   StatusGreen,
		Message: message,
		Source:  source,
		Link:    link,
	})
}

func (o *Output) Warning(id, message, source string, link string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.checks = append(o.checks, Check{
		ID:      id,
		Level:   StatusYellow,
		Message: message,
		Source:  source,
		Link:    link,
	})
	if !o.ignores[id] && o.status != StatusRed {
		o.status = StatusYellow
	}
}

func (o *Output) Error(id, message, source string, link string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.checks = append(o.checks, Check{
		ID:      id,
		Level:   StatusRed,
		Message: message,
		Source:  source,
		Link:    link,
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
			check(ctx, input, output)
		}(fn)
	}
	wg.Wait()

	return output.Result()
}
