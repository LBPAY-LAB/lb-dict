package usecases

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-bridge/internal/domain/interfaces"
	"github.com/lbpay-lab/conn-bridge/internal/domain/valueobjects"
)

// DeleteEntryUseCase handles the deletion of DICT entries
type DeleteEntryUseCase struct {
	bacenClient      interfaces.BacenClient
	messagePublisher interfaces.MessagePublisher
	circuitBreaker   interfaces.CircuitBreaker
}

// NewDeleteEntryUseCase creates a new DeleteEntryUseCase
func NewDeleteEntryUseCase(
	bacenClient interfaces.BacenClient,
	messagePublisher interfaces.MessagePublisher,
	circuitBreaker interfaces.CircuitBreaker,
) *DeleteEntryUseCase {
	return &DeleteEntryUseCase{
		bacenClient:      bacenClient,
		messagePublisher: messagePublisher,
		circuitBreaker:   circuitBreaker,
	}
}

// DeleteEntryRequest represents the request to delete a DICT entry
type DeleteEntryRequest struct {
	Key           string
	Reason        string
	CorrelationID string
}

// DeleteEntryResponse represents the response of deleting a DICT entry
type DeleteEntryResponse struct {
	Success       bool
	ErrorCode     string
	ErrorMessage  string
	CorrelationID string
}

// Execute executes the delete entry use case
func (uc *DeleteEntryUseCase) Execute(ctx context.Context, req *DeleteEntryRequest) (*DeleteEntryResponse, error) {
	// TODO: Build delete payload
	payload := []byte{} // Placeholder

	// Create Bacen request
	bacenReq := valueobjects.NewBacenRequest(
		valueobjects.OperationDeleteEntry,
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
		Topic:         "dict-entry-deleted",
		Key:           req.Key,
		Payload:       bacenResp.Payload,
		CorrelationID: req.CorrelationID,
	}

	if err := uc.messagePublisher.Publish(ctx, message); err != nil {
		// Log error but don't fail the operation
		// TODO: Implement proper error handling and retry logic
	}

	return &DeleteEntryResponse{
		Success:       bacenResp.IsSuccess(),
		ErrorCode:     bacenResp.ErrorCode,
		ErrorMessage:  bacenResp.ErrorMessage,
		CorrelationID: req.CorrelationID,
	}, nil
}
