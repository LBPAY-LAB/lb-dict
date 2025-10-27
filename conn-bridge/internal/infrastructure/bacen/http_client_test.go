package bacen

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lbpay-lab/conn-bridge/internal/domain/entities"
	xmlstructs "github.com/lbpay-lab/conn-bridge/internal/xml"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPClient(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantErr   bool
		errString string
	}{
		{
			name: "valid config with dev mode",
			config: &Config{
				BaseURL: "https://dict-hom.bcb.gov.br",
				DevMode: true,
				Logger:  logrus.New(),
			},
			wantErr: false,
		},
		{
			name: "valid config with defaults",
			config: &Config{
				BaseURL: "https://dict-hom.bcb.gov.br",
				DevMode: true,
			},
			wantErr: false,
		},
		{
			name:      "missing base URL",
			config:    &Config{},
			wantErr:   true,
			errString: "base URL is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewHTTPClient(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errString != "" {
					assert.Contains(t, err.Error(), tt.errString)
				}
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestConfigureTLS(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantErr   bool
		errString string
	}{
		{
			name: "dev mode - insecure skip verify",
			config: &Config{
				DevMode: true,
				Logger:  logrus.New(),
			},
			wantErr: false,
		},
		{
			name: "production mode - missing certs",
			config: &Config{
				DevMode: false,
				Logger:  logrus.New(),
			},
			wantErr:   true,
			errString: "certificate and key paths are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tlsConfig, err := configureTLS(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errString != "" {
					assert.Contains(t, err.Error(), tt.errString)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tlsConfig)
				if tt.config.DevMode {
					assert.True(t, tlsConfig.InsecureSkipVerify)
				}
				assert.GreaterOrEqual(t, tlsConfig.MinVersion, uint16(tls.VersionTLS12))
			}
		})
	}
}

func TestCreateEntry(t *testing.T) {
	// Create a test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and headers
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/xml; charset=utf-8", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/xml", r.Header.Get("Accept"))

		// Create response
		response := &xmlstructs.XMLCreateEntryResponse{
			ResponseTime:  time.Now().Format(time.RFC3339),
			CorrelationId: "test-correlation-id",
			Entry: xmlstructs.XMLExtendedEntry{
				Key:     "12345678901",
				KeyType: "CPF",
				Account: xmlstructs.XMLAccount{
					Participant:   "12345678",
					Branch:        "0001",
					AccountNumber: "123456",
					AccountType:   "CHECKING",
				},
				Owner: xmlstructs.XMLOwner{
					Type:        "PERSON",
					TaxIdNumber: "12345678901",
					Name:        "Test User",
				},
				CreationTime:     time.Now().Format(time.RFC3339),
				KeyOwnershipDate: time.Now().Format(time.RFC3339),
			},
		}

		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusCreated)
		xml.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewHTTPClient(&Config{
		BaseURL: server.URL,
		DevMode: true,
		Logger:  logrus.New(),
	})
	require.NoError(t, err)

	// Create test entry
	entry := &entities.DictEntry{
		Key:  "12345678901",
		Type: entities.KeyTypeCPF,
		Account: entities.Account{
			ISPB:        "12345678",
			Branch:      "0001",
			Number:      "123456",
			Type:        entities.AccountTypeChecking,
			OpeningDate: time.Now(),
		},
		Owner: entities.Owner{
			Type:     entities.OwnerTypePerson,
			Document: "12345678901",
			Name:     "Test User",
		},
	}

	// Test CreateEntry
	ctx := context.Background()
	httpClient := client.(*HTTPClient)
	response, err := httpClient.CreateEntry(ctx, entry)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "test-correlation-id", response.CorrelationId)
	assert.Equal(t, "12345678901", response.Entry.Key)
}

func TestUpdateEntry(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)

		response := &xmlstructs.XMLUpdateEntryResponse{
			ResponseTime:  time.Now().Format(time.RFC3339),
			CorrelationId: "update-correlation-id",
			Entry: xmlstructs.XMLExtendedEntry{
				Key:     "12345678901",
				KeyType: "CPF",
			},
		}

		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		xml.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewHTTPClient(&Config{
		BaseURL: server.URL,
		DevMode: true,
		Logger:  logrus.New(),
	})
	require.NoError(t, err)

	entry := &entities.DictEntry{
		Key:  "12345678901",
		Type: entities.KeyTypeCPF,
		Account: entities.Account{
			ISPB:   "12345678",
			Branch: "0002",
			Number: "654321",
			Type:   entities.AccountTypeSavings,
		},
	}

	ctx := context.Background()
	httpClient := client.(*HTTPClient)
	response, err := httpClient.UpdateEntry(ctx, entry)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "update-correlation-id", response.CorrelationId)
}

