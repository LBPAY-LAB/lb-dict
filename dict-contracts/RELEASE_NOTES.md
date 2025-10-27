# Release Notes - v0.1.0

**Release Date**: October 26, 2025
**Project**: DICT Contracts - Protocol Buffers for LBPay DICT System
**Status**: Initial Release

---

## Overview

This is the **initial release** of the DICT Contracts module, providing comprehensive Protocol Buffers definitions for all gRPC services in the LBPay DICT (Diretório de Identificadores de Contas Transacionais) system.

The contracts define the communication layer between:
- **FrontEnd** <-> **Core DICT API** (via CoreDictService)
- **RSFN Connect** <-> **Bridge** <-> **Bacen DICT** (via BridgeService)

---

## What's Included

### Proto Files

#### 1. proto/common.proto
Shared types and enums used across all services.

**Enums** (5):
- `KeyType`: PIX key types (CPF, CNPJ, EMAIL, PHONE, EVP)
- `AccountType`: Bank account types (CHECKING, SAVINGS, PAYMENT, SALARY)
- `DocumentType`: Document types (CPF, CNPJ)
- `EntryStatus`: DICT entry status (ACTIVE, PORTABILITY_PENDING, CLAIM_PENDING, DELETED)
- `ClaimStatus`: Claim status (OPEN, WAITING_RESOLUTION, CONFIRMED, CANCELLED, COMPLETED, EXPIRED)

**Messages** (7):
- `Account`: Complete bank account representation
- `DictKey`: DICT key (type + value)
- `ValidationError`: Field validation errors
- `BusinessError`: Business rule errors
- `InfrastructureError`: Infrastructure errors
- `BacenError`: Bacen-specific errors
- `ErrorResponse`: Unified error wrapper

---

#### 2. proto/core_dict.proto
CoreDictService - FrontEnd to Core DICT API communication.

**Total Methods**: 15

**Key Operations** (4 methods):
1. `CreateKey` - Create new PIX key
2. `ListKeys` - List user's PIX keys
3. `GetKey` - Get key details
4. `DeleteKey` - Delete PIX key

**Claim Operations** (6 methods):
5. `StartClaim` - Initiate 30-day claim for a key
6. `GetClaimStatus` - Check claim status
7. `ListIncomingClaims` - List received claims (as key owner)
8. `ListOutgoingClaims` - List sent claims (as claimer)
9. `RespondToClaim` - Accept or reject claim
10. `CancelClaim` - Cancel sent claim

**Portability Operations** (3 methods):
11. `StartPortability` - Initiate account portability
12. `ConfirmPortability` - Confirm portability
13. `CancelPortability` - Cancel portability

**Query Operations** (1 method):
14. `LookupKey` - Query third-party DICT keys (for PIX transactions)

**Health Check** (1 method):
15. `HealthCheck` - Service health status

---

#### 3. proto/bridge.proto
BridgeService - RSFN Connect to Bridge to Bacen DICT communication.

**Total Methods**: 14

**Entry Operations** (4 methods):
1. `CreateEntry` - Create key in Bacen DICT
2. `GetEntry` - Fetch key from Bacen DICT
3. `DeleteEntry` - Delete key in Bacen DICT
4. `UpdateEntry` - Update account data

**Claim Operations** (4 methods):
5. `CreateClaim` - Create claim in Bacen (30-day period)
6. `GetClaim` - Fetch claim status from Bacen
7. `CompleteClaim` - Complete claim (approval)
8. `CancelClaim` - Cancel claim (rejection or timeout)

**Portability Operations** (3 methods):
9. `InitiatePortability` - Start portability process
10. `ConfirmPortability` - Confirm account portability
11. `CancelPortability` - Cancel portability

**Directory Queries** (2 methods):
12. `GetDirectory` - Query complete directory
13. `SearchEntries` - Search keys by criteria

**Health Check** (1 method):
14. `HealthCheck` - Check Bridge and Bacen connectivity

---

## Statistics

- **Total gRPC Methods**: 29
- **Total Proto Files**: 3
- **Total Enums**: 7
- **Total Messages**: 60+
- **Supported Go Version**: 1.24.0
- **Package Version**: v0.1.0

