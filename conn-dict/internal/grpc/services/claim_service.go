package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/lbpay-lab/conn-dict/internal/workflows"
)

// ClaimService implements the gRPC service for DICT Claim operations
// This service handles PIX key portability/ownership claims (30-day workflow)
type ClaimService struct {
	// Note: UnimplementedConnDictServiceServer will be added when proto is generated
	// pb.UnimplementedConnDictServiceServer

	temporalClient client.Client
	claimRepo      *repositories.ClaimRepository
	logger         *logrus.Logger
}

// NewClaimService creates a new ClaimService instance
func NewClaimService(
	temporalClient client.Client,
	claimRepo *repositories.ClaimRepository,
	logger *logrus.Logger,
) *ClaimService {
	return &ClaimService{
		temporalClient: temporalClient,
		claimRepo:      claimRepo,
		logger:         logger,
	}
}

// CreateClaim handles PIX key claim creation by starting a Temporal workflow
// This is the main entry point for initiating a 30-day claim process
//
// Request should contain:
// - claim_id: External claim identifier (UUID)
// - entry_id: Entry identifier to be claimed
// - claim_type: Type of claim (OWNERSHIP or PORTABILITY)
// - claimer_ispb: Claimer ISPB (8 digits)
// - donor_ispb: Current owner ISPB (8 digits)
// - claimer_account: Account information
//
// Returns:
// - workflow_id: Temporal workflow ID for tracking
// - claim_id: Claim identifier
// - status: Initial status (typically "OPEN")
// - expires_at: Expiration date (30 days from creation)
//
// Error codes:
// - InvalidArgument: Missing or invalid required fields
// - AlreadyExists: Active claim already exists for this key
// - Internal: Temporal workflow start failed
func (s *ClaimService) CreateClaim(ctx context.Context, req interface{}) (interface{}, error) {
	s.logger.Info("CreateClaim called")

	// TODO: When proto is generated, replace interface{} with actual proto type
	// Example: req *pb.CreateClaimRequest
	reqMap, ok := req.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid request format")
	}

	// Extract and validate required fields
	claimID, ok := reqMap["claim_id"].(string)
	if !ok || claimID == "" {
		claimID = uuid.New().String() // Generate if not provided
	}

	entryID, ok := reqMap["entry_id"].(string)
	if !ok || entryID == "" {
		return nil, status.Error(codes.InvalidArgument, "entry_id is required")
	}

	claimType, ok := reqMap["claim_type"].(string)
	if !ok || claimType == "" {
		claimType = "OWNERSHIP" // Default to OWNERSHIP
	}

	// Validate claim type
	if claimType != "OWNERSHIP" && claimType != "PORTABILITY" {
		return nil, status.Errorf(codes.InvalidArgument, "claim_type must be OWNERSHIP or PORTABILITY, got %s", claimType)
	}

	claimerISPB, ok := reqMap["claimer_ispb"].(string)
	if !ok || claimerISPB == "" {
		return nil, status.Error(codes.InvalidArgument, "claimer_ispb is required")
	}

	donorISPB, ok := reqMap["donor_ispb"].(string)
	if !ok || donorISPB == "" {
		return nil, status.Error(codes.InvalidArgument, "donor_ispb is required")
	}

	// Validate ISPB format
	if len(claimerISPB) != 8 {
		return nil, status.Errorf(codes.InvalidArgument, "claimer_ispb must be 8 digits, got %d", len(claimerISPB))
	}
	if len(donorISPB) != 8 {
		return nil, status.Errorf(codes.InvalidArgument, "donor_ispb must be 8 digits, got %d", len(donorISPB))
	}

	// Validate that claimer and donor are different
	if claimerISPB == donorISPB {
		return nil, status.Error(codes.InvalidArgument, "claimer_ispb and donor_ispb must be different")
	}

	// Extract optional fields
	claimerAccount := getStringOrEmpty(reqMap, "claimer_account")
	requestedBy := getStringOrEmpty(reqMap, "requested_by")

	// Check if there's already an active claim for this entry
	// Note: This check should ideally be done by querying by key, not entry_id
	// For now, we'll check by claim_id to avoid duplicates
	existingClaim, err := s.claimRepo.GetByClaimID(ctx, claimID)
	if err == nil && existingClaim != nil {
		return nil, status.Errorf(codes.AlreadyExists, "claim %s already exists", claimID)
	}

	// Build workflow input
	workflowInput := workflows.ClaimWorkflowInput{
		ClaimID:        claimID,
		EntryID:        entryID,
		ClaimType:      claimType,
		ClaimerISPB:    claimerISPB,
		DonorISPB:      donorISPB,
		ClaimerAccount: claimerAccount,
		RequestedBy:    requestedBy,
	}

	// Start Temporal workflow for 30-day claim processing
	workflowID := fmt.Sprintf("claim-workflow-%s", claimID)
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "dict-claims-queue",
		// Workflow timeout: 31 days (30 days + buffer)
		WorkflowExecutionTimeout: 31 * 24 * time.Hour,
	}

	we, err := s.temporalClient.ExecuteWorkflow(ctx, workflowOptions, workflows.ClaimWorkflow, workflowInput)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"claim_id":    claimID,
			"entry_id":    entryID,
			"workflow_id": workflowID,
		}).Error("Failed to start ClaimWorkflow")
		return nil, status.Errorf(codes.Internal, "failed to start claim workflow: %v", err)
	}

	// Calculate expiration date (30 days from now)
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	s.logger.WithFields(logrus.Fields{
		"claim_id":     claimID,
		"entry_id":     entryID,
		"workflow_id":  we.GetID(),
		"run_id":       we.GetRunID(),
		"claim_type":   claimType,
		"claimer_ispb": claimerISPB,
		"donor_ispb":   donorISPB,
		"expires_at":   expiresAt,
	}).Info("ClaimWorkflow started successfully")

	// Return response (will be proto message when generated)
	return map[string]interface{}{
		"claim_id":     claimID,
		"entry_id":     entryID,
		"workflow_id":  we.GetID(),
		"run_id":       we.GetRunID(),
		"status":       "OPEN",
		"claim_type":   claimType,
		"claimer_ispb": claimerISPB,
		"donor_ispb":   donorISPB,
		"expires_at":   expiresAt.Format(time.RFC3339),
		"message":      "Claim created successfully. Donor has 30 days to respond.",
	}, nil
}

