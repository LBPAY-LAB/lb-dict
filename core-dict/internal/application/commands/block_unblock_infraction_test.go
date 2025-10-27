package commands_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/lbpay-lab/core-dict/internal/application/commands"
)

// ===== BLOCK ENTRY TESTS =====

// Test 1: TestBlockEntryHandler_Success
func TestBlockEntryHandler_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockEntryRepository)
	mockEventPublisher := new(MockEventPublisher)
	mockCacheService := new(MockCacheService)

	entryID := uuid.New()
	entry := &commands.Entry{
		ID:        entryID,
		KeyValue:  "12345678901",
		Status:    "ACTIVE",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("FindByID", mock.Anything, entryID).Return(entry, nil)
	mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
	mockEventPublisher.On("Publish", mock.Anything, mock.Anything).Return(nil)
	mockCacheService.On("InvalidateKey", mock.Anything, "entry:12345678901").Return(nil)

	handler := &BlockEntryCommandHandler{
		EntryRepo:      mockRepo,
		EventPublisher: mockEventPublisher,
		CacheService:   mockCacheService,
	}

	cmd := BlockEntryCommand{
		EntryID:   entryID,
		Reason:    "Suspicious activity",
		BlockedBy: uuid.New(),
	}

	// Act
	err := handler.Handle(context.Background(), cmd)

	// Assert
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockEventPublisher.AssertExpectations(t)
	mockCacheService.AssertExpectations(t)
}

// Test 2: TestBlockEntryHandler_NotFound
func TestBlockEntryHandler_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockEntryRepository)
	mockEventPublisher := new(MockEventPublisher)
	mockCacheService := new(MockCacheService)

	entryID := uuid.New()
	mockRepo.On("FindByID", mock.Anything, entryID).Return(nil, errors.New("not found"))

	handler := &BlockEntryCommandHandler{
		EntryRepo:      mockRepo,
		EventPublisher: mockEventPublisher,
		CacheService:   mockCacheService,
	}

	cmd := BlockEntryCommand{
		EntryID: entryID,
		Reason:  "Test",
	}

	// Act
	err := handler.Handle(context.Background(), cmd)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// Test 3: TestBlockEntryHandler_AlreadyBlocked
