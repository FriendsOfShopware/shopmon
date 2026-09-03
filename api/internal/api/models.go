// Package api contains hand-written HTTP DTO types shared by handlers and read models.
package api

import (
	"time"

	"github.com/danielgtaylor/huma/v2"
)

func namedStringEnumSchema(r huma.Registry, name string, values ...string) *huma.Schema {
	m := r.Map()
	if _, ok := m[name]; !ok {
		enum := make([]any, len(values))
		for i, value := range values {
			enum[i] = value
		}
		m[name] = &huma.Schema{Type: huma.TypeString, Enum: enum}
	}
	return &huma.Schema{Ref: "#/components/schemas/" + name}
}

// Defines values for SeverityLevel.
const (
	SeverityLevelCritical SeverityLevel = "critical"
	SeverityLevelHigh     SeverityLevel = "high"
	SeverityLevelLow      SeverityLevel = "low"
	SeverityLevelMedium   SeverityLevel = "medium"
	SeverityLevelNone     SeverityLevel = "none"
)

// Valid indicates whether the value is a known member of the SeverityLevel enum.
func (e SeverityLevel) Valid() bool {
	switch e {
	case SeverityLevelCritical:
		return true
	case SeverityLevelHigh:
		return true
	case SeverityLevelLow:
		return true
	case SeverityLevelMedium:
		return true
	case SeverityLevelNone:
		return true
	default:
		return false
	}
}

func (SeverityLevel) Schema(r huma.Registry) *huma.Schema {
	return namedStringEnumSchema(r, "SeverityLevel", "none", "low", "medium", "high", "critical")
}

// Defines values for AdvisoryScope.
const (
	AdvisoryScopeAffected   AdvisoryScope = "affected"
	AdvisoryScopeAll        AdvisoryScope = "all"
	AdvisoryScopeSuppressed AdvisoryScope = "suppressed"
)

// Valid indicates whether the value is a known member of the AdvisoryScope enum.
func (e AdvisoryScope) Valid() bool {
	switch e {
	case AdvisoryScopeAffected:
		return true
	case AdvisoryScopeAll:
		return true
	case AdvisoryScopeSuppressed:
		return true
	default:
		return false
	}
}

// Defines values for AdvisorySeverity.
const (
	AdvisorySeverityCritical AdvisorySeverity = "critical"
	AdvisorySeverityHigh     AdvisorySeverity = "high"
	AdvisorySeverityLow      AdvisorySeverity = "low"
	AdvisorySeverityMedium   AdvisorySeverity = "medium"
	AdvisorySeverityNone     AdvisorySeverity = "none"
)

// Valid indicates whether the value is a known member of the AdvisorySeverity enum.
func (e AdvisorySeverity) Valid() bool {
	switch e {
	case AdvisorySeverityCritical:
		return true
	case AdvisorySeverityHigh:
		return true
	case AdvisorySeverityLow:
		return true
	case AdvisorySeverityMedium:
		return true
	case AdvisorySeverityNone:
		return true
	default:
		return false
	}
}

// Defines values for AdvisorySort.
const (
	AdvisorySortAffected AdvisorySort = "affected"
	AdvisorySortCvss     AdvisorySort = "cvss"
	AdvisorySortReported AdvisorySort = "reported"
	AdvisorySortSeverity AdvisorySort = "severity"
)

// Valid indicates whether the value is a known member of the AdvisorySort enum.
func (e AdvisorySort) Valid() bool {
	switch e {
	case AdvisorySortAffected:
		return true
	case AdvisorySortCvss:
		return true
	case AdvisorySortReported:
		return true
	case AdvisorySortSeverity:
		return true
	default:
		return false
	}
}

// Defines values for LanguageParam.
const (
	LanguageParamDe LanguageParam = "de"
	LanguageParamEn LanguageParam = "en"
)

// Valid indicates whether the value is a known member of the LanguageParam enum.
func (e LanguageParam) Valid() bool {
	switch e {
	case LanguageParamDe:
		return true
	case LanguageParamEn:
		return true
	default:
		return false
	}
}

// Defines values for GetAccountExtensionsParamsLanguage.
const (
	GetAccountExtensionsParamsLanguageDe GetAccountExtensionsParamsLanguage = "de"
	GetAccountExtensionsParamsLanguageEn GetAccountExtensionsParamsLanguage = "en"
)

// Valid indicates whether the value is a known member of the GetAccountExtensionsParamsLanguage enum.
func (e GetAccountExtensionsParamsLanguage) Valid() bool {
	switch e {
	case GetAccountExtensionsParamsLanguageDe:
		return true
	case GetAccountExtensionsParamsLanguageEn:
		return true
	default:
		return false
	}
}

// Defines values for GetAccountExtensionParamsLanguage.
const (
	GetAccountExtensionParamsLanguageDe GetAccountExtensionParamsLanguage = "de"
	GetAccountExtensionParamsLanguageEn GetAccountExtensionParamsLanguage = "en"
)

// Valid indicates whether the value is a known member of the GetAccountExtensionParamsLanguage enum.
func (e GetAccountExtensionParamsLanguage) Valid() bool {
	switch e {
	case GetAccountExtensionParamsLanguageDe:
		return true
	case GetAccountExtensionParamsLanguageEn:
		return true
	default:
		return false
	}
}

// Defines values for AdminListAdvisoriesParamsSeverity.
const (
	AdminListAdvisoriesParamsSeverityCritical AdminListAdvisoriesParamsSeverity = "critical"
	AdminListAdvisoriesParamsSeverityHigh     AdminListAdvisoriesParamsSeverity = "high"
	AdminListAdvisoriesParamsSeverityLow      AdminListAdvisoriesParamsSeverity = "low"
	AdminListAdvisoriesParamsSeverityMedium   AdminListAdvisoriesParamsSeverity = "medium"
	AdminListAdvisoriesParamsSeverityNone     AdminListAdvisoriesParamsSeverity = "none"
)

// Valid indicates whether the value is a known member of the AdminListAdvisoriesParamsSeverity enum.
func (e AdminListAdvisoriesParamsSeverity) Valid() bool {
	switch e {
	case AdminListAdvisoriesParamsSeverityCritical:
		return true
	case AdminListAdvisoriesParamsSeverityHigh:
		return true
	case AdminListAdvisoriesParamsSeverityLow:
		return true
	case AdminListAdvisoriesParamsSeverityMedium:
		return true
	case AdminListAdvisoriesParamsSeverityNone:
		return true
	default:
		return false
	}
}

// Defines values for AdminGetEnvironmentsParamsSortDirection.
const (
	AdminGetEnvironmentsParamsSortDirectionAsc  AdminGetEnvironmentsParamsSortDirection = "asc"
	AdminGetEnvironmentsParamsSortDirectionDesc AdminGetEnvironmentsParamsSortDirection = "desc"
)

// Valid indicates whether the value is a known member of the AdminGetEnvironmentsParamsSortDirection enum.
func (e AdminGetEnvironmentsParamsSortDirection) Valid() bool {
	switch e {
	case AdminGetEnvironmentsParamsSortDirectionAsc:
		return true
	case AdminGetEnvironmentsParamsSortDirectionDesc:
		return true
	default:
		return false
	}
}

// Defines values for AdminGetOrganizationsParamsSortBy.
const (
	CreatedAt   AdminGetOrganizationsParamsSortBy = "createdAt"
	MemberCount AdminGetOrganizationsParamsSortBy = "memberCount"
	Name        AdminGetOrganizationsParamsSortBy = "name"
	ShopCount   AdminGetOrganizationsParamsSortBy = "shopCount"
)

// Valid indicates whether the value is a known member of the AdminGetOrganizationsParamsSortBy enum.
func (e AdminGetOrganizationsParamsSortBy) Valid() bool {
	switch e {
	case CreatedAt:
		return true
	case MemberCount:
		return true
	case Name:
		return true
	case ShopCount:
		return true
	default:
		return false
	}
}

// Defines values for AdminGetOrganizationsParamsSortDirection.
const (
	AdminGetOrganizationsParamsSortDirectionAsc  AdminGetOrganizationsParamsSortDirection = "asc"
	AdminGetOrganizationsParamsSortDirectionDesc AdminGetOrganizationsParamsSortDirection = "desc"
)

// Valid indicates whether the value is a known member of the AdminGetOrganizationsParamsSortDirection enum.
func (e AdminGetOrganizationsParamsSortDirection) Valid() bool {
	switch e {
	case AdminGetOrganizationsParamsSortDirectionAsc:
		return true
	case AdminGetOrganizationsParamsSortDirectionDesc:
		return true
	default:
		return false
	}
}

// Defines values for ListAdvisoriesParamsSeverity.
const (
	ListAdvisoriesParamsSeverityCritical ListAdvisoriesParamsSeverity = "critical"
	ListAdvisoriesParamsSeverityHigh     ListAdvisoriesParamsSeverity = "high"
	ListAdvisoriesParamsSeverityLow      ListAdvisoriesParamsSeverity = "low"
	ListAdvisoriesParamsSeverityMedium   ListAdvisoriesParamsSeverity = "medium"
	ListAdvisoriesParamsSeverityNone     ListAdvisoriesParamsSeverity = "none"
)

