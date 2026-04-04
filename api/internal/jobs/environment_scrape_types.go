package jobs

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
	Name string `json:"name"`
	Size int32  `json:"size"`
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

// Store API response types
type storePlugin struct {
	Name          string           `json:"name"`
	Label         string           `json:"label"`
	Version       string           `json:"version"`
	LatestVersion string           `json:"latestVersion"`
	RatingAverage float64          `json:"ratingAverage"`
	StoreLink     string           `json:"link"`
	Changelogs    []storeChangelog `json:"changelog"`
}

type storeChangelog struct {
	Version      string `json:"version"`
	Text         string `json:"text"`
	CreationDate struct {
		Date string `json:"date"`
	} `json:"creationDate"`
}

// extensionEntry is a combined representation of plugins and apps.
type extensionEntry struct {
	Name          string               `json:"name"`
	Label         string               `json:"label"`
	Active        bool                 `json:"active"`
	Version       string               `json:"version"`
	LatestVersion *string              `json:"latestVersion"`
	Installed     bool                 `json:"installed"`
	RatingAverage *float64             `json:"ratingAverage"`
	StoreLink     *string              `json:"storeLink"`
	Changelog     []extensionChangelog `json:"changelog,omitempty"`
	InstalledAt   *string              `json:"installedAt"`
}

type extensionChangelog struct {
	Version      string `json:"version"`
	Text         string `json:"text"`
	CreationDate string `json:"creationDate"`
	IsCompatible bool   `json:"isCompatible"`
}

type extensionDiff struct {
	Name       string               `json:"name"`
	Label      string               `json:"label"`
	State      string               `json:"state"`
	OldVersion *string              `json:"old_version"`
	NewVersion *string              `json:"new_version"`
	Changelog  []extensionChangelog `json:"changelog,omitempty"`
	Active     bool                 `json:"active"`
}

type notificationLink struct {
	Name   string            `json:"name"`
	Params map[string]string `json:"params"`
}
