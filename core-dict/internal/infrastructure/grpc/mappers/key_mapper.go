package mappers

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	corev1 "github.com/lbpay-lab/dict-contracts/gen/proto/core/v1"
	commonv1 "github.com/lbpay-lab/dict-contracts/gen/proto/common/v1"

	"github.com/lbpay-lab/core-dict/internal/application/commands"
	"github.com/lbpay-lab/core-dict/internal/application/queries"
	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/domain/valueobjects"
)

// ============================================================================
// Proto KeyType → Domain KeyType
// ============================================================================

func MapProtoKeyTypeToDomain(kt commonv1.KeyType) valueobjects.KeyType {
	switch kt {
	case commonv1.KeyType_KEY_TYPE_CPF:
		return valueobjects.KeyTypeCPF
	case commonv1.KeyType_KEY_TYPE_CNPJ:
		return valueobjects.KeyTypeCNPJ
	case commonv1.KeyType_KEY_TYPE_EMAIL:
		return valueobjects.KeyTypeEmail
	case commonv1.KeyType_KEY_TYPE_PHONE:
		return valueobjects.KeyTypePhone
	case commonv1.KeyType_KEY_TYPE_EVP:
		return valueobjects.KeyTypeEVP
	default:
		return ""
	}
}

func MapDomainKeyTypeToProto(kt valueobjects.KeyType) commonv1.KeyType {
	switch kt {
	case valueobjects.KeyTypeCPF:
		return commonv1.KeyType_KEY_TYPE_CPF
	case valueobjects.KeyTypeCNPJ:
		return commonv1.KeyType_KEY_TYPE_CNPJ
	case valueobjects.KeyTypeEmail:
		return commonv1.KeyType_KEY_TYPE_EMAIL
	case valueobjects.KeyTypePhone:
		return commonv1.KeyType_KEY_TYPE_PHONE
	case valueobjects.KeyTypeEVP:
		return commonv1.KeyType_KEY_TYPE_EVP
	default:
		return commonv1.KeyType_KEY_TYPE_UNSPECIFIED
	}
}

// ============================================================================
// Proto EntryStatus → Domain KeyStatus
// ============================================================================

func MapProtoStatusToDomain(st commonv1.EntryStatus) valueobjects.KeyStatus {
	switch st {
	case commonv1.EntryStatus_ENTRY_STATUS_ACTIVE:
		return valueobjects.KeyStatusActive
	case commonv1.EntryStatus_ENTRY_STATUS_PORTABILITY_PENDING:
		return valueobjects.KeyStatusPending // Map portability to pending
	case commonv1.EntryStatus_ENTRY_STATUS_PORTABILITY_CONFIRMED:
		return valueobjects.KeyStatusActive // Map confirmed portability to active
	case commonv1.EntryStatus_ENTRY_STATUS_CLAIM_PENDING:
		return valueobjects.KeyStatusClaimPending
	case commonv1.EntryStatus_ENTRY_STATUS_DELETED:
		return valueobjects.KeyStatusDeleted
	default:
		return valueobjects.KeyStatusActive
	}
}

func MapDomainStatusToProto(st valueobjects.KeyStatus) commonv1.EntryStatus {
	switch st {
	case valueobjects.KeyStatusPending:
		return commonv1.EntryStatus_ENTRY_STATUS_PORTABILITY_PENDING
	case valueobjects.KeyStatusActive:
		return commonv1.EntryStatus_ENTRY_STATUS_ACTIVE
	case valueobjects.KeyStatusBlocked:
		return commonv1.EntryStatus_ENTRY_STATUS_DELETED // Map blocked to deleted
	case valueobjects.KeyStatusDeleted:
		return commonv1.EntryStatus_ENTRY_STATUS_DELETED
	case valueobjects.KeyStatusClaimPending:
		return commonv1.EntryStatus_ENTRY_STATUS_CLAIM_PENDING
	case valueobjects.KeyStatusFailed:
		return commonv1.EntryStatus_ENTRY_STATUS_DELETED // Map failed to deleted
	default:
		return commonv1.EntryStatus_ENTRY_STATUS_UNSPECIFIED
	}
}

