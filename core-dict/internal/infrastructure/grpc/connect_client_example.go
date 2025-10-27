package grpc

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	commonv1 "github.com/lbpay-lab/dict-contracts/gen/proto/common/v1"
	connectv1 "github.com/lbpay-lab/dict-contracts/gen/proto/connect/v1"
)

// Example usage of ConnectClient with all features

func ExampleConnectClient_BasicUsage() {
	// Create client with default configuration
	client, err := NewConnectClient("localhost:9092")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	requestID := uuid.New().String()

	// Get entry by ID
	entry, err := client.GetEntry(ctx, "entry-123", requestID)
	if err != nil {
		log.Printf("GetEntry failed: %v", err)
		return
	}
	fmt.Printf("Entry found: %s\n", entry.EntryId)
}

func ExampleConnectClient_CustomConfiguration() {
	// Create client with custom configuration
	client, err := NewConnectClient(
		"localhost:9092",
		WithTimeout(10*time.Second),
		WithMaxMessageSize(20*1024*1024), // 20MB
		WithHealthCheck(true, 1*time.Minute),
		WithCircuitBreaker(CircuitBreakerConfig{
			Threshold:     10, // Open after 10 consecutive failures
			Timeout:       2 * time.Minute,
			HalfOpenTests: 2,
		}),
		WithRetryPolicy(RetryConfig{
			MaxRetries: 5,
			BaseDelay:  200 * time.Millisecond,
			MaxDelay:   5 * time.Second,
			Multiplier: 2.0,
			Jitter:     0.3,
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	requestID := uuid.New().String()

	// List entries with filters
	filters := ListEntriesFilters{
		ParticipantISPB: "12345678",
		KeyType:         ptr(commonv1.KeyType_KEY_TYPE_CPF),
		Status:          ptr(commonv1.EntryStatus_ENTRY_STATUS_ACTIVE),
		Limit:           100,
		Offset:          0,
		RequestID:       requestID,
	}

	entries, totalCount, err := client.ListEntries(ctx, filters)
	if err != nil {
		log.Printf("ListEntries failed: %v", err)
		return
	}

	fmt.Printf("Found %d entries (total: %d)\n", len(entries), totalCount)
}

func ExampleConnectClient_CreateClaim() {
	client, err := NewConnectClient("localhost:9092")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	requestID := uuid.New().String()

	// Create a claim
	claimReq := CreateClaimRequest{
		EntryID:     "entry-123",
		ClaimerISPB: "87654321",
		OwnerISPB:   "12345678",
		ClaimerAccount: &commonv1.Account{
			Ispb:                  "87654321",
			AccountType:           commonv1.AccountType_ACCOUNT_TYPE_CHECKING,
			AccountNumber:         "12345678",
			AccountCheckDigit:     "0",
			BranchCode:            "0001",
			AccountHolderName:     "John Doe",
			AccountHolderDocument: "12345678901",
			DocumentType:          commonv1.DocumentType_DOCUMENT_TYPE_CPF,
		},
		ClaimType: connectv1.CreateClaimRequest_CLAIM_TYPE_OWNERSHIP,
		RequestID: requestID,
	}

	claimResp, err := client.CreateClaim(ctx, claimReq)
	if err != nil {
		log.Printf("CreateClaim failed: %v", err)
		return
	}

	fmt.Printf("Claim created: %s (expires at %s)\n",
		claimResp.ClaimId,
		claimResp.ExpiresAt.AsTime().Format(time.RFC3339))
}

func ExampleConnectClient_CreateInfraction() {
	client, err := NewConnectClient("localhost:9092")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	requestID := uuid.New().String()

	// Create an infraction
	infractionReq := CreateInfractionRequest{
		Key: &commonv1.DictKey{
			KeyType:  commonv1.KeyType_KEY_TYPE_CPF,
			KeyValue: "12345678901",
		},
		ParticipantISPB: "12345678",
		InfractionType:  connectv1.CreateInfractionRequest_INFRACTION_TYPE_FRAUD,
		Description:     "Suspected fraudulent activity detected",
		ReporterISPB:    "87654321",
		RequestID:       requestID,
	}

	infractionResp, err := client.CreateInfraction(ctx, infractionReq)
	if err != nil {
		log.Printf("CreateInfraction failed: %v", err)
		return
	}

	fmt.Printf("Infraction created: %s (status: %s)\n",
		infractionResp.InfractionId,
		infractionResp.Status)
}

func ExampleConnectClient_HealthCheck() {
	client, err := NewConnectClient("localhost:9092")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Perform health check
	health, err := client.HealthCheck(ctx)
	if err != nil {
		log.Printf("HealthCheck failed: %v", err)
		return
	}

	fmt.Printf("Service status: %s\n", health.Status)
	fmt.Printf("PostgreSQL: reachable=%v, latency=%dms\n",
		health.Postgresql.Reachable,
		health.Postgresql.LatencyMs)
	fmt.Printf("Redis: reachable=%v, latency=%dms\n",
		health.Redis.Reachable,
		health.Redis.LatencyMs)
	fmt.Printf("Temporal: reachable=%v, latency=%dms\n",
		health.Temporal.Reachable,
		health.Temporal.LatencyMs)
	fmt.Printf("Pulsar: reachable=%v, latency=%dms\n",
		health.Pulsar.Reachable,
		health.Pulsar.LatencyMs)
	fmt.Printf("Bridge: reachable=%v, latency=%dms\n",
		health.Bridge.Reachable,
		health.Bridge.LatencyMs)
}

func ExampleConnectClient_CircuitBreaker() {
	client, err := NewConnectClient("localhost:9092")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	requestID := uuid.New().String()

	// Check circuit breaker state before making request
	state := client.GetCircuitBreakerState()
	fmt.Printf("Circuit breaker state: %s\n", state)

	// Make request (may fail if circuit is open)
	_, err = client.GetEntry(ctx, "entry-123", requestID)
	if err == ErrCircuitOpen {
		fmt.Println("Circuit breaker is OPEN, service is unavailable")
		return
	}

	// Get circuit breaker metrics
	metrics := client.GetCircuitBreakerMetrics()
	fmt.Printf("Circuit breaker metrics: %+v\n", metrics)

	// Manually reset circuit breaker if needed
	client.ResetCircuitBreaker()
	fmt.Println("Circuit breaker reset to CLOSED state")
}

// Helper function to create pointer
func ptr[T any](v T) *T {
	return &v
}
