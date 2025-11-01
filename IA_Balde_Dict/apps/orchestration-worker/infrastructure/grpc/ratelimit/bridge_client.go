package ratelimit

import (
	"context"
	"fmt"
	"time"

	pb "github.com/lb-conn/rsfn-connect-bacen-bridge/proto/bacen/dict/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/lb-conn/connector-dict/domain/ratelimit"
)

const (
	tracerName = "github.com/lb-conn/connector-dict/infrastructure/grpc/ratelimit"
)

var tracer = otel.Tracer(tracerName)

// BridgeRateLimitClient wraps Bridge gRPC client for rate limit operations
// Provides type conversion from proto to domain entities and error handling
type BridgeRateLimitClient struct {
	grpcClient pb.RateLimitServiceClient
}

// NewBridgeRateLimitClient creates a new Bridge rate limit client
// Reuses existing grpc.ClientConn from grpcGateway
func NewBridgeRateLimitClient(conn *grpc.ClientConn) *BridgeRateLimitClient {
	return &BridgeRateLimitClient{
		grpcClient: pb.NewRateLimitServiceClient(conn),
	}
}

// GetAllPolicies retrieves all rate limit policies for the PSP
// Returns domain entities converted from gRPC response
func (c *BridgeRateLimitClient) GetAllPolicies(ctx context.Context) ([]*ratelimit.Policy, *ratelimit.PolicyState, string, error) {
	ctx, span := tracer.Start(ctx, "BridgeRateLimitClient.GetAllPolicies")
	defer span.End()

	// Call Bridge gRPC
	req := &pb.GetRateLimitPoliciesRequest{}

	resp, err := c.grpcClient.GetRateLimitPolicies(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Bridge gRPC call failed")
		return nil, nil, "", handleBridgeError(err)
	}

	// Extract response metadata
	responseTime := resp.ResponseTime.AsTime()
	pspCategory := resp.Category

	span.SetAttributes(
		attribute.String("psp_category", pspCategory),
		attribute.Int("policy_count", len(resp.Policies)),
		attribute.String("response_time", responseTime.Format(time.RFC3339)),
	)

	// Convert proto policies to domain policies
	policies := make([]*ratelimit.Policy, 0, len(resp.Policies))
	states := make([]*ratelimit.PolicyState, 0, len(resp.Policies))

	for _, p := range resp.Policies {
		// Create Policy entity (static configuration)
		policy, err := ratelimit.NewPolicy(
			p.Name,                    // EndpointID
			p.Name,                    // EndpointPath (use name as path for now)
			"",                        // HTTPMethod (not provided by DICT)
			int(p.Capacity),
			int(p.RefillTokens),
			int(p.RefillPeriodSec),
			p.PolicyCategory,          // May be empty
		)
		if err != nil {
			span.RecordError(err)
			return nil, nil, "", fmt.Errorf("failed to create policy for %s: %w", p.Name, err)
		}

		policies = append(policies, policy)

		// Create PolicyState entity (current snapshot)
		state, err := ratelimit.NewPolicyState(
			p.Name,                     // EndpointID
			int(p.AvailableTokens),
			int(p.Capacity),
			int(p.RefillTokens),
			int(p.RefillPeriodSec),
			pspCategory,                // PSP category from response
			responseTime,               // From DICT <ResponseTime>
		)
		if err != nil {
			span.RecordError(err)
			return nil, nil, "", fmt.Errorf("failed to create state for %s: %w", p.Name, err)
		}

		states = append(states, state)
	}

	span.SetStatus(codes.Ok, "Policies retrieved successfully")

	// Return first state as representative (caller should iterate if needed)
	var firstState *ratelimit.PolicyState
	if len(states) > 0 {
		firstState = states[0]
	}

	return policies, firstState, pspCategory, nil
}