// Valid indicates whether the value is a known member of the ListAdvisoriesParamsSeverity enum.
func (e ListAdvisoriesParamsSeverity) Valid() bool {
	switch e {
	case ListAdvisoriesParamsSeverityCritical:
		return true
	case ListAdvisoriesParamsSeverityHigh:
		return true
	case ListAdvisoriesParamsSeverityLow:
		return true
	case ListAdvisoriesParamsSeverityMedium:
		return true
	case ListAdvisoriesParamsSeverityNone:
		return true
	default:
		return false
	}
}

// Defines values for ListAdvisoriesParamsScope.
const (
	ListAdvisoriesParamsScopeAffected   ListAdvisoriesParamsScope = "affected"
	ListAdvisoriesParamsScopeAll        ListAdvisoriesParamsScope = "all"
	ListAdvisoriesParamsScopeSuppressed ListAdvisoriesParamsScope = "suppressed"
)

// Valid indicates whether the value is a known member of the ListAdvisoriesParamsScope enum.
func (e ListAdvisoriesParamsScope) Valid() bool {
	switch e {
	case ListAdvisoriesParamsScopeAffected:
		return true
	case ListAdvisoriesParamsScopeAll:
		return true
	case ListAdvisoriesParamsScopeSuppressed:
		return true
	default:
		return false
	}
}

// Defines values for ListAdvisoriesParamsSort.
const (
	ListAdvisoriesParamsSortAffected ListAdvisoriesParamsSort = "affected"
	ListAdvisoriesParamsSortCvss     ListAdvisoriesParamsSort = "cvss"
	ListAdvisoriesParamsSortReported ListAdvisoriesParamsSort = "reported"
	ListAdvisoriesParamsSortSeverity ListAdvisoriesParamsSort = "severity"
)

// Valid indicates whether the value is a known member of the ListAdvisoriesParamsSort enum.
func (e ListAdvisoriesParamsSort) Valid() bool {
	switch e {
	case ListAdvisoriesParamsSortAffected:
		return true
	case ListAdvisoriesParamsSortCvss:
		return true
	case ListAdvisoriesParamsSortReported:
		return true
	case ListAdvisoriesParamsSortSeverity:
		return true
	default:
		return false
	}
}

// Defines values for GetEnvironmentParamsLanguage.
const (
	GetEnvironmentParamsLanguageDe GetEnvironmentParamsLanguage = "de"
	GetEnvironmentParamsLanguageEn GetEnvironmentParamsLanguage = "en"
)

// Valid indicates whether the value is a known member of the GetEnvironmentParamsLanguage enum.
func (e GetEnvironmentParamsLanguage) Valid() bool {
	switch e {
	case GetEnvironmentParamsLanguageDe:
		return true
	case GetEnvironmentParamsLanguageEn:
		return true
	default:
		return false
	}
}

// AccountChangelog defines model for AccountChangelog.
type AccountChangelog struct {
	Date                        time.Time `json:"date"`
	EnvironmentId               int       `json:"environmentId"`
	EnvironmentName             string    `json:"environmentName"`
	EnvironmentOrganizationId   string    `json:"environmentOrganizationId"`
	EnvironmentOrganizationName string    `json:"environmentOrganizationName"`
	EnvironmentShopName         string    `json:"environmentShopName"`

	// Extensions JSON data describing extension changes
	Extensions         []ExtensionDiff `json:"extensions"`
	Id                 int             `json:"id"`
	NewShopwareVersion *string         `json:"newShopwareVersion"`
	OldShopwareVersion *string         `json:"oldShopwareVersion"`
}

// AccountEnvironment defines model for AccountEnvironment.
type AccountEnvironment struct {
	Favicon          *string    `json:"favicon"`
	Id               int        `json:"id"`
	LastScrapedAt    *time.Time `json:"lastScrapedAt"`
	LastScrapedError *string    `json:"lastScrapedError"`
	Name             string     `json:"name"`
	OrganizationId   string     `json:"organizationId"`
	OrganizationName string     `json:"organizationName"`
	ShopId           *int       `json:"shopId"`
	ShopName         *string    `json:"shopName"`
	ShopwareVersion  string     `json:"shopwareVersion"`
	Status           string     `json:"status"`
	Url              string     `json:"url"`
}

// AccountExtension defines model for AccountExtension.
type AccountExtension struct {
	Active    bool                       `json:"active"`
	Changelog *[]ExtensionChangelogEntry `json:"changelog"`

	// Description Full store description (HTML) in the requested language (falls back to English).
	Description  *string                       `json:"description,omitempty"`
	Environments []AccountExtensionEnvironment `json:"environments"`

	// IconUrl URL of the extension's store icon, when known.
	IconUrl *string `json:"iconUrl,omitempty"`

	// InstallationManual Installation manual (HTML) in the requested language (falls back to English).
	InstallationManual *string    `json:"installationManual,omitempty"`
	Installed          bool       `json:"installed"`
	InstalledAt        *time.Time `json:"installedAt"`
	Label              string     `json:"label"`
	LatestVersion      string     `json:"latestVersion"`
	Name               string     `json:"name"`
	ProducerName       *string    `json:"producerName,omitempty"`
	ProducerWebsite    *string    `json:"producerWebsite,omitempty"`
	RatingAverage      *float32   `json:"ratingAverage"`

	// ReleaseDate Release date of the latest store version (raw store value).
	ReleaseDate *string                `json:"releaseDate,omitempty"`
	Screenshots *[]ExtensionScreenshot `json:"screenshots,omitempty"`

	// ShortDescription Short store description in the requested language (falls back to English).
	ShortDescription *string `json:"shortDescription,omitempty"`
	StoreLink        *string `json:"storeLink"`
	Version          string  `json:"version"`
}

// AccountExtensionEnvironment defines model for AccountExtensionEnvironment.
type AccountExtensionEnvironment struct {
	Active                      bool   `json:"active"`
	EnvironmentId               int    `json:"environmentId"`
	EnvironmentName             string `json:"environmentName"`
	EnvironmentOrganizationId   string `json:"environmentOrganizationId"`
	EnvironmentOrganizationName string `json:"environmentOrganizationName"`
	EnvironmentShopName         string `json:"environmentShopName"`
	EnvironmentUrl              string `json:"environmentUrl"`
	Installed                   bool   `json:"installed"`
	LatestVersion               string `json:"latestVersion"`

	// ShopId ID of the shop this environment belongs to.
	ShopId int `json:"shopId"`

	// ShopName Name of the shop this environment belongs to (environments may share a name across shops).
	ShopName string `json:"shopName"`
	Version  string `json:"version"`
}

// AccountOrganization defines model for AccountOrganization.
type AccountOrganization struct {
	CreatedAt        time.Time `json:"createdAt"`
	EnvironmentCount int       `json:"environmentCount"`
	Id               string    `json:"id"`
	Logo             *string   `json:"logo"`
	MemberCount      int       `json:"memberCount"`
	Name             string    `json:"name"`
}

// AccountShop defines model for AccountShop.
type AccountShop struct {
	DefaultEnvironmentId *int    `json:"defaultEnvironmentId"`
	Description          *string `json:"description"`
	GitUrl               *string `json:"gitUrl"`
	Id                   int     `json:"id"`
	Name                 string  `json:"name"`
	OrganizationId       string  `json:"organizationId"`
	OrganizationName     string  `json:"organizationName"`
}

