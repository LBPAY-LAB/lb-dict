package bacen

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-bridge/internal/domain/interfaces"
	"github.com/lbpay-lab/conn-bridge/internal/domain/valueobjects"
	"github.com/sirupsen/logrus"
)

// CircuitBreakerClient wraps a BacenClient with circuit breaker protection
type CircuitBreakerClient struct {
	client         interfaces.BacenClient
	circuitBreaker interfaces.CircuitBreaker
	logger         *logrus.Logger
}

// NewCircuitBreakerClient creates a new BacenClient wrapper with circuit breaker protection
func NewCircuitBreakerClient(client interfaces.BacenClient, circuitBreaker interfaces.CircuitBreaker, logger *logrus.Logger) interfaces.BacenClient {
	if logger == nil {
		logger = logrus.New()
	}

	return &CircuitBreakerClient{
		client:         client,
		circuitBreaker: circuitBreaker,
		logger:         logger,
	}
}

// SendRequest sends a request to Bacen through the circuit breaker
func (c *CircuitBreakerClient) SendRequest(ctx context.Context, request *valueobjects.BacenRequest) (*valueobjects.BacenResponse, error) {
	// Execute request through circuit breaker
	result, err := c.circuitBreaker.Execute(ctx, func() (interface{}, error) {
		c.logger.WithFields(logrus.Fields{
			"request_id":     request.ID,
			"correlation_id": request.CorrelationID,
		}).Debug("Executing Bacen request through circuit breaker")

		return c.client.SendRequest(ctx, request)
	})

	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"request_id":     request.ID,
			"correlation_id": request.CorrelationID,
			"error":          err.Error(),
		}).Error("Bacen request failed through circuit breaker")
		return nil, fmt.Errorf("circuit breaker execution failed: %w", err)
	}

	// Type assertion to convert result back to BacenResponse
	response, ok := result.(*valueobjects.BacenResponse)
	if !ok {
		c.logger.WithFields(logrus.Fields{
			"request_id":     request.ID,
			"correlation_id": request.CorrelationID,
		}).Error("Failed to cast circuit breaker result to BacenResponse")
		return nil, fmt.Errorf("invalid response type from circuit breaker")
	}

	c.logger.WithFields(logrus.Fields{
		"request_id":     request.ID,
		"correlation_id": request.CorrelationID,
		"status_code":    response.StatusCode,
	}).Debug("Bacen request succeeded through circuit breaker")

	return response, nil
}

// HealthCheck checks if the Bacen API is healthy through the circuit breaker
func (c *CircuitBreakerClient) HealthCheck(ctx context.Context) error {
	_, err := c.circuitBreaker.Execute(ctx, func() (interface{}, error) {
		return nil, c.client.HealthCheck(ctx)
	})

	return err
}

// GetEndpoint returns the current endpoint being used
func (c *CircuitBreakerClient) GetEndpoint() string {
	return c.client.GetEndpoint()
}

// SetTimeout sets the timeout for requests
func (c *CircuitBreakerClient) SetTimeout(timeout int) error {
	return c.client.SetTimeout(timeout)
}
