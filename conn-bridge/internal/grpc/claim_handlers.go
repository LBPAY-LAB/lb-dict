package grpc

import (
	"context"
	"fmt"
	"time"

	pb "github.com/lbpay-lab/dict-contracts/gen/proto/bridge/v1"
	commonv1 "github.com/lbpay-lab/dict-contracts/gen/proto/common/v1"
	"github.com/lbpay-lab/conn-bridge/internal/xml"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateClaim handles the CreateClaim RPC call
// Creates a new claim (reivindicação) for a DICT key with 30-day ownership claim period
func (s *Server) CreateClaim(ctx context.Context, req *pb.CreateClaimRequest) (*pb.CreateClaimResponse, error) {
	s.logger.Infof("CreateClaim called: entry_id=%s, key_type=%v, key_value=%s, claimer_ispb=%s",
		req.EntryId, req.KeyType, maskKey(req.KeyValue), req.ClaimerIspb)

	// Validate request
	if err := s.validateCreateClaimRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
	}

	// Step 1: Convert gRPC request to XML
	xmlData, err := xml.CreateClaimRequestToXML(req)
	if err != nil {
		s.logger.Errorf("Failed to convert CreateClaim request to XML: %v", err)
		return nil, status.Errorf(codes.Internal, "XML conversion failed: %v", err)
	}

	// Step 2: Sign XML with ICP-Brasil A3 (XML Signer)
	// TODO: Integrate with XML Signer service
	// For now, we'll proceed with unsigned XML (development mode)
	_ = xmlData // signedXML will be used when XML signer is integrated
	s.logger.Warn("XML signing not yet implemented - using unsigned XML (DEV MODE)")

	// Step 3: Send signed XML to Bacen via SOAP/mTLS
	// TODO: Integrate with HTTPClient/SOAP Client
	// For now, return placeholder response
	s.logger.Info("SOAP call to Bacen not yet implemented - returning placeholder (DEV MODE)")

	// Step 4: Parse Bacen response
	// Placeholder response for now
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour) // 30 days from now

	return &pb.CreateClaimResponse{
		ClaimId:              fmt.Sprintf("claim-%d", now.UnixNano()),
		ExternalId:           fmt.Sprintf("bacen-claim-%d", now.UnixNano()),
		Status:               commonv1.ClaimStatus_CLAIM_STATUS_OPEN,
		CompletionPeriodDays: 30,
		ExpiresAt:            timestamppb.New(expiresAt),
		CreatedAt:            timestamppb.New(now),
		BacenClaimId:         fmt.Sprintf("bacen-claim-id-%d", now.UnixNano()),
	}, nil
}

// GetClaim handles the GetClaim RPC call
// Retrieves the current status and details of a claim
func (s *Server) GetClaim(ctx context.Context, req *pb.GetClaimRequest) (*pb.GetClaimResponse, error) {
	s.logger.Infof("GetClaim called: identifier=%v", req.Identifier)

	// Validate request
	if err := s.validateGetClaimRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
	}

	// Step 1: Convert gRPC request to XML
	xmlData, err := xml.GetClaimRequestToXML(req)
	if err != nil {
		s.logger.Errorf("Failed to convert GetClaim request to XML: %v", err)
		return nil, status.Errorf(codes.Internal, "XML conversion failed: %v", err)
	}

	// Step 2: Sign XML (if required by Bacen for GET operations)
	_ = xmlData // signedXML will be used when XML signer is integrated
	s.logger.Debug("XML signing not required for GET operations")

	// Step 3: Send request to Bacen via SOAP/mTLS
	// TODO: Integrate with HTTPClient/SOAP Client
	s.logger.Info("SOAP call to Bacen not yet implemented - returning placeholder (DEV MODE)")

	// Step 4: Parse Bacen response
	// Placeholder response
	now := time.Now()
	createdAt := now.Add(-10 * 24 * time.Hour) // 10 days ago
	expiresAt := createdAt.Add(30 * 24 * time.Hour)
	daysRemaining := int32(expiresAt.Sub(now).Hours() / 24)
	expired := now.After(expiresAt)

	// Get claim ID from request
	var claimId string
	switch id := req.Identifier.(type) {
	case *pb.GetClaimRequest_ClaimId:
		claimId = id.ClaimId
	case *pb.GetClaimRequest_ExternalId:
		claimId = "claim-from-external-" + id.ExternalId
	default:
		claimId = "unknown-claim-id"
	}

	return &pb.GetClaimResponse{
		ClaimId:              claimId,
		ExternalId:           "bacen-external-id",
		EntryId:              "entry-id-placeholder",
		Status:               commonv1.ClaimStatus_CLAIM_STATUS_WAITING_RESOLUTION,
		CompletionPeriodDays: 30,
		ClaimerIspb:          "12345678",
		OwnerIspb:            "87654321",
		CreatedAt:            timestamppb.New(createdAt),
		ExpiresAt:            timestamppb.New(expiresAt),
		Found:                true,
		DaysRemaining:        daysRemaining,
		Expired:              expired,
	}, nil
}

