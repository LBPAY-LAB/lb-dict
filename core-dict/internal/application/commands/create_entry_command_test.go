package commands_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/lbpay-lab/core-dict/internal/application/commands"
)

// Mock EntryRepository
type MockEntryRepository struct {
	mock.Mock
}

func (m *MockEntryRepository) Create(ctx context.Context, entry *commands.Entry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *MockEntryRepository) FindByID(ctx context.Context, id uuid.UUID) (*commands.Entry, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*commands.Entry), args.Error(1)
}

func (m *MockEntryRepository) FindByKeyValue(ctx context.Context, keyValue string) (*commands.Entry, error) {
	args := m.Called(ctx, keyValue)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*commands.Entry), args.Error(1)
}

func (m *MockEntryRepository) Update(ctx context.Context, entry *commands.Entry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *MockEntryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEntryRepository) CountByOwnerAndType(ctx context.Context, ownerTaxID string, keyType commands.KeyType) (int, error) {
	args := m.Called(ctx, ownerTaxID, keyType)
	return args.Int(0), args.Error(1)
}

// Mock EventPublisher
type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) Publish(ctx context.Context, event commands.DomainEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// Mock KeyValidatorService
type MockKeyValidatorService struct {
	mock.Mock
}

func (m *MockKeyValidatorService) ValidateFormat(keyType commands.KeyType, keyValue string) error {
	args := m.Called(keyType, keyValue)
	return args.Error(0)
}

func (m *MockKeyValidatorService) ValidateLimits(ctx context.Context, keyType commands.KeyType, ownerTaxID string) error {
	args := m.Called(ctx, keyType, ownerTaxID)
	return args.Error(0)
}

// Mock OwnershipService
type MockOwnershipService struct {
	mock.Mock
}

func (m *MockOwnershipService) ValidateOwnership(ctx context.Context, keyType commands.KeyType, keyValue, ownerTaxID string) error {
	args := m.Called(ctx, keyType, keyValue, ownerTaxID)
	return args.Error(0)
}

// Mock DuplicateCheckerService
type MockDuplicateCheckerService struct {
	mock.Mock
}

func (m *MockDuplicateCheckerService) IsDuplicate(ctx context.Context, keyValue string) (bool, error) {
	args := m.Called(ctx, keyValue)
	return args.Bool(0), args.Error(1)
}

// Mock CacheService
type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) Get(ctx context.Context, key string) (interface{}, error) {
	args := m.Called(ctx, key)
	return args.Get(0), args.Error(1)
}