// AdminAdvisory defines model for AdminAdvisory.
type AdminAdvisory struct {
	// AdvisoryId Canonical id (CVE, else GHSA, else Packagist PKSA)
	AdvisoryId         string   `json:"advisoryId"`
	AffectedComponents []string `json:"affectedComponents"`

	// AffectedEnvironmentCount Number of the caller's environments whose Composer package inventory matches this advisory and that are not suppressed. Null when no SBOM data has been collected.
	AffectedEnvironmentCount *int           `json:"affectedEnvironmentCount,omitempty"`
	ComposerRepository       *string        `json:"composerRepository,omitempty"`
	CreatedAt                time.Time      `json:"createdAt"`
	Cve                      *string        `json:"cve,omitempty"`
	CvssScore                *float64       `json:"cvssScore,omitempty"`
	CvssVector               *string        `json:"cvssVector,omitempty"`
	Cwes                     *[]AdvisoryCWE `json:"cwes,omitempty"`

	// Description Full advisory write-up in Markdown (source form)
	Description *string `json:"description,omitempty"`

	// DescriptionHtml description rendered as HTML for safe display after client sanitization
	DescriptionHtml *string `json:"descriptionHtml,omitempty"`

	// DetailsSource Where disclosure details were loaded from (e.g. github)
	DetailsSource     *string        `json:"detailsSource,omitempty"`
	EffectiveSeverity *SeverityLevel `json:"effectiveSeverity"`
	EnrichedAt        *time.Time     `json:"enrichedAt,omitempty"`
	EnrichedBy        *string        `json:"enrichedBy,omitempty"`

	// ExternalReferences External reference URLs (GHSA page, commits, etc.)
	ExternalReferences *[]string `json:"externalReferences,omitempty"`

	// FirstPatchedVersions First Shopware release per line that closes this advisory, e.g. {"6.7": "6.7.10.1"}. Machine-derived from the GitHub advisory.
	FirstPatchedVersions *map[string]string `json:"firstPatchedVersions,omitempty"`
	GhsaId               *string            `json:"ghsaId,omitempty"`
	IsVisible            bool               `json:"isVisible"`
	Link                 *string            `json:"link,omitempty"`
	NotesInternal        *string            `json:"notesInternal,omitempty"`
	NotesPublic          *string            `json:"notesPublic,omitempty"`

	// Packages Affected Composer packages (one CVE may span multiple packages)
	Packages           []AdvisoryAffectedPackage `json:"packages"`
	RecommendedUpgrade *string                   `json:"recommendedUpgrade,omitempty"`
	RemediationSummary *string                   `json:"remediationSummary,omitempty"`
	RemediationUrl     *string                   `json:"remediationUrl,omitempty"`
	ReportedAt         *time.Time                `json:"reportedAt,omitempty"`

	// SecurityPluginFixes Which SwagPlatformSecurity releases backport this advisory, per branch. Empty when it is not backportable or not yet derived.
	SecurityPluginFixes   *[]AdvisorySecurityPluginFix `json:"securityPluginFixes,omitempty"`
	Severity              *SeverityLevel               `json:"severity,omitempty"`
	SeverityOverride      *SeverityLevel               `json:"severityOverride,omitempty"`
	ShopwareImpactSummary *string                      `json:"shopwareImpactSummary,omitempty"`
	Sources               []AdvisorySource             `json:"sources"`

	// Summary Short summary from the disclosure source (e.g. GitHub Advisory)
	Summary *string `json:"summary,omitempty"`

	// SuppressedEnvironmentCount Number of the caller's environments where this advisory has been acknowledged via an active suppression.
	SuppressedEnvironmentCount *int      `json:"suppressedEnvironmentCount,omitempty"`
	SyncedAt                   time.Time `json:"syncedAt"`
	Tags                       []string  `json:"tags"`
	Title                      string    `json:"title"`
	UpdatedAt                  time.Time `json:"updatedAt"`
}

// AdminAdvisoryListResponse defines model for AdminAdvisoryListResponse.
type AdminAdvisoryListResponse struct {
	Advisories []AdminAdvisory `json:"advisories"`
	Total      int             `json:"total"`
}

// AdminAdvisorySyncResponse defines model for AdminAdvisorySyncResponse.
type AdminAdvisorySyncResponse struct {
	Enqueued bool `json:"enqueued"`
}

// AdminAuditLogEntry defines model for AdminAuditLogEntry.
type AdminAuditLogEntry struct {
	Action       string    `json:"action"`
	ActorEmail   *string   `json:"actorEmail,omitempty"`
	ActorName    *string   `json:"actorName,omitempty"`
	ActorUserId  *string   `json:"actorUserId,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	Detail       *string   `json:"detail,omitempty"`
	Id           int64     `json:"id"`
	IpAddress    *string   `json:"ipAddress,omitempty"`
	TargetEmail  *string   `json:"targetEmail,omitempty"`
	TargetName   *string   `json:"targetName,omitempty"`
	TargetUserId *string   `json:"targetUserId,omitempty"`
}

// AdminAuditLogResponse defines model for AdminAuditLogResponse.
type AdminAuditLogResponse struct {
	Entries []AdminAuditLogEntry `json:"entries"`
	Total   int                  `json:"total"`
}

// AdminEnvironmentCheck defines model for AdminEnvironmentCheck.
type AdminEnvironmentCheck struct {
	CheckId string  `json:"checkId"`
	Id      int     `json:"id"`
	Level   string  `json:"level"`
	Link    *string `json:"link,omitempty"`
	Message string  `json:"message"`
	Source  string  `json:"source"`
}

// AdminEnvironmentDeployment defines model for AdminEnvironmentDeployment.
type AdminEnvironmentDeployment struct {
	Command       string    `json:"command"`
	CreatedAt     time.Time `json:"createdAt"`
	ExecutionTime float32   `json:"executionTime"`
	Id            int       `json:"id"`
	Name          string    `json:"name"`
	Reference     *string   `json:"reference,omitempty"`
	ReturnCode    int       `json:"returnCode"`
}

// AdminEnvironmentDetail defines model for AdminEnvironmentDetail.
type AdminEnvironmentDetail struct {
	Checks               []AdminEnvironmentCheck     `json:"checks"`
	ConnectionIssueCount int                         `json:"connectionIssueCount"`
	CreatedAt            time.Time                   `json:"createdAt"`
	Extensions           []AdminEnvironmentExtension `json:"extensions"`
	Id                   int                         `json:"id"`
	LastDeployment       *AdminEnvironmentDeployment `json:"lastDeployment,omitempty"`
	LastScrapedAt        *time.Time                  `json:"lastScrapedAt,omitempty"`
	LastScrapedError     *string                     `json:"lastScrapedError,omitempty"`
	Name                 string                      `json:"name"`
	OrganizationId       string                      `json:"organizationId"`
	OrganizationName     string                      `json:"organizationName"`
	ScheduledTasks       []AdminEnvironmentTask      `json:"scheduledTasks"`
	ShopId               int                         `json:"shopId"`
	ShopName             string                      `json:"shopName"`
	ShopwareVersion      string                      `json:"shopwareVersion"`
	Status               string                      `json:"status"`
	Url                  string                      `json:"url"`
}

// AdminEnvironmentExtension defines model for AdminEnvironmentExtension.
type AdminEnvironmentExtension struct {
	Active        bool    `json:"active"`
	Id            int     `json:"id"`
	Installed     bool    `json:"installed"`
	Label         string  `json:"label"`
	LatestVersion *string `json:"latestVersion,omitempty"`
	Name          string  `json:"name"`
	StoreLink     *string `json:"storeLink,omitempty"`
	Version       string  `json:"version"`
}

// AdminEnvironmentTask defines model for AdminEnvironmentTask.
type AdminEnvironmentTask struct {
	Id                int     `json:"id"`
	Interval          int     `json:"interval"`
	LastExecutionTime *string `json:"lastExecutionTime,omitempty"`
	Name              string  `json:"name"`
	NextExecutionTime *string `json:"nextExecutionTime,omitempty"`
	Overdue           bool    `json:"overdue"`
	Status            string  `json:"status"`
	TaskId            string  `json:"taskId"`
}

// AdminEnvironmentsResponse defines model for AdminEnvironmentsResponse.
type AdminEnvironmentsResponse struct {
	Environments []AccountEnvironment `json:"environments"`
	Total        int                  `json:"total"`
}

// AdminGrowth defines model for AdminGrowth.
type AdminGrowth struct {
	Environments []GrowthDataPoint `json:"environments"`
	Users        []GrowthDataPoint `json:"users"`
}

// AdminOrganizationDetail defines model for AdminOrganizationDetail.
type AdminOrganizationDetail struct {
	CreatedAt        time.Time                      `json:"createdAt"`
	EnvironmentCount int                            `json:"environmentCount"`
	Environments     []AdminOrganizationEnvironment `json:"environments"`
	Id               string                         `json:"id"`
	Invitations      []AdminOrganizationInvitation  `json:"invitations"`
	Logo             *string                        `json:"logo,omitempty"`
	MemberCount      int                            `json:"memberCount"`
	Members          []AdminOrganizationMember      `json:"members"`
	Name             string                         `json:"name"`
	Shops            []AdminOrganizationShop        `json:"shops"`
	Slug             string                         `json:"slug"`
	SsoProviders     []AdminOrganizationSSO         `json:"ssoProviders"`
}

// AdminOrganizationEnvironment defines model for AdminOrganizationEnvironment.
type AdminOrganizationEnvironment struct {
	Id              int        `json:"id"`
	LastScrapedAt   *time.Time `json:"lastScrapedAt,omitempty"`
	Name            string     `json:"name"`
	ShopId          int        `json:"shopId"`
	ShopName        string     `json:"shopName"`
	ShopwareVersion string     `json:"shopwareVersion"`
	Status          string     `json:"status"`
	Url             string     `json:"url"`
}

// AdminOrganizationInvitation defines model for AdminOrganizationInvitation.
type AdminOrganizationInvitation struct {
	CreatedAt    time.Time `json:"createdAt"`
	Email        string    `json:"email"`
	ExpiresAt    time.Time `json:"expiresAt"`
	Id           string    `json:"id"`
	InviterEmail string    `json:"inviterEmail"`
	InviterName  string    `json:"inviterName"`
	Role         *string   `json:"role,omitempty"`
	Status       string    `json:"status"`
}

// AdminOrganizationMember defines model for AdminOrganizationMember.
type AdminOrganizationMember struct {
	CreatedAt time.Time `json:"createdAt"`
	Email     string    `json:"email"`
	Image     *string   `json:"image,omitempty"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	UserId    string    `json:"userId"`
}