---

## Generated Code

After running `make proto-gen`, the following Go code is generated:

```
gen/proto/
├── common/v1/
│   └── common.pb.go
├── core/v1/
│   ├── core_dict.pb.go
│   └── core_dict_grpc.pb.go
└── bridge/v1/
    ├── bridge.pb.go
    └── bridge_grpc.pb.go
```

---

## Usage Examples

### Import Packages

```go
import (
    commonv1 "github.com/lbpay/dict-contracts/gen/proto/common/v1"
    corev1 "github.com/lbpay/dict-contracts/gen/proto/core/v1"
    bridgev1 "github.com/lbpay/dict-contracts/gen/proto/bridge/v1"
)
```

### Example 1: Create PIX Key (CoreDictService)

```go
// Create gRPC client
conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

client := corev1.NewCoreDictServiceClient(conn)

// Create CPF-type PIX key
resp, err := client.CreateKey(context.Background(), &corev1.CreateKeyRequest{
    KeyType:   commonv1.KeyType_KEY_TYPE_CPF,
    KeyValue:  "12345678900",
    AccountId: "acc-550e8400-e29b-41d4-a716-446655440000",
})

if err != nil {
    log.Fatalf("CreateKey failed: %v", err)
}

fmt.Printf("Key created: %s\n", resp.KeyId)
fmt.Printf("Status: %s\n", resp.Status)
```

### Example 2: Start 30-Day Claim

```go
// Start claim for a key owned by another user
claimResp, err := client.StartClaim(context.Background(), &corev1.StartClaimRequest{
    Key: &commonv1.DictKey{
        KeyType:  commonv1.KeyType_KEY_TYPE_EMAIL,
        KeyValue: "user@example.com",
    },
    AccountId: "acc-my-account-id",
})

if err != nil {
    log.Fatalf("StartClaim failed: %v", err)
}

fmt.Printf("Claim ID: %s\n", claimResp.ClaimId)
fmt.Printf("Expires at: %s\n", claimResp.ExpiresAt)
fmt.Printf("Message: %s\n", claimResp.Message)
```

### Example 3: Query Third-Party Key

```go
// Lookup a key before making a PIX transaction
lookupResp, err := client.LookupKey(context.Background(), &corev1.LookupKeyRequest{
    Key: &commonv1.DictKey{
        KeyType:  commonv1.KeyType_KEY_TYPE_PHONE,
        KeyValue: "+5511999999999",
    },
})

if err != nil {
    log.Fatalf("LookupKey failed: %v", err)
}

fmt.Printf("Account holder: %s\n", lookupResp.AccountHolderName)
fmt.Printf("ISPB: %s\n", lookupResp.Account.Ispb)
fmt.Printf("Account: %s-%s\n",
    lookupResp.Account.AccountNumber,
    lookupResp.Account.AccountCheckDigit)
```

### Example 4: Bridge Entry Creation

```go
// Create entry in Bacen via Bridge
bridgeClient := bridgev1.NewBridgeServiceClient(bridgeConn)

entryResp, err := bridgeClient.CreateEntry(context.Background(), &bridgev1.CreateEntryRequest{
    Key: &commonv1.DictKey{
        KeyType:  commonv1.KeyType_KEY_TYPE_CPF,
        KeyValue: "12345678900",
    },
    Account: &commonv1.Account{
        Ispb:                  "12345678",
        AccountType:           commonv1.AccountType_ACCOUNT_TYPE_CHECKING,
        AccountNumber:         "123456",
        AccountCheckDigit:     "7",
        BranchCode:            "0001",
        AccountHolderName:     "Jose Silva",
        AccountHolderDocument: "12345678900",
        DocumentType:          commonv1.DocumentType_DOCUMENT_TYPE_CPF,
    },
    IdempotencyKey: "req-550e8400-e29b-41d4-a716-446655440000",
    RequestId:      "trace-550e8400",
})

if err != nil {
    log.Fatalf("CreateEntry failed: %v", err)
}

fmt.Printf("Entry ID: %s\n", entryResp.EntryId)
fmt.Printf("Bacen Transaction ID: %s\n", entryResp.BacenTransactionId)
```

