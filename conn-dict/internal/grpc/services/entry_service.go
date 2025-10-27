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
)

// EntryService implements the gRPC service for DICT Entry operations
// This service handles PIX key registration, updates, and queries
type EntryService struct {
	// Note: UnimplementedConnDictServiceServer will be added when proto is generated
	// pb.UnimplementedConnDictServiceServer

	temporalClient client.Client
	entryRepo      *repositories.EntryRepository
	logger         *logrus.Logger
}

// NewEntryService creates a new EntryService instance
func NewEntryService(
	temporalClient client.Client,
	entryRepo *repositories.EntryRepository,
	logger *logrus.Logger,
) *EntryService {
	return &EntryService{
		temporalClient: temporalClient,
		entryRepo:      entryRepo,
		logger:         logger,
	}
}

// CreateEntry handles PIX key creation by starting a Temporal workflow
// This is the main entry point for registering new PIX keys in DICT
//
// Request should contain:
// - entry_id: External entry identifier
// - key: PIX key value (CPF, phone, email, etc.)
// - key_type: Type of PIX key
// - participant_ispb: Bank ISPB (8 digits)
// - account information (branch, number, type)
// - owner information (name, tax ID, type)
//
// Returns:
// - workflow_id: Temporal workflow ID for tracking
// - entry_id: Entry identifier
// - status: Initial status (typically "PROCESSING")
//
// Error codes:
// - InvalidArgument: Missing or invalid required fields
// - AlreadyExists: Key already registered
// - Internal: Temporal workflow start failed
func (s *EntryService) CreateEntry(ctx context.Context, req interface{}) (interface{}, error) {
	s.logger.Info("CreateEntry called")

	// TODO: When proto is generated, replace interface{} with actual proto type
	// Example: req *pb.CreateEntryRequest

	// For now, we'll use a map for demonstration
	// In production, this will be strongly typed proto messages
	reqMap, ok := req.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid request format")
	}

	// Extract and validate required fields
	entryID, ok := reqMap["entry_id"].(string)
	if !ok || entryID == "" {
		entryID = uuid.New().String() // Generate if not provided
	}

	key, ok := reqMap["key"].(string)
	if !ok || key == "" {
		return nil, status.Error(codes.InvalidArgument, "key is required")
	}

	keyType, ok := reqMap["key_type"].(string)
	if !ok || keyType == "" {
		return nil, status.Error(codes.InvalidArgument, "key_type is required")
	}

	participantISPB, ok := reqMap["participant_ispb"].(string)
	if !ok || participantISPB == "" {
		return nil, status.Error(codes.InvalidArgument, "participant_ispb is required")
	}

	// Validate ISPB format
	if len(participantISPB) != 8 {
		return nil, status.Errorf(codes.InvalidArgument, "participant_ispb must be 8 digits, got %d", len(participantISPB))
	}

	// Check if key already exists
	hasActiveKey, err := s.entryRepo.HasActiveKey(ctx, key)
	if err != nil {
		s.logger.WithError(err).Error("Failed to check if key exists")
		return nil, status.Error(codes.Internal, "failed to check key existence")
	}
	if hasActiveKey {
		return nil, status.Errorf(codes.AlreadyExists, "key %s is already registered", key)
	}

	// TODO: Remove Temporal workflow - CreateEntry should use Pulsar Consumer directly
	// This is a placeholder until Pulsar Consumer is implemented
	// See: ANALISE_SYNC_VS_ASYNC_OPERATIONS.md for architecture decision

	// For now, return a simulated response
	// In production: Core DICT will publish to Pulsar, Consumer will handle Bridge call
	workflowID := fmt.Sprintf("create-entry-%s", entryID)

	// TEMPORARY: Simulate workflow execution
	// TODO: Replace with actual Pulsar publish + immediate response
	var we client.WorkflowRun

	// Commented out old workflow code - to be replaced by Pulsar Consumer
	/*
	workflowInput := workflows.CreateEntryWorkflowInput{
		EntryID:         entryID,
		Key:             key,
		KeyType:         keyType,
		ParticipantISPB: participantISPB,
		AccountBranch:   getStringOrEmpty(reqMap, "account_branch"),
		AccountNumber:   getStringOrEmpty(reqMap, "account_number"),
		AccountType:     getStringOrEmpty(reqMap, "account_type"),
		OwnerType:       getStringOrEmpty(reqMap, "owner_type"),
		OwnerName:       getStringOrEmpty(reqMap, "owner_name"),
		OwnerTaxID:      getStringOrEmpty(reqMap, "owner_tax_id"),
		RequestedBy:     getStringOrEmpty(reqMap, "requested_by"),
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "dict-task-queue",
		WorkflowExecutionTimeout: 2 * time.Minute,
	}

	we, err = s.temporalClient.ExecuteWorkflow(ctx, workflowOptions, workflows.CreateEntryWorkflow, workflowInput)
	*/

	// Simulate error for now
	createErr := fmt.Errorf("CreateEntry workflow removed - use Pulsar Consumer (see GAP #2)")
	if createErr != nil {
		s.logger.WithError(createErr).WithFields(logrus.Fields{
			"entry_id":   entryID,
			"key":        key,
			"workflow_id": workflowID,
		}).Error("Failed to start CreateEntryWorkflow")
		return nil, status.Errorf(codes.Internal, "failed to start entry creation workflow: %v", createErr)
	}

	s.logger.WithFields(logrus.Fields{
		"entry_id":    entryID,
		"workflow_id": we.GetID(),
		"run_id":      we.GetRunID(),
	}).Info("CreateEntryWorkflow started successfully")

	// Return response (will be proto message when generated)
	return map[string]interface{}{
		"entry_id":    entryID,
		"workflow_id": we.GetID(),
		"run_id":      we.GetRunID(),
		"status":      "PROCESSING",
		"message":     "Entry creation workflow started successfully",
	}, nil
}

