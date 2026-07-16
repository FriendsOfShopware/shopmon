package monitoring

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubRepository struct {
	shop                   Shop
	shopErr                error
	environment            Environment
	environmentErr         error
	createShopResult       CreateShopResult
	createShopErr          error
	createEnvironmentID    int32
	createEnvironmentErr   error
	createdShop            CreateShopRecord
	createdEnvironment     CreateEnvironmentRecord
	updatedShop            Shop
	updatedDefaultID       *int32
	updateShopCalls        int
	shopEnvironmentCount   int32
	countErr               error
	deleteShopCalls        int
	deleteShopErr          error
	updatedEnvironment     Environment
	updateEnvironmentCalls int
	updateEnvironmentErr   error
	deleteEnvironmentCalls int
	deleteEnvironmentErr   error
	deploymentIDs          []int32
	deploymentIDsErr       error
	sitespeedCalls         int
	sitespeedEnabled       bool
	sitespeedURLs          []string
	sitespeedErr           error
}

func (r *stubRepository) FindShop(context.Context, int32) (Shop, error) {
	return r.shop, r.shopErr
}

func (r *stubRepository) FindEnvironment(context.Context, int32) (Environment, error) {
	return r.environment, r.environmentErr
}

func (r *stubRepository) CreateShopWithEnvironment(_ context.Context, shop CreateShopRecord, environment CreateEnvironmentRecord) (int32, int32, error) {
	r.createdShop = shop
	r.createdEnvironment = environment
	return r.createShopResult.ShopID, r.createShopResult.EnvironmentID, r.createShopErr
}

func (r *stubRepository) CreateEnvironment(_ context.Context, environment CreateEnvironmentRecord) (int32, error) {
	r.createdEnvironment = environment
	return r.createEnvironmentID, r.createEnvironmentErr
}

func (r *stubRepository) UpdateShop(_ context.Context, shop Shop, defaultEnvironmentID *int32) error {
	r.updateShopCalls++
	r.updatedShop = shop
	r.updatedDefaultID = defaultEnvironmentID
	return nil
}

func (r *stubRepository) CountShopEnvironments(context.Context, int32) (int32, error) {
	return r.shopEnvironmentCount, r.countErr
}

func (r *stubRepository) DeleteShop(context.Context, int32, string) error {
	r.deleteShopCalls++
	return r.deleteShopErr
}

func (r *stubRepository) UpdateEnvironment(_ context.Context, environment Environment) error {
	r.updateEnvironmentCalls++
	r.updatedEnvironment = environment
	return r.updateEnvironmentErr
}

func (r *stubRepository) DeleteEnvironment(context.Context, int32) error {
	r.deleteEnvironmentCalls++
	return r.deleteEnvironmentErr
}

func (r *stubRepository) ListDeploymentIDs(context.Context, int32) ([]int32, error) {
	return r.deploymentIDs, r.deploymentIDsErr
}

func (r *stubRepository) UpdateSitespeedSettings(_ context.Context, _ int32, enabled bool, urls []string) error {
	r.sitespeedCalls++
	r.sitespeedEnabled = enabled
	r.sitespeedURLs = urls
	return r.sitespeedErr
}

func (r *stubRepository) UpdateUserEnvironmentSubscriptions(context.Context, string, []string) error {
	return nil
}

type stubAuthorizer struct {
	member bool
	err    error
	calls  int
}

func (a *stubAuthorizer) IsMember(context.Context, string, string) (bool, error) {
	a.calls++
	return a.member, a.err
}

type stubEnvironmentGateway struct {
	shopInfo        ShopInfo
	validateErr     error
	validateCalls   int
	credentials     ConnectionCredentials
	encryptedSecret string
	encryptErr      error
	clearErr        error
	rescheduleErr   error
}

func (g *stubEnvironmentGateway) ValidateConnection(_ context.Context, credentials ConnectionCredentials) (ShopInfo, error) {
	g.validateCalls++
	g.credentials = credentials
	return g.shopInfo, g.validateErr
}

func (g *stubEnvironmentGateway) EncryptSecret(string) (string, error) {
	return g.encryptedSecret, g.encryptErr
}

func (g *stubEnvironmentGateway) ClearCache(context.Context, StoredCredentials) error {
	return g.clearErr
}

func (g *stubEnvironmentGateway) RescheduleTask(context.Context, StoredCredentials, string, time.Time) error {
	return g.rescheduleErr
}

type stubJobDispatcher struct {
	scrapeIDs    []int32
	sitespeedIDs []int32
	scrapeErr    error
	sitespeedErr error
}

func (d *stubJobDispatcher) DispatchEnvironmentScrape(_ context.Context, environmentID int32) error {
	d.scrapeIDs = append(d.scrapeIDs, environmentID)
	return d.scrapeErr
}

func (d *stubJobDispatcher) DispatchSitespeedScrape(_ context.Context, environmentID int32) error {
	d.sitespeedIDs = append(d.sitespeedIDs, environmentID)
	return d.sitespeedErr
}

