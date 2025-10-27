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

	"github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"
)

const (
	// SOAP HTTP timeouts
	soapConnectionTimeout = 30 * time.Second
	soapRequestTimeout    = 60 * time.Second
	soapKeepAlive         = 30 * time.Second
	soapIdleConnTimeout   = 90 * time.Second

	// Connection pool settings
	soapMaxIdleConns        = 20
	soapMaxIdleConnsPerHost = 20
	soapMaxConnsPerHost     = 20

	// SOAP namespaces
	soapEnvelopeNS = "http://www.w3.org/2003/05/soap-envelope"
	soapSecurityNS = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd"
	bacenDictNS    = "http://www.bcb.gov.br/dict/api/v1"
)

// SOAPClient implements SOAP calls to Bacen DICT API with mTLS
type SOAPClient struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
	certPath   string
	keyPath    string
	caPath     string
	devMode    bool
	logger     *logrus.Logger
	cb         *gobreaker.CircuitBreaker
}

// SOAPClientConfig holds the configuration for the SOAP client
type SOAPClientConfig struct {
	BaseURL  string
	Timeout  time.Duration
	CertPath string
	KeyPath  string
	CAPath   string
	DevMode  bool
	Logger   *logrus.Logger
}

// SOAPEnvelope represents a SOAP 1.2 envelope
type SOAPEnvelope struct {
	XMLName xml.Name    `xml:"soap:Envelope"`
	SoapNS  string      `xml:"xmlns:soap,attr"`
	DictNS  string      `xml:"xmlns:dict,attr"`
	Header  *SOAPHeader `xml:"soap:Header,omitempty"`
	Body    SOAPBody    `xml:"soap:Body"`
}

// SOAPHeader represents the SOAP header with security
type SOAPHeader struct {
	Security *SOAPSecurity `xml:"wsse:Security,omitempty"`
}

// SOAPSecurity represents WS-Security header
type SOAPSecurity struct {
	WSSENS    string `xml:"xmlns:wsse,attr"`
	Signature string `xml:"wsse:Signature,omitempty"`
}

// SOAPBody represents the SOAP body
type SOAPBody struct {
	Content interface{} `xml:",innerxml"`
}

// SOAPFault represents a SOAP fault response
type SOAPFault struct {
	XMLName xml.Name `xml:"Fault"`
	Code    string   `xml:"Code>Value"`
	Reason  string   `xml:"Reason>Text"`
	Detail  string   `xml:"Detail,omitempty"`
}

// SOAPFaultResponse represents a SOAP fault envelope
type SOAPFaultResponse struct {
	XMLName xml.Name  `xml:"Envelope"`
	Body    SOAPFBody `xml:"Body"`
}

// SOAPFBody represents SOAP body with fault
type SOAPFBody struct {
	Fault SOAPFault `xml:"Fault"`
}

// NewSOAPClient creates a new SOAP client with mTLS support and circuit breaker
func NewSOAPClient(config *SOAPClientConfig) (*SOAPClient, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	if config.Logger == nil {
		config.Logger = logrus.New()
		config.Logger.SetLevel(logrus.InfoLevel)
	}

	if config.Timeout == 0 {
		config.Timeout = soapRequestTimeout
	}

	// Configure TLS for mTLS
	tlsConfig, err := configureSOAPTLS(config)
	if err != nil {
		return nil, fmt.Errorf("failed to configure TLS: %w", err)
	}

	// Create HTTP transport with connection pooling
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
		DialContext: (&net.Dialer{
			Timeout:   soapConnectionTimeout,
			KeepAlive: soapKeepAlive,
		}).DialContext,
		MaxIdleConns:          soapMaxIdleConns,
		MaxIdleConnsPerHost:   soapMaxIdleConnsPerHost,
		MaxConnsPerHost:       soapMaxConnsPerHost,
		IdleConnTimeout:       soapIdleConnTimeout,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
	}

	httpClient := &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
	}

	// Configure circuit breaker
	cbSettings := gobreaker.Settings{
		Name:        "BacenSOAPClient",
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 5 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			config.Logger.WithFields(logrus.Fields{
				"circuitBreaker": name,
				"from":           from.String(),
				"to":             to.String(),
			}).Warn("Circuit breaker state changed")
		},
	}

	return &SOAPClient{
		baseURL:    config.BaseURL,
		httpClient: httpClient,
		timeout:    config.Timeout,
		certPath:   config.CertPath,
		keyPath:    config.KeyPath,
		caPath:     config.CAPath,
		devMode:    config.DevMode,
		logger:     config.Logger,
		cb:         gobreaker.NewCircuitBreaker(cbSettings),
	}, nil
}

