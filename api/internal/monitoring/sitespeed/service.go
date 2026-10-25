package sitespeed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	goqueue "github.com/shyim/go-queue"

	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/metrics"
)

// Repository is the persistence boundary used by sitespeed jobs.
type Repository interface {
	GetEnvironmentsWithSitespeedEnabled(context.Context) ([]queries.GetEnvironmentsWithSitespeedEnabledRow, error)
	InsertEnvironmentSitespeed(context.Context, queries.InsertEnvironmentSitespeedParams) error
	UpdateEnvironmentImage(context.Context, queries.UpdateEnvironmentImageParams) error
}

type Service struct {
	queries Repository
	cfg     *config.Config
}

func NewService(q Repository, cfg *config.Config) *Service {
	return &Service{queries: q, cfg: cfg}
}

func (s *Service) Scrape(ctx context.Context, environmentID int32) (err error) {
	environments, err := s.queries.GetEnvironmentsWithSitespeedEnabled(ctx)
	if err != nil {
		metrics.RecordSitespeedOutcome(ctx, metrics.OutcomeError)
		return fmt.Errorf("list sitespeed-enabled environments: %w", err)
	}

	for _, env := range environments {
		if env.ID == environmentID {
			return s.scrapeEnvironment(ctx, env)
		}
	}

	slog.InfoContext(ctx, "sitespeed environment deleted or disabled, skipping", "environmentId", environmentID)
	metrics.RecordSitespeedOutcome(ctx, metrics.OutcomeSkipped)
	return nil
}

func (s *Service) scrapeEnvironment(ctx context.Context, env queries.GetEnvironmentsWithSitespeedEnabledRow) (err error) {
	log := slog.With("environmentId", env.ID)

	if s.cfg.SitespeedEndpoint == "" || s.cfg.SitespeedAPIKey == "" {
		log.Info("sitespeed not configured, skipping")
		metrics.RecordSitespeedOutcome(ctx, metrics.OutcomeSkipped)
		return nil
	}

	defer func() {
		if err != nil {
			metrics.RecordSitespeedOutcome(ctx, metrics.OutcomeError)
			return
		}
		metrics.RecordSitespeedOutcome(ctx, metrics.OutcomeOK)
	}()

	// Get URLs to test - default is the environment URL
	var urls []string
	if env.SitespeedUrls != nil {
		if err := json.Unmarshal(env.SitespeedUrls, &urls); err != nil {
			return goqueue.Unrecoverable(fmt.Errorf("invalid sitespeed URL configuration for environment %d", env.ID))
		}
	}
	if len(urls) == 0 {
		urls = []string{env.Url}
	}
	urls, err = NormalizeURLs(urls)
	if err != nil {
		return goqueue.Unrecoverable(err)
	}
	if _, err := NormalizeURLs([]string{s.cfg.SitespeedEndpoint}); err != nil {
		return goqueue.Unrecoverable(fmt.Errorf("invalid sitespeed service endpoint"))
	}

	for i, u := range urls {
		urls[i] = strings.Replace(u, "http://localhost:3889", "http://demoshop:8000", 1)
	}

	recordID := fmt.Sprintf("%senvironment-%d", s.cfg.SitespeedPrefix, env.ID)

	// Call sitespeed service
	log.Info("calling sitespeed service", "urls", len(urls), "recordId", recordID)
	endpoint := fmt.Sprintf("%s/api/result/%s", s.cfg.SitespeedEndpoint, recordID)

	reqBody, err := json.Marshal(map[string]any{
		"urls": urls,
		"headers": map[string]string{
			"shopmon-shop-token": env.EnvironmentToken,
		},
	})
	if err != nil {
		return fmt.Errorf("marshal sitespeed request for environment %d: %w", env.ID, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return goqueue.Unrecoverable(fmt.Errorf("invalid sitespeed service request for environment %d", env.ID))
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.SitespeedAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := httputil.NewHTTPClient(httputil.WithTimeout(300 * time.Second))
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("call sitespeed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Do not include upstream bodies: they may contain URLs, headers or credentials.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		statusErr := fmt.Errorf("sitespeed service returned HTTP %d", resp.StatusCode)
		if resp.StatusCode >= 300 && resp.StatusCode < 500 && resp.StatusCode != http.StatusRequestTimeout && resp.StatusCode != http.StatusTooManyRequests {
			return goqueue.Unrecoverable(statusErr)
		}
		return statusErr
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return fmt.Errorf("read sitespeed response for environment %d: %w", env.ID, err)
	}

	// Parse response and save metrics
	log.Info("saving sitespeed metrics")
	var result struct {
		TTFB                   *int32   `json:"ttfb"`
		FullyLoaded            *int32   `json:"fullyLoaded"`
		LargestContentfulPaint *int32   `json:"largestContentfulPaint"`
		FirstContentfulPaint   *int32   `json:"firstContentfulPaint"`
		CumulativeLayoutShift  *float32 `json:"cumulativeLayoutShift"`
		TransferSize           *int32   `json:"transferSize"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("unmarshal sitespeed result for environment %d: %w", env.ID, err)
	}

	if err := s.queries.InsertEnvironmentSitespeed(ctx, queries.InsertEnvironmentSitespeedParams{
		EnvironmentID:          &env.ID,
		DeploymentID:           env.ActiveDeploymentID,
		Ttfb:                   result.TTFB,
		FullyLoaded:            result.FullyLoaded,
		LargestContentfulPaint: result.LargestContentfulPaint,
		FirstContentfulPaint:   result.FirstContentfulPaint,
		CumulativeLayoutShift:  result.CumulativeLayoutShift,
		TransferSize:           result.TransferSize,
	}); err != nil {
		return fmt.Errorf("insert environment sitespeed for environment %d: %w", env.ID, err)
	}

	screenshotURL := fmt.Sprintf("%s/screenshot/%s", s.cfg.SitespeedEndpoint, recordID)
	if err := s.queries.UpdateEnvironmentImage(ctx, queries.UpdateEnvironmentImageParams{
		EnvironmentImage: &screenshotURL,
		ID:               env.ID,
	}); err != nil {
		return fmt.Errorf("update environment image for environment %d: %w", env.ID, err)
	}

	log.Info("sitespeed scrape completed")
	return nil
}