// AdminOrganizationSSO defines model for AdminOrganizationSSO.
type AdminOrganizationSSO struct {
	Domain     string `json:"domain"`
	Id         string `json:"id"`
	Issuer     string `json:"issuer"`
	ProviderId string `json:"providerId"`
}

// AdminOrganizationShop defines model for AdminOrganizationShop.
type AdminOrganizationShop struct {
	CreatedAt            time.Time `json:"createdAt"`
	DefaultEnvironmentId *int      `json:"defaultEnvironmentId,omitempty"`
	Description          *string   `json:"description,omitempty"`
	Id                   int       `json:"id"`
	Name                 string    `json:"name"`
}

// AdminOrganizationsResponse defines model for AdminOrganizationsResponse.
type AdminOrganizationsResponse struct {
	Organizations []AccountOrganization `json:"organizations"`
	Total         int                   `json:"total"`
}

// AdminRecentActivity defines model for AdminRecentActivity.
type AdminRecentActivity struct {
	RecentEnvironments []AccountEnvironment `json:"recentEnvironments"`
	RecentUsers        []UserProfile        `json:"recentUsers"`
}

// AdminStats defines model for AdminStats.
type AdminStats struct {
	EnvironmentsByStatus struct {
		Green  int `json:"green"`
		Red    int `json:"red"`
		Yellow int `json:"yellow"`
	} `json:"environmentsByStatus"`
	TotalEnvironments  int `json:"totalEnvironments"`
	TotalOrganizations int `json:"totalOrganizations"`
	TotalUsers         int `json:"totalUsers"`
}

// Advisory defines model for Advisory.
type Advisory struct {
	// AdvisoryId Canonical id (CVE, else GHSA, else Packagist PKSA)
	AdvisoryId         string   `json:"advisoryId"`
	AffectedComponents []string `json:"affectedComponents"`

	// AffectedEnvironmentCount Number of the caller's environments whose Composer package inventory matches this advisory and that are not suppressed. Null when no SBOM data has been collected.
	AffectedEnvironmentCount *int           `json:"affectedEnvironmentCount,omitempty"`
	ComposerRepository       *string        `json:"composerRepository,omitempty"`
	CreatedAt                time.Time      `json:"createdAt"`
	Cve                      *string        `json:"cve,omitempty"`
	CvssScore                *float64       `json:"cvssScore,omitempty"`
	CvssVector               *string        `json:"cvssVector,omitempty"`
	Cwes                     *[]AdvisoryCWE `json:"cwes,omitempty"`

	// Description Full advisory write-up in Markdown (source form)
	Description *string `json:"description,omitempty"`

	// DescriptionHtml description rendered as HTML for safe display after client sanitization
	DescriptionHtml *string `json:"descriptionHtml,omitempty"`

	// DetailsSource Where disclosure details were loaded from (e.g. github)
	DetailsSource     *string        `json:"detailsSource,omitempty"`
	EffectiveSeverity *SeverityLevel `json:"effectiveSeverity"`

	// ExternalReferences External reference URLs (GHSA page, commits, etc.)
	ExternalReferences *[]string `json:"externalReferences,omitempty"`

	// FirstPatchedVersions First Shopware release per line that closes this advisory, e.g. {"6.7": "6.7.10.1"}. Machine-derived from the GitHub advisory.
	FirstPatchedVersions *map[string]string `json:"firstPatchedVersions,omitempty"`
	GhsaId               *string            `json:"ghsaId,omitempty"`
	IsVisible            bool               `json:"isVisible"`
	Link                 *string            `json:"link,omitempty"`
	NotesPublic          *string            `json:"notesPublic,omitempty"`

	// Packages Affected Composer packages (one CVE may span multiple packages)
	Packages           []AdvisoryAffectedPackage `json:"packages"`
	RecommendedUpgrade *string                   `json:"recommendedUpgrade,omitempty"`
	RemediationSummary *string                   `json:"remediationSummary,omitempty"`
	RemediationUrl     *string                   `json:"remediationUrl,omitempty"`
	ReportedAt         *time.Time                `json:"reportedAt,omitempty"`

	// SecurityPluginFixes Which SwagPlatformSecurity releases backport this advisory, per branch. Empty when it is not backportable or not yet derived.
	SecurityPluginFixes   *[]AdvisorySecurityPluginFix `json:"securityPluginFixes,omitempty"`
	Severity              *SeverityLevel               `json:"severity,omitempty"`
	SeverityOverride      *SeverityLevel               `json:"severityOverride,omitempty"`
	ShopwareImpactSummary *string                      `json:"shopwareImpactSummary,omitempty"`
	Sources               []AdvisorySource             `json:"sources"`

	// Summary Short summary from the disclosure source (e.g. GitHub Advisory)
	Summary *string `json:"summary,omitempty"`

	// SuppressedEnvironmentCount Number of the caller's environments where this advisory has been acknowledged via an active suppression.
	SuppressedEnvironmentCount *int      `json:"suppressedEnvironmentCount,omitempty"`
	SyncedAt                   time.Time `json:"syncedAt"`
	Tags                       []string  `json:"tags"`
	Title                      string    `json:"title"`
	UpdatedAt                  time.Time `json:"updatedAt"`
}

// AdvisoryAffectedEnvironment defines model for AdvisoryAffectedEnvironment.
type AdvisoryAffectedEnvironment struct {
	// AffectedVersions The advisory constraint the installed version falls into
	AffectedVersions  string    `json:"affectedVersions"`
	EnvironmentId     int       `json:"environmentId"`
	EnvironmentName   string    `json:"environmentName"`
	EnvironmentStatus string    `json:"environmentStatus"`
	InstalledVersion  string    `json:"installedVersion"`
	MatchedAt         time.Time `json:"matchedAt"`
	OrganizationId    string    `json:"organizationId"`
	OrganizationName  string    `json:"organizationName"`

	// PackageName Installed Composer package that matched the advisory
	PackageName     string  `json:"packageName"`
	ShopId          int     `json:"shopId"`
	ShopName        string  `json:"shopName"`
	ShopwareVersion *string `json:"shopwareVersion,omitempty"`

	// Suppressed Whether an active suppression covers this environment
	Suppressed bool `json:"suppressed"`
}

// AdvisoryAffectedPackage defines model for AdvisoryAffectedPackage.
type AdvisoryAffectedPackage struct {
	// AffectedVersions Composer constraint for this package
	AffectedVersions string `json:"affectedVersions"`

	// PackageName Composer package name (e.g. shopware/core)
	PackageName string `json:"packageName"`

	// PackagistAdvisoryId Original Packagist PKSA id for this package row
	PackagistAdvisoryId string `json:"packagistAdvisoryId"`
}

// AdvisoryAffectedResponse defines model for AdvisoryAffectedResponse.
type AdvisoryAffectedResponse struct {
	// Environments Affected environments within the caller's organizations
	Environments []AdvisoryAffectedEnvironment `json:"environments"`

	// GlobalTotal Fleet-wide affected environment count; admins only
	GlobalTotal *int `json:"globalTotal,omitempty"`

	// Total Number of affected environments visible to the caller
	Total int `json:"total"`
}

// AdvisoryCWE defines model for AdvisoryCWE.
type AdvisoryCWE struct {
	// Id CWE identifier, e.g. CWE-918
	Id   string `json:"id"`
	Name string `json:"name"`
}

// AdvisoryListResponse defines model for AdvisoryListResponse.
type AdvisoryListResponse struct {
	Advisories []Advisory `json:"advisories"`

	// ScopeCounts Totals per scope, ignoring the scope filter, for tab badges
	ScopeCounts *AdvisoryScopeCounts `json:"scopeCounts,omitempty"`

	// Total Advisories matching the current scope and filters
	Total int `json:"total"`
}

// AdvisoryScopeCounts defines model for AdvisoryScopeCounts.
type AdvisoryScopeCounts struct {
	// Affected Of those, how many affect the caller's environments and are not suppressed
	Affected int `json:"affected"`

	// All All visible advisories matching the non-scope filters
	All int `json:"all"`

	// Suppressed Of those, how many the caller has acknowledged
	Suppressed int `json:"suppressed"`
}

// AdvisorySecurityPluginFix defines model for AdvisorySecurityPluginFix.
type AdvisorySecurityPluginFix struct {
	// PluginBranch SwagPlatformSecurity major version, e.g. "4"
	PluginBranch string `json:"pluginBranch"`

	// PluginVersion Lowest plugin version on this branch that backports the fix
	PluginVersion string `json:"pluginVersion"`

	// ShopwareBranch Shopware line this branch serves, e.g. "6.7"
	ShopwareBranch string `json:"shopwareBranch"`
}

// AdvisorySource defines model for AdvisorySource.
type AdvisorySource struct {
	Name     string `json:"name"`
	RemoteId string `json:"remoteId"`
}

