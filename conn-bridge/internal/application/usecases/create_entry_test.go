package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lbpay-lab/conn-bridge/internal/domain/entities"
	"github.com/lbpay-lab/conn-bridge/internal/domain/interfaces"
	"github.com/lbpay-lab/conn-bridge/internal/domain/valueobjects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockBacenClient is a mock implementation of BacenClient
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

// MockMessagePublisher is a mock implementation of MessagePublisher
type MockMessagePublisher struct {
	mock.Mock
}

func (m *MockMessagePublisher) Publish(ctx context.Context, message *interfaces.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessagePublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockCircuitBreaker is a mock implementation of CircuitBreaker
type MockCircuitBreaker struct {
	mock.Mock
}

func (m *MockCircuitBreaker) Execute(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	args := m.Called(ctx, fn)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	// Actually execute the function for testing
	return fn()
}

func (m *MockCircuitBreaker) State() string {
	args := m.Called()
	return args.String(0)
}

func TestNewCreateEntryUseCase(t *testing.T) {
	bacenClient := new(MockBacenClient)
	publisher := new(MockMessagePublisher)
	breaker := new(MockCircuitBreaker)

	uc := NewCreateEntryUseCase(bacenClient, publisher, breaker)

	require.NotNil(t, uc)
	assert.Equal(t, bacenClient, uc.bacenClient)
	assert.Equal(t, publisher, uc.messagePublisher)
	assert.Equal(t, breaker, uc.circuitBreaker)
}

func TestCreateEntryUseCase_Execute_Success(t *testing.T) {
	// Setup
	bacenClient := new(MockBacenClient)
	publisher := new(MockMessagePublisher)
	breaker := new(MockCircuitBreaker)

	uc := NewCreateEntryUseCase(bacenClient, publisher, breaker)

	entry := &entities.DictEntry{
		Key:         "test@example.com",
		Type:        entities.KeyTypeEmail,
		Participant: "60701190",
		Account: entities.Account{
			ISPB:        "60701190",
			Branch:      "0001",
			Number:      "123456",
			Type:        entities.AccountTypeChecking,
			OpeningDate: time.Now(),
		},
		Owner: entities.Owner{
			Type:     entities.OwnerTypePerson,
			Document: "12345678901",
			Name:     "Test User",
		},
		Status:    entities.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	request := &CreateEntryRequest{
		Entry:         entry,
		CorrelationID: "corr-123",
	}

	bacenResponse := valueobjects.NewBacenResponse(
		"req-123",
		200,
		[]byte("success"),
		"corr-123",
	)

	// Mock expectations
	breaker.On("Execute", mock.Anything, mock.AnythingOfType("func() (interface {}, error)")).Return(bacenResponse, nil)
	publisher.On("Publish", mock.Anything, mock.AnythingOfType("*interfaces.Message")).Return(nil)

	// Execute
	ctx := context.Background()
	response, err := uc.Execute(ctx, request)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.Success)
	assert.Equal(t, "test@example.com", response.EntryID)
	assert.Equal(t, "corr-123", response.CorrelationID)

	publisher.AssertExpectations(t)
	breaker.AssertExpectations(t)
}

func TestCreateEntryUseCase_Execute_BacenError(t *testing.T) {
	// Setup
	bacenClient := new(MockBacenClient)
	publisher := new(MockMessagePublisher)
	breaker := new(MockCircuitBreaker)

	uc := NewCreateEntryUseCase(bacenClient, publisher, breaker)

	entry := &entities.DictEntry{
		Key:         "test@example.com",
		Type:        entities.KeyTypeEmail,
		Participant: "60701190",
		Status:      entities.StatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	request := &CreateEntryRequest{
		Entry:         entry,
		CorrelationID: "corr-123",
	}

	// Mock expectations - circuit breaker returns error
	breaker.On("Execute", mock.Anything, mock.AnythingOfType("func() (interface {}, error)")).
		Return(nil, errors.New("bacen service unavailable"))

	// Execute
	ctx := context.Background()
	response, err := uc.Execute(ctx, request)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to send request to Bacen")

	breaker.AssertExpectations(t)
}

func TestCreateEntryUseCase_Execute_PublisherError(t *testing.T) {
	// Setup - publisher error should not fail the operation
	bacenClient := new(MockBacenClient)
	publisher := new(MockMessagePublisher)
	breaker := new(MockCircuitBreaker)

	uc := NewCreateEntryUseCase(bacenClient, publisher, breaker)

	entry := &entities.DictEntry{
		Key:         "test@example.com",
		Type:        entities.KeyTypeEmail,
		Participant: "60701190",
		Status:      entities.StatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	request := &CreateEntryRequest{
		Entry:         entry,
		CorrelationID: "corr-123",
	}

	bacenResponse := valueobjects.NewBacenResponse(
		"req-123",
		200,
		[]byte("success"),
		"corr-123",
	)

	// Mock expectations
	breaker.On("Execute", mock.Anything, mock.AnythingOfType("func() (interface {}, error)")).Return(bacenResponse, nil)
	publisher.On("Publish", mock.Anything, mock.AnythingOfType("*interfaces.Message")).
		Return(errors.New("pulsar connection failed"))

	// Execute
	ctx := context.Background()
	response, err := uc.Execute(ctx, request)

	// Assert - should still succeed even if publisher fails
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.Success)

	publisher.AssertExpectations(t)
	breaker.AssertExpectations(t)
}

func TestCreateEntryUseCase_Execute_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		entry          *entities.DictEntry
		correlationID  string
		bacenStatus    int
		bacenErr       error
		publishErr     error
		wantSuccess    bool
		wantErr        bool
		wantErrContain string
	}{
		{
			name: "successful creation",
			entry: &entities.DictEntry{
				Key:         "12345678901",
				Type:        entities.KeyTypeCPF,
				Participant: "60701190",
				Status:      entities.StatusActive,
			},
			correlationID: "corr-001",
			bacenStatus:   200,
			bacenErr:      nil,
			publishErr:    nil,
			wantSuccess:   true,
			wantErr:       false,
		},
		{
			name: "bacen service error",
			entry: &entities.DictEntry{
				Key:         "12345678901",
				Type:        entities.KeyTypeCPF,
				Participant: "60701190",
				Status:      entities.StatusActive,
			},
			correlationID:  "corr-002",
			bacenErr:       errors.New("service timeout"),
			wantSuccess:    false,
			wantErr:        true,
			wantErrContain: "failed to send request",
		},
		{
			name: "publisher error - non-critical",
			entry: &entities.DictEntry{
				Key:         "test@example.com",
				Type:        entities.KeyTypeEmail,
				Participant: "60701190",
				Status:      entities.StatusActive,
			},
			correlationID: "corr-003",
			bacenStatus:   200,
			bacenErr:      nil,
			publishErr:    errors.New("pulsar down"),
			wantSuccess:   true,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bacenClient := new(MockBacenClient)
			publisher := new(MockMessagePublisher)
			breaker := new(MockCircuitBreaker)

			uc := NewCreateEntryUseCase(bacenClient, publisher, breaker)

			request := &CreateEntryRequest{
				Entry:         tt.entry,
				CorrelationID: tt.correlationID,
			}

			if tt.bacenErr != nil {
				breaker.On("Execute", mock.Anything, mock.AnythingOfType("func() (interface {}, error)")).
					Return(nil, tt.bacenErr)
			} else {
				bacenResponse := valueobjects.NewBacenResponse(
					"req-123",
					tt.bacenStatus,
					[]byte("response"),
					tt.correlationID,
				)
				breaker.On("Execute", mock.Anything, mock.AnythingOfType("func() (interface {}, error)")).
					Return(bacenResponse, nil)
				publisher.On("Publish", mock.Anything, mock.AnythingOfType("*interfaces.Message")).
					Return(tt.publishErr)
			}

			ctx := context.Background()
			response, err := uc.Execute(ctx, request)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrContain != "" {
					assert.Contains(t, err.Error(), tt.wantErrContain)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, tt.wantSuccess, response.Success)
				assert.Equal(t, tt.correlationID, response.CorrelationID)
			}
		})
	}
}
