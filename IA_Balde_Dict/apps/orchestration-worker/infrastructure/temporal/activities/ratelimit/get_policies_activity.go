package ratelimit

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	"github.com/lb-conn/connector-dict/apps/orchestration-worker/application/ports"
	"github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/grpc/ratelimit"
	domainRL "github.com/lb-conn/connector-dict/domain/ratelimit"
)

// GetPoliciesActivity retrieves rate limit policies from Bridge and stores them
type GetPoliciesActivity struct {
	bridgeClient *ratelimit.BridgeRateLimitClient
	policyRepo   ports.PolicyRepository
	stateRepo    ports.StateRepository
}

// GetPoliciesResult contains the result of getting policies
type GetPoliciesResult struct {
	PSPCategory  string `json:"psp_category"`
	PolicyCount  int    `json:"policy_count"`
	StateCount   int    `json:"state_count"`
}

// NewGetPoliciesActivity creates a new GetPoliciesActivity
func NewGetPoliciesActivity(
	bridgeClient *ratelimit.BridgeRateLimitClient,
	policyRepo ports.PolicyRepository,
	stateRepo ports.StateRepository,
) *GetPoliciesActivity {
	return &GetPoliciesActivity{
		bridgeClient: bridgeClient,
		policyRepo:   policyRepo,
		stateRepo:    stateRepo,
	}
}

// Execute retrieves policies from Bridge and stores them in the database
func (a *GetPoliciesActivity) Execute(ctx context.Context) (*GetPoliciesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("GetPoliciesActivity started")

	// Call Bridge gRPC to get all policies
	policies, _, pspCategory, err := a.bridgeClient.GetAllPolicies(ctx)
	if err != nil {
		logger.Error("Failed to get policies from Bridge", "error", err)
		return nil, fmt.Errorf("failed to get policies from bridge: %w", err)
	}

	logger.Info("Retrieved policies from Bridge",
		"category", pspCategory,
		"count", len(policies))

	// Store policies in database (upsert batch)
	if err := a.policyRepo.UpsertBatch(ctx, policies); err != nil {
		logger.Error("Failed to store policies", "error", err)
		return nil, fmt.Errorf("failed to store policies: %w", err)
	}

	logger.Info("Stored policies in database", "count", len(policies))

	// Create and store initial state snapshots for each policy
	states := make([]*domainRL.PolicyState, 0, len(policies))

	for _, policy := range policies {
		// Get current state from Bridge
		state, _, err := a.bridgeClient.GetPolicyState(ctx, policy.EndpointID)
		if err != nil {
			logger.Warn("Failed to get state for policy",
				"endpoint_id", policy.EndpointID,
				"error", err)
			continue
		}

		states = append(states, state)
	}

	// Store states in database
	if len(states) > 0 {
		if err := a.stateRepo.SaveBatch(ctx, states); err != nil {
			logger.Error("Failed to store states", "error", err)
			return nil, fmt.Errorf("failed to store states: %w", err)
		}

		logger.Info("Stored initial states in database", "count", len(states))
	}

	result := &GetPoliciesResult{
		PSPCategory: pspCategory,
		PolicyCount: len(policies),
		StateCount:  len(states),
	}

	logger.Info("GetPoliciesActivity completed successfully",
		"category", pspCategory,
		"policies", len(policies),
		"states", len(states))

	return result, nil
}
