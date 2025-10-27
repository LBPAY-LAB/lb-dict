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

// InitiatePortability handles the InitiatePortability RPC call
// Initiates a portability process to move a DICT key to a new account
func (s *Server) InitiatePortability(ctx context.Context, req *pb.InitiatePortabilityRequest) (*pb.InitiatePortabilityResponse, error) {
	s.logger.Infof("InitiatePortability called: entry_id=%s, key=%v, new_account=%v",
		req.EntryId, req.Key, req.NewAccount)

	// Validate request
	if err := s.validateInitiatePortabilityRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
	}

	// Step 1: Convert gRPC request to XML
	xmlData, err := xml.InitiatePortabilityRequestToXML(req)
	if err != nil {
		s.logger.Errorf("Failed to convert InitiatePortability request to XML: %v", err)
		return nil, status.Errorf(codes.Internal, "XML conversion failed: %v", err)
	}

	// Step 2: Sign XML with ICP-Brasil A3
	_ = xmlData // TODO: Integrate with XML signer when portability handlers are fully implemented
	s.logger.Warn("XML signing not yet implemented - using unsigned XML (DEV MODE)")

	// Step 3: Send signed XML to Bacen via SOAP/mTLS
	// TODO: Integrate with HTTPClient/SOAP Client
	s.logger.Info("SOAP call to Bacen not yet implemented - returning placeholder (DEV MODE)")

	// Step 4: Parse Bacen response
	now := time.Now()

	return &pb.InitiatePortabilityResponse{
		PortabilityId:      fmt.Sprintf("port-%d", now.UnixNano()),
		EntryId:            req.EntryId,
		Status:             commonv1.EntryStatus_ENTRY_STATUS_PORTABILITY_PENDING,
		InitiatedAt:        timestamppb.New(now),
		BacenTransactionId: fmt.Sprintf("tx-port-init-%d", now.UnixNano()),
	}, nil
}

// ConfirmPortability handles the ConfirmPortability RPC call
// Confirms and completes a portability process
func (s *Server) ConfirmPortability(ctx context.Context, req *pb.ConfirmPortabilityRequest) (*pb.ConfirmPortabilityResponse, error) {
	s.logger.Infof("ConfirmPortability called: entry_id=%s, portability_id=%s",
		req.EntryId, req.PortabilityId)

	// Validate request
	if err := s.validateConfirmPortabilityRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
	}

	// Step 1: Convert gRPC request to XML
	xmlData, err := xml.ConfirmPortabilityRequestToXML(req)
	if err != nil {
		s.logger.Errorf("Failed to convert ConfirmPortability request to XML: %v", err)
		return nil, status.Errorf(codes.Internal, "XML conversion failed: %v", err)
	}

	// Step 2: Sign XML with ICP-Brasil A3
	_ = xmlData // TODO: Integrate with XML signer when portability handlers are fully implemented
	s.logger.Warn("XML signing not yet implemented - using unsigned XML (DEV MODE)")

	// Step 3: Send signed XML to Bacen via SOAP/mTLS
	s.logger.Info("SOAP call to Bacen not yet implemented - returning placeholder (DEV MODE)")

	// Step 4: Parse Bacen response and return updated account
	now := time.Now()

	return &pb.ConfirmPortabilityResponse{
		EntryId:       req.EntryId,
		PortabilityId: req.PortabilityId,
		Status:        commonv1.EntryStatus_ENTRY_STATUS_PORTABILITY_CONFIRMED,
		Account: &commonv1.Account{
			Ispb:                   req.NewAccount.Ispb,
			AccountType:            req.NewAccount.AccountType,
			AccountNumber:          req.NewAccount.AccountNumber,
			AccountCheckDigit:      req.NewAccount.AccountCheckDigit,
			BranchCode:             req.NewAccount.BranchCode,
			AccountHolderName:      req.NewAccount.AccountHolderName,
			AccountHolderDocument:  req.NewAccount.AccountHolderDocument,
			DocumentType:           req.NewAccount.DocumentType,
		},
		ConfirmedAt:        timestamppb.New(now),
		BacenTransactionId: fmt.Sprintf("tx-port-confirm-%d", now.UnixNano()),
	}, nil
}

