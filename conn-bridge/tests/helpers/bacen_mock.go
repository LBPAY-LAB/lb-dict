package helpers

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/lbpay-lab/conn-bridge/internal/domain/entities"
	xmlstructs "github.com/lbpay-lab/conn-bridge/internal/xml"
)

// MockBacenServer represents a mock Bacen API server for testing
type MockBacenServer struct {
	Server        *httptest.Server
	RequestCount  int
	LastRequest   interface{}
	NextResponse  interface{}
	NextError     error
	ResponseDelay time.Duration
}

// NewMockBacenServer creates a new mock Bacen server
func NewMockBacenServer() *MockBacenServer {
	mock := &MockBacenServer{
		RequestCount: 0,
	}

	server := httptest.NewTLSServer(http.HandlerFunc(mock.handler))
	mock.Server = server

	return mock
}

// handler handles all requests to the mock server
func (m *MockBacenServer) handler(w http.ResponseWriter, r *http.Request) {
	m.RequestCount++

	// Add delay if configured
	if m.ResponseDelay > 0 {
		time.Sleep(m.ResponseDelay)
	}

	// Return error if configured
	if m.NextError != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set headers
	w.Header().Set("Content-Type", "application/xml")

	// Route based on method and path
	switch r.Method {
	case http.MethodPost:
		m.handleCreateEntry(w, r)
	case http.MethodPut:
		m.handleUpdateEntry(w, r)
	case http.MethodDelete:
		m.handleDeleteEntry(w, r)
	case http.MethodGet:
		m.handleGetEntry(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// handleCreateEntry handles create entry requests
func (m *MockBacenServer) handleCreateEntry(w http.ResponseWriter, r *http.Request) {
	var request xmlstructs.XMLCreateEntryRequest
	if err := xml.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	m.LastRequest = request

	// Use custom response if provided
	if m.NextResponse != nil {
		if resp, ok := m.NextResponse.(*xmlstructs.XMLCreateEntryResponse); ok {
			w.WriteHeader(http.StatusCreated)
			xml.NewEncoder(w).Encode(resp)
			return
		}
	}

	// Default response
	response := &xmlstructs.XMLCreateEntryResponse{
		ResponseTime:  time.Now().Format(time.RFC3339),
		CorrelationId: "mock-correlation-id",
		Entry: xmlstructs.XMLExtendedEntry{
			Key:              request.Entry.Key,
			KeyType:          request.Entry.KeyType,
			Account:          request.Entry.Account,
			Owner:            request.Entry.Owner,
			CreationTime:     time.Now().Format(time.RFC3339),
			KeyOwnershipDate: time.Now().Format(time.RFC3339),
		},
	}

	w.WriteHeader(http.StatusCreated)
	xml.NewEncoder(w).Encode(response)
}

// handleUpdateEntry handles update entry requests
func (m *MockBacenServer) handleUpdateEntry(w http.ResponseWriter, r *http.Request) {
	var request xmlstructs.XMLUpdateEntryRequest
	if err := xml.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	m.LastRequest = request

	response := &xmlstructs.XMLUpdateEntryResponse{
		ResponseTime:  time.Now().Format(time.RFC3339),
		CorrelationId: "mock-correlation-id",
		Entry: xmlstructs.XMLExtendedEntry{
			Key:     request.Key,
			KeyType: request.KeyType,
			Account: request.NewAccount,
		},
	}

	w.WriteHeader(http.StatusOK)
	xml.NewEncoder(w).Encode(response)
}

// handleDeleteEntry handles delete entry requests
func (m *MockBacenServer) handleDeleteEntry(w http.ResponseWriter, r *http.Request) {
	response := &xmlstructs.XMLDeleteEntryResponse{
		ResponseTime:  time.Now().Format(time.RFC3339),
		CorrelationId: "mock-correlation-id",
		Deleted:       true,
		Key:           "deleted-key",
		KeyType:       "CPF",
	}

	w.WriteHeader(http.StatusOK)
	xml.NewEncoder(w).Encode(response)
}

// handleGetEntry handles get entry requests
func (m *MockBacenServer) handleGetEntry(w http.ResponseWriter, r *http.Request) {
	response := &xmlstructs.XMLGetEntryResponse{
		ResponseTime:  time.Now().Format(time.RFC3339),
		CorrelationId: "mock-correlation-id",
		Entry: xmlstructs.XMLExtendedEntry{
			Key:     "12345678901",
			KeyType: "CPF",
			Account: xmlstructs.XMLAccount{
				Participant:   "60701190",
				Branch:        "0001",
				AccountNumber: "123456",
				AccountType:   "CHECKING",
			},
			Owner: xmlstructs.XMLOwner{
				Type:        "PERSON",
				TaxIdNumber: "12345678901",
				Name:        "Mock User",
			},
			CreationTime:     time.Now().Format(time.RFC3339),
			KeyOwnershipDate: time.Now().Format(time.RFC3339),
		},
	}

	w.WriteHeader(http.StatusOK)
	xml.NewEncoder(w).Encode(response)
}

// SetNextResponse sets the response for the next request
func (m *MockBacenServer) SetNextResponse(response interface{}) {
	m.NextResponse = response
}

// SetNextError configures the server to return an error
func (m *MockBacenServer) SetNextError(err error) {
	m.NextError = err
}

// SetResponseDelay sets a delay before responding
func (m *MockBacenServer) SetResponseDelay(delay time.Duration) {
	m.ResponseDelay = delay
}

// Reset resets the mock server state
func (m *MockBacenServer) Reset() {
	m.RequestCount = 0
	m.LastRequest = nil
	m.NextResponse = nil
	m.NextError = nil
	m.ResponseDelay = 0
}

// Close closes the mock server
func (m *MockBacenServer) Close() {
	m.Server.Close()
}

// URL returns the mock server URL
func (m *MockBacenServer) URL() string {
	return m.Server.URL
}

// CreateSuccessResponse creates a successful create entry response
func CreateSuccessResponse(entry *entities.DictEntry) *xmlstructs.XMLCreateEntryResponse {
	return &xmlstructs.XMLCreateEntryResponse{
		ResponseTime:  time.Now().Format(time.RFC3339),
		CorrelationId: "test-correlation-id",
		Entry: xmlstructs.XMLExtendedEntry{
			Key:     entry.Key,
			KeyType: string(entry.Type),
			Account: xmlstructs.XMLAccount{
				Participant:   entry.Account.ISPB,
				Branch:        entry.Account.Branch,
				AccountNumber: entry.Account.Number,
				AccountType:   string(entry.Account.Type),
			},
			Owner: xmlstructs.XMLOwner{
				Type:        string(entry.Owner.Type),
				TaxIdNumber: entry.Owner.Document,
				Name:        entry.Owner.Name,
			},
			CreationTime:     time.Now().Format(time.RFC3339),
			KeyOwnershipDate: time.Now().Format(time.RFC3339),
		},
	}
}

// CreateErrorResponse creates an error response
func CreateErrorResponse(errorCode, errorMessage string) *xmlstructs.XMLCreateEntryResponse {
	return &xmlstructs.XMLCreateEntryResponse{
		ResponseTime:  time.Now().Format(time.RFC3339),
		CorrelationId: "error-correlation-id",
		Entry: xmlstructs.XMLExtendedEntry{
			// Empty entry for error case
		},
	}
}

// MockBacenResponses provides common mock responses
type MockBacenResponses struct{}

// GetStandardCreateResponse returns a standard create entry response
func (MockBacenResponses) GetStandardCreateResponse() *xmlstructs.XMLCreateEntryResponse {
	return &xmlstructs.XMLCreateEntryResponse{
		ResponseTime:  time.Now().Format(time.RFC3339),
		CorrelationId: "standard-correlation-id",
		Entry: xmlstructs.XMLExtendedEntry{
			Key:     "test@example.com",
			KeyType: "EMAIL",
			Account: xmlstructs.XMLAccount{
				Participant:   "60701190",
				Branch:        "0001",
				AccountNumber: "123456",
				AccountType:   "CHECKING",
			},
			Owner: xmlstructs.XMLOwner{
				Type:        "PERSON",
				TaxIdNumber: "12345678901",
				Name:        "Standard User",
			},
			CreationTime:     time.Now().Format(time.RFC3339),
			KeyOwnershipDate: time.Now().Format(time.RFC3339),
		},
	}
}

// GetStandardGetResponse returns a standard get entry response
func (MockBacenResponses) GetStandardGetResponse() *xmlstructs.XMLGetEntryResponse {
	return &xmlstructs.XMLGetEntryResponse{
		ResponseTime:  time.Now().Format(time.RFC3339),
		CorrelationId: "standard-correlation-id",
		Entry: xmlstructs.XMLExtendedEntry{
			Key:     "12345678901",
			KeyType: "CPF",
			Account: xmlstructs.XMLAccount{
				Participant:   "60701190",
				Branch:        "0001",
				AccountNumber: "123456",
				AccountType:   "CHECKING",
			},
			Owner: xmlstructs.XMLOwner{
				Type:        "PERSON",
				TaxIdNumber: "12345678901",
				Name:        "Standard User",
			},
			CreationTime:     time.Now().Format(time.RFC3339),
			KeyOwnershipDate: time.Now().Format(time.RFC3339),
		},
	}
}