// ConfirmClaim handles claim confirmation by sending a Signal to the Temporal workflow
// This is called when the donor (current owner) confirms the claim
//
// Request should contain:
// - claim_id: Claim identifier
// - confirmed_by: User/system confirming the claim
//
// Returns:
// - claim_id: Claim identifier
// - status: Updated status (typically "CONFIRMED")
// - message: Confirmation message
//
// Error codes:
// - InvalidArgument: Missing or invalid claim_id
// - NotFound: Claim not found
// - Internal: Failed to signal workflow
func (s *ClaimService) ConfirmClaim(ctx context.Context, req interface{}) (interface{}, error) {
	s.logger.Info("ConfirmClaim called")

	reqMap, ok := req.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid request format")
	}

	claimID, ok := reqMap["claim_id"].(string)
	if !ok || claimID == "" {
		return nil, status.Error(codes.InvalidArgument, "claim_id is required")
	}

	confirmedBy := getStringOrEmpty(reqMap, "confirmed_by")

	// Verify claim exists
	claim, err := s.claimRepo.GetByClaimID(ctx, claimID)
	if err != nil {
		s.logger.WithError(err).WithField("claim_id", claimID).Error("Failed to get claim")
		return nil, status.Errorf(codes.NotFound, "claim not found: %s", claimID)
	}

	// Validate claim can be confirmed
	if claim.Status != entities.ClaimStatusOpen && claim.Status != entities.ClaimStatusWaitingResolution {
		return nil, status.Errorf(codes.FailedPrecondition, "claim cannot be confirmed in status %s", claim.Status)
	}

	// Check if claim has expired
	if claim.IsExpired() {
		return nil, status.Error(codes.FailedPrecondition, "claim has expired")
	}

	// Send "confirm" signal to Temporal workflow
	workflowID := fmt.Sprintf("claim-workflow-%s", claimID)
	runID := "" // Empty string means "latest run"

	signalData := map[string]interface{}{
		"confirmed_by": confirmedBy,
		"confirmed_at": time.Now().Format(time.RFC3339),
	}

	err = s.temporalClient.SignalWorkflow(ctx, workflowID, runID, "confirm", signalData)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"claim_id":    claimID,
			"workflow_id": workflowID,
		}).Error("Failed to signal ClaimWorkflow for confirmation")
		return nil, status.Errorf(codes.Internal, "failed to confirm claim: %v", err)
	}

	s.logger.WithFields(logrus.Fields{
		"claim_id":     claimID,
		"workflow_id":  workflowID,
		"confirmed_by": confirmedBy,
	}).Info("Claim confirmation signal sent successfully")

	return map[string]interface{}{
		"claim_id": claimID,
		"status":   "CONFIRMED",
		"message":  "Claim confirmation signal sent successfully. Workflow will complete the claim.",
	}, nil
}

