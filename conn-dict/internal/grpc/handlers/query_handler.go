package handlers

import (
	"context"
	"fmt"

	connectv1 "github.com/lbpay-lab/dict-contracts/gen/proto/conn_dict/v1"
	commonv1 "github.com/lbpay-lab/dict-contracts/gen/proto/common/v1"
	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/cache"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// QueryHandler handles read-only Entry operations for ConnectService
type QueryHandler struct {
	entryRepo *repositories.EntryRepository
	cache     *cache.RedisClient
	logger    *logrus.Logger
	tracer    trace.Tracer
}

// NewQueryHandler creates a new QueryHandler
func NewQueryHandler(
	entryRepo *repositories.EntryRepository,
	cache *cache.RedisClient,
	logger *logrus.Logger,
	tracer trace.Tracer,
) *QueryHandler {
	return &QueryHandler{
		entryRepo: entryRepo,
		cache:     cache,
		logger:    logger,
		tracer:    tracer,
	}
}

// GetEntry retrieves an entry by ID
func (h *QueryHandler) GetEntry(ctx context.Context, req *connectv1.GetEntryRequest) (*connectv1.GetEntryResponse, error) {
	ctx, span := h.tracer.Start(ctx, "QueryHandler.GetEntry")
	defer span.End()

	h.logger.WithFields(logrus.Fields{
		"entry_id": req.EntryId,
	}).Info("GetEntry called")

	// Validate request
	if req.EntryId == "" {
		return nil, status.Error(codes.InvalidArgument, "entry_id is required")
	}

	// Try cache first (skip cache for now - requires serialization logic)
	// cacheKey := fmt.Sprintf("entry:%s", req.EntryId)
	// var entry entities.Entry
	// err := h.cache.Get(ctx, cacheKey, &entry)
	// if err == nil {
	//     h.logger.Debug("Cache hit for entry")
	//     return &connectv1.GetEntryResponse{Entry: h.convertEntryToProto(&entry)}, nil
	// }

	// Query database by EntryID (external ID)
	entry, err := h.entryRepo.GetByEntryID(ctx, req.EntryId)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get entry from database")
		return nil, status.Error(codes.NotFound, fmt.Sprintf("entry not found: %v", err))
	}

	// Cache the result (5 minutes TTL)
	// TODO: Enable caching once serialization is working
	// h.cache.Set(ctx, cacheKey, entry, 5*time.Minute)

	// Convert to proto
	protoEntry := h.convertEntryToProto(entry)

	return &connectv1.GetEntryResponse{
		Entry: protoEntry,
	}, nil
}

// GetEntryByKey retrieves an entry by DICT key (CPF, email, phone, etc)
func (h *QueryHandler) GetEntryByKey(ctx context.Context, req *connectv1.GetEntryByKeyRequest) (*connectv1.GetEntryByKeyResponse, error) {
	ctx, span := h.tracer.Start(ctx, "QueryHandler.GetEntryByKey")
	defer span.End()

	h.logger.WithFields(logrus.Fields{
		"key_type":  req.Key.GetKeyType(),
		"key_value": maskKey(req.Key.GetKeyValue()),
	}).Info("GetEntryByKey called")

	// Validate request
	if req.Key == nil {
		return nil, status.Error(codes.InvalidArgument, "key is required")
	}
	if req.Key.GetKeyValue() == "" {
		return nil, status.Error(codes.InvalidArgument, "key_value is required")
	}

	// Try cache first (skip cache for now - requires serialization logic)
	// cacheKey := fmt.Sprintf("entry:key:%s:%s", req.Key.GetKeyType(), req.Key.GetKeyValue())
	// var entry entities.Entry
	// err := h.cache.Get(ctx, cacheKey, &entry)
	// if err == nil {
	//     h.logger.Debug("Cache hit for entry by key")
	//     return &connectv1.GetEntryByKeyResponse{Entry: h.convertEntryToProto(&entry)}, nil
	// }

	// Query database by key
	entry, err := h.entryRepo.GetByKey(ctx, req.Key.GetKeyValue())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get entry by key from database")
		return nil, status.Error(codes.NotFound, fmt.Sprintf("entry not found for key: %v", err))
	}

	// Cache the result (5 minutes TTL)
	// TODO: Enable caching once serialization is working
	// h.cache.Set(ctx, cacheKey, entry, 5*time.Minute)

	// Convert to proto
	protoEntry := h.convertEntryToProto(entry)

	return &connectv1.GetEntryByKeyResponse{
		Entry: protoEntry,
	}, nil
}

