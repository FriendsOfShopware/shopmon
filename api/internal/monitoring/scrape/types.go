package scrape

import (
	"encoding/json"

	"github.com/friendsofshopware/shopmon/api/internal/catalog/storemodel"
)

type catalogState struct {
	member      map[string]bool
	compatKnown map[string]*string
}

// ---- Shopware API response types ----

type shopwareConfig struct {
	Version     string `json:"version"`
	Environment string `json:"environment"`
	HttpCache   bool   `json:"httpCache"`
	AdminWorker struct {
		EnableAdminWorker bool `json:"enableAdminWorker"`
	} `json:"adminWorker"`
	Settings struct {
		CacheTtl int `json:"cacheTtl"`
	} `json:"settings"`
}

type shopwarePlugin struct {
	Name           string  `json:"name"`
	Label          string  `json:"label"`
	Active         bool    `json:"active"`
	Version        string  `json:"version"`
	UpgradeVersion *string `json:"upgradeVersion"`
	InstalledAt    *string `json:"installedAt"`
}

type shopwareApp struct {
	Name      string `json:"name"`
	Label     string `json:"label"`
	Active    bool   `json:"active"`
	Version   string `json:"version"`
	CreatedAt string `json:"createdAt"`
}

type shopwareSearchResponse[T any] struct {
	Data []T `json:"data"`
}

type shopwareQueueEntry struct {
	Name string      `json:"name"`
	Size json.Number `json:"size"`
}

type shopwareScheduledTask struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Status            string  `json:"status"`
	RunInterval       int32   `json:"runInterval"`
	NextExecutionTime *string `json:"nextExecutionTime"`
	LastExecutionTime *string `json:"lastExecutionTime"`
}

type shopwareCacheInfo struct {
	Environment  string `json:"environment"`
	HttpCache    bool   `json:"httpCache"`
	CacheAdapter string `json:"cacheAdapter"`
}

// extensionEntry is a combined representation of plugins and apps. isStore is
// set when the extension exists in the shared store catalog (store_extension);
// such extensions are linked via environment_store_extension, while the rest
// land in environment_extension.
type extensionEntry struct {
	Name          string  `json:"name"`
	Label         string  `json:"label"`
	Active        bool    `json:"active"`
	Version       string  `json:"version"`
	LatestVersion *string `json:"latestVersion"`
	Installed     bool    `json:"installed"`
	InstalledAt   *string `json:"installedAt"`
	isStore       bool
}

type extensionChangelog = storemodel.Changelog

type extensionDiff struct {
	Name       string               `json:"name"`
	Label      string               `json:"label"`
	State      string               `json:"state"`
	OldVersion *string              `json:"oldVersion,omitempty"`
	NewVersion *string              `json:"newVersion,omitempty"`
	Changelog  []extensionChangelog `json:"changelog,omitempty"`
	Active     bool                 `json:"active"`
}
