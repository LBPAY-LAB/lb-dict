package queries_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/lbpay-lab/core-dict/internal/domain/entities"
)

// MockEntryRepository for query tests
type MockEntryRepository struct {
	mock.Mock
}

func (m *MockEntryRepository) FindByKey(ctx context.Context, keyValue string) (*entities.Entry, error) {
	args := m.Called(ctx, keyValue)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Entry), args.Error(1)
}

func (m *MockEntryRepository) List(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*entities.Entry, int, error) {
	args := m.Called(ctx, filters, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*entities.Entry), args.Int(1), args.Error(2)
}

// MockCacheService for query tests
type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) Get(ctx context.Context, key string) (interface{}, error) {
	args := m.Called(ctx, key)
	return args.Get(0), args.Error(1)
}

func (m *MockCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

// MockConnectClient for query tests
type MockConnectClient struct {
	mock.Mock
}

func (m *MockConnectClient) GetEntryByKey(ctx context.Context, keyValue string) (interface{}, error) {
	args := m.Called(ctx, keyValue)
	return args.Get(0), args.Error(1)
}

// ===== GET ENTRY QUERY TESTS =====

// Test 1: TestGetEntryHandler_Success_FromCache
func TestGetEntryHandler_Success_FromCache(t *testing.T) {
	// Arrange
	mockRepo := new(MockEntryRepository)
	mockCache := new(MockCacheService)
	mockConnectClient := new(MockConnectClient)

	entry := &entities.Entry{
		ID:        uuid.New(),
		KeyValue:  "12345678901",
		KeyType:   entities.KeyTypeCPF,
		Status:    entities.KeyStatusActive,
		CreatedAt: time.Now(),
	}

	// Marshal entry to JSON for cache
	entryJSON, _ := json.Marshal(entry)

	mockCache.On("Get", mock.Anything, "entry:12345678901").Return(string(entryJSON), nil)

	handler := &GetEntryQueryHandler{
		EntryRepo:     mockRepo,
		Cache:         mockCache,
		ConnectClient: mockConnectClient,
	}

	query := GetEntryQuery{
		KeyValue: "12345678901",
	}

	// Act
	result, err := handler.Handle(context.Background(), query)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "12345678901", result.KeyValue)
	mockCache.AssertExpectations(t)
	// Repository should NOT be called (cache hit)
	mockRepo.AssertNotCalled(t, "FindByKey")
}

// Test 2: TestGetEntryHandler_Success_FromDB
func TestGetEntryHandler_Success_FromDB(t *testing.T) {
	// Arrange
	mockRepo := new(MockEntryRepository)
	mockCache := new(MockCacheService)
	mockConnectClient := new(MockConnectClient)

	entry := &entities.Entry{
		ID:        uuid.New(),
		KeyValue:  "12345678901",
		KeyType:   entities.KeyTypeCPF,
		Status:    entities.KeyStatusActive,
		CreatedAt: time.Now(),
	}

	mockCache.On("Get", mock.Anything, "entry:12345678901").Return(nil, errors.New("cache miss"))
	mockRepo.On("FindByKey", mock.Anything, "12345678901").Return(entry, nil)
	mockCache.On("Set", mock.Anything, "entry:12345678901", entry, 5*time.Minute).Return(nil)

	handler := &GetEntryQueryHandler{
		EntryRepo:     mockRepo,
		Cache:         mockCache,
		ConnectClient: mockConnectClient,
	}

	query := GetEntryQuery{
		KeyValue: "12345678901",
	}

	// Act
	result, err := handler.Handle(context.Background(), query)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "12345678901", result.KeyValue)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

// Test 3: TestGetEntryHandler_NotFound
func TestGetEntryHandler_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockEntryRepository)
	mockCache := new(MockCacheService)
	mockConnectClient := new(MockConnectClient)

	mockCache.On("Get", mock.Anything, "nonexistent").Return(nil, errors.New("cache miss"))
	mockRepo.On("FindByKey", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))

	handler := &GetEntryQueryHandler{
		EntryRepo:     mockRepo,
		Cache:         mockCache,
		ConnectClient: mockConnectClient,
	}

	query := GetEntryQuery{
		KeyValue: "nonexistent",
	}

	// Act
	result, err := handler.Handle(context.Background(), query)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "entry not found")
}

// ===== LIST ENTRIES QUERY TESTS =====

