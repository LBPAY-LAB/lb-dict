package entities

import (
	"time"

	"github.com/google/uuid"
)

// KeyType representa o tipo de chave PIX
type KeyType string

const (
	KeyTypeCPF   KeyType = "CPF"
	KeyTypeCNPJ  KeyType = "CNPJ"
	KeyTypeEmail KeyType = "EMAIL"
	KeyTypePhone KeyType = "PHONE"
	KeyTypeEVP   KeyType = "EVP"
)

// KeyStatus representa o status de uma chave PIX
type KeyStatus string

const (
	KeyStatusPending KeyStatus = "PENDING"
	KeyStatusActive  KeyStatus = "ACTIVE"
	KeyStatusBlocked KeyStatus = "BLOCKED"
	KeyStatusDeleted KeyStatus = "DELETED"
)

// Entry representa uma chave PIX no sistema DICT
type Entry struct {
	ID            uuid.UUID  `json:"id"`
	KeyType       KeyType    `json:"key_type"`
	KeyValue      string     `json:"key_value"`
	Status        KeyStatus  `json:"status"`
	AccountID     uuid.UUID  `json:"account_id"`
	ISPB          string     `json:"ispb"`
	Branch        string     `json:"branch"`
	AccountNumber string     `json:"account_number"`
	AccountType   string     `json:"account_type"`
	OwnerName     string     `json:"owner_name"`
	OwnerTaxID    string     `json:"owner_tax_id"`
	OwnerType     string     `json:"owner_type"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}

// Note: Account, Claim, Infraction, and AuditEvent
// are defined in their respective files (account.go, claim.go, infraction.go, audit_event.go)

// AuditLog is an alias for AuditEvent for backwards compatibility
type AuditLog struct {
	ID         uuid.UUID
	EntityType string
	EntityID   uuid.UUID
	Action     string
	ActorID    uuid.UUID
	ActorType  string
	Changes    map[string]interface{}
	Metadata   map[string]interface{}
	Timestamp  time.Time
}

// Statistics representa estatísticas agregadas do sistema
type Statistics struct {
	TotalKeys             int64            `json:"total_keys"`
	ActiveKeys            int64            `json:"active_keys"`
	BlockedKeys           int64            `json:"blocked_keys"`
	DeletedKeys           int64            `json:"deleted_keys"`
	TotalClaims           int64            `json:"total_claims"`
	PendingClaims         int64            `json:"pending_claims"`
	CompletedClaims       int64            `json:"completed_claims"`
	KeysByType            map[string]int64 `json:"keys_by_type"`
	ClaimsByType          map[string]int64 `json:"claims_by_type"`
	TotalInfractions      int64            `json:"total_infractions"`
	InfractionsBySeverity map[string]int64 `json:"infractions_by_severity"`
	LastUpdated           time.Time        `json:"last_updated"`
}

// HealthStatus representa o status de saúde do sistema
type HealthStatus struct {
	Status         string                 `json:"status"`
	Version        string                 `json:"version"`
	Uptime         int64                  `json:"uptime"`
	DatabaseStatus string                 `json:"database_status"`
	RedisStatus    string                 `json:"redis_status"`
	PulsarStatus   string                 `json:"pulsar_status"`
	Dependencies   map[string]interface{} `json:"dependencies"`
	Timestamp      time.Time              `json:"timestamp"`
}
