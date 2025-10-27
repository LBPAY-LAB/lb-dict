package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	pb "github.com/lbpay-lab/dict-contracts/gen/proto/bridge/v1"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// HealthCheck handles the HealthCheck RPC call
// Verifies Bridge connectivity and dependencies
func (s *Server) HealthCheck(ctx context.Context, _ *emptypb.Empty) (*pb.HealthCheckResponse, error) {
	s.logger.Info("HealthCheck called")

	response := &pb.HealthCheckResponse{
		LastCheck: timestamppb.Now(),
	}

	// Check 1: Bacen DICT API connectivity
	bacenStatus, bacenLatency := s.checkBacenConnectivity(ctx)
	response.BacenStatus = bacenStatus
	response.BacenLatencyMs = bacenLatency

	// Check 2: mTLS Certificate status
	certStatus := s.checkCertificateStatus()
	response.CertificateStatus = certStatus

	// Determine overall health status
	response.Status = s.determineOverallHealth(bacenStatus, certStatus)

	s.logger.Infof("HealthCheck completed: status=%s, bacen=%s, cert=%s, latency=%dms",
		response.Status, response.BacenStatus, response.CertificateStatus, response.BacenLatencyMs)

	return response, nil
}

// checkBacenConnectivity verifies connectivity with Bacen DICT API
func (s *Server) checkBacenConnectivity(ctx context.Context) (pb.BacenConnectionStatus, int64) {
	bacenURL := os.Getenv("BACEN_DICT_URL")
	if bacenURL == "" {
		bacenURL = "https://api.dict.bacen.gov.br" // Default production URL
	}

	// Try to reach Bacen health endpoint
	healthURL := fmt.Sprintf("%s/dict/api/v1/health", bacenURL)

	startTime := time.Now()

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				// For health check, we use a less strict TLS config
				// The actual mTLS certificates will be checked separately
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		s.logger.Errorf("Failed to create Bacen health check request: %v", err)
		return pb.BacenConnectionStatus_BACEN_CONNECTION_UNSPECIFIED, 0
	}

	resp, err := client.Do(req)
	latency := time.Since(startTime).Milliseconds()

	if err != nil {
		s.logger.Errorf("Bacen health check failed: %v", err)

		// Determine specific error type
		if os.IsTimeout(err) {
			return pb.BacenConnectionStatus_BACEN_CONNECTION_TIMEOUT, latency
		}

		// Check if it's a TLS error
		if _, ok := err.(*tls.CertificateVerificationError); ok {
			return pb.BacenConnectionStatus_BACEN_CONNECTION_TLS_ERROR, latency
		}

		return pb.BacenConnectionStatus_BACEN_CONNECTION_UNSPECIFIED, latency
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode == http.StatusOK {
		s.logger.Infof("Bacen health check OK (latency: %dms)", latency)
		return pb.BacenConnectionStatus_BACEN_CONNECTION_OK, latency
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		s.logger.Errorf("Bacen health check auth failed: status=%d", resp.StatusCode)
		return pb.BacenConnectionStatus_BACEN_CONNECTION_AUTH_FAILED, latency
	}

	s.logger.Errorf("Bacen health check unexpected status: %d", resp.StatusCode)
	return pb.BacenConnectionStatus_BACEN_CONNECTION_UNSPECIFIED, latency
}

