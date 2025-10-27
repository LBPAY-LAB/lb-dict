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

// Test 1: TestDeleteEntryHandler_Success
func TestDeleteEntryHandler_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockEntryRepository)
	mockEventPublisher := new(MockEventPublisher)
	mockCacheService := new(MockCacheService)

	entryID := uuid.New()
	existingEntry := &commands.Entry{
		ID:        entryID,
		KeyValue:  "12345678901",
		Status:    "ACTIVE",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("FindByID", mock.Anything, entryID).Return(existingEntry, nil)
	mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
	mockEventPublisher.On("Publish", mock.Anything, mock.Anything).Return(nil)
	mockCacheService.On("InvalidateKey", mock.Anything, "entry:12345678901").Return(nil)

	handler := &commands.DeleteEntryCommandHandler{
		EntryRepo:      mockRepo,
		EventPublisher: mockEventPublisher,
		CacheService:   mockCacheService,
	}

	cmd := commands.DeleteEntryCommand{
		EntryID:     entryID,
		Reason:      "User request",
		RequestedBy: uuid.New(),
	}

	// Act
	err := handler.Handle(context.Background(), cmd)

	// Assert
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockEventPublisher.AssertExpectations(t)
	mockCacheService.AssertExpectations(t)
}

// Test 2: TestDeleteEntryHandler_NotFound
func TestDeleteEntryHandler_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockEntryRepository)
	mockEventPublisher := new(MockEventPublisher)
	mockCacheService := new(MockCacheService)

	entryID := uuid.New()

	mockRepo.On("FindByID", mock.Anything, entryID).Return(nil, errors.New("not found"))

	handler := &commands.DeleteEntryCommandHandler{
		EntryRepo:      mockRepo,
		EventPublisher: mockEventPublisher,
		CacheService:   mockCacheService,
	}

	cmd := commands.DeleteEntryCommand{
		EntryID:     entryID,
		Reason:      "User request",
		RequestedBy: uuid.New(),
	}

	// Act
	err := handler.Handle(context.Background(), cmd)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

// Test 3: TestDeleteEntryHandler_AlreadyDeleted
func TestDeleteEntryHandler_AlreadyDeleted(t *testing.T) {
	// Arrange
	mockRepo := new(MockEntryRepository)
	mockEventPublisher := new(MockEventPublisher)
	mockCacheService := new(MockCacheService)

	entryID := uuid.New()
	deletedTime := time.Now()
	existingEntry := &commands.Entry{
		ID:        entryID,
		KeyValue:  "12345678901",
		Status:    "DELETED",
		DeletedAt: &deletedTime,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("FindByID", mock.Anything, entryID).Return(existingEntry, nil)

	handler := &commands.DeleteEntryCommandHandler{
		EntryRepo:      mockRepo,
		EventPublisher: mockEventPublisher,
		CacheService:   mockCacheService,
	}

	cmd := commands.DeleteEntryCommand{
		EntryID:     entryID,
		Reason:      "User request",
		RequestedBy: uuid.New(),
	}

	// Act
	err := handler.Handle(context.Background(), cmd)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already deleted")
	mockRepo.AssertExpectations(t)
}

// DeleteEntryCommandHandler stub (to be implemented by backend agent)
type DeleteEntryCommand struct {
	EntryID     uuid.UUID
	Reason      string
	RequestedBy uuid.UUID
}

type DeleteEntryCommandHandler struct {
	EntryRepo      commands.EntryRepository
	EventPublisher commands.EventPublisher
	CacheService   commands.CacheService
}

func (h *DeleteEntryCommandHandler) Handle(ctx context.Context, cmd DeleteEntryCommand) error {
	// Find entry
	entry, err := h.EntryRepo.FindByID(ctx, cmd.EntryID)
	if err != nil {
		return errors.New("entry not found: " + err.Error())
	}

	// Check if already deleted
	if entry.DeletedAt != nil || entry.Status == "DELETED" {
		return errors.New("entry already deleted")
	}

	// Mark as deleted (soft delete)
	now := time.Now()
	entry.Status = "DELETED"
	entry.DeletedAt = &now
	entry.UpdatedAt = now

	// Persist
	if err := h.EntryRepo.Update(ctx, entry); err != nil {
		return errors.New("failed to delete entry: " + err.Error())
	}

	// Publish event
	h.EventPublisher.Publish(ctx, commands.DomainEvent{
		EventType:     "EntryDeleted",
		AggregateID:   entry.ID.String(),
		AggregateType: "Entry",
		OccurredAt:    now,
		Payload: map[string]interface{}{
			"entry_id":   entry.ID.String(),
			"key_value":  entry.KeyValue,
			"reason":     cmd.Reason,
			"deleted_by": cmd.RequestedBy.String(),
		},
	})

	// Invalidate cache
	h.CacheService.InvalidateKey(ctx, "entry:"+entry.KeyValue)

	return nil
}