// AdvisorySuppression defines model for AdvisorySuppression.
type AdvisorySuppression struct {
	AdvisoryId    string    `json:"advisoryId"`
	CreatedAt     time.Time `json:"createdAt"`
	CreatedBy     *string   `json:"createdBy,omitempty"`
	CreatedByName *string   `json:"createdByName,omitempty"`

	// EnvironmentId Null means every environment of the shop
	EnvironmentId   *int    `json:"environmentId,omitempty"`
	EnvironmentName *string `json:"environmentName,omitempty"`

	// ExpiresAt Null means the suppression does not expire
	ExpiresAt      *time.Time `json:"expiresAt,omitempty"`
	Id             int64      `json:"id"`
	OrganizationId string     `json:"organizationId"`
	Reason         string     `json:"reason"`
	RevokedAt      *time.Time `json:"revokedAt,omitempty"`
	ShopId         int        `json:"shopId"`
	ShopName       string     `json:"shopName"`
}

// AdvisorySuppressionListResponse defines model for AdvisorySuppressionListResponse.
type AdvisorySuppressionListResponse struct {
	Suppressions []AdvisorySuppression `json:"suppressions"`
}

// ApiKey defines model for ApiKey.
type ApiKey struct {
	CreatedAt  time.Time  `json:"createdAt"`
	Id         string     `json:"id"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
	Name       string     `json:"name"`
	Scopes     []string   `json:"scopes"`
}

// ApiKeyScope defines model for ApiKeyScope.
type ApiKeyScope struct {
	Description string `json:"description"`
	Label       string `json:"label"`
	Value       string `json:"value"`
}

// CacheInfo defines model for CacheInfo.
type CacheInfo struct {
	CacheAdapter *string `json:"cacheAdapter,omitempty"`
	Environment  *string `json:"environment,omitempty"`
	HttpCache    *bool   `json:"httpCache,omitempty"`
	Id           *int    `json:"id,omitempty"`
}

// CreateAdvisorySuppressionRequest defines model for CreateAdvisorySuppressionRequest.
type CreateAdvisorySuppressionRequest struct {
	// EnvironmentId Omit to suppress across every environment of the shop
	EnvironmentId *int `json:"environmentId,omitempty"`

	// ExpiresAt Omit for a suppression that does not expire
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`

	// Reason Why the advisory is accepted or how it is mitigated
	Reason string `json:"reason"`
	ShopId int    `json:"shopId"`
}