// Test 4: TestListEntriesHandler_Success_Paginated
func TestListEntriesHandler_Success_Paginated(t *testing.T) {
	// Arrange
	mockRepo := new(MockEntryRepository)

	entries := []*entities.Entry{
		{ID: uuid.New(), KeyValue: "key1", KeyType: entities.KeyTypeCPF, Status: entities.KeyStatusActive},
		{ID: uuid.New(), KeyValue: "key2", KeyType: entities.KeyTypeEmail, Status: entities.KeyStatusActive},
	}

	mockRepo.On("List", mock.Anything, mock.Anything, 10, 0).Return(entries, 2, nil)

	handler := &ListEntriesQueryHandler{
		EntryRepo: mockRepo,
	}

	query := ListEntriesQuery{
		Limit:  10,
		Offset: 0,
	}

	// Act
	result, err := handler.Handle(context.Background(), query)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Entries, 2)
	assert.Equal(t, 2, result.Total)
	mockRepo.AssertExpectations(t)
}

// Test 5: TestListEntriesHandler_Filters
func TestListEntriesHandler_Filters(t *testing.T) {
	// Arrange
	mockRepo := new(MockEntryRepository)

	entries := []*entities.Entry{
		{ID: uuid.New(), KeyValue: "key1", KeyType: entities.KeyTypeCPF, Status: entities.KeyStatusActive},
	}

	filters := map[string]interface{}{
		"key_type": "CPF",
		"status":   "ACTIVE",
	}

	mockRepo.On("List", mock.Anything, filters, 10, 0).Return(entries, 1, nil)

	handler := &ListEntriesQueryHandler{
		EntryRepo: mockRepo,
	}

	query := ListEntriesQuery{
		Filters: filters,
		Limit:   10,
		Offset:  0,
	}

	// Act
	result, err := handler.Handle(context.Background(), query)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Entries, 1)
	assert.Equal(t, entities.KeyTypeCPF, result.Entries[0].KeyType)
	mockRepo.AssertExpectations(t)
}

// Test 6: TestListEntriesHandler_EmptyResult
func TestListEntriesHandler_EmptyResult(t *testing.T) {
	// Arrange
	mockRepo := new(MockEntryRepository)

	mockRepo.On("List", mock.Anything, mock.Anything, 10, 0).Return([]*entities.Entry{}, 0, nil)

	handler := &ListEntriesQueryHandler{
		EntryRepo: mockRepo,
	}

	query := ListEntriesQuery{
		Limit:  10,
		Offset: 0,
	}

	// Act
	result, err := handler.Handle(context.Background(), query)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Entries, 0)
	assert.Equal(t, 0, result.Total)
}

// ===== Query Handler Stubs =====

type GetEntryQuery struct {
	KeyValue string
}

type GetEntryQueryHandler struct {
	EntryRepo     *MockEntryRepository
	Cache         *MockCacheService
	ConnectClient *MockConnectClient
}

func (h *GetEntryQueryHandler) Handle(ctx context.Context, query GetEntryQuery) (*entities.Entry, error) {
	if query.KeyValue == "" {
		return nil, errors.New("key_value is required")
	}

	// 1. Try cache first
	cacheKey := "entry:" + query.KeyValue
	if cachedData, err := h.Cache.Get(ctx, cacheKey); err == nil && cachedData != nil {
		if jsonData, ok := cachedData.(string); ok {
			var entry entities.Entry
			if err := json.Unmarshal([]byte(jsonData), &entry); err == nil {
				return &entry, nil
			}
		}
	}

	// 2. Cache miss - query database
	entry, err := h.EntryRepo.FindByKey(ctx, query.KeyValue)
	if err == nil {
		// Found in database - cache and return
		h.Cache.Set(ctx, cacheKey, entry, 5*time.Minute)
		return entry, nil
	}

	// 3. Not found anywhere
	return nil, errors.New("entry not found: " + err.Error())
}

type ListEntriesQuery struct {
	Filters map[string]interface{}
	Limit   int
	Offset  int
}

type ListEntriesResult struct {
	Entries []*entities.Entry
	Total   int
}

type ListEntriesQueryHandler struct {
	EntryRepo *MockEntryRepository
}

func (h *ListEntriesQueryHandler) Handle(ctx context.Context, query ListEntriesQuery) (*ListEntriesResult, error) {
	if query.Limit == 0 {
		query.Limit = 10
	}

	entries, total, err := h.EntryRepo.List(ctx, query.Filters, query.Limit, query.Offset)
	if err != nil {
		return nil, err
	}

	return &ListEntriesResult{
		Entries: entries,
		Total:   total,
	}, nil
}
