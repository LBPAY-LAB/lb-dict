package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	connectv1 "github.com/lbpay-lab/dict-contracts/gen/proto/conn_dict/v1"
	commonv1 "github.com/lbpay-lab/dict-contracts/gen/proto/common/v1"
	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"
)

// MockEntryRepository is a mock implementation of EntryRepository
type MockEntryRepository struct {
	mock.Mock
}

func (m *MockEntryRepository) GetByEntryID(ctx context.Context, entryID string) (*entities.Entry, error) {
	args := m.Called(ctx, entryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Entry), args.Error(1)
}

func (m *MockEntryRepository) GetByKey(ctx context.Context, key string) (*entities.Entry, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Entry), args.Error(1)
}

func (m *MockEntryRepository) ListByParticipant(ctx context.Context, ispb string, limit, offset int) ([]*entities.Entry, error) {
	args := m.Called(ctx, ispb, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Entry), args.Error(1)
}

func (m *MockEntryRepository) CountByParticipant(ctx context.Context, ispb string) (int64, error) {
	args := m.Called(ctx, ispb)
	return args.Get(0).(int64), args.Error(1)
}

// MockCacheClient is a mock implementation of cache client
type MockCacheClient struct {
	mock.Mock
}

func (m *MockCacheClient) Get(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func TestQueryHandler_GetEntry(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	tracer := otel.Tracer("test")

	t.Run("success - entry found", func(t *testing.T) {
		mockRepo := new(MockEntryRepository)
		mockCache := new(MockCacheClient)

		now := time.Now()
		expectedEntry := &entities.Entry{
			ID:           uuid.New(),
			EntryID:      "entry-123",
			Key:          "12345678901",
			KeyType:      entities.KeyTypeCPF,
			Participant:  "12345678",
			AccountType:  entities.AccountTypeCACC,
			OwnerType:    entities.OwnerTypeNaturalPerson,
			Status:       entities.EntryStatusActive,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		mockRepo.On("GetByEntryID", mock.Anything, "entry-123").Return(expectedEntry, nil)

		handler := NewQueryHandler(mockRepo, mockCache, logger, tracer)

		req := &connectv1.GetEntryRequest{
			EntryId: "entry-123",
		}

		resp, err := handler.GetEntry(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "entry-123", resp.Entry.EntryId)
		assert.Equal(t, "12345678", resp.Entry.ParticipantIspb)
		assert.Equal(t, commonv1.KeyType_KEY_TYPE_CPF, resp.Entry.KeyType)
		assert.Equal(t, "12345678901", resp.Entry.KeyValue)
		assert.Equal(t, commonv1.EntryStatus_ENTRY_STATUS_ACTIVE, resp.Entry.Status)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - entry not found", func(t *testing.T) {
		mockRepo := new(MockEntryRepository)
		mockCache := new(MockCacheClient)

		mockRepo.On("GetByEntryID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

		handler := NewQueryHandler(mockRepo, mockCache, logger, tracer)

		req := &connectv1.GetEntryRequest{
			EntryId: "nonexistent",
		}

		resp, err := handler.GetEntry(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - missing entry_id", func(t *testing.T) {
		mockRepo := new(MockEntryRepository)
		mockCache := new(MockCacheClient)

		handler := NewQueryHandler(mockRepo, mockCache, logger, tracer)

		req := &connectv1.GetEntryRequest{
			EntryId: "",
		}

		resp, err := handler.GetEntry(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestQueryHandler_GetEntryByKey(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	tracer := otel.Tracer("test")

	t.Run("success - entry found by key", func(t *testing.T) {
		mockRepo := new(MockEntryRepository)
		mockCache := new(MockCacheClient)

		now := time.Now()
		expectedEntry := &entities.Entry{
			ID:           uuid.New(),
			EntryID:      "entry-123",
			Key:          "12345678901",
			KeyType:      entities.KeyTypeCPF,
			Participant:  "12345678",
			AccountType:  entities.AccountTypeCACC,
			OwnerType:    entities.OwnerTypeNaturalPerson,
			Status:       entities.EntryStatusActive,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		mockRepo.On("GetByKey", mock.Anything, "12345678901").Return(expectedEntry, nil)

		handler := NewQueryHandler(mockRepo, mockCache, logger, tracer)

		req := &connectv1.GetEntryByKeyRequest{
			Key: &commonv1.DictKey{
				KeyType:  commonv1.KeyType_KEY_TYPE_CPF,
				KeyValue: "12345678901",
			},
		}

		resp, err := handler.GetEntryByKey(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "entry-123", resp.Entry.EntryId)
		assert.Equal(t, "12345678901", resp.Entry.KeyValue)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - entry not found by key", func(t *testing.T) {
		mockRepo := new(MockEntryRepository)
		mockCache := new(MockCacheClient)

		mockRepo.On("GetByKey", mock.Anything, "00000000000").Return(nil, assert.AnError)

		handler := NewQueryHandler(mockRepo, mockCache, logger, tracer)

		req := &connectv1.GetEntryByKeyRequest{
			Key: &commonv1.DictKey{
				KeyType:  commonv1.KeyType_KEY_TYPE_CPF,
				KeyValue: "00000000000",
			},
		}

		resp, err := handler.GetEntryByKey(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - missing key", func(t *testing.T) {
		mockRepo := new(MockEntryRepository)
		mockCache := new(MockCacheClient)

		handler := NewQueryHandler(mockRepo, mockCache, logger, tracer)

		req := &connectv1.GetEntryByKeyRequest{
			Key: nil,
		}

		resp, err := handler.GetEntryByKey(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestQueryHandler_ListEntries(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	tracer := otel.Tracer("test")

	t.Run("success - list entries with pagination", func(t *testing.T) {
		mockRepo := new(MockEntryRepository)
		mockCache := new(MockCacheClient)

		now := time.Now()
		expectedEntries := []*entities.Entry{
			{
				ID:           uuid.New(),
				EntryID:      "entry-1",
				Key:          "11111111111",
				KeyType:      entities.KeyTypeCPF,
				Participant:  "12345678",
				AccountType:  entities.AccountTypeCACC,
				OwnerType:    entities.OwnerTypeNaturalPerson,
				Status:       entities.EntryStatusActive,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			{
				ID:           uuid.New(),
				EntryID:      "entry-2",
				Key:          "22222222222",
				KeyType:      entities.KeyTypeCPF,
				Participant:  "12345678",
				AccountType:  entities.AccountTypeCACC,
				OwnerType:    entities.OwnerTypeNaturalPerson,
				Status:       entities.EntryStatusActive,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
		}

		mockRepo.On("ListByParticipant", mock.Anything, "12345678", 100, 0).Return(expectedEntries, nil)
		mockRepo.On("CountByParticipant", mock.Anything, "12345678").Return(int64(2), nil)

		handler := NewQueryHandler(mockRepo, mockCache, logger, tracer)

		req := &connectv1.ListEntriesRequest{
			ParticipantIspb: "12345678",
			Limit:           100,
			Offset:          0,
		}

		resp, err := handler.ListEntries(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Entries, 2)
		assert.Equal(t, int32(2), resp.TotalCount)
		assert.Equal(t, int32(100), resp.Limit)
		assert.Equal(t, int32(0), resp.Offset)

		mockRepo.AssertExpectations(t)
	})

	t.Run("success - default limit applied", func(t *testing.T) {
		mockRepo := new(MockEntryRepository)
		mockCache := new(MockCacheClient)

		mockRepo.On("ListByParticipant", mock.Anything, "12345678", 100, 0).Return([]*entities.Entry{}, nil)
		mockRepo.On("CountByParticipant", mock.Anything, "12345678").Return(int64(0), nil)

		handler := NewQueryHandler(mockRepo, mockCache, logger, tracer)

		req := &connectv1.ListEntriesRequest{
			ParticipantIspb: "12345678",
			Limit:           0, // Should default to 100
			Offset:          0,
		}

		resp, err := handler.ListEntries(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(100), resp.Limit)

		mockRepo.AssertExpectations(t)
	})

	t.Run("error - missing participant_ispb", func(t *testing.T) {
		mockRepo := new(MockEntryRepository)
		mockCache := new(MockCacheClient)

		handler := NewQueryHandler(mockRepo, mockCache, logger, tracer)

		req := &connectv1.ListEntriesRequest{
			ParticipantIspb: "",
		}

		resp, err := handler.ListEntries(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestConvertKeyTypeToProto(t *testing.T) {
	tests := []struct {
		name     string
		keyType  entities.KeyType
		expected commonv1.KeyType
	}{
		{"CPF", entities.KeyTypeCPF, commonv1.KeyType_KEY_TYPE_CPF},
		{"CNPJ", entities.KeyTypeCNPJ, commonv1.KeyType_KEY_TYPE_CNPJ},
		{"EMAIL", entities.KeyTypeEMAIL, commonv1.KeyType_KEY_TYPE_EMAIL},
		{"PHONE", entities.KeyTypePHONE, commonv1.KeyType_KEY_TYPE_PHONE},
		{"EVP", entities.KeyTypeEVP, commonv1.KeyType_KEY_TYPE_EVP},
		{"Unknown", entities.KeyType("UNKNOWN"), commonv1.KeyType_KEY_TYPE_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertKeyTypeToProto(tt.keyType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertAccountTypeToProto(t *testing.T) {
	tests := []struct {
		name        string
		accountType entities.AccountType
		expected    commonv1.AccountType
	}{
		{"CACC", entities.AccountTypeCACC, commonv1.AccountType_ACCOUNT_TYPE_CHECKING},
		{"SLRY", entities.AccountTypeSLRY, commonv1.AccountType_ACCOUNT_TYPE_SALARY},
		{"SVGS", entities.AccountTypeSVGS, commonv1.AccountType_ACCOUNT_TYPE_SAVINGS},
		{"TRAN", entities.AccountTypeTRAN, commonv1.AccountType_ACCOUNT_TYPE_PAYMENT},
		{"Unknown", entities.AccountType("UNKNOWN"), commonv1.AccountType_ACCOUNT_TYPE_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertAccountTypeToProto(tt.accountType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertStatusToProto(t *testing.T) {
	tests := []struct {
		name     string
		status   entities.EntryStatus
		expected commonv1.EntryStatus
	}{
		{"Active", entities.EntryStatusActive, commonv1.EntryStatus_ENTRY_STATUS_ACTIVE},
		{"Portability Pending", entities.EntryStatusPortabilityPending, commonv1.EntryStatus_ENTRY_STATUS_PORTABILITY_PENDING},
		{"Ownership Change Pending", entities.EntryStatusOwnershipChangePending, commonv1.EntryStatus_ENTRY_STATUS_CLAIM_PENDING},
		{"Inactive", entities.EntryStatusInactive, commonv1.EntryStatus_ENTRY_STATUS_DELETED},
		{"Blocked", entities.EntryStatusBlocked, commonv1.EntryStatus_ENTRY_STATUS_DELETED},
		{"Unknown", entities.EntryStatus("UNKNOWN"), commonv1.EntryStatus_ENTRY_STATUS_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertStatusToProto(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMaskKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"Long key", "12345678901", "12****01"},
		{"Short key", "123", "***"},
		{"Four chars", "1234", "12****34"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskKey(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}
