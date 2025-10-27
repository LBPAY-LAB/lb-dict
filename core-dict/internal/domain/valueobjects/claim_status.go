package valueobjects

import "errors"

// ClaimStatus representa o status de uma reivindicação de chave PIX
type ClaimStatus string

const (
	ClaimStatusOpen              ClaimStatus = "OPEN"               // Reivindicação aberta
	ClaimStatusWaitingResolution ClaimStatus = "WAITING_RESOLUTION" // Aguardando resolução
	ClaimStatusConfirmed         ClaimStatus = "CONFIRMED"          // Confirmada pelo doador
	ClaimStatusCancelled         ClaimStatus = "CANCELLED"          // Cancelada
	ClaimStatusCompleted         ClaimStatus = "COMPLETED"          // Completada com sucesso
	ClaimStatusExpired           ClaimStatus = "EXPIRED"            // Expirada (30 dias)
	ClaimStatusAutoConfirmed     ClaimStatus = "AUTO_CONFIRMED"     // Auto-confirmada por timeout
)

var validClaimStatuses = map[ClaimStatus]bool{
	ClaimStatusOpen:              true,
	ClaimStatusWaitingResolution: true,
	ClaimStatusConfirmed:         true,
	ClaimStatusCancelled:         true,
	ClaimStatusCompleted:         true,
	ClaimStatusExpired:           true,
	ClaimStatusAutoConfirmed:     true,
}

// NewClaimStatus cria e valida um status de reivindicação
func NewClaimStatus(s string) (ClaimStatus, error) {
	cs := ClaimStatus(s)
	if !validClaimStatuses[cs] {
		return "", errors.New("invalid claim status")
	}
	return cs, nil
}

// String retorna a representação string do status
func (cs ClaimStatus) String() string {
	return string(cs)
}

// IsValid verifica se o status é válido
func (cs ClaimStatus) IsValid() bool {
	return validClaimStatuses[cs]
}

// IsFinal indica se o status é um estado final
func (cs ClaimStatus) IsFinal() bool {
	finalStatuses := map[ClaimStatus]bool{
		ClaimStatusCompleted:     true,
		ClaimStatusCancelled:     true,
		ClaimStatusExpired:       true,
		ClaimStatusAutoConfirmed: true,
	}
	return finalStatuses[cs]
}

// CanTransitionTo verifica se é possível transitar para outro status
func (cs ClaimStatus) CanTransitionTo(newStatus ClaimStatus) bool {
	validTransitions := map[ClaimStatus][]ClaimStatus{
		ClaimStatusOpen: {
			ClaimStatusWaitingResolution,
			ClaimStatusConfirmed,
			ClaimStatusCancelled,
			ClaimStatusExpired,
		},
		ClaimStatusWaitingResolution: {
			ClaimStatusConfirmed,
			ClaimStatusCancelled,
			ClaimStatusCompleted,
			ClaimStatusExpired,
			ClaimStatusAutoConfirmed,
		},
		ClaimStatusConfirmed: {
			ClaimStatusCompleted,
		},
	}

	allowedTransitions, exists := validTransitions[cs]
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
