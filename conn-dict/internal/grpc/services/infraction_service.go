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

// InfractionService implements the gRPC service for DICT Infraction operations
// This service handles fraud reports, infraction investigations, and escalations to Bacen
type InfractionService struct {
	// Note: UnimplementedConnDictServiceServer will be added when proto is generated
	// pb.UnimplementedConnDictServiceServer

	temporalClient   client.Client
	infractionRepo   *repositories.InfractionRepository
	logger           *logrus.Logger
}

// NewInfractionService creates a new InfractionService instance
func NewInfractionService(
	temporalClient client.Client,
	infractionRepo *repositories.InfractionRepository,
	logger *logrus.Logger,
) *InfractionService {
	return &InfractionService{
		temporalClient: temporalClient,
		infractionRepo: infractionRepo,
		logger:         logger,
	}
}

// CreateInfraction handles infraction creation by starting the InvestigateInfractionWorkflow
// This is the main entry point for reporting fraud or infractions in the DICT system
//
// Request should contain:
// - infraction_id: External infraction identifier
// - key: PIX key involved in the infraction
// - type: Type of infraction (FRAUD, ACCOUNT_CLOSED, INCORRECT_DATA, UNAUTHORIZED_USE, DUPLICATE_KEY, OTHER)
// - description: Detailed description of the infraction
// - reporter_ispb: Bank ISPB reporting the infraction (8 digits)
// - reported_ispb: Bank ISPB being reported (optional, 8 digits)
// - evidence_urls: Array of evidence URLs (optional)
// - related_entry_id: Related entry ID (optional)
// - related_claim_id: Related claim ID (optional)
//
// Returns:
// - workflow_id: Temporal workflow ID for tracking
// - infraction_id: Infraction identifier
// - status: Initial status (OPEN)
// - message: Success message
//
// Error codes:
// - InvalidArgument: Missing or invalid required fields
// - Internal: Temporal workflow start failed or database error
func (s *InfractionService) CreateInfraction(ctx context.Context, req interface{}) (interface{}, error) {
	s.logger.Info("CreateInfraction called")

	// TODO: When proto is generated, replace interface{} with actual proto type
	// Example: req *pb.CreateInfractionRequest

	// For now, we'll use a map for demonstration
	// In production, this will be strongly typed proto messages
	reqMap, ok := req.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid request format")
	}

	// Extract and validate required fields
	infractionID, ok := reqMap["infraction_id"].(string)
	if !ok || infractionID == "" {
		infractionID = uuid.New().String() // Generate if not provided
	}

	key, ok := reqMap["key"].(string)
	if !ok || key == "" {
		return nil, status.Error(codes.InvalidArgument, "key is required")
	}

	infractionType, ok := reqMap["type"].(string)
	if !ok || infractionType == "" {
		return nil, status.Error(codes.InvalidArgument, "type is required")
	}

	// Validate infraction type
	validTypes := map[string]bool{
		"FRAUD":             true,
		"ACCOUNT_CLOSED":    true,
		"INCORRECT_DATA":    true,
		"UNAUTHORIZED_USE":  true,
		"DUPLICATE_KEY":     true,
		"OTHER":             true,
	}
	if !validTypes[infractionType] {
		return nil, status.Errorf(codes.InvalidArgument,
			"invalid infraction type: %s (must be FRAUD, ACCOUNT_CLOSED, INCORRECT_DATA, UNAUTHORIZED_USE, DUPLICATE_KEY, or OTHER)",
			infractionType)
	}

	description, ok := reqMap["description"].(string)
	if !ok || description == "" {
		return nil, status.Error(codes.InvalidArgument, "description is required")
	}

	reporterISPB, ok := reqMap["reporter_ispb"].(string)
	if !ok || reporterISPB == "" {
		return nil, status.Error(codes.InvalidArgument, "reporter_ispb is required")
	}

	// Validate ISPB format
	if len(reporterISPB) != 8 {
		return nil, status.Errorf(codes.InvalidArgument, "reporter_ispb must be 8 digits, got %d", len(reporterISPB))
	}

	// Extract optional fields
	reportedISPB := getStringOrEmpty(reqMap, "reported_ispb")
	if reportedISPB != "" && len(reportedISPB) != 8 {
		return nil, status.Errorf(codes.InvalidArgument, "reported_ispb must be 8 digits, got %d", len(reportedISPB))
	}

	// Reporter and reported participant must be different
	if reportedISPB != "" && reporterISPB == reportedISPB {
		return nil, status.Error(codes.InvalidArgument, "reporter_ispb and reported_ispb must be different")
	}

	relatedEntryID := getStringOrEmpty(reqMap, "related_entry_id")
	relatedClaimID := getStringOrEmpty(reqMap, "related_claim_id")

	// Extract evidence URLs (array)
	var evidenceURLs []string
	if evidenceList, ok := reqMap["evidence_urls"].([]interface{}); ok {
		for _, ev := range evidenceList {
			if evStr, ok := ev.(string); ok {
				evidenceURLs = append(evidenceURLs, evStr)
			}
		}
	}

	// Build workflow input
	workflowInput := workflows.InvestigateInfractionInput{
		InfractionID:        infractionID,
		Key:                 key,
		Type:                infractionType,
		Description:         description,
		ReporterISPB:        reporterISPB,
		ReportedISPB:        reportedISPB,
		EvidenceURLs:        evidenceURLs,
		RelatedEntryID:      relatedEntryID,
		RelatedClaimID:      relatedClaimID,
	}

	// Start Temporal workflow for async infraction investigation
	workflowID := fmt.Sprintf("infraction-%s", infractionID)
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "dict-task-queue",
		// Workflow timeout: 30 days (infraction investigation period)
		WorkflowExecutionTimeout: 30 * 24 * time.Hour,
	}

	we, err := s.temporalClient.ExecuteWorkflow(ctx, workflowOptions, workflows.InvestigateInfractionWorkflow, workflowInput)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"infraction_id": infractionID,
			"key":           key,
			"workflow_id":   workflowID,
		}).Error("Failed to start InvestigateInfractionWorkflow")
		return nil, status.Errorf(codes.Internal, "failed to start infraction investigation workflow: %v", err)
	}

	s.logger.WithFields(logrus.Fields{
		"infraction_id": infractionID,
		"workflow_id":   we.GetID(),
		"run_id":        we.GetRunID(),
		"key":           key,
		"type":          infractionType,
	}).Info("InvestigateInfractionWorkflow started successfully")

	// Return response (will be proto message when generated)
	return map[string]interface{}{
		"infraction_id": infractionID,
		"workflow_id":   we.GetID(),
		"run_id":        we.GetRunID(),
		"status":        "OPEN",
		"message":       "Infraction created successfully. Investigation workflow started.",
	}, nil
}

