// Package slackwebhook manages Slack incoming webhooks of an organization and
// of its shops, the destinations of alerts (see notify.Dispatcher).
package slackwebhook

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/crypto"
	"github.com/friendsofshopware/shopmon/api/internal/organization"
	"github.com/friendsofshopware/shopmon/api/internal/slack"
)

var (
	ErrNotConfigured  = errors.New("slack webhook not configured")
	ErrShopNotFound   = errors.New("shop not found")
	ErrDeliveryFailed = errors.New("slack rejected the message")
)

var managerRoles = []organization.Role{organization.RoleOwner, organization.RoleAdmin}

// Target is the webhook owner: the organization, or one of its shops when
// ShopID is set.
type Target struct {
	OrganizationID string
	ShopID         int32
}

func (t Target) isShop() bool { return t.ShopID != 0 }

// Status is what the UI may know about a webhook: never the URL itself, which
// is the credential for posting into the channel.
type Status struct {
	Configured bool
	CanManage  bool
}

type Repository interface {
	// Find returns the encrypted webhook URL or ErrNotConfigured.
	Find(ctx context.Context, target Target) (string, error)
	Upsert(ctx context.Context, target Target, encryptedURL string) error
	Delete(ctx context.Context, target Target) error
	ShopBelongsToOrganization(ctx context.Context, organizationID string, shopID int32) (bool, error)
}

type Authorizer interface {
	Membership(ctx context.Context, userID, organizationID string) (organization.Role, error)
	RequireRole(ctx context.Context, userID, organizationID string, allowedRoles ...organization.Role) (organization.Role, error)
}

type Poster interface {
	Post(ctx context.Context, webhookURL, text string) error
}

type Service struct {
	repository Repository
	authorizer Authorizer
	poster     Poster
	appSecret  string
}

func NewService(repository Repository, authorizer Authorizer, poster Poster, appSecret string) *Service {
	return &Service{repository: repository, authorizer: authorizer, poster: poster, appSecret: appSecret}
}

func (s *Service) Status(ctx context.Context, userID string, target Target) (Status, error) {
	role, err := s.authorizer.Membership(ctx, userID, target.OrganizationID)
	if err != nil {
		return Status{}, err
	}
	if err := s.requireShop(ctx, target); err != nil {
		return Status{}, err
	}

	status := Status{CanManage: slices.Contains(managerRoles, role)}
	_, err = s.repository.Find(ctx, target)
	switch {
	case errors.Is(err, ErrNotConfigured):
	case err != nil:
		return Status{}, err
	default:
		status.Configured = true
	}
	return status, nil
}

func (s *Service) Set(ctx context.Context, userID string, target Target, webhookURL string) error {
	if err := s.requireManager(ctx, userID, target); err != nil {
		return err
	}
	webhookURL = strings.TrimSpace(webhookURL)
	if err := slack.ValidateWebhookURL(webhookURL); err != nil {
		return err
	}
	encrypted, err := crypto.Encrypt(webhookURL, s.appSecret)
	if err != nil {
		return fmt.Errorf("encrypt slack webhook: %w", err)
	}
	if err := s.repository.Upsert(ctx, target, encrypted); err != nil {
		return fmt.Errorf("save slack webhook: %w", err)
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, userID string, target Target) error {
	if err := s.requireManager(ctx, userID, target); err != nil {
		return err
	}
	if err := s.repository.Delete(ctx, target); err != nil {
		return fmt.Errorf("delete slack webhook: %w", err)
	}
	return nil
}

// SendTest posts a confirmation message so admins can verify the channel.
func (s *Service) SendTest(ctx context.Context, userID string, target Target) error {
	if err := s.requireManager(ctx, userID, target); err != nil {
		return err
	}
	encrypted, err := s.repository.Find(ctx, target)
	if err != nil {
		return err
	}
	webhookURL, err := crypto.Decrypt(encrypted, s.appSecret)
	if err != nil {
		return fmt.Errorf("decrypt slack webhook: %w", err)
	}
	if err := s.poster.Post(ctx, webhookURL, testMessage(target)); err != nil {
		return fmt.Errorf("%w: %w", ErrDeliveryFailed, err)
	}
	return nil
}

func testMessage(target Target) string {
	if target.isShop() {
		return ":white_check_mark: Shopmon alerts for this shop will be posted to this channel."
	}
	return ":white_check_mark: Shopmon alerts for this organization will be posted to this channel."
}

func (s *Service) requireManager(ctx context.Context, userID string, target Target) error {
	if _, err := s.authorizer.RequireRole(ctx, userID, target.OrganizationID, managerRoles...); err != nil {
		return err
	}
	return s.requireShop(ctx, target)
}

// requireShop keeps a shop ID from another organization out of reach of this
// organization's admins.
func (s *Service) requireShop(ctx context.Context, target Target) error {
	if !target.isShop() {
		return nil
	}
	ok, err := s.repository.ShopBelongsToOrganization(ctx, target.OrganizationID, target.ShopID)
	if err != nil {
		return fmt.Errorf("check shop: %w", err)
	}
	if !ok {
		return ErrShopNotFound
	}
	return nil
}
