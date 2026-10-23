package handler

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func registerAdvisories(api huma.API, h *Handler) {
	huma.Register(api, huma.Operation{
		OperationID: "listAdvisories",
		Method:      http.MethodGet,
		Path:        "/advisories",
		Summary:     "List visible Composer security advisories for Shopware packages",
		Tags:        []string{"Advisories"},
		Security:    []map[string][]string{},
	}, h.ListAdvisories)

	huma.Register(api, huma.Operation{
		OperationID: "listAdvisoryPackages",
		Method:      http.MethodGet,
		Path:        "/advisories/packages",
		Summary:     "List package names that have visible advisories",
		Tags:        []string{"Advisories"},
		Security:    []map[string][]string{},
	}, h.ListAdvisoryPackages)

	huma.Register(api, huma.Operation{
		OperationID: "getAdvisory",
		Method:      http.MethodGet,
		Path:        "/advisories/{advisoryId}",
		Summary:     "Get a visible advisory by Packagist advisory ID",
		Tags:        []string{"Advisories"},
		Security:    []map[string][]string{},
		Errors:      []int{http.StatusNotFound},
	}, h.GetAdvisory)

	huma.Register(api, huma.Operation{
		OperationID:   "createAdvisorySuppression",
		Method:        http.MethodPost,
		Path:          "/advisories/{advisoryId}/suppressions",
		Summary:       "Record that an advisory is accepted or mitigated for a shop",
		Description:   "Suppressing an advisory removes it from the \"affecting my shops\" list and silences its alerts. Shop-wide suppression requires the owner or admin role; narrowing to a single environment is open to any member.",
		Tags:          []string{"Advisories"},
		DefaultStatus: http.StatusCreated,
		Errors: []int{
			http.StatusBadRequest,
			http.StatusUnauthorized,
			http.StatusForbidden,
			http.StatusNotFound,
			http.StatusConflict,
		},
	}, h.CreateAdvisorySuppression)

	huma.Register(api, huma.Operation{
		OperationID: "listSuppressions",
		Method:      http.MethodGet,
		Path:        "/suppressions",
		Summary:     "List advisory suppressions for the caller's organizations",
		Tags:        []string{"Advisories"},
		Errors:      []int{http.StatusUnauthorized},
	}, h.ListSuppressions)

	huma.Register(api, huma.Operation{
		OperationID:   "revokeAdvisorySuppression",
		Method:        http.MethodDelete,
		Path:          "/suppressions/{suppressionId}",
		Summary:       "Revoke a suppression so the advisory resurfaces",
		Tags:          []string{"Advisories"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusUnauthorized, http.StatusNotFound},
	}, h.RevokeAdvisorySuppression)

	huma.Register(api, huma.Operation{
		OperationID: "listAdvisoryAffectedEnvironments",
		Method:      http.MethodGet,
		Path:        "/advisories/{advisoryId}/affected",
		Summary:     "List the caller's environments affected by an advisory",
		Description: "Environments whose Composer package inventory (from the FroshTools SBOM) matches this advisory. Restricted to organizations the caller belongs to; admins additionally receive a fleet-wide affected count.",
		Tags:        []string{"Advisories"},
		Errors:      []int{http.StatusUnauthorized, http.StatusNotFound},
	}, h.ListAdvisoryAffectedEnvironments)

	huma.Register(api, huma.Operation{
		OperationID: "adminListAdvisories",
		Method:      http.MethodGet,
		Path:        "/admin/advisories",
		Summary:     "List all Composer advisories including hidden ones (admin only)",
		Tags:        []string{"Admin"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden},
	}, h.AdminListAdvisories)

	huma.Register(api, huma.Operation{
		OperationID: "adminPreviewAdvisoryImpact",
		Method:      http.MethodPost,
		Path:        "/admin/advisories/impact",
		Summary:     "Preview affected shops for package constraints (admin only)",
		Description: "Checks the submitted Composer package constraints against collected inventories across all organizations. Returns distinct shops and their matching environments and packages, including suppressed matches. Does not save or publish an advisory.",
		Tags:        []string{"Admin"},
		Errors:      []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden},
	}, h.AdminPreviewAdvisoryImpact)

	huma.Register(api, huma.Operation{
		OperationID:   "adminCreateAdvisory",
		Method:        http.MethodPost,
		Path:          "/admin/advisories",
		Summary:       "Create a Shopmon-authored advisory (admin only)",
		Description:   "Creates an advisory with a stable SHOPMON-YYYY-NNNN id, marked as a Shopmon report until an official CVE/GHSA exists. Starts hidden as a draft unless isVisible is set; publishing matches and alerts affected shops like an official advisory.",
		Tags:          []string{"Admin"},
		DefaultStatus: http.StatusCreated,
		Errors:        []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusConflict},
	}, h.AdminCreateAdvisory)

	huma.Register(api, huma.Operation{
		OperationID: "adminPreviewGhsaImport",
		Method:      http.MethodPost,
		Path:        "/admin/advisories/import/ghsa/preview",
		Summary:     "Fetch a GitHub advisory for import review (admin only)",
		Description: "Loads a GitHub security advisory by GHSA id or advisory URL and projects it onto the create form's fields. Nothing is persisted; if the advisory is already tracked, existingAdvisoryId names the catalog row.",
		Tags:        []string{"Admin"},
		Errors:      []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusBadGateway, http.StatusServiceUnavailable},
	}, h.AdminPreviewGhsaImport)

	huma.Register(api, huma.Operation{
		OperationID:   "adminImportGhsaAdvisory",
		Method:        http.MethodPost,
		Path:          "/admin/advisories/import/ghsa",
		Summary:       "Create an advisory imported from GitHub (admin only)",
		Description:   "Creates the reviewed GitHub advisory with origin 'github' and the GHSA id as its advisory id, so a later CVE assignment links and merges without duplicates. Starts hidden as a draft unless isVisible is set.",
		Tags:          []string{"Admin"},
		DefaultStatus: http.StatusCreated,
		Errors:        []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusConflict},
	}, h.AdminImportGhsaAdvisory)

	huma.Register(api, huma.Operation{
		OperationID: "adminLinkAdvisoryOfficial",
		Method:      http.MethodPost,
		Path:        "/admin/advisories/{advisoryId}/link-official",
		Summary:     "Link a Shopmon-authored advisory to its official CVE/GHSA (admin only)",
		Description: "Records the official identifiers on the Shopmon advisory without changing its id. If the official advisory was already synced as its own row, it is merged into the Shopmon advisory: upstream content wins, matches, suppressions, and alert history carry over.",
		Tags:        []string{"Admin"},
		Errors:      []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusConflict},
	}, h.AdminLinkAdvisoryOfficial)

	huma.Register(api, huma.Operation{
		OperationID:   "adminSyncAdvisories",
		Method:        http.MethodPost,
		Path:          "/admin/advisories/sync",
		Summary:       "Enqueue a Packagist advisory sync job",
		Tags:          []string{"Admin"},
		DefaultStatus: http.StatusAccepted,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden},
	}, h.AdminSyncAdvisories)

	huma.Register(api, huma.Operation{
		OperationID: "adminGetAdvisory",
		Method:      http.MethodGet,
		Path:        "/admin/advisories/{advisoryId}",
		Summary:     "Get a single advisory with internal notes (admin only)",
		Tags:        []string{"Admin"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
	}, h.AdminGetAdvisory)

	huma.Register(api, huma.Operation{
		OperationID: "adminUpdateAdvisory",
		Method:      http.MethodPatch,
		Path:        "/admin/advisories/{advisoryId}",
		Summary:     "Update Shopmon enrichment fields for an advisory",
		Description: "Enrichment fields apply to every advisory. Core fields (title, packages, severity, disclosure details) are only accepted for Shopmon-authored advisories that are not yet linked to an official CVE/GHSA.",
		Tags:        []string{"Admin"},
		Errors:      []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusConflict},
	}, h.AdminUpdateAdvisory)
}
