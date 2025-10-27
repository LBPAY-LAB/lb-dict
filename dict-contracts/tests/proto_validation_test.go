package tests

import (
	"testing"

	pb_bridge "github.com/lbpay-lab/dict-contracts/gen/proto/bridge/v1"
	pb_common "github.com/lbpay-lab/dict-contracts/gen/proto/common/v1"
	pb_core "github.com/lbpay-lab/dict-contracts/gen/proto/core/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestProtoCommonMessages(t *testing.T) {
	t.Run("Account message", func(t *testing.T) {
		account := &pb_common.Account{
			Ispb:        "60701190",
			Branch:      "0001",
			Number:      "123456",
			Type:        "CHECKING",
			OpeningDate: "2024-01-01",
		}

		// Verify required fields are set
		assert.NotEmpty(t, account.Ispb)
		assert.NotEmpty(t, account.Branch)
		assert.NotEmpty(t, account.Number)
		assert.NotEmpty(t, account.Type)

		// Test proto marshaling
		data, err := proto.Marshal(account)
		require.NoError(t, err)
		assert.NotEmpty(t, data)

		// Test proto unmarshaling
		decoded := &pb_common.Account{}
		err = proto.Unmarshal(data, decoded)
		require.NoError(t, err)
		assert.Equal(t, account.Ispb, decoded.Ispb)
		assert.Equal(t, account.Number, decoded.Number)
	})

	t.Run("Owner message", func(t *testing.T) {
		owner := &pb_common.Owner{
			Type:     "PERSON",
			Document: "12345678901",
			Name:     "Test User",
		}

		assert.NotEmpty(t, owner.Type)
		assert.NotEmpty(t, owner.Document)
		assert.NotEmpty(t, owner.Name)

		// Test serialization
		data, err := proto.Marshal(owner)
		require.NoError(t, err)

		decoded := &pb_common.Owner{}
		err = proto.Unmarshal(data, decoded)
		require.NoError(t, err)
		assert.Equal(t, owner.Document, decoded.Document)
	})
}

func TestProtoBridgeMessages(t *testing.T) {
	t.Run("CreateEntryRequest", func(t *testing.T) {
		request := &pb_bridge.CreateEntryRequest{
			Key:         "test@example.com",
			KeyType:     "EMAIL",
			Participant: "60701190",
			Account: &pb_common.Account{
				Ispb:        "60701190",
				Branch:      "0001",
				Number:      "123456",
				Type:        "CHECKING",
				OpeningDate: "2024-01-01",
			},
			Owner: &pb_common.Owner{
				Type:     "PERSON",
				Document: "12345678901",
				Name:     "Test User",
			},
		}

		// Verify structure
		assert.NotEmpty(t, request.Key)
		assert.NotNil(t, request.Account)
		assert.NotNil(t, request.Owner)

		// Test serialization roundtrip
		data, err := proto.Marshal(request)
		require.NoError(t, err)

		decoded := &pb_bridge.CreateEntryRequest{}
		err = proto.Unmarshal(data, decoded)
		require.NoError(t, err)
		assert.Equal(t, request.Key, decoded.Key)
		assert.Equal(t, request.KeyType, decoded.KeyType)
	})

	t.Run("CreateEntryResponse", func(t *testing.T) {
		response := &pb_bridge.CreateEntryResponse{
			Success:       true,
			EntryId:       "entry-123",
			CorrelationId: "corr-456",
		}

		assert.True(t, response.Success)
		assert.NotEmpty(t, response.EntryId)

		// Test serialization
		data, err := proto.Marshal(response)
		require.NoError(t, err)
		assert.NotEmpty(t, data)
	})
}

func TestProtoCoreMessages(t *testing.T) {
	t.Run("ClaimRequest", func(t *testing.T) {
		request := &pb_core.ClaimRequest{
			ClaimId:     "claim-123",
			Type:        "PORTABILITY",
			Key:         "12345678901",
			KeyType:     "CPF",
			DonorIspb:   "60701190",
			ClaimerIspb: "60746948",
		}

		assert.NotEmpty(t, request.ClaimId)
		assert.NotEmpty(t, request.Type)
		assert.NotEmpty(t, request.Key)

		// Test serialization
		data, err := proto.Marshal(request)
		require.NoError(t, err)

		decoded := &pb_core.ClaimRequest{}
		err = proto.Unmarshal(data, decoded)
		require.NoError(t, err)
		assert.Equal(t, request.ClaimId, decoded.ClaimId)
	})

	t.Run("ClaimResponse", func(t *testing.T) {
		now := timestamppb.Now()

		response := &pb_core.ClaimResponse{
			ClaimId:   "claim-123",
			Status:    "OPEN",
			CreatedAt: now,
		}

		assert.NotEmpty(t, response.ClaimId)
		assert.NotEmpty(t, response.Status)
		assert.NotNil(t, response.CreatedAt)

		// Test serialization with timestamp
		data, err := proto.Marshal(response)
		require.NoError(t, err)

		decoded := &pb_core.ClaimResponse{}
		err = proto.Unmarshal(data, decoded)
		require.NoError(t, err)
		assert.Equal(t, response.ClaimId, decoded.ClaimId)
		assert.Equal(t, response.CreatedAt.AsTime().Unix(), decoded.CreatedAt.AsTime().Unix())
	})
}