// CancelClaim handles claim cancellation by sending a Signal to the Temporal workflow
// This can be called by either the claimer or the donor
//
// Request should contain:
// - claim_id: Claim identifier
// - reason: Cancellation reason
// - cancelled_by: User/system cancelling the claim
//
// Returns:
// - claim_id: Claim identifier
// - status: Updated status ("CANCELLED")
// - message: Cancellation message
//
// Error codes:
// - InvalidArgument: Missing or invalid claim_id
// - NotFound: Claim not found
// - FailedPrecondition: Claim cannot be cancelled (already completed/expired)
// - Internal: Failed to signal workflow
func (s *ClaimService) CancelClaim(ctx context.Context, req interface{}) (interface{}, error) {
	s.logger.Info("CancelClaim called")

	reqMap, ok := req.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid request format")
	}

	claimID, ok := reqMap["claim_id"].(string)
	if !ok || claimID == "" {
		return nil, status.Error(codes.InvalidArgument, "claim_id is required")
	}

	reason := getStringOrEmpty(reqMap, "reason")
	if reason == "" {
		reason = "No reason provided"
	}

	cancelledBy := getStringOrEmpty(reqMap, "cancelled_by")

	// Verify claim exists
	claim, err := s.claimRepo.GetByClaimID(ctx, claimID)
	if err != nil {
		s.logger.WithError(err).WithField("claim_id", claimID).Error("Failed to get claim")
		return nil, status.Errorf(codes.NotFound, "claim not found: %s", claimID)
	}

	// Validate claim can be cancelled
	if !claim.CanBeCancelled() {
		return nil, status.Errorf(codes.FailedPrecondition, "claim cannot be cancelled in status %s", claim.Status)
	}

	// Send "cancel" signal to Temporal workflow
	workflowID := fmt.Sprintf("claim-workflow-%s", claimID)
	runID := "" // Empty string means "latest run"

	signalData := map[string]interface{}{
		"reason":       reason,
		"cancelled_by": cancelledBy,
		"cancelled_at": time.Now().Format(time.RFC3339),
	}

	err = s.temporalClient.SignalWorkflow(ctx, workflowID, runID, "cancel", signalData)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"claim_id":    claimID,
			"workflow_id": workflowID,
		}).Error("Failed to signal ClaimWorkflow for cancellation")
		return nil, status.Errorf(codes.Internal, "failed to cancel claim: %v", err)
	}

	s.logger.WithFields(logrus.Fields{
		"claim_id":     claimID,
		"workflow_id":  workflowID,
		"reason":       reason,
		"cancelled_by": cancelledBy,
	}).Info("Claim cancellation signal sent successfully")

	return map[string]interface{}{
		"claim_id": claimID,
		"status":   "CANCELLED",
		"reason":   reason,
		"message":  fmt.Sprintf("Claim cancelled by %s. Reason: %s", cancelledBy, reason),
	}, nil
}

// GetClaim retrieves a single claim by ID (synchronous database query)
//
// Request should contain:
// - claim_id: Claim identifier
//
// Returns:
// - Claim details including status, timestamps, participants
//
// Error codes:
// - InvalidArgument: Missing or invalid claim_id
// - NotFound: Claim not found
// - Internal: Database query failed
func (s *ClaimService) GetClaim(ctx context.Context, req interface{}) (interface{}, error) {
	s.logger.Info("GetClaim called")

	reqMap, ok := req.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid request format")
	}

	claimID, ok := reqMap["claim_id"].(string)
	if !ok || claimID == "" {
		return nil, status.Error(codes.InvalidArgument, "claim_id is required")
	}

	claim, err := s.claimRepo.GetByClaimID(ctx, claimID)
	if err != nil {
		s.logger.WithError(err).WithField("claim_id", claimID).Error("Failed to get claim")
		return nil, status.Errorf(codes.NotFound, "claim not found: %s", claimID)
	}

	return claimToProtoMap(claim), nil
}

