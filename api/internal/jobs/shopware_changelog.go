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

	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

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

// phpCodeRe extracts the individual <code>X.Y</code> PHP versions from the
// "tested on PHP" list item present in newer release bodies.
var phpCodeRe = regexp.MustCompile(`<code>([^<]+)</code>`)

func (h *ShopwareChangelogHandler) HandleSync(ctx context.Context, _ ShopwareChangelogSync) error {
	versions, err := h.fetchIndex(ctx)
	if err != nil {
		return fmt.Errorf("fetch changelog index: %w", err)
	}

	knownList, err := h.queries.GetKnownShopwareVersions(ctx)
	if err != nil {
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

		entry, err := h.fetchVersion(ctx, version)
		if err != nil {
			// Don't fail the whole job for a single bad release; log and move on
			// so the next run can retry it.
			slog.Warn("failed to fetch shopware changelog entry", "version", version, "error", err)
			failed++
			continue
		}

		phpVersions, err := json.Marshal(parsePHPVersions(entry.Body))
		if err != nil {
			return fmt.Errorf("marshal php versions: %w", err)
		}

		if err := h.queries.UpsertShopwareVersion(ctx, queries.UpsertShopwareVersionParams{
			Version:     version,
			ReleaseDate: pgtype.Timestamp{Time: entry.Date, Valid: true},
			PhpVersions: phpVersions,
			Title:       entry.Title,
			Body:        entry.Body,
		}); err != nil {
			return fmt.Errorf("upsert shopware version %s: %w", version, err)
		}
		added++
	}

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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, dst)
}

// parsePHPVersions extracts the supported PHP versions from a release body. Newer
// releases contain a list item like:
//
//	<li>tested on PHP <code>8.2</code>, <code>8.4</code> and <code>8.5</code></li>
//
// Older releases don't advertise PHP versions, in which case an empty slice is
// returned. The result is always non-nil so it marshals to a JSON array.
func parsePHPVersions(body string) []string {
	versions := []string{}

	idx := strings.Index(body, "tested on PHP")
	if idx == -1 {
		return versions
	}

	segment := body[idx:]
	// Bound the search to the single list item so we don't pick up the
	// MySQL/MariaDB <code> entries on the following line.
	if end := strings.Index(segment, "</li>"); end != -1 {
		segment = segment[:end]
	}

	for _, m := range phpCodeRe.FindAllStringSubmatch(segment, -1) {
		if v := strings.TrimSpace(m[1]); v != "" {
			versions = append(versions, v)
		}
	}

	return versions
}
