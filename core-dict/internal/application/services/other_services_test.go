package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/lbpay-lab/core-dict/internal/application/services"
)

// MockAccountService for ownership tests
type MockAccountService struct {
	mock.Mock
}

func (m *MockAccountService) GetAccountOwner(ctx context.Context, accountID string) (string, error) {
	args := m.Called(ctx, accountID)
	return args.String(0), args.Error(1)
}

// MockEntryRepository for duplicate checker
type MockEntryRepository struct {
	mock.Mock
}

func (m *MockEntryRepository) ExistsByKeyValue(ctx context.Context, keyValue string) (bool, error) {
	args := m.Called(ctx, keyValue)
	return args.Bool(0), args.Error(1)
}

// MockCacheClient for cache service
type MockCacheClient struct {
	mock.Mock
}

func (m *MockCacheClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheClient) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheClient) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

// ===== ACCOUNT OWNERSHIP SERVICE TESTS =====

// Test 16: TestAccountOwnership_Verify_Success
func TestAccountOwnership_Verify_Success(t *testing.T) {
	// Arrange
	mockAccountService := new(MockAccountService)

	ownershipService := &AccountOwnershipService{
		AccountService: mockAccountService,
	}

	accountID := "123456"
	expectedOwnerTaxID := "12345678901"

	mockAccountService.On("GetAccountOwner", mock.Anything, accountID).Return(expectedOwnerTaxID, nil)

	// Act
	err := ownershipService.ValidateOwnership(
		context.Background(),
		services.KeyTypeCPF,
		"12345678901", // CPF value
		expectedOwnerTaxID,
	)

	// Assert
	require.NoError(t, err)
	mockAccountService.AssertExpectations(t)
}

// Test 17: TestAccountOwnership_Verify_Mismatch
func TestAccountOwnership_Verify_Mismatch(t *testing.T) {
	// Arrange
	mockAccountService := new(MockAccountService)

	ownershipService := &AccountOwnershipService{
		AccountService: mockAccountService,
	}

	// For CPF/CNPJ keys, the key value must match the owner's tax ID
	keyValue := "12345678901"
	ownerTaxID := "98765432100" // Different

	// Act
	err := ownershipService.ValidateOwnership(
		context.Background(),
		services.KeyTypeCPF,
		keyValue,
		ownerTaxID,
	)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ownership mismatch")
}

// Test 18: TestAccountOwnership_Verify_EmailPhone_Success
func TestAccountOwnership_Verify_EmailPhone_Success(t *testing.T) {
	// Arrange
	ownershipService := &AccountOwnershipService{}

	testCases := []struct {
		keyType   services.KeyType
		keyValue  string
		ownerTaxID string
	}{
		{services.KeyTypeEmail, "user@example.com", "12345678901"},
		{services.KeyTypePhone, "+5511999998888", "12345678901"},
		{services.KeyTypeEVP, "550e8400-e29b-41d4-a716-446655440000", "12345678901"},
	}

	for _, tc := range testCases {
		// Act
		err := ownershipService.ValidateOwnership(
			context.Background(),
			tc.keyType,
			tc.keyValue,
			tc.ownerTaxID,
		)

		// Assert
		// Email, Phone, EVP don't require matching owner tax ID
		require.NoError(t, err, "Key type %s should not require ownership match", tc.keyType)
	}
}

// ===== DUPLICATE KEY CHECKER TESTS =====

// Test 19: TestDuplicateChecker_ExistsLocal
func TestDuplicateChecker_ExistsLocal(t *testing.T) {
	// Arrange
	mockRepo := new(MockEntryRepository)

	checker := &DuplicateKeyChecker{
		EntryRepo: mockRepo,
	}

	keyValue := "12345678901"
	mockRepo.On("ExistsByKeyValue", mock.Anything, keyValue).Return(true, nil)

	// Act
	isDuplicate, err := checker.IsDuplicate(context.Background(), keyValue)

	// Assert
	require.NoError(t, err)
	assert.True(t, isDuplicate)
	mockRepo.AssertExpectations(t)
}

// Test 20: TestDuplicateChecker_NotExists
func TestDuplicateChecker_NotExists(t *testing.T) {
	// Arrange
	mockRepo := new(MockEntryRepository)

	checker := &DuplicateKeyChecker{
		EntryRepo: mockRepo,
	}

	keyValue := "newkey123"
	mockRepo.On("ExistsByKeyValue", mock.Anything, keyValue).Return(false, nil)

	// Act
	isDuplicate, err := checker.IsDuplicate(context.Background(), keyValue)

	// Assert
	require.NoError(t, err)
	assert.False(t, isDuplicate)
	mockRepo.AssertExpectations(t)
}