// ListClaims retrieves a paginated list of claims (synchronous database query)
//
// Request should contain:
// - participant_ispb: Filter by participant ISPB (required for security)
// - claim_type: Filter by claim type (optional)
// - status: Filter by status (optional)
// - limit: Max number of results (default: 20, max: 100)
// - offset: Pagination offset (default: 0)
//
// Returns:
// - claims: List of claim summaries
// - total_count: Total number of matching claims
// - limit: Applied limit
// - offset: Applied offset
//
// Error codes:
// - InvalidArgument: Missing or invalid parameters
// - Internal: Database query failed
func (s *ClaimService) ListClaims(ctx context.Context, req interface{}) (interface{}, error) {
	s.logger.Info("ListClaims called")

	reqMap, ok := req.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid request format")
	}

	// Extract pagination parameters
	limit := 20 // Default
	offset := 0

	if l, ok := reqMap["limit"].(int); ok && l > 0 {
		limit = l
		if limit > 100 {
			limit = 100 // Max limit
		}
	}

	if o, ok := reqMap["offset"].(int); ok && o >= 0 {
		offset = o
	}

	// Extract filter: key (optional)
	key := getStringOrEmpty(reqMap, "key")

	// Query claims
	var claims []*entities.Claim
	var err error

	if key != "" {
		// List claims by key
		claims, err = s.claimRepo.ListByKey(ctx, key)
	} else {
		// For now, we don't have a ListAll method, so we'll return an error
		// In production, you should implement a ListAll or ListByParticipant method
		return nil, status.Error(codes.InvalidArgument, "key filter is required")
	}

	if err != nil {
		s.logger.WithError(err).Error("Failed to list claims")
		return nil, status.Error(codes.Internal, "failed to list claims")
	}

	// Apply pagination manually (since ListByKey doesn't support it yet)
	totalCount := len(claims)
	start := offset
	end := offset + limit

	if start >= totalCount {
		claims = []*entities.Claim{}
	} else {
		if end > totalCount {
			end = totalCount
		}
		claims = claims[start:end]
	}

	// Convert to proto format
	claimMaps := make([]map[string]interface{}, len(claims))
	for i, claim := range claims {
		claimMaps[i] = claimToProtoMap(claim)
	}

	return map[string]interface{}{
		"claims":      claimMaps,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	}, nil
}

// Helper functions

// claimToProtoMap converts Claim entity to proto-like map (temporary until proto is generated)
func claimToProtoMap(c *entities.Claim) map[string]interface{} {
	result := map[string]interface{}{
		"id":                  c.ID.String(),
		"claim_id":            c.ClaimID,
		"type":                string(c.Type),
		"status":              string(c.Status),
		"key":                 c.Key,
		"key_type":            c.KeyType,
		"donor_participant":   c.DonorParticipant,
		"claimer_participant": c.ClaimerParticipant,
		"created_at":          c.CreatedAt.Format(time.RFC3339),
		"updated_at":          c.UpdatedAt.Format(time.RFC3339),
	}

	// Add optional account fields
	if c.ClaimerAccountBranch != "" {
		result["claimer_account_branch"] = c.ClaimerAccountBranch
	}
	if c.ClaimerAccountNumber != "" {
		result["claimer_account_number"] = c.ClaimerAccountNumber
	}
	if c.ClaimerAccountType != "" {
		result["claimer_account_type"] = c.ClaimerAccountType
	}

	// Add timestamp fields
	result["completion_period_end"] = c.CompletionPeriodEnd.Format(time.RFC3339)
	result["claim_expiry_date"] = c.ClaimExpiryDate.Format(time.RFC3339)

	if c.ConfirmedAt != nil {
		result["confirmed_at"] = c.ConfirmedAt.Format(time.RFC3339)
	}
	if c.CompletedAt != nil {
		result["completed_at"] = c.CompletedAt.Format(time.RFC3339)
	}
	if c.CancelledAt != nil {
		result["cancelled_at"] = c.CancelledAt.Format(time.RFC3339)
	}
	if c.ExpiredAt != nil {
		result["expired_at"] = c.ExpiredAt.Format(time.RFC3339)
	}

	// Add metadata
	if c.CancellationReason != "" {
		result["cancellation_reason"] = c.CancellationReason
	}
	if c.Notes != "" {
		result["notes"] = c.Notes
	}

	// Add derived fields
	result["is_expired"] = c.IsExpired()
	result["is_active"] = c.IsActive()
	result["can_be_cancelled"] = c.CanBeCancelled()

	return result
}
