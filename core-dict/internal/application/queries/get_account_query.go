package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lbpay-lab/core-dict/internal/application/services"
	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/domain/repositories"
)

// GetAccountQuery representa a query para buscar uma conta CID
type GetAccountQuery struct {
	AccountID uuid.UUID
}

// GetAccountByNumberQuery representa a query para buscar conta por ISPB + agência + número
type GetAccountByNumberQuery struct {
	ISPB          string
	Branch        string
	AccountNumber string
}

// GetAccountQueryHandler lida com queries de contas
type GetAccountQueryHandler struct {
	accountRepo repositories.AccountRepository
	cache       services.CacheService
}

// NewGetAccountQueryHandler cria um novo handler para GetAccount
func NewGetAccountQueryHandler(
	accountRepo repositories.AccountRepository,
	cache services.CacheService,
) *GetAccountQueryHandler {
	return &GetAccountQueryHandler{
		accountRepo: accountRepo,
		cache:       cache,
	}
}

// Handle executa a query GetAccount por ID
func (h *GetAccountQueryHandler) Handle(ctx context.Context, query GetAccountQuery) (*entities.Account, error) {
	// Validação
	if query.AccountID == uuid.Nil {
		return nil, fmt.Errorf("account_id is required")
	}

	// 1. Try cache first
	cacheKey := fmt.Sprintf("account:id:%s", query.AccountID.String())
	if cachedData, err := h.cache.Get(ctx, cacheKey); err == nil && cachedData != nil {
		// Cache hit
		if account, ok := cachedData.(*entities.Account); ok {
			return account, nil
		}
		// Try to unmarshal
		if jsonData, ok := cachedData.(string); ok {
			var account entities.Account
			if err := json.Unmarshal([]byte(jsonData), &account); err == nil {
				return &account, nil
			}
		}
	}

	// 2. Cache miss - query database
	account, err := h.accountRepo.FindByID(ctx, query.AccountID)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	// 3. Store in cache (TTL: 5 minutes)
	if err := h.cache.Set(ctx, cacheKey, account, 5*time.Minute); err != nil {
		_ = err
	}

	return account, nil
}

// HandleByNumber executa a query GetAccount por ISPB + agência + número
func (h *GetAccountQueryHandler) HandleByNumber(ctx context.Context, query GetAccountByNumberQuery) (*entities.Account, error) {
	// Validação
	if query.ISPB == "" || query.Branch == "" || query.AccountNumber == "" {
		return nil, fmt.Errorf("ispb, branch, and account_number are required")
	}

	// 1. Try cache first
	cacheKey := fmt.Sprintf("account:number:%s:%s:%s", query.ISPB, query.Branch, query.AccountNumber)
	if cachedData, err := h.cache.Get(ctx, cacheKey); err == nil && cachedData != nil {
		// Cache hit
		if account, ok := cachedData.(*entities.Account); ok {
			return account, nil
		}
		// Try to unmarshal
		if jsonData, ok := cachedData.(string); ok {
			var account entities.Account
			if err := json.Unmarshal([]byte(jsonData), &account); err == nil {
				return &account, nil
			}
		}
	}

	// 2. Cache miss - query database
	account, err := h.accountRepo.FindByAccountNumber(ctx, query.ISPB, query.Branch, query.AccountNumber)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	// 3. Store in cache (TTL: 5 minutes)
	// Cache by both ID and account number
	cacheKeyByID := fmt.Sprintf("account:id:%s", account.ID.String())
	_ = h.cache.Set(ctx, cacheKey, account, 5*time.Minute)
	_ = h.cache.Set(ctx, cacheKeyByID, account, 5*time.Minute)

	return account, nil
}

// InvalidateCache invalida o cache de uma conta
func (h *GetAccountQueryHandler) InvalidateCache(ctx context.Context, accountID uuid.UUID) error {
	pattern := fmt.Sprintf("account:*")
	return h.cache.Invalidate(ctx, pattern)
}
