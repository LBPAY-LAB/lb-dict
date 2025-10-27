package grpc

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/types/known/emptypb"

	commonv1 "github.com/lbpay-lab/dict-contracts/gen/proto/common/v1"
	connectv1 "github.com/lbpay-lab/dict-contracts/gen/proto/connect/v1"
)

// ConnectClient is a gRPC client for communicating with conn-dict service
type ConnectClient struct {
	conn           *grpc.ClientConn
	client         connectv1.ConnectServiceClient
	circuitBreaker *CircuitBreaker
	retryPolicy    *RetryPolicy
	config         ClientConfig
}

// ClientConfig holds configuration for the ConnectClient
type ClientConfig struct {
	Address           string
	Timeout           time.Duration
	MaxMessageSize    int
	PoolSize          int
	EnableHealthCheck bool
	HealthCheckPeriod time.Duration
	CircuitBreaker    CircuitBreakerConfig
	Retry             RetryConfig
}

// ClientOption is a functional option for configuring the client
type ClientOption func(*ClientConfig)

// WithTimeout sets the default timeout for requests
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *ClientConfig) {
		c.Timeout = timeout
	}
}

// WithMaxMessageSize sets the maximum message size
func WithMaxMessageSize(size int) ClientOption {
	return func(c *ClientConfig) {
		c.MaxMessageSize = size
	}
}

// WithHealthCheck enables periodic health checks
func WithHealthCheck(enabled bool, period time.Duration) ClientOption {
	return func(c *ClientConfig) {
		c.EnableHealthCheck = enabled
		c.HealthCheckPeriod = period
	}
}

// WithCircuitBreaker configures the circuit breaker
func WithCircuitBreaker(config CircuitBreakerConfig) ClientOption {
	return func(c *ClientConfig) {
		c.CircuitBreaker = config
	}
}

// WithRetryPolicy configures the retry policy
func WithRetryPolicy(config RetryConfig) ClientOption {
	return func(c *ClientConfig) {
		c.Retry = config
	}
}