// InvestigateInfraction handles sending investigation decision signal to the workflow
// This RPC is called when an investigation decision is made (RESOLVE, DISMISS, or ESCALATE)
//
// Request should contain:
// - infraction_id: Infraction identifier
// - decision: Investigation decision (RESOLVE, DISMISS, ESCALATE)
// - notes: Decision notes (required)
//
// Returns:
// - infraction_id: Infraction identifier
// - decision: Decision sent
// - message: Success message
//
// Error codes:
// - InvalidArgument: Missing or invalid required fields
// - NotFound: Workflow not found
// - Internal: Signal send failed
func (s *InfractionService) InvestigateInfraction(ctx context.Context, req interface{}) (interface{}, error) {
	s.logger.Info("InvestigateInfraction called")

	reqMap, ok := req.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid request format")
	}

	// Extract and validate fields
	infractionID, ok := reqMap["infraction_id"].(string)
	if !ok || infractionID == "" {
		return nil, status.Error(codes.InvalidArgument, "infraction_id is required")
	}

	decision, ok := reqMap["decision"].(string)
	if !ok || decision == "" {
		return nil, status.Error(codes.InvalidArgument, "decision is required")
	}

	// Validate decision
	validDecisions := map[string]bool{
		"RESOLVE":  true,
		"DISMISS":  true,
		"ESCALATE": true,
	}
	if !validDecisions[decision] {
		return nil, status.Errorf(codes.InvalidArgument,
			"invalid decision: %s (must be RESOLVE, DISMISS, or ESCALATE)", decision)
	}

	notes, ok := reqMap["notes"].(string)
	if !ok || notes == "" {
		return nil, status.Error(codes.InvalidArgument, "notes are required")
	}

	// Build signal data
	investigationDecision := workflows.InvestigationDecision{
		Decision: decision,
		Notes:    notes,
	}

	// Send signal to workflow
	workflowID := fmt.Sprintf("infraction-%s", infractionID)
	err := s.temporalClient.SignalWorkflow(ctx, workflowID, "", "investigation_complete", investigationDecision)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"infraction_id": infractionID,
			"workflow_id":   workflowID,
			"decision":      decision,
		}).Error("Failed to send investigation_complete signal")
		return nil, status.Errorf(codes.Internal, "failed to send investigation decision: %v", err)
	}

	s.logger.WithFields(logrus.Fields{
		"infraction_id": infractionID,
		"workflow_id":   workflowID,
		"decision":      decision,
	}).Info("Investigation decision signal sent successfully")

	return map[string]interface{}{
		"infraction_id": infractionID,
		"decision":      decision,
		"message":       fmt.Sprintf("Investigation decision '%s' sent successfully", decision),
	}, nil
}

