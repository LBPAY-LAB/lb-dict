package entities

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// AccountType representa os tipos de conta suportados
type AccountType string

const (
	AccountTypeCACC AccountType = "CACC" // Conta Corrente
	AccountTypeSVGS AccountType = "SVGS" // Poupança
	AccountTypeSLRY AccountType = "SLRY" // Salário
	AccountTypeTRAN AccountType = "TRAN" // Transacional
)

// OwnerType representa o tipo de titular da conta
type OwnerType string

const (
	OwnerTypeNaturalPerson OwnerType = "NATURAL_PERSON" // Pessoa Física
	OwnerTypeLegalEntity   OwnerType = "LEGAL_ENTITY"   // Pessoa Jurídica
)

// Account representa uma conta CID (Conta de Identificação de Depósito)
type Account struct {
	ID            uuid.UUID
	ISPB          string      // ISPB da instituição (8 dígitos)
	Branch        string      // Agência
	AccountNumber string      // Número da conta
	AccountType   AccountType // Tipo de conta
	Status        string      // Status da conta (ACTIVE, BLOCKED, CLOSED)
	Owner         Owner       // Titular da conta
	OpenedAt      time.Time
	ClosedAt      *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

// Owner representa o titular de uma conta
type Owner struct {
	TaxID       string    // CPF (11) ou CNPJ (14)
	Type        OwnerType // Tipo de titular
	Name        string    // Nome completo
	NameEncoded string    // Nome codificado (LGPD)
}

var (
	ispbRegex = regexp.MustCompile(`^\d{8}$`)
	cpfRegex  = regexp.MustCompile(`^\d{11}$`)
	cnpjRegex = regexp.MustCompile(`^\d{14}$`)
)

// NewAccount cria uma nova conta
func NewAccount(
	ispb string,
	branch string,
	accountNumber string,
	accountType AccountType,
	owner Owner,
) (*Account, error) {
	if !ispbRegex.MatchString(ispb) {
		return nil, errors.New("invalid ISPB: must be 8 digits")
	}
	if branch == "" {
		return nil, errors.New("branch cannot be empty")
	}
	if accountNumber == "" {
		return nil, errors.New("account number cannot be empty")
	}
	if err := validateAccountType(accountType); err != nil {
		return nil, err
	}
	if err := owner.Validate(); err != nil {
		return nil, errors.New("invalid owner: " + err.Error())
	}

	now := time.Now()
	return &Account{
		ID:            uuid.New(),
		ISPB:          ispb,
		Branch:        branch,
		AccountNumber: accountNumber,
		AccountType:   accountType,
		Status:        "ACTIVE",
		Owner:         owner,
		OpenedAt:      now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// Validate valida os campos da conta
func (a *Account) Validate() error {
	if a.ID == uuid.Nil {
		return errors.New("account ID cannot be nil")
	}
	if !ispbRegex.MatchString(a.ISPB) {
		return errors.New("invalid ISPB")
	}
	if a.Branch == "" {
		return errors.New("branch cannot be empty")
	}
	if a.AccountNumber == "" {
		return errors.New("account number cannot be empty")
	}
	if err := validateAccountType(a.AccountType); err != nil {
		return err
	}
	return a.Owner.Validate()
}

// Validate valida os campos do titular
func (o *Owner) Validate() error {
	if o.Name == "" {
		return errors.New("owner name cannot be empty")
	}
	if o.TaxID == "" {
		return errors.New("tax ID cannot be empty")
	}

	switch o.Type {
	case OwnerTypeNaturalPerson:
		if !cpfRegex.MatchString(o.TaxID) {
			return errors.New("invalid CPF: must be 11 digits")
		}
	case OwnerTypeLegalEntity:
		if !cnpjRegex.MatchString(o.TaxID) {
			return errors.New("invalid CNPJ: must be 14 digits")
		}
	default:
		return errors.New("invalid owner type")
	}

	return nil
}

func validateAccountType(at AccountType) error {
	validTypes := map[AccountType]bool{
		AccountTypeCACC: true,
		AccountTypeSVGS: true,
		AccountTypeSLRY: true,
		AccountTypeTRAN: true,
	}
	if !validTypes[at] {
		return errors.New("invalid account type")
	}
	return nil
}

// IsActive verifica se a conta está ativa
func (a *Account) IsActive() bool {
	return a.Status == "ACTIVE" && a.DeletedAt == nil && a.ClosedAt == nil
}

// IsClosed verifica se a conta está fechada
func (a *Account) IsClosed() bool {
	return a.ClosedAt != nil || a.Status == "CLOSED"
}

// Close fecha a conta
func (a *Account) Close() error {
	if a.IsClosed() {
		return errors.New("account is already closed")
	}
	now := time.Now()
	a.Status = "CLOSED"
	a.ClosedAt = &now
	a.UpdatedAt = now
	return nil
}

// Block bloqueia a conta
func (a *Account) Block() error {
	if !a.IsActive() {
		return errors.New("only active accounts can be blocked")
	}
	a.Status = "BLOCKED"
	a.UpdatedAt = time.Now()
	return nil
}

// Unblock desbloqueia a conta
func (a *Account) Unblock() error {
	if a.Status != "BLOCKED" {
		return errors.New("account is not blocked")
	}
	a.Status = "ACTIVE"
	a.UpdatedAt = time.Now()
	return nil
}
