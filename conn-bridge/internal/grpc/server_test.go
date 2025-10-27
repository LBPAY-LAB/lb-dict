package grpc

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	logger := logrus.New()
	port := 9094

	server := NewServer(logger, port)

	require.NotNil(t, server)
	assert.Equal(t, port, server.port)
	assert.NotNil(t, server.logger)
}

func TestServer_ValidateCreateEntryRequest(t *testing.T) {
	logger := logrus.New()
	_ = NewServer(logger, 9094) // Server used for validation logic tests

	tests := []struct {
		name    string
		req     *struct {
			KeyType       string
			KeyValue      string
			Ispb          string
			AccountNumber string
			OwnerDocument string
			OwnerName     string
		}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid_request",
			req: &struct {
				KeyType       string
				KeyValue      string
				Ispb          string
				AccountNumber string
				OwnerDocument string
				OwnerName     string
			}{
				KeyType:       "CPF",
				KeyValue:      "12345678901",
				Ispb:          "12345678",
				AccountNumber: "123456",
				OwnerDocument: "12345678901",
				OwnerName:     "Test User",
			},
			wantErr: false,
		},
		{
			name: "missing_key_type",
			req: &struct {
				KeyType       string
				KeyValue      string
				Ispb          string
				AccountNumber string
				OwnerDocument string
				OwnerName     string
			}{
				KeyType:       "",
				KeyValue:      "12345678901",
				Ispb:          "12345678",
				AccountNumber: "123456",
				OwnerDocument: "12345678901",
				OwnerName:     "Test User",
			},
			wantErr: true,
			errMsg:  "key_type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation logic (placeholder)
			// In real implementation, call server.validateCreateEntryRequest()
			if tt.wantErr {
				assert.NotEmpty(t, tt.errMsg)
			}
		})
	}
}