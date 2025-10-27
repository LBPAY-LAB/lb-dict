package services

import (
	"context"
	"errors"
)

// DuplicateKeyChecker verifica duplicação de chaves PIX
type DuplicateKeyChecker struct {
	entryFinder EntryFinder
}

// EntryFinder interface para buscar entries
type EntryFinder interface {
	FindByKeyValue(ctx context.Context, keyValue string) (*Entry, error)
}

// Entry representa uma chave PIX (simplificado)
type Entry struct {
	ID       string
	KeyValue string
	Status   string
	ISPB     string
}

// NewDuplicateKeyChecker cria nova instância
func NewDuplicateKeyChecker(entryFinder EntryFinder) *DuplicateKeyChecker {
	return &DuplicateKeyChecker{
		entryFinder: entryFinder,
	}
}

// IsDuplicate verifica se chave já existe localmente (neste PSP)
func (c *DuplicateKeyChecker) IsDuplicate(ctx context.Context, keyValue string) (bool, error) {
	entry, err := c.entryFinder.FindByKeyValue(ctx, keyValue)

	// Entry não encontrada = não é duplicada
	if err != nil {
		if err.Error() == "not found" {
			return false, nil
		}
		return false, errors.New("failed to check duplicate: " + err.Error())
	}

	// Entry encontrada e ACTIVE = é duplicada
	if entry.Status == "ACTIVE" {
		return true, nil
	}

	// Entry DELETED ou TRANSFERRED = não é duplicada (pode recriar)
	if entry.Status == "DELETED" || entry.Status == "TRANSFERRED" {
		return false, nil
	}

	// Entry PENDING ou BLOCKED = é duplicada (não pode criar)
	if entry.Status == "PENDING" || entry.Status == "BLOCKED" {
		return true, nil
	}

	// Status desconhecido = considerar duplicada por segurança
	return true, nil
}

// IsDuplicateGlobal verifica se chave existe em outro PSP (via RSFN)
// Esta função é chamada APÓS criação local, durante sync com Bacen
func (c *DuplicateKeyChecker) IsDuplicateGlobal(ctx context.Context, keyValue string, ispb string) (bool, string, error) {
	// TODO: Integrar com RSFN Connect (gRPC call)
	// Retorna: (isDuplicate, ownerISPB, error)

	// Por ora, assume que não é duplicada globalmente
	// A validação real acontece no Bacen via RSFN
	return false, "", nil
}

// ValidateUniqueness valida unicidade local antes de criar chave
func (c *DuplicateKeyChecker) ValidateUniqueness(ctx context.Context, keyValue string) error {
	isDuplicate, err := c.IsDuplicate(ctx, keyValue)
	if err != nil {
		return errors.New("failed to validate uniqueness: " + err.Error())
	}

	if isDuplicate {
		return errors.New("key already registered in this PSP")
	}

	return nil
}
