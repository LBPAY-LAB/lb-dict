package usecases

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-bridge/internal/domain/entities"
	"github.com/lbpay-lab/conn-bridge/internal/domain/interfaces"
	"github.com/lbpay-lab/conn-bridge/internal/domain/valueobjects"
)

// CreateClaimUseCase handles the creation of DICT key claims
type CreateClaimUseCase struct {
	bacenClient      interfaces.BacenClient
	messagePublisher interfaces.MessagePublisher
	circuitBreaker   interfaces.CircuitBreaker
}

// NewCreateClaimUseCase creates a new CreateClaimUseCase
func NewCreateClaimUseCase(
	bacenClient interfaces.BacenClient,
	messagePublisher interfaces.MessagePublisher,
	circuitBreaker interfaces.CircuitBreaker,
) *CreateClaimUseCase {
	return &CreateClaimUseCase{
		bacenClient:      bacenClient,
		messagePublisher: messagePublisher,
		circuitBreaker:   circuitBreaker,
	}
}

// CreateClaimRequest represents the request to create a claim
type CreateClaimRequest struct {
	Claim         *entities.Claim
	CorrelationID string
}

// CreateClaimResponse represents the response of creating a claim
type CreateClaimResponse struct {
	Success       bool
	ClaimID       string
	ErrorCode     string
	ErrorMessage  string
	CorrelationID string
}

// Execute executes the create claim use case
func (uc *CreateClaimUseCase) Execute(ctx context.Context, req *CreateClaimRequest) (*CreateClaimResponse, error) {
	// Validate claim
	if err := req.Claim.Validate(); err != nil {
		return &CreateClaimResponse{
			Success:       false,
			ErrorCode:     "VALIDATION_ERROR",
			ErrorMessage:  err.Error(),
			CorrelationID: req.CorrelationID,
		}, nil
	}

	// TODO: Convert claim to Bacen request format
	payload := []byte{} // Placeholder

	// Create Bacen request
	bacenReq := valueobjects.NewBacenRequest(
		valueobjects.OperationCreateClaim,
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
		Topic:         "dict-claim-created",
		Key:           req.Claim.Key,
		Payload:       bacenResp.Payload,
		CorrelationID: req.CorrelationID,
	}

	if err := uc.messagePublisher.Publish(ctx, message); err != nil {
		// Log error but don't fail the operation
		// TODO: Implement proper error handling and retry logic
	}

	return &CreateClaimResponse{
		Success:       bacenResp.IsSuccess(),
		ClaimID:       req.Claim.ID,
		ErrorCode:     bacenResp.ErrorCode,
		ErrorMessage:  bacenResp.ErrorMessage,
		CorrelationID: req.CorrelationID,
	}, nil
}
