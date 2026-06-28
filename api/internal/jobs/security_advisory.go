package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/jackc/pgx/v5/pgtype"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// securityAdvisoryPackages is the set of first-party Shopware packages whose
// advisories are imported from Packagist. Extension/plugin advisories are a
// separate, future concern (user-submitted disclosures, origin != 'packagist').
var securityAdvisoryPackages = []string{
	"shopware/core",
	"shopware/administration",
	"shopware/storefront",
	"shopware/elasticsearch",
}

// packagistSecurityAdvisoriesURL is the Packagist Security Advisories API.
const packagistSecurityAdvisoriesURL = "https://packagist.org/api/security-advisories/"

// packagistReportedAtLayout matches the timestamp format Packagist returns,
// e.g. "2021-09-15 12:00:00".
const packagistReportedAtLayout = "2006-01-02 15:04:05"

// SecurityAdvisorySyncHandler imports security advisories from Packagist.
type SecurityAdvisorySyncHandler struct {
	queries *queries.Queries
	baseURL string
}

func NewSecurityAdvisorySyncHandler(q *queries.Queries) *SecurityAdvisorySyncHandler {
	return &SecurityAdvisorySyncHandler{queries: q, baseURL: packagistSecurityAdvisoriesURL}
}

// packagistAdvisory mirrors a single advisory record from the Packagist response.
type packagistAdvisory struct {
	AdvisoryID       string `json:"advisoryId"`
	PackageName      string `json:"packageName"`
	Title            string `json:"title"`
	Link             string `json:"link"`
	CVE              string `json:"cve"`
	AffectedVersions string `json:"affectedVersions"`
	ReportedAt       string `json:"reportedAt"`
	Severity         string `json:"severity"`
	Sources          []struct {
		Name     string `json:"name"`
		RemoteID string `json:"remoteId"`
	} `json:"sources"`
}

type packagistAdvisoriesResponse struct {
	Advisories map[string][]packagistAdvisory `json:"advisories"`
}

// HandleSync fetches advisories for the tracked Shopware packages from Packagist
// and upserts them into the security_advisory catalog.
func (h *SecurityAdvisorySyncHandler) HandleSync(ctx context.Context, _ SecurityAdvisorySync) error {
	ctx, span := tracer.Start(ctx, "security_advisory.sync")
	defer span.End()

	resp, err := h.fetch(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("fetch security advisories: %w", err)
	}

	count := 0
	for _, advisories := range resp.Advisories {
		for _, adv := range advisories {
			if adv.AdvisoryID == "" {
				continue
			}
			if err := h.queries.UpsertSecurityAdvisory(ctx, toUpsertParams(adv)); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return fmt.Errorf("upsert security advisory %s: %w", adv.AdvisoryID, err)
			}
			count++
		}
	}

	span.SetAttributes(attribute.Int("advisories.count", count))
	slog.Info("synced security advisories", "count", count)
	return nil
}

func (h *SecurityAdvisorySyncHandler) fetch(ctx context.Context) (*packagistAdvisoriesResponse, error) {
	q := url.Values{}
	for _, pkg := range securityAdvisoryPackages {
		q.Add("packages[]", pkg)
	}
	reqURL := h.baseURL + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := httputil.NewHTTPClient(httputil.WithTimeout(30 * time.Second)).Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("packagist returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed packagistAdvisoriesResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("parse packagist response: %w", err)
	}
	return &parsed, nil
}

func toUpsertParams(adv packagistAdvisory) queries.UpsertSecurityAdvisoryParams {
	params := queries.UpsertSecurityAdvisoryParams{
		AdvisoryID:       adv.AdvisoryID,
		Origin:           "packagist",
		PackageName:      adv.PackageName,
		Title:            adv.Title,
		Link:             nilIfEmpty(adv.Link),
		Cve:              nilIfEmpty(adv.CVE),
		AffectedVersions: adv.AffectedVersions,
		Severity:         nilIfEmpty(adv.Severity),
		ReportedAt:       parsePackagistTime(adv.ReportedAt),
	}
	if len(adv.Sources) > 0 {
		params.SourceName = nilIfEmpty(adv.Sources[0].Name)
		params.SourceRemoteID = nilIfEmpty(adv.Sources[0].RemoteID)
	}
	return params
}

func parsePackagistTime(s string) pgtype.Timestamp {
	if s == "" {
		return pgtype.Timestamp{}
	}
	t, err := time.Parse(packagistReportedAtLayout, s)
	if err != nil {
		return pgtype.Timestamp{}
	}
	return pgtype.Timestamp{Time: t, Valid: true}
}
