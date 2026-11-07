// Package destination manages organization and shop notification destinations.
package destination

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/crypto"
	"github.com/friendsofshopware/shopmon/api/internal/delivery"
	"github.com/friendsofshopware/shopmon/api/internal/organization"
	"github.com/google/uuid"
)

var (
	ErrNotConfigured  = errors.New("notification destination not found")
	ErrShopNotFound   = errors.New("shop not found")
	ErrDeliveryFailed = errors.New("notification test delivery failed")
	ErrInvalidName    = errors.New("destination name must contain between 1 and 100 characters")
)
var managerRoles = []organization.Role{organization.RoleOwner, organization.RoleAdmin}

type Target struct {
	OrganizationID string
	ShopID         int32
}

func (t Target) isShop() bool { return t.ShopID != 0 }

type Record struct {
	ID   string
	Name string
	DSN  string
}

type Destination struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Summary  string `json:"summary"`
}

type Status struct {
	CanManage    bool          `json:"canManage"`
	Destinations []Destination `json:"destinations"`
}

type Repository interface {
	List(context.Context, Target) ([]Record, error)
	Save(context.Context, Target, Record) error
	Delete(context.Context, Target, string) error
	ShopBelongsToOrganization(context.Context, string, int32) (bool, error)
}

type Authorizer interface {
	Membership(context.Context, string, string) (organization.Role, error)
	RequireRole(context.Context, string, string, ...organization.Role) (organization.Role, error)
}

type Poster interface {
	Post(context.Context, string, string) error
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
	rows, err := s.repository.List(ctx, target)
	if err != nil {
		return Status{}, fmt.Errorf("list destinations: %w", err)
	}
	status := Status{CanManage: slices.Contains(managerRoles, role), Destinations: []Destination{}}
	for _, row := range rows {
		raw, err := crypto.Decrypt(row.DSN, s.appSecret)
		if err != nil {
			return Status{}, fmt.Errorf("decrypt destination: %w", err)
		}
		details, err := delivery.Describe(raw)
		if err != nil {
			return Status{}, err
		}
		status.Destinations = append(status.Destinations, Destination{ID: row.ID, Name: row.Name, Provider: details.Provider, Summary: details.Summary})
	}
	return status, nil
}

func (s *Service) find(ctx context.Context, target Target, id string) (Record, error) {
	rows, err := s.repository.List(ctx, target)
	if err != nil {
		return Record{}, fmt.Errorf("find destination: %w", err)
	}
	for _, row := range rows {
		if row.ID == id {
			return row, nil
		}
	}
	return Record{}, ErrNotConfigured
}

// Save creates a destination when id is empty. An empty DSN on update keeps the stored secret.
func (s *Service) Save(ctx context.Context, userID string, target Target, id, name, dsn string) error {
	if err := s.requireManager(ctx, userID, target); err != nil {
		return err
	}
	name = strings.TrimSpace(name)
	dsn = strings.TrimSpace(dsn)
	if len([]rune(name)) < 1 || len([]rune(name)) > 100 {
		return ErrInvalidName
	}
	var record Record
	if id != "" {
		var err error
		record, err = s.find(ctx, target, id)
		if err != nil {
			return err
		}
	} else {
		record.ID = uuid.NewString()
	}
	if dsn != "" || id == "" {
		if _, err := delivery.Validate(dsn); err != nil {
			return err
		}
		encrypted, err := crypto.Encrypt(dsn, s.appSecret)
		if err != nil {
			return fmt.Errorf("encrypt destination: %w", err)
		}
		record.DSN = encrypted
	}
	record.Name = name
	if err := s.repository.Save(ctx, target, record); err != nil {
		return fmt.Errorf("save destination: %w", err)
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, userID string, target Target, id string) error {
	if err := s.requireManager(ctx, userID, target); err != nil {
		return err
	}
	return s.repository.Delete(ctx, target, id)
}

func (s *Service) SendTest(ctx context.Context, userID string, target Target, id string) error {
	if err := s.requireManager(ctx, userID, target); err != nil {
		return err
	}
	row, err := s.find(ctx, target, id)
	if err != nil {
		return err
	}
	dsn, err := crypto.Decrypt(row.DSN, s.appSecret)
	if err != nil {
		return fmt.Errorf("decrypt destination: %w", err)
	}
	scope := "organization"
	if target.isShop() {
		scope = "shop"
	}
	if err := s.poster.Post(ctx, dsn, "✅ Shopmon alerts for this "+scope+" will be sent to this destination."); err != nil {
		return ErrDeliveryFailed
	}
	return nil
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
