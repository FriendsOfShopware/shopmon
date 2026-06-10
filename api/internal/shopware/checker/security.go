package checker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

const securityDataURL = "https://raw.githubusercontent.com/FriendsOfShopware/shopware-static-data/main/data/security.json"

// securityDataTTL is how long a fetched security.json is reused before refetching.
// The upstream file changes infrequently, so caching avoids hammering GitHub on
// every environment scrape.
const securityDataTTL = 30 * time.Minute

type securityData struct {
	LatestPluginVersion string                      `json:"latestPluginVersion"`
	Advisories          map[string]securityAdvisory `json:"advisories"`
	VersionToAdvisories map[string][]string         `json:"versionToAdvisories"`
}

type securityAdvisory struct {
	Title  string `json:"title"`
	Link   string `json:"link"`
	CVE    string `json:"cve"`
	Source string `json:"source"`
}

var securityCache struct {
	mu        sync.Mutex
	data      *securityData
	fetchedAt time.Time
}

// fetchSecurityData returns the security advisories, served from an in-memory
// cache when a previous fetch is still within securityDataTTL.
func fetchSecurityData(ctx context.Context) (*securityData, error) {
	securityCache.mu.Lock()
	defer securityCache.mu.Unlock()

	if securityCache.data != nil && time.Since(securityCache.fetchedAt) < securityDataTTL {
		return securityCache.data, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, securityDataURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create security advisories request: %w", err)
	}

	resp, err := httputil.NewHTTPClient(httputil.WithTimeout(15 * time.Second)).Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch security advisories: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("security advisories endpoint returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read security response: %w", err)
	}

	var data securityData
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("parse security data: %w", err)
	}

	securityCache.data = &data
	securityCache.fetchedAt = time.Now()
	return &data, nil
}

func checkSecurity(ctx context.Context, input Input, output *Output) {
	data, err := fetchSecurityData(ctx)
	if err != nil {
		slog.Warn("failed to load security advisories", "error", err)
		return
	}

	version := input.Config.Version
	advisoryIDs, hasAdvisories := data.VersionToAdvisories[version]
	if !hasAdvisories || len(advisoryIDs) == 0 {
		return
	}

	for _, ext := range input.Extensions {
		if ext.Name == "SwagPlatformSecurity" && ext.Active && ext.Installed {
			if ext.LatestVersion != nil && ext.Version >= *ext.LatestVersion {
				// Plugin is at latest version, skip security warnings
				return
			}
			if ext.LatestVersion != nil && ext.Version >= data.LatestPluginVersion {
				return
			}
		}
	}

	for _, advisoryID := range advisoryIDs {
		advisory, exists := data.Advisories[advisoryID]
		if !exists {
			continue
		}

		id := fmt.Sprintf("security.%s", advisoryID)
		message := advisory.Title
		if advisory.CVE != "" {
			message = fmt.Sprintf("%s (%s)", advisory.Title, advisory.CVE)
		}

		output.Error(id, message, "Security", advisory.Link)
	}
}
