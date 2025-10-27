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

	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/domain/valueobjects"
)

// MockClaimRepository for claim tests
type MockClaimRepository struct {
	mock.Mock
}

func (m *MockClaimRepository) Create(ctx context.Context, claim *entities.Claim) error {
	args := m.Called(ctx, claim)
	return args.Error(0)
}

func (m *MockClaimRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Claim, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Claim), args.Error(1)
}

func (m *MockClaimRepository) Update(ctx context.Context, claim *entities.Claim) error {
	args := m.Called(ctx, claim)
	return args.Error(0)
}

func (m *MockClaimRepository) FindActiveByEntryKey(ctx context.Context, entryKey string) (*entities.Claim, error) {
	args := m.Called(ctx, entryKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Claim), args.Error(1)
}

// ===== CREATE CLAIM TESTS =====

// Test 1: TestCreateClaimHandler_Success_Ownership
func TestCreateClaimHandler_Success_Ownership(t *testing.T) {
	// Arrange
	mockClaimRepo := new(MockClaimRepository)
	mockEntryRepo := new(MockEntryRepository)
	mockEventPublisher := new(MockEventPublisher)

	entryKey := "12345678901"
	existingEntry := &commands.Entry{
		ID:         uuid.New(),
		KeyValue:   entryKey,
		Status:     "ACTIVE",
		AccountID:  uuid.New(),
		Account:    commands.Account{ISPB: "99999999"},
		Owner:      commands.Owner{TaxID: entryKey},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	mockEntryRepo.On("FindByKeyValue", mock.Anything, entryKey).Return(existingEntry, nil)
	mockClaimRepo.On("FindActiveByEntryKey", mock.Anything, entryKey).Return(nil, errors.New("not found"))
	mockClaimRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockEventPublisher.On("Publish", mock.Anything, mock.Anything).Return(nil)

	handler := &CreateClaimCommandHandler{
		ClaimRepo:      mockClaimRepo,
		EntryRepo:      mockEntryRepo,
		EventPublisher: mockEventPublisher,
	}

	cmd := CreateClaimCommand{
		EntryKey:      entryKey,
		ClaimType:     "OWNERSHIP",
		ClaimerISPB:   "12345678",
		DonorISPB:     "99999999",
		ClaimerAccount: uuid.New(),
		DonorAccount:   existingEntry.AccountID,
		RequestedBy:    uuid.New(),
	}

	// Act
	result, err := handler.Handle(context.Background(), cmd)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, result.ClaimID)
	assert.Equal(t, "OPEN", result.Status)
	mockClaimRepo.AssertExpectations(t)
	mockEntryRepo.AssertExpectations(t)
}

// Test 2: TestCreateClaimHandler_Success_Portability
func TestCreateClaimHandler_Success_Portability(t *testing.T) {
	// Arrange
	mockClaimRepo := new(MockClaimRepository)
	mockEntryRepo := new(MockEntryRepository)
	mockEventPublisher := new(MockEventPublisher)

	entryKey := "12345678901"
	existingEntry := &commands.Entry{
		ID:         uuid.New(),
		KeyValue:   entryKey,
		Status:     "ACTIVE",
		AccountID:  uuid.New(),
		Account:    commands.Account{ISPB: "99999999"},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	mockEntryRepo.On("FindByKeyValue", mock.Anything, entryKey).Return(existingEntry, nil)
	mockClaimRepo.On("FindActiveByEntryKey", mock.Anything, entryKey).Return(nil, errors.New("not found"))
	mockClaimRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockEventPublisher.On("Publish", mock.Anything, mock.Anything).Return(nil)

	handler := &CreateClaimCommandHandler{
		ClaimRepo:      mockClaimRepo,
		EntryRepo:      mockEntryRepo,
		EventPublisher: mockEventPublisher,
	}

	cmd := CreateClaimCommand{
		EntryKey:       entryKey,
		ClaimType:      "PORTABILITY",
		ClaimerISPB:    "12345678",
		DonorISPB:      "99999999",
		ClaimerAccount: uuid.New(),
		DonorAccount:   existingEntry.AccountID,
		RequestedBy:    uuid.New(),
	}

	// Act
	result, err := handler.Handle(context.Background(), cmd)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "PORTABILITY", result.ClaimType)
	mockClaimRepo.AssertExpectations(t)
}

