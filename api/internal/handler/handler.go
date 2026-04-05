package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/auth"
	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/mail"
	"github.com/friendsofshopware/shopmon/api/internal/middleware"
	"github.com/friendsofshopware/shopmon/api/internal/storage"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	goqueue "github.com/shyim/go-queue"
)

// Ensure Handler implements ServerInterface at compile time
var _ api.ServerInterface = (*Handler)(nil)

type Handler struct {
	pool    *pgxpool.Pool
	queries *queries.Queries
	storage *storage.S3Storage
	cfg     *config.Config
	mail    *mail.Service
	bus     *goqueue.Bus
}

func New(pool *pgxpool.Pool, q *queries.Queries, s *storage.S3Storage, cfg *config.Config, mail *mail.Service, bus *goqueue.Bus) *Handler {
	return &Handler{
		pool:    pool,
		queries: q,
		storage: s,
		cfg:     cfg,
		mail:    mail,
		bus:     bus,
	}
}

func pgtimeToTime(t pgtype.Timestamp) time.Time {
	return t.Time
}

func pgtimeToTimePtr(t pgtype.Timestamp) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func sitespeedDetailUrl(cfg *config.Config, environmentID int32, enabled bool) *string {
	if !enabled || cfg.SitespeedEndpoint == "" {
		return nil
	}
	url := fmt.Sprintf("%s/result/%senvironment-%d/index.html", cfg.SitespeedEndpoint, cfg.SitespeedPrefix, environmentID)
	return &url
}

// requireUser extracts the authenticated user from context.
// Returns nil and writes a 401 response if not authenticated.
func (h *Handler) requireUser(w http.ResponseWriter, r *http.Request) *auth.User {
	user := middleware.GetUser(r.Context())
	if user == nil {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return nil
	}
	return user
}

// getActiveOrganizationID returns the active organization ID from the session, or nil if not set.
func (h *Handler) getActiveOrganizationID(r *http.Request) *string {
	sess := middleware.GetSession(r.Context())
	if sess == nil {
		return nil
	}
	return sess.ActiveOrganizationID
}

// requireOrgMembership checks if the user is a member of the given organization.
// Returns false and writes a 403 response if not a member.
func (h *Handler) requireOrgMembership(w http.ResponseWriter, r *http.Request, user *auth.User, orgID string) bool {
	isMember, err := h.queries.IsOrganizationMember(r.Context(), queries.IsOrganizationMemberParams{
		OrganizationID: orgID,
		UserID:         user.ID,
	})
	if err != nil || !isMember {
		httputil.WriteError(w, http.StatusForbidden, "not a member of this organization")
		return false
	}
	return true
}

func (h *Handler) loadAuthorizedEnvironment(w http.ResponseWriter, r *http.Request, user *auth.User, environmentID int32) (*queries.GetEnvironmentByIDRow, bool) {
	environment, err := h.queries.GetEnvironmentByID(r.Context(), environmentID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "environment not found")
		return nil, false
	}

	if !h.requireOrgMembership(w, r, user, environment.OrganizationID) {
		return nil, false
	}

	return &environment, true
}

func (h *Handler) loadAuthorizedEnvironmentCredentials(w http.ResponseWriter, r *http.Request, user *auth.User, environmentID int32) (*queries.GetEnvironmentCredentialsRow, bool) {
	creds, err := h.queries.GetEnvironmentCredentials(r.Context(), environmentID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "environment not found")
		return nil, false
	}

	orgID, err := h.queries.GetEnvironmentOrganizationID(r.Context(), environmentID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "environment not found")
		return nil, false
	}

	if !h.requireOrgMembership(w, r, user, orgID) {
		return nil, false
	}

	return &creds, true
}

func int32PtrToIntPtr(v *int32) *int {
	if v == nil {
		return nil
	}
	i := int(*v)
	return &i
}

func derefInt32(v *int32) int {
	if v == nil {
		return 0
	}
	return int(*v)
}

func (h *Handler) requireShopInOrganization(w http.ResponseWriter, r *http.Request, shopID int32, orgID string) bool {
	shop, err := h.queries.GetShopByID(r.Context(), shopID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "shop not found")
		return false
	}

	if shop.OrganizationID != orgID {
		httputil.WriteError(w, http.StatusForbidden, "shop does not belong to this organization")
		return false
	}

	return true
}
