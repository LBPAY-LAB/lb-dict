package bacen

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/lbpay-lab/conn-bridge/internal/domain/entities"
	"github.com/lbpay-lab/conn-bridge/internal/domain/interfaces"
	"github.com/lbpay-lab/conn-bridge/internal/domain/valueobjects"
	xmlstructs "github.com/lbpay-lab/conn-bridge/internal/xml"
	"github.com/sirupsen/logrus"
)

const (
	// HTTP timeouts
	defaultConnectionTimeout = 30 * time.Second
	defaultRequestTimeout    = 60 * time.Second
	defaultKeepAlive         = 30 * time.Second
	defaultIdleConnTimeout   = 90 * time.Second

	// Connection pool settings
	maxIdleConns        = 20
	maxIdleConnsPerHost = 20
	maxConnsPerHost     = 20

	// Retry settings
	maxRetries         = 3
	initialBackoff     = 1 * time.Second
	maxBackoff         = 10 * time.Second
	backoffMultiplier  = 2.0

	// DICT API endpoints
	endpointCreateEntry = "/api/v1/dict/entries"
	endpointUpdateEntry = "/api/v1/dict/entries/%s"
	endpointDeleteEntry = "/api/v1/dict/entries/%s"
	endpointGetEntry    = "/api/v1/dict/entries/%s"
	endpointHealthCheck = "/health"
)

// HTTPClient implements the BacenClient interface using HTTP with mTLS
type HTTPClient struct {
	baseURL     string
	httpClient  *http.Client
	timeout     time.Duration
	apiKey      string
	certPath    string
	keyPath     string
	caPath      string
	devMode     bool
	logger      *logrus.Logger
	maxRetries  int
}

// Config holds the configuration for the HTTP client
type Config struct {
	BaseURL     string
	Timeout     time.Duration
	APIKey      string
	CertPath    string
	KeyPath     string
	CAPath      string
	DevMode     bool
	Logger      *logrus.Logger
	MaxRetries  int
}

// NewHTTPClient creates a new Bacen HTTP client with mTLS support
func NewHTTPClient(config *Config) (interfaces.BacenClient, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	if config.Logger == nil {
		config.Logger = logrus.New()
		config.Logger.SetLevel(logrus.InfoLevel)
	}

	if config.Timeout == 0 {
		config.Timeout = defaultRequestTimeout
	}

	if config.MaxRetries == 0 {
		config.MaxRetries = maxRetries
	}

	// Configure TLS
	tlsConfig, err := configureTLS(config)
	if err != nil {
		return nil, fmt.Errorf("failed to configure TLS: %w", err)
	}

	// Create HTTP transport with connection pooling and timeouts
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
		DialContext: (&net.Dialer{
			Timeout:   defaultConnectionTimeout,
			KeepAlive: defaultKeepAlive,
		}).DialContext,
		MaxIdleConns:          maxIdleConns,
		MaxIdleConnsPerHost:   maxIdleConnsPerHost,
		MaxConnsPerHost:       maxConnsPerHost,
		IdleConnTimeout:       defaultIdleConnTimeout,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
	}

	httpClient := &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
	}

	return &HTTPClient{
		baseURL:    config.BaseURL,
		httpClient: httpClient,
		timeout:    config.Timeout,
		apiKey:     config.APIKey,
		certPath:   config.CertPath,
		keyPath:    config.KeyPath,
		caPath:     config.CAPath,
		devMode:    config.DevMode,
		logger:     config.Logger,
		maxRetries: config.MaxRetries,
	}, nil
}

