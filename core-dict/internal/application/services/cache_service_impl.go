package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// CacheServiceImpl implementação do CacheService com Redis
type CacheServiceImpl struct {
	redisClient RedisClient
}

// RedisClient interface para operações Redis
type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	DelPattern(ctx context.Context, pattern string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
}

// NewCacheServiceImpl cria nova instância
func NewCacheServiceImpl(redisClient RedisClient) *CacheServiceImpl {
	return &CacheServiceImpl{
		redisClient: redisClient,
	}
}

// CacheTTL define TTLs padrão por tipo de cache
var CacheTTL = map[string]time.Duration{
	"entry":      5 * time.Minute,  // Entry (chave PIX)
	"account":    10 * time.Minute, // Conta CID
	"claim":      2 * time.Minute,  // Claim (alta volatilidade)
	"statistics": 1 * time.Minute,  // Estatísticas agregadas
	"metadata":   30 * time.Minute, // Metadata (baixa volatilidade)
}

// Get busca valor no cache
func (s *CacheServiceImpl) Get(ctx context.Context, key string) (interface{}, error) {
	value, err := s.redisClient.Get(ctx, key)
	if err != nil {
		if err.Error() == "redis: nil" {
			return nil, errors.New("cache miss")
		}
		return nil, errors.New("cache get error: " + err.Error())
	}

	// Deserializar JSON
	var result interface{}
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return nil, errors.New("cache deserialize error: " + err.Error())
	}

	return result, nil
}

// Set armazena valor no cache com TTL
func (s *CacheServiceImpl) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Serializar para JSON
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return errors.New("cache serialize error: " + err.Error())
	}

	// Armazenar no Redis
	if err := s.redisClient.Set(ctx, key, string(valueJSON), ttl); err != nil {
		return errors.New("cache set error: " + err.Error())
	}

	return nil
}

// Delete remove valor do cache
func (s *CacheServiceImpl) Delete(ctx context.Context, key string) error {
	if err := s.redisClient.Del(ctx, key); err != nil {
		return errors.New("cache delete error: " + err.Error())
	}
	return nil
}

// Exists verifica se chave existe
func (s *CacheServiceImpl) Exists(ctx context.Context, key string) (bool, error) {
	exists, err := s.redisClient.Exists(ctx, key)
	if err != nil {
		return false, errors.New("cache exists error: " + err.Error())
	}
	return exists, nil
}

// Invalidate invalida padrão de chaves
func (s *CacheServiceImpl) Invalidate(ctx context.Context, pattern string) error {
	if err := s.redisClient.DelPattern(ctx, pattern); err != nil {
		return errors.New("cache invalidate error: " + err.Error())
	}
	return nil
}

// InvalidateKey invalida chave específica (alias para Delete)
func (s *CacheServiceImpl) InvalidateKey(ctx context.Context, key string) error {
	return s.Delete(ctx, key)
}

// InvalidatePattern invalida padrão (alias para Invalidate)
func (s *CacheServiceImpl) InvalidatePattern(ctx context.Context, pattern string) error {
	return s.Invalidate(ctx, pattern)
}

// --- Estratégias de Cache ---

// 1. Cache-Aside (Lazy Loading)
func (s *CacheServiceImpl) GetOrLoad(ctx context.Context, key string, loader func() (interface{}, error), ttl time.Duration) (interface{}, error) {
	// 1. Tentar buscar no cache
	cached, err := s.Get(ctx, key)
	if err == nil {
		return cached, nil // Cache hit
	}

	// 2. Cache miss - buscar do source (DB)
	value, err := loader()
	if err != nil {
		return nil, errors.New("loader error: " + err.Error())
	}

	// 3. Armazenar no cache
	if err := s.Set(ctx, key, value, ttl); err != nil {
		fmt.Printf("Warning: failed to cache key %s: %v\n", key, err)
	}

	return value, nil
}

// 2. Write-Through Cache
func (s *CacheServiceImpl) WriteThrough(ctx context.Context, key string, value interface{}, ttl time.Duration, writer func(interface{}) error) error {
	// 1. Escrever no DB primeiro
	if err := writer(value); err != nil {
		return errors.New("db write error: " + err.Error())
	}

	// 2. Escrever no cache
	if err := s.Set(ctx, key, value, ttl); err != nil {
		fmt.Printf("Warning: failed to write-through cache key %s: %v\n", key, err)
	}

	return nil
}

// GetTTL retorna TTL padrão baseado no tipo
func (s *CacheServiceImpl) GetTTL(cacheType string) time.Duration {
	if ttl, ok := CacheTTL[cacheType]; ok {
		return ttl
	}
	return 5 * time.Minute // Default
}

// WarmUp pré-popula cache
func (s *CacheServiceImpl) WarmUp(ctx context.Context, keys []string, loader func(string) (interface{}, error)) error {
	for _, key := range keys {
		value, err := loader(key)
		if err != nil {
			fmt.Printf("Warning: failed to warm up cache key %s: %v\n", key, err)
			continue
		}

		ttl := s.GetTTL("entry")
		if err := s.Set(ctx, key, value, ttl); err != nil {
			fmt.Printf("Warning: failed to set warm up cache key %s: %v\n", key, err)
		}
	}
	return nil
}

// Clear limpa todo o cache
func (s *CacheServiceImpl) Clear(ctx context.Context) error {
	patterns := []string{
		"entry:*",
		"account:*",
		"claim:*",
		"statistics:*",
	}

	for _, pattern := range patterns {
		if err := s.InvalidatePattern(ctx, pattern); err != nil {
			return errors.New("failed to clear cache pattern " + pattern + ": " + err.Error())
		}
	}

	return nil
}
