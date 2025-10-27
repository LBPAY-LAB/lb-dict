package adapters

import (
	"context"
	"github.com/lbpay-lab/core-dict/internal/application/services"
)

// ConnectServiceAdapter adapts ConnectClient to ConnectService interface
// This is a temporary adapter until ConnectClient is fully implemented
type ConnectServiceAdapter struct {
	client services.ConnectClient
}

// NewConnectServiceAdapter creates a new adapter
func NewConnectServiceAdapter(client services.ConnectClient) services.ConnectService {
	return &ConnectServiceAdapter{
		client: client,
	}
}

// VerifyAccount implements ConnectService.VerifyAccount
// For now, it returns true (optimistic) since ConnectClient doesn't have this method yet
func (a *ConnectServiceAdapter) VerifyAccount(ctx context.Context, ispb, branch, accountNumber string) (bool, error) {
	// If client is nil, assume account is valid (optimistic verification)
	if a.client == nil {
		return true, nil
	}

	// Check if Connect service is reachable
	if err := a.client.HealthCheck(ctx); err != nil {
		// If Connect is down, still allow operation (degraded mode)
		return true, nil
	}

	// TODO: When ConnectClient has VerifyAccount method, use it here
	// For now, assume account is valid if Connect is reachable
	return true, nil
}
