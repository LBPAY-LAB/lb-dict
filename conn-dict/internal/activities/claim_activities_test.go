package activities

import (
	"context"
	"errors"
	"testing"

	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/pulsar"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockClaimRepository is a mock implementation of ClaimRepository
type MockClaimRepository struct {
	mock.Mock
}

func (m *MockClaimRepository) Create(ctx context.Context, claim *entities.Claim) error {
	args := m.Called(ctx, claim)
	return args.Error(0)
}

func (m *MockClaimRepository) Update(ctx context.Context, claim *entities.Claim) error {
	args := m.Called(ctx, claim)
	return args.Error(0)
}

func (m *MockClaimRepository) GetByClaimID(ctx context.Context, claimID string) (*entities.Claim, error) {
	args := m.Called(ctx, claimID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Claim), args.Error(1)
}

func (m *MockClaimRepository) HasActiveClaim(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockClaimRepository) GetByKey(ctx context.Context, key string) ([]*entities.Claim, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Claim), args.Error(1)
}

// MockPulsarProducer is a mock implementation of Pulsar Producer
type MockPulsarProducer struct {
	mock.Mock
}

func (m *MockPulsarProducer) PublishEvent(ctx context.Context, event map[string]interface{}, key string) error {
	args := m.Called(ctx, event, key)
	return args.Error(0)
}

func (m *MockPulsarProducer) Close() {
	m.Called()
}

func TestNewClaimActivities(t *testing.T) {
	logger := logrus.New()
	repo := new(MockClaimRepository)
	producer := new(MockPulsarProducer)

	activities := NewClaimActivities(logger, repo, producer)

	require.NotNil(t, activities)
	assert.Equal(t, logger, activities.logger)
	assert.Equal(t, repo, activities.claimRepo)
	assert.Equal(t, producer, activities.pulsarProducer)
}

func TestCreateClaimActivity_Success(t *testing.T) {
	logger := logrus.New()
	repo := new(MockClaimRepository)
	producer := new(MockPulsarProducer)

	activities := NewClaimActivities(logger, repo, producer)

	input := CreateClaimInput{
		ClaimID:              "claim-123",
		Type:                 "PORTABILITY",
		Key:                  "12345678901",
		KeyType:              "CPF",
		DonorISPB:            "60701190",
		ClaimerISPB:          "60746948",
		ClaimerAccountBranch: "0001",
		ClaimerAccountNumber: "123456",
		ClaimerAccountType:   "CHECKING",
	}

	// Setup mocks
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Claim")).Return(nil)
	producer.On("PublishEvent", mock.Anything, mock.AnythingOfType("map[string]interface {}"), mock.Anything).Return(nil)

	// Execute
	ctx := context.Background()
	err := activities.CreateClaimActivity(ctx, input)

	// Assert
	require.NoError(t, err)
	repo.AssertExpectations(t)
	producer.AssertExpectations(t)
}