func TestDeleteEntry(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)

		response := &xmlstructs.XMLDeleteEntryResponse{
			ResponseTime:  time.Now().Format(time.RFC3339),
			CorrelationId: "delete-correlation-id",
			Deleted:       true,
			Key:           "12345678901",
			KeyType:       "CPF",
		}

		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		xml.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewHTTPClient(&Config{
		BaseURL: server.URL,
		DevMode: true,
		Logger:  logrus.New(),
	})
	require.NoError(t, err)

	ctx := context.Background()
	httpClient := client.(*HTTPClient)
	response, err := httpClient.DeleteEntry(ctx, "12345678901", entities.KeyTypeCPF)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.True(t, response.Deleted)
	assert.Equal(t, "delete-correlation-id", response.CorrelationId)
}

func TestGetEntry(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		response := &xmlstructs.XMLGetEntryResponse{
			ResponseTime:  time.Now().Format(time.RFC3339),
			CorrelationId: "get-correlation-id",
			Entry: xmlstructs.XMLExtendedEntry{
				Key:     "12345678901",
				KeyType: "CPF",
				Account: xmlstructs.XMLAccount{
					Participant:   "12345678",
					Branch:        "0001",
					AccountNumber: "123456",
					AccountType:   "CHECKING",
				},
				Owner: xmlstructs.XMLOwner{
					Type:        "PERSON",
					TaxIdNumber: "12345678901",
					Name:        "Test User",
				},
			},
		}

		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		xml.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewHTTPClient(&Config{
		BaseURL: server.URL,
		DevMode: true,
		Logger:  logrus.New(),
	})
	require.NoError(t, err)

	ctx := context.Background()
	httpClient := client.(*HTTPClient)
	response, err := httpClient.GetEntry(ctx, "12345678901", entities.KeyTypeCPF)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "get-correlation-id", response.CorrelationId)
	assert.Equal(t, "12345678901", response.Entry.Key)
}