// ListEntries lists entries for a participant with pagination
func (h *QueryHandler) ListEntries(ctx context.Context, req *connectv1.ListEntriesRequest) (*connectv1.ListEntriesResponse, error) {
	ctx, span := h.tracer.Start(ctx, "QueryHandler.ListEntries")
	defer span.End()

	h.logger.WithFields(logrus.Fields{
		"participant_ispb": req.ParticipantIspb,
		"limit":            req.Limit,
		"offset":           req.Offset,
	}).Info("ListEntries called")

	// Validate request
	if req.ParticipantIspb == "" {
		return nil, status.Error(codes.InvalidArgument, "participant_ispb is required")
	}

	// Set defaults
	limit := req.Limit
	if limit == 0 || limit > 1000 {
		limit = 100 // Default and max
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	// Query database
	entries, err := h.entryRepo.ListByParticipant(ctx, req.ParticipantIspb, int(limit), int(offset))
	if err != nil {
		h.logger.WithError(err).Error("Failed to list entries from database")
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list entries: %v", err))
	}

	// Convert to proto
	protoEntries := make([]*connectv1.Entry, len(entries))
	for i, entry := range entries {
		protoEntries[i] = h.convertEntryToProto(entry)
	}

	// Get total count for pagination
	totalCount, err := h.entryRepo.CountByParticipant(ctx, req.ParticipantIspb)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to get total count, using entries length")
		totalCount = int64(len(entries))
	}

	return &connectv1.ListEntriesResponse{
		Entries:    protoEntries,
		TotalCount: int32(totalCount),
		Limit:      int32(limit),
		Offset:     int32(offset),
	}, nil
}

// convertEntryToProto converts domain Entry to proto Entry
func (h *QueryHandler) convertEntryToProto(entry *entities.Entry) *connectv1.Entry {
	if entry == nil {
		return nil
	}

	return &connectv1.Entry{
		EntryId:         entry.EntryID,
		ParticipantIspb: entry.Participant,
		KeyType:         convertKeyTypeToProto(entry.KeyType),
		KeyValue:        entry.Key,
		Account: &commonv1.Account{
			Ispb:          entry.Participant,
			AccountType:   convertAccountTypeToProto(entry.AccountType),
			AccountNumber: stringPtrOrEmpty(entry.AccountNumber),
		},
		Status:    convertStatusToProto(entry.Status),
		CreatedAt: timestamppb.New(entry.CreatedAt),
		UpdatedAt: timestamppb.New(entry.UpdatedAt),
	}
}

// Helper functions for enum conversion
func convertKeyTypeToProto(keyType entities.KeyType) commonv1.KeyType {
	switch keyType {
	case entities.KeyTypeCPF:
		return commonv1.KeyType_KEY_TYPE_CPF
	case entities.KeyTypeCNPJ:
		return commonv1.KeyType_KEY_TYPE_CNPJ
	case entities.KeyTypeEMAIL:
		return commonv1.KeyType_KEY_TYPE_EMAIL
	case entities.KeyTypePHONE:
		return commonv1.KeyType_KEY_TYPE_PHONE
	case entities.KeyTypeEVP:
		return commonv1.KeyType_KEY_TYPE_EVP
	default:
		return commonv1.KeyType_KEY_TYPE_UNSPECIFIED
	}
}

func convertAccountTypeToProto(accountType entities.AccountType) commonv1.AccountType {
	switch accountType {
	case entities.AccountTypeCACC:
		return commonv1.AccountType_ACCOUNT_TYPE_CHECKING
	case entities.AccountTypeSLRY:
		return commonv1.AccountType_ACCOUNT_TYPE_SALARY
	case entities.AccountTypeSVGS:
		return commonv1.AccountType_ACCOUNT_TYPE_SAVINGS
	case entities.AccountTypeTRAN:
		return commonv1.AccountType_ACCOUNT_TYPE_PAYMENT
	default:
		return commonv1.AccountType_ACCOUNT_TYPE_UNSPECIFIED
	}
}

func convertStatusToProto(status entities.EntryStatus) commonv1.EntryStatus {
	switch status {
	case entities.EntryStatusActive:
		return commonv1.EntryStatus_ENTRY_STATUS_ACTIVE
	case entities.EntryStatusPortabilityPending:
		return commonv1.EntryStatus_ENTRY_STATUS_PORTABILITY_PENDING
	case entities.EntryStatusOwnershipChangePending:
		return commonv1.EntryStatus_ENTRY_STATUS_CLAIM_PENDING
	case entities.EntryStatusInactive:
		return commonv1.EntryStatus_ENTRY_STATUS_DELETED
	case entities.EntryStatusBlocked:
		return commonv1.EntryStatus_ENTRY_STATUS_DELETED
	default:
		return commonv1.EntryStatus_ENTRY_STATUS_UNSPECIFIED
	}
}

// maskKey masks sensitive parts of the key for logging
func maskKey(key string) string {
	if len(key) < 4 {
		return "***"
	}
	return key[:2] + "****" + key[len(key)-2:]
}

// stringPtrOrEmpty returns the string value or empty string if nil
func stringPtrOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
