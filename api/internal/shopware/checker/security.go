package checker

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

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

func checkSecurity(input Input, output *Output) {
	resp, err := httputil.NewHTTPClient().Get("https://raw.githubusercontent.com/FriendsOfShopware/shopware-static-data/main/data/security.json")
	if err != nil {
		slog.Warn("failed to fetch security advisories", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("security advisories endpoint returned non-200", "status", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Warn("failed to read security response", "error", err)
		return
	}

	var data securityData
	if err := json.Unmarshal(body, &data); err != nil {
		slog.Warn("failed to parse security data", "error", err)
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
