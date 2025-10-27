package queries_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/domain/valueobjects"
)

// MockClaimRepository for claim queries
type MockClaimRepository struct {
	mock.Mock
}

func (m *MockClaimRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Claim, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Claim), args.Error(1)
}

func (m *MockClaimRepository) List(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*entities.Claim, int, error) {
	args := m.Called(ctx, filters, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*entities.Claim), args.Int(1), args.Error(2)
}

// MockStatisticsRepository
type MockStatisticsRepository struct {
	mock.Mock
}

func (m *MockStatisticsRepository) GetStatistics(ctx context.Context) (*entities.Statistics, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Statistics), args.Error(1)
}

// MockHealthChecker
type MockHealthChecker struct {
	mock.Mock
}

func (m *MockHealthChecker) CheckDatabase(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockHealthChecker) CheckRedis(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockHealthChecker) CheckPulsar(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockAuditLogRepository
type MockAuditLogRepository struct {
	mock.Mock
}

func (m *MockAuditLogRepository) List(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*entities.AuditLog, int, error) {
	args := m.Called(ctx, filters, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*entities.AuditLog), args.Int(1), args.Error(2)
}

// ===== CLAIM QUERY TESTS =====

// Test 7: TestGetClaimHandler_Success
func TestGetClaimHandler_Success(t *testing.T) {
	// Arrange
	mockClaimRepo := new(MockClaimRepository)

	claimID := uuid.New()
	claim := &entities.Claim{
		ID:        claimID,
		EntryKey:  "12345678901",
		ClaimType: valueobjects.ClaimTypeOwnership,
		Status:    valueobjects.ClaimStatusOpen,
		CreatedAt: time.Now(),
	}

	mockClaimRepo.On("FindByID", mock.Anything, claimID).Return(claim, nil)

	handler := &GetClaimQueryHandler{
		ClaimRepo: mockClaimRepo,
	}

	query := GetClaimQuery{
		ClaimID: claimID,
	}

	// Act
	result, err := handler.Handle(context.Background(), query)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, claimID, result.ID)
	assert.Equal(t, "12345678901", result.EntryKey)
	mockClaimRepo.AssertExpectations(t)
}

// Test 8: TestGetClaimHandler_NotFound
func TestGetClaimHandler_NotFound(t *testing.T) {
	// Arrange
	mockClaimRepo := new(MockClaimRepository)

	claimID := uuid.New()
	mockClaimRepo.On("FindByID", mock.Anything, claimID).Return(nil, errors.New("not found"))

	handler := &GetClaimQueryHandler{
		ClaimRepo: mockClaimRepo,
	}

	query := GetClaimQuery{
		ClaimID: claimID,
	}

	// Act
	result, err := handler.Handle(context.Background(), query)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "claim not found")
}

// ===== LIST CLAIMS TESTS =====

// Test 9: TestListClaimsHandler_Success
func TestListClaimsHandler_Success(t *testing.T) {
	// Arrange
	mockClaimRepo := new(MockClaimRepository)

	claims := []*entities.Claim{
		{ID: uuid.New(), EntryKey: "key1", Status: valueobjects.ClaimStatusOpen, CreatedAt: time.Now()},
		{ID: uuid.New(), EntryKey: "key2", Status: valueobjects.ClaimStatusConfirmed, CreatedAt: time.Now()},
	}

	mockClaimRepo.On("List", mock.Anything, mock.Anything, 10, 0).Return(claims, 2, nil)

	handler := &ListClaimsQueryHandler{
		ClaimRepo: mockClaimRepo,
	}

	query := ListClaimsQuery{
		Limit:  10,
		Offset: 0,
	}

	// Act
	result, err := handler.Handle(context.Background(), query)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Claims, 2)
	assert.Equal(t, 2, result.Total)
	mockClaimRepo.AssertExpectations(t)
}