// configureSOAPTLS sets up TLS configuration with mTLS support
func configureSOAPTLS(config *SOAPClientConfig) (*tls.Config, error) {
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

	// Load client certificate and key for mTLS (ICP-Brasil A3)
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

// BuildSOAPEnvelope builds a SOAP envelope with the given body content
func (c *SOAPClient) BuildSOAPEnvelope(bodyXML string, signedXML string) ([]byte, error) {
	envelope := SOAPEnvelope{
		SoapNS: soapEnvelopeNS,
		DictNS: bacenDictNS,
		Body: SOAPBody{
			Content: bodyXML,
		},
	}

	// Add signature to header if provided
	if signedXML != "" {
		envelope.Header = &SOAPHeader{
			Security: &SOAPSecurity{
				WSSENS:    soapSecurityNS,
				Signature: signedXML,
			},
		}
	}

	// Marshal to XML
	xmlData, err := xml.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SOAP envelope: %w", err)
	}

	// Add XML header
	xmlHeader := []byte(xml.Header)
	return append(xmlHeader, xmlData...), nil
}

// SendSOAPRequest sends a SOAP request to Bacen and returns the response
func (c *SOAPClient) SendSOAPRequest(ctx context.Context, endpoint string, soapEnvelope []byte) ([]byte, error) {
	c.logger.WithFields(logrus.Fields{
		"endpoint":     endpoint,
		"envelopeSize": len(soapEnvelope),
	}).Debug("Sending SOAP request")

	// Execute request through circuit breaker
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.doSOAPRequest(ctx, endpoint, soapEnvelope)
	})

	if err != nil {
		c.logger.WithError(err).Error("SOAP request failed")
		return nil, err
	}

	return result.([]byte), nil
}

// doSOAPRequest performs the actual SOAP HTTP request
func (c *SOAPClient) doSOAPRequest(ctx context.Context, endpoint string, soapEnvelope []byte) ([]byte, error) {
	url := c.baseURL + endpoint

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(soapEnvelope))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set SOAP-specific headers
	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	req.Header.Set("Accept", "application/soap+xml")
	req.Header.Set("User-Agent", "LBPay-DICT-Bridge/1.0")

	// Add correlation ID from context if available
	if correlationID := ctx.Value("correlationID"); correlationID != nil {
		req.Header.Set("X-Correlation-ID", correlationID.(string))
	}

	// Log request
	c.logger.WithFields(logrus.Fields{
		"method":       req.Method,
		"url":          url,
		"contentType":  req.Header.Get("Content-Type"),
		"envelopeSize": len(soapEnvelope),
	}).Info("Sending SOAP/HTTPS request to Bacen")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Log response
	c.logger.WithFields(logrus.Fields{
		"statusCode": resp.StatusCode,
		"status":     resp.Status,
		"bodySize":   len(body),
	}).Info("Received SOAP response from Bacen")

	// Check for HTTP errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, c.handleSOAPError(resp.StatusCode, body)
	}

	// Check for SOAP Fault
	if err := c.checkSOAPFault(body); err != nil {
		return nil, err
	}

	return body, nil
}

// ParseSOAPResponse extracts the body content from a SOAP envelope response
func (c *SOAPClient) ParseSOAPResponse(soapResponse []byte) ([]byte, error) {
	var envelope struct {
		XMLName xml.Name `xml:"Envelope"`
		Body    struct {
			Content []byte `xml:",innerxml"`
		} `xml:"Body"`
	}

	if err := xml.Unmarshal(soapResponse, &envelope); err != nil {
		return nil, fmt.Errorf("failed to parse SOAP envelope: %w", err)
	}

	return envelope.Body.Content, nil
}

// checkSOAPFault checks if the response contains a SOAP Fault
func (c *SOAPClient) checkSOAPFault(body []byte) error {
	var faultResp SOAPFaultResponse
	if err := xml.Unmarshal(body, &faultResp); err != nil {
		// Not a fault response, this is ok
		return nil
	}

	// Check if it contains a fault
	if faultResp.Body.Fault.Code != "" {
		c.logger.WithFields(logrus.Fields{
			"faultCode":   faultResp.Body.Fault.Code,
			"faultReason": faultResp.Body.Fault.Reason,
			"faultDetail": faultResp.Body.Fault.Detail,
		}).Error("SOAP Fault received")

		return fmt.Errorf("SOAP Fault: %s - %s (Detail: %s)",
			faultResp.Body.Fault.Code,
			faultResp.Body.Fault.Reason,
			faultResp.Body.Fault.Detail)
	}

	return nil
}

// handleSOAPError processes SOAP error responses
func (c *SOAPClient) handleSOAPError(statusCode int, body []byte) error {
	c.logger.WithFields(logrus.Fields{
		"statusCode": statusCode,
		"body":       string(body),
	}).Error("SOAP request failed with HTTP error")

	return fmt.Errorf("SOAP HTTP error %d: %s", statusCode, string(body))
}

// HealthCheck performs a health check on the SOAP endpoint
func (c *SOAPClient) HealthCheck(ctx context.Context) error {
	c.logger.Debug("Performing SOAP health check")

	// Simple HTTP GET to health endpoint (not SOAP)
	url := c.baseURL + "/health"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	c.logger.Debug("SOAP health check passed")
	return nil
}

// GetCircuitBreakerState returns the current state of the circuit breaker
func (c *SOAPClient) GetCircuitBreakerState() gobreaker.State {
	return c.cb.State()
}

// GetBaseURL returns the base URL being used
func (c *SOAPClient) GetBaseURL() string {
	return c.baseURL
}