func TestBlockEntryHandler_AlreadyBlocked(t *testing.T) {
	// Arrange
	mockRepo := new(MockEntryRepository)
	mockEventPublisher := new(MockEventPublisher)
	mockCacheService := new(MockCacheService)

	entryID := uuid.New()
	entry := &commands.Entry{
		ID:        entryID,
		KeyValue:  "12345678901",
		Status:    "BLOCKED",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("FindByID", mock.Anything, entryID).Return(entry, nil)

	handler := &BlockEntryCommandHandler{
		EntryRepo:      mockRepo,
		EventPublisher: mockEventPublisher,
		CacheService:   mockCacheService,
	}

	cmd := BlockEntryCommand{
		EntryID: entryID,
		Reason:  "Test",
	}

	// Act
	err := handler.Handle(context.Background(), cmd)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already blocked")
}

// ===== UNBLOCK ENTRY TESTS =====

// Test 4: TestUnblockEntryHandler_Success
func TestUnblockEntryHandler_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockEntryRepository)
	mockEventPublisher := new(MockEventPublisher)
	mockCacheService := new(MockCacheService)

	entryID := uuid.New()
	entry := &commands.Entry{
		ID:        entryID,
		KeyValue:  "12345678901",
		Status:    "BLOCKED",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("FindByID", mock.Anything, entryID).Return(entry, nil)
	mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
	mockEventPublisher.On("Publish", mock.Anything, mock.Anything).Return(nil)
	mockCacheService.On("InvalidateKey", mock.Anything, "entry:12345678901").Return(nil)

	handler := &UnblockEntryCommandHandler{
		EntryRepo:      mockRepo,
		EventPublisher: mockEventPublisher,
		CacheService:   mockCacheService,
	}

	cmd := UnblockEntryCommand{
		EntryID:     entryID,
		Reason:      "Issue resolved",
		UnblockedBy: uuid.New(),
	}

	// Act
	err := handler.Handle(context.Background(), cmd)

	// Assert
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// Test 5: TestUnblockEntryHandler_NotBlocked
func TestUnblockEntryHandler_NotBlocked(t *testing.T) {
	// Arrange
	mockRepo := new(MockEntryRepository)
	mockEventPublisher := new(MockEventPublisher)
	mockCacheService := new(MockCacheService)

	entryID := uuid.New()
	entry := &commands.Entry{
		ID:        entryID,
		KeyValue:  "12345678901",
		Status:    "ACTIVE",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("FindByID", mock.Anything, entryID).Return(entry, nil)

	handler := &UnblockEntryCommandHandler{
		EntryRepo:      mockRepo,
		EventPublisher: mockEventPublisher,
		CacheService:   mockCacheService,
	}

	cmd := UnblockEntryCommand{
		EntryID: entryID,
	}

	// Act
	err := handler.Handle(context.Background(), cmd)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not blocked")
}

// Test 6: TestUnblockEntryHandler_NotFound
func TestUnblockEntryHandler_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockEntryRepository)
	mockEventPublisher := new(MockEventPublisher)
	mockCacheService := new(MockCacheService)

	entryID := uuid.New()
	mockRepo.On("FindByID", mock.Anything, entryID).Return(nil, errors.New("not found"))

	handler := &UnblockEntryCommandHandler{
		EntryRepo:      mockRepo,
		EventPublisher: mockEventPublisher,
		CacheService:   mockCacheService,
	}

	cmd := UnblockEntryCommand{
		EntryID: entryID,
	}

	// Act
	err := handler.Handle(context.Background(), cmd)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ===== CREATE INFRACTION TESTS =====

// Test 7: TestCreateInfractionHandler_Success
func TestCreateInfractionHandler_Success(t *testing.T) {
	// Arrange
	mockInfractionRepo := new(MockInfractionRepository)
	mockEventPublisher := new(MockEventPublisher)

	mockInfractionRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockEventPublisher.On("Publish", mock.Anything, mock.Anything).Return(nil)

	handler := &CreateInfractionCommandHandler{
		InfractionRepo: mockInfractionRepo,
		EventPublisher: mockEventPublisher,
	}

	cmd := CreateInfractionCommand{
		EntryID:     uuid.New(),
		InfractionType: "FRAUD",
		Description: "Fraudulent activity detected",
		Severity:    "HIGH",
		ReportedBy:  uuid.New(),
	}

	// Act
	result, err := handler.Handle(context.Background(), cmd)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, result.InfractionID)
	mockInfractionRepo.AssertExpectations(t)
}

// Test 8: TestCreateInfractionHandler_InvalidReason
func TestCreateInfractionHandler_InvalidReason(t *testing.T) {
	// Arrange
	mockInfractionRepo := new(MockInfractionRepository)
	mockEventPublisher := new(MockEventPublisher)

	handler := &CreateInfractionCommandHandler{
		InfractionRepo: mockInfractionRepo,
		EventPublisher: mockEventPublisher,
	}

	cmd := CreateInfractionCommand{
		EntryID:        uuid.New(),
		InfractionType: "INVALID_TYPE",
		Severity:       "HIGH",
	}

	// Act
	result, err := handler.Handle(context.Background(), cmd)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid infraction type")
}

// ===== Mock InfractionRepository =====

type MockInfractionRepository struct {
	mock.Mock
}

func (m *MockInfractionRepository) Create(ctx context.Context, infraction interface{}) error {
	args := m.Called(ctx, infraction)
	return args.Error(0)
}

// ===== Command Handler Stubs =====

type BlockEntryCommand struct {
	EntryID   uuid.UUID
	Reason    string
	BlockedBy uuid.UUID
}

type BlockEntryCommandHandler struct {
	EntryRepo      *MockEntryRepository
	EventPublisher *MockEventPublisher
	CacheService   *MockCacheService
}

