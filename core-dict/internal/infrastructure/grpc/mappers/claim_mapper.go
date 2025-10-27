package mappers

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	corev1 "github.com/lbpay-lab/dict-contracts/gen/proto/core/v1"
	commonv1 "github.com/lbpay-lab/dict-contracts/gen/proto/common/v1"

	"github.com/lbpay-lab/core-dict/internal/application/commands"
	"github.com/lbpay-lab/core-dict/internal/application/queries"
	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/domain/valueobjects"
)

// NOTE: ClaimType mappers removed - ClaimType does not exist in common.proto
// It only exists in conn_dict events. Core DICT does not use ClaimType.

// ============================================================================
// Proto ClaimStatus → Domain ClaimStatus
// ============================================================================

func MapProtoClaimStatusToDomain(cs commonv1.ClaimStatus) valueobjects.ClaimStatus {
	switch cs {
	case commonv1.ClaimStatus_CLAIM_STATUS_OPEN:
		return valueobjects.ClaimStatusOpen
	case commonv1.ClaimStatus_CLAIM_STATUS_WAITING_RESOLUTION:
		return valueobjects.ClaimStatusWaitingResolution
	case commonv1.ClaimStatus_CLAIM_STATUS_CONFIRMED:
		return valueobjects.ClaimStatusConfirmed
	case commonv1.ClaimStatus_CLAIM_STATUS_CANCELLED:
		return valueobjects.ClaimStatusCancelled
	case commonv1.ClaimStatus_CLAIM_STATUS_COMPLETED:
		return valueobjects.ClaimStatusCompleted
	default:
		return valueobjects.ClaimStatusOpen
	}
}

func MapDomainClaimStatusToProto(cs valueobjects.ClaimStatus) commonv1.ClaimStatus {
	switch cs {
	case valueobjects.ClaimStatusOpen:
		return commonv1.ClaimStatus_CLAIM_STATUS_OPEN
	case valueobjects.ClaimStatusWaitingResolution:
		return commonv1.ClaimStatus_CLAIM_STATUS_WAITING_RESOLUTION
	case valueobjects.ClaimStatusConfirmed:
		return commonv1.ClaimStatus_CLAIM_STATUS_CONFIRMED
	case valueobjects.ClaimStatusCancelled:
		return commonv1.ClaimStatus_CLAIM_STATUS_CANCELLED
	case valueobjects.ClaimStatusCompleted:
		return commonv1.ClaimStatus_CLAIM_STATUS_COMPLETED
	default:
		return commonv1.ClaimStatus_CLAIM_STATUS_UNSPECIFIED
	}
}

// ============================================================================
// Proto StartClaimRequest → Application CreateClaimCommand
// ============================================================================

func MapProtoStartClaimRequestToCommand(req *corev1.StartClaimRequest, userID string) (commands.CreateClaimCommand, error) {
	accountID, err := parseUUID(req.GetAccountId())
	if err != nil {
		return commands.CreateClaimCommand{}, fmt.Errorf("invalid account_id: %w", err)
	}

	requestedBy, err := parseUUID(userID)
	if err != nil {
		return commands.CreateClaimCommand{}, fmt.Errorf("invalid user_id: %w", err)
	}

	// Proto only has: key (DictKey), account_id
	// Handler will determine: ClaimType, ISPBs, OwnerTaxID from context
	return commands.CreateClaimCommand{
		KeyValue:    req.GetKey().GetKeyValue(),
		// ClaimType will be determined by handler (ownership vs portability)
		// ClaimerISPB, ClaimedISPB, OwnerTaxID determined by handler from account/entry lookup
		AccountID:   accountID,
		RequestedBy: requestedBy,
		// BacenClaimID will be set by the handler after RSFN call
	}, nil
}

// ============================================================================
// Proto RespondToClaimRequest → Application Commands
// ============================================================================

func MapProtoRespondToClaimRequestToConfirmCommand(req *corev1.RespondToClaimRequest, userID string) (commands.ConfirmClaimCommand, error) {
	claimID, err := parseUUID(req.GetClaimId())
	if err != nil {
		return commands.ConfirmClaimCommand{}, fmt.Errorf("invalid claim_id: %w", err)
	}

	requestedBy, err := parseUUID(userID)
	if err != nil {
		return commands.ConfirmClaimCommand{}, fmt.Errorf("invalid user_id: %w", err)
	}

	// Proto has: claim_id, response (ACCEPT/REJECT), optional reason
	// TwoFactorCode and ConfirmedBy will come from auth context
	return commands.ConfirmClaimCommand{
		ClaimID:     claimID,
		RequestedBy: requestedBy,
		// TwoFactorCode: from auth context/header
		// ConfirmedBy: from auth context (user's name)
	}, nil
}

