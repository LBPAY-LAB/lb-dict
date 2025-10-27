package integration

import (
	"context"
	"testing"
	"time"

	pb "github.com/lbpay-lab/dict-contracts/gen/proto/bridge/v1"
	commonv1 "github.com/lbpay-lab/dict-contracts/gen/proto/common/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/lbpay-lab/conn-bridge/tests/helpers"
)

// TestCreateEntry_E2E tests the complete CreateEntry flow with mock Bacen
func TestCreateEntry_E2E(t *testing.T) {
	// Setup
	client, cleanup := helpers.SetupTestClient(t)
	defer cleanup()

	// Create request
	req := &pb.CreateEntryRequest{
		Key: &commonv1.DictKey{
			KeyType:  commonv1.KeyType_KEY_TYPE_CPF,
			KeyValue: "12345678900",
		},
		Account: &commonv1.Account{
			Ispb:              "12345678",
			AccountType:       commonv1.AccountType_ACCOUNT_TYPE_CHECKING,
			AccountNumber:     "123456",
			AccountCheckDigit: "7",
			BranchCode:        "0001",
		},
		IdempotencyKey: "test-idempotency-key-001",
		RequestId:      "test-request-001",
	}

	// Execute
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.CreateEntry(ctx, req)

	// Assert
	require.NoError(t, err, "CreateEntry should not return error")
	assert.NotEmpty(t, resp.EntryId, "EntryId should be set")
	assert.NotEmpty(t, resp.ExternalId, "ExternalId should be set")
	assert.Equal(t, commonv1.EntryStatus_ENTRY_STATUS_ACTIVE, resp.Status, "Status should be ACTIVE")
	assert.NotEmpty(t, resp.BacenTransactionId, "BacenTransactionId should be set")
}

// TestGetEntry_E2E tests the complete GetEntry flow with mock Bacen
func TestGetEntry_E2E(t *testing.T) {
	// Setup
	client, cleanup := helpers.SetupTestClient(t)
	defer cleanup()

	// Test cases for different query types
	testCases := []struct {
		name       string
		request    *pb.GetEntryRequest
		wantFound  bool
		wantKeyType commonv1.KeyType
	}{
		{
			name: "GetEntry by entry_id",
			request: &pb.GetEntryRequest{
				Identifier: &pb.GetEntryRequest_EntryId{
					EntryId: "entry-123",
				},
				RequestId: "test-request-002",
			},
			wantFound:  true,
			wantKeyType: commonv1.KeyType_KEY_TYPE_CPF,
		},
		{
			name: "GetEntry by key_query",
			request: &pb.GetEntryRequest{
				Identifier: &pb.GetEntryRequest_KeyQuery{
					KeyQuery: &pb.KeyQuery{
						KeyType:  commonv1.KeyType_KEY_TYPE_CPF,
						KeyValue: "12345678900",
					},
				},
				RequestId: "test-request-003",
			},
			wantFound:  true,
			wantKeyType: commonv1.KeyType_KEY_TYPE_CPF,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Execute
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			resp, err := client.GetEntry(ctx, tc.request)

			// Assert
			require.NoError(t, err, "GetEntry should not return error")
			assert.NotEmpty(t, resp.EntryId, "EntryId should be set")
			assert.NotNil(t, resp.Key, "Key should be set")
			assert.NotNil(t, resp.Account, "Account should be set")
			assert.Equal(t, commonv1.EntryStatus_ENTRY_STATUS_ACTIVE, resp.Status, "Status should be ACTIVE")
		})
	}
}

// TestGetDirectory_E2E tests the GetDirectory RPC
func TestGetDirectory_E2E(t *testing.T) {
	// Setup
	client, cleanup := helpers.SetupTestClient(t)
	defer cleanup()

	// Create request
	req := &pb.GetDirectoryRequest{
		PageSize:  100,
		PageToken: "",
		RequestId: "test-request-004",
	}

	// Execute
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.GetDirectory(ctx, req)

	// Assert
	require.NoError(t, err, "GetDirectory should not return error")
	assert.NotNil(t, resp.Entries, "Entries should not be nil")
	assert.GreaterOrEqual(t, len(resp.Entries), 0, "Should return entries list (may be empty)")
	assert.GreaterOrEqual(t, resp.TotalCount, int32(0), "TotalCount should be non-negative")
}

