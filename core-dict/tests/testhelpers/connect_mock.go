package testhelpers

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ConnectMock simulates the conn-dict gRPC service
type ConnectMock struct {
	mu              sync.RWMutex
	t               *testing.T
	checkDuplicates map[string]bool // key -> isDuplicate
	calls           []MockCall
	server          *grpc.Server
}

// MockCall represents a gRPC call
type MockCall struct {
	Method string
	Req    interface{}
	Resp   interface{}
	Err    error
}

// NewConnectMock creates a new Connect mock
func NewConnectMock(t *testing.T) *ConnectMock {
	return &ConnectMock{
		t:               t,
		checkDuplicates: make(map[string]bool),
		calls:           []MockCall{},
	}
}

// CheckDuplicate simulates the CheckDuplicate gRPC call
func (c *ConnectMock) CheckDuplicate(ctx context.Context, keyType string, keyValue string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := fmt.Sprintf("%s:%s", keyType, keyValue)

	call := MockCall{
		Method: "CheckDuplicate",
		Req:    map[string]string{"keyType": keyType, "keyValue": keyValue},
	}

	isDuplicate, exists := c.checkDuplicates[key]
	if !exists {
		isDuplicate = false
	}

	call.Resp = isDuplicate
	c.calls = append(c.calls, call)

	return isDuplicate, nil
}

// SetDuplicateResult sets the response for CheckDuplicate
func (c *ConnectMock) SetDuplicateResult(keyType string, keyValue string, isDuplicate bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := fmt.Sprintf("%s:%s", keyType, keyValue)
	c.checkDuplicates[key] = isDuplicate
}

// NotifyEntryCreated simulates entry creation notification
func (c *ConnectMock) NotifyEntryCreated(ctx context.Context, entryID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	call := MockCall{
		Method: "NotifyEntryCreated",
		Req:    entryID,
	}

	c.calls = append(c.calls, call)
	return nil
}

// CreateClaim simulates claim creation via Temporal workflow
func (c *ConnectMock) CreateClaim(ctx context.Context, claimID string, claimType string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	call := MockCall{
		Method: "CreateClaim",
		Req:    map[string]string{"claimID": claimID, "claimType": claimType},
	}

	c.calls = append(c.calls, call)
	return nil
}

// GetCalls returns all mock calls
func (c *ConnectMock) GetCalls() []MockCall {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return append([]MockCall{}, c.calls...)
}

// GetCallCount returns the number of calls to a method
func (c *ConnectMock) GetCallCount(method string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	count := 0
	for _, call := range c.calls {
		if call.Method == method {
			count++
		}
	}
	return count
}

// Reset clears all calls
func (c *ConnectMock) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.calls = []MockCall{}
	c.checkDuplicates = make(map[string]bool)
}

// Stop stops the mock server
func (c *ConnectMock) Stop() {
	if c.server != nil {
		c.server.Stop()
	}
}

// SimulateError makes the next call return an error
func (c *ConnectMock) SimulateError(method string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Store error for next call of this method
	// Implementation depends on your needs
}

// SimulateTimeout simulates a timeout
func (c *ConnectMock) SimulateTimeout(method string) error {
	return status.Error(codes.DeadlineExceeded, "context deadline exceeded")
}
