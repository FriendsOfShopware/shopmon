package shopwareaccount

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// StorePlugin is a plugin entry returned by the store pluginsByName endpoint,
// flattened to the fields shopmon stores. Localized fields (label, description,
// short description, installation manual, changelog text) reflect the locale
// passed to PluginsByName.
type StorePlugin struct {
	ID                 int
	Name               string
	Label              string
	Description        string
	ShortDescription   string
	InstallationManual string
	Version            string
	RatingAverage      float64
	StoreLink          string
	IconPath           string
	ReleaseDate        string
	ProducerName       string
	ProducerWebsite    string
	Pictures           []StorePicture
	Changelogs         []StoreChangelog
}

// StorePicture is a listing image (screenshot) of a store plugin.
type StorePicture struct {
	URL      string
	Preview  bool
	Priority int
}

// StoreChangelog is a single changelog entry for a store plugin.
type StoreChangelog struct {
	Version      string
	Text         string
	CreationDate StoreDate
}

// StoreDate is the date wrapper the store API uses for timestamp fields.
type StoreDate struct {
	Date string `json:"date"`
}

// rawStorePlugin mirrors the raw pluginsByName response shape.
type rawStorePlugin struct {
	ID                 int       `json:"id"`
	Name               string    `json:"name"`
	Label              string    `json:"label"`
	Description        string    `json:"description"`
	InstallationManual string    `json:"installationManual"`
	Version            string    `json:"version"`
	RatingAverage      float64   `json:"ratingAverage"`
	Link               string    `json:"link"`
	IconPath           string    `json:"iconPath"`
	ReleaseDate        StoreDate `json:"releaseDate"`
	Producer           struct {
		Name    string `json:"name"`
		Website string `json:"website"`
	} `json:"producer"`
	Infos []struct {
		ShortDescription string `json:"shortDescription"`
	} `json:"infos"`
	Pictures []struct {
		RemoteLink string `json:"remoteLink"`
		Preview    bool   `json:"preview"`
		Priority   int    `json:"priority"`
	} `json:"pictures"`
	Changelog []struct {
		Version      string    `json:"version"`
		Text         string    `json:"text"`
		CreationDate StoreDate `json:"creationDate"`
	} `json:"changelog"`
}

// PluginsByName fetches store metadata (versions, rating, changelogs, pictures,
// localized descriptions, store link, ...) for the given technical names, scoped
// to a Shopware version. The locale must be in the store's underscore form
// (e.g. "en_GB", "de_DE"); hyphenated locales are silently ignored by the API.
func (c *Client) PluginsByName(ctx context.Context, locale, shopwareVersion string, technicalNames []string) ([]StorePlugin, error) {
	q := url.Values{}
	q.Set("locale", locale)
	q.Set("shopwareVersion", shopwareVersion)
	for _, name := range technicalNames {
		q.Add("technicalNames[]", name)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/pluginStore/pluginsByName?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch store plugins: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("store api returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var raw []rawStorePlugin
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse store plugins response: %w", err)
	}

	plugins := make([]StorePlugin, 0, len(raw))
	for i := range raw {
		plugins = append(plugins, raw[i].toStorePlugin())
	}
	return plugins, nil
}

func (r *rawStorePlugin) toStorePlugin() StorePlugin {
	sp := StorePlugin{
		ID:                 r.ID,
		Name:               r.Name,
		Label:              r.Label,
		Description:        r.Description,
		InstallationManual: r.InstallationManual,
		Version:            r.Version,
		RatingAverage:      r.RatingAverage,
		StoreLink:          normalizeStoreLink(r.Link),
		IconPath:           r.IconPath,
		ReleaseDate:        r.ReleaseDate.Date,
		ProducerName:       r.Producer.Name,
		ProducerWebsite:    r.Producer.Website,
	}
	if len(r.Infos) > 0 {
		sp.ShortDescription = r.Infos[0].ShortDescription
	}
	for _, p := range r.Pictures {
		if p.RemoteLink == "" {
			continue
		}
		sp.Pictures = append(sp.Pictures, StorePicture{
			URL:      p.RemoteLink,
			Preview:  p.Preview,
			Priority: p.Priority,
		})
	}
	for _, cl := range r.Changelog {
		sp.Changelogs = append(sp.Changelogs, StoreChangelog{
			Version:      cl.Version,
			Text:         cl.Text,
			CreationDate: cl.CreationDate,
		})
	}
	return sp
}

// normalizeStoreLink rewrites the explicit-port store URLs the API returns
// (e.g. http://store.shopware.com:80/... or https://store.shopware.com:443/...)
// to a clean https URL.
func normalizeStoreLink(link string) string {
	replacer := []struct{ from, to string }{
		{"http://store.shopware.com:80", "https://store.shopware.com"},
		{"https://store.shopware.com:443", "https://store.shopware.com"},
		{"http://store.shopware.com", "https://store.shopware.com"},
	}
	for _, rep := range replacer {
		if len(link) >= len(rep.from) && link[:len(rep.from)] == rep.from {
			return rep.to + link[len(rep.from):]
		}
	}
	return link
}