func (m *MockCacheService) Set(ctx context.Context, key string, value interface{}, ttl interface{}) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheService) InvalidateKey(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheService) InvalidatePattern(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

// Mock ConnectClient
type MockConnectClient struct {
	mock.Mock
}

func (m *MockConnectClient) GetEntryByKey(ctx context.Context, keyValue string) (interface{}, error) {
	args := m.Called(ctx, keyValue)
	return args.Get(0), args.Error(1)
}

func (m *MockConnectClient) CreateEntry(ctx context.Context, keyType, keyValue, accountISPB string) (string, error) {
	args := m.Called(ctx, keyType, keyValue, accountISPB)
	return args.String(0), args.Error(1)
}

func (m *MockConnectClient) UpdateEntry(ctx context.Context, entryID, newAccountISPB string) error {
	args := m.Called(ctx, entryID, newAccountISPB)
	return args.Error(0)
}

func (m *MockConnectClient) DeleteEntry(ctx context.Context, entryID, reason string) error {
	args := m.Called(ctx, entryID, reason)
	return args.Error(0)
}

// Mock EntryEventProducer
type MockEntryEventProducer struct {
	mock.Mock
}

func (m *MockEntryEventProducer) PublishCreated(ctx context.Context, entry interface{}, userID string) error {
	args := m.Called(ctx, entry, userID)
	return args.Error(0)
}

func (m *MockEntryEventProducer) PublishUpdated(ctx context.Context, entry interface{}, userID string) error {
	args := m.Called(ctx, entry, userID)
	return args.Error(0)
}

func (m *MockEntryEventProducer) PublishDeleted(ctx context.Context, entryID, keyValue, reason, userID string) error {
	args := m.Called(ctx, entryID, keyValue, reason, userID)
	return args.Error(0)
}

// Test 1: TestCreateEntryHandler_Success
func TestCreateEntryHandler_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockEntryRepository)
	mockEventPublisher := new(MockEventPublisher)
	mockKeyValidator := new(MockKeyValidatorService)
	mockOwnershipChecker := new(MockOwnershipService)
	mockDuplicateChecker := new(MockDuplicateCheckerService)
	mockCacheService := new(MockCacheService)
	mockConnectClient := new(MockConnectClient)
	mockEntryProducer := new(MockEntryEventProducer)

	mockKeyValidator.On("ValidateFormat", commands.KeyTypeCPF, "12345678901").Return(nil)
	mockOwnershipChecker.On("ValidateOwnership", mock.Anything, commands.KeyTypeCPF, "12345678901", "12345678901").Return(nil)
	mockDuplicateChecker.On("IsDuplicate", mock.Anything, "12345678901").Return(false, nil)
	mockConnectClient.On("GetEntryByKey", mock.Anything, "12345678901").Return(nil, errors.New("not found"))
	mockKeyValidator.On("ValidateLimits", mock.Anything, commands.KeyTypeCPF, "12345678901").Return(nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockEventPublisher.On("Publish", mock.Anything, mock.Anything).Return(nil)
	mockCacheService.On("InvalidateKey", mock.Anything, "entry:12345678901").Return(nil)

	handler := commands.NewCreateEntryCommandHandler(
		mockRepo,
		mockEventPublisher,
		mockKeyValidator,
		mockOwnershipChecker,
		mockDuplicateChecker,
		mockCacheService,
		mockConnectClient,
		mockEntryProducer,
	)

	cmd := commands.CreateEntryCommand{
		KeyType:       commands.KeyTypeCPF,
		KeyValue:      "12345678901",
		AccountID:     uuid.New(),
		AccountISPB:   "12345678",
		AccountBranch: "0001",
		AccountNumber: "123456",
		AccountType:   "CHECKING",
		OwnerType:     "NATURAL_PERSON",
		OwnerTaxID:    "12345678901",
		OwnerName:     "Test User",
		RequestedBy:   uuid.New(),
	}

	// Act
	result, err := handler.Handle(context.Background(), cmd)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, result.EntryID)
	assert.Equal(t, "PENDING", result.Status)
	mockRepo.AssertExpectations(t)
	mockKeyValidator.AssertExpectations(t)
	mockOwnershipChecker.AssertExpectations(t)
	mockDuplicateChecker.AssertExpectations(t)
}

// Test 2: TestCreateEntryHandler_DuplicateKeyLocal
func TestCreateEntryHandler_DuplicateKeyLocal(t *testing.T) {
	// Arrange
	mockRepo := new(MockEntryRepository)
	mockEventPublisher := new(MockEventPublisher)
	mockKeyValidator := new(MockKeyValidatorService)
	mockOwnershipChecker := new(MockOwnershipService)
	mockDuplicateChecker := new(MockDuplicateCheckerService)
	mockCacheService := new(MockCacheService)
	mockConnectClient := new(MockConnectClient)
	mockEntryProducer := new(MockEntryEventProducer)

	mockKeyValidator.On("ValidateFormat", commands.KeyTypeCPF, "12345678901").Return(nil)
	mockOwnershipChecker.On("ValidateOwnership", mock.Anything, commands.KeyTypeCPF, "12345678901", "12345678901").Return(nil)
	mockDuplicateChecker.On("IsDuplicate", mock.Anything, "12345678901").Return(true, nil)

	handler := commands.NewCreateEntryCommandHandler(
		mockRepo,
		mockEventPublisher,
		mockKeyValidator,
		mockOwnershipChecker,
		mockDuplicateChecker,
		mockCacheService,
		mockConnectClient,
		mockEntryProducer,
	)

	cmd := commands.CreateEntryCommand{
		KeyType:    commands.KeyTypeCPF,
		KeyValue:   "12345678901",
		OwnerTaxID: "12345678901",
	}

	// Act
	result, err := handler.Handle(context.Background(), cmd)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "already registered in this PSP")
	mockDuplicateChecker.AssertExpectations(t)
}

