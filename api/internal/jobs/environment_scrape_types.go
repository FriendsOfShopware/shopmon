package jobs

import (
	"encoding/json"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/shopwareaccount"
)

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

// extensionEntry is a combined representation of plugins and apps. Store is set
// when the extension is known to the Shopware store (api.shopware.com); such
// extensions are persisted into the normalized store_extension* tables, while
// the rest land in environment_extension.
type extensionEntry struct {
	Name          string              `json:"name"`
	Label         string              `json:"label"`
	Active        bool                `json:"active"`
	Version       string              `json:"version"`
	LatestVersion *string             `json:"latestVersion"`
	Installed     bool                `json:"installed"`
	InstalledAt   *string             `json:"installedAt"`
	Store         *storeExtensionData `json:"-"`
}

// storeExtensionData holds the per-locale store metadata fetched for an
// extension. en is the en_GB response, de the de_DE response; either may be nil
// if that locale call failed.
type storeExtensionData struct {
	en *shopwareaccount.StorePlugin
	de *shopwareaccount.StorePlugin
}

// primary returns the locale response to use for locale-independent fields,
// preferring en_GB and falling back to de_DE.
func (s *storeExtensionData) primary() *shopwareaccount.StorePlugin {
	if s == nil {
		return nil
	}
	if s.en != nil {
		return s.en
	}
	return s.de
}

type extensionChangelog struct {
	Version      string    `json:"version"`
	Text         string    `json:"text"`
	TextDe       string    `json:"textDe,omitempty"`
	CreationDate time.Time `json:"creationDate"`
}

type extensionDiff struct {
	Name       string               `json:"name"`
	Label      string               `json:"label"`
	State      string               `json:"state"`
	OldVersion *string              `json:"oldVersion,omitempty"`
	NewVersion *string              `json:"newVersion,omitempty"`
	Changelog  []extensionChangelog `json:"changelog,omitempty"`
	Active     bool                 `json:"active"`
}
