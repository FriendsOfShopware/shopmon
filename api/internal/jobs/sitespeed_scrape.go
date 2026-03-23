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
	shops, err := h.queries.GetShopsWithSitespeedEnabled(ctx)
	if err != nil {
		return fmt.Errorf("get shops with sitespeed: %w", err)
	}

	for _, shop := range shops {
		if err := h.scrapeShop(ctx, shop); err != nil {
			slog.Error("failed to scrape sitespeed", "shopId", shop.ID, "error", err)
		}
	}
	return nil
}

func (h *SitespeedScrapeHandler) HandleScrape(ctx context.Context, msg SitespeedScrape) error {
	shops, err := h.queries.GetShopsWithSitespeedEnabled(ctx)
	if err != nil {
		return err
	}

	for _, shop := range shops {
		if shop.ID == msg.ShopID {
			return h.scrapeShop(ctx, shop)
		}
	}

	return fmt.Errorf("shop %d not found or sitespeed not enabled", msg.ShopID)
}

func (h *SitespeedScrapeHandler) scrapeShop(ctx context.Context, shop queries.GetShopsWithSitespeedEnabledRow) error {
	log := slog.With("shopId", shop.ID)

	// Get URLs to test - default is the shop URL
	var urls []string
	if shop.SitespeedUrls != nil {
		json.Unmarshal(shop.SitespeedUrls, &urls)
	}
	for i, u := range urls {
		urls[i] = strings.Replace(u, "http://localhost:3889", "http://demoshop:8000", 1)
	}

	recordID := fmt.Sprintf("%sshop-%d", h.cfg.SitespeedPrefix, shop.ID)

	// Call sitespeed service
	log.Info("calling sitespeed service", "urls", len(urls), "recordId", recordID)
	endpoint := fmt.Sprintf("%s/api/result/%s", h.cfg.SitespeedEndpoint, recordID)

	reqBody, _ := json.Marshal(map[string]interface{}{
		"urls": urls,
		"headers": map[string]string{
			"shopmon-shop-token": shop.ShopToken,
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
	defer resp.Body.Close()

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
	json.Unmarshal(body, &result)

	h.queries.InsertShopSitespeed(ctx, queries.InsertShopSitespeedParams{
		ShopID:                 &shop.ID,
		DeploymentID:           shop.ActiveDeploymentID,
		Ttfb:                   result.TTFB,
		FullyLoaded:            result.FullyLoaded,
		LargestContentfulPaint: result.LargestContentfulPaint,
		FirstContentfulPaint:   result.FirstContentfulPaint,
		CumulativeLayoutShift:  result.CumulativeLayoutShift,
		TransferSize:           result.TransferSize,
	})

	screenshotURL := fmt.Sprintf("%s/screenshot/%s", h.cfg.SitespeedEndpoint, recordID)
	h.queries.UpdateShopImage(ctx, queries.UpdateShopImageParams{
		ShopImage: &screenshotURL,
		ID:        shop.ID,
	})

	log.Info("sitespeed scrape completed")
	return nil
}