// configureTLS sets up the TLS configuration with mTLS support
func configureTLS(config *Config) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
	}

	// In dev mode, accept self-signed certificates
	if config.DevMode {
		config.Logger.Warn("Running in DEV MODE - accepting self-signed certificates")
		tlsConfig.InsecureSkipVerify = true
		return tlsConfig, nil
	}

	// Load client certificate and key for mTLS
	if config.CertPath == "" || config.KeyPath == "" {
		return nil, fmt.Errorf("certificate and key paths are required for mTLS")
	}

	cert, err := tls.LoadX509KeyPair(config.CertPath, config.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}
	tlsConfig.Certificates = []tls.Certificate{cert}

	// Load CA certificate for server verification
	if config.CAPath != "" {
		caCert, err := os.ReadFile(config.CAPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	return tlsConfig, nil
}

// CreateEntry creates a new DICT entry in Bacen
func (c *HTTPClient) CreateEntry(ctx context.Context, entry *entities.DictEntry) (*xmlstructs.XMLCreateEntryResponse, error) {
	c.logger.WithFields(logrus.Fields{
		"key":     maskSensitiveData(entry.Key),
		"keyType": entry.Type,
	}).Info("Creating DICT entry")

	// Build XML request
	xmlReq := &xmlstructs.XMLCreateEntryRequest{
		Entry: xmlstructs.XMLEntry{
			Key:     entry.Key,
			KeyType: string(entry.Type),
			Account: xmlstructs.XMLAccount{
				Participant:   entry.Account.ISPB,
				Branch:        entry.Account.Branch,
				AccountNumber: entry.Account.Number,
				AccountType:   string(entry.Account.Type),
				OpeningDate:   entry.Account.OpeningDate.Format(time.RFC3339),
			},
			Owner: xmlstructs.XMLOwner{
				Type:        string(entry.Owner.Type),
				TaxIdNumber: entry.Owner.Document,
				Name:        entry.Owner.Name,
			},
		},
		RequestId: generateRequestID(),
	}

	// Marshal to XML
	xmlData, err := xml.MarshalIndent(xmlReq, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal XML request: %w", err)
	}

	// Add XML header
	xmlData = append([]byte(xml.Header), xmlData...)

	// Send request with retry
	var response *xmlstructs.XMLCreateEntryResponse
	err = c.retryWithBackoff(ctx, func() error {
		resp, err := c.doRequest(ctx, http.MethodPost, endpointCreateEntry, xmlData)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			return c.handleErrorResponse(resp)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		response = &xmlstructs.XMLCreateEntryResponse{}
		if err := xml.Unmarshal(body, response); err != nil {
			return fmt.Errorf("failed to unmarshal XML response: %w", err)
		}

		return nil
	})

	if err != nil {
		c.logger.WithError(err).Error("Failed to create DICT entry")
		return nil, err
	}

	c.logger.WithFields(logrus.Fields{
		"key":           maskSensitiveData(entry.Key),
		"correlationId": response.CorrelationId,
	}).Info("DICT entry created successfully")

	return response, nil
}

// UpdateEntry updates an existing DICT entry in Bacen
func (c *HTTPClient) UpdateEntry(ctx context.Context, entry *entities.DictEntry) (*xmlstructs.XMLUpdateEntryResponse, error) {
	c.logger.WithFields(logrus.Fields{
		"key":     maskSensitiveData(entry.Key),
		"keyType": entry.Type,
	}).Info("Updating DICT entry")

	// Build XML request
	xmlReq := &xmlstructs.XMLUpdateEntryRequest{
		Key:     entry.Key,
		KeyType: string(entry.Type),
		NewAccount: xmlstructs.XMLAccount{
			Participant:   entry.Account.ISPB,
			Branch:        entry.Account.Branch,
			AccountNumber: entry.Account.Number,
			AccountType:   string(entry.Account.Type),
		},
		RequestId: generateRequestID(),
	}

	// Marshal to XML
	xmlData, err := xml.MarshalIndent(xmlReq, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal XML request: %w", err)
	}

	xmlData = append([]byte(xml.Header), xmlData...)

	// Send request with retry
	endpoint := fmt.Sprintf(endpointUpdateEntry, entry.Key)
	var response *xmlstructs.XMLUpdateEntryResponse
	err = c.retryWithBackoff(ctx, func() error {
		resp, err := c.doRequest(ctx, http.MethodPut, endpoint, xmlData)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return c.handleErrorResponse(resp)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		response = &xmlstructs.XMLUpdateEntryResponse{}
		if err := xml.Unmarshal(body, response); err != nil {
			return fmt.Errorf("failed to unmarshal XML response: %w", err)
		}

		return nil
	})

	if err != nil {
		c.logger.WithError(err).Error("Failed to update DICT entry")
		return nil, err
	}

	c.logger.WithFields(logrus.Fields{
		"key":           maskSensitiveData(entry.Key),
		"correlationId": response.CorrelationId,
	}).Info("DICT entry updated successfully")

	return response, nil
}

// DeleteEntry deletes a DICT entry from Bacen
func (c *HTTPClient) DeleteEntry(ctx context.Context, keyID string, keyType entities.KeyType) (*xmlstructs.XMLDeleteEntryResponse, error) {
	c.logger.WithFields(logrus.Fields{
		"key":     maskSensitiveData(keyID),
		"keyType": keyType,
	}).Info("Deleting DICT entry")

	// Build XML request
	xmlReq := &xmlstructs.XMLDeleteEntryRequest{
		Key:       keyID,
		KeyType:   string(keyType),
		RequestId: generateRequestID(),
	}

	// Marshal to XML
	xmlData, err := xml.MarshalIndent(xmlReq, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal XML request: %w", err)
	}

	xmlData = append([]byte(xml.Header), xmlData...)

	// Send request with retry
	endpoint := fmt.Sprintf(endpointDeleteEntry, keyID)
	var response *xmlstructs.XMLDeleteEntryResponse
	err = c.retryWithBackoff(ctx, func() error {
		resp, err := c.doRequest(ctx, http.MethodDelete, endpoint, xmlData)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
			return c.handleErrorResponse(resp)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		response = &xmlstructs.XMLDeleteEntryResponse{}
		if err := xml.Unmarshal(body, response); err != nil {
			return fmt.Errorf("failed to unmarshal XML response: %w", err)
		}

		return nil
	})

	if err != nil {
		c.logger.WithError(err).Error("Failed to delete DICT entry")
		return nil, err
	}

	c.logger.WithFields(logrus.Fields{
		"key":           maskSensitiveData(keyID),
		"correlationId": response.CorrelationId,
		"deleted":       response.Deleted,
	}).Info("DICT entry deleted successfully")

	return response, nil
}

// GetEntry retrieves a DICT entry from Bacen
func (c *HTTPClient) GetEntry(ctx context.Context, keyID string, keyType entities.KeyType) (*xmlstructs.XMLGetEntryResponse, error) {
	c.logger.WithFields(logrus.Fields{
		"key":     maskSensitiveData(keyID),
		"keyType": keyType,
	}).Info("Getting DICT entry")

	// Build XML request
	xmlReq := &xmlstructs.XMLGetEntryRequest{
		Key:       keyID,
		KeyType:   string(keyType),
		RequestId: generateRequestID(),
	}

	// Marshal to XML
	xmlData, err := xml.MarshalIndent(xmlReq, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal XML request: %w", err)
	}

	xmlData = append([]byte(xml.Header), xmlData...)

	// Send request with retry
	endpoint := fmt.Sprintf(endpointGetEntry, keyID)
	var response *xmlstructs.XMLGetEntryResponse
	err = c.retryWithBackoff(ctx, func() error {
		resp, err := c.doRequest(ctx, http.MethodGet, endpoint, xmlData)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return c.handleErrorResponse(resp)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		response = &xmlstructs.XMLGetEntryResponse{}
		if err := xml.Unmarshal(body, response); err != nil {
			return fmt.Errorf("failed to unmarshal XML response: %w", err)
		}

		return nil
	})

	if err != nil {
		c.logger.WithError(err).Error("Failed to get DICT entry")
		return nil, err
	}

	c.logger.WithFields(logrus.Fields{
		"key":           maskSensitiveData(keyID),
		"correlationId": response.CorrelationId,
	}).Info("DICT entry retrieved successfully")

	return response, nil
}

// SendRequest sends a request to Bacen and returns the response
// This implements the BacenClient interface for backwards compatibility
func (c *HTTPClient) SendRequest(ctx context.Context, request *valueobjects.BacenRequest) (*valueobjects.BacenResponse, error) {
	c.logger.WithFields(logrus.Fields{
		"requestId":     request.ID,
		"correlationId": request.CorrelationID,
	}).Info("Sending generic request to Bacen")

	var response *valueobjects.BacenResponse
	err := c.retryWithBackoff(ctx, func() error {
		resp, err := c.doRequest(ctx, http.MethodPost, string(request.Payload), request.Payload)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return c.handleErrorResponse(resp)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		response = valueobjects.NewBacenResponse(
			request.ID,
			resp.StatusCode,
			body,
			request.CorrelationID,
		)

		return nil
	})

	if err != nil {
		c.logger.WithError(err).Error("Failed to send request")
		return nil, err
	}

	return response, nil
}

// HealthCheck checks if the Bacen API is healthy
func (c *HTTPClient) HealthCheck(ctx context.Context) error {
	c.logger.Debug("Performing health check")

	resp, err := c.doRequest(ctx, http.MethodGet, endpointHealthCheck, nil)
	if err != nil {
		c.logger.WithError(err).Error("Health check failed")
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.WithField("statusCode", resp.StatusCode).Error("Health check returned non-OK status")
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	c.logger.Debug("Health check passed")
	return nil
}

// GetEndpoint returns the current endpoint being used
func (c *HTTPClient) GetEndpoint() string {
	return c.baseURL
}

// SetTimeout sets the timeout for requests
func (c *HTTPClient) SetTimeout(timeout int) error {
	if timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	c.timeout = time.Duration(timeout) * time.Second
	c.httpClient.Timeout = c.timeout
	return nil
}

// doRequest performs the actual HTTP request with proper headers and context
func (c *HTTPClient) doRequest(ctx context.Context, method, endpoint string, body []byte) (*http.Response, error) {
	url := c.baseURL + endpoint

	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/xml; charset=utf-8")
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("User-Agent", "LBPay-DICT-Bridge/1.0")

	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	// Add correlation ID from context if available
	if correlationID := ctx.Value("correlationID"); correlationID != nil {
		req.Header.Set("X-Correlation-ID", correlationID.(string))
	}

	// Log request (with sensitive data masked)
	c.logger.WithFields(logrus.Fields{
		"method":   method,
		"url":      url,
		"bodySize": len(body),
	}).Debug("Sending HTTP request")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	// Log response
	c.logger.WithFields(logrus.Fields{
		"statusCode": resp.StatusCode,
		"status":     resp.Status,
	}).Debug("Received HTTP response")

	return resp, nil
}

// retryWithBackoff implements exponential backoff retry logic
func (c *HTTPClient) retryWithBackoff(ctx context.Context, fn func() error) error {
	backoff := initialBackoff

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		// Don't retry on context errors
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Don't retry on 4xx errors (client errors)
		if isClientError(err) {
			return err
		}

		// Last attempt, return error
		if attempt == c.maxRetries {
			return fmt.Errorf("max retries exceeded: %w", err)
		}

		// Wait before retry
		c.logger.WithFields(logrus.Fields{
			"attempt": attempt + 1,
			"backoff": backoff,
			"error":   err,
		}).Warn("Request failed, retrying...")

		select {
		case <-time.After(backoff):
			// Increase backoff for next attempt
			backoff = time.Duration(float64(backoff) * backoffMultiplier)
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return fmt.Errorf("unexpected retry loop exit")
}

// handleErrorResponse processes error responses from Bacen
func (c *HTTPClient) handleErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("HTTP %d: failed to read error response", resp.StatusCode)
	}

	c.logger.WithFields(logrus.Fields{
		"statusCode": resp.StatusCode,
		"body":       string(body),
	}).Error("Bacen API returned error")

	return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
}

// isClientError determines if an error is a client error (4xx) that shouldn't be retried
func isClientError(err error) bool {
	if err == nil {
		return false
	}
	// Check if error message contains "HTTP 4" for 4xx errors
	errMsg := err.Error()
	return len(errMsg) > 6 && errMsg[0:6] == "HTTP 4"
}

// maskSensitiveData masks sensitive data for logging
func maskSensitiveData(data string) string {
	if len(data) <= 4 {
		return "****"
	}
	return data[:2] + "****" + data[len(data)-2:]
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}