func MapProtoRespondToClaimRequestToCancelCommand(req *corev1.RespondToClaimRequest, userID string) (commands.CancelClaimCommand, error) {
	claimID, err := parseUUID(req.GetClaimId())
	if err != nil {
		return commands.CancelClaimCommand{}, fmt.Errorf("invalid claim_id: %w", err)
	}

	requestedBy, err := parseUUID(userID)
	if err != nil {
		return commands.CancelClaimCommand{}, fmt.Errorf("invalid user_id: %w", err)
	}

	// Proto has: claim_id, response (ACCEPT/REJECT), optional reason
	reason := ""
	if req.Reason != nil {
		reason = *req.Reason
	}

	return commands.CancelClaimCommand{
		ClaimID:     claimID,
		RequestedBy: requestedBy,
		// TwoFactorCode: from auth context/header
		Reason: reason,
	}, nil
}

// ============================================================================
// Proto CancelClaimRequest → Application CancelClaimCommand
// ============================================================================

func MapProtoCancelClaimRequestToCommand(req *corev1.CancelClaimRequest, userID string) (commands.CancelClaimCommand, error) {
	claimID, err := parseUUID(req.GetClaimId())
	if err != nil {
		return commands.CancelClaimCommand{}, fmt.Errorf("invalid claim_id: %w", err)
	}

	requestedBy, err := parseUUID(userID)
	if err != nil {
		return commands.CancelClaimCommand{}, fmt.Errorf("invalid user_id: %w", err)
	}

	// Proto has: claim_id, optional reason
	reason := ""
	if req.Reason != nil {
		reason = *req.Reason
	}

	return commands.CancelClaimCommand{
		ClaimID:     claimID,
		RequestedBy: requestedBy,
		// TwoFactorCode: from auth context/header
		Reason: reason,
	}, nil
}

// ============================================================================
// Domain Claim → Proto ClaimSummary
// ============================================================================

func MapDomainClaimToProtoSummary(claim *entities.Claim) *corev1.ClaimSummary {
	if claim == nil {
		return nil
	}

	// TODO: Claim entity doesn't have separate KeyType - only EntryKey (string)
	// Need to either: 1) Add KeyType to Claim entity, or 2) Parse from EntryKey format
	return &corev1.ClaimSummary{
		ClaimId: claim.ID.String(),
		EntryId: "", // TODO: Claim doesn't store EntryID separately - needs to be looked up
		Key: &commonv1.DictKey{
			KeyType:  commonv1.KeyType_KEY_TYPE_UNSPECIFIED, // TODO: Parse from EntryKey
			KeyValue: claim.EntryKey,
		},
		Status:        MapDomainClaimStatusToProto(claim.Status),
		CreatedAt:     timestamppb.New(claim.CreatedAt),
		ExpiresAt:     timestamppb.New(claim.ExpiresAt),
		DaysRemaining: CalculateDaysRemaining(claim.ExpiresAt),
	}
}

// ============================================================================
// Domain Claim → Proto GetClaimStatusResponse
// ============================================================================

func MapDomainClaimToProtoGetClaimStatusResponse(claim *entities.Claim, entry *entities.Entry) *corev1.GetClaimStatusResponse {
	if claim == nil {
		return nil
	}

	resp := &corev1.GetClaimStatusResponse{
		ClaimId: claim.ID.String(),
		EntryId: "", // TODO: Not stored in Claim entity - needs lookup or add to entity
		Key: &commonv1.DictKey{
			KeyType:  commonv1.KeyType_KEY_TYPE_UNSPECIFIED, // TODO: Parse from EntryKey
			KeyValue: claim.EntryKey,
		},
		Status:        MapDomainClaimStatusToProto(claim.Status),
		ClaimerIspb:   claim.ClaimerParticipant.ISPB,
		OwnerIspb:     claim.DonorParticipant.ISPB, // Current owner
		CreatedAt:     timestamppb.New(claim.CreatedAt),
		ExpiresAt:     timestamppb.New(claim.ExpiresAt),
		DaysRemaining: CalculateDaysRemaining(claim.ExpiresAt),
	}

	if claim.ResolutionDate != nil && !claim.ResolutionDate.IsZero() {
		resp.CompletedAt = timestamppb.New(*claim.ResolutionDate)
	}

	return resp
}

// ============================================================================
// Proto StartClaimResponse helpers
// ============================================================================

func MapDomainClaimToProtoStartClaimResponse(claim *entities.Claim) *corev1.StartClaimResponse {
	if claim == nil {
		return nil
	}

	return &corev1.StartClaimResponse{
		ClaimId:   claim.ID.String(),
		EntryId:   "", // TODO: Not stored - needs lookup
		Status:    MapDomainClaimStatusToProto(claim.Status),
		ExpiresAt: timestamppb.New(claim.ExpiresAt),
		CreatedAt: timestamppb.New(claim.CreatedAt),
		Message:   FormatClaimMessage(claim.Status, claim.ExpiresAt),
	}
}