// TestSearchEntries_E2E tests the SearchEntries RPC
func TestSearchEntries_E2E(t *testing.T) {
	// Setup
	client, cleanup := helpers.SetupTestClient(t)
	defer cleanup()

	// Test cases for different search criteria
	testCases := []struct {
		name    string
		request *pb.SearchEntriesRequest
		wantErr bool
	}{
		{
			name: "Search by ISPB",
			request: &pb.SearchEntriesRequest{
				Ispb:      stringPtr("12345678"),
				PageSize:  50,
				RequestId: "test-request-005",
			},
			wantErr: false,
		},
		{
			name: "Search by account holder document",
			request: &pb.SearchEntriesRequest{
				AccountHolderDocument: stringPtr("12345678900"),
				PageSize:              50,
				RequestId:             "test-request-006",
			},
			wantErr: false,
		},
		{
			name: "Search by account number",
			request: &pb.SearchEntriesRequest{
				AccountNumber: stringPtr("123456"),
				PageSize:      50,
				RequestId:     "test-request-007",
			},
			wantErr: false,
		},
		{
			name: "Search without criteria should fail",
			request: &pb.SearchEntriesRequest{
				PageSize:  50,
				RequestId: "test-request-008",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Execute
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			resp, err := client.SearchEntries(ctx, tc.request)

			// Assert
			if tc.wantErr {
				assert.Error(t, err, "SearchEntries should return error for invalid request")
			} else {
				require.NoError(t, err, "SearchEntries should not return error")
				assert.NotNil(t, resp.Entries, "Entries should not be nil")
				assert.GreaterOrEqual(t, resp.TotalCount, int32(0), "TotalCount should be non-negative")
			}
		})
	}
}

// TestHealthCheck_E2E tests the HealthCheck RPC
func TestHealthCheck_E2E(t *testing.T) {
	// Setup
	client, cleanup := helpers.SetupTestClient(t)
	defer cleanup()

	// Execute
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.HealthCheck(ctx, &emptypb.Empty{})

	// Assert
	require.NoError(t, err, "HealthCheck should not return error")
	assert.NotNil(t, resp, "Response should not be nil")

	// Validate status fields
	assert.NotEqual(t, pb.HealthStatus_HEALTH_STATUS_UNSPECIFIED, resp.Status,
		"Overall status should be set")
	assert.NotEqual(t, pb.BacenConnectionStatus_BACEN_CONNECTION_UNSPECIFIED, resp.BacenStatus,
		"Bacen connection status should be set")
	assert.NotEqual(t, pb.CertificateStatus_CERTIFICATE_STATUS_UNSPECIFIED, resp.CertificateStatus,
		"Certificate status should be set")

	// Validate timestamp
	assert.NotNil(t, resp.LastCheck, "LastCheck timestamp should be set")

	// Log results for debugging
	t.Logf("Health Check Results:")
	t.Logf("  Overall Status: %s", resp.Status)
	t.Logf("  Bacen Status: %s", resp.BacenStatus)
	t.Logf("  Certificate Status: %s", resp.CertificateStatus)
	t.Logf("  Bacen Latency: %dms", resp.BacenLatencyMs)
}

// TestDeleteEntry_E2E tests the complete DeleteEntry flow
func TestDeleteEntry_E2E(t *testing.T) {
	// Setup
	client, cleanup := helpers.SetupTestClient(t)
	defer cleanup()

	// Create request
	req := &pb.DeleteEntryRequest{
		EntryId:        "entry-to-delete-001",
		IdempotencyKey: "test-idempotency-key-delete-001",
		RequestId:      "test-request-009",
	}

	// Execute
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.DeleteEntry(ctx, req)

	// Assert
	require.NoError(t, err, "DeleteEntry should not return error")
	assert.True(t, resp.Deleted, "Deleted should be true")
	assert.NotEmpty(t, resp.BacenTransactionId, "BacenTransactionId should be set")
}

// TestUpdateEntry_E2E tests the complete UpdateEntry flow
func TestUpdateEntry_E2E(t *testing.T) {
	// Setup
	client, cleanup := helpers.SetupTestClient(t)
	defer cleanup()

	// Create request
	req := &pb.UpdateEntryRequest{
		EntryId: "entry-to-update-001",
		NewAccount: &commonv1.Account{
			Ispb:              "87654321",
			AccountType:       commonv1.AccountType_ACCOUNT_TYPE_SAVINGS,
			AccountNumber:     "654321",
			AccountCheckDigit: "0",
			BranchCode:        "0002",
		},
		IdempotencyKey: "test-idempotency-key-update-001",
		RequestId:      "test-request-010",
	}

	// Execute
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.UpdateEntry(ctx, req)

	// Assert
	require.NoError(t, err, "UpdateEntry should not return error")
	assert.Equal(t, req.EntryId, resp.EntryId, "EntryId should match request")
	assert.NotEmpty(t, resp.BacenTransactionId, "BacenTransactionId should be set")
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}