// Test 3: TestCreateClaimHandler_EntryNotFound
func TestCreateClaimHandler_EntryNotFound(t *testing.T) {
	// Arrange
	mockClaimRepo := new(MockClaimRepository)
	mockEntryRepo := new(MockEntryRepository)
	mockEventPublisher := new(MockEventPublisher)

	mockEntryRepo.On("FindByKeyValue", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))

	handler := &CreateClaimCommandHandler{
		ClaimRepo:      mockClaimRepo,
		EntryRepo:      mockEntryRepo,
		EventPublisher: mockEventPublisher,
	}

	cmd := CreateClaimCommand{
		EntryKey:  "nonexistent",
		ClaimType: "OWNERSHIP",
	}

	// Act
	result, err := handler.Handle(context.Background(), cmd)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "entry not found")
}

// Test 4: TestCreateClaimHandler_ActiveClaimExists
func TestCreateClaimHandler_ActiveClaimExists(t *testing.T) {
	// Arrange
	mockClaimRepo := new(MockClaimRepository)
	mockEntryRepo := new(MockEntryRepository)
	mockEventPublisher := new(MockEventPublisher)

	entryKey := "12345678901"
	existingEntry := &commands.Entry{
		ID:        uuid.New(),
		KeyValue:  entryKey,
		Status:    "ACTIVE",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	activeClaim := &entities.Claim{
		ID:        uuid.New(),
		EntryKey:  entryKey,
		Status:    valueobjects.ClaimStatusOpen,
		CreatedAt: time.Now(),
	}

	mockEntryRepo.On("FindByKeyValue", mock.Anything, entryKey).Return(existingEntry, nil)
	mockClaimRepo.On("FindActiveByEntryKey", mock.Anything, entryKey).Return(activeClaim, nil)

	handler := &CreateClaimCommandHandler{
		ClaimRepo:      mockClaimRepo,
		EntryRepo:      mockEntryRepo,
		EventPublisher: mockEventPublisher,
	}

	cmd := CreateClaimCommand{
		EntryKey:  entryKey,
		ClaimType: "OWNERSHIP",
	}

	// Act
	result, err := handler.Handle(context.Background(), cmd)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "active claim already exists")
}

// Test 5: TestCreateClaimHandler_InvalidClaimType
func TestCreateClaimHandler_InvalidClaimType(t *testing.T) {
	// Arrange
	mockClaimRepo := new(MockClaimRepository)
	mockEntryRepo := new(MockEntryRepository)
	mockEventPublisher := new(MockEventPublisher)

	handler := &CreateClaimCommandHandler{
		ClaimRepo:      mockClaimRepo,
		EntryRepo:      mockEntryRepo,
		EventPublisher: mockEventPublisher,
	}

	cmd := CreateClaimCommand{
		EntryKey:  "12345678901",
		ClaimType: "INVALID_TYPE",
	}

	// Act
	result, err := handler.Handle(context.Background(), cmd)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid claim type")
}

// ===== CONFIRM CLAIM TESTS =====

// Test 6: TestConfirmClaimHandler_Success
func TestConfirmClaimHandler_Success(t *testing.T) {
	// Arrange
	mockClaimRepo := new(MockClaimRepository)
	mockEventPublisher := new(MockEventPublisher)

	claimID := uuid.New()
	claim := &entities.Claim{
		ID:        claimID,
		EntryKey:  "12345678901",
		Status:    valueobjects.ClaimStatusOpen,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockClaimRepo.On("FindByID", mock.Anything, claimID).Return(claim, nil)
	mockClaimRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
	mockEventPublisher.On("Publish", mock.Anything, mock.Anything).Return(nil)

	handler := &ConfirmClaimCommandHandler{
		ClaimRepo:      mockClaimRepo,
		EventPublisher: mockEventPublisher,
	}

	cmd := ConfirmClaimCommand{
		ClaimID:     claimID,
		Reason:      "Approved by donor",
		ConfirmedBy: uuid.New(),
	}

	// Act
	err := handler.Handle(context.Background(), cmd)

	// Assert
	require.NoError(t, err)
	mockClaimRepo.AssertExpectations(t)
	mockEventPublisher.AssertExpectations(t)
}

// Test 7: TestConfirmClaimHandler_ClaimNotFound
func TestConfirmClaimHandler_ClaimNotFound(t *testing.T) {
	// Arrange
	mockClaimRepo := new(MockClaimRepository)
	mockEventPublisher := new(MockEventPublisher)

	claimID := uuid.New()
	mockClaimRepo.On("FindByID", mock.Anything, claimID).Return(nil, errors.New("not found"))

	handler := &ConfirmClaimCommandHandler{
		ClaimRepo:      mockClaimRepo,
		EventPublisher: mockEventPublisher,
	}

	cmd := ConfirmClaimCommand{
		ClaimID: claimID,
	}

	// Act
	err := handler.Handle(context.Background(), cmd)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "claim not found")
}