// UpdateEntry handles PIX key updates (stub - will be implemented later)
func (s *EntryService) UpdateEntry(ctx context.Context, req interface{}) (interface{}, error) {
	s.logger.Info("UpdateEntry called (stub)")
	// TODO: Implement update logic
	// - Start UpdateEntryWorkflow
	// - Validate entry exists
	// - Update entry fields
	// - Notify Bacen
	return nil, status.Error(codes.Unimplemented, "UpdateEntry not yet implemented")
}

// DeleteEntry handles PIX key deletion (stub - will be implemented later)
func (s *EntryService) DeleteEntry(ctx context.Context, req interface{}) (interface{}, error) {
	s.logger.Info("DeleteEntry called (stub)")
	// TODO: Implement delete logic
	// - Start DeleteEntryWorkflow
	// - Validate entry exists and can be deleted
	// - Soft delete entry
	// - Notify Bacen
	return nil, status.Error(codes.Unimplemented, "DeleteEntry not yet implemented")
}

// GetEntry retrieves a single entry by ID or key (stub with partial implementation)
func (s *EntryService) GetEntry(ctx context.Context, req interface{}) (interface{}, error) {
	s.logger.Info("GetEntry called")

	reqMap, ok := req.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid request format")
	}

	// Try to get by entry_id first
	if entryID, ok := reqMap["entry_id"].(string); ok && entryID != "" {
		entry, err := s.entryRepo.GetByEntryID(ctx, entryID)
		if err != nil {
			s.logger.WithError(err).WithField("entry_id", entryID).Error("Failed to get entry by ID")
			return nil, status.Errorf(codes.NotFound, "entry not found: %s", entryID)
		}
		return entryToProtoMap(entry), nil
	}

	// Try to get by key
	if key, ok := reqMap["key"].(string); ok && key != "" {
		entry, err := s.entryRepo.GetByKey(ctx, key)
		if err != nil {
			s.logger.WithError(err).WithField("key", key).Error("Failed to get entry by key")
			return nil, status.Errorf(codes.NotFound, "entry not found for key: %s", key)
		}
		return entryToProtoMap(entry), nil
	}

	return nil, status.Error(codes.InvalidArgument, "either entry_id or key must be provided")
}

// ListEntries retrieves a paginated list of entries (stub with partial implementation)
func (s *EntryService) ListEntries(ctx context.Context, req interface{}) (interface{}, error) {
	s.logger.Info("ListEntries called")

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

	// Extract ISPB filter (required for security - users can only see their own entries)
	ispb, ok := reqMap["participant_ispb"].(string)
	if !ok || ispb == "" {
		return nil, status.Error(codes.InvalidArgument, "participant_ispb is required")
	}

	// Query entries
	entries, err := s.entryRepo.ListByParticipant(ctx, ispb, limit, offset)
	if err != nil {
		s.logger.WithError(err).WithField("ispb", ispb).Error("Failed to list entries")
		return nil, status.Error(codes.Internal, "failed to list entries")
	}

	// Convert to proto format
	entryMaps := make([]map[string]interface{}, len(entries))
	for i, entry := range entries {
		entryMaps[i] = entryToProtoMap(entry)
	}

	return map[string]interface{}{
		"entries":     entryMaps,
		"total_count": len(entryMaps),
		"limit":       limit,
		"offset":      offset,
	}, nil
}

// Helper functions

// Note: getStringOrEmpty is defined in claim_service.go to avoid duplication

// entryToProtoMap converts Entry entity to proto-like map (temporary until proto is generated)
func entryToProtoMap(e *entities.Entry) map[string]interface{} {
	result := map[string]interface{}{
		"id":       e.ID.String(),
		"entry_id": e.EntryID,
		"key":      e.Key,
		"key_type": string(e.KeyType),
		"participant": e.Participant,
		"account_type": string(e.AccountType),
		"owner_type": string(e.OwnerType),
		"status": string(e.Status),
		"created_at": e.CreatedAt.Format(time.RFC3339),
		"updated_at": e.UpdatedAt.Format(time.RFC3339),
	}

	// Add optional fields if present
	if e.AccountBranch != nil {
		result["account_branch"] = *e.AccountBranch
	}
	if e.AccountNumber != nil {
		result["account_number"] = *e.AccountNumber
	}
	if e.OwnerName != nil {
		result["owner_name"] = *e.OwnerName
	}
	if e.OwnerTaxID != nil {
		result["owner_tax_id"] = *e.OwnerTaxID
	}
	if e.RegisteredAt != nil {
		result["registered_at"] = e.RegisteredAt.Format(time.RFC3339)
	}
	if e.ActivatedAt != nil {
		result["activated_at"] = e.ActivatedAt.Format(time.RFC3339)
	}
	if e.BacenEntryID != nil {
		result["bacen_entry_id"] = *e.BacenEntryID
	}

	return result
}