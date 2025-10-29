# Task 2: Entry Domain Model Analysis

## Overview
Analysis of the Entry domain model in connector-dict to understand data structures and their relationship to CID generation requirements.

## Domain Models Structure

### 1. Entry Entity
Location: `/apps/dict/domain/entry.go`

```go
type Entry struct {
    Key                   Key
    Account               Account
    Owner                 Owner
    CreationDate          time.Time
    KeyOwnershipDate      time.Time
    OpenClaimCreationDate time.Time
}
```

### 2. Key Model
Location: `/apps/dict/domain/key_type.go`

```go
type Key struct {
    Value string
    Type  KeyType
}

type KeyType string

const (
    KeyCPF   KeyType = "CPF"    // 11 digits
    KeyCNPJ  KeyType = "CNPJ"   // 14 digits
    KeyPHONE KeyType = "PHONE"  // +[country][number]
    KeyEMAIL KeyType = "EMAIL"  // lowercase, max 77 chars
    KeyEVP   KeyType = "EVP"    // UUID format
)
```

**Key Features:**
- Automatic type detection based on value pattern
- Built-in validation using regex patterns
- CPF/CNPJ validation using brdoc library
- Normalization handled in validation

### 3. Account Model
```go
type Account struct {
    Participant   string       // ISPB - 8 digits
    Branch        *string      // Optional branch number
    AccountNumber string       // Account number
    AccountType   AccountType  // CACC or SVGS
    OpeningDate   time.Time    // Account opening date
}

type AccountType string
const (
    AccountCACC AccountType = "CACC" // Checking account
    AccountSVGS AccountType = "SVGS" // Savings account
)
```

### 4. Owner Model
```go
type Owner struct {
    Type        string  // "NATURAL_PERSON" or "LEGAL_PERSON"
    TaxIDNumber string  // CPF or CNPJ
    Name        string  // Legal name
    TradeName   string  // Trade name (for companies)
}
```

## Data Normalization

### Key Normalization
The system handles normalization at the Key level:
- **CPF/CNPJ**: Numbers only, validated
- **Phone**: E.164 format with country code
- **Email**: Lowercase, validated against RFC spec
- **EVP**: UUID format, lowercase

### Entry Lifecycle
1. **Creation**: Entry created with all required fields
2. **Update**: Account and Owner information can be updated
3. **Delete**: Entry removed with reason code

## Mapping to CID Fields

### CID Required Fields from Entry Model

| CID Field | Entry Model Source | Notes |
|-----------|-------------------|-------|
| Key Value | `Entry.Key.Value` | Already normalized |
| Key Type | `Entry.Key.Type` | Enum values match BACEN |
| ISPB | `Entry.Account.Participant` | 8-digit institution code |
| Branch | `Entry.Account.Branch` | Optional, may be null |
| Account Number | `Entry.Account.AccountNumber` | |
| Account Type | `Entry.Account.AccountType` | CACC or SVGS |
| Person Type | `Entry.Owner.Type` | Natural or Legal |
| Tax ID | `Entry.Owner.TaxIDNumber` | CPF or CNPJ |
| Name | `Entry.Owner.Name` | Legal name |
| Creation Date | `Entry.CreationDate` | ISO 8601 format |
| Ownership Date | `Entry.KeyOwnershipDate` | ISO 8601 format |

## SDK Integration

The domain models integrate with the SDK types:
```go
// SDK types from github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/bacen
type Entry struct {
    Key     string
    KeyType KeyType
    Account BrazilianAccount
    Owner   Person
}
```

### Conversion Pattern
```go
// Domain to SDK conversion in application layer
payload := pkgEntry.CreateEntryRequest{
    Entry: bacen.Entry{
        Key:     entry.Key,
        KeyType: bacen.KeyType(entry.KeyType),
        Account: bacen.BrazilianAccount{
            Participant:   entry.Account.Participant,
            Branch:        entry.Account.Branch,
            AccountNumber: entry.Account.AccountNumber,
            AccountType:   bacen.AccountType(entry.Account.AccountType),
            OpeningDate:   entry.Account.OpeningDate,
        },
        Owner: bacen.Person{
            Type:        bacen.PersonType(entry.Owner.Type),
            TaxIDNumber: entry.Owner.TaxIDNumber,
            Name:        entry.Owner.Name,
            TradeName:   entry.Owner.TradeName,
        },
    },
}
```

## Data Completeness for CID

### Required for CID Generation
✅ **All fields available**:
- Key identifier (value + type)
- Financial institution (ISPB)
- Account details (branch, number, type)
- Owner identification (tax ID, name, type)
- Temporal data (creation, ownership dates)

### CID Generation Logic
Based on the Entry model, CID can be generated as:
```
CID = Hash(
    KeyValue +
    KeyType +
    ISPB +
    Branch +
    AccountNumber +
    AccountType +
    TaxID +
    CreationTimestamp
)
```

## Validation Rules

### Business Rules in Domain
1. **Key Validation**:
   - CPF: 11 digits with valid checksum
   - CNPJ: 14 digits with valid checksum
   - Phone: E.164 format
   - Email: RFC 5322 compliant
   - EVP: Valid UUID v4

2. **Account Validation**:
   - Participant: Must be valid ISPB
   - AccountType: Only CACC or SVGS
   - OpeningDate: Cannot be future date

3. **Owner Validation**:
   - TaxIDNumber must match Type (CPF for Natural, CNPJ for Legal)
   - Name is required
   - TradeName only for Legal persons

## Reusability Assessment

### Can we reuse Entry domain model?
**YES** - The Entry domain model is perfectly suitable for CID generation:

1. **Complete Data**: Contains all fields required by BACEN spec
2. **Validated Data**: Built-in validation ensures data quality
3. **Normalized Format**: Keys and data already normalized
4. **Type Safety**: Strong typing prevents data errors
5. **SDK Compatible**: Maps directly to BACEN SDK types

### Recommended Approach
1. **Import Domain Types**: Reuse Entry, Key, Account, Owner types
2. **Add CID Logic**: Create CID generation functions using Entry data
3. **Maintain Separation**: Keep CID logic in VSync container
4. **Event Processing**: Deserialize events to Entry types for processing

## Relationships

### Entry ↔ Key
- One-to-one relationship
- Key is embedded in Entry
- Key type determines validation rules

### Entry ↔ Account
- One-to-one relationship
- Account represents where funds are held
- ISPB identifies the institution

### Entry ↔ Owner
- One-to-one relationship
- Owner can be person or company
- Tax ID is unique identifier

### Entry ↔ Claims
- One Entry can have multiple claims
- Claims tracked separately
- OpenClaimCreationDate tracks active claims

## Conclusion

The Entry domain model is **well-designed and complete** for CID generation:

1. **All required fields are present** in the model
2. **Data is already normalized** and validated
3. **Type safety** prevents data corruption
4. **Direct mapping** to BACEN SDK types
5. **Reusable** across containers

The VSync system can:
- Import and reuse the domain types
- Process events containing Entry data
- Generate CID using the complete Entry information
- Maintain consistency with the main system

## Recommendations

1. **Reuse Domain Types**: Import Entry domain types into VSync
2. **Leverage Validation**: Use existing validation logic
3. **Maintain Compatibility**: Keep field mappings consistent
4. **Event Serialization**: Use JSON for Entry event payloads
5. **Type Converters**: Create helpers for SDK ↔ Domain conversion