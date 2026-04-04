package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SitespeedScrapeHandler struct {
	pool    *pgxpool.Pool
	queries *queries.Queries
	cfg     *config.Config
}

func NewSitespeedScrapeHandler(pool *pgxpool.Pool, q *queries.Queries, cfg *config.Config) *SitespeedScrapeHandler {
	return &SitespeedScrapeHandler{pool: pool, queries: q, cfg: cfg}
}

func (h *SitespeedScrapeHandler) HandleScrapeAll(ctx context.Context, _ SitespeedScrapeAll) error {
	environments, err := h.queries.GetEnvironmentsWithSitespeedEnabled(ctx)
	if err != nil {
		return fmt.Errorf("get environments with sitespeed: %w", err)
	}

	for _, env := range environments {
		if err := h.scrapeEnvironment(ctx, env); err != nil {
			slog.Error("failed to scrape sitespeed", "environmentId", env.ID, "error", err)
		}
	}
	return nil
}

func (h *SitespeedScrapeHandler) HandleScrape(ctx context.Context, msg SitespeedScrape) error {
	environments, err := h.queries.GetEnvironmentsWithSitespeedEnabled(ctx)
	if err != nil {
		return err
	}

	for _, env := range environments {
		if env.ID == msg.EnvironmentID {
			return h.scrapeEnvironment(ctx, env)
		}
	}

	return fmt.Errorf("environment %d not found or sitespeed not enabled", msg.EnvironmentID)
}

func (h *SitespeedScrapeHandler) scrapeEnvironment(ctx context.Context, env queries.GetEnvironmentsWithSitespeedEnabledRow) error {
	log := slog.With("environmentId", env.ID)

	// Get URLs to test - default is the environment URL
	var urls []string
	if env.SitespeedUrls != nil {
		if err := json.Unmarshal(env.SitespeedUrls, &urls); err != nil {
			slog.Error("failed to unmarshal sitespeed urls", "environmentId", env.ID, "error", err)
			return nil
		}
	}
	for i, u := range urls {
		urls[i] = strings.Replace(u, "http://localhost:3889", "http://demoshop:8000", 1)
	}

	recordID := fmt.Sprintf("%senvironment-%d", h.cfg.SitespeedPrefix, env.ID)

	// Call sitespeed service
	log.Info("calling sitespeed service", "urls", len(urls), "recordId", recordID)
	endpoint := fmt.Sprintf("%s/api/result/%s", h.cfg.SitespeedEndpoint, recordID)

	reqBody, _ := json.Marshal(map[string]interface{}{
		"urls": urls,
		"headers": map[string]string{
			"shopmon-shop-token": env.EnvironmentToken,
		},
	})

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+h.cfg.SitespeedAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httputil.NewHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("call sitespeed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
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
		slog.Error("failed to unmarshal sitespeed result", "environmentId", env.ID, "error", err)
	}

	if err := h.queries.InsertEnvironmentSitespeed(ctx, queries.InsertEnvironmentSitespeedParams{
		EnvironmentID:          &env.ID,
		DeploymentID:           env.ActiveDeploymentID,
		Ttfb:                   result.TTFB,
		FullyLoaded:            result.FullyLoaded,
		LargestContentfulPaint: result.LargestContentfulPaint,
		FirstContentfulPaint:   result.FirstContentfulPaint,
		CumulativeLayoutShift:  result.CumulativeLayoutShift,
		TransferSize:           result.TransferSize,
	}); err != nil {
		slog.Error("failed to insert environment sitespeed", "environmentId", env.ID, "error", err)
	}

	screenshotURL := fmt.Sprintf("%s/screenshot/%s", h.cfg.SitespeedEndpoint, recordID)
	if err := h.queries.UpdateEnvironmentImage(ctx, queries.UpdateEnvironmentImageParams{
		EnvironmentImage: &screenshotURL,
		ID:               env.ID,
	}); err != nil {
		slog.Error("failed to update environment image", "environmentId", env.ID, "error", err)
	}

	log.Info("sitespeed scrape completed")
	return nil
}
