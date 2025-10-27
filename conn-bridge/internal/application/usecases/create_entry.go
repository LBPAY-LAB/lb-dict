package usecases

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-bridge/internal/domain/entities"
	"github.com/lbpay-lab/conn-bridge/internal/domain/interfaces"
	"github.com/lbpay-lab/conn-bridge/internal/domain/valueobjects"
)

// CreateEntryUseCase handles the creation of DICT entries
type CreateEntryUseCase struct {
	bacenClient      interfaces.BacenClient
	messagePublisher interfaces.MessagePublisher
	circuitBreaker   interfaces.CircuitBreaker
}

// NewCreateEntryUseCase creates a new CreateEntryUseCase
func NewCreateEntryUseCase(
	bacenClient interfaces.BacenClient,
	messagePublisher interfaces.MessagePublisher,
	circuitBreaker interfaces.CircuitBreaker,
) *CreateEntryUseCase {
	return &CreateEntryUseCase{
		bacenClient:      bacenClient,
		messagePublisher: messagePublisher,
		circuitBreaker:   circuitBreaker,
	}
}

// CreateEntryRequest represents the request to create a DICT entry
type CreateEntryRequest struct {
	Entry         *entities.DictEntry
	CorrelationID string
}

// CreateEntryResponse represents the response of creating a DICT entry
type CreateEntryResponse struct {
	Success       bool
	EntryID       string
	ErrorCode     string
	ErrorMessage  string
	CorrelationID string
}

// Execute executes the create entry use case
func (uc *CreateEntryUseCase) Execute(ctx context.Context, req *CreateEntryRequest) (*CreateEntryResponse, error) {
	// Validate entry
	if err := req.Entry.Validate(); err != nil {
		return &CreateEntryResponse{
			Success:       false,
			ErrorCode:     "VALIDATION_ERROR",
			ErrorMessage:  err.Error(),
			CorrelationID: req.CorrelationID,
		}, nil
	}

	// TODO: Convert entry to Bacen request format
	payload := []byte{} // Placeholder

	// Create Bacen request
	bacenReq := valueobjects.NewBacenRequest(
		valueobjects.OperationCreateEntry,
		payload,
		req.CorrelationID,
	)

	// Send request to Bacen with circuit breaker protection
	result, err := uc.circuitBreaker.Execute(ctx, func() (interface{}, error) {
		return uc.bacenClient.SendRequest(ctx, bacenReq)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to send request to Bacen: %w", err)
	}

	bacenResp := result.(*valueobjects.BacenResponse)

	// Publish event to Pulsar
	message := &interfaces.Message{
		Topic:         "dict-entry-created",
		Key:           req.Entry.Key,
		Payload:       bacenResp.Payload,
		CorrelationID: req.CorrelationID,
	}

	if err := uc.messagePublisher.Publish(ctx, message); err != nil {
		// Log error but don't fail the operation
		// TODO: Implement proper error handling and retry logic
	}

	// Build response
	return &CreateEntryResponse{
		Success:       bacenResp.IsSuccess(),
		EntryID:       req.Entry.Key,
		ErrorCode:     bacenResp.ErrorCode,
		ErrorMessage:  bacenResp.ErrorMessage,
		CorrelationID: req.CorrelationID,
	}, nil
}
