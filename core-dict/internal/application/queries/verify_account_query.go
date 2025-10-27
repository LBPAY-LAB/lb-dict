package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lbpay-lab/core-dict/internal/application/services"
	"github.com/lbpay-lab/core-dict/internal/domain/repositories"
)

// VerifyAccountQuery representa a query para verificar uma conta no RSFN
type VerifyAccountQuery struct {
	ISPB          string
	Branch        string
	AccountNumber string
}

// VerifyAccountResult representa o resultado da verificação
type VerifyAccountResult struct {
	Valid  bool   `json:"valid"`
	Source string `json:"source"` // "local" ou "rsfn"
	Reason string `json:"reason,omitempty"`
}

// VerifyAccountQueryHandler lida com a query VerifyAccount
type VerifyAccountQueryHandler struct {
	accountRepo    repositories.AccountRepository
	connectService services.ConnectService
	cache          services.CacheService
}

// NewVerifyAccountQueryHandler cria um novo handler para VerifyAccount
func NewVerifyAccountQueryHandler(
	accountRepo repositories.AccountRepository,
	connectService services.ConnectService,
	cache services.CacheService,
) *VerifyAccountQueryHandler {
	return &VerifyAccountQueryHandler{
		accountRepo:    accountRepo,
		connectService: connectService,
		cache:          cache,
	}
}

// Handle executa a query VerifyAccount
// 1. Tenta verificar localmente (cache + database)
// 2. Se não encontrar, chama RSFN via Connect service
func (h *VerifyAccountQueryHandler) Handle(ctx context.Context, query VerifyAccountQuery) (*VerifyAccountResult, error) {
	// Validação
	if query.ISPB == "" || query.Branch == "" || query.AccountNumber == "" {
		return nil, fmt.Errorf("ispb, branch, and account_number are required")
	}

	// 1. Try cache first (verification results cache por 10 minutos)
	cacheKey := fmt.Sprintf("verify:account:%s:%s:%s", query.ISPB, query.Branch, query.AccountNumber)
	if cachedData, err := h.cache.Get(ctx, cacheKey); err == nil && cachedData != nil {
		// Cache hit
		if result, ok := cachedData.(*VerifyAccountResult); ok {
			return result, nil
		}
		// Try to unmarshal
		if jsonData, ok := cachedData.(string); ok {
			var result VerifyAccountResult
			if err := json.Unmarshal([]byte(jsonData), &result); err == nil {
				return &result, nil
			}
		}
	}

	// 2. Try local database (verificação rápida)
	valid, err := h.accountRepo.ExistsByAccountNumber(ctx, query.ISPB, query.Branch, query.AccountNumber)
	if err == nil && valid {
		result := &VerifyAccountResult{
			Valid:  true,
			Source: "local",
			Reason: "Account found in local database",
		}

		// Cache result (TTL: 10 minutes)
		_ = h.cache.Set(ctx, cacheKey, result, 10*time.Minute)

		return result, nil
	}

	// 3. Not found locally - call RSFN via Connect service
	// This triggers: Connect → Bridge → Bacen SOAP API
	// Expected latency: ~500ms (includes network + Bacen processing)
	valid, err = h.connectService.VerifyAccount(ctx, query.ISPB, query.Branch, query.AccountNumber)
	if err != nil {
		// RSFN unavailable or error - this is NOT a hard failure
		// We return false but indicate the verification source
		result := &VerifyAccountResult{
			Valid:  false,
			Source: "rsfn",
			Reason: fmt.Sprintf("RSFN verification failed: %v", err),
		}
		// Cache negative result for shorter TTL (1 minute) to allow retry
		_ = h.cache.Set(ctx, cacheKey, result, 1*time.Minute)
		return result, nil
	}

	result := &VerifyAccountResult{
		Valid:  valid,
		Source: "rsfn",
		Reason: "Verified via RSFN Connect",
	}

	// 4. Cache result (TTL: 10 minutes for positive, 1 minute for negative)
	cacheTTL := 10 * time.Minute
	if !valid {
		cacheTTL = 1 * time.Minute // Shorter TTL for negative results
	}
	if err := h.cache.Set(ctx, cacheKey, result, cacheTTL); err != nil {
		// Log cache error but don't fail the request
		_ = err
	}

	return result, nil
}

// InvalidateCache invalida o cache de verificação de uma conta
func (h *VerifyAccountQueryHandler) InvalidateCache(ctx context.Context, ispb, branch, accountNumber string) error {
	cacheKey := fmt.Sprintf("verify:account:%s:%s:%s", ispb, branch, accountNumber)
	return h.cache.Delete(ctx, cacheKey)
}