// Test 8: TestConfirmClaimHandler_AlreadyConfirmed
func TestConfirmClaimHandler_AlreadyConfirmed(t *testing.T) {
	// Arrange
	mockClaimRepo := new(MockClaimRepository)
	mockEventPublisher := new(MockEventPublisher)

	claimID := uuid.New()
	claim := &entities.Claim{
		ID:       claimID,
		Status:   valueobjects.ClaimStatusConfirmed,
		UpdatedAt: time.Now(),
	}

	mockClaimRepo.On("FindByID", mock.Anything, claimID).Return(claim, nil)

	handler := &ConfirmClaimCommandHandler{
		ClaimRepo:      mockClaimRepo,
		EventPublisher: mockEventPublisher,
	}

	cmd := ConfirmClaimCommand{
		ClaimID: claimID,
	}

	// Act
	err := handler.Handle(context.Background(), cmd)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot confirm")
}

// ===== CANCEL CLAIM TESTS =====

// Test 9: TestCancelClaimHandler_Success
func TestCancelClaimHandler_Success(t *testing.T) {
	// Arrange
	mockClaimRepo := new(MockClaimRepository)
	mockEventPublisher := new(MockEventPublisher)

	claimID := uuid.New()
	claim := &entities.Claim{
		ID:        claimID,
		Status:    valueobjects.ClaimStatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockClaimRepo.On("FindByID", mock.Anything, claimID).Return(claim, nil)
	mockClaimRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
	mockEventPublisher.On("Publish", mock.Anything, mock.Anything).Return(nil)

	handler := &CancelClaimCommandHandler{
		ClaimRepo:      mockClaimRepo,
		EventPublisher: mockEventPublisher,
	}

	cmd := CancelClaimCommand{
		ClaimID:     claimID,
		Reason:      "User request",
		CancelledBy: uuid.New(),
	}

	// Act
	err := handler.Handle(context.Background(), cmd)

	// Assert
	require.NoError(t, err)
	mockClaimRepo.AssertExpectations(t)
}

// Test 10: TestCancelClaimHandler_ClaimNotFound
func TestCancelClaimHandler_ClaimNotFound(t *testing.T) {
	// Arrange
	mockClaimRepo := new(MockClaimRepository)
	mockEventPublisher := new(MockEventPublisher)

	claimID := uuid.New()
	mockClaimRepo.On("FindByID", mock.Anything, claimID).Return(nil, errors.New("not found"))

	handler := &CancelClaimCommandHandler{
		ClaimRepo:      mockClaimRepo,
		EventPublisher: mockEventPublisher,
	}

	cmd := CancelClaimCommand{
		ClaimID: claimID,
	}

	// Act
	err := handler.Handle(context.Background(), cmd)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "claim not found")
}

