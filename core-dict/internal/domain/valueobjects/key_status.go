package valueobjects

import "errors"

// KeyStatus representa o status de uma chave PIX no DICT
type KeyStatus string

const (
	KeyStatusPending              KeyStatus = "PENDING"                // Aguardando confirmação Bacen
	KeyStatusActive               KeyStatus = "ACTIVE"                 // Ativa e utilizável
	KeyStatusBlocked              KeyStatus = "BLOCKED"                // Bloqueada temporariamente
	KeyStatusDeleted              KeyStatus = "DELETED"                // Deletada (soft delete)
	KeyStatusClaimPending         KeyStatus = "CLAIM_PENDING"          // Com reivindicação pendente
	KeyStatusPortabilityRequested KeyStatus = "PORTABILITY_REQUESTED"  // Portabilidade em andamento
	KeyStatusOwnershipConfirmed   KeyStatus = "OWNERSHIP_CONFIRMED"    // Propriedade confirmada
	KeyStatusFailed               KeyStatus = "FAILED"                 // Falha no registro
)

var validKeyStatuses = map[KeyStatus]bool{
	KeyStatusPending:              true,
	KeyStatusActive:               true,
	KeyStatusBlocked:              true,
	KeyStatusDeleted:              true,
	KeyStatusClaimPending:         true,
	KeyStatusPortabilityRequested: true,
	KeyStatusOwnershipConfirmed:   true,
	KeyStatusFailed:               true,
}

// NewKeyStatus cria e valida um status de chave
func NewKeyStatus(s string) (KeyStatus, error) {
	ks := KeyStatus(s)
	if !validKeyStatuses[ks] {
		return "", errors.New("invalid key status")
	}
	return ks, nil
}

// String retorna a representação string do status
func (ks KeyStatus) String() string {
	return string(ks)
}

// IsValid verifica se o status é válido
func (ks KeyStatus) IsValid() bool {
	return validKeyStatuses[ks]
}

// CanTransitionTo verifica se é possível transitar para outro status
func (ks KeyStatus) CanTransitionTo(newStatus KeyStatus) bool {
	validTransitions := map[KeyStatus][]KeyStatus{
		KeyStatusPending: {
			KeyStatusActive,
			KeyStatusFailed,
			KeyStatusDeleted,
		},
		KeyStatusActive: {
			KeyStatusBlocked,
			KeyStatusDeleted,
			KeyStatusClaimPending,
			KeyStatusPortabilityRequested,
		},
		KeyStatusBlocked: {
			KeyStatusActive,
			KeyStatusDeleted,
		},
		KeyStatusClaimPending: {
			KeyStatusActive,
			KeyStatusOwnershipConfirmed,
			KeyStatusDeleted,
		},
		KeyStatusPortabilityRequested: {
			KeyStatusActive,
			KeyStatusDeleted,
		},
		KeyStatusOwnershipConfirmed: {
			KeyStatusActive,
		},
	}

	allowedTransitions, exists := validTransitions[ks]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == newStatus {
			return true
		}
	}
	return false
}