// ResolveInfraction is a convenience RPC for resolving an infraction
// Internally sends InvestigationDecision signal with decision="RESOLVE"
//
// Request should contain:
// - infraction_id: Infraction identifier
// - resolution_notes: Resolution notes (required)
//
// Returns:
// - infraction_id: Infraction identifier
// - status: New status (RESOLVED)
// - message: Success message
//
// Error codes:
// - InvalidArgument: Missing or invalid required fields
// - NotFound: Workflow not found
// - Internal: Signal send failed
func (s *InfractionService) ResolveInfraction(ctx context.Context, req interface{}) (interface{}, error) {
	s.logger.Info("ResolveInfraction called")

	reqMap, ok := req.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid request format")
	}

	infractionID, ok := reqMap["infraction_id"].(string)
	if !ok || infractionID == "" {
		return nil, status.Error(codes.InvalidArgument, "infraction_id is required")
	}

	resolutionNotes, ok := reqMap["resolution_notes"].(string)
	if !ok || resolutionNotes == "" {
		return nil, status.Error(codes.InvalidArgument, "resolution_notes are required")
	}

	// Send RESOLVE decision
	investigationDecision := workflows.InvestigationDecision{
		Decision: "RESOLVE",
		Notes:    resolutionNotes,
	}

	workflowID := fmt.Sprintf("infraction-%s", infractionID)
	err := s.temporalClient.SignalWorkflow(ctx, workflowID, "", "investigation_complete", investigationDecision)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"infraction_id": infractionID,
			"workflow_id":   workflowID,
		}).Error("Failed to send resolve signal")
		return nil, status.Errorf(codes.Internal, "failed to resolve infraction: %v", err)
	}

	s.logger.WithFields(logrus.Fields{
		"infraction_id": infractionID,
		"workflow_id":   workflowID,
	}).Info("Infraction resolved successfully")

	return map[string]interface{}{
		"infraction_id": infractionID,
		"status":        "RESOLVED",
		"message":       "Infraction resolved successfully",
	}, nil
}

// DismissInfraction is a convenience RPC for dismissing an infraction
// Internally sends InvestigationDecision signal with decision="DISMISS"
//
// Request should contain:
// - infraction_id: Infraction identifier
// - dismissal_notes: Dismissal notes (required)
//
// Returns:
// - infraction_id: Infraction identifier
// - status: New status (DISMISSED)
// - message: Success message
//
// Error codes:
// - InvalidArgument: Missing or invalid required fields
// - NotFound: Workflow not found
// - Internal: Signal send failed
func (s *InfractionService) DismissInfraction(ctx context.Context, req interface{}) (interface{}, error) {
	s.logger.Info("DismissInfraction called")

	reqMap, ok := req.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid request format")
	}

	infractionID, ok := reqMap["infraction_id"].(string)
	if !ok || infractionID == "" {
		return nil, status.Error(codes.InvalidArgument, "infraction_id is required")
	}

	dismissalNotes, ok := reqMap["dismissal_notes"].(string)
	if !ok || dismissalNotes == "" {
		return nil, status.Error(codes.InvalidArgument, "dismissal_notes are required")
	}

	// Send DISMISS decision
	investigationDecision := workflows.InvestigationDecision{
		Decision: "DISMISS",
		Notes:    dismissalNotes,
	}

	workflowID := fmt.Sprintf("infraction-%s", infractionID)
	err := s.temporalClient.SignalWorkflow(ctx, workflowID, "", "investigation_complete", investigationDecision)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"infraction_id": infractionID,
			"workflow_id":   workflowID,
		}).Error("Failed to send dismiss signal")
		return nil, status.Errorf(codes.Internal, "failed to dismiss infraction: %v", err)
	}

	s.logger.WithFields(logrus.Fields{
		"infraction_id": infractionID,
		"workflow_id":   workflowID,
	}).Info("Infraction dismissed successfully")

	return map[string]interface{}{
		"infraction_id": infractionID,
		"status":        "DISMISSED",
		"message":       "Infraction dismissed successfully",
	}, nil
}

