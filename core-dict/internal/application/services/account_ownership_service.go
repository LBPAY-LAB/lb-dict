package services

import (
	"context"
	"errors"
)

// AccountOwnershipService valida ownership de chave PIX
type AccountOwnershipService struct {
	accountRepo AccountRepository
}

// AccountRepository interface para buscar contas
type AccountRepository interface {
	FindByID(ctx context.Context, accountID string) (*Account, error)
	VerifyOwnership(ctx context.Context, accountID, taxID string) (bool, error)
}

// Account representa uma conta CID
type Account struct {
	ID            string
	ISPB          string
	Branch        string
	AccountNumber string
	OwnerTaxID    string
	OwnerName     string
	Status        string
}

// NewAccountOwnershipService cria nova instância
func NewAccountOwnershipService(accountRepo AccountRepository) *AccountOwnershipService {
	return &AccountOwnershipService{
		accountRepo: accountRepo,
	}
}

// ValidateOwnership valida se a chave pertence ao titular da conta
func (s *AccountOwnershipService) ValidateOwnership(ctx context.Context, keyType KeyType, keyValue, ownerTaxID string) error {
	switch keyType {
	case KeyTypeCPF:
		return s.validateCPFOwnership(keyValue, ownerTaxID)
	case KeyTypeCNPJ:
		return s.validateCNPJOwnership(keyValue, ownerTaxID)
	case KeyTypeEmail:
		// Email ownership é validado via OTP (fora deste serviço)
		return nil
	case KeyTypePhone:
		// Phone ownership é validado via OTP (fora deste serviço)
		return nil
	case KeyTypeEVP:
		// EVP não requer validação de ownership (gerado aleatoriamente)
		return nil
	default:
		return errors.New("unsupported key type for ownership validation")
	}
}

// validateCPFOwnership valida que CPF da chave pertence ao titular da conta
func (s *AccountOwnershipService) validateCPFOwnership(cpf, ownerTaxID string) error {
	if cpf != ownerTaxID {
		return errors.New("CPF key must match account owner CPF")
	}
	return nil
}

// validateCNPJOwnership valida que CNPJ da chave pertence ao titular da conta
func (s *AccountOwnershipService) validateCNPJOwnership(cnpj, ownerTaxID string) error {
	if cnpj != ownerTaxID {
		return errors.New("CNPJ key must match account owner CNPJ")
	}
	return nil
}

// ValidateAccountStatus valida que a conta está ativa e válida
func (s *AccountOwnershipService) ValidateAccountStatus(ctx context.Context, accountID string) error {
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return errors.New("account not found")
	}

	if account.Status != "ACTIVE" {
		return errors.New("account must be ACTIVE to register keys")
	}

	return nil
}

// VerifyOwnership verifica ownership de conta com TaxID
func (s *AccountOwnershipService) VerifyOwnership(ctx context.Context, accountID, taxID string) (bool, error) {
	isOwner, err := s.accountRepo.VerifyOwnership(ctx, accountID, taxID)
	if err != nil {
		return false, errors.New("failed to verify ownership: " + err.Error())
	}
	return isOwner, nil
}