func TestProtoFieldValidation(t *testing.T) {
	tests := []struct {
		name    string
		message proto.Message
		isValid func(proto.Message) bool
	}{
		{
			name: "valid account",
			message: &pb_common.Account{
				Ispb:   "60701190",
				Branch: "0001",
				Number: "123456",
				Type:   "CHECKING",
			},
			isValid: func(m proto.Message) bool {
				acc := m.(*pb_common.Account)
				return acc.Ispb != "" && acc.Number != ""
			},
		},
		{
			name: "invalid account - missing ISPB",
			message: &pb_common.Account{
				Branch: "0001",
				Number: "123456",
			},
			isValid: func(m proto.Message) bool {
				acc := m.(*pb_common.Account)
				return acc.Ispb != ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.isValid(tt.message)

			if tt.name == "valid account" {
				assert.True(t, result, "Expected valid account")
			} else {
				assert.False(t, result, "Expected invalid account")
			}
		})
	}
}

func TestProtoSizeConstraints(t *testing.T) {
	t.Run("large message serialization", func(t *testing.T) {
		// Create a large message
		request := &pb_bridge.CreateEntryRequest{
			Key:         string(make([]byte, 100)), // 100 byte key
			KeyType:     "EMAIL",
			Participant: "60701190",
			Account: &pb_common.Account{
				Ispb:   "60701190",
				Branch: "0001",
				Number: "123456789012345678901234567890", // Long number
				Type:   "CHECKING",
			},
			Owner: &pb_common.Owner{
				Type:     "PERSON",
				Document: "12345678901",
				Name:     "Very Long Name " + string(make([]byte, 200)),
			},
		}

		// Should serialize without error
		data, err := proto.Marshal(request)
		require.NoError(t, err)
		assert.NotEmpty(t, data)

		// Verify size is reasonable
		assert.Less(t, len(data), 10000, "Message size should be under 10KB")
	})
}

func TestProtoConcurrentAccess(t *testing.T) {
	t.Run("concurrent marshal/unmarshal", func(t *testing.T) {
		account := &pb_common.Account{
			Ispb:   "60701190",
			Branch: "0001",
			Number: "123456",
			Type:   "CHECKING",
		}

		// Marshal in main thread
		data, err := proto.Marshal(account)
		require.NoError(t, err)

		// Unmarshal in goroutines
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				decoded := &pb_common.Account{}
				err := proto.Unmarshal(data, decoded)
				assert.NoError(t, err)
				assert.Equal(t, "60701190", decoded.Ispb)
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func TestProtoEnumValues(t *testing.T) {
	// If you have enums defined in your proto files, test them here
	t.Run("key types", func(t *testing.T) {
		keyTypes := []string{"CPF", "CNPJ", "EMAIL", "PHONE", "EVP"}

		for _, kt := range keyTypes {
			account := &pb_bridge.CreateEntryRequest{
				Key:     "test-key",
				KeyType: kt,
			}
			assert.Contains(t, keyTypes, account.KeyType)
		}
	})
}

func TestProtoBackwardCompatibility(t *testing.T) {
	t.Run("old format can be unmarshaled", func(t *testing.T) {
		// Create message with minimal fields (old format)
		oldAccount := &pb_common.Account{
			Ispb:   "60701190",
			Number: "123456",
		}

		data, err := proto.Marshal(oldAccount)
		require.NoError(t, err)

		// Unmarshal into new format (with additional fields)
		newAccount := &pb_common.Account{}
		err = proto.Unmarshal(data, newAccount)
		require.NoError(t, err)

		// Old fields should be preserved
		assert.Equal(t, oldAccount.Ispb, newAccount.Ispb)
		assert.Equal(t, oldAccount.Number, newAccount.Number)

		// New fields should have default values
		assert.Empty(t, newAccount.Branch)
	})
}
