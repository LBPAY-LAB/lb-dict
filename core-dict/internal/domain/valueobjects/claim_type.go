package valueobjects

import "errors"

// ClaimType representa os tipos de reivindicação de chave PIX
type ClaimType string

const (
	ClaimTypeOwnership   ClaimType = "OWNERSHIP"   // Reivindicação de propriedade
	ClaimTypePortability ClaimType = "PORTABILITY" // Portabilidade de conta
)

var validClaimTypes = map[ClaimType]bool{
	ClaimTypeOwnership:   true,
	ClaimTypePortability: true,
}

// NewClaimType cria e valida um tipo de reivindicação
func NewClaimType(s string) (ClaimType, error) {
	ct := ClaimType(s)
	if !validClaimTypes[ct] {
		return "", errors.New("invalid claim type: must be OWNERSHIP or PORTABILITY")
	}
	return ct, nil
}

// String retorna a representação string do tipo de reivindicação
func (ct ClaimType) String() string {
	return string(ct)
}

// IsValid verifica se o tipo de reivindicação é válido
func (ct ClaimType) IsValid() bool {
	return validClaimTypes[ct]
}
