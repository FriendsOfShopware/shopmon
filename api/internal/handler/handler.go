package handler

import (
	"context"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/access"
	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/catalog"
	"github.com/friendsofshopware/shopmon/api/internal/deployment"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/monitoring"
	"github.com/friendsofshopware/shopmon/api/internal/notification"
	organizationsso "github.com/friendsofshopware/shopmon/api/internal/organization/sso"
	"github.com/friendsofshopware/shopmon/api/internal/packagesmirror"
	"github.com/friendsofshopware/shopmon/api/internal/suppression"
	accountread "github.com/friendsofshopware/shopmon/api/internal/readmodel/account"
	adminread "github.com/friendsofshopware/shopmon/api/internal/readmodel/admin"
	advisoryread "github.com/friendsofshopware/shopmon/api/internal/readmodel/advisory"
	environmentread "github.com/friendsofshopware/shopmon/api/internal/readmodel/environment"
)

// AdvisorySyncDispatcher enqueues a Packagist advisory sync job.
type AdvisorySyncDispatcher interface {
	EnqueueComposerAdvisorySync(ctx context.Context) error
}

// Ensure Handler implements ServerInterface at compile time.
var _ api.ServerInterface = (*Handler)(nil)

type DatabaseHealth interface {
	Ping(ctx context.Context) error
}

// InstanceFeatures is the transport-facing projection of instance-wide
// feature flags. Keeping this small prevents the handler from depending on the
// complete environment configuration.
type InstanceFeatures struct {
	RegistrationEnabled  bool
	GithubAuthEnabled    bool
	SitespeedEnabled     bool
	PackageMirrorEnabled bool
}

// Dependencies is the HTTP composition boundary. Business operations are
// exposed as capability services and query-side read models; raw persistence,
// queues, mail, storage, and cache clients remain outside the handler package.
type Dependencies struct {
	Database      DatabaseHealth
	Features      InstanceFeatures
	Monitoring    *monitoring.Service
	Notifications *notification.Service
	Packages      *packagesmirror.Service
	Deployments   *deployment.Service
	Catalog       *catalog.Service
	SSO           *organizationsso.Service
	Account       *accountread.Service
	Admin         *adminread.Service
	Environments  *environmentread.Service
	Advisories    *advisoryread.Service
	AdvisorySync  AdvisorySyncDispatcher
	Suppressions  *suppression.Service
}

type Handler struct {
	database      DatabaseHealth
	features      InstanceFeatures
	monitoring    *monitoring.Service
	notifications *notification.Service
	packages      *packagesmirror.Service
	deployments   *deployment.Service
	catalog       *catalog.Service
	sso           *organizationsso.Service
	account       *accountread.Service
	admin         *adminread.Service
	environments  *environmentread.Service
	advisories    *advisoryread.Service
	advisorySync  AdvisorySyncDispatcher
	suppressions  *suppression.Service
}

func New(dependencies Dependencies) *Handler {
	return &Handler{
		database:      dependencies.Database,
		features:      dependencies.Features,
		monitoring:    dependencies.Monitoring,
		notifications: dependencies.Notifications,
		packages:      dependencies.Packages,
		deployments:   dependencies.Deployments,
		catalog:       dependencies.Catalog,
		sso:           dependencies.SSO,
		account:       dependencies.Account,
		admin:         dependencies.Admin,
		environments:  dependencies.Environments,
		advisories:    dependencies.Advisories,
		advisorySync:  dependencies.AdvisorySync,
		suppressions:  dependencies.Suppressions,
	}
}

// requireUser extracts the authenticated principal populated by auth
// middleware. It writes the transport response because authentication itself
// is an HTTP boundary concern.
func (h *Handler) requireUser(w http.ResponseWriter, r *http.Request) *access.User {
	user := access.UserFromContext(r.Context())
	if user == nil {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return nil
	}
	return user
}

func (h *Handler) getActiveOrganizationID(r *http.Request) *string {
	session := access.SessionFromContext(r.Context())
	if session == nil {
		return nil
	}
	return session.ActiveOrganizationID
}