func TestHealthCheck(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "healthy",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "unhealthy",
			statusCode: http.StatusServiceUnavailable,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client, err := NewHTTPClient(&Config{
				BaseURL: server.URL,
				DevMode: true,
				Logger:  logrus.New(),
			})
			require.NoError(t, err)

			ctx := context.Background()
			err = client.HealthCheck(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRetryWithBackoff(t *testing.T) {
	attempts := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		response := &xmlstructs.XMLGetEntryResponse{
			ResponseTime:  time.Now().Format(time.RFC3339),
			CorrelationId: "retry-correlation-id",
			Entry: xmlstructs.XMLExtendedEntry{
				Key:     "12345678901",
				KeyType: "CPF",
			},
		}

		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		xml.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewHTTPClient(&Config{
		BaseURL:    server.URL,
		DevMode:    true,
		Logger:     logrus.New(),
		MaxRetries: 3,
	})
	require.NoError(t, err)

	ctx := context.Background()
	httpClient := client.(*HTTPClient)
	response, err := httpClient.GetEntry(ctx, "12345678901", entities.KeyTypeCPF)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 3, attempts, "should have retried twice before succeeding")
}

func TestSetTimeout(t *testing.T) {
	client, err := NewHTTPClient(&Config{
		BaseURL: "https://dict-hom.bcb.gov.br",
		DevMode: true,
		Logger:  logrus.New(),
	})
	require.NoError(t, err)

	// Test valid timeout
	err = client.SetTimeout(30)
	assert.NoError(t, err)

	// Test invalid timeout
	err = client.SetTimeout(0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout must be positive")

	err = client.SetTimeout(-1)
	assert.Error(t, err)
}

func TestGetEndpoint(t *testing.T) {
	baseURL := "https://dict-hom.bcb.gov.br"
	client, err := NewHTTPClient(&Config{
		BaseURL: baseURL,
		DevMode: true,
		Logger:  logrus.New(),
	})
	require.NoError(t, err)

	assert.Equal(t, baseURL, client.GetEndpoint())
}

func TestMaskSensitiveData(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "CPF",
			input:    "12345678901",
			expected: "12****01",
		},
		{
			name:     "short data",
			input:    "123",
			expected: "****",
		},
		{
			name:     "email",
			input:    "test@example.com",
			expected: "te****om",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskSensitiveData(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsClientError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "4xx error",
			err:      assert.AnError,
			expected: false,
		},
		{
			name:     "HTTP 400 error",
			err:      &clientError{msg: "HTTP 400: Bad Request"},
			expected: true,
		},
		{
			name:     "HTTP 404 error",
			err:      &clientError{msg: "HTTP 404: Not Found"},
			expected: true,
		},
		{
			name:     "HTTP 500 error",
			err:      &clientError{msg: "HTTP 500: Internal Server Error"},
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isClientError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper type for testing
type clientError struct {
	msg string
}

func (e *clientError) Error() string {
	return e.msg
}

func TestGenerateRequestID(t *testing.T) {
	id1 := generateRequestID()
	time.Sleep(1 * time.Millisecond)
	id2 := generateRequestID()

	assert.NotEqual(t, id1, id2, "request IDs should be unique")
	assert.Contains(t, id1, "req-")
	assert.Contains(t, id2, "req-")
}

func TestTLSConfiguration(t *testing.T) {
	t.Run("TLS 1.2 minimum version", func(t *testing.T) {
		config := &Config{
			DevMode: true,
			Logger:  logrus.New(),
		}

		tlsConfig, err := configureTLS(config)
		require.NoError(t, err)
		assert.Equal(t, uint16(tls.VersionTLS12), tlsConfig.MinVersion)
		assert.Equal(t, uint16(tls.VersionTLS13), tlsConfig.MaxVersion)
	})
}

func TestConnectionPoolSettings(t *testing.T) {
	client, err := NewHTTPClient(&Config{
		BaseURL: "https://dict-hom.bcb.gov.br",
		DevMode: true,
		Logger:  logrus.New(),
	})
	require.NoError(t, err)

	httpClient := client.(*HTTPClient)
	transport := httpClient.httpClient.Transport.(*http.Transport)

	assert.Equal(t, maxIdleConns, transport.MaxIdleConns)
	assert.Equal(t, maxIdleConnsPerHost, transport.MaxIdleConnsPerHost)
	assert.Equal(t, maxConnsPerHost, transport.MaxConnsPerHost)
}

func TestContextCancellation(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewHTTPClient(&Config{
		BaseURL: server.URL,
		DevMode: true,
		Logger:  logrus.New(),
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	httpClient := client.(*HTTPClient)
	_, err = httpClient.GetEntry(ctx, "12345678901", entities.KeyTypeCPF)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestCACertificateValidation(t *testing.T) {
	t.Run("invalid CA certificate", func(t *testing.T) {
		config := &Config{
			BaseURL:  "https://dict-hom.bcb.gov.br",
			CertPath: "/tmp/cert.pem",
			KeyPath:  "/tmp/key.pem",
			CAPath:   "/tmp/invalid-ca.pem",
			DevMode:  false,
			Logger:   logrus.New(),
		}

		_, err := configureTLS(config)
		assert.Error(t, err)
	})

	t.Run("missing CA certificate", func(t *testing.T) {
		config := &Config{
			BaseURL:  "https://dict-hom.bcb.gov.br",
			CertPath: "/tmp/cert.pem",
			KeyPath:  "/tmp/key.pem",
			DevMode:  false,
			Logger:   logrus.New(),
		}

		_, err := configureTLS(config)
		assert.Error(t, err)
	})
}

func TestCorrelationIDPropagation(t *testing.T) {
	receivedCorrelationID := ""
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedCorrelationID = r.Header.Get("X-Correlation-ID")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewHTTPClient(&Config{
		BaseURL: server.URL,
		DevMode: true,
		Logger:  logrus.New(),
	})
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), "correlationID", "test-correlation-123")
	httpClient := client.(*HTTPClient)
	httpClient.HealthCheck(ctx)

	assert.Equal(t, "test-correlation-123", receivedCorrelationID)
}

func TestAPIKeyHeader(t *testing.T) {
	receivedAPIKey := ""
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAPIKey = r.Header.Get("X-API-Key")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewHTTPClient(&Config{
		BaseURL: server.URL,
		DevMode: true,
		APIKey:  "test-api-key-123",
		Logger:  logrus.New(),
	})
	require.NoError(t, err)

	ctx := context.Background()
	httpClient := client.(*HTTPClient)
	httpClient.HealthCheck(ctx)

	assert.Equal(t, "test-api-key-123", receivedAPIKey)
}