---

## Key Features

### Idempotency Support
All write operations include `idempotency_key` field for safe retries:
```protobuf
string idempotency_key = 3;  // For retry safety
```

### Request Tracing
All operations include `request_id` for distributed tracing:
```protobuf
string request_id = 4;  // For log correlation
```

### Structured Error Handling
Comprehensive error types with context:
```protobuf
message ErrorResponse {
  int32 grpc_code = 1;
  string message = 2;
  oneof details {
    ValidationError validation = 3;
    BusinessError business = 4;
    InfrastructureError infrastructure = 5;
    BacenError bacen = 6;
  }
  string request_id = 7;
  google.protobuf.Timestamp timestamp = 8;
}
```

### Type Safety
All domain concepts represented as typed enums:
- Key types (CPF, CNPJ, EMAIL, PHONE, EVP)
- Account types (CHECKING, SAVINGS, PAYMENT, SALARY)
- Status values (entry status, claim status)

### Pagination Support
List operations support cursor-based pagination:
```protobuf
message ListKeysRequest {
  int32 page_size = 1;     // Default: 20, Max: 100
  string page_token = 2;
}

message ListKeysResponse {
  repeated KeySummary keys = 1;
  string next_page_token = 2;
  int32 total_count = 3;
}
```

---

## Dependencies

### Runtime Dependencies
```go
require (
    google.golang.org/grpc v1.76.0
    google.golang.org/protobuf v1.36.10
)
```

### Development Tools
- `protoc` (Protocol Buffers compiler)
- `protoc-gen-go` (Go plugin for protoc)
- `protoc-gen-go-grpc` (gRPC plugin for protoc)
- `buf` (optional, for linting and breaking change detection)

---

## Installation

### Using Go Modules

```bash
go get github.com/lbpay/dict-contracts@v0.1.0
```

### Development Setup

```bash
# Clone repository
git clone https://github.com/lbpay/dict-contracts.git
cd dict-contracts

# Run setup script
./setup.sh

# Generate code
make proto-gen

# Run linting
make proto-lint
```

---

## Versioning

This project follows [Semantic Versioning](https://semver.org/):
- **v0.1.0**: Initial release (current)
- Future **v0.x.x**: Pre-1.0 releases may include breaking changes
- Future **v1.0.0**: Stable API, backward compatibility guaranteed

---

## Architecture

```
┌─────────────────────┐
│  FrontEnd (Web/App) │
└──────────┬──────────┘
           │ gRPC (CoreDictService)
           │ 15 methods
           ▼
┌─────────────────────┐
│   Core DICT API     │
└──────────┬──────────┘
           │ gRPC (internal)
           │
           ▼
┌─────────────────────┐
│   RSFN Connect      │
└──────────┬──────────┘
           │ gRPC (BridgeService)
           │ 14 methods
           ▼
┌─────────────────────┐
│   Bridge DICT       │
└──────────┬──────────┘
           │ SOAP/mTLS
           │
           ▼
┌─────────────────────┐
│   Bacen DICT        │
└─────────────────────┘
```

---

## Documentation

- **README.md**: API reference and quick start
- **IMPLEMENTATION.md**: Technical implementation details
- **CHANGELOG.md**: Version history
- **Proto files**: Inline documentation with comments

---

## Support

**Project**: DICT LBPay
**Squad**: Implementation
**Contact**: api-specialist

---

## License

Proprietary - LBPay
Internal use only

---

## Next Steps

After installing v0.1.0:

1. **Implement Server**: Create gRPC servers implementing the services
2. **Implement Client**: Create gRPC clients consuming the services
3. **Add Tests**: Write integration tests using the contracts
4. **Monitor Performance**: Track RPC performance and latency
5. **Provide Feedback**: Report issues or suggest improvements

---

**Release Tag**: `v0.1.0`
**Go Module**: `github.com/lbpay/dict-contracts@v0.1.0`
