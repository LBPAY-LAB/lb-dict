package signer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// Default configuration
	defaultSignerTimeout = 30 * time.Second
	defaultSignerURL     = "http://localhost:8081"
	signEndpoint         = "/sign"
	healthEndpoint       = "/health"
)

// XMLSignerClient is a client for the Java XML Signer service
type XMLSignerClient struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
	logger     *logrus.Logger
}

// Config holds the configuration for the XML Signer client
type Config struct {
	BaseURL string
	Timeout time.Duration
	Logger  *logrus.Logger
}

// SignRequest represents the JSON request to the XML Signer service
type SignRequest struct {
	XML string `json:"xml"`
}

// SignResponse represents the JSON response from the XML Signer service
type SignResponse struct {
	SignedXML string `json:"signedXml"`
	Signature string `json:"signature,omitempty"` // XML Signature element
	Error     string `json:"error,omitempty"`
}

// ErrorResponse represents an error response from the XML Signer service
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// NewXMLSignerClient creates a new XML Signer client
func NewXMLSignerClient(config *Config) (*XMLSignerClient, error) {
	if config.BaseURL == "" {
		config.BaseURL = defaultSignerURL
	}

	if config.Logger == nil {
		config.Logger = logrus.New()
		config.Logger.SetLevel(logrus.InfoLevel)
	}

	if config.Timeout == 0 {
		config.Timeout = defaultSignerTimeout
	}

	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	return &XMLSignerClient{
		baseURL:    config.BaseURL,
		httpClient: httpClient,
		timeout:    config.Timeout,
		logger:     config.Logger,
	}, nil
}

// SignXML signs an XML document using the Java XML Signer service with ICP-Brasil A3
func (c *XMLSignerClient) SignXML(ctx context.Context, xmlData string) (string, error) {
	c.logger.WithFields(logrus.Fields{
		"xmlSize": len(xmlData),
	}).Debug("Signing XML with ICP-Brasil A3")

	// Prepare request
	signReq := SignRequest{
		XML: xmlData,
	}

	jsonData, err := json.Marshal(signReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal sign request: %w", err)
	}

	// Create HTTP request
	url := c.baseURL + signEndpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add correlation ID from context if available
	if correlationID := ctx.Value("correlationID"); correlationID != nil {
		req.Header.Set("X-Correlation-ID", correlationID.(string))
	}

	// Log request
	c.logger.WithFields(logrus.Fields{
		"url":         url,
		"method":      req.Method,
		"xmlSize":     len(xmlData),
		"requestSize": len(jsonData),
	}).Info("Calling XML Signer service")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.WithError(err).Error("Failed to call XML Signer service")
		return "", fmt.Errorf("HTTP request to signer failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Log response
	c.logger.WithFields(logrus.Fields{
		"statusCode": resp.StatusCode,
		"bodySize":   len(body),
	}).Debug("Received response from XML Signer service")

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return "", fmt.Errorf("signer returned HTTP %d: %s", resp.StatusCode, string(body))
		}
		c.logger.WithFields(logrus.Fields{
			"statusCode": resp.StatusCode,
			"error":      errResp.Error,
			"message":    errResp.Message,
			"code":       errResp.Code,
		}).Error("XML Signer service returned error")
		return "", fmt.Errorf("signer error (%d): %s - %s", resp.StatusCode, errResp.Error, errResp.Message)
	}

	// Parse response
	var signResp SignResponse
	if err := json.Unmarshal(body, &signResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal signer response: %w", err)
	}

	// Check for application-level errors
	if signResp.Error != "" {
		c.logger.WithField("error", signResp.Error).Error("XML signing failed")
		return "", fmt.Errorf("XML signing failed: %s", signResp.Error)
	}

	// Validate signed XML
	if signResp.SignedXML == "" {
		return "", fmt.Errorf("signer returned empty signed XML")
	}

	c.logger.WithFields(logrus.Fields{
		"originalSize": len(xmlData),
		"signedSize":   len(signResp.SignedXML),
	}).Info("XML signed successfully with ICP-Brasil A3")

	return signResp.SignedXML, nil
}

// SignXMLAndGetSignature signs an XML document and returns both the signed XML and the signature element
func (c *XMLSignerClient) SignXMLAndGetSignature(ctx context.Context, xmlData string) (signedXML string, signature string, err error) {
	c.logger.WithFields(logrus.Fields{
		"xmlSize": len(xmlData),
	}).Debug("Signing XML and extracting signature")

	signedXML, err = c.SignXML(ctx, xmlData)
	if err != nil {
		return "", "", err
	}

	// TODO: Extract signature element from signed XML if needed
	// For now, we return the full signed XML
	return signedXML, "", nil
}

// HealthCheck checks if the XML Signer service is healthy
func (c *XMLSignerClient) HealthCheck(ctx context.Context) error {
	c.logger.Debug("Performing XML Signer health check")

	url := c.baseURL + healthEndpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.WithError(err).Error("XML Signer health check failed")
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.WithFields(logrus.Fields{
			"statusCode": resp.StatusCode,
			"body":       string(body),
		}).Error("XML Signer health check failed")
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	c.logger.Debug("XML Signer health check passed")
	return nil
}

// GetBaseURL returns the base URL being used
func (c *XMLSignerClient) GetBaseURL() string {
	return c.baseURL
}

// SetTimeout sets the timeout for requests
func (c *XMLSignerClient) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	c.httpClient.Timeout = timeout
}
