package deployment

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeDeploymentRepository is a minimal in-memory Repository for CreateCLI.
type fakeDeploymentRepository struct {
	apiKey      APIKey
	environment Environment
	created     []CreateDeploymentRecord
}

func (f *fakeDeploymentRepository) FindAPIKeyByTokenHash(context.Context, string) (APIKey, error) {
	return f.apiKey, nil
}

func (f *fakeDeploymentRepository) TouchAPIKey(context.Context, string) error { return nil }

func (f *fakeDeploymentRepository) FindEnvironment(context.Context, int32) (Environment, error) {
	return f.environment, nil
}

func (f *fakeDeploymentRepository) CreateDeployment(_ context.Context, record CreateDeploymentRecord) (int32, error) {
	f.created = append(f.created, record)
	return 42, nil
}

func (f *fakeDeploymentRepository) OrganizationIDForShop(context.Context, int32) (string, error) {
	return f.environment.OrganizationID, nil
}
func (f *fakeDeploymentRepository) ListDeployments(context.Context, int32, int32, int32) ([]Deployment, error) {
	return nil, nil
}
func (f *fakeDeploymentRepository) FindDeployment(context.Context, int32, int32) (Deployment, error) {
	return Deployment{}, nil
}
func (f *fakeDeploymentRepository) DeleteDeployment(context.Context, int32, int32) error {
	return nil
}
func (f *fakeDeploymentRepository) ListAPIKeys(context.Context, int32) ([]APIKey, error) {
	return nil, nil
}
func (f *fakeDeploymentRepository) CreateAPIKey(context.Context, APIKey) error { return nil }
func (f *fakeDeploymentRepository) DeleteAPIKey(context.Context, int32, string) error {
	return nil
}

// fakePostDispatcher records post-deployment dispatches.
type fakePostDispatcher struct {
	scrapes      []int32
	resumes      []int32
	resumeDelays []time.Duration
}

func (f *fakePostDispatcher) DispatchPostDeploymentScrape(_ context.Context, environmentID int32, _ time.Duration) error {
	f.scrapes = append(f.scrapes, environmentID)
	return nil
}

func (f *fakePostDispatcher) DispatchUptimeResume(_ context.Context, environmentID int32, delay time.Duration) error {
	f.resumes = append(f.resumes, environmentID)
	f.resumeDelays = append(f.resumeDelays, delay)
	return nil
}

// fakeUptimePauser records pause calls.
type fakeUptimePauser struct {
	paused []int32
}

func (f *fakeUptimePauser) PauseMonitor(_ context.Context, environmentID int32) error {
	f.paused = append(f.paused, environmentID)
	return nil
}

func newCLIService(repo Repository, dispatcher PostDeploymentDispatcher) *Service {
	return NewService(repo, nil, nil, dispatcher, "http://localhost:3000", 5*time.Minute)
}

func validCLICommand() CreateCLICommand {
	return CreateCLICommand{
		Token:         "token",
		EnvironmentID: 7,
		Command:       "bin/deploy.sh",
		ReturnCode:    0,
		StartDate:     time.Now().Add(-time.Minute),
		EndDate:       time.Now(),
	}
}

func TestCreateCLIPausesUptimeAndSchedulesResume(t *testing.T) {
	repo := &fakeDeploymentRepository{
		apiKey:      APIKey{ID: "key-1", ShopID: 3, Scopes: []string{ScopeDeployments}},
		environment: Environment{ID: 7, OrganizationID: "org-1", ShopID: 3},
	}
	dispatcher := &fakePostDispatcher{}
	pauser := &fakeUptimePauser{}
	service := newCLIService(repo, dispatcher).WithUptimePauser(pauser)

	result, err := service.CreateCLI(context.Background(), validCLICommand())
	require.NoError(t, err)
	assert.EqualValues(t, 42, result.DeploymentID)

	assert.Equal(t, []int32{7}, pauser.paused, "monitor must be paused for the settle window")
	assert.Equal(t, []int32{7}, dispatcher.resumes, "a delayed resume must be scheduled")
	require.Len(t, dispatcher.resumeDelays, 1)
	assert.Equal(t, 5*time.Minute, dispatcher.resumeDelays[0], "resume delay must match the scrape settle delay")
	assert.Equal(t, []int32{7}, dispatcher.scrapes)
}

func TestCreateCLIWithoutPauserSkipsUptimeHandling(t *testing.T) {
	repo := &fakeDeploymentRepository{
		apiKey:      APIKey{ID: "key-1", ShopID: 3, Scopes: []string{ScopeDeployments}},
		environment: Environment{ID: 7, OrganizationID: "org-1", ShopID: 3},
	}
	dispatcher := &fakePostDispatcher{}
	service := newCLIService(repo, dispatcher)

	_, err := service.CreateCLI(context.Background(), validCLICommand())
	require.NoError(t, err)

	assert.Equal(t, []int32{7}, dispatcher.scrapes)
	assert.Empty(t, dispatcher.resumes, "no resume when no pauser is attached")
}