// CreateApiKeyRequest defines model for CreateApiKeyRequest.
type CreateApiKeyRequest struct {
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

// CreateApiKeyResponse defines model for CreateApiKeyResponse.
type CreateApiKeyResponse struct {
	Id     string   `json:"id"`
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
	Token  string   `json:"token"`
}

// CreateCliDeploymentRequest defines model for CreateCliDeploymentRequest.
type CreateCliDeploymentRequest struct {
	Command       string    `json:"command"`
	Composer      *string   `json:"composer,omitempty"`
	EndDate       time.Time `json:"end_date"`
	EnvironmentId int       `json:"environment_id"`
	ExecutionTime float32   `json:"execution_time"`
	Name          *string   `json:"name,omitempty"`
	Reference     *string   `json:"reference,omitempty"`
	ReturnCode    int       `json:"return_code"`
	StartDate     time.Time `json:"start_date"`
}

// CreateCliDeploymentResponse defines model for CreateCliDeploymentResponse.
type CreateCliDeploymentResponse struct {
	DeploymentId int    `json:"deployment_id"`
	Name         string `json:"name"`
	Success      bool   `json:"success"`
	UploadUrl    string `json:"upload_url"`
	Url          string `json:"url"`
}

// CreateEnvironmentRequest defines model for CreateEnvironmentRequest.
type CreateEnvironmentRequest struct {
	ClientId         string  `json:"clientId"`
	ClientSecret     string  `json:"clientSecret"`
	EnvironmentToken *string `json:"environmentToken,omitempty"`
	Name             string  `json:"name"`
	ShopId           int     `json:"shopId"`
	ShopUrl          string  `json:"shopUrl"`
}

// CreatePackagesTokenRequest defines model for CreatePackagesTokenRequest.
type CreatePackagesTokenRequest struct {
	Token string `json:"token"`
}

// CreateShopRequest defines model for CreateShopRequest.
type CreateShopRequest struct {
	ClientId        string  `json:"clientId"`
	ClientSecret    string  `json:"clientSecret"`
	Description     *string `json:"description,omitempty"`
	EnvironmentName string  `json:"environmentName"`
	EnvironmentUrl  string  `json:"environmentUrl"`
	GitUrl          *string `json:"gitUrl,omitempty"`
	Name            string  `json:"name"`
}

// Deployment defines model for Deployment.
type Deployment struct {
	Command   string    `json:"command"`
	CreatedAt time.Time `json:"createdAt"`
	EndDate   time.Time `json:"endDate"`

	// ExecutionTime Execution time in seconds
	ExecutionTime float32   `json:"executionTime"`
	Id            int       `json:"id"`
	Name          *string   `json:"name"`
	Reference     *string   `json:"reference"`
	ReturnCode    int       `json:"returnCode"`
	StartDate     time.Time `json:"startDate"`
}

// DeploymentDetail defines model for DeploymentDetail.
type DeploymentDetail struct {
	Command   string    `json:"command"`
	CreatedAt time.Time `json:"createdAt"`
	EndDate   time.Time `json:"endDate"`

	// ExecutionTime Execution time in seconds
	ExecutionTime float32   `json:"executionTime"`
	Id            int       `json:"id"`
	Name          *string   `json:"name"`
	Output        *string   `json:"output,omitempty"`
	Reference     *string   `json:"reference"`
	ReturnCode    int       `json:"returnCode"`
	StartDate     time.Time `json:"startDate"`
}

// EcosystemStats defines model for EcosystemStats.
type EcosystemStats struct {
	Growth           AdminGrowth            `json:"growth"`
	ShopwareVersions []ShopwareVersionCount `json:"shopwareVersions"`
}

// EnvironmentChangelogsResponse defines model for EnvironmentChangelogsResponse.
type EnvironmentChangelogsResponse struct {
	Entries []AccountChangelog `json:"entries"`
	Total   int                `json:"total"`
}

// EnvironmentCheck defines model for EnvironmentCheck.
type EnvironmentCheck struct {
	Id      string  `json:"id"`
	Level   string  `json:"level"`
	Link    *string `json:"link,omitempty"`
	Message string  `json:"message"`

	// MessageKey Translation key for the message; render with params client-side. Falls back to message.
	MessageKey *string `json:"messageKey,omitempty"`

	// Params Structured params for interpolating messageKey.
	Params *map[string]any `json:"params,omitempty"`
}

// EnvironmentDetail defines model for EnvironmentDetail.
type EnvironmentDetail struct {
	Cache *CacheInfo `json:"cache"`

	// ChangelogsCount Total number of recorded changelog entries. The entries themselves are paginated via GET /environments/{environmentId}/changelogs.
	ChangelogsCount    int                    `json:"changelogsCount"`
	Checks             []EnvironmentCheck     `json:"checks"`
	CreatedAt          time.Time              `json:"createdAt"`
	DeploymentsCount   int                    `json:"deploymentsCount"`
	EnvironmentImage   *string                `json:"environmentImage"`
	EnvironmentToken   string                 `json:"environmentToken"`
	Extensions         []EnvironmentExtension `json:"extensions"`
	Favicon            *string                `json:"favicon"`
	Id                 int                    `json:"id"`
	Ignores            *[]string              `json:"ignores"`
	LastChangelog      *AccountChangelog      `json:"lastChangelog"`
	LastScrapedAt      *time.Time             `json:"lastScrapedAt"`
	LastScrapedError   *string                `json:"lastScrapedError"`
	Name               string                 `json:"name"`
	OrganizationId     string                 `json:"organizationId"`
	OrganizationName   string                 `json:"organizationName"`
	Queues             []Queue                `json:"queues"`
	ScheduledTasks     []ScheduledTask        `json:"scheduledTasks"`
	ShopDescription    *string                `json:"shopDescription"`
	ShopId             *int                   `json:"shopId"`
	ShopName           *string                `json:"shopName"`
	ShopwareVersion    string                 `json:"shopwareVersion"`
	SitespeedDetailUrl *string                `json:"sitespeedDetailUrl"`
	SitespeedEnabled   bool                   `json:"sitespeedEnabled"`
	SitespeedUrls      *[]string              `json:"sitespeedUrls"`
	Sitespeeds         []Sitespeed            `json:"sitespeeds"`
	Status             string                 `json:"status"`
	Subscribed         bool                   `json:"subscribed"`
	Url                string                 `json:"url"`
	TaskGraceMinutes   int                    `json:"taskGraceMinutes"`
}

// EnvironmentExtension defines model for EnvironmentExtension.
type EnvironmentExtension struct {
	Active        bool                       `json:"active"`
	Changelog     *[]ExtensionChangelogEntry `json:"changelog,omitempty"`
	Installed     bool                       `json:"installed"`
	InstalledAt   *time.Time                 `json:"installedAt,omitempty"`
	Label         string                     `json:"label"`
	LatestVersion string                     `json:"latestVersion"`
	Name          string                     `json:"name"`
	RatingAverage *float32                   `json:"ratingAverage,omitempty"`
	StoreLink     *string                    `json:"storeLink,omitempty"`
	Version       string                     `json:"version"`
}

// ErrorResponse defines model for ErrorResponse.
type ErrorResponse struct {
	Message string `json:"message"`
}

// ExtensionChangelogEntry defines model for ExtensionChangelogEntry.
type ExtensionChangelogEntry struct {
	CreationDate time.Time `json:"creationDate"`

	// Text Changelog text. For store catalog endpoints this is already resolved to the requested language (English fallback). For stored environment-changelog history it holds the English (en_GB) text, with textDe carrying the German variant when available.
	Text string `json:"text"`

	// TextDe German (de_DE) changelog text for stored history entries, when available.
	TextDe  *string `json:"textDe,omitempty"`
	Version string  `json:"version"`
}

// ExtensionCompatibilityRequest defines model for ExtensionCompatibilityRequest.
type ExtensionCompatibilityRequest struct {
	CurrentVersion string `json:"currentVersion"`
	Extensions     []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"extensions"`
	FutureVersion string `json:"futureVersion"`
}

// ExtensionCompatibilityResult defines model for ExtensionCompatibilityResult.
type ExtensionCompatibilityResult struct {
	IconPath *string `json:"iconPath"`
	Label    string  `json:"label"`
	Name     string  `json:"name"`
	Status   struct {
		Label string `json:"label"`
		Name  string `json:"name"`
		Type  string `json:"type"`
	} `json:"status"`
}

// ExtensionDiff defines model for ExtensionDiff.
type ExtensionDiff struct {
	Active     bool                       `json:"active"`
	Changelog  *[]ExtensionChangelogEntry `json:"changelog,omitempty"`
	Label      string                     `json:"label"`
	Name       string                     `json:"name"`
	NewVersion *string                    `json:"newVersion,omitempty"`
	OldVersion *string                    `json:"oldVersion,omitempty"`
	State      string                     `json:"state"`
}

// ExtensionScreenshot defines model for ExtensionScreenshot.
type ExtensionScreenshot struct {
	// Preview Whether this image is the store's primary preview image.
	Preview bool   `json:"preview"`
	Url     string `json:"url"`
}

// GrowthDataPoint defines model for GrowthDataPoint.
type GrowthDataPoint struct {
	Count int    `json:"count"`
	Month string `json:"month"`
}

// InstanceConfig defines model for InstanceConfig.
type InstanceConfig struct {
	GithubAuthEnabled    bool `json:"githubAuthEnabled"`
	PackageMirrorEnabled bool `json:"packageMirrorEnabled"`
	RegistrationEnabled  bool `json:"registrationEnabled"`
	SitespeedEnabled     bool `json:"sitespeedEnabled"`
}

// Notification defines model for Notification.
type Notification struct {
	CreatedAt time.Time         `json:"createdAt"`
	Id        int               `json:"id"`
	Key       string            `json:"key"`
	Level     string            `json:"level"`
	Link      *NotificationLink `json:"link"`
	Message   string            `json:"message"`

	// MessageKey Translation key for the message; render with params client-side. Falls back to message.
	MessageKey *string `json:"messageKey,omitempty"`

	// Params Structured params for interpolating titleKey/messageKey.
	Params *map[string]any `json:"params,omitempty"`
	Read   bool            `json:"read"`
	Title  string          `json:"title"`

	// TitleKey Translation key for the title; render with params client-side. Falls back to title.
	TitleKey *string `json:"titleKey,omitempty"`
	UserId   string  `json:"userId"`
}

// NotificationEventType defines model for NotificationEventType.
type NotificationEventType struct {
	// DefaultChannels Channels this event is delivered on by default.
	DefaultChannels []string `json:"defaultChannels"`

	// Type Stable event type identifier (e.g. status_degraded).
	Type string `json:"type"`
}

// NotificationLink defines model for NotificationLink.
type NotificationLink struct {
	Label string `json:"label"`
	Url   string `json:"url"`
}

// NotificationPreference defines model for NotificationPreference.
type NotificationPreference struct {
	// Channel Empty is a subscription marker; otherwise in_app or email.
	Channel string `json:"channel"`
	Enabled bool   `json:"enabled"`

	// EventType Empty matches all event types; otherwise a specific event type.
	EventType string `json:"eventType"`

	// ScopeId Empty for global; organization or environment id otherwise.
	ScopeId string `json:"scopeId"`

	// ScopeType Scope of the preference: global, organization, or environment.
	ScopeType string `json:"scopeType"`
}

// NotificationPreferenceInput defines model for NotificationPreferenceInput.
type NotificationPreferenceInput struct {
	Channel   string  `json:"channel"`
	Enabled   bool    `json:"enabled"`
	EventType *string `json:"eventType,omitempty"`
	ScopeId   *string `json:"scopeId,omitempty"`
	ScopeType string  `json:"scopeType"`
}

// PackagesToken defines model for PackagesToken.
type PackagesToken struct {
	Id int `json:"id"`

	// LastSyncedAt Unix timestamp in seconds of the last successful sync, or null if never synced.
	LastSyncedAt *int64 `json:"lastSyncedAt"`
	Source       string `json:"source"`
}

// PackagesTokenConfiguration defines model for PackagesTokenConfiguration.
type PackagesTokenConfiguration struct {
	ComposerUrl *string `json:"composerUrl"`
	Configured  bool    `json:"configured"`
}

// Queue defines model for Queue.
type Queue struct {
	Name string `json:"name"`
	Size int    `json:"size"`
}

// ScheduledTask defines model for ScheduledTask.
type ScheduledTask struct {
	Id                string     `json:"id"`
	LastExecutionTime *time.Time `json:"lastExecutionTime,omitempty"`
	Name              string     `json:"name"`
	NextExecutionTime *time.Time `json:"nextExecutionTime,omitempty"`
	Overdue           bool       `json:"overdue"`
	RunInterval       int        `json:"runInterval"`
	Status            string     `json:"status"`
}

// SeverityLevel defines model for SeverityLevel.
type SeverityLevel string

// Shop defines model for Shop.
type Shop struct {
	DefaultEnvironmentId *int    `json:"defaultEnvironmentId"`
	Description          *string `json:"description"`
	GitUrl               *string `json:"gitUrl"`
	Id                   int     `json:"id"`
	Name                 string  `json:"name"`
	OrganizationId       string  `json:"organizationId"`
}

// ShopwareVersion defines model for ShopwareVersion.
type ShopwareVersion struct {
	Name string `json:"name"`
}

// ShopwareVersionCount defines model for ShopwareVersionCount.
type ShopwareVersionCount struct {
	Count   int    `json:"count"`
	Version string `json:"version"`
}

// Sitespeed defines model for Sitespeed.
type Sitespeed struct {
	CreatedAt              time.Time            `json:"createdAt"`
	CumulativeLayoutShift  *float32             `json:"cumulativeLayoutShift,omitempty"`
	Deployment             *SitespeedDeployment `json:"deployment,omitempty"`
	FirstContentfulPaint   *float32             `json:"firstContentfulPaint,omitempty"`
	FullyLoaded            *float32             `json:"fullyLoaded,omitempty"`
	LargestContentfulPaint *float32             `json:"largestContentfulPaint,omitempty"`
	TransferSize           *float32             `json:"transferSize,omitempty"`
	Ttfb                   *float32             `json:"ttfb,omitempty"`
}

// SitespeedDeployment defines model for SitespeedDeployment.
type SitespeedDeployment struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// SitespeedSettingsRequest defines model for SitespeedSettingsRequest.
type SitespeedSettingsRequest struct {
	Enabled bool      `json:"enabled"`
	Urls    *[]string `json:"urls,omitempty"`
}

// SsoDiscovery defines model for SsoDiscovery.
type SsoDiscovery struct {
	AuthorizationEndpoint string   `json:"authorizationEndpoint"`
	Issuer                string   `json:"issuer"`
	JwksEndpoint          string   `json:"jwksEndpoint"`
	Scopes                []string `json:"scopes"`
	TokenEndpoint         string   `json:"tokenEndpoint"`
	UserInfoEndpoint      string   `json:"userInfoEndpoint"`
}

// SsoProvider defines model for SsoProvider.
type SsoProvider struct {
	AuthorizationEndpoint string `json:"authorizationEndpoint"`
	ClientId              string `json:"clientId"`
	Domain                string `json:"domain"`
	Id                    string `json:"id"`
	Issuer                string `json:"issuer"`
	JwksEndpoint          string `json:"jwksEndpoint"`
	TokenEndpoint         string `json:"tokenEndpoint"`
}

// StatusEvent defines model for StatusEvent.
type StatusEvent struct {
	CreatedAt time.Time      `json:"createdAt"`
	Id        int            `json:"id"`
	NewStatus string         `json:"newStatus"`
	OldStatus string         `json:"oldStatus"`
	Reasons   []StatusReason `json:"reasons"`
}

// StatusReason defines model for StatusReason.
type StatusReason struct {
	Level      string          `json:"level"`
	MessageKey string          `json:"messageKey"`
	Params     *map[string]any `json:"params,omitempty"`
	Source     *string         `json:"source,omitempty"`
}

// SubscribedEnvironment defines model for SubscribedEnvironment.
type SubscribedEnvironment struct {
	Id   int    `json:"id"`
	Name string `json:"name"`

	// ShopId The shop this environment belongs to, if any.
	ShopId   *int    `json:"shopId,omitempty"`
	ShopName *string `json:"shopName,omitempty"`
}

// UpdateAdvisoryEnrichmentRequest defines model for UpdateAdvisoryEnrichmentRequest.
type UpdateAdvisoryEnrichmentRequest struct {
	AffectedComponents    *[]string `json:"affectedComponents,omitempty"`
	IsVisible             *bool     `json:"isVisible,omitempty"`
	NotesInternal         *string   `json:"notesInternal,omitempty"`
	NotesPublic           *string   `json:"notesPublic,omitempty"`
	RecommendedUpgrade    *string   `json:"recommendedUpgrade,omitempty"`
	RemediationSummary    *string   `json:"remediationSummary,omitempty"`
	RemediationUrl        *string   `json:"remediationUrl,omitempty"`
	SeverityOverride      *string   `json:"severityOverride,omitempty" enum:"none,low,medium,high,critical" nullable:"true"`
	ShopwareImpactSummary *string   `json:"shopwareImpactSummary,omitempty"`
	Tags                  *[]string `json:"tags,omitempty"`
}

// UpdateEnvironmentRequest defines model for UpdateEnvironmentRequest.
type UpdateEnvironmentRequest struct {
	ClientId     *string   `json:"clientId,omitempty"`
	ClientSecret *string   `json:"clientSecret,omitempty"`
	Ignores      *[]string `json:"ignores,omitempty"`
	Name         *string   `json:"name,omitempty"`
	ShopId       int       `json:"shopId"`
	ShopUrl      *string   `json:"shopUrl,omitempty"`
	TaskGraceMinutes *int `json:"taskGraceMinutes,omitempty" minimum:"0"`
}

// UpdateShopRequest defines model for UpdateShopRequest.
type UpdateShopRequest struct {
	DefaultEnvironmentId *int    `json:"defaultEnvironmentId,omitempty"`
	Description          *string `json:"description,omitempty"`
	GitUrl               *string `json:"gitUrl,omitempty"`
	Name                 *string `json:"name,omitempty"`
}

// UpdateSsoProviderRequest defines model for UpdateSsoProviderRequest.
type UpdateSsoProviderRequest struct {
	AuthorizationEndpoint string  `json:"authorizationEndpoint"`
	ClientId              string  `json:"clientId"`
	ClientSecret          *string `json:"clientSecret,omitempty"`
	Domain                string  `json:"domain"`
	Issuer                string  `json:"issuer"`
	JwksEndpoint          string  `json:"jwksEndpoint"`
	TokenEndpoint         string  `json:"tokenEndpoint"`
}

// UserProfile defines model for UserProfile.
type UserProfile struct {
	CreatedAt   time.Time `json:"createdAt"`
	DisplayName string    `json:"displayName"`
	Email       string    `json:"email"`
	Id          string    `json:"id"`

	// Locale The user's preferred language, used to localize notification emails.
	Locale string `json:"locale"`
}

// AdvisoryId defines model for AdvisoryId.
type AdvisoryId = string

// AdvisoryLimit defines model for AdvisoryLimit.
type AdvisoryLimit = int

// AdvisoryOffset defines model for AdvisoryOffset.
type AdvisoryOffset = int

// AdvisoryPackage defines model for AdvisoryPackage.
type AdvisoryPackage = string

// AdvisoryScope defines model for AdvisoryScope.
type AdvisoryScope string

func (AdvisoryScope) Schema(r huma.Registry) *huma.Schema {
	return namedStringEnumSchema(r, "AdvisoryScope", "affected", "all", "suppressed")
}

// AdvisorySearch defines model for AdvisorySearch.
type AdvisorySearch = string

// AdvisorySeverity defines model for AdvisorySeverity.
type AdvisorySeverity string

func (AdvisorySeverity) Schema(r huma.Registry) *huma.Schema {
	return namedStringEnumSchema(r, "AdvisorySeverity", "none", "low", "medium", "high", "critical")
}

// AdvisorySort defines model for AdvisorySort.
type AdvisorySort string

// AdvisoryTag defines model for AdvisoryTag.
type AdvisoryTag = string

// DeploymentId defines model for DeploymentId.
type DeploymentId = int

// EnvironmentId defines model for EnvironmentId.
type EnvironmentId = int

// KeyId defines model for KeyId.
type KeyId = string

// LanguageParam defines model for LanguageParam.
type LanguageParam string

// NotificationId defines model for NotificationId.
type NotificationId = int

// OrgId defines model for OrgId.
type OrgId = string

// ProviderId defines model for ProviderId.
type ProviderId = string

// ShopId defines model for ShopId.
type ShopId = int

// TaskId defines model for TaskId.
type TaskId = string

// TokenId defines model for TokenId.
type TokenId = int

// Forbidden defines model for Forbidden.
type Forbidden = ErrorResponse

// NotFound defines model for NotFound.
type NotFound = ErrorResponse

// Unauthorized defines model for Unauthorized.
type Unauthorized = ErrorResponse

// ValidationError defines model for ValidationError.
type ValidationError = ErrorResponse

// GetAccountExtensionsParams defines parameters for GetAccountExtensions.
type GetAccountExtensionsParams struct {
	// Language Language for localized store text (label, description, manual, changelog). Falls back to English.
	Language *GetAccountExtensionsParamsLanguage `form:"language,omitempty" json:"language,omitempty"`
}

// GetAccountExtensionsParamsLanguage defines parameters for GetAccountExtensions.
type GetAccountExtensionsParamsLanguage string

// GetAccountExtensionParams defines parameters for GetAccountExtension.
type GetAccountExtensionParams struct {
	// Language Language for localized store text (label, description, manual, changelog). Falls back to English.
	Language *GetAccountExtensionParamsLanguage `form:"language,omitempty" json:"language,omitempty"`
}

// GetAccountExtensionParamsLanguage defines parameters for GetAccountExtension.
type GetAccountExtensionParamsLanguage string

// UpdateAccountMeJSONBody defines parameters for UpdateAccountMe.
type UpdateAccountMeJSONBody struct {
	// Locale Preferred language (e.g. "en", "de").
	Locale *string `json:"locale,omitempty"`
}

// DeleteNotificationPreferenceParams defines parameters for DeleteNotificationPreference.
type DeleteNotificationPreferenceParams struct {
	ScopeType string  `form:"scopeType" json:"scopeType"`
	ScopeId   *string `form:"scopeId,omitempty" json:"scopeId,omitempty"`
	EventType *string `form:"eventType,omitempty" json:"eventType,omitempty"`
	Channel   string  `form:"channel" json:"channel"`
}

// AdminListAdvisoriesParams defines parameters for AdminListAdvisories.
type AdminListAdvisoriesParams struct {
	Limit  *AdvisoryLimit  `form:"limit,omitempty" json:"limit,omitempty"`
	Offset *AdvisoryOffset `form:"offset,omitempty" json:"offset,omitempty"`

	// Package Filter by Composer package name (e.g. shopware/core)
	Package *AdvisoryPackage `form:"package,omitempty" json:"package,omitempty"`

	// Severity Filter by effective severity
	Severity *AdminListAdvisoriesParamsSeverity `form:"severity,omitempty" json:"severity,omitempty"`

	// Tag Filter by admin tag
	Tag *AdvisoryTag `form:"tag,omitempty" json:"tag,omitempty"`

	// Q Search title, CVE, GHSA, package, or advisory ID
	Q *AdvisorySearch `form:"q,omitempty" json:"q,omitempty"`

	// Visible When set, filter by visibility
	Visible *bool `form:"visible,omitempty" json:"visible,omitempty"`
}

// AdminListAdvisoriesParamsSeverity defines parameters for AdminListAdvisories.
type AdminListAdvisoriesParamsSeverity string

// AdminGetAuditLogParams defines parameters for AdminGetAuditLog.
type AdminGetAuditLogParams struct {
	Limit        *int    `form:"limit,omitempty" json:"limit,omitempty"`
	Offset       *int    `form:"offset,omitempty" json:"offset,omitempty"`
	Action       *string `form:"action,omitempty" json:"action,omitempty"`
	ActorUserId  *string `form:"actorUserId,omitempty" json:"actorUserId,omitempty"`
	TargetUserId *string `form:"targetUserId,omitempty" json:"targetUserId,omitempty"`
}

// AdminGetEnvironmentsParams defines parameters for AdminGetEnvironments.
type AdminGetEnvironmentsParams struct {
	Limit          *int                                     `form:"limit,omitempty" json:"limit,omitempty"`
	Offset         *int                                     `form:"offset,omitempty" json:"offset,omitempty"`
	SortBy         *string                                  `form:"sortBy,omitempty" json:"sortBy,omitempty"`
	SortDirection  *AdminGetEnvironmentsParamsSortDirection `form:"sortDirection,omitempty" json:"sortDirection,omitempty"`
	SearchField    *string                                  `form:"searchField,omitempty" json:"searchField,omitempty"`
	SearchOperator *string                                  `form:"searchOperator,omitempty" json:"searchOperator,omitempty"`
	SearchValue    *string                                  `form:"searchValue,omitempty" json:"searchValue,omitempty"`
	FilterField    *string                                  `form:"filterField,omitempty" json:"filterField,omitempty"`
	FilterOperator *string                                  `form:"filterOperator,omitempty" json:"filterOperator,omitempty"`
	FilterValue    *string                                  `form:"filterValue,omitempty" json:"filterValue,omitempty"`
}

// AdminGetEnvironmentsParamsSortDirection defines parameters for AdminGetEnvironments.
type AdminGetEnvironmentsParamsSortDirection string

// AdminGetOrganizationsParams defines parameters for AdminGetOrganizations.
type AdminGetOrganizationsParams struct {
	Limit          *int                                      `form:"limit,omitempty" json:"limit,omitempty"`
	Offset         *int                                      `form:"offset,omitempty" json:"offset,omitempty"`
	SortBy         *AdminGetOrganizationsParamsSortBy        `form:"sortBy,omitempty" json:"sortBy,omitempty"`
	SortDirection  *AdminGetOrganizationsParamsSortDirection `form:"sortDirection,omitempty" json:"sortDirection,omitempty"`
	SearchField    *string                                   `form:"searchField,omitempty" json:"searchField,omitempty"`
	SearchOperator *string                                   `form:"searchOperator,omitempty" json:"searchOperator,omitempty"`
	SearchValue    *string                                   `form:"searchValue,omitempty" json:"searchValue,omitempty"`
	FilterField    *string                                   `form:"filterField,omitempty" json:"filterField,omitempty"`
	FilterOperator *string                                   `form:"filterOperator,omitempty" json:"filterOperator,omitempty"`
	FilterValue    *string                                   `form:"filterValue,omitempty" json:"filterValue,omitempty"`
}

// AdminGetOrganizationsParamsSortBy defines parameters for AdminGetOrganizations.
type AdminGetOrganizationsParamsSortBy string

// AdminGetOrganizationsParamsSortDirection defines parameters for AdminGetOrganizations.
type AdminGetOrganizationsParamsSortDirection string

// ListAdvisoriesParams defines parameters for ListAdvisories.
type ListAdvisoriesParams struct {
	Limit  *AdvisoryLimit  `form:"limit,omitempty" json:"limit,omitempty"`
	Offset *AdvisoryOffset `form:"offset,omitempty" json:"offset,omitempty"`

	// Package Filter by Composer package name (e.g. shopware/core)
	Package *AdvisoryPackage `form:"package,omitempty" json:"package,omitempty"`

	// Severity Filter by effective severity
	Severity *ListAdvisoriesParamsSeverity `form:"severity,omitempty" json:"severity,omitempty"`

	// Tag Filter by admin tag
	Tag *AdvisoryTag `form:"tag,omitempty" json:"tag,omitempty"`

	// Q Search title, CVE, GHSA, package, or advisory ID
	Q *AdvisorySearch `form:"q,omitempty" json:"q,omitempty"`

	// Scope "affected" limits results to advisories matching the Composer inventory of the caller's own environments, excluding suppressed ones; "suppressed" returns only those the caller has acknowledged; "all" returns the full catalog.
	Scope *ListAdvisoriesParamsScope `form:"scope,omitempty" json:"scope,omitempty"`

	// Sort Sort order: "reported" newest first, "severity" most severe first, "affected" most affected environments first, "cvss" highest score first.
	Sort *ListAdvisoriesParamsSort `form:"sort,omitempty" json:"sort,omitempty"`
}

// ListAdvisoriesParamsSeverity defines parameters for ListAdvisories.
type ListAdvisoriesParamsSeverity string

// ListAdvisoriesParamsScope defines parameters for ListAdvisories.
type ListAdvisoriesParamsScope string

// ListAdvisoriesParamsSort defines parameters for ListAdvisories.
type ListAdvisoriesParamsSort string

// GetEnvironmentParams defines parameters for GetEnvironment.
type GetEnvironmentParams struct {
	// Language Language for localized store text (label, description, manual, changelog). Falls back to English.
	Language *GetEnvironmentParamsLanguage `form:"language,omitempty" json:"language,omitempty"`
}

// GetEnvironmentParamsLanguage defines parameters for GetEnvironment.
type GetEnvironmentParamsLanguage string

// GetEnvironmentChangelogsParams defines parameters for GetEnvironmentChangelogs.
type GetEnvironmentChangelogsParams struct {
	Limit  *int `form:"limit,omitempty" json:"limit,omitempty"`
	Offset *int `form:"offset,omitempty" json:"offset,omitempty"`
}

// GetDeploymentsParams defines parameters for GetDeployments.
type GetDeploymentsParams struct {
	Limit  *int `form:"limit,omitempty" json:"limit,omitempty"`
	Offset *int `form:"offset,omitempty" json:"offset,omitempty"`
}

// RefreshEnvironmentJSONBody defines parameters for RefreshEnvironment.
type RefreshEnvironmentJSONBody struct {
	// Sitespeed Also run sitespeed check
	Sitespeed *bool `json:"sitespeed,omitempty"`
}

// DiscoverSsoParams defines parameters for DiscoverSso.
type DiscoverSsoParams struct {
	Issuer string `form:"issuer" json:"issuer"`
}

// ListSuppressionsParams defines parameters for ListSuppressions.
type ListSuppressionsParams struct {
	// IncludeInactive Include revoked and expired suppressions
	IncludeInactive *bool `form:"includeInactive,omitempty" json:"includeInactive,omitempty"`
}

// UpdateAccountMeJSONRequestBody defines body for UpdateAccountMe for application/json ContentType.
type UpdateAccountMeJSONRequestBody UpdateAccountMeJSONBody

// SetNotificationPreferenceJSONRequestBody defines body for SetNotificationPreference for application/json ContentType.
type SetNotificationPreferenceJSONRequestBody = NotificationPreferenceInput

// AdminUpdateAdvisoryJSONRequestBody defines body for AdminUpdateAdvisory for application/json ContentType.
type AdminUpdateAdvisoryJSONRequestBody = UpdateAdvisoryEnrichmentRequest

// CreateAdvisorySuppressionJSONRequestBody defines body for CreateAdvisorySuppression for application/json ContentType.
type CreateAdvisorySuppressionJSONRequestBody = CreateAdvisorySuppressionRequest

// CreateCliDeploymentJSONRequestBody defines body for CreateCliDeployment for application/json ContentType.
type CreateCliDeploymentJSONRequestBody = CreateCliDeploymentRequest

// CreateEnvironmentJSONRequestBody defines body for CreateEnvironment for application/json ContentType.
type CreateEnvironmentJSONRequestBody = CreateEnvironmentRequest

// UpdateEnvironmentJSONRequestBody defines body for UpdateEnvironment for application/json ContentType.
type UpdateEnvironmentJSONRequestBody = UpdateEnvironmentRequest

// RefreshEnvironmentJSONRequestBody defines body for RefreshEnvironment for application/json ContentType.
type RefreshEnvironmentJSONRequestBody RefreshEnvironmentJSONBody

// UpdateSitespeedSettingsJSONRequestBody defines body for UpdateSitespeedSettings for application/json ContentType.
type UpdateSitespeedSettingsJSONRequestBody = SitespeedSettingsRequest

// CheckExtensionCompatibilityJSONRequestBody defines body for CheckExtensionCompatibility for application/json ContentType.
type CheckExtensionCompatibilityJSONRequestBody = ExtensionCompatibilityRequest

// CreateShopJSONRequestBody defines body for CreateShop for application/json ContentType.
type CreateShopJSONRequestBody = CreateShopRequest

// UpdateShopJSONRequestBody defines body for UpdateShop for application/json ContentType.
type UpdateShopJSONRequestBody = UpdateShopRequest

// CreateApiKeyJSONRequestBody defines body for CreateApiKey for application/json ContentType.
type CreateApiKeyJSONRequestBody = CreateApiKeyRequest

// CreatePackagesTokenJSONRequestBody defines body for CreatePackagesToken for application/json ContentType.
type CreatePackagesTokenJSONRequestBody = CreatePackagesTokenRequest

// UpdateSsoProviderJSONRequestBody defines body for UpdateSsoProvider for application/json ContentType.
type UpdateSsoProviderJSONRequestBody = UpdateSsoProviderRequest