func (h *BlockEntryCommandHandler) Handle(ctx context.Context, cmd BlockEntryCommand) error {
	entry, err := h.EntryRepo.FindByID(ctx, cmd.EntryID)
	if err != nil {
		return errors.New("entry not found: " + err.Error())
	}

	if entry.Status == "BLOCKED" {
		return errors.New("entry already blocked")
	}

	entry.Status = "BLOCKED"
	entry.UpdatedAt = time.Now()

	if err := h.EntryRepo.Update(ctx, entry); err != nil {
		return err
	}

	h.EventPublisher.Publish(ctx, commands.DomainEvent{
		EventType:     "EntryBlocked",
		AggregateID:   entry.ID.String(),
		AggregateType: "Entry",
		OccurredAt:    time.Now(),
	})

	h.CacheService.InvalidateKey(ctx, "entry:"+entry.KeyValue)

	return nil
}

type UnblockEntryCommand struct {
	EntryID     uuid.UUID
	Reason      string
	UnblockedBy uuid.UUID
}

type UnblockEntryCommandHandler struct {
	EntryRepo      *MockEntryRepository
	EventPublisher *MockEventPublisher
	CacheService   *MockCacheService
}

func (h *UnblockEntryCommandHandler) Handle(ctx context.Context, cmd UnblockEntryCommand) error {
	entry, err := h.EntryRepo.FindByID(ctx, cmd.EntryID)
	if err != nil {
		return errors.New("entry not found: " + err.Error())
	}

	if entry.Status != "BLOCKED" {
		return errors.New("entry is not blocked")
	}

	entry.Status = "ACTIVE"
	entry.UpdatedAt = time.Now()

	if err := h.EntryRepo.Update(ctx, entry); err != nil {
		return err
	}

	h.EventPublisher.Publish(ctx, commands.DomainEvent{
		EventType:     "EntryUnblocked",
		AggregateID:   entry.ID.String(),
		AggregateType: "Entry",
		OccurredAt:    time.Now(),
	})

	h.CacheService.InvalidateKey(ctx, "entry:"+entry.KeyValue)

	return nil
}

type CreateInfractionCommand struct {
	EntryID        uuid.UUID
	InfractionType string
	Description    string
	Severity       string
	ReportedBy     uuid.UUID
}

type CreateInfractionResult struct {
	InfractionID uuid.UUID
	Status       string
	CreatedAt    time.Time
}

type CreateInfractionCommandHandler struct {
	InfractionRepo *MockInfractionRepository
	EventPublisher *MockEventPublisher
}

func (h *CreateInfractionCommandHandler) Handle(ctx context.Context, cmd CreateInfractionCommand) (*CreateInfractionResult, error) {
	validTypes := map[string]bool{
		"FRAUD":          true,
		"MONEY_LAUNDER":  true,
		"DUPLICATE":      true,
		"INVALID_DATA":   true,
		"UNAUTHORIZED":   true,
	}

	if !validTypes[cmd.InfractionType] {
		return nil, errors.New("invalid infraction type")
	}

	infraction := struct {
		ID             uuid.UUID
		EntryID        uuid.UUID
		InfractionType string
		Description    string
		Severity       string
		Status         string
		CreatedAt      time.Time
	}{
		ID:             uuid.New(),
		EntryID:        cmd.EntryID,
		InfractionType: cmd.InfractionType,
		Description:    cmd.Description,
		Severity:       cmd.Severity,
		Status:         "OPEN",
		CreatedAt:      time.Now(),
	}

	if err := h.InfractionRepo.Create(ctx, infraction); err != nil {
		return nil, err
	}

	h.EventPublisher.Publish(ctx, commands.DomainEvent{
		EventType:     "InfractionCreated",
		AggregateID:   infraction.ID.String(),
		AggregateType: "Infraction",
		OccurredAt:    time.Now(),
	})

	return &CreateInfractionResult{
		InfractionID: infraction.ID,
		Status:       "OPEN",
		CreatedAt:    infraction.CreatedAt,
	}, nil
}