func TestCreateClaimActivity_InvalidType(t *testing.T) {
	logger := logrus.New()
	repo := new(MockClaimRepository)
	producer := new(MockPulsarProducer)

	activities := NewClaimActivities(logger, repo, producer)

	input := CreateClaimInput{
		ClaimID:     "claim-123",
		Type:        "INVALID_TYPE",
		Key:         "12345678901",
		KeyType:     "CPF",
		DonorISPB:   "60701190",
		ClaimerISPB: "60746948",
	}

	// Execute
	ctx := context.Background()
	err := activities.CreateClaimActivity(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid claim type")
}

func TestCreateClaimActivity_RepositoryError(t *testing.T) {
	logger := logrus.New()
	repo := new(MockClaimRepository)
	producer := new(MockPulsarProducer)

	activities := NewClaimActivities(logger, repo, producer)

	input := CreateClaimInput{
		ClaimID:     "claim-123",
		Type:        "PORTABILITY",
		Key:         "12345678901",
		KeyType:     "CPF",
		DonorISPB:   "60701190",
		ClaimerISPB: "60746948",
	}

	// Setup mocks
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Claim")).
		Return(errors.New("database error"))

	// Execute
	ctx := context.Background()
	err := activities.CreateClaimActivity(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	repo.AssertExpectations(t)
}

func TestGetClaimStatusActivity_Success(t *testing.T) {
	logger := logrus.New()
	repo := new(MockClaimRepository)
	producer := new(MockPulsarProducer)

	activities := NewClaimActivities(logger, repo, producer)

	claim := &entities.Claim{
		ClaimID: "claim-123",
		Status:  entities.ClaimStatusOpen,
	}

	// Setup mocks
	repo.On("GetByClaimID", mock.Anything, "claim-123").Return(claim, nil)

	// Execute
	ctx := context.Background()
	status, err := activities.GetClaimStatusActivity(ctx, "claim-123")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, string(entities.ClaimStatusOpen), status)
	repo.AssertExpectations(t)
}

func TestCompleteClaimActivity_Success(t *testing.T) {
	logger := logrus.New()
	repo := new(MockClaimRepository)
	producer := new(MockPulsarProducer)

	activities := NewClaimActivities(logger, repo, producer)

	claim, _ := entities.NewClaim(
		"claim-123",
		entities.ClaimTypePortability,
		"12345678901",
		"CPF",
		"60701190",
		"60746948",
	)
	claim.Confirm() // Move to confirmed state first

	// Setup mocks
	repo.On("GetByClaimID", mock.Anything, "claim-123").Return(claim, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Claim")).Return(nil)
	producer.On("PublishEvent", mock.Anything, mock.AnythingOfType("map[string]interface {}"), mock.Anything).Return(nil)

	// Execute
	ctx := context.Background()
	err := activities.CompleteClaimActivity(ctx, "claim-123")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, entities.ClaimStatusCompleted, claim.Status)
	repo.AssertExpectations(t)
	producer.AssertExpectations(t)
}

func TestCancelClaimActivity_Success(t *testing.T) {
	logger := logrus.New()
	repo := new(MockClaimRepository)
	producer := new(MockPulsarProducer)

	activities := NewClaimActivities(logger, repo, producer)

	claim, _ := entities.NewClaim(
		"claim-123",
		entities.ClaimTypePortability,
		"12345678901",
		"CPF",
		"60701190",
		"60746948",
	)

	// Setup mocks
	repo.On("GetByClaimID", mock.Anything, "claim-123").Return(claim, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Claim")).Return(nil)
	producer.On("PublishEvent", mock.Anything, mock.AnythingOfType("map[string]interface {}"), mock.Anything).Return(nil)

	// Execute
	ctx := context.Background()
	err := activities.CancelClaimActivity(ctx, "claim-123", "user requested")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, entities.ClaimStatusCancelled, claim.Status)
	repo.AssertExpectations(t)
	producer.AssertExpectations(t)
}

func TestExpireClaimActivity_Success(t *testing.T) {
	logger := logrus.New()
	repo := new(MockClaimRepository)
	producer := new(MockPulsarProducer)

	activities := NewClaimActivities(logger, repo, producer)

	claim, _ := entities.NewClaim(
		"claim-123",
		entities.ClaimTypePortability,
		"12345678901",
		"CPF",
		"60701190",
		"60746948",
	)

	// Setup mocks
	repo.On("GetByClaimID", mock.Anything, "claim-123").Return(claim, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Claim")).Return(nil)
	producer.On("PublishEvent", mock.Anything, mock.AnythingOfType("map[string]interface {}"), mock.Anything).Return(nil)

	// Execute
	ctx := context.Background()
	err := activities.ExpireClaimActivity(ctx, "claim-123")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, entities.ClaimStatusExpired, claim.Status)
	repo.AssertExpectations(t)
	producer.AssertExpectations(t)
}

func TestValidateClaimEligibilityActivity_Success(t *testing.T) {
	logger := logrus.New()
	repo := new(MockClaimRepository)
	producer := new(MockPulsarProducer)

	activities := NewClaimActivities(logger, repo, producer)

	// Setup mocks - no active claim exists
	repo.On("HasActiveClaim", mock.Anything, "12345678901").Return(false, nil)

	// Execute
	ctx := context.Background()
	err := activities.ValidateClaimEligibilityActivity(ctx, "12345678901")

	// Assert
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestValidateClaimEligibilityActivity_AlreadyHasClaim(t *testing.T) {
	logger := logrus.New()
	repo := new(MockClaimRepository)
	producer := new(MockPulsarProducer)

	activities := NewClaimActivities(logger, repo, producer)

	// Setup mocks - active claim exists
	repo.On("HasActiveClaim", mock.Anything, "12345678901").Return(true, nil)

	// Execute
	ctx := context.Background()
	err := activities.ValidateClaimEligibilityActivity(ctx, "12345678901")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already has an active claim")
	repo.AssertExpectations(t)
}

func TestClaimActivitiesTableDriven(t *testing.T) {
	tests := []struct {
		name           string
		activity       func(*ClaimActivities, context.Context) error
		setupMocks     func(*MockClaimRepository, *MockPulsarProducer)
		wantErr        bool
		wantErrContain string
	}{
		{
			name: "notify donor success",
			activity: func(a *ClaimActivities, ctx context.Context) error {
				return a.NotifyDonorActivity(ctx, "claim-123")
			},
			setupMocks: func(repo *MockClaimRepository, producer *MockPulsarProducer) {
				claim, _ := entities.NewClaim("claim-123", entities.ClaimTypePortability, "key", "CPF", "donor", "claimer")
				repo.On("GetByClaimID", mock.Anything, "claim-123").Return(claim, nil)
				producer.On("PublishEvent", mock.Anything, mock.AnythingOfType("map[string]interface {}"), mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "notify donor - claim not found",
			activity: func(a *ClaimActivities, ctx context.Context) error {
				return a.NotifyDonorActivity(ctx, "claim-404")
			},
			setupMocks: func(repo *MockClaimRepository, producer *MockPulsarProducer) {
				repo.On("GetByClaimID", mock.Anything, "claim-404").Return(nil, errors.New("not found"))
			},
			wantErr:        true,
			wantErrContain: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logrus.New()
			repo := new(MockClaimRepository)
			producer := new(MockPulsarProducer)

			activities := NewClaimActivities(logger, repo, producer)

			if tt.setupMocks != nil {
				tt.setupMocks(repo, producer)
			}

			ctx := context.Background()
			err := tt.activity(activities, ctx)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrContain != "" {
					assert.Contains(t, err.Error(), tt.wantErrContain)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