// NewConnectClient creates a new ConnectClient with the given address and options
func NewConnectClient(address string, opts ...ClientOption) (*ConnectClient, error) {
	// Default configuration
	config := ClientConfig{
		Address:           address,
		Timeout:           5 * time.Second,
		MaxMessageSize:    10 * 1024 * 1024, // 10MB
		PoolSize:          10,
		EnableHealthCheck: true,
		HealthCheckPeriod: 30 * time.Second,
		CircuitBreaker: CircuitBreakerConfig{
			Threshold:     5,
			Timeout:       60 * time.Second,
			HalfOpenTests: 1,
			OnStateChange: func(from, to State) {
				log.Printf("[ConnectClient] Circuit breaker state changed: %s -> %s", from, to)
			},
		},
		Retry: RetryConfig{
			MaxRetries: 3,
			BaseDelay:  100 * time.Millisecond,
			MaxDelay:   2 * time.Second,
			Multiplier: 2.0,
			Jitter:     0.2,
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(&config)
	}

	// Create gRPC connection with keep-alive
	conn, err := grpc.NewClient(
		config.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(config.MaxMessageSize),
			grpc.MaxCallSendMsgSize(config.MaxMessageSize),
		),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                15 * time.Minute,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	client := &ConnectClient{
		conn:           conn,
		client:         connectv1.NewConnectServiceClient(conn),
		circuitBreaker: NewCircuitBreaker(config.CircuitBreaker),
		retryPolicy:    NewRetryPolicy(config.Retry),
		config:         config,
	}

	// Start health check goroutine if enabled
	if config.EnableHealthCheck {
		go client.healthCheckLoop()
	}

	return client, nil
}

// Close closes the gRPC connection
func (c *ConnectClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// =============================================================================
// ENTRY OPERATIONS (3 methods)
// =============================================================================

// GetEntry retrieves an entry by its ID
func (c *ConnectClient) GetEntry(ctx context.Context, entryID string, requestID string) (*connectv1.Entry, error) {
	var entry *connectv1.Entry
	var found bool

	err := c.executeWithRetry(ctx, func() error {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()

		resp, err := c.client.GetEntry(ctxWithTimeout, &connectv1.GetEntryRequest{
			EntryId:   entryID,
			RequestId: requestID,
		})
		if err != nil {
			return mapGRPCError(err)
		}

		entry = resp.Entry
		found = resp.Found
		return nil
	})

	if err != nil {
		return nil, err
	}

	if !found {
		return nil, ErrEntryNotFound
	}

	return entry, nil
}

// GetEntryByKey retrieves an entry by its DICT key
func (c *ConnectClient) GetEntryByKey(ctx context.Context, key *commonv1.DictKey, requestID string) (*connectv1.Entry, error) {
	var entry *connectv1.Entry
	var found bool

	err := c.executeWithRetry(ctx, func() error {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()

		resp, err := c.client.GetEntryByKey(ctxWithTimeout, &connectv1.GetEntryByKeyRequest{
			Key:       key,
			RequestId: requestID,
		})
		if err != nil {
			return mapGRPCError(err)
		}

		entry = resp.Entry
		found = resp.Found
		return nil
	})

	if err != nil {
		return nil, err
	}

	if !found {
		return nil, ErrEntryNotFound
	}

	return entry, nil
}

// ListEntriesFilters holds filters for listing entries
type ListEntriesFilters struct {
	ParticipantISPB string
	KeyType         *commonv1.KeyType
	Status          *commonv1.EntryStatus
	Limit           int32
	Offset          int32
	RequestID       string
}

// ListEntries retrieves a list of entries with optional filters
func (c *ConnectClient) ListEntries(ctx context.Context, filters ListEntriesFilters) ([]*connectv1.Entry, int32, error) {
	var entries []*connectv1.Entry
	var totalCount int32

	err := c.executeWithRetry(ctx, func() error {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()

		req := &connectv1.ListEntriesRequest{
			ParticipantIspb: filters.ParticipantISPB,
			Limit:           filters.Limit,
			Offset:          filters.Offset,
			RequestId:       filters.RequestID,
		}

		if filters.KeyType != nil {
			req.KeyType = filters.KeyType
		}
		if filters.Status != nil {
			req.Status = filters.Status
		}

		resp, err := c.client.ListEntries(ctxWithTimeout, req)
		if err != nil {
			return mapGRPCError(err)
		}

		entries = resp.Entries
		totalCount = resp.TotalCount
		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return entries, totalCount, nil
}

// =============================================================================
// CLAIM OPERATIONS (5 methods)
// =============================================================================

// CreateClaimRequest holds parameters for creating a claim
type CreateClaimRequest struct {
	EntryID        string
	ClaimerISPB    string
	OwnerISPB      string
	ClaimerAccount *commonv1.Account
	ClaimType      connectv1.CreateClaimRequest_ClaimType
	RequestID      string
}

// CreateClaim initiates a new claim (starts 30-day ClaimWorkflow)
func (c *ConnectClient) CreateClaim(ctx context.Context, req CreateClaimRequest) (*connectv1.CreateClaimResponse, error) {
	var resp *connectv1.CreateClaimResponse

	err := c.executeWithRetry(ctx, func() error {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()

		result, err := c.client.CreateClaim(ctxWithTimeout, &connectv1.CreateClaimRequest{
			EntryId:        req.EntryID,
			ClaimerIspb:    req.ClaimerISPB,
			OwnerIspb:      req.OwnerISPB,
			ClaimerAccount: req.ClaimerAccount,
			ClaimType:      req.ClaimType,
			RequestId:      req.RequestID,
		})
		if err != nil {
			return mapGRPCError(err)
		}

		resp = result
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// ConfirmClaim confirms a claim (owner accepts)
func (c *ConnectClient) ConfirmClaim(ctx context.Context, claimID string, reason string, requestID string) (*connectv1.ConfirmClaimResponse, error) {
	var resp *connectv1.ConfirmClaimResponse

	err := c.executeWithRetry(ctx, func() error {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()

		req := &connectv1.ConfirmClaimRequest{
			ClaimId:   claimID,
			RequestId: requestID,
		}
		if reason != "" {
			req.Reason = &reason
		}

		result, err := c.client.ConfirmClaim(ctxWithTimeout, req)
		if err != nil {
			return mapGRPCError(err)
		}

		resp = result
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CancelClaim cancels a claim (owner rejects or claimer withdraws)
func (c *ConnectClient) CancelClaim(ctx context.Context, claimID string, reason string, requestID string) (*connectv1.CancelClaimResponse, error) {
	var resp *connectv1.CancelClaimResponse

	err := c.executeWithRetry(ctx, func() error {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()

		result, err := c.client.CancelClaim(ctxWithTimeout, &connectv1.CancelClaimRequest{
			ClaimId:   claimID,
			Reason:    reason,
			RequestId: requestID,
		})
		if err != nil {
			return mapGRPCError(err)
		}

		resp = result
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetClaim retrieves a claim by its ID
func (c *ConnectClient) GetClaim(ctx context.Context, claimID string, requestID string) (*connectv1.Claim, error) {
	var claim *connectv1.Claim
	var found bool

	err := c.executeWithRetry(ctx, func() error {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()

		resp, err := c.client.GetClaim(ctxWithTimeout, &connectv1.GetClaimRequest{
			ClaimId:   claimID,
			RequestId: requestID,
		})
		if err != nil {
			return mapGRPCError(err)
		}

		claim = resp.Claim
		found = resp.Found
		return nil
	})

	if err != nil {
		return nil, err
	}

	if !found {
		return nil, ErrClaimNotFound
	}

	return claim, nil
}

// ListClaimsFilters holds filters for listing claims
type ListClaimsFilters struct {
	EntryID     *string
	ClaimerISPB *string
	OwnerISPB   *string
	Status      *commonv1.ClaimStatus
	Limit       int32
	Offset      int32
	RequestID   string
}

// ListClaims retrieves a list of claims with optional filters
func (c *ConnectClient) ListClaims(ctx context.Context, filters ListClaimsFilters) ([]*connectv1.Claim, int32, error) {
	var claims []*connectv1.Claim
	var totalCount int32

	err := c.executeWithRetry(ctx, func() error {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()

		req := &connectv1.ListClaimsRequest{
			Limit:     filters.Limit,
			Offset:    filters.Offset,
			RequestId: filters.RequestID,
		}

		if filters.EntryID != nil {
			req.EntryId = filters.EntryID
		}
		if filters.ClaimerISPB != nil {
			req.ClaimerIspb = filters.ClaimerISPB
		}
		if filters.OwnerISPB != nil {
			req.OwnerIspb = filters.OwnerISPB
		}
		if filters.Status != nil {
			req.Status = filters.Status
		}

		resp, err := c.client.ListClaims(ctxWithTimeout, req)
		if err != nil {
			return mapGRPCError(err)
		}

		claims = resp.Claims
		totalCount = resp.TotalCount
		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return claims, totalCount, nil
}

// =============================================================================
// INFRACTION OPERATIONS (6 methods)
// =============================================================================

// CreateInfractionRequest holds parameters for creating an infraction
type CreateInfractionRequest struct {
	Key             *commonv1.DictKey
	ParticipantISPB string
	InfractionType  connectv1.CreateInfractionRequest_InfractionType
	Description     string
	ReporterISPB    string
	RequestID       string
}

// CreateInfraction creates a new infraction report
func (c *ConnectClient) CreateInfraction(ctx context.Context, req CreateInfractionRequest) (*connectv1.CreateInfractionResponse, error) {
	var resp *connectv1.CreateInfractionResponse

	err := c.executeWithRetry(ctx, func() error {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()

		result, err := c.client.CreateInfraction(ctxWithTimeout, &connectv1.CreateInfractionRequest{
			Key:              req.Key,
			ParticipantIspb:  req.ParticipantISPB,
			InfractionType:   req.InfractionType,
			Description:      req.Description,
			ReporterIspb:     req.ReporterISPB,
			RequestId:        req.RequestID,
		})
		if err != nil {
			return mapGRPCError(err)
		}

		resp = result
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// InvestigateInfraction processes an infraction investigation
func (c *ConnectClient) InvestigateInfraction(ctx context.Context, infractionID string, decision connectv1.InvestigateInfractionRequest_InvestigationDecision, analystNotes string, requestID string) (*connectv1.InvestigateInfractionResponse, error) {
	var resp *connectv1.InvestigateInfractionResponse

	err := c.executeWithRetry(ctx, func() error {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()

		result, err := c.client.InvestigateInfraction(ctxWithTimeout, &connectv1.InvestigateInfractionRequest{
			InfractionId: infractionID,
			Decision:     decision,
			AnalystNotes: analystNotes,
			RequestId:    requestID,
		})
		if err != nil {
			return mapGRPCError(err)
		}

		resp = result
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// ResolveInfraction marks an infraction as resolved
func (c *ConnectClient) ResolveInfraction(ctx context.Context, infractionID string, resolution string, requestID string) (*connectv1.ResolveInfractionResponse, error) {
	var resp *connectv1.ResolveInfractionResponse

	err := c.executeWithRetry(ctx, func() error {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()

		result, err := c.client.ResolveInfraction(ctxWithTimeout, &connectv1.ResolveInfractionRequest{
			InfractionId: infractionID,
			Resolution:   resolution,
			RequestId:    requestID,
		})
		if err != nil {
			return mapGRPCError(err)
		}

		resp = result
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// DismissInfraction dismisses an infraction as unfounded
func (c *ConnectClient) DismissInfraction(ctx context.Context, infractionID string, reason string, requestID string) (*connectv1.DismissInfractionResponse, error) {
	var resp *connectv1.DismissInfractionResponse

	err := c.executeWithRetry(ctx, func() error {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()

		result, err := c.client.DismissInfraction(ctxWithTimeout, &connectv1.DismissInfractionRequest{
			InfractionId: infractionID,
			Reason:       reason,
			RequestId:    requestID,
		})
		if err != nil {
			return mapGRPCError(err)
		}

		resp = result
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetInfraction retrieves an infraction by its ID
func (c *ConnectClient) GetInfraction(ctx context.Context, infractionID string, requestID string) (*connectv1.Infraction, error) {
	var infraction *connectv1.Infraction
	var found bool

	err := c.executeWithRetry(ctx, func() error {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()

		resp, err := c.client.GetInfraction(ctxWithTimeout, &connectv1.GetInfractionRequest{
			InfractionId: infractionID,
			RequestId:    requestID,
		})
		if err != nil {
			return mapGRPCError(err)
		}

		infraction = resp.Infraction
		found = resp.Found
		return nil
	})

	if err != nil {
		return nil, err
	}

	if !found {
		return nil, ErrInfractionNotFound
	}

	return infraction, nil
}

// ListInfractionsFilters holds filters for listing infractions
type ListInfractionsFilters struct {
	ParticipantISPB *string
	Status          *connectv1.InfractionStatus
	Limit           int32
	Offset          int32
	RequestID       string
}

// ListInfractions retrieves a list of infractions with optional filters
func (c *ConnectClient) ListInfractions(ctx context.Context, filters ListInfractionsFilters) ([]*connectv1.Infraction, int32, error) {
	var infractions []*connectv1.Infraction
	var totalCount int32

	err := c.executeWithRetry(ctx, func() error {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()

		req := &connectv1.ListInfractionsRequest{
			Limit:     filters.Limit,
			Offset:    filters.Offset,
			RequestId: filters.RequestID,
		}

		if filters.ParticipantISPB != nil {
			req.ParticipantIspb = filters.ParticipantISPB
		}
		if filters.Status != nil {
			req.Status = filters.Status
		}

		resp, err := c.client.ListInfractions(ctxWithTimeout, req)
		if err != nil {
			return mapGRPCError(err)
		}

		infractions = resp.Infractions
		totalCount = resp.TotalCount
		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return infractions, totalCount, nil
}

// =============================================================================
// HEALTH CHECK (1 method)
// =============================================================================

// HealthCheck performs a health check on the Connect service
func (c *ConnectClient) HealthCheck(ctx context.Context) (*connectv1.HealthCheckResponse, error) {
	var resp *connectv1.HealthCheckResponse

	// Health check doesn't use circuit breaker to allow monitoring even when circuit is open
	err := c.retryPolicy.Execute(ctx, func() error {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()

		result, err := c.client.HealthCheck(ctxWithTimeout, &emptypb.Empty{})
		if err != nil {
			return mapGRPCError(err)
		}

		resp = result
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// =============================================================================
// HELPER METHODS
// =============================================================================

// executeWithRetry executes a function with retry and circuit breaker protection
func (c *ConnectClient) executeWithRetry(ctx context.Context, fn func() error) error {
	return c.circuitBreaker.Call(func() error {
		return c.retryPolicy.Execute(ctx, fn)
	})
}

// healthCheckLoop periodically checks the health of the Connect service
func (c *ConnectClient) healthCheckLoop() {
	ticker := time.NewTicker(c.config.HealthCheckPeriod)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		resp, err := c.HealthCheck(ctx)
		cancel()

		if err != nil {
			log.Printf("[ConnectClient] Health check failed: %v", err)
			continue
		}

		if resp.Status != connectv1.HealthCheckResponse_HEALTH_STATUS_HEALTHY {
			log.Printf("[ConnectClient] Service unhealthy: %s", resp.Status)
		}
	}
}

// GetCircuitBreakerState returns the current circuit breaker state
func (c *ConnectClient) GetCircuitBreakerState() State {
	return c.circuitBreaker.GetState()
}

// GetCircuitBreakerMetrics returns circuit breaker metrics
func (c *ConnectClient) GetCircuitBreakerMetrics() map[string]interface{} {
	return c.circuitBreaker.GetMetrics()
}

// ResetCircuitBreaker manually resets the circuit breaker to closed state
func (c *ConnectClient) ResetCircuitBreaker() {
	c.circuitBreaker.Reset()
}