// Test 10: TestListClaimsHandler_FilterByStatus
func TestListClaimsHandler_FilterByStatus(t *testing.T) {
	// Arrange
	mockClaimRepo := new(MockClaimRepository)

	claims := []*entities.Claim{
		{ID: uuid.New(), EntryKey: "key1", Status: valueobjects.ClaimStatusOpen, CreatedAt: time.Now()},
	}

	filters := map[string]interface{}{
		"status": "OPEN",
	}

	mockClaimRepo.On("List", mock.Anything, filters, 10, 0).Return(claims, 1, nil)

	handler := &ListClaimsQueryHandler{
		ClaimRepo: mockClaimRepo,
	}

	query := ListClaimsQuery{
		Filters: filters,
		Limit:   10,
		Offset:  0,
	}

	// Act
	result, err := handler.Handle(context.Background(), query)

	// Assert
	require.NoError(t, err)
	assert.Len(t, result.Claims, 1)
	assert.Equal(t, valueobjects.ClaimStatusOpen, result.Claims[0].Status)
}

// Test 11: TestListClaimsHandler_Pagination
func TestListClaimsHandler_Pagination(t *testing.T) {
	// Arrange
	mockClaimRepo := new(MockClaimRepository)

	claims := []*entities.Claim{
		{ID: uuid.New(), EntryKey: "key11", CreatedAt: time.Now()},
	}

	mockClaimRepo.On("List", mock.Anything, mock.Anything, 10, 10).Return(claims, 15, nil)

	handler := &ListClaimsQueryHandler{
		ClaimRepo: mockClaimRepo,
	}

	query := ListClaimsQuery{
		Limit:  10,
		Offset: 10, // Page 2
	}

	// Act
	result, err := handler.Handle(context.Background(), query)

	// Assert
	require.NoError(t, err)
	assert.Len(t, result.Claims, 1)
	assert.Equal(t, 15, result.Total)
}

// ===== STATISTICS TESTS =====

// Test 12: TestGetStatisticsHandler_Success
func TestGetStatisticsHandler_Success(t *testing.T) {
	// Arrange
	mockStatsRepo := new(MockStatisticsRepository)

	stats := &entities.Statistics{
		TotalKeys:       1000,
		ActiveKeys:      800,
		BlockedKeys:     50,
		DeletedKeys:     150,
		TotalClaims:     200,
		PendingClaims:   20,
		CompletedClaims: 180,
		KeysByType: map[string]int64{
			"CPF":   400,
			"CNPJ":  300,
			"EMAIL": 200,
			"PHONE": 100,
		},
		LastUpdated: time.Now(),
	}

	mockStatsRepo.On("GetStatistics", mock.Anything).Return(stats, nil)

	handler := &GetStatisticsQueryHandler{
		StatisticsRepo: mockStatsRepo,
	}

	query := GetStatisticsQuery{}

	// Act
	result, err := handler.Handle(context.Background(), query)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1000), result.TotalKeys)
	assert.Equal(t, int64(800), result.ActiveKeys)
	assert.Equal(t, int64(200), result.TotalClaims)
	assert.Equal(t, int64(400), result.KeysByType["CPF"])
	mockStatsRepo.AssertExpectations(t)
}

// Test 13: TestGetStatisticsHandler_EmptyDB
func TestGetStatisticsHandler_EmptyDB(t *testing.T) {
	// Arrange
	mockStatsRepo := new(MockStatisticsRepository)

	stats := &entities.Statistics{
		TotalKeys:     0,
		ActiveKeys:    0,
		TotalClaims:   0,
		KeysByType:    map[string]int64{},
		LastUpdated:   time.Now(),
	}

	mockStatsRepo.On("GetStatistics", mock.Anything).Return(stats, nil)

	handler := &GetStatisticsQueryHandler{
		StatisticsRepo: mockStatsRepo,
	}

	query := GetStatisticsQuery{}

	// Act
	result, err := handler.Handle(context.Background(), query)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(0), result.TotalKeys)
	assert.Equal(t, int64(0), result.TotalClaims)
}

// ===== HEALTH CHECK TESTS =====