// ============================================================================
// Proto RespondToClaimResponse helpers
// ============================================================================

func MapDomainClaimToProtoRespondToClaimResponse(claim *entities.Claim) *corev1.RespondToClaimResponse {
	if claim == nil {
		return nil
	}

	respondedAt := claim.UpdatedAt
	if claim.ResolutionDate != nil && !claim.ResolutionDate.IsZero() {
		respondedAt = *claim.ResolutionDate
	}

	return &corev1.RespondToClaimResponse{
		ClaimId:     claim.ID.String(),
		NewStatus:   MapDomainClaimStatusToProto(claim.Status),
		RespondedAt: timestamppb.New(respondedAt),
		Message:     FormatClaimResponseMessage(claim.Status),
	}
}

// ============================================================================
// Proto ListIncomingClaimsRequest → Application ListClaimsQuery
// ============================================================================

func MapProtoListIncomingClaimsRequestToQuery(req *corev1.ListIncomingClaimsRequest, ispb string) queries.ListClaimsQuery {
	// Proto has: optional status, page_size, page_token
	// Page is 1-indexed (computed from page_token)
	page := 1 // TODO: Extract from page_token

	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = 100
	}
	if pageSize > 1000 {
		pageSize = 1000
	}

	// ISPB comes from auth context (authenticated participant)
	return queries.ListClaimsQuery{
		ISPB:     ispb, // ISPB of the donor (current owner)
		Page:     page,
		PageSize: pageSize,
	}
}

// ============================================================================
// Proto ListOutgoingClaimsRequest → Application ListClaimsQuery
// ============================================================================

func MapProtoListOutgoingClaimsRequestToQuery(req *corev1.ListOutgoingClaimsRequest, ispb string) queries.ListClaimsQuery {
	// Proto has: optional status, page_size, page_token
	// Page is 1-indexed (computed from page_token)
	page := 1 // TODO: Extract from page_token

	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = 100
	}
	if pageSize > 1000 {
		pageSize = 1000
	}

	// ISPB comes from auth context (authenticated participant)
	return queries.ListClaimsQuery{
		ISPB:     ispb, // ISPB of the claimer (requesting ownership)
		Page:     page,
		PageSize: pageSize,
	}
}

// ============================================================================
// Helpers
// ============================================================================

// CalculateDaysRemaining calculates days remaining until expiration
func CalculateDaysRemaining(expiresAt time.Time) int32 {
	if expiresAt.IsZero() {
		return 0
	}

	now := time.Now()
	if expiresAt.Before(now) {
		return 0 // Expired
	}

	duration := expiresAt.Sub(now)
	days := int32(duration.Hours() / 24)

	// Round up partial days
	if duration.Hours() > float64(days*24) {
		days++
	}

	return days
}

// FormatClaimMessage formats user-friendly message based on claim status
func FormatClaimMessage(status valueobjects.ClaimStatus, expiresAt time.Time) string {
	daysRemaining := CalculateDaysRemaining(expiresAt)

	switch status {
	case valueobjects.ClaimStatusOpen:
		return "Claim created. The current owner has 30 days to respond. Days remaining: " + string(rune(daysRemaining))
	case valueobjects.ClaimStatusWaitingResolution:
		return "Claim is waiting for resolution by the current owner."
	case valueobjects.ClaimStatusConfirmed:
		return "Claim confirmed by the owner. Transfer will be completed."
	case valueobjects.ClaimStatusCancelled:
		return "Claim cancelled. Key remains with current owner."
	case valueobjects.ClaimStatusCompleted:
		return "Claim completed successfully. Key ownership transferred."
	default:
		return "Claim status unknown."
	}
}

// FormatClaimResponseMessage formats message after responding to claim
func FormatClaimResponseMessage(status valueobjects.ClaimStatus) string {
	switch status {
	case valueobjects.ClaimStatusConfirmed:
		return "Claim accepted successfully. Key will be transferred."
	case valueobjects.ClaimStatusCancelled:
		return "Claim rejected. Key remains with current owner."
	case valueobjects.ClaimStatusCompleted:
		return "Claim transfer completed."
	default:
		return "Claim response processed."
	}
}

// ============================================================================
// Helper: ClaimType conversion (proto string → commands.ClaimType)
// ============================================================================

func mapProtoClaimTypeString(claimTypeStr string) commands.ClaimType {
	switch claimTypeStr {
	case "OWNERSHIP":
		return commands.ClaimTypeOwnership
	case "PORTABILITY":
		return commands.ClaimTypePortability
	default:
		return commands.ClaimTypePortability // Default
	}
}