// CancelPortability handles the CancelPortability RPC call
// Cancels a portability process and reverts to original state
func (s *Server) CancelPortability(ctx context.Context, req *pb.CancelPortabilityRequest) (*pb.CancelPortabilityResponse, error) {
	s.logger.Infof("CancelPortability called: entry_id=%s, portability_id=%s, reason=%s",
		req.EntryId, req.PortabilityId, req.Reason)

	// Validate request
	if err := s.validateCancelPortabilityRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
	}

	// Step 1: Convert gRPC request to XML
	xmlData, err := xml.CancelPortabilityRequestToXML(req)
	if err != nil {
		s.logger.Errorf("Failed to convert CancelPortability request to XML: %v", err)
		return nil, status.Errorf(codes.Internal, "XML conversion failed: %v", err)
	}

	// Step 2: Sign XML with ICP-Brasil A3
	_ = xmlData // TODO: Integrate with XML signer when portability handlers are fully implemented
	s.logger.Warn("XML signing not yet implemented - using unsigned XML (DEV MODE)")

	// Step 3: Send signed XML to Bacen via SOAP/mTLS
	s.logger.Info("SOAP call to Bacen not yet implemented - returning placeholder (DEV MODE)")

	// Step 4: Parse Bacen response
	now := time.Now()

	return &pb.CancelPortabilityResponse{
		EntryId:            req.EntryId,
		PortabilityId:      req.PortabilityId,
		Status:             commonv1.EntryStatus_ENTRY_STATUS_ACTIVE, // Reverts to ACTIVE
		CancelledAt:        timestamppb.New(now),
		BacenTransactionId: fmt.Sprintf("tx-port-cancel-%d", now.UnixNano()),
	}, nil
}

// ========== VALIDATION FUNCTIONS ==========

// validateInitiatePortabilityRequest validates the InitiatePortability request
func (s *Server) validateInitiatePortabilityRequest(req *pb.InitiatePortabilityRequest) error {
	if req.EntryId == "" {
		return fmt.Errorf("entry_id is required")
	}
	if req.Key == nil {
		return fmt.Errorf("key is required")
	}
	if req.Key.KeyType == commonv1.KeyType_KEY_TYPE_UNSPECIFIED {
		return fmt.Errorf("key.key_type is required")
	}
	if req.Key.KeyValue == "" {
		return fmt.Errorf("key.key_value is required")
	}
	if req.NewAccount == nil {
		return fmt.Errorf("new_account is required")
	}
	if req.NewAccount.Ispb == "" {
		return fmt.Errorf("new_account.ispb is required")
	}
	if req.NewAccount.AccountNumber == "" {
		return fmt.Errorf("new_account.account_number is required")
	}
	if req.NewAccount.BranchCode == "" {
		return fmt.Errorf("new_account.branch_code is required")
	}
	return nil
}

// validateConfirmPortabilityRequest validates the ConfirmPortability request
func (s *Server) validateConfirmPortabilityRequest(req *pb.ConfirmPortabilityRequest) error {
	if req.EntryId == "" {
		return fmt.Errorf("entry_id is required")
	}
	if req.PortabilityId == "" {
		return fmt.Errorf("portability_id is required")
	}
	if req.NewAccount == nil {
		return fmt.Errorf("new_account is required")
	}
	if req.NewAccount.Ispb == "" {
		return fmt.Errorf("new_account.ispb is required")
	}
	if req.NewAccount.AccountNumber == "" {
		return fmt.Errorf("new_account.account_number is required")
	}
	return nil
}

// validateCancelPortabilityRequest validates the CancelPortability request
func (s *Server) validateCancelPortabilityRequest(req *pb.CancelPortabilityRequest) error {
	if req.EntryId == "" {
		return fmt.Errorf("entry_id is required")
	}
	if req.PortabilityId == "" {
		return fmt.Errorf("portability_id is required")
	}
	if req.Reason == "" {
		return fmt.Errorf("reason is required")
	}
	return nil
}
