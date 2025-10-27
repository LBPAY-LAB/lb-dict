package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	bridgev1 "github.com/lbpay-lab/dict-contracts/gen/proto/bridge/v1"
	commonv1 "github.com/lbpay-lab/dict-contracts/gen/proto/common/v1"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// BridgeClient interface for calling Bridge service
type BridgeClient interface {
	CreateEntry(ctx context.Context, req *bridgev1.CreateEntryRequest) (*bridgev1.CreateEntryResponse, error)
	GetEntry(ctx context.Context, req *bridgev1.GetEntryRequest) (*bridgev1.GetEntryResponse, error)
	UpdateEntry(ctx context.Context, req *bridgev1.UpdateEntryRequest) (*bridgev1.UpdateEntryResponse, error)
	DeleteEntry(ctx context.Context, req *bridgev1.DeleteEntryRequest) (*bridgev1.DeleteEntryResponse, error)
}

// EntryRepository interface for database operations
type EntryRepository interface {
	Create(ctx context.Context, entry *Entry) error
	GetByID(ctx context.Context, entryID string) (*Entry, error)
	GetByKey(ctx context.Context, keyType commonv1.KeyType, keyValue string) (*Entry, error)
	Update(ctx context.Context, entry *Entry) error
	SoftDelete(ctx context.Context, entryID string) error
}

// CacheRepository interface for Redis operations
type CacheRepository interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// EventPublisher interface for Pulsar operations
type EventPublisher interface {
	PublishEntryCreated(ctx context.Context, entryID string, key *commonv1.DictKey, account *commonv1.Account) error
	PublishEntryUpdated(ctx context.Context, entryID string, account *commonv1.Account) error
	PublishEntryDeleted(ctx context.Context, entryID string) error
}

