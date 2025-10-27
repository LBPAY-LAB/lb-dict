package services

import (
	"context"
	"errors"
	"regexp"
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

// KeyLimits define limites por tipo de chave
var KeyLimits = map[KeyType]int{
	KeyTypeCPF:   5,  // Máximo 5 chaves CPF por titular
	KeyTypeCNPJ:  20, // Máximo 20 chaves CNPJ por CNPJ
	KeyTypeEmail: 20, // Máximo 20 chaves Email
	KeyTypePhone: 20, // Máximo 20 chaves Telefone
	KeyTypeEVP:   20, // Máximo 20 chaves EVP (aleatórias)
}

// KeyValidatorService serviço de validação de chaves PIX
type KeyValidatorService struct {
	entryCounter EntryCounter
}

// EntryCounter interface para contar entries
type EntryCounter interface {
	CountByOwnerAndType(ctx context.Context, ownerTaxID string, keyType KeyType) (int, error)
}

// NewKeyValidatorService cria nova instância
func NewKeyValidatorService(entryCounter EntryCounter) *KeyValidatorService {
	return &KeyValidatorService{
		entryCounter: entryCounter,
	}
}

// ValidateFormat valida formato da chave conforme tipo
func (s *KeyValidatorService) ValidateFormat(keyType KeyType, keyValue string) error {
	switch keyType {
	case KeyTypeCPF:
		return s.validateCPF(keyValue)
	case KeyTypeCNPJ:
		return s.validateCNPJ(keyValue)
	case KeyTypeEmail:
		return s.validateEmail(keyValue)
	case KeyTypePhone:
		return s.validatePhone(keyValue)
	case KeyTypeEVP:
		return s.validateEVP(keyValue)
	default:
		return errors.New("invalid key type")
	}
}

// ValidateLimits valida limites de chaves por titular
func (s *KeyValidatorService) ValidateLimits(ctx context.Context, keyType KeyType, ownerTaxID string) error {
	limit, ok := KeyLimits[keyType]
	if !ok {
		return errors.New("unknown key type")
	}

	count, err := s.entryCounter.CountByOwnerAndType(ctx, ownerTaxID, keyType)
	if err != nil {
		return errors.New("failed to count existing keys: " + err.Error())
	}

	if count >= limit {
		return errors.New("key limit exceeded: maximum " + string(rune(limit)) + " keys allowed for this type")
	}

	return nil
}

// validateCPF valida formato de CPF
func (s *KeyValidatorService) validateCPF(cpf string) error {
	// 1. Validar comprimento
	if len(cpf) != 11 {
		return errors.New("CPF must have 11 digits")
	}

	// 2. Validar apenas números
	if !regexp.MustCompile(`^\d{11}$`).MatchString(cpf) {
		return errors.New("CPF must contain only digits")
	}

	// 3. Rejeitar CPFs conhecidos inválidos
	invalidCPFs := []string{
		"00000000000", "11111111111", "22222222222", "33333333333",
		"44444444444", "55555555555", "66666666666", "77777777777",
		"88888888888", "99999999999",
	}
	for _, invalid := range invalidCPFs {
		if cpf == invalid {
			return errors.New("invalid CPF pattern")
		}
	}

	// 4. Validar dígitos verificadores (algoritmo oficial)
	if !s.validateCPFCheckDigits(cpf) {
		return errors.New("invalid CPF check digits")
	}

	return nil
}

// validateCPFCheckDigits valida dígitos verificadores do CPF
func (s *KeyValidatorService) validateCPFCheckDigits(cpf string) bool {
	// Primeiro dígito verificador
	sum := 0
	for i := 0; i < 9; i++ {
		digit := int(cpf[i] - '0')
		sum += digit * (10 - i)
	}
	firstDigit := (sum * 10) % 11
	if firstDigit == 10 {
		firstDigit = 0
	}
	if firstDigit != int(cpf[9]-'0') {
		return false
	}

	// Segundo dígito verificador
	sum = 0
	for i := 0; i < 10; i++ {
		digit := int(cpf[i] - '0')
		sum += digit * (11 - i)
	}
	secondDigit := (sum * 10) % 11
	if secondDigit == 10 {
		secondDigit = 0
	}
	if secondDigit != int(cpf[10]-'0') {
		return false
	}

	return true
}

// validateCNPJ valida formato de CNPJ
func (s *KeyValidatorService) validateCNPJ(cnpj string) error {
	// 1. Validar comprimento
	if len(cnpj) != 14 {
		return errors.New("CNPJ must have 14 digits")
	}

	// 2. Validar apenas números
	if !regexp.MustCompile(`^\d{14}$`).MatchString(cnpj) {
		return errors.New("CNPJ must contain only digits")
	}

	// 3. Rejeitar CNPJs conhecidos inválidos
	invalidCNPJs := []string{
		"00000000000000", "11111111111111", "22222222222222",
		"33333333333333", "44444444444444", "55555555555555",
		"66666666666666", "77777777777777", "88888888888888",
		"99999999999999",
	}
	for _, invalid := range invalidCNPJs {
		if cnpj == invalid {
			return errors.New("invalid CNPJ pattern")
		}
	}

	// 4. Validar dígitos verificadores
	if !s.validateCNPJCheckDigits(cnpj) {
		return errors.New("invalid CNPJ check digits")
	}

	return nil
}

// validateCNPJCheckDigits valida dígitos verificadores do CNPJ
func (s *KeyValidatorService) validateCNPJCheckDigits(cnpj string) bool {
	// Primeiro dígito verificador
	weights := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i := 0; i < 12; i++ {
		digit := int(cnpj[i] - '0')
		sum += digit * weights[i]
	}
	remainder := sum % 11
	firstDigit := 0
	if remainder >= 2 {
		firstDigit = 11 - remainder
	}
	if firstDigit != int(cnpj[12]-'0') {
		return false
	}

	// Segundo dígito verificador
	weights = []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum = 0
	for i := 0; i < 13; i++ {
		digit := int(cnpj[i] - '0')
		sum += digit * weights[i]
	}
	remainder = sum % 11
	secondDigit := 0
	if remainder >= 2 {
		secondDigit = 11 - remainder
	}
	if secondDigit != int(cnpj[13]-'0') {
		return false
	}

	return true
}

// validateEmail valida formato de email
func (s *KeyValidatorService) validateEmail(email string) error {
	// RFC 5322 simplified
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	// Validar tamanho máximo
	if len(email) > 254 {
		return errors.New("email exceeds maximum length (254 characters)")
	}

	return nil
}

// validatePhone valida formato de telefone
func (s *KeyValidatorService) validatePhone(phone string) error {
	// E.164 format: +5511999998888 (13-14 dígitos com +55)
	phoneRegex := regexp.MustCompile(`^\+55[1-9]{2}9?[0-9]{8}$`)
	if !phoneRegex.MatchString(phone) {
		return errors.New("invalid phone format (must be E.164: +5511999998888)")
	}
	return nil
}

// validateEVP valida formato de EVP (UUID v4)
func (s *KeyValidatorService) validateEVP(evp string) error {
	// UUID v4 format
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	if !uuidRegex.MatchString(evp) {
		return errors.New("invalid EVP format (must be UUID v4)")
	}
	return nil
}