// Test 11: TestCancelClaimHandler_CannotCancel
func TestCancelClaimHandler_CannotCancel(t *testing.T) {
	// Arrange
	mockClaimRepo := new(MockClaimRepository)
	mockEventPublisher := new(MockEventPublisher)

	claimID := uuid.New()
	claim := &entities.Claim{
		ID:        claimID,
		Status:    valueobjects.ClaimStatusCompleted,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockClaimRepo.On("FindByID", mock.Anything, claimID).Return(claim, nil)

	handler := &CancelClaimCommandHandler{
		ClaimRepo:      mockClaimRepo,
		EventPublisher: mockEventPublisher,
	}

	cmd := CancelClaimCommand{
		ClaimID: claimID,
	}

	// Act
	err := handler.Handle(context.Background(), cmd)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot cancel")
}

// ===== COMPLETE CLAIM TESTS =====

// Test 12: TestCompleteClaimHandler_Success
func TestCompleteClaimHandler_Success(t *testing.T) {
	// Arrange
	mockClaimRepo := new(MockClaimRepository)
	mockEventPublisher := new(MockEventPublisher)

	claimID := uuid.New()
	claim := &entities.Claim{
		ID:        claimID,
		Status:    valueobjects.ClaimStatusConfirmed,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockClaimRepo.On("FindByID", mock.Anything, claimID).Return(claim, nil)
	mockClaimRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
	mockEventPublisher.On("Publish", mock.Anything, mock.Anything).Return(nil)

	handler := &CompleteClaimCommandHandler{
		ClaimRepo:      mockClaimRepo,
		EventPublisher: mockEventPublisher,
	}

	cmd := CompleteClaimCommand{
		ClaimID:     claimID,
		CompletedBy: uuid.New(),
	}

	// Act
	err := handler.Handle(context.Background(), cmd)

	// Assert
	require.NoError(t, err)
	mockClaimRepo.AssertExpectations(t)
}

// Test 13: TestCompleteClaimHandler_NotConfirmed
func TestCompleteClaimHandler_NotConfirmed(t *testing.T) {
	// Arrange
	mockClaimRepo := new(MockClaimRepository)
	mockEventPublisher := new(MockEventPublisher)

	claimID := uuid.New()
	claim := &entities.Claim{
		ID:        claimID,
		Status:    valueobjects.ClaimStatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockClaimRepo.On("FindByID", mock.Anything, claimID).Return(claim, nil)

	handler := &CompleteClaimCommandHandler{
		ClaimRepo:      mockClaimRepo,
		EventPublisher: mockEventPublisher,
	}

	cmd := CompleteClaimCommand{
		ClaimID: claimID,
	}

	// Act
	err := handler.Handle(context.Background(), cmd)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be confirmed")
}

// Test 14: TestCompleteClaimHandler_Expired
func TestCompleteClaimHandler_Expired(t *testing.T) {
	// Arrange
	mockClaimRepo := new(MockClaimRepository)
	mockEventPublisher := new(MockEventPublisher)

	claimID := uuid.New()
	claim := &entities.Claim{
		ID:        claimID,
		Status:    valueobjects.ClaimStatusConfirmed,
		ExpiresAt: time.Now().Add(-24 * time.Hour), // Expired
		CreatedAt: time.Now().Add(-48 * time.Hour),
		UpdatedAt: time.Now(),
	}

	mockClaimRepo.On("FindByID", mock.Anything, claimID).Return(claim, nil)

	handler := &CompleteClaimCommandHandler{
		ClaimRepo:      mockClaimRepo,
		EventPublisher: mockEventPublisher,
	}

	cmd := CompleteClaimCommand{
		ClaimID: claimID,
	}

	// Act
	err := handler.Handle(context.Background(), cmd)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

// ===== Command Handler Stubs =====

type CreateClaimCommand struct {
	EntryKey       string
	ClaimType      string
	ClaimerISPB    string
	DonorISPB      string
	ClaimerAccount uuid.UUID
	DonorAccount   uuid.UUID
	RequestedBy    uuid.UUID
}

type CreateClaimResult struct {
	ClaimID   uuid.UUID
	Status    string
	ClaimType string
	ExpiresAt time.Time
}

type CreateClaimCommandHandler struct {
	ClaimRepo      *MockClaimRepository
	EntryRepo      *MockEntryRepository
	EventPublisher *MockEventPublisher
}

func (h *CreateClaimCommandHandler) Handle(ctx context.Context, cmd CreateClaimCommand) (*CreateClaimResult, error) {
	// Validate claim type
	if cmd.ClaimType != "OWNERSHIP" && cmd.ClaimType != "PORTABILITY" {
		return nil, errors.New("invalid claim type")
	}

	// Find entry
	_, err := h.EntryRepo.FindByKeyValue(ctx, cmd.EntryKey)
	if err != nil {
		return nil, errors.New("entry not found: " + err.Error())
	}

	// Check for active claim
	_, err = h.ClaimRepo.FindActiveByEntryKey(ctx, cmd.EntryKey)
	if err == nil {
		return nil, errors.New("active claim already exists for this entry")
	}

	// Create claim
	claimType := valueobjects.ClaimTypeOwnership
	if cmd.ClaimType == "PORTABILITY" {
		claimType = valueobjects.ClaimTypePortability
	}

	claim, err := entities.NewClaim(
		cmd.EntryKey,
		claimType,
		valueobjects.Participant{ISPB: cmd.ClaimerISPB},
		valueobjects.Participant{ISPB: cmd.DonorISPB},
		cmd.ClaimerAccount,
		cmd.DonorAccount,
	)
	if err != nil {
		return nil, err
	}

	// Persist
	if err := h.ClaimRepo.Create(ctx, claim); err != nil {
		return nil, err
	}

	// Publish event
	h.EventPublisher.Publish(ctx, commands.DomainEvent{
		EventType:     "ClaimCreated",
		AggregateID:   claim.ID.String(),
		AggregateType: "Claim",
		OccurredAt:    time.Now(),
	})

	return &CreateClaimResult{
		ClaimID:   claim.ID,
		Status:    "OPEN",
		ClaimType: cmd.ClaimType,
		ExpiresAt: claim.ExpiresAt,
	}, nil
}

type ConfirmClaimCommand struct {
	ClaimID     uuid.UUID
	Reason      string
	ConfirmedBy uuid.UUID
}

type ConfirmClaimCommandHandler struct {
	ClaimRepo      *MockClaimRepository
	EventPublisher *MockEventPublisher
}

func (h *ConfirmClaimCommandHandler) Handle(ctx context.Context, cmd ConfirmClaimCommand) error {
	claim, err := h.ClaimRepo.FindByID(ctx, cmd.ClaimID)
	if err != nil {
		return errors.New("claim not found: " + err.Error())
	}

	if err := claim.Confirm(cmd.Reason); err != nil {
		return err
	}

	if err := h.ClaimRepo.Update(ctx, claim); err != nil {
		return err
	}

	h.EventPublisher.Publish(ctx, commands.DomainEvent{
		EventType:     "ClaimConfirmed",
		AggregateID:   claim.ID.String(),
		AggregateType: "Claim",
		OccurredAt:    time.Now(),
	})

	return nil
}

type CancelClaimCommand struct {
	ClaimID     uuid.UUID
	Reason      string
	CancelledBy uuid.UUID
}

type CancelClaimCommandHandler struct {
	ClaimRepo      *MockClaimRepository
	EventPublisher *MockEventPublisher
}

func (h *CancelClaimCommandHandler) Handle(ctx context.Context, cmd CancelClaimCommand) error {
	claim, err := h.ClaimRepo.FindByID(ctx, cmd.ClaimID)
	if err != nil {
		return errors.New("claim not found: " + err.Error())
	}

	if err := claim.Cancel(cmd.Reason); err != nil {
		return err
	}

	if err := h.ClaimRepo.Update(ctx, claim); err != nil {
		return err
	}

	h.EventPublisher.Publish(ctx, commands.DomainEvent{
		EventType:     "ClaimCancelled",
		AggregateID:   claim.ID.String(),
		AggregateType: "Claim",
		OccurredAt:    time.Now(),
	})

	return nil
}

type CompleteClaimCommand struct {
	ClaimID     uuid.UUID
	CompletedBy uuid.UUID
}

type CompleteClaimCommandHandler struct {
	ClaimRepo      *MockClaimRepository
	EventPublisher *MockEventPublisher
}

func (h *CompleteClaimCommandHandler) Handle(ctx context.Context, cmd CompleteClaimCommand) error {
	claim, err := h.ClaimRepo.FindByID(ctx, cmd.ClaimID)
	if err != nil {
		return errors.New("claim not found: " + err.Error())
	}

	if claim.IsExpired() {
		return errors.New("claim has expired")
	}

	if err := claim.Complete(); err != nil {
		return err
	}

	if err := h.ClaimRepo.Update(ctx, claim); err != nil {
		return err
	}

	h.EventPublisher.Publish(ctx, commands.DomainEvent{
		EventType:     "ClaimCompleted",
		AggregateID:   claim.ID.String(),
		AggregateType: "Claim",
		OccurredAt:    time.Now(),
	})

	return nil
}