// Entry represents a DICT entry in database
type Entry struct {
	EntryID      string
	ExternalID   string
	KeyType      commonv1.KeyType
	KeyValue     string
	AccountISPB  string
	AccountType  commonv1.AccountType
	AccountNum   string
	BranchCode   string
	HolderName   string
	HolderDoc    string
	Status       commonv1.EntryStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

// EntryUseCase handles business logic for entry operations
type EntryUseCase struct {
	bridgeClient BridgeClient
	entryRepo    EntryRepository
	cache        CacheRepository
	publisher    EventPublisher
	logger       *logrus.Logger
	tracer       trace.Tracer
}

// NewEntryUseCase creates a new EntryUseCase
func NewEntryUseCase(
	bridgeClient BridgeClient,
	entryRepo EntryRepository,
	cache CacheRepository,
	publisher EventPublisher,
	logger *logrus.Logger,
	tracer trace.Tracer,
) *EntryUseCase {
	return &EntryUseCase{
		bridgeClient: bridgeClient,
		entryRepo:    entryRepo,
		cache:        cache,
		publisher:    publisher,
		logger:       logger,
		tracer:       tracer,
	}
}

// CreateEntry handles the business logic for creating a new entry
func (uc *EntryUseCase) CreateEntry(ctx context.Context, req *bridgev1.CreateEntryRequest) (*bridgev1.CreateEntryResponse, error) {
	ctx, span := uc.tracer.Start(ctx, "EntryUseCase.CreateEntry")
	defer span.End()

	uc.logger.WithFields(logrus.Fields{
		"key_type":   req.Key.KeyType.String(),
		"key_value":  req.Key.KeyValue,
		"request_id": req.RequestId,
	}).Info("Creating DICT entry")

	// Step 1: Call Bridge to create entry in Bacen
	bridgeResp, err := uc.bridgeClient.CreateEntry(ctx, req)
	if err != nil {
		uc.logger.WithError(err).Error("Failed to create entry in Bridge")
		return nil, fmt.Errorf("bridge service failed: %w", err)
	}

	// Step 2: Persist entry in PostgreSQL
	entry := &Entry{
		EntryID:      bridgeResp.EntryId,
		ExternalID:   bridgeResp.ExternalId,
		KeyType:      req.Key.KeyType,
		KeyValue:     req.Key.KeyValue,
		AccountISPB:  req.Account.Ispb,
		AccountType:  req.Account.AccountType,
		AccountNum:   req.Account.AccountNumber,
		BranchCode:   req.Account.BranchCode,
		HolderName:   req.Account.AccountHolderName,
		HolderDoc:    req.Account.AccountHolderDocument,
		Status:       bridgeResp.Status,
		CreatedAt:    bridgeResp.CreatedAt.AsTime(),
		UpdatedAt:    bridgeResp.CreatedAt.AsTime(),
	}

	if err := uc.entryRepo.Create(ctx, entry); err != nil {
		uc.logger.WithError(err).Error("Failed to persist entry in database")
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Step 3: Publish event to Pulsar
	if err := uc.publisher.PublishEntryCreated(ctx, entry.EntryID, req.Key, req.Account); err != nil {
		uc.logger.WithError(err).Warn("Failed to publish entry created event")
		// Don't fail the request if event publishing fails
	}

	uc.logger.WithField("entry_id", entry.EntryID).Info("Entry created successfully")

	return bridgeResp, nil
}

// GetEntry retrieves an entry with caching strategy
func (uc *EntryUseCase) GetEntry(ctx context.Context, req *bridgev1.GetEntryRequest) (*bridgev1.GetEntryResponse, error) {
	ctx, span := uc.tracer.Start(ctx, "EntryUseCase.GetEntry")
	defer span.End()

	var cacheKey string
	var entry *Entry
	var err error

	// Determine cache key based on identifier type
	switch id := req.Identifier.(type) {
	case *bridgev1.GetEntryRequest_EntryId:
		cacheKey = fmt.Sprintf("entry:id:%s", id.EntryId)
		uc.logger.WithField("entry_id", id.EntryId).Debug("Getting entry by ID")
	case *bridgev1.GetEntryRequest_KeyQuery:
		cacheKey = fmt.Sprintf("entry:key:%s:%s", id.KeyQuery.KeyType.String(), id.KeyQuery.KeyValue)
		uc.logger.WithFields(logrus.Fields{
			"key_type":  id.KeyQuery.KeyType.String(),
			"key_value": id.KeyQuery.KeyValue,
		}).Debug("Getting entry by key")
	default:
		return nil, fmt.Errorf("invalid identifier type")
	}

	// Step 1: Check Redis cache
	if cached, err := uc.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		uc.logger.Debug("Cache hit")

		// Unmarshal cached entry
		if err := json.Unmarshal(cached, &entry); err == nil {
			return uc.entryToProtoResponse(entry), nil
		}
	}

	uc.logger.Debug("Cache miss, checking database")

	// Step 2: Query PostgreSQL
	switch id := req.Identifier.(type) {
	case *bridgev1.GetEntryRequest_EntryId:
		entry, err = uc.entryRepo.GetByID(ctx, id.EntryId)
	case *bridgev1.GetEntryRequest_KeyQuery:
		entry, err = uc.entryRepo.GetByKey(ctx, id.KeyQuery.KeyType, id.KeyQuery.KeyValue)
	}

	if err == nil && entry != nil {
		uc.logger.Debug("Found in database")

		// Cache the result (TTL: 5 minutes)
		if cached, err := json.Marshal(entry); err == nil {
			uc.cache.Set(ctx, cacheKey, cached, 5*time.Minute)
		}

		return uc.entryToProtoResponse(entry), nil
	}

	uc.logger.Debug("Not found in database, querying Bridge")

	// Step 3: Call Bridge service
	bridgeResp, err := uc.bridgeClient.GetEntry(ctx, req)
	if err != nil {
		uc.logger.WithError(err).Error("Failed to get entry from Bridge")
		return nil, fmt.Errorf("bridge service failed: %w", err)
	}

	if !bridgeResp.Found {
		uc.logger.Debug("Entry not found")
		return &bridgev1.GetEntryResponse{Found: false}, nil
	}

	// Step 4: Cache result from Bridge (TTL: 5 minutes)
	if bridgeResp.Found {
		entry = &Entry{
			EntryID:    bridgeResp.EntryId,
			ExternalID: bridgeResp.ExternalId,
			KeyType:    bridgeResp.Key.KeyType,
			KeyValue:   bridgeResp.Key.KeyValue,
			Status:     bridgeResp.Status,
			CreatedAt:  bridgeResp.CreatedAt.AsTime(),
			UpdatedAt:  bridgeResp.UpdatedAt.AsTime(),
		}

		if cached, err := json.Marshal(entry); err == nil {
			uc.cache.Set(ctx, cacheKey, cached, 5*time.Minute)
		}
	}

	return bridgeResp, nil
}

// UpdateEntry handles the business logic for updating an entry
func (uc *EntryUseCase) UpdateEntry(ctx context.Context, req *bridgev1.UpdateEntryRequest) (*bridgev1.UpdateEntryResponse, error) {
	ctx, span := uc.tracer.Start(ctx, "EntryUseCase.UpdateEntry")
	defer span.End()

	uc.logger.WithFields(logrus.Fields{
		"entry_id":   req.EntryId,
		"request_id": req.RequestId,
	}).Info("Updating DICT entry")

	// Step 1: Validate entry exists
	existing, err := uc.entryRepo.GetByID(ctx, req.EntryId)
	if err != nil || existing == nil {
		uc.logger.WithError(err).Error("Entry not found")
		return nil, fmt.Errorf("entry not found: %s", req.EntryId)
	}

	// Step 2: Call Bridge to update entry in Bacen
	bridgeResp, err := uc.bridgeClient.UpdateEntry(ctx, req)
	if err != nil {
		uc.logger.WithError(err).Error("Failed to update entry in Bridge")
		return nil, fmt.Errorf("bridge service failed: %w", err)
	}

	// Step 3: Update in PostgreSQL
	existing.AccountISPB = req.NewAccount.Ispb
	existing.AccountType = req.NewAccount.AccountType
	existing.AccountNum = req.NewAccount.AccountNumber
	existing.BranchCode = req.NewAccount.BranchCode
	existing.HolderName = req.NewAccount.AccountHolderName
	existing.HolderDoc = req.NewAccount.AccountHolderDocument
	existing.UpdatedAt = bridgeResp.UpdatedAt.AsTime()

	if err := uc.entryRepo.Update(ctx, existing); err != nil {
		uc.logger.WithError(err).Error("Failed to update entry in database")
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Step 4: Invalidate cache
	cacheKey := fmt.Sprintf("entry:id:%s", req.EntryId)
	uc.cache.Delete(ctx, cacheKey)

	// Step 5: Publish event to Pulsar
	if err := uc.publisher.PublishEntryUpdated(ctx, req.EntryId, req.NewAccount); err != nil {
		uc.logger.WithError(err).Warn("Failed to publish entry updated event")
	}

	uc.logger.WithField("entry_id", req.EntryId).Info("Entry updated successfully")

	return bridgeResp, nil
}

// DeleteEntry handles the business logic for deleting an entry
func (uc *EntryUseCase) DeleteEntry(ctx context.Context, req *bridgev1.DeleteEntryRequest) (*bridgev1.DeleteEntryResponse, error) {
	ctx, span := uc.tracer.Start(ctx, "EntryUseCase.DeleteEntry")
	defer span.End()

	uc.logger.WithFields(logrus.Fields{
		"entry_id":   req.EntryId,
		"request_id": req.RequestId,
	}).Info("Deleting DICT entry")

	// Step 1: Validate entry exists
	existing, err := uc.entryRepo.GetByID(ctx, req.EntryId)
	if err != nil || existing == nil {
		uc.logger.WithError(err).Error("Entry not found")
		return nil, fmt.Errorf("entry not found: %s", req.EntryId)
	}

	// Step 2: Call Bridge to delete entry in Bacen
	bridgeResp, err := uc.bridgeClient.DeleteEntry(ctx, req)
	if err != nil {
		uc.logger.WithError(err).Error("Failed to delete entry in Bridge")
		return nil, fmt.Errorf("bridge service failed: %w", err)
	}

	// Step 3: Soft delete in PostgreSQL (mark as deleted)
	if err := uc.entryRepo.SoftDelete(ctx, req.EntryId); err != nil {
		uc.logger.WithError(err).Error("Failed to delete entry in database")
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Step 4: Invalidate cache
	cacheKey := fmt.Sprintf("entry:id:%s", req.EntryId)
	uc.cache.Delete(ctx, cacheKey)

	// Step 5: Publish event to Pulsar
	if err := uc.publisher.PublishEntryDeleted(ctx, req.EntryId); err != nil {
		uc.logger.WithError(err).Warn("Failed to publish entry deleted event")
	}

	uc.logger.WithField("entry_id", req.EntryId).Info("Entry deleted successfully")

	return bridgeResp, nil
}

// entryToProtoResponse converts Entry to GetEntryResponse
func (uc *EntryUseCase) entryToProtoResponse(entry *Entry) *bridgev1.GetEntryResponse {
	return &bridgev1.GetEntryResponse{
		EntryId:    entry.EntryID,
		ExternalId: entry.ExternalID,
		Key: &commonv1.DictKey{
			KeyType:  entry.KeyType,
			KeyValue: entry.KeyValue,
		},
		Account: &commonv1.Account{
			Ispb:                  entry.AccountISPB,
			AccountType:           entry.AccountType,
			AccountNumber:         entry.AccountNum,
			BranchCode:            entry.BranchCode,
			AccountHolderName:     entry.HolderName,
			AccountHolderDocument: entry.HolderDoc,
		},
		Status:    entry.Status,
		CreatedAt: timestamppb.New(entry.CreatedAt),
		UpdatedAt: timestamppb.New(entry.UpdatedAt),
		Found:     true,
	}
}