// MapStringStatusToProto converts string status (from command result) to proto EntryStatus
func MapStringStatusToProto(status string) commonv1.EntryStatus {
	switch status {
	case "pending", "PENDING":
		return commonv1.EntryStatus_ENTRY_STATUS_PORTABILITY_PENDING
	case "active", "ACTIVE":
		return commonv1.EntryStatus_ENTRY_STATUS_ACTIVE
	case "blocked", "BLOCKED":
		return commonv1.EntryStatus_ENTRY_STATUS_DELETED // Map blocked to deleted
	case "deleted", "DELETED":
		return commonv1.EntryStatus_ENTRY_STATUS_DELETED
	case "claim_pending", "CLAIM_PENDING":
		return commonv1.EntryStatus_ENTRY_STATUS_CLAIM_PENDING
	case "failed", "FAILED":
		return commonv1.EntryStatus_ENTRY_STATUS_DELETED // Map failed to deleted
	default:
		return commonv1.EntryStatus_ENTRY_STATUS_UNSPECIFIED
	}
}

// ============================================================================
// Proto AccountType → Domain AccountType (String-based until domain implements)
// ============================================================================

// TODO: Remove these when domain/valueobjects/account_type.go is implemented
// For now, Account.AccountType is just a string in commands

func MapProtoAccountTypeToDomain(at commonv1.AccountType) string {
	switch at {
	case commonv1.AccountType_ACCOUNT_TYPE_CHECKING:
		return "CHECKING"
	case commonv1.AccountType_ACCOUNT_TYPE_SAVINGS:
		return "SAVINGS"
	case commonv1.AccountType_ACCOUNT_TYPE_PAYMENT:
		return "PAYMENT"
	default:
		return "CHECKING"
	}
}

func MapDomainAccountTypeToProto(at string) commonv1.AccountType {
	switch at {
	case "CHECKING":
		return commonv1.AccountType_ACCOUNT_TYPE_CHECKING
	case "SAVINGS":
		return commonv1.AccountType_ACCOUNT_TYPE_SAVINGS
	case "PAYMENT":
		return commonv1.AccountType_ACCOUNT_TYPE_PAYMENT
	default:
		return commonv1.AccountType_ACCOUNT_TYPE_UNSPECIFIED
	}
}

// ============================================================================
// Proto Account → Domain Account
// ============================================================================

func MapProtoAccountToDomain(acc *commonv1.Account) *entities.Account {
	if acc == nil {
		return nil
	}

	// TODO: Account entity uses AccountType (string constants), not the proto's AccountType enum
	return &entities.Account{
		ISPB:          acc.Ispb,
		Branch:        acc.BranchCode,
		AccountNumber: acc.AccountNumber,
		// AccountType: needs proper mapping to entities.AccountType constants
	}
}

func MapDomainAccountToProto(acc *entities.Account) *commonv1.Account {
	if acc == nil {
		return nil
	}

	// Map entities.AccountType (string constants) to proto AccountType enum
	protoAccountType := mapEntityAccountTypeToProto(acc.AccountType)

	return &commonv1.Account{
		Ispb:          acc.ISPB,
		BranchCode:    acc.Branch,
		AccountNumber: acc.AccountNumber,
		AccountType:   protoAccountType,
		// TODO: Add other proto fields: account_check_digit, account_holder_name, account_holder_document, document_type
	}
}

// Helper to map entities.AccountType to proto AccountType
func mapEntityAccountTypeToProto(at entities.AccountType) commonv1.AccountType {
	switch at {
	case entities.AccountTypeCACC:
		return commonv1.AccountType_ACCOUNT_TYPE_CHECKING
	case entities.AccountTypeSVGS:
		return commonv1.AccountType_ACCOUNT_TYPE_SAVINGS
	case entities.AccountTypeSLRY:
		return commonv1.AccountType_ACCOUNT_TYPE_SALARY
	case entities.AccountTypeTRAN:
		return commonv1.AccountType_ACCOUNT_TYPE_PAYMENT
	default:
		return commonv1.AccountType_ACCOUNT_TYPE_UNSPECIFIED
	}
}

// ============================================================================
// Domain Entry → Proto KeySummary
// ============================================================================

