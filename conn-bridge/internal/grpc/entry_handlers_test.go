package grpc

import (
	"context"
	"testing"

	pb "github.com/lbpay-lab/dict-contracts/gen/proto/bridge/v1"
	commonv1 "github.com/lbpay-lab/dict-contracts/gen/proto/common/v1"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateEntry(t *testing.T) {
	// Setup
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	server := &Server{
		logger: logger,
	}

	tests := []struct {
		name    string
		req     *pb.CreateEntryRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid_cpf_entry",
			req: &pb.CreateEntryRequest{
				Key: &commonv1.DictKey{
					KeyType:  commonv1.KeyType_KEY_TYPE_CPF,
					KeyValue: "12345678900",
				},
				Account: &commonv1.Account{
					Ispb:              "12345678",
					AccountType:       commonv1.AccountType_ACCOUNT_TYPE_CHECKING,
					AccountNumber:     "123456",
					AccountCheckDigit: "7",
					BranchCode:        "0001",
				},
				IdempotencyKey: "idempotency-123",
				RequestId:      "request-456",
			},
			wantErr: false,
		},
		{
			name: "missing_key",
			req: &pb.CreateEntryRequest{
				Account: &commonv1.Account{
					Ispb:          "12345678",
					AccountNumber: "123456",
				},
			},
			wantErr: true,
			errMsg:  "key is required",
		},
		{
			name: "missing_account",
			req: &pb.CreateEntryRequest{
				Key: &commonv1.DictKey{
					KeyType:  commonv1.KeyType_KEY_TYPE_CPF,
					KeyValue: "12345678900",
				},
			},
			wantErr: true,
			errMsg:  "account is required",
		},
		{
			name: "unspecified_key_type",
			req: &pb.CreateEntryRequest{
				Key: &commonv1.DictKey{
					KeyType:  commonv1.KeyType_KEY_TYPE_UNSPECIFIED,
					KeyValue: "12345678900",
				},
				Account: &commonv1.Account{
					Ispb:          "12345678",
					AccountNumber: "123456",
				},
			},
			wantErr: true,
			errMsg:  "key type is required",
		},
		{
			name: "empty_key_value",
			req: &pb.CreateEntryRequest{
				Key: &commonv1.DictKey{
					KeyType:  commonv1.KeyType_KEY_TYPE_CPF,
					KeyValue: "",
				},
				Account: &commonv1.Account{
					Ispb:          "12345678",
					AccountNumber: "123456",
				},
			},
			wantErr: true,
			errMsg:  "key value is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			resp, err := server.CreateEntry(context.Background(), tt.req)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.NotEmpty(t, resp.EntryId)
				assert.NotEmpty(t, resp.ExternalId)
				assert.Equal(t, commonv1.EntryStatus_ENTRY_STATUS_ACTIVE, resp.Status)
			}
		})
	}
}

func TestUpdateEntry(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	server := &Server{
		logger: logger,
	}

	tests := []struct {
		name    string
		req     *pb.UpdateEntryRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid_update",
			req: &pb.UpdateEntryRequest{
				EntryId: "entry-123",
				NewAccount: &commonv1.Account{
					Ispb:          "12345678",
					AccountNumber: "654321",
				},
			},
			wantErr: false,
		},
		{
			name: "missing_entry_id",
			req: &pb.UpdateEntryRequest{
				NewAccount: &commonv1.Account{
					Ispb:          "12345678",
					AccountNumber: "654321",
				},
			},
			wantErr: true,
			errMsg:  "entry_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.UpdateEntry(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tt.req.EntryId, resp.EntryId)
			}
		})
	}
}

func TestDeleteEntry(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	server := &Server{
		logger: logger,
	}

	tests := []struct {
		name    string
		req     *pb.DeleteEntryRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid_delete",
			req: &pb.DeleteEntryRequest{
				EntryId: "entry-123",
			},
			wantErr: false,
		},
		{
			name: "missing_entry_id",
			req: &pb.DeleteEntryRequest{
				EntryId: "",
			},
			wantErr: true,
			errMsg:  "entry_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.DeleteEntry(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.True(t, resp.Deleted)
			}
		})
	}
}

func TestGetEntry(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	server := &Server{
		logger: logger,
	}

	tests := []struct {
		name    string
		req     *pb.GetEntryRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "get_by_entry_id",
			req: &pb.GetEntryRequest{
				Identifier: &pb.GetEntryRequest_EntryId{
					EntryId: "entry-123",
				},
				RequestId: "request-456",
			},
			wantErr: false,
		},
		{
			name: "get_by_external_id",
			req: &pb.GetEntryRequest{
				Identifier: &pb.GetEntryRequest_ExternalId{
					ExternalId: "bacen-456",
				},
				RequestId: "request-789",
			},
			wantErr: false,
		},
		{
			name: "missing_identifier",
			req: &pb.GetEntryRequest{
				RequestId: "request-789",
			},
			wantErr: true,
			errMsg:  "one of entry_id, external_id, or key_query is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.GetEntry(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.NotEmpty(t, resp.EntryId)
				assert.NotNil(t, resp.Key)
				assert.NotNil(t, resp.Account)
			}
		})
	}
}

func TestValidateCreateEntryRequest(t *testing.T) {
	server := &Server{
		logger: logrus.New(),
	}

	tests := []struct {
		name    string
		req     *pb.CreateEntryRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid_request",
			req: &pb.CreateEntryRequest{
				Key: &commonv1.DictKey{
					KeyType:  commonv1.KeyType_KEY_TYPE_EMAIL,
					KeyValue: "test@example.com",
				},
				Account: &commonv1.Account{
					Ispb:          "12345678",
					AccountNumber: "123456",
				},
			},
			wantErr: false,
		},
		{
			name: "nil_key",
			req: &pb.CreateEntryRequest{
				Account: &commonv1.Account{
					Ispb:          "12345678",
					AccountNumber: "123456",
				},
			},
			wantErr: true,
			errMsg:  "key is required",
		},
		{
			name: "nil_account",
			req: &pb.CreateEntryRequest{
				Key: &commonv1.DictKey{
					KeyType:  commonv1.KeyType_KEY_TYPE_CPF,
					KeyValue: "12345678900",
				},
			},
			wantErr: true,
			errMsg:  "account is required",
		},
		{
			name: "empty_ispb",
			req: &pb.CreateEntryRequest{
				Key: &commonv1.DictKey{
					KeyType:  commonv1.KeyType_KEY_TYPE_CPF,
					KeyValue: "12345678900",
				},
				Account: &commonv1.Account{
					Ispb:          "",
					AccountNumber: "123456",
				},
			},
			wantErr: true,
			errMsg:  "ISPB is required",
		},
		{
			name: "empty_account_number",
			req: &pb.CreateEntryRequest{
				Key: &commonv1.DictKey{
					KeyType:  commonv1.KeyType_KEY_TYPE_CPF,
					KeyValue: "12345678900",
				},
				Account: &commonv1.Account{
					Ispb:          "12345678",
					AccountNumber: "",
				},
			},
			wantErr: true,
			errMsg:  "account number is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := server.validateCreateEntryRequest(tt.req)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}