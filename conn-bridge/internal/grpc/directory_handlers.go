package grpc

import (
	"context"

	pb "github.com/lbpay-lab/dict-contracts/gen/proto/bridge/v1"
	commonv1 "github.com/lbpay-lab/dict-contracts/gen/proto/common/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GetDirectory handles the GetDirectory RPC call
// Queries the complete DICT directory from Bacen with optional filters
func (s *Server) GetDirectory(ctx context.Context, req *pb.GetDirectoryRequest) (*pb.GetDirectoryResponse, error) {
	s.logger.Infof("GetDirectory called: request_id=%s, key_type=%v, status=%v, page_size=%d",
		req.RequestId, req.KeyType, req.Status, req.PageSize)

	// Validate request
	if req.RequestId == "" {
		return nil, status.Error(codes.InvalidArgument, "request_id is required")
	}

	// Validate page size (max 1000 for performance)
	if req.PageSize <= 0 {
		req.PageSize = 100 // Default page size
	}
	if req.PageSize > 1000 {
		return nil, status.Error(codes.InvalidArgument, "page_size cannot exceed 1000")
	}

	// TODO: Step 1: Convert gRPC request to SOAP XML (GetDirectory)
	// TODO: Step 2: Sign XML with ICP-Brasil A3
	// TODO: Step 3: Query Bacen DICT directory
	// TODO: Step 4: Parse SOAP response
	// TODO: Step 5: Convert XML entries to proto Entry messages

	s.logger.Warnf("GetDirectory not fully implemented - returning mock data")

	// Placeholder response (mock data for now)
	mockEntries := []*pb.Entry{
		{
			EntryId:    "entry-001",
			ExternalId: "bacen-001",
			KeyType:    commonv1.KeyType_KEY_TYPE_CPF,
			KeyValue:   "12345678900",
			Account: &commonv1.Account{
				Ispb:              "12345678",
				AccountType:       commonv1.AccountType_ACCOUNT_TYPE_CHECKING,
				AccountNumber:     "123456",
				AccountCheckDigit: "7",
				BranchCode:        "0001",
			},
			Status:    commonv1.EntryStatus_ENTRY_STATUS_ACTIVE,
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
		},
		{
			EntryId:    "entry-002",
			ExternalId: "bacen-002",
			KeyType:    commonv1.KeyType_KEY_TYPE_EMAIL,
			KeyValue:   "user@example.com",
			Account: &commonv1.Account{
				Ispb:              "12345678",
				AccountType:       commonv1.AccountType_ACCOUNT_TYPE_CHECKING,
				AccountNumber:     "654321",
				AccountCheckDigit: "0",
				BranchCode:        "0001",
			},
			Status:    commonv1.EntryStatus_ENTRY_STATUS_ACTIVE,
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
		},
	}

	return &pb.GetDirectoryResponse{
		Entries:       mockEntries,
		NextPageToken: "", // No more pages in this mock
		TotalCount:    int32(len(mockEntries)),
	}, nil
}

// SearchEntries handles the SearchEntries RPC call
// Searches for entries by specific criteria (account holder, account number, ISPB)
func (s *Server) SearchEntries(ctx context.Context, req *pb.SearchEntriesRequest) (*pb.SearchEntriesResponse, error) {
	s.logger.Infof("SearchEntries called: request_id=%s, document=%s, account=%s, ispb=%s, page_size=%d",
		req.RequestId, req.GetAccountHolderDocument(), req.GetAccountNumber(), req.GetIspb(), req.PageSize)

	// Validate request
	if req.RequestId == "" {
		return nil, status.Error(codes.InvalidArgument, "request_id is required")
	}

	// At least one search criterion must be provided
	if req.GetAccountHolderDocument() == "" && req.GetAccountNumber() == "" && req.GetIspb() == "" {
		return nil, status.Error(codes.InvalidArgument, "at least one search criterion is required (document, account_number, or ispb)")
	}

	// Validate page size
	if req.PageSize <= 0 {
		req.PageSize = 100 // Default page size
	}
	if req.PageSize > 1000 {
		return nil, status.Error(codes.InvalidArgument, "page_size cannot exceed 1000")
	}

	// TODO: Step 1: Convert gRPC request to SOAP XML (SearchEntries)
	// TODO: Step 2: Sign XML with ICP-Brasil A3
	// TODO: Step 3: Execute search query on Bacen
	// TODO: Step 4: Parse SOAP response
	// TODO: Step 5: Convert XML results to proto Entry messages

	s.logger.Warnf("SearchEntries not fully implemented - returning mock data")

	// Placeholder response (mock data filtered by criteria)
	var mockResults []*pb.Entry

	// Simulate search by ISPB
	if req.GetIspb() == "12345678" {
		mockResults = append(mockResults, &pb.Entry{
			EntryId:    "entry-search-001",
			ExternalId: "bacen-search-001",
			KeyType:    commonv1.KeyType_KEY_TYPE_PHONE,
			KeyValue:   "+5511999999999",
			Account: &commonv1.Account{
				Ispb:              "12345678",
				AccountType:       commonv1.AccountType_ACCOUNT_TYPE_SAVINGS,
				AccountNumber:     "789012",
				AccountCheckDigit: "3",
				BranchCode:        "0002",
			},
			Status:    commonv1.EntryStatus_ENTRY_STATUS_ACTIVE,
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
		})
	}

	// Simulate search by account holder document
	if req.GetAccountHolderDocument() == "12345678900" {
		mockResults = append(mockResults, &pb.Entry{
			EntryId:    "entry-search-002",
			ExternalId: "bacen-search-002",
			KeyType:    commonv1.KeyType_KEY_TYPE_CPF,
			KeyValue:   "12345678900",
			Account: &commonv1.Account{
				Ispb:              "12345678",
				AccountType:       commonv1.AccountType_ACCOUNT_TYPE_CHECKING,
				AccountNumber:     "123456",
				AccountCheckDigit: "7",
				BranchCode:        "0001",
			},
			Status:    commonv1.EntryStatus_ENTRY_STATUS_ACTIVE,
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
		})
	}

	return &pb.SearchEntriesResponse{
		Entries:       mockResults,
		NextPageToken: "", // No more pages in this mock
		TotalCount:    int32(len(mockResults)),
	}, nil
}