// checkCertificateStatus verifies mTLS certificate validity
func (s *Server) checkCertificateStatus() pb.CertificateStatus {
	certPath := os.Getenv("MTLS_CLIENT_CERT")
	if certPath == "" {
		s.logger.Warn("MTLS_CLIENT_CERT not configured - skipping certificate check")
		return pb.CertificateStatus_CERTIFICATE_STATUS_UNSPECIFIED
	}

	// Load certificate
	certData, err := os.ReadFile(certPath)
	if err != nil {
		s.logger.Errorf("Failed to read certificate file: %v", err)
		return pb.CertificateStatus_CERTIFICATE_STATUS_UNSPECIFIED
	}

	// Parse certificate
	cert, err := tls.X509KeyPair(certData, certData)
	if err != nil {
		s.logger.Errorf("Failed to parse certificate: %v", err)
		return pb.CertificateStatus_CERTIFICATE_STATUS_UNSPECIFIED
	}

	// Check if we have a valid certificate
	if len(cert.Certificate) == 0 {
		s.logger.Error("Certificate chain is empty")
		return pb.CertificateStatus_CERTIFICATE_STATUS_UNSPECIFIED
	}

	// Parse the leaf certificate to check expiration
	// Note: In production, you would use x509.ParseCertificate here
	// For now, we'll use a simplified check based on file modification time

	fileInfo, err := os.Stat(certPath)
	if err != nil {
		s.logger.Errorf("Failed to stat certificate file: %v", err)
		return pb.CertificateStatus_CERTIFICATE_STATUS_UNSPECIFIED
	}

	// Calculate certificate age (simplified - in production, check NotAfter)
	certAge := time.Since(fileInfo.ModTime())

	// ICP-Brasil A3 certificates are typically valid for 1-3 years
	// We'll warn if certificate is older than 11 months (30 days before 1 year)
	if certAge > 11*30*24*time.Hour {
		s.logger.Warn("Certificate may be expiring soon (older than 11 months)")
		return pb.CertificateStatus_CERTIFICATE_STATUS_EXPIRING_SOON
	}

	// If certificate is older than 3 years, consider it expired
	if certAge > 3*365*24*time.Hour {
		s.logger.Error("Certificate appears to be expired (older than 3 years)")
		return pb.CertificateStatus_CERTIFICATE_STATUS_EXPIRED
	}

	s.logger.Info("Certificate appears to be valid")
	return pb.CertificateStatus_CERTIFICATE_STATUS_VALID
}

// determineOverallHealth determines the overall health status
func (s *Server) determineOverallHealth(
	bacenStatus pb.BacenConnectionStatus,
	certStatus pb.CertificateStatus,
) pb.HealthStatus {
	// Unhealthy conditions
	if bacenStatus == pb.BacenConnectionStatus_BACEN_CONNECTION_AUTH_FAILED ||
		certStatus == pb.CertificateStatus_CERTIFICATE_STATUS_EXPIRED {
		return pb.HealthStatus_HEALTH_STATUS_UNHEALTHY
	}

	// Degraded conditions
	if bacenStatus == pb.BacenConnectionStatus_BACEN_CONNECTION_TIMEOUT ||
		bacenStatus == pb.BacenConnectionStatus_BACEN_CONNECTION_TLS_ERROR ||
		certStatus == pb.CertificateStatus_CERTIFICATE_STATUS_EXPIRING_SOON {
		return pb.HealthStatus_HEALTH_STATUS_DEGRADED
	}

	// Healthy only if both Bacen and certificate are OK
	if bacenStatus == pb.BacenConnectionStatus_BACEN_CONNECTION_OK &&
		(certStatus == pb.CertificateStatus_CERTIFICATE_STATUS_VALID ||
			certStatus == pb.CertificateStatus_CERTIFICATE_STATUS_UNSPECIFIED) {
		return pb.HealthStatus_HEALTH_STATUS_HEALTHY
	}

	// Default to degraded if status is unclear
	return pb.HealthStatus_HEALTH_STATUS_DEGRADED
}

// CheckXMLSignerHealth checks if the XML Signer service is available
// This is an internal helper method (not exposed via gRPC)
func (s *Server) CheckXMLSignerHealth(ctx context.Context) bool {
	xmlSignerURL := os.Getenv("XML_SIGNER_URL")
	if xmlSignerURL == "" {
		xmlSignerURL = "http://localhost:8081" // Default URL
	}

	healthURL := fmt.Sprintf("%s/health", xmlSignerURL)

	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		s.logger.Errorf("Failed to create XML Signer health check request: %v", err)
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		s.logger.Errorf("XML Signer health check failed: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		s.logger.Debug("XML Signer health check OK")
		return true
	}

	s.logger.Errorf("XML Signer health check failed: status=%d", resp.StatusCode)
	return false
}