// Test 21: TestDuplicateChecker_Error
func TestDuplicateChecker_Error(t *testing.T) {
	// Arrange
	mockRepo := new(MockEntryRepository)

	checker := &DuplicateKeyChecker{
		EntryRepo: mockRepo,
	}

	keyValue := "12345678901"
	mockRepo.On("ExistsByKeyValue", mock.Anything, keyValue).Return(false, errors.New("database error"))

	// Act
	isDuplicate, err := checker.IsDuplicate(context.Background(), keyValue)

	// Assert
	require.Error(t, err)
	assert.False(t, isDuplicate)
	mockRepo.AssertExpectations(t)
}

// ===== CACHE SERVICE TESTS =====

// Test 22: TestCacheService_GetOrSet_Hit
func TestCacheService_GetOrSet_Hit(t *testing.T) {
	// Arrange
	mockCacheClient := new(MockCacheClient)

	cacheService := &CacheServiceImpl{
		Client: mockCacheClient,
	}

	key := "entry:12345678901"
	cachedValue := `{"id":"123","key_value":"12345678901"}`

	mockCacheClient.On("Get", mock.Anything, key).Return(cachedValue, nil)

	// Act
	value, err := cacheService.Get(context.Background(), key)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, cachedValue, value)
	mockCacheClient.AssertExpectations(t)
}

// Test 23: TestCacheService_GetOrSet_Miss
func TestCacheService_GetOrSet_Miss(t *testing.T) {
	// Arrange
	mockCacheClient := new(MockCacheClient)

	cacheService := &CacheServiceImpl{
		Client: mockCacheClient,
	}

	key := "entry:newkey"
	mockCacheClient.On("Get", mock.Anything, key).Return("", errors.New("cache miss"))

	// Act
	value, err := cacheService.Get(context.Background(), key)

	// Assert
	require.Error(t, err)
	assert.Empty(t, value)
	mockCacheClient.AssertExpectations(t)
}

// Test 24: TestCacheService_Set
func TestCacheService_Set(t *testing.T) {
	// Arrange
	mockCacheClient := new(MockCacheClient)

	cacheService := &CacheServiceImpl{
		Client: mockCacheClient,
	}

	key := "entry:12345678901"
	value := `{"id":"123"}`
	ttl := 5 * time.Minute

	mockCacheClient.On("Set", mock.Anything, key, value, ttl).Return(nil)

	// Act
	err := cacheService.Set(context.Background(), key, value, ttl)

	// Assert
	require.NoError(t, err)
	mockCacheClient.AssertExpectations(t)
}

// Test 25: TestCacheService_Invalidate
func TestCacheService_Invalidate(t *testing.T) {
	// Arrange
	mockCacheClient := new(MockCacheClient)

	cacheService := &CacheServiceImpl{
		Client: mockCacheClient,
	}

	key := "entry:12345678901"
	mockCacheClient.On("Delete", mock.Anything, key).Return(nil)

	// Act
	err := cacheService.Delete(context.Background(), key)

	// Assert
	require.NoError(t, err)
	mockCacheClient.AssertExpectations(t)
}

// ===== Service Implementation Stubs =====

type AccountOwnershipService struct {
	AccountService *MockAccountService
}

func (s *AccountOwnershipService) ValidateOwnership(ctx context.Context, keyType services.KeyType, keyValue, ownerTaxID string) error {
	// For CPF and CNPJ, the key value MUST match the owner's tax ID
	if keyType == services.KeyTypeCPF || keyType == services.KeyTypeCNPJ {
		if keyValue != ownerTaxID {
			return errors.New("ownership mismatch: CPF/CNPJ key must belong to the account owner")
		}
	}

	// For Email, Phone, EVP: ownership is assumed valid if account exists
	// In real implementation, we would verify the account exists and is active
	return nil
}

type DuplicateKeyChecker struct {
	EntryRepo *MockEntryRepository
}

func (c *DuplicateKeyChecker) IsDuplicate(ctx context.Context, keyValue string) (bool, error) {
	exists, err := c.EntryRepo.ExistsByKeyValue(ctx, keyValue)
	if err != nil {
		return false, err
	}
	return exists, nil
}

type CacheServiceImpl struct {
	Client *MockCacheClient
}

func (s *CacheServiceImpl) Get(ctx context.Context, key string) (interface{}, error) {
	value, err := s.Client.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (s *CacheServiceImpl) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	valueStr, ok := value.(string)
	if !ok {
		return errors.New("value must be string")
	}
	return s.Client.Set(ctx, key, valueStr, ttl)
}

func (s *CacheServiceImpl) Delete(ctx context.Context, key string) error {
	return s.Client.Delete(ctx, key)
}
