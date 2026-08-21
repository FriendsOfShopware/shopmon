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
		Errors:      []int{http.StatusUnauthorized},
	}, h.ListAdvisories)

	huma.Register(api, huma.Operation{
		OperationID: "listAdvisoryPackages",
		Method:      http.MethodGet,
		Path:        "/advisories/packages",
		Summary:     "List package names that have visible advisories",
		Tags:        []string{"Advisories"},
		Errors:      []int{http.StatusUnauthorized},
	}, h.ListAdvisoryPackages)

	huma.Register(api, huma.Operation{
		OperationID: "getAdvisory",
		Method:      http.MethodGet,
		Path:        "/advisories/{advisoryId}",
		Summary:     "Get a visible advisory by Packagist advisory ID",
		Tags:        []string{"Advisories"},
		Errors:      []int{http.StatusUnauthorized, http.StatusNotFound},
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
		Tags:        []string{"Admin"},
		Errors:      []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
	}, h.AdminUpdateAdvisory)
}