func TestCreateShopOrchestratesLifecycle(t *testing.T) {
	repository := &stubRepository{createShopResult: CreateShopResult{ShopID: 10, EnvironmentID: 20}}
	authorizer := &stubAuthorizer{member: true}
	gateway := &stubEnvironmentGateway{shopInfo: ShopInfo{Version: "6.7.0"}, encryptedSecret: "encrypted"}
	jobs := &stubJobDispatcher{}
	service := NewService(repository, authorizer, gateway, jobs, nil)

	result, err := service.CreateShop(t.Context(), CreateShopCommand{
		UserID:          "user-1",
		OrganizationID:  "org-1",
		Name:            "Shop",
		EnvironmentName: "Production",
		EnvironmentURL:  "https://shop.example.com",
		ClientID:        "client",
		ClientSecret:    "secret",
	})

	require.NoError(t, err)
	assert.Equal(t, CreateShopResult{ShopID: 10, EnvironmentID: 20}, result)
	assert.Equal(t, "org-1", repository.createdShop.OrganizationID)
	assert.Equal(t, "encrypted", repository.createdEnvironment.EncryptedClientSecret)
	assert.Equal(t, "6.7.0", repository.createdEnvironment.ShopwareVersion)
	assert.Equal(t, []int32{20}, jobs.scrapeIDs)
}

func TestCreateShopStopsBeforeSideEffectsWhenUnauthorized(t *testing.T) {
	repository := &stubRepository{}
	authorizer := &stubAuthorizer{}
	gateway := &stubEnvironmentGateway{}
	service := NewService(repository, authorizer, gateway, &stubJobDispatcher{}, nil)

	_, err := service.CreateShop(t.Context(), CreateShopCommand{UserID: "user-1", OrganizationID: "org-1"})

	require.ErrorIs(t, err, ErrNotAuthorized)
	assert.Equal(t, 0, gateway.validateCalls)
	assert.Empty(t, repository.createdShop.OrganizationID)
}

func TestUpdateShopValidatesDefaultEnvironmentBeforeWriting(t *testing.T) {
	defaultID := int32(22)
	repository := &stubRepository{
		shop:        Shop{ID: 10, OrganizationID: "org-1", Name: "Shop"},
		environment: Environment{ID: defaultID, ShopID: 99, OrganizationID: "org-1"},
	}
	service := NewService(repository, &stubAuthorizer{member: true}, &stubEnvironmentGateway{}, &stubJobDispatcher{}, nil)

	err := service.UpdateShop(t.Context(), UpdateShopCommand{
		UserID:               "user-1",
		OrganizationID:       "org-1",
		ShopID:               10,
		DefaultEnvironmentID: &defaultID,
	})

	require.ErrorIs(t, err, ErrDefaultEnvironmentMismatch)
	assert.Equal(t, 0, repository.updateShopCalls)
}

func TestUpdateEnvironmentValidatesAndEncryptsNewSecret(t *testing.T) {
	repository := &stubRepository{environment: Environment{
		ID:                    20,
		OrganizationID:        "org-1",
		ShopID:                10,
		URL:                   "https://old.example.com",
		ClientID:              "old-client",
		EncryptedClientSecret: "old-encrypted",
	}}
	gateway := &stubEnvironmentGateway{encryptedSecret: "new-encrypted"}
	service := NewService(repository, &stubAuthorizer{member: true}, gateway, &stubJobDispatcher{}, nil)
	url := "https://new.example.com"
	clientID := "new-client"
	secret := "new-secret"

	err := service.UpdateEnvironment(t.Context(), UpdateEnvironmentCommand{
		UserID:        "user-1",
		EnvironmentID: 20,
		ShopID:        10,
		URL:           &url,
		ClientID:      &clientID,
		ClientSecret:  &secret,
	})

	require.NoError(t, err)
	assert.Equal(t, ConnectionCredentials{URL: url, ClientID: clientID, ClientSecret: secret}, gateway.credentials)
	assert.Equal(t, "new-encrypted", repository.updatedEnvironment.EncryptedClientSecret)
	assert.Equal(t, 1, repository.updateEnvironmentCalls)
}

func TestDeleteShopRejectsExistingEnvironments(t *testing.T) {
	repository := &stubRepository{
		shop:                 Shop{ID: 10, OrganizationID: "org-1"},
		shopEnvironmentCount: 1,
	}
	service := NewService(repository, &stubAuthorizer{member: true}, &stubEnvironmentGateway{}, &stubJobDispatcher{}, nil)

	err := service.DeleteShop(t.Context(), "user-1", "org-1", 10)

	require.ErrorIs(t, err, ErrShopHasEnvironments)
	assert.Equal(t, 0, repository.deleteShopCalls)
}

func TestRefreshEnvironmentDispatchesRequestedJobs(t *testing.T) {
	repository := &stubRepository{environment: Environment{ID: 20, OrganizationID: "org-1"}}
	jobs := &stubJobDispatcher{}
	service := NewService(repository, &stubAuthorizer{member: true}, &stubEnvironmentGateway{}, jobs, nil)

	err := service.RefreshEnvironment(t.Context(), "user-1", 20, true)

	require.NoError(t, err)
	assert.Equal(t, []int32{20}, jobs.scrapeIDs)
	assert.Equal(t, []int32{20}, jobs.sitespeedIDs)
}

func TestRefreshEnvironmentReturnsMandatoryDispatchError(t *testing.T) {
	dispatchErr := errors.New("queue unavailable")
	repository := &stubRepository{environment: Environment{ID: 20, OrganizationID: "org-1"}}
	jobs := &stubJobDispatcher{scrapeErr: dispatchErr}
	service := NewService(repository, &stubAuthorizer{member: true}, &stubEnvironmentGateway{}, jobs, nil)

	err := service.RefreshEnvironment(t.Context(), "user-1", 20, true)

	require.ErrorIs(t, err, dispatchErr)
	assert.Empty(t, jobs.sitespeedIDs)
}
