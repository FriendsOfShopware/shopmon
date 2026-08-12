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

	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/metrics"
)

type Service struct {
	queries *queries.Queries
	cfg     *config.Config
}

func NewService(q *queries.Queries, cfg *config.Config) *Service {
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

	metrics.RecordSitespeedOutcome(ctx, metrics.OutcomeError)
	return fmt.Errorf("environment %d not found or sitespeed not enabled", environmentID)
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
			return fmt.Errorf("unmarshal sitespeed urls for environment %d: %w", env.ID, err)
		}
	}
	if len(urls) == 0 {
		urls = []string{env.Url}
	}
	for i, u := range urls {
		urls[i] = strings.Replace(u, "http://localhost:3889", "http://demoshop:8000", 1)
	}

	recordID := fmt.Sprintf("%senvironment-%d", s.cfg.SitespeedPrefix, env.ID)

	// Call sitespeed service
	log.Info("calling sitespeed service", "urls", len(urls), "recordId", recordID)
	endpoint := fmt.Sprintf("%s/api/result/%s", s.cfg.SitespeedEndpoint, recordID)

	reqBody, err := json.Marshal(map[string]interface{}{
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
		return fmt.Errorf("create sitespeed request for environment %d: %w", env.ID, err)
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.SitespeedAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httputil.NewHTTPClient(httputil.WithTimeout(300 * time.Second)).Do(req)
	if err != nil {
		return fmt.Errorf("call sitespeed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return fmt.Errorf("read sitespeed response for environment %d: %w", env.ID, err)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("sitespeed error (%d): %s", resp.StatusCode, string(body))
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
