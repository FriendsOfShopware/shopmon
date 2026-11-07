package api

import "time"

// ComposerAdvisoryResponse follows the Composer security-advisories API format.
type ComposerAdvisoryResponse struct {
	Advisories map[string][]ComposerAdvisory `json:"advisories"`
}

type ComposerAdvisory struct {
	AdvisoryID       string           `json:"advisoryId"`
	PackageName      string           `json:"packageName"`
	AffectedVersions string           `json:"affectedVersions"`
	Title            string           `json:"title"`
	CVE              *string          `json:"cve"`
	Link             *string          `json:"link"`
	ReportedAt       time.Time        `json:"reportedAt"`
	Sources          []AdvisorySource `json:"sources"`
	Severity         *string          `json:"severity"`
}