// GetInfraction retrieves a single infraction by ID
// This is a synchronous database query (no workflow involved)
//
// Request should contain:
// - infraction_id: Infraction identifier
//
// Returns:
// - Infraction details (id, key, type, status, timestamps, etc.)
//
// Error codes:
// - InvalidArgument: Missing infraction_id
// - NotFound: Infraction not found
// - Internal: Database query failed
func (s *InfractionService) GetInfraction(ctx context.Context, req interface{}) (interface{}, error) {
	s.logger.Info("GetInfraction called")

	reqMap, ok := req.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid request format")
	}

	infractionID, ok := reqMap["infraction_id"].(string)
	if !ok || infractionID == "" {
		return nil, status.Error(codes.InvalidArgument, "infraction_id is required")
	}

	// Query database
	infraction, err := s.infractionRepo.GetByInfractionID(ctx, infractionID)
	if err != nil {
		s.logger.WithError(err).WithField("infraction_id", infractionID).Error("Failed to get infraction")
		return nil, status.Errorf(codes.NotFound, "infraction not found: %s", infractionID)
	}

	return infractionToProtoMap(infraction), nil
}

// ListInfractions retrieves a paginated list of infractions
// This is a synchronous database query with filtering options
//
// Request should contain:
// - limit: Number of results (default: 20, max: 100)
// - offset: Pagination offset (default: 0)
// - key: Filter by PIX key (optional)
// - reporter_ispb: Filter by reporter ISPB (optional)
// - status: Filter by status (optional)
//
// Returns:
// - infractions: Array of infraction summaries
// - total_count: Total number of infractions
// - limit: Applied limit
// - offset: Applied offset
//
// Error codes:
// - InvalidArgument: Invalid parameters
// - Internal: Database query failed
func (s *InfractionService) ListInfractions(ctx context.Context, req interface{}) (interface{}, error) {
	s.logger.Info("ListInfractions called")

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

	// Extract optional filters
	key := getStringOrEmpty(reqMap, "key")
	reporterISPB := getStringOrEmpty(reqMap, "reporter_ispb")
	statusFilter := getStringOrEmpty(reqMap, "status")

	var infractions []*entities.Infraction
	var err error

	// Apply filters
	if key != "" {
		// List by key
		infractions, err = s.infractionRepo.ListByKey(ctx, key, limit, offset)
	} else if reporterISPB != "" {
		// List by reporter
		if len(reporterISPB) != 8 {
			return nil, status.Errorf(codes.InvalidArgument, "reporter_ispb must be 8 digits, got %d", len(reporterISPB))
		}
		infractions, err = s.infractionRepo.ListByReporter(ctx, reporterISPB, limit, offset)
	} else if statusFilter != "" {
		// List by status
		infractions, err = s.infractionRepo.ListByStatus(ctx, entities.InfractionStatus(statusFilter), limit, offset)
	} else {
		// List open infractions (default)
		infractions, err = s.infractionRepo.ListOpen(ctx, limit)
	}

	if err != nil {
		s.logger.WithError(err).Error("Failed to list infractions")
		return nil, status.Error(codes.Internal, "failed to list infractions")
	}

	// Convert to proto format
	infractionMaps := make([]map[string]interface{}, len(infractions))
	for i, infraction := range infractions {
		infractionMaps[i] = infractionToProtoMap(infraction)
	}

	return map[string]interface{}{
		"infractions":  infractionMaps,
		"total_count":  len(infractionMaps),
		"limit":        limit,
		"offset":       offset,
	}, nil
}

// Helper functions

// infractionToProtoMap converts Infraction entity to proto-like map (temporary until proto is generated)
func infractionToProtoMap(i *entities.Infraction) map[string]interface{} {
	result := map[string]interface{}{
		"id":                    i.ID.String(),
		"infraction_id":         i.InfractionID,
		"key":                   i.Key,
		"type":                  string(i.Type),
		"description":           i.Description,
		"evidence_urls":         i.EvidenceURLs,
		"reporter_participant":  i.ReporterParticipant,
		"status":                string(i.Status),
		"reported_at":           i.ReportedAt.Format(time.RFC3339),
		"created_at":            i.CreatedAt.Format(time.RFC3339),
		"updated_at":            i.UpdatedAt.Format(time.RFC3339),
	}

	// Add optional fields if present
	if i.EntryID != nil {
		result["entry_id"] = *i.EntryID
	}
	if i.ClaimID != nil {
		result["claim_id"] = *i.ClaimID
	}
	if i.ReportedParticipant != nil {
		result["reported_participant"] = *i.ReportedParticipant
	}
	if i.ResolutionNotes != nil {
		result["resolution_notes"] = *i.ResolutionNotes
	}
	if i.InvestigatedAt != nil {
		result["investigated_at"] = i.InvestigatedAt.Format(time.RFC3339)
	}
	if i.ResolvedAt != nil {
		result["resolved_at"] = i.ResolvedAt.Format(time.RFC3339)
	}
	if i.DeletedAt != nil {
		result["deleted_at"] = i.DeletedAt.Format(time.RFC3339)
	}

	return result
}
