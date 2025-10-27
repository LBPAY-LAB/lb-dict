package valueobjects

import (
	"time"
)

// BacenResponse represents a response from Bacen
type BacenResponse struct {
	ID            string
	RequestID     string
	StatusCode    int
	Success       bool
	Payload       []byte
	ErrorCode     string
	ErrorMessage  string
	Timestamp     time.Time
	CorrelationID string
}

// NewBacenResponse creates a new Bacen response
func NewBacenResponse(requestID string, statusCode int, payload []byte, correlationID string) *BacenResponse {
	return &BacenResponse{
		RequestID:     requestID,
		StatusCode:    statusCode,
		Success:       statusCode >= 200 && statusCode < 300,
		Payload:       payload,
		Timestamp:     time.Now(),
		CorrelationID: correlationID,
	}
}

// WithError sets error information on the response
func (r *BacenResponse) WithError(code, message string) *BacenResponse {
	r.Success = false
	r.ErrorCode = code
	r.ErrorMessage = message
	return r
}

// IsSuccess returns true if the response is successful
func (r *BacenResponse) IsSuccess() bool {
	return r.Success && r.ErrorCode == ""
}