// CompleteClaim handles the CompleteClaim RPC call
// Completes a claim, transferring ownership to the claimer
func (s *Server) CompleteClaim(ctx context.Context, req *pb.CompleteClaimRequest) (*pb.CompleteClaimResponse, error) {
	s.logger.Infof("CompleteClaim called: claim_id=%s, external_id=%s", req.ClaimId, req.ExternalId)

	// Validate request
	if err := s.validateCompleteClaimRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
	}

	// Step 1: Convert gRPC request to XML
	xmlData, err := xml.CompleteClaimRequestToXML(req)
	if err != nil {
		s.logger.Errorf("Failed to convert CompleteClaim request to XML: %v", err)
		return nil, status.Errorf(codes.Internal, "XML conversion failed: %v", err)
	}

	// Step 2: Sign XML with ICP-Brasil A3
	_ = xmlData // TODO: Integrate with XML signer when claim handlers are fully implemented
	s.logger.Warn("XML signing not yet implemented - using unsigned XML (DEV MODE)")

	// Step 3: Send signed XML to Bacen via SOAP/mTLS
	s.logger.Info("SOAP call to Bacen not yet implemented - returning placeholder (DEV MODE)")

	// Step 4: Parse Bacen response and return updated entry
	now := time.Now()

	return &pb.CompleteClaimResponse{
		ClaimId: req.ClaimId,
		Status:  commonv1.ClaimStatus_CLAIM_STATUS_COMPLETED,
		UpdatedEntry: &pb.Entry{
			EntryId:    "entry-id-updated",
			ExternalId: "bacen-entry-id",
			KeyType:    commonv1.KeyType_KEY_TYPE_CPF,
			KeyValue:   "12345678900",
			Account: &commonv1.Account{
				Ispb:          "12345678",
				AccountType:   commonv1.AccountType_ACCOUNT_TYPE_CHECKING,
				AccountNumber: "123456",
				BranchCode:    "0001",
			},
			Status:    commonv1.EntryStatus_ENTRY_STATUS_ACTIVE,
			CreatedAt: timestamppb.New(now.Add(-30 * 24 * time.Hour)),
			UpdatedAt: timestamppb.New(now),
		},
		CompletedAt:        timestamppb.New(now),
		BacenTransactionId: fmt.Sprintf("tx-complete-%d", now.UnixNano()),
	}, nil
}