// GetPolicyState retrieves current state for a specific policy
func (c *BridgeRateLimitClient) GetPolicyState(ctx context.Context, policyName string) (*ratelimit.PolicyState, string, error) {
	ctx, span := tracer.Start(ctx, "BridgeRateLimitClient.GetPolicyState")
	defer span.End()

	span.SetAttributes(attribute.String("policy_name", policyName))

	// Call Bridge gRPC
	req := &pb.GetRateLimitPolicyRequest{
		PolicyName: policyName,
	}

	resp, err := c.grpcClient.GetRateLimitPolicy(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Bridge gRPC call failed")
		return nil, "", handleBridgeError(err)
	}

	// Extract response metadata
	responseTime := resp.ResponseTime.AsTime()
	pspCategory := resp.Category
	policy := resp.Policy

	span.SetAttributes(
		attribute.String("psp_category", pspCategory),
		attribute.Int("available_tokens", int(policy.AvailableTokens)),
		attribute.Int("capacity", int(policy.Capacity)),
		attribute.String("response_time", responseTime.Format(time.RFC3339)),
	)

	// Create PolicyState entity
	state, err := ratelimit.NewPolicyState(
		policy.Name,
		int(policy.AvailableTokens),
		int(policy.Capacity),
		int(policy.RefillTokens),
		int(policy.RefillPeriodSec),
		pspCategory,
		responseTime,
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create policy state")
		return nil, "", fmt.Errorf("failed to create state for %s: %w", policyName, err)
	}

	span.SetStatus(codes.Ok, "Policy state retrieved successfully")

	return state, pspCategory, nil
}

// handleBridgeError converts gRPC errors to domain errors
func handleBridgeError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("bridge unknown error: %w", err)
	}

	switch st.Code() {
	case 16: // Unauthenticated
		return &BridgeAuthError{
			Message: "bridge authentication failed - check mTLS certificates",
			Cause:   err,
		}

	case 7: // PermissionDenied
		return &BridgePermissionError{
			Message: "bridge permission denied - check PSP authorization",
			Cause:   err,
		}

	case 14: // Unavailable
		return &BridgeUnavailableError{
			Message:   "bridge or DICT unavailable - retry later",
			Cause:     err,
			Retryable: true,
		}

	case 4: // DeadlineExceeded
		return &BridgeTimeoutError{
			Message:   "bridge request timeout",
			Cause:     err,
			Retryable: true,
		}

	case 5: // NotFound
		return &PolicyNotFoundError{
			Message: "policy not found in DICT",
			Cause:   err,
		}

	case 13: // Internal
		return &BridgeInternalError{
			Message:   st.Message(),
			Cause:     err,
			Retryable: false,
		}

	default:
		return fmt.Errorf("bridge error: %s (code: %v)", st.Message(), st.Code())
	}
}

// Bridge-specific error types

type BridgeAuthError struct {
	Message string
	Cause   error
}

func (e *BridgeAuthError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Cause)
}

func (e *BridgeAuthError) Unwrap() error {
	return e.Cause
}

type BridgePermissionError struct {
	Message string
	Cause   error
}

func (e *BridgePermissionError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Cause)
}

func (e *BridgePermissionError) Unwrap() error {
	return e.Cause
}

type BridgeUnavailableError struct {
	Message   string
	Cause     error
	Retryable bool
}

func (e *BridgeUnavailableError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Cause)
}

func (e *BridgeUnavailableError) Unwrap() error {
	return e.Cause
}

type BridgeTimeoutError struct {
	Message   string
	Cause     error
	Retryable bool
}

func (e *BridgeTimeoutError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Cause)
}

func (e *BridgeTimeoutError) Unwrap() error {
	return e.Cause
}

type PolicyNotFoundError struct {
	Message string
	Cause   error
}

func (e *PolicyNotFoundError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Cause)
}

func (e *PolicyNotFoundError) Unwrap() error {
	return e.Cause
}

type BridgeInternalError struct {
	Message   string
	Cause     error
	Retryable bool
}

func (e *BridgeInternalError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Cause)
}

func (e *BridgeInternalError) Unwrap() error {
	return e.Cause
}
