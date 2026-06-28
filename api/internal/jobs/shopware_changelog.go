package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

// maxChangelogResponseBytes bounds how much we read from a single changelog
// response, so a misbehaving upstream can't exhaust worker memory.
const maxChangelogResponseBytes = 10 << 20 // 10 MB

// shopwareVersionRe guards the version strings taken from the external index
// before they are interpolated into a request URL, rejecting anything that
// isn't a plain dotted version with an optional pre-release suffix (e.g.
// 6.7.11.1 or 6.7.0.0-rc1).
var shopwareVersionRe = regexp.MustCompile(`^\d+(\.\d+){1,3}(-[0-9A-Za-z.]+)?$`)

// ShopwareChangelogHandler crawls the Shopware release changelog API hourly and
// caches every release in the shopware_version table. This removes the need to
// call any external service when answering version queries at request time.
type ShopwareChangelogHandler struct {
	queries *queries.Queries
	baseURL string
	client  *http.Client
}

func NewShopwareChangelogHandler(q *queries.Queries, cfg *config.Config) *ShopwareChangelogHandler {
	return &ShopwareChangelogHandler{
		queries: q,
		baseURL: strings.TrimSuffix(cfg.ShopwareChangelogURL, "/"),
		client:  httputil.NewHTTPClient(httputil.WithTimeout(30 * time.Second)),
	}
}

// changelogEntry mirrors the per-version JSON document served at
// {baseURL}/{version}.json.
type changelogEntry struct {
	Title   string    `json:"title"`
	Body    string    `json:"body"`
	Date    time.Time `json:"date"`
	Version string    `json:"version"`
}

func (h *ShopwareChangelogHandler) HandleSync(ctx context.Context, _ ShopwareChangelogSync) error {
	ctx, span := tracer.Start(ctx, "shopware.changelog.sync")
	defer span.End()

	versions, err := h.fetchIndex(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("fetch changelog index: %w", err)
	}

	knownList, err := h.queries.GetKnownShopwareVersions(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("load known versions: %w", err)
	}
	known := make(map[string]struct{}, len(knownList))
	for _, v := range knownList {
		known[v] = struct{}{}
	}

	var added, failed int
	for _, version := range versions {
		if _, ok := known[version]; ok {
			continue
		}

		if !shopwareVersionRe.MatchString(version) {
			slog.Warn("skipping malformed shopware version", "version", version)
			failed++
			continue
		}

		entry, err := h.fetchVersion(ctx, version)
		if err != nil {
			// Don't fail the whole job for a single bad release; log and move on
			// so the next run can retry it.
			slog.Warn("failed to fetch shopware changelog entry", "version", version, "error", err)
			failed++
			continue
		}

		if err := h.queries.UpsertShopwareVersion(ctx, queries.UpsertShopwareVersionParams{
			Version:     version,
			ReleaseDate: pgtype.Timestamp{Time: entry.Date, Valid: true},
			Title:       entry.Title,
			Body:        entry.Body,
		}); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("upsert shopware version %s: %w", version, err)
		}
		added++
	}

	span.SetAttributes(
		attribute.Int("shopware.versions.total", len(versions)),
		attribute.Int("shopware.versions.added", added),
		attribute.Int("shopware.versions.failed", failed),
	)
	slog.Info("synced shopware changelog", "total", len(versions), "added", added, "failed", failed)
	return nil
}

func (h *ShopwareChangelogHandler) fetchIndex(ctx context.Context) ([]string, error) {
	var versions []string
	if err := h.getJSON(ctx, h.baseURL+"/index.json", &versions); err != nil {
		return nil, err
	}
	return versions, nil
}

func (h *ShopwareChangelogHandler) fetchVersion(ctx context.Context, version string) (*changelogEntry, error) {
	var entry changelogEntry
	if err := h.getJSON(ctx, fmt.Sprintf("%s/%s.json", h.baseURL, version), &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

func (h *ShopwareChangelogHandler) getJSON(ctx context.Context, url string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxChangelogResponseBytes))
	if err != nil {
		return err
	}

	return json.Unmarshal(body, dst)
}
