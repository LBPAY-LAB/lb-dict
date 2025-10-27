package valueobjects

import "errors"

// KeyType representa os tipos de chave PIX suportados pelo DICT
type KeyType string

const (
	KeyTypeCPF   KeyType = "CPF"
	KeyTypeCNPJ  KeyType = "CNPJ"
	KeyTypeEmail KeyType = "EMAIL"
	KeyTypePhone KeyType = "PHONE"
	KeyTypeEVP   KeyType = "EVP"
)

var validKeyTypes = map[KeyType]bool{
	KeyTypeCPF:   true,
	KeyTypeCNPJ:  true,
	KeyTypeEmail: true,
	KeyTypePhone: true,
	KeyTypeEVP:   true,
}

// NewKeyType cria e valida um tipo de chave
func NewKeyType(s string) (KeyType, error) {
	kt := KeyType(s)
	if !validKeyTypes[kt] {
		return "", errors.New("invalid key type: must be CPF, CNPJ, EMAIL, PHONE, or EVP")
	}
	return kt, nil
}

// String retorna a representação string do tipo de chave
func (kt KeyType) String() string {
	return string(kt)
}

// IsValid verifica se o tipo de chave é válido
func (kt KeyType) IsValid() bool {
	return validKeyTypes[kt]
}

// AllKeyTypes retorna todos os tipos de chave válidos
func AllKeyTypes() []KeyType {
	return []KeyType{
		KeyTypeCPF,
		KeyTypeCNPJ,
		KeyTypeEmail,
		KeyTypePhone,
		KeyTypeEVP,
	}
}