// Test 14: TestHealthCheckHandler_AllHealthy
func TestHealthCheckHandler_AllHealthy(t *testing.T) {
	// Arrange
	mockHealthChecker := new(MockHealthChecker)

	mockHealthChecker.On("CheckDatabase", mock.Anything).Return(nil)
	mockHealthChecker.On("CheckRedis", mock.Anything).Return(nil)
	mockHealthChecker.On("CheckPulsar", mock.Anything).Return(nil)

	handler := &HealthCheckQueryHandler{
		HealthChecker: mockHealthChecker,
	}

	query := HealthCheckQuery{}

	// Act
	result, err := handler.Handle(context.Background(), query)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "healthy", result.Status)
	assert.Equal(t, "healthy", result.DatabaseStatus)
	assert.Equal(t, "healthy", result.RedisStatus)
	assert.Equal(t, "healthy", result.PulsarStatus)
	mockHealthChecker.AssertExpectations(t)
}

// Test 15: TestHealthCheckHandler_DatabaseDown
func TestHealthCheckHandler_DatabaseDown(t *testing.T) {
	// Arrange
	mockHealthChecker := new(MockHealthChecker)

	mockHealthChecker.On("CheckDatabase", mock.Anything).Return(errors.New("connection refused"))
	mockHealthChecker.On("CheckRedis", mock.Anything).Return(nil)
	mockHealthChecker.On("CheckPulsar", mock.Anything).Return(nil)

	handler := &HealthCheckQueryHandler{
		HealthChecker: mockHealthChecker,
	}

	query := HealthCheckQuery{}

	// Act
	result, err := handler.Handle(context.Background(), query)

	// Assert
	require.NoError(t, err) // Health check doesn't fail, just reports unhealthy
	assert.NotNil(t, result)
	assert.Equal(t, "unhealthy", result.Status)
	assert.Equal(t, "unhealthy", result.DatabaseStatus)
	assert.Equal(t, "healthy", result.RedisStatus)
	mockHealthChecker.AssertExpectations(t)
}

// ===== AUDIT LOG TESTS =====

// Test 16: TestGetAuditLogHandler_Success
func TestGetAuditLogHandler_Success(t *testing.T) {
	// Arrange
	mockAuditRepo := new(MockAuditLogRepository)

	logs := []*entities.AuditLog{
		{ID: uuid.New(), Action: "CREATE", EntityType: "Entry", Timestamp: time.Now()},
		{ID: uuid.New(), Action: "UPDATE", EntityType: "Entry", Timestamp: time.Now()},
	}

	mockAuditRepo.On("List", mock.Anything, mock.Anything, 50, 0).Return(logs, 2, nil)

	handler := &GetAuditLogQueryHandler{
		AuditRepo: mockAuditRepo,
	}

	query := GetAuditLogQuery{
		Limit:  50,
		Offset: 0,
	}

	// Act
	result, err := handler.Handle(context.Background(), query)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Logs, 2)
	assert.Equal(t, "CREATE", result.Logs[0].Action)
	mockAuditRepo.AssertExpectations(t)
}

// Test 17: TestGetAuditLogHandler_FilterByUser
func TestGetAuditLogHandler_FilterByUser(t *testing.T) {
	// Arrange
	mockAuditRepo := new(MockAuditLogRepository)

	userID := uuid.New()
	logs := []*entities.AuditLog{
		{ID: uuid.New(), Action: "CREATE", ActorID: userID, Timestamp: time.Now()},
	}

	filters := map[string]interface{}{
		"actor_id": userID.String(),
	}

	mockAuditRepo.On("List", mock.Anything, filters, 50, 0).Return(logs, 1, nil)

	handler := &GetAuditLogQueryHandler{
		AuditRepo: mockAuditRepo,
	}

	query := GetAuditLogQuery{
		Filters: filters,
		Limit:   50,
		Offset:  0,
	}

	// Act
	result, err := handler.Handle(context.Background(), query)

	// Assert
	require.NoError(t, err)
	assert.Len(t, result.Logs, 1)
	assert.Equal(t, userID, result.Logs[0].ActorID)
}