func MapDomainEntryToProtoKeySummary(entry *entities.Entry) *corev1.KeySummary {
	if entry == nil {
		return nil
	}

	// Map entities.KeyType (string) to proto KeyType (enum)
	protoKeyType := mapEntityKeyTypeToProto(entry.KeyType)
	// Map entities.KeyStatus (string) to proto EntryStatus (enum)
	protoStatus := mapEntityStatusToProto(entry.Status)

	return &corev1.KeySummary{
		KeyId: entry.ID.String(),
		Key: &commonv1.DictKey{
			KeyType:  protoKeyType,
			KeyValue: entry.KeyValue,
		},
		Status:    protoStatus,
		AccountId: entry.AccountID.String(),
		CreatedAt: timestamppb.New(entry.CreatedAt),
		UpdatedAt: timestamppb.New(entry.UpdatedAt),
	}
}

// ============================================================================
// Domain Entry → Proto GetKeyResponse
// ============================================================================

func MapDomainEntryToProtoGetKeyResponse(entry *entities.Entry, account *entities.Account) *corev1.GetKeyResponse {
	if entry == nil {
		return nil
	}

	// Map entities types to proto enums
	protoKeyType := mapEntityKeyTypeToProto(entry.KeyType)
	protoStatus := mapEntityStatusToProto(entry.Status)

	resp := &corev1.GetKeyResponse{
		KeyId: entry.ID.String(),
		Key: &commonv1.DictKey{
			KeyType:  protoKeyType,
			KeyValue: entry.KeyValue,
		},
		Status:    protoStatus,
		CreatedAt: timestamppb.New(entry.CreatedAt),
		UpdatedAt: timestamppb.New(entry.UpdatedAt),
	}

	if account != nil {
		resp.Account = MapDomainAccountToProto(account)
	}

	// TODO: Add portability history if needed
	resp.PortabilityHistory = []*corev1.PortabilityHistory{}

	return resp
}

// ============================================================================
// Proto CreateKeyRequest → Application CreateEntryCommand
// ============================================================================

func MapProtoCreateKeyRequestToCommand(req *corev1.CreateKeyRequest, userID string) (commands.CreateEntryCommand, error) {
	// Proto only has: key_type, key_value, account_id
	// All account details (ISPB, branch, etc.) will be fetched by handler from account_id
	accountID, err := parseUUID(req.GetAccountId())
	if err != nil {
		return commands.CreateEntryCommand{}, fmt.Errorf("invalid account_id: %w", err)
	}

	requestedBy, err := parseUUID(userID)
	if err != nil {
		return commands.CreateEntryCommand{}, fmt.Errorf("invalid user_id: %w", err)
	}

	// Map KeyType from proto to commands.KeyType (string-based)
	keyType := mapProtoKeyTypeToCommandKeyType(req.GetKeyType())

	return commands.CreateEntryCommand{
		KeyType:     keyType,
		KeyValue:    req.GetKeyValue(),
		AccountID:   accountID,
		RequestedBy: requestedBy,
		// AccountISPB, AccountBranch, AccountNumber, AccountType, OwnerType, OwnerTaxID, OwnerName
		// will all be populated by handler after fetching account details
		// OTP: handled separately by auth middleware
	}, nil
}

// ============================================================================
// Proto ListKeysRequest → Application ListEntriesQuery
// ============================================================================

