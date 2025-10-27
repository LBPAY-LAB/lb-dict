package bacen

import (
	"context"
	"errors"
	"testing"

	"github.com/lbpay-lab/conn-bridge/internal/domain/interfaces"
	"github.com/lbpay-lab/conn-bridge/internal/domain/valueobjects"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBacenClient is a mock implementation of BacenClient for testing
type MockBacenClient struct {
	mock.Mock
}

func (m *MockBacenClient) SendRequest(ctx context.Context, request *valueobjects.BacenRequest) (*valueobjects.BacenResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*valueobjects.BacenResponse), args.Error(1)
}

func (m *MockBacenClient) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockBacenClient) GetEndpoint() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockBacenClient) SetTimeout(timeout int) error {
	args := m.Called(timeout)
	return args.Error(0)
}

// MockCircuitBreaker is a mock implementation of CircuitBreaker for testing
type MockCircuitBreaker struct {
	mock.Mock
}

func (m *MockCircuitBreaker) Execute(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	// Actually execute the function for testing
	return fn()
}

func (m *MockCircuitBreaker) GetState() interfaces.CircuitBreakerState {
	args := m.Called()
	return args.Get(0).(interfaces.CircuitBreakerState)
}

func (m *MockCircuitBreaker) GetStats() interfaces.CircuitBreakerStats {
	args := m.Called()
	return args.Get(0).(interfaces.CircuitBreakerStats)
}

func (m *MockCircuitBreaker) Reset() {
	m.Called()
}

func (m *MockCircuitBreaker) IsOpen() bool {
	args := m.Called()
	return args.Bool(0)
}

func TestNewCircuitBreakerClient(t *testing.T) {
	mockClient := new(MockBacenClient)
	mockCB := new(MockCircuitBreaker)
	logger := logrus.New()

	client := NewCircuitBreakerClient(mockClient, mockCB, logger)

	assert.NotNil(t, client)
}

func TestCircuitBreakerClient_SendRequest_Success(t *testing.T) {
	mockClient := new(MockBacenClient)
	mockCB := new(MockCircuitBreaker)
	logger := logrus.New()

	client := NewCircuitBreakerClient(mockClient, mockCB, logger)

	ctx := context.Background()
	request := &valueobjects.BacenRequest{
		ID:            "test-id",
		CorrelationID: "test-correlation-id",
		Payload:       []byte("test payload"),
	}

	expectedResponse := valueobjects.NewBacenResponse(
		request.ID,
		200,
		[]byte("response payload"),
		request.CorrelationID,
	)

	mockClient.On("SendRequest", ctx, request).Return(expectedResponse, nil)

	response, err := client.SendRequest(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, expectedResponse.ID, response.ID)
	assert.Equal(t, expectedResponse.StatusCode, response.StatusCode)
	assert.Equal(t, expectedResponse.CorrelationID, response.CorrelationID)

	mockClient.AssertExpectations(t)
}

func TestCircuitBreakerClient_SendRequest_Failure(t *testing.T) {
	mockClient := new(MockBacenClient)
	mockCB := new(MockCircuitBreaker)
	logger := logrus.New()

	client := NewCircuitBreakerClient(mockClient, mockCB, logger)

	ctx := context.Background()
	request := &valueobjects.BacenRequest{
		ID:            "test-id",
		CorrelationID: "test-correlation-id",
		Payload:       []byte("test payload"),
	}

	expectedErr := errors.New("connection failed")
	mockClient.On("SendRequest", ctx, request).Return(nil, expectedErr)

	response, err := client.SendRequest(ctx, request)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "circuit breaker execution failed")

	mockClient.AssertExpectations(t)
}

func TestCircuitBreakerClient_HealthCheck(t *testing.T) {
	mockClient := new(MockBacenClient)
	mockCB := new(MockCircuitBreaker)
	logger := logrus.New()

	client := NewCircuitBreakerClient(mockClient, mockCB, logger)

	ctx := context.Background()
	mockClient.On("HealthCheck", ctx).Return(nil)

	err := client.HealthCheck(ctx)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestCircuitBreakerClient_GetEndpoint(t *testing.T) {
	mockClient := new(MockBacenClient)
	mockCB := new(MockCircuitBreaker)
	logger := logrus.New()

	client := NewCircuitBreakerClient(mockClient, mockCB, logger)

	expectedEndpoint := "https://api.bacen.gov.br"
	mockClient.On("GetEndpoint").Return(expectedEndpoint)

	endpoint := client.GetEndpoint()

	assert.Equal(t, expectedEndpoint, endpoint)
	mockClient.AssertExpectations(t)
}

func TestCircuitBreakerClient_SetTimeout(t *testing.T) {
	mockClient := new(MockBacenClient)
	mockCB := new(MockCircuitBreaker)
	logger := logrus.New()

	client := NewCircuitBreakerClient(mockClient, mockCB, logger)

	timeout := 30
	mockClient.On("SetTimeout", timeout).Return(nil)

	err := client.SetTimeout(timeout)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}
