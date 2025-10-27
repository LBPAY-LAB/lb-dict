package usecases

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-bridge/internal/domain/entities"
	"github.com/lbpay-lab/conn-bridge/internal/domain/interfaces"
	"github.com/lbpay-lab/conn-bridge/internal/domain/valueobjects"
)

// QueryEntryUseCase handles querying DICT entries
type QueryEntryUseCase struct {
	bacenClient    interfaces.BacenClient
	circuitBreaker interfaces.CircuitBreaker
}

// NewQueryEntryUseCase creates a new QueryEntryUseCase
func NewQueryEntryUseCase(
	bacenClient interfaces.BacenClient,
	circuitBreaker interfaces.CircuitBreaker,
) *QueryEntryUseCase {
	return &QueryEntryUseCase{
		bacenClient:    bacenClient,
		circuitBreaker: circuitBreaker,
	}
}

// QueryEntryRequest represents the request to query a DICT entry
type QueryEntryRequest struct {
	Key           string
	KeyType       entities.KeyType
	CorrelationID string
}

// QueryEntryResponse represents the response of querying a DICT entry
type QueryEntryResponse struct {
	Success       bool
	Entry         *entities.DictEntry
	ErrorCode     string
	ErrorMessage  string
	CorrelationID string
}

// Execute executes the query entry use case
func (uc *QueryEntryUseCase) Execute(ctx context.Context, req *QueryEntryRequest) (*QueryEntryResponse, error) {
	// TODO: Build query payload
	payload := []byte{} // Placeholder

	// Create Bacen request
	bacenReq := valueobjects.NewBacenRequest(
		valueobjects.OperationQueryEntry,
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

	if !bacenResp.IsSuccess() {
		return &QueryEntryResponse{
			Success:       false,
			ErrorCode:     bacenResp.ErrorCode,
			ErrorMessage:  bacenResp.ErrorMessage,
			CorrelationID: req.CorrelationID,
		}, nil
	}

	// TODO: Parse response and convert to DictEntry
	entry := &entities.DictEntry{
		Key:     req.Key,
		Type:    req.KeyType,
		Status:  entities.StatusActive,
	}

	return &QueryEntryResponse{
		Success:       true,
		Entry:         entry,
		CorrelationID: req.CorrelationID,
	}, nil
}