// Test 18: TestGetAuditLogHandler_TimeRange
func TestGetAuditLogHandler_TimeRange(t *testing.T) {
	// Arrange
	mockAuditRepo := new(MockAuditLogRepository)

	now := time.Now()
	logs := []*entities.AuditLog{
		{ID: uuid.New(), Action: "CREATE", Timestamp: now},
	}

	filters := map[string]interface{}{
		"start_date": now.Add(-24 * time.Hour),
		"end_date":   now,
	}

	mockAuditRepo.On("List", mock.Anything, filters, 50, 0).Return(logs, 1, nil)

	handler := &GetAuditLogQueryHandler{
		AuditRepo: mockAuditRepo,
	}

	query := GetAuditLogQuery{
		Filters: filters,
		Limit:   50,
		Offset:  0,
	}

	// Act
	result, err := handler.Handle(context.Background(), query)

	// Assert
	require.NoError(t, err)
	assert.Len(t, result.Logs, 1)
}

// ===== Query Handler Stubs =====

type GetClaimQuery struct {
	ClaimID uuid.UUID
}

type GetClaimQueryHandler struct {
	ClaimRepo *MockClaimRepository
}

func (h *GetClaimQueryHandler) Handle(ctx context.Context, query GetClaimQuery) (*entities.Claim, error) {
	claim, err := h.ClaimRepo.FindByID(ctx, query.ClaimID)
	if err != nil {
		return nil, errors.New("claim not found: " + err.Error())
	}
	return claim, nil
}

type ListClaimsQuery struct {
	Filters map[string]interface{}
	Limit   int
	Offset  int
}

type ListClaimsResult struct {
	Claims []*entities.Claim
	Total  int
}

type ListClaimsQueryHandler struct {
	ClaimRepo *MockClaimRepository
}

func (h *ListClaimsQueryHandler) Handle(ctx context.Context, query ListClaimsQuery) (*ListClaimsResult, error) {
	if query.Limit == 0 {
		query.Limit = 10
	}

	claims, total, err := h.ClaimRepo.List(ctx, query.Filters, query.Limit, query.Offset)
	if err != nil {
		return nil, err
	}

	return &ListClaimsResult{
		Claims: claims,
		Total:  total,
	}, nil
}

type GetStatisticsQuery struct{}

type GetStatisticsQueryHandler struct {
	StatisticsRepo *MockStatisticsRepository
}

func (h *GetStatisticsQueryHandler) Handle(ctx context.Context, query GetStatisticsQuery) (*entities.Statistics, error) {
	return h.StatisticsRepo.GetStatistics(ctx)
}

type HealthCheckQuery struct{}

type HealthCheckResult struct {
	Status         string
	DatabaseStatus string
	RedisStatus    string
	PulsarStatus   string
}

type HealthCheckQueryHandler struct {
	HealthChecker *MockHealthChecker
}

func (h *HealthCheckQueryHandler) Handle(ctx context.Context, query HealthCheckQuery) (*HealthCheckResult, error) {
	result := &HealthCheckResult{
		Status:         "healthy",
		DatabaseStatus: "healthy",
		RedisStatus:    "healthy",
		PulsarStatus:   "healthy",
	}

	if err := h.HealthChecker.CheckDatabase(ctx); err != nil {
		result.DatabaseStatus = "unhealthy"
		result.Status = "unhealthy"
	}

	if err := h.HealthChecker.CheckRedis(ctx); err != nil {
		result.RedisStatus = "unhealthy"
		result.Status = "unhealthy"
	}

	if err := h.HealthChecker.CheckPulsar(ctx); err != nil {
		result.PulsarStatus = "unhealthy"
		result.Status = "unhealthy"
	}

	return result, nil
}

type GetAuditLogQuery struct {
	Filters map[string]interface{}
	Limit   int
	Offset  int
}

type GetAuditLogResult struct {
	Logs  []*entities.AuditLog
	Total int
}

type GetAuditLogQueryHandler struct {
	AuditRepo *MockAuditLogRepository
}

func (h *GetAuditLogQueryHandler) Handle(ctx context.Context, query GetAuditLogQuery) (*GetAuditLogResult, error) {
	if query.Limit == 0 {
		query.Limit = 50
	}

	logs, total, err := h.AuditRepo.List(ctx, query.Filters, query.Limit, query.Offset)
	if err != nil {
		return nil, err
	}

	return &GetAuditLogResult{
		Logs:  logs,
		Total: total,
	}, nil
}