// CancelClaim handles the CancelClaim RPC call
// Cancels an existing claim (user requested, timeout, or error)
func (s *Server) CancelClaim(ctx context.Context, req *pb.CancelClaimRequest) (*pb.CancelClaimResponse, error) {
	s.logger.Infof("CancelClaim called: claim_id=%s, external_id=%s, reason=%s",
		req.ClaimId, req.ExternalId, req.CancellationReason)

	// Validate request
	if err := s.validateCancelClaimRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
	}

	// Step 1: Convert gRPC request to XML
	xmlData, err := xml.CancelClaimRequestToXML(req)
	if err != nil {
		s.logger.Errorf("Failed to convert CancelClaim request to XML: %v", err)
		return nil, status.Errorf(codes.Internal, "XML conversion failed: %v", err)
	}

	// Step 2: Sign XML with ICP-Brasil A3
	_ = xmlData // TODO: Integrate with XML signer when claim handlers are fully implemented
	s.logger.Warn("XML signing not yet implemented - using unsigned XML (DEV MODE)")

	// Step 3: Send signed XML to Bacen via SOAP/mTLS
	s.logger.Info("SOAP call to Bacen not yet implemented - returning placeholder (DEV MODE)")

	// Step 4: Parse Bacen response
	now := time.Now()

	return &pb.CancelClaimResponse{
		ClaimId:            req.ClaimId,
		Status:             commonv1.ClaimStatus_CLAIM_STATUS_CANCELLED,
		CancelledAt:        timestamppb.New(now),
		BacenTransactionId: fmt.Sprintf("tx-cancel-%d", now.UnixNano()),
	}, nil
}

// ========== VALIDATION FUNCTIONS ==========

// validateCreateClaimRequest validates the CreateClaim request
func (s *Server) validateCreateClaimRequest(req *pb.CreateClaimRequest) error {
	if req.EntryId == "" {
		return fmt.Errorf("entry_id is required")
	}
	if req.KeyType == commonv1.KeyType_KEY_TYPE_UNSPECIFIED {
		return fmt.Errorf("key_type is required")
	}
	if req.KeyValue == "" {
		return fmt.Errorf("key_value is required")
	}
	if req.ClaimerIspb == "" {
		return fmt.Errorf("claimer_ispb is required")
	}
	if req.OwnerIspb == "" {
		return fmt.Errorf("owner_ispb is required")
	}
	if req.ClaimerAccount == nil {
		return fmt.Errorf("claimer_account is required")
	}
	if req.CompletionPeriodDays != 30 {
		return fmt.Errorf("completion_period_days must be 30 (Bacen requirement)")
	}
	return nil
}

// validateGetClaimRequest validates the GetClaim request
func (s *Server) validateGetClaimRequest(req *pb.GetClaimRequest) error {
	if req.Identifier == nil {
		return fmt.Errorf("identifier (claim_id or external_id) is required")
	}

	switch id := req.Identifier.(type) {
	case *pb.GetClaimRequest_ClaimId:
		if id.ClaimId == "" {
			return fmt.Errorf("claim_id cannot be empty")
		}
	case *pb.GetClaimRequest_ExternalId:
		if id.ExternalId == "" {
			return fmt.Errorf("external_id cannot be empty")
		}
	default:
		return fmt.Errorf("unknown identifier type")
	}

	return nil
}

// validateCompleteClaimRequest validates the CompleteClaim request
func (s *Server) validateCompleteClaimRequest(req *pb.CompleteClaimRequest) error {
	if req.ClaimId == "" && req.ExternalId == "" {
		return fmt.Errorf("either claim_id or external_id is required")
	}
	// resolution_reason is optional but recommended
	return nil
}

// validateCancelClaimRequest validates the CancelClaim request
func (s *Server) validateCancelClaimRequest(req *pb.CancelClaimRequest) error {
	if req.ClaimId == "" && req.ExternalId == "" {
		return fmt.Errorf("either claim_id or external_id is required")
	}
	if req.CancellationReason == "" {
		return fmt.Errorf("cancellation_reason is required")
	}
	return nil
}

// ========== HELPER FUNCTIONS ==========

// maskKey masks sensitive key data for logging
func maskKey(key string) string {
	if len(key) <= 4 {
		return "****"
	}
	return key[:2] + "****" + key[len(key)-2:]
}