func MapProtoListKeysRequestToQuery(req *corev1.ListKeysRequest, accountID string) (queries.ListEntriesQuery, error) {
	// Proto has: page_size, page_token, optional key_type, optional status
	// Account ID comes from auth context (which account the user is querying)
	accID, err := parseUUID(accountID)
	if err != nil {
		return queries.ListEntriesQuery{}, fmt.Errorf("invalid account_id: %w", err)
	}

	// Page is 1-indexed in the query (extract from page_token)
	page := 1 // TODO: Extract from page_token

	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = 100 // Default
	}
	if pageSize > 1000 {
		pageSize = 1000 // Max
	}

	return queries.ListEntriesQuery{
		AccountID: accID,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

// ============================================================================
// Proto DeleteKeyRequest → Application DeleteEntryCommand
// ============================================================================

func MapProtoDeleteKeyRequestToCommand(req *corev1.DeleteKeyRequest, userID string) (commands.DeleteEntryCommand, error) {
	// Proto only has: key_id
	entryID, err := parseUUID(req.GetKeyId())
	if err != nil {
		return commands.DeleteEntryCommand{}, fmt.Errorf("invalid key_id: %w", err)
	}

	requestedBy, err := parseUUID(userID)
	if err != nil {
		return commands.DeleteEntryCommand{}, fmt.Errorf("invalid user_id: %w", err)
	}

	return commands.DeleteEntryCommand{
		EntryID:     entryID,
		RequestedBy: requestedBy,
		// TwoFactorCode: from auth context/header
		// Reason: "USER_REQUESTED" (default by handler)
	}, nil
}

// ============================================================================
// Proto LookupKeyRequest → Application GetEntryQuery (by key value)
// ============================================================================

func MapProtoLookupKeyRequestToQuery(req *corev1.LookupKeyRequest) queries.GetEntryQuery {
	return queries.GetEntryQuery{
		KeyValue: req.GetKey().GetKeyValue(),
	}
}

// ============================================================================
// Helper: Timestamp conversion
// ============================================================================

func TimeToTimestamppb(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

// ============================================================================
// Helper: UUID parsing
// ============================================================================

func parseUUID(s string) (uuid.UUID, error) {
	if s == "" {
		return uuid.Nil, fmt.Errorf("empty UUID string")
	}
	return uuid.Parse(s)
}

// ============================================================================
// Helper: KeyType conversion (proto enum → commands.KeyType string)
// ============================================================================

func mapProtoKeyTypeToCommandKeyType(kt commonv1.KeyType) commands.KeyType {
	switch kt {
	case commonv1.KeyType_KEY_TYPE_CPF:
		return commands.KeyTypeCPF
	case commonv1.KeyType_KEY_TYPE_CNPJ:
		return commands.KeyTypeCNPJ
	case commonv1.KeyType_KEY_TYPE_EMAIL:
		return commands.KeyTypeEmail
	case commonv1.KeyType_KEY_TYPE_PHONE:
		return commands.KeyTypePhone
	case commonv1.KeyType_KEY_TYPE_EVP:
		return commands.KeyTypeEVP
	default:
		return ""
	}
}

// ============================================================================
// Helper: AccountType conversion (proto enum → string)
// ============================================================================

func mapProtoAccountTypeToString(at commonv1.AccountType) string {
	switch at {
	case commonv1.AccountType_ACCOUNT_TYPE_CHECKING:
		return "CHECKING"
	case commonv1.AccountType_ACCOUNT_TYPE_SAVINGS:
		return "SAVINGS"
	case commonv1.AccountType_ACCOUNT_TYPE_PAYMENT:
		return "PAYMENT"
	default:
		return "CHECKING" // Default
	}
}

// ============================================================================
// Helper: entities.KeyType (string) → proto KeyType (enum)
// ============================================================================

func mapEntityKeyTypeToProto(kt entities.KeyType) commonv1.KeyType {
	switch kt {
	case entities.KeyTypeCPF:
		return commonv1.KeyType_KEY_TYPE_CPF
	case entities.KeyTypeCNPJ:
		return commonv1.KeyType_KEY_TYPE_CNPJ
	case entities.KeyTypeEmail:
		return commonv1.KeyType_KEY_TYPE_EMAIL
	case entities.KeyTypePhone:
		return commonv1.KeyType_KEY_TYPE_PHONE
	case entities.KeyTypeEVP:
		return commonv1.KeyType_KEY_TYPE_EVP
	default:
		return commonv1.KeyType_KEY_TYPE_UNSPECIFIED
	}
}

// ============================================================================
// Helper: entities.KeyStatus (string) → proto EntryStatus (enum)
// ============================================================================

func mapEntityStatusToProto(st entities.KeyStatus) commonv1.EntryStatus {
	switch st {
	case entities.KeyStatusActive:
		return commonv1.EntryStatus_ENTRY_STATUS_ACTIVE
	case entities.KeyStatusPending:
		return commonv1.EntryStatus_ENTRY_STATUS_PORTABILITY_PENDING
	case entities.KeyStatusBlocked:
		return commonv1.EntryStatus_ENTRY_STATUS_DELETED // Map blocked to deleted
	case entities.KeyStatusDeleted:
		return commonv1.EntryStatus_ENTRY_STATUS_DELETED
	default:
		return commonv1.EntryStatus_ENTRY_STATUS_UNSPECIFIED
	}
}
