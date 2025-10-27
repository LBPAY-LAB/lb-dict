package interfaces

import (
	"context"

	"github.com/lbpay-lab/conn-bridge/internal/domain/valueobjects"
)

// BacenClient defines the interface for communicating with Bacen DICT API
type BacenClient interface {
	// SendRequest sends a request to Bacen and returns the response
	SendRequest(ctx context.Context, request *valueobjects.BacenRequest) (*valueobjects.BacenResponse, error)

	// HealthCheck checks if the Bacen API is healthy
	HealthCheck(ctx context.Context) error

	// GetEndpoint returns the current endpoint being used
	GetEndpoint() string

	// SetTimeout sets the timeout for requests
	SetTimeout(timeout int) error
}
