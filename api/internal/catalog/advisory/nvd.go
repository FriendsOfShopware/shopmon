package advisory

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	defaultNVDAPIBase = "https://services.nvd.nist.gov/rest/json/cves/2.0"
	maxNVDBodyBytes   = 5 << 20
	// Public NVD limit without a key is ~5 requests / 30s.
	nvdDelayNoKey   = 7 * time.Second
	nvdDelayWithKey = 700 * time.Millisecond
	nvdBatchNoKey   = 10
	nvdBatchWithKey = 40
)

type nvdClient struct {
	baseURL    string
	apiKey     string
	userAgent  string
	httpClient *http.Client
}

func newNVDClient(apiKey, userAgent string, httpClient *http.Client) *nvdClient {
	if userAgent == "" {
		userAgent = defaultUserAgent
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &nvdClient{
		baseURL:    defaultNVDAPIBase,
		apiKey:     strings.TrimSpace(apiKey),
		userAgent:  userAgent,
		httpClient: httpClient,
	}
}

func (c *nvdClient) hasAPIKey() bool {
	return c.apiKey != ""
}

func (c *nvdClient) batchSize() int32 {
	if c.hasAPIKey() {
		return nvdBatchWithKey
	}
	return nvdBatchNoKey
}

func (c *nvdClient) fetchDelay() time.Duration {
	if c.hasAPIKey() {
		return nvdDelayWithKey
	}
	return nvdDelayNoKey
}

// fetchCVE loads a single CVE from NVD and maps it into our shared details shape.
func (c *nvdClient) fetchCVE(ctx context.Context, cveID string) (*githubAdvisoryDetails, error) {
	cveID = strings.TrimSpace(cveID)
	if cveID == "" {
		return nil, fmt.Errorf("empty cve id")
	}

	endpoint := c.baseURL + "?cveId=" + url.QueryEscape(cveID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create nvd request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.apiKey != "" {
		req.Header.Set("apiKey", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch nvd cve %s: %w", cveID, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxNVDBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("read nvd cve %s: %w", cveID, err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("nvd cve %s not found", cveID)
	}
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("nvd rate limited or forbidden (%d) for %s", resp.StatusCode, cveID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nvd cve %s status %d", cveID, resp.StatusCode)
	}

	var raw nvdResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode nvd cve %s: %w", cveID, err)
	}
	if len(raw.Vulnerabilities) == 0 {
		return nil, fmt.Errorf("nvd cve %s returned no results", cveID)
	}

	cve := raw.Vulnerabilities[0].CVE
	details := &githubAdvisoryDetails{}

	// Prefer English description; fall back to the first available language.
	var enDesc, anyDesc string
	for _, d := range cve.Descriptions {
		if strings.TrimSpace(d.Value) == "" {
			continue
		}
		if anyDesc == "" {
			anyDesc = strings.TrimSpace(d.Value)
		}
		if strings.EqualFold(d.Lang, "en") {
			enDesc = strings.TrimSpace(d.Value)
			break
		}
	}
	if enDesc != "" {
		details.Description = enDesc
		details.Summary = truncateSummary(enDesc)
	} else if anyDesc != "" {
		details.Description = anyDesc
		details.Summary = truncateSummary(anyDesc)
	}

	score, vector := pickNVDCVSS(cve.Metrics)
	details.CVSSScore = score
	details.CVSSVector = vector

	seenCWE := map[string]struct{}{}
	for _, w := range cve.Weaknesses {
		for _, d := range w.Description {
			id := strings.TrimSpace(d.Value)
			if id == "" || !strings.HasPrefix(strings.ToUpper(id), "CWE-") {
				continue
			}
			if _, ok := seenCWE[id]; ok {
				continue
			}
			seenCWE[id] = struct{}{}
			details.CWEs = append(details.CWEs, githubCWE{ID: id, Name: id})
		}
	}

	seenRef := map[string]struct{}{}
	for _, r := range cve.References {
		u := strings.TrimSpace(r.URL)
		if u == "" {
			continue
		}
		if _, ok := seenRef[u]; ok {
			continue
		}
		seenRef[u] = struct{}{}
		details.References = append(details.References, u)
	}
	// Stable NVD detail page first when not already present.
	nvdPage := "https://nvd.nist.gov/vuln/detail/" + cveID
	if _, ok := seenRef[nvdPage]; !ok {
		details.References = append([]string{nvdPage}, details.References...)
	}

	return details, nil
}

type nvdResponse struct {
	Vulnerabilities []struct {
		CVE nvdCVE `json:"cve"`
	} `json:"vulnerabilities"`
}

type nvdCVE struct {
	ID           string `json:"id"`
	Descriptions []struct {
		Lang  string `json:"lang"`
		Value string `json:"value"`
	} `json:"descriptions"`
	Metrics    nvdMetrics `json:"metrics"`
	Weaknesses []struct {
		Description []struct {
			Lang  string `json:"lang"`
			Value string `json:"value"`
		} `json:"description"`
	} `json:"weaknesses"`
	References []struct {
		URL string `json:"url"`
	} `json:"references"`
}

type nvdMetrics struct {
	CVSSMetricV31 []nvdCVSSMetric `json:"cvssMetricV31"`
	CVSSMetricV30 []nvdCVSSMetric `json:"cvssMetricV30"`
	CVSSMetricV2  []nvdCVSSMetric `json:"cvssMetricV2"`
}

type nvdCVSSMetric struct {
	CVSSData *struct {
		BaseScore    float64 `json:"baseScore"`
		VectorString string  `json:"vectorString"`
	} `json:"cvssData"`
}

func pickNVDCVSS(m nvdMetrics) (*float64, *string) {
	for _, group := range [][]nvdCVSSMetric{m.CVSSMetricV31, m.CVSSMetricV30, m.CVSSMetricV2} {
		for _, metric := range group {
			if metric.CVSSData == nil || metric.CVSSData.BaseScore <= 0 {
				continue
			}
			score := metric.CVSSData.BaseScore
			var vector *string
			if v := strings.TrimSpace(metric.CVSSData.VectorString); v != "" {
				vector = &v
			}
			return &score, vector
		}
	}
	return nil, nil
}

func truncateSummary(text string) string {
	text = strings.TrimSpace(text)
	// Use first sentence-ish chunk as a short summary for list views.
	if i := strings.IndexAny(text, ".!?"); i > 40 && i < 240 {
		return strings.TrimSpace(text[:i+1])
	}
	if len(text) > 200 {
		// Cut on a rune boundary: NVD descriptions are arbitrary UTF-8 (this
		// client even falls back to non-English text), and slicing mid-sequence
		// would store invalid UTF-8 that Postgres rejects, failing enrichment
		// for that CVE outright.
		cut := 197
		for cut > 0 && !utf8.RuneStart(text[cut]) {
			cut--
		}
		return strings.TrimSpace(text[:cut]) + "..."
	}
	return text
}
