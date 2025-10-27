package valueobjects

import (
	"time"
)

// BacenRequest represents a request to Bacen
type BacenRequest struct {
	ID            string
	Operation     OperationType
	Payload       []byte
	Headers       map[string]string
	Timestamp     time.Time
	CorrelationID string
}

// OperationType represents the type of Bacen operation
type OperationType string

const (
	OperationCreateEntry  OperationType = "CREATE_ENTRY"
	OperationUpdateEntry  OperationType = "UPDATE_ENTRY"
	OperationDeleteEntry  OperationType = "DELETE_ENTRY"
	OperationQueryEntry   OperationType = "QUERY_ENTRY"
	OperationCreateClaim  OperationType = "CREATE_CLAIM"
	OperationConfirmClaim OperationType = "CONFIRM_CLAIM"
	OperationCancelClaim  OperationType = "CANCEL_CLAIM"
)

// NewBacenRequest creates a new Bacen request
func NewBacenRequest(operation OperationType, payload []byte, correlationID string) *BacenRequest {
	return &BacenRequest{
		Operation:     operation,
		Payload:       payload,
		Headers:       make(map[string]string),
		Timestamp:     time.Now(),
		CorrelationID: correlationID,
	}
}

// AddHeader adds a header to the request
func (r *BacenRequest) AddHeader(key, value string) {
	r.Headers[key] = value
}

// GetHeader gets a header from the request
func (r *BacenRequest) GetHeader(key string) string {
	return r.Headers[key]
}