// Test 3: TestCreateEntryHandler_DuplicateKeyGlobal
func TestCreateEntryHandler_DuplicateKeyGlobal(t *testing.T) {
	// Arrange
	mockRepo := new(MockEntryRepository)
	mockEventPublisher := new(MockEventPublisher)
	mockKeyValidator := new(MockKeyValidatorService)
	mockOwnershipChecker := new(MockOwnershipService)
	mockDuplicateChecker := new(MockDuplicateCheckerService)
	mockCacheService := new(MockCacheService)
	mockConnectClient := new(MockConnectClient)
	mockEntryProducer := new(MockEntryEventProducer)

	existingEntry := &commands.Entry{ID: uuid.New(), KeyValue: "12345678901"}

	mockKeyValidator.On("ValidateFormat", commands.KeyTypeCPF, "12345678901").Return(nil)
	mockOwnershipChecker.On("ValidateOwnership", mock.Anything, commands.KeyTypeCPF, "12345678901", "12345678901").Return(nil)
	mockDuplicateChecker.On("IsDuplicate", mock.Anything, "12345678901").Return(false, nil)
	mockConnectClient.On("GetEntryByKey", mock.Anything, "12345678901").Return(existingEntry, nil)

	handler := commands.NewCreateEntryCommandHandler(
		mockRepo,
		mockEventPublisher,
		mockKeyValidator,
		mockOwnershipChecker,
		mockDuplicateChecker,
		mockCacheService,
		mockConnectClient,
		mockEntryProducer,
	)

	cmd := commands.CreateEntryCommand{
		KeyType:    commands.KeyTypeCPF,
		KeyValue:   "12345678901",
		OwnerTaxID: "12345678901",
	}

	// Act
	result, err := handler.Handle(context.Background(), cmd)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "already registered in RSFN DICT")
	mockConnectClient.AssertExpectations(t)
}

// Test 4: TestCreateEntryHandler_MaxKeysExceeded
func TestCreateEntryHandler_MaxKeysExceeded(t *testing.T) {
	// Arrange
	mockRepo := new(MockEntryRepository)
	mockEventPublisher := new(MockEventPublisher)
	mockKeyValidator := new(MockKeyValidatorService)
	mockOwnershipChecker := new(MockOwnershipService)
	mockDuplicateChecker := new(MockDuplicateCheckerService)
	mockCacheService := new(MockCacheService)
	mockConnectClient := new(MockConnectClient)
	mockEntryProducer := new(MockEntryEventProducer)

	mockKeyValidator.On("ValidateFormat", commands.KeyTypeCPF, "12345678901").Return(nil)
	mockOwnershipChecker.On("ValidateOwnership", mock.Anything, commands.KeyTypeCPF, "12345678901", "12345678901").Return(nil)
	mockDuplicateChecker.On("IsDuplicate", mock.Anything, "12345678901").Return(false, nil)
	mockConnectClient.On("GetEntryByKey", mock.Anything, "12345678901").Return(nil, errors.New("not found"))
	mockKeyValidator.On("ValidateLimits", mock.Anything, commands.KeyTypeCPF, "12345678901").Return(errors.New("key limit exceeded"))

	handler := commands.NewCreateEntryCommandHandler(
		mockRepo,
		mockEventPublisher,
		mockKeyValidator,
		mockOwnershipChecker,
		mockDuplicateChecker,
		mockCacheService,
		mockConnectClient,
		mockEntryProducer,
	)

	cmd := commands.CreateEntryCommand{
		KeyType:    commands.KeyTypeCPF,
		KeyValue:   "12345678901",
		OwnerTaxID: "12345678901",
	}

	// Act
	result, err := handler.Handle(context.Background(), cmd)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "key limit exceeded")
	mockKeyValidator.AssertExpectations(t)
}

// Test 5: TestCreateEntryHandler_InvalidKeyValue
func TestCreateEntryHandler_InvalidKeyValue(t *testing.T) {
	// Arrange
	mockRepo := new(MockEntryRepository)
	mockEventPublisher := new(MockEventPublisher)
	mockKeyValidator := new(MockKeyValidatorService)
	mockOwnershipChecker := new(MockOwnershipService)
	mockDuplicateChecker := new(MockDuplicateCheckerService)
	mockCacheService := new(MockCacheService)
	mockConnectClient := new(MockConnectClient)
	mockEntryProducer := new(MockEntryEventProducer)

	mockKeyValidator.On("ValidateFormat", commands.KeyTypeCPF, "invalid").Return(errors.New("invalid format"))

	handler := commands.NewCreateEntryCommandHandler(
		mockRepo,
		mockEventPublisher,
		mockKeyValidator,
		mockOwnershipChecker,
		mockDuplicateChecker,
		mockCacheService,
		mockConnectClient,
		mockEntryProducer,
	)

	cmd := commands.CreateEntryCommand{
		KeyType:  commands.KeyTypeCPF,
		KeyValue: "invalid",
	}

	// Act
	result, err := handler.Handle(context.Background(), cmd)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid key format")
	mockKeyValidator.AssertExpectations(t)
}
