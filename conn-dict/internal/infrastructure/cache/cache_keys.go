package cache

import (
	"fmt"
	"time"
)

// CacheKeyPattern defines standard key patterns and TTLs for different data types
type CacheKeyPattern struct {
	Pattern  string
	TTL      time.Duration
	Strategy CacheStrategy
}

// Standard cache key patterns and their configurations
var (
	// Entry lookups - Cache-Aside
	EntryKeyPattern = CacheKeyPattern{
		Pattern:  "entry:%s",
		TTL:      5 * time.Minute,
		Strategy: StrategyCacheAside,
	}

	// Entry creation - Write-Through
	EntryCreatePattern = CacheKeyPattern{
		Pattern:  "entry:%s",
		TTL:      5 * time.Minute,
		Strategy: StrategyWriteThrough,
	}

	// Metrics and counters - Write-Behind
	MetricsPattern = CacheKeyPattern{
		Pattern:  "metrics:%s",
		TTL:      10 * time.Minute,
		Strategy: StrategyWriteBehind,
	}

	// Configuration data - Read-Through
	ConfigPattern = CacheKeyPattern{
		Pattern:  "config:%s",
		TTL:      1 * time.Hour,
		Strategy: StrategyReadThrough,
	}

	// Bulk operations - Write-Around
	BulkOpPattern = CacheKeyPattern{
		Pattern:  "bulk:%s",
		TTL:      0, // No TTL, invalidation-based
		Strategy: StrategyWriteAround,
	}

	// Participant data
	ParticipantPattern = CacheKeyPattern{
		Pattern:  "participant:%s",
		TTL:      15 * time.Minute,
		Strategy: StrategyCacheAside,
	}

	// Dictionary entries by key
	DictEntryByKeyPattern = CacheKeyPattern{
		Pattern:  "dict:key:%s",
		TTL:      5 * time.Minute,
		Strategy: StrategyCacheAside,
	}

	// Dictionary entries by participant ISPB
	DictEntryByISPBPattern = CacheKeyPattern{
		Pattern:  "dict:ispb:%s",
		TTL:      5 * time.Minute,
		Strategy: StrategyCacheAside,
	}

	// VSYNC bulk operations
	VSyncPattern = CacheKeyPattern{
		Pattern:  "vsync:%s",
		TTL:      0,
		Strategy: StrategyWriteAround,
	}
)

// CacheKeyBuilder provides methods to build cache keys
type CacheKeyBuilder struct{}

// NewCacheKeyBuilder creates a new cache key builder
func NewCacheKeyBuilder() *CacheKeyBuilder {
	return &CacheKeyBuilder{}
}

// BuildEntryKey builds a cache key for an entry lookup
func (ckb *CacheKeyBuilder) BuildEntryKey(keyID string) string {
	return fmt.Sprintf(EntryKeyPattern.Pattern, keyID)
}

// BuildMetricsKey builds a cache key for metrics
func (ckb *CacheKeyBuilder) BuildMetricsKey(participantISPB string) string {
	return fmt.Sprintf(MetricsPattern.Pattern, participantISPB)
}

// BuildConfigKey builds a cache key for configuration
func (ckb *CacheKeyBuilder) BuildConfigKey(configName string) string {
	return fmt.Sprintf(ConfigPattern.Pattern, configName)
}

// BuildParticipantKey builds a cache key for participant data
func (ckb *CacheKeyBuilder) BuildParticipantKey(participantISPB string) string {
	return fmt.Sprintf(ParticipantPattern.Pattern, participantISPB)
}

// BuildDictEntryByKeyKey builds a cache key for dictionary entry by key
func (ckb *CacheKeyBuilder) BuildDictEntryByKeyKey(key string) string {
	return fmt.Sprintf(DictEntryByKeyPattern.Pattern, key)
}

// BuildDictEntryByISPBKey builds a cache key for dictionary entries by ISPB
func (ckb *CacheKeyBuilder) BuildDictEntryByISPBKey(ispb string) string {
	return fmt.Sprintf(DictEntryByISPBPattern.Pattern, ispb)
}

// BuildVSyncKey builds a cache key for VSYNC operations
func (ckb *CacheKeyBuilder) BuildVSyncKey(syncID string) string {
	return fmt.Sprintf(VSyncPattern.Pattern, syncID)
}

// GetTTLForPattern returns the TTL for a specific pattern
func (ckb *CacheKeyBuilder) GetTTLForPattern(pattern CacheKeyPattern) time.Duration {
	return pattern.TTL
}

// GetStrategyForPattern returns the recommended strategy for a pattern
func (ckb *CacheKeyBuilder) GetStrategyForPattern(pattern CacheKeyPattern) CacheStrategy {
	return pattern.Strategy
}

// InvalidationPatterns defines patterns for bulk cache invalidation
type InvalidationPatterns struct {
	// Invalidate all entries for a participant
	ParticipantEntries string // "dict:ispb:{ispb}:*"

	// Invalidate all metrics for a participant
	ParticipantMetrics string // "metrics:{ispb}:*"

	// Invalidate all config
	AllConfig string // "config:*"

	// Invalidate all entries
	AllEntries string // "entry:*"
}

// GetInvalidationPatterns returns standard invalidation patterns
func GetInvalidationPatterns() *InvalidationPatterns {
	return &InvalidationPatterns{
		ParticipantEntries: "dict:ispb:%s:*",
		ParticipantMetrics: "metrics:%s:*",
		AllConfig:          "config:*",
		AllEntries:         "entry:*",
	}
}

// BuildInvalidationPattern builds an invalidation pattern
func (ip *InvalidationPatterns) BuildParticipantEntriesPattern(ispb string) string {
	return fmt.Sprintf(ip.ParticipantEntries, ispb)
}

// BuildParticipantMetricsPattern builds a participant metrics invalidation pattern
func (ip *InvalidationPatterns) BuildParticipantMetricsPattern(ispb string) string {
	return fmt.Sprintf(ip.ParticipantMetrics, ispb)
}

// CacheKeyHelper provides utility functions for cache key operations
type CacheKeyHelper struct {
	builder *CacheKeyBuilder
}

// NewCacheKeyHelper creates a new cache key helper
func NewCacheKeyHelper() *CacheKeyHelper {
	return &CacheKeyHelper{
		builder: NewCacheKeyBuilder(),
	}
}

// ExtractKeyFromPattern extracts the dynamic part from a cache key
// For example: "entry:123" -> "123"
func (ckh *CacheKeyHelper) ExtractKeyFromPattern(cacheKey, pattern string) (string, error) {
	var extracted string
	_, err := fmt.Sscanf(cacheKey, pattern, &extracted)
	if err != nil {
		return "", fmt.Errorf("failed to extract key: %w", err)
	}
	return extracted, nil
}

// IsValidKey validates if a key matches expected format
func (ckh *CacheKeyHelper) IsValidKey(key string) bool {
	return len(key) > 0 && len(key) < 256 // Redis key limit is typically 512MB, but we use 256 chars
}
