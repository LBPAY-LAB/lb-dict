# TST-001: Test Cases - CreateEntry Operation

**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: QA Team
**Status**: ✅ Completo

---

## Sumário Executivo

Este documento apresenta os **test cases completos** para a operação **CreateEntry** (criação de chave PIX) no sistema DICT LBPay.

**Objetivo**: Garantir cobertura de testes abrangente para todos os cenários da operação CreateEntry, incluindo happy paths, edge cases, e error handling.

**Cobertura**:
- Happy Path (P0)
- Invalid Key Formats (P0)
- Duplicate Key (P0)
- Invalid Account (P1)
- Authentication/Authorization Errors (P0)
- Performance Tests (P1)
- Security Tests (P1)

**Referências**:
- [INT-001: Flow CreateEntry E2E](../../12_Integracao/Fluxos/INT-001_Flow_CreateEntry_E2E.md)
- [API-002: Core DICT REST API](../../04_APIs/REST/API-002_Core_DICT_REST_API.md)

---

## Test Environment Setup

### Prerequisites
```yaml
Environment: test
Core DICT API: https://dict-api-test.lbpay.com.br
Auth Service: https://auth-test.lbpay.com.br
Database: PostgreSQL 16.4 (test)
Temporal: test.temporal.lbpay.com.br
Bridge Mock: Bacen mock enabled
```

### Test Users
```yaml
User 1:
  email: test.user@lbpay.com.br
  password: Test@1234
  scopes: [dict:read, dict:write]
  role: customer

User 2 (No Write Permission):
  email: readonly@lbpay.com.br
  password: Test@1234
  scopes: [dict:read]
  role: customer

Admin User:
  email: admin@lbpay.com.br
  password: Admin@1234
  scopes: [dict:read, dict:write, dict:admin]
  role: admin
```

### Test Accounts
```json
{
  "account_1": {
    "ispb": "12345678",
    "account_number": "123456",
    "branch": "0001",
    "account_type": "CACC",
    "status": "ACTIVE"
  },
  "account_2": {
    "ispb": "12345678",
    "account_number": "654321",
    "branch": "0002",
    "account_type": "SVGS",
    "status": "ACTIVE"
  },
  "account_inactive": {
    "ispb": "12345678",
    "account_number": "999999",
    "branch": "0999",
    "account_type": "CACC",
    "status": "INACTIVE"
  }
}
```

---

## P0 Test Cases (Critical Path)

### TC-001-001: Create Entry - CPF (Happy Path)

**Priority**: P0 (Critical)
**Type**: E2E Functional
**Estimated Duration**: 2-3 seconds

#### Preconditions
- User is authenticated with valid JWT token
- User has scope `dict:write`
- Account exists in LBPay Ledger and is ACTIVE
- CPF key does not exist in DICT

#### Test Data
```json
{
  "key_type": "CPF",
  "key_value": "12345678900",
  "account": {
    "ispb": "12345678",
    "account_number": "123456",
    "branch": "0001",
    "account_type": "CACC"
  }
}
```

#### BDD Format (Given/When/Then)
```gherkin
Given the user is authenticated with valid JWT token
And the user has scope "dict:write"
And the account "123456" exists and is ACTIVE
And the CPF "12345678900" does not exist in DICT
When the user sends POST /api/v1/keys with CPF data
Then the API returns 201 Created
And the response contains entry_id
And the response contains status "PENDING"
And within 2 seconds the entry status changes to "ACTIVE"
And the entry has a valid external_id from Bacen
And an audit log entry is created
And a notification is sent to the user
```

#### Steps
1. Authenticate user and obtain JWT token:
   ```bash
   POST /auth/login
   {
     "email": "test.user@lbpay.com.br",
     "password": "Test@1234"
   }
   ```
   Expected: 200 OK with `access_token`

2. Create CPF entry:
   ```bash
   POST /api/v1/keys
   Authorization: Bearer {access_token}
   Content-Type: application/json

   {
     "key_type": "CPF",
     "key_value": "12345678900",
     "account": {
       "ispb": "12345678",
       "account_number": "123456",
       "branch": "0001",
       "account_type": "CACC"
     }
   }
   ```

3. Verify synchronous response:
   Expected: 201 Created
   ```json
   {
     "entry_id": "550e8400-e29b-41d4-a716-446655440000",
     "key_type": "CPF",
     "key_value": "12345678900",
     "account": {
       "ispb": "12345678",
       "account_number": "123456",
       "branch": "0001",
       "account_type": "CACC"
     },
     "status": "PENDING",
     "created_at": "2025-10-25T10:00:00Z"
   }
   ```

4. Poll entry status (retry up to 10 times, 200ms interval):
   ```bash
   GET /api/v1/keys/CPF/12345678900
   Authorization: Bearer {access_token}
   ```

5. Verify final status:
   Expected: 200 OK
   ```json
   {
     "entry_id": "550e8400-e29b-41d4-a716-446655440000",
     "key_type": "CPF",
     "key_value": "12345678900",
     "status": "ACTIVE",
     "external_id": "bacen-uuid-123",
     "created_at": "2025-10-25T10:00:00Z",
     "updated_at": "2025-10-25T10:00:01Z"
   }
   ```

6. Verify audit log:
   ```sql
   SELECT * FROM dict.audit_logs
   WHERE entity_type = 'entry'
   AND entity_id = '550e8400-e29b-41d4-a716-446655440000'
   AND action = 'CREATE';
   ```
   Expected: 1 row with correct user_id, timestamp

#### Expected Result
- ✅ Entry created successfully with status PENDING
- ✅ Entry transitions to ACTIVE within 2 seconds
- ✅ Entry has valid external_id from Bacen
- ✅ Audit log created
- ✅ User notification sent
- ✅ Total time < 2.5 seconds (p95)

#### Actual Result
[To be filled during execution]

#### Status
⬜ Not Run | 🟡 In Progress | ✅ Pass | ❌ Fail | 🚫 Blocked

---

### TC-001-002: Create Entry - EMAIL (Happy Path)

**Priority**: P0 (Critical)
**Type**: E2E Functional
**Estimated Duration**: 2-3 seconds

#### Preconditions
- User is authenticated with valid JWT token
- User has scope `dict:write`
- Account exists and is ACTIVE
- Email key does not exist in DICT

#### Test Data
```json
{
  "key_type": "EMAIL",
  "key_value": "user@example.com",
  "account": {
    "ispb": "12345678",
    "account_number": "123456",
    "branch": "0001",
    "account_type": "CACC"
  }
}
```

#### BDD Format
```gherkin
Given the user is authenticated
And the user has scope "dict:write"
And the email "user@example.com" does not exist in DICT
When the user sends POST /api/v1/keys with EMAIL data
Then the API returns 201 Created
And the entry status is "PENDING"
And within 2 seconds the entry status changes to "ACTIVE"
```

#### Steps
1. Authenticate user (same as TC-001-001)
2. Create EMAIL entry:
   ```bash
   POST /api/v1/keys
   Authorization: Bearer {access_token}

   {
     "key_type": "EMAIL",
     "key_value": "user@example.com",
     "account": { ... }
   }
   ```
3. Verify 201 Created response
4. Poll until status = ACTIVE (max 2 seconds)
5. Verify external_id exists

#### Expected Result
- ✅ Email entry created successfully
- ✅ Entry transitions to ACTIVE
- ✅ Valid external_id from Bacen

#### Status
⬜ Not Run

---

### TC-001-003: Create Entry - PHONE (Happy Path)

**Priority**: P0 (Critical)
**Type**: E2E Functional

#### Test Data
```json
{
  "key_type": "PHONE",
  "key_value": "+5511999999999",
  "account": {
    "ispb": "12345678",
    "account_number": "123456",
    "branch": "0001",
    "account_type": "CACC"
  }
}
```

#### BDD Format
```gherkin
Given the user is authenticated
And the phone "+5511999999999" does not exist in DICT
When the user creates a PHONE entry
Then the entry is created with status PENDING
And within 2 seconds becomes ACTIVE
```

#### Expected Result
- ✅ Phone entry created successfully
- ✅ Phone format E.164 validated
- ✅ Entry synced with Bacen

#### Status
⬜ Not Run

---

### TC-001-004: Create Entry - CNPJ (Happy Path)

**Priority**: P0 (Critical)
**Type**: E2E Functional

#### Test Data
```json
{
  "key_type": "CNPJ",
  "key_value": "12345678000190",
  "account": {
    "ispb": "12345678",
    "account_number": "654321",
    "branch": "0002",
    "account_type": "CACC"
  }
}
```

#### BDD Format
```gherkin
Given the user is authenticated
And the CNPJ "12345678000190" is valid and does not exist
When the user creates a CNPJ entry
Then the entry is created successfully
```

#### Expected Result
- ✅ CNPJ entry created
- ✅ CNPJ format validated (14 digits)
- ✅ Entry synced with Bacen

#### Status
⬜ Not Run

---

### TC-001-005: Create Entry - EVP (Random Key)

**Priority**: P0 (Critical)
**Type**: E2E Functional

#### Test Data
```json
{
  "key_type": "EVP",
  "key_value": "550e8400-e29b-41d4-a716-446655440000",
  "account": {
    "ispb": "12345678",
    "account_number": "123456",
    "branch": "0001",
    "account_type": "CACC"
  }
}
```

#### BDD Format
```gherkin
Given the user is authenticated
And the EVP key is a valid UUID v4
When the user creates an EVP entry
Then the entry is created successfully
```

#### Expected Result
- ✅ EVP entry created
- ✅ UUID v4 format validated
- ✅ Entry synced with Bacen

#### Status
⬜ Not Run

---

## P0 Test Cases (Error Handling - Critical)

### TC-001-006: Invalid Key Format - CPF Invalid

**Priority**: P0 (Critical)
**Type**: Negative Test

#### Test Data
```json
{
  "key_type": "CPF",
  "key_value": "11111111111",
  "account": { ... }
}
```

#### BDD Format
```gherkin
Given the user is authenticated
When the user attempts to create a CPF entry with invalid CPF "11111111111"
Then the API returns 400 Bad Request
And the error code is "INVALID_KEY_FORMAT"
And the error message contains "Invalid CPF format"
```

#### Steps
1. Authenticate user
2. Send POST /api/v1/keys with invalid CPF (all same digits)
3. Verify response:
   ```json
   {
     "error": "INVALID_KEY_FORMAT",
     "message": "Invalid CPF format: CPF cannot have all same digits",
     "key_type": "CPF",
     "key_value": "11111111111"
   }
   ```

#### Expected Result
- ✅ API returns 400 Bad Request
- ✅ Error code is "INVALID_KEY_FORMAT"
- ✅ Clear error message provided
- ✅ Entry NOT created in database

#### Status
⬜ Not Run

---

### TC-001-007: Invalid Key Format - EMAIL Invalid

**Priority**: P0 (Critical)
**Type**: Negative Test

#### Test Data
```json
{
  "key_type": "EMAIL",
  "key_value": "invalid-email",
  "account": { ... }
}
```

#### BDD Format
```gherkin
Given the user is authenticated
When the user attempts to create an EMAIL entry with invalid format "invalid-email"
Then the API returns 400 Bad Request
And the error code is "INVALID_KEY_FORMAT"
```

#### Expected Result
- ✅ 400 Bad Request returned
- ✅ Error: Invalid email format (missing @)

#### Status
⬜ Not Run

---

### TC-001-008: Invalid Key Format - PHONE Invalid

**Priority**: P0 (Critical)
**Type**: Negative Test

#### Test Data
```json
{
  "key_type": "PHONE",
  "key_value": "11999999999",
  "account": { ... }
}
```

#### BDD Format
```gherkin
Given the user is authenticated
When the user creates PHONE entry without E.164 format (missing +55)
Then the API returns 400 Bad Request
And the error message is "Phone must be in E.164 format: +5511999999999"
```

#### Expected Result
- ✅ 400 Bad Request
- ✅ Error: Phone must start with +55 (E.164)

#### Status
⬜ Not Run

---

### TC-001-009: Duplicate Key - CPF Already Exists

**Priority**: P0 (Critical)
**Type**: Negative Test

#### Preconditions
- CPF "12345678900" already exists in DICT (associated with another account)

#### Test Data
```json
{
  "key_type": "CPF",
  "key_value": "12345678900",
  "account": { ... }
}
```

#### BDD Format
```gherkin
Given the CPF "12345678900" already exists in DICT
When the user attempts to create an entry with this CPF
Then the API returns 409 Conflict
And the error code is "KEY_ALREADY_EXISTS"
And the response suggests claiming the key
```

#### Steps
1. Ensure CPF "12345678900" exists (run TC-001-001 first)
2. Attempt to create same CPF with different account
3. Verify response:
   ```json
   {
     "error": "KEY_ALREADY_EXISTS",
     "message": "This key is already registered. You can claim it if it belongs to you.",
     "key_type": "CPF",
     "key_value": "12345678900",
     "actions": {
       "claim": "/api/v1/claims?entry_id={existing_entry_id}"
     }
   }
   ```

#### Expected Result
- ✅ 409 Conflict returned
- ✅ Clear error message
- ✅ Link to claim workflow provided
- ✅ Entry NOT created

#### Status
⬜ Not Run

---

### TC-001-010: Invalid Account - Account Does Not Exist

**Priority**: P0 (Critical)
**Type**: Negative Test

#### Test Data
```json
{
  "key_type": "CPF",
  "key_value": "98765432100",
  "account": {
    "ispb": "12345678",
    "account_number": "000000",
    "branch": "9999",
    "account_type": "CACC"
  }
}
```

#### BDD Format
```gherkin
Given the user is authenticated
When the user attempts to create an entry with non-existent account "000000"
Then the API returns 400 Bad Request
And the error code is "INVALID_ACCOUNT"
And the error message is "Account does not exist or is not active"
```

#### Expected Result
- ✅ 400 Bad Request
- ✅ Error: Account not found
- ✅ Entry NOT created

#### Status
⬜ Not Run

---

### TC-001-011: Invalid Account - Account Inactive

**Priority**: P0 (Critical)
**Type**: Negative Test

#### Test Data
```json
{
  "key_type": "CPF",
  "key_value": "98765432100",
  "account": {
    "ispb": "12345678",
    "account_number": "999999",
    "branch": "0999",
    "account_type": "CACC"
  }
}
```

#### Preconditions
- Account "999999" exists but has status INACTIVE

#### BDD Format
```gherkin
Given the account "999999" exists but is INACTIVE
When the user attempts to create an entry with this account
Then the API returns 400 Bad Request
And the error code is "INVALID_ACCOUNT"
```

#### Expected Result
- ✅ 400 Bad Request
- ✅ Error: Account is not active
- ✅ Entry NOT created

#### Status
⬜ Not Run

---

## P0 Test Cases (Authentication & Authorization)

### TC-001-012: Authentication Error - Invalid JWT Token

**Priority**: P0 (Critical)
**Type**: Security Test

#### Test Data
```bash
Authorization: Bearer invalid-token-xyz123
```

#### BDD Format
```gherkin
Given the user has an invalid JWT token
When the user attempts to create an entry
Then the API returns 401 Unauthorized
And the error message is "Invalid or expired token"
```

#### Steps
1. Send POST /api/v1/keys with invalid token:
   ```bash
   POST /api/v1/keys
   Authorization: Bearer invalid-token-xyz123

   {
     "key_type": "CPF",
     "key_value": "12345678900",
     "account": { ... }
   }
   ```
2. Verify response:
   ```json
   {
     "error": "Unauthorized",
     "message": "Invalid or expired token",
     "timestamp": "2025-10-25T10:00:00Z"
   }
   ```

#### Expected Result
- ✅ 401 Unauthorized
- ✅ Clear error message
- ✅ No entry created

#### Status
⬜ Not Run

---

### TC-001-013: Authentication Error - Expired JWT Token

**Priority**: P0 (Critical)
**Type**: Security Test

#### Preconditions
- User has a JWT token that expired 1 hour ago

#### BDD Format
```gherkin
Given the user has an expired JWT token
When the user attempts to create an entry
Then the API returns 401 Unauthorized
And the error message indicates token expiration
```

#### Expected Result
- ✅ 401 Unauthorized
- ✅ Error: Token expired
- ✅ Suggest token refresh

#### Status
⬜ Not Run

---

### TC-001-014: Authentication Error - Missing Authorization Header

**Priority**: P0 (Critical)
**Type**: Security Test

#### BDD Format
```gherkin
Given the user does not provide Authorization header
When the user attempts to create an entry
Then the API returns 401 Unauthorized
And the error message is "Missing authorization header"
```

#### Steps
1. Send POST /api/v1/keys without Authorization header:
   ```bash
   POST /api/v1/keys
   Content-Type: application/json

   {
     "key_type": "CPF",
     "key_value": "12345678900",
     "account": { ... }
   }
   ```

#### Expected Result
- ✅ 401 Unauthorized
- ✅ Error: Missing authorization

#### Status
⬜ Not Run

---

### TC-001-015: Authorization Error - Missing Scope

**Priority**: P0 (Critical)
**Type**: Security Test

#### Preconditions
- User "readonly@lbpay.com.br" has only `dict:read` scope (no `dict:write`)

#### Test Data
- User: readonly@lbpay.com.br
- Scopes: [dict:read]

#### BDD Format
```gherkin
Given the user has valid JWT token
But the user does not have scope "dict:write"
When the user attempts to create an entry
Then the API returns 403 Forbidden
And the error message is "Required scope: dict:write"
```

#### Steps
1. Authenticate as readonly user
2. Attempt to create entry
3. Verify response:
   ```json
   {
     "error": "Forbidden",
     "message": "Required scope: dict:write",
     "required_scopes": ["dict:write"],
     "user_scopes": ["dict:read"]
   }
   ```

#### Expected Result
- ✅ 403 Forbidden
- ✅ Clear error message
- ✅ No entry created

#### Status
⬜ Not Run

---

## P1 Test Cases (Edge Cases)

### TC-001-016: Key Limit - CPF Maximum Keys (5 keys)

**Priority**: P1 (High)
**Type**: Business Rule Test

#### Preconditions
- User already has 5 CPF keys registered

#### BDD Format
```gherkin
Given the user has 5 CPF keys registered (Bacen limit)
When the user attempts to create a 6th CPF key
Then the API returns 400 Bad Request
And the error code is "KEY_LIMIT_REACHED"
And the error message is "CPF can have maximum 5 keys"
```

#### Expected Result
- ✅ 400 Bad Request
- ✅ Error: Key limit reached
- ✅ Message shows current count (5/5)

#### Status
⬜ Not Run

---

### TC-001-017: Key Limit - CNPJ Maximum Keys (20 keys)

**Priority**: P1 (High)
**Type**: Business Rule Test

#### Preconditions
- CNPJ already has 20 keys registered

#### BDD Format
```gherkin
Given the CNPJ has 20 keys registered (Bacen limit)
When attempting to create 21st key
Then the API returns 400 Bad Request
And the error code is "KEY_LIMIT_REACHED"
```

#### Expected Result
- ✅ 400 Bad Request
- ✅ Error: CNPJ limit is 20 keys

#### Status
⬜ Not Run

---

### TC-001-018: Concurrent Requests - Same Key

**Priority**: P1 (High)
**Type**: Race Condition Test

#### Test Data
- 2 concurrent requests for same CPF "11122233344"

#### BDD Format
```gherkin
Given two users attempt to create the same CPF key simultaneously
When both requests arrive at the same time
Then only one request succeeds with 201 Created
And the other request fails with 409 Conflict
```

#### Steps
1. Send 2 parallel POST requests for same CPF
2. Verify:
   - Request A: 201 Created OR 409 Conflict
   - Request B: 409 Conflict OR 201 Created
   - Only ONE entry created in database

#### Expected Result
- ✅ One request succeeds (201)
- ✅ One request fails (409)
- ✅ Database has only 1 entry

#### Status
⬜ Not Run

---

### TC-001-019: Special Characters - Email with Plus Sign

**Priority**: P1 (High)
**Type**: Edge Case

#### Test Data
```json
{
  "key_type": "EMAIL",
  "key_value": "user+tag@example.com",
  "account": { ... }
}
```

#### BDD Format
```gherkin
Given the email "user+tag@example.com" is valid RFC 5322
When the user creates EMAIL entry
Then the entry is created successfully
And the email is stored as-is (not normalized)
```

#### Expected Result
- ✅ Email with + accepted
- ✅ Email not normalized
- ✅ Entry created successfully

#### Status
⬜ Not Run

---

### TC-001-020: UTF-8 Characters - Email with Accents

**Priority**: P1 (High)
**Type**: Edge Case

#### Test Data
```json
{
  "key_type": "EMAIL",
  "key_value": "josé@example.com",
  "account": { ... }
}
```

#### BDD Format
```gherkin
Given the email contains UTF-8 characters (accents)
When the user creates EMAIL entry
Then the email is properly encoded and stored
```

#### Expected Result
- ✅ UTF-8 email accepted
- ✅ Stored correctly with encoding

#### Status
⬜ Not Run

---

## P1 Test Cases (Performance)

### TC-001-021: Performance - Response Time p95 < 300ms

**Priority**: P1 (High)
**Type**: Performance Test

#### Test Setup
- Load: 100 requests/second
- Duration: 60 seconds
- Total requests: 6000

#### BDD Format
```gherkin
Given the system is under normal load (100 TPS)
When 6000 CreateEntry requests are sent over 60 seconds
Then p50 response time is < 150ms
And p95 response time is < 300ms
And p99 response time is < 500ms
```

#### Expected Result
- ✅ p50 < 150ms
- ✅ p95 < 300ms
- ✅ p99 < 500ms
- ✅ 0 errors

#### Status
⬜ Not Run

---

### TC-001-022: Performance - Throughput 1000 TPS

**Priority**: P1 (High)
**Type**: Load Test

#### Test Setup
- Ramp-up: 0 → 1000 TPS over 5 minutes
- Sustained: 1000 TPS for 10 minutes
- Ramp-down: 1000 → 0 TPS over 2 minutes

#### BDD Format
```gherkin
Given the system is load tested with 1000 TPS
When CreateEntry requests are sent at 1000 TPS for 10 minutes
Then the success rate is > 99%
And p95 latency is < 2 seconds
And no system errors occur
```

#### Expected Result
- ✅ Success rate > 99%
- ✅ p95 < 2s
- ✅ CPU < 80%
- ✅ Memory stable

#### Status
⬜ Not Run

---

## P2 Test Cases (Integration)

### TC-001-023: Integration - Bacen Timeout Retry

**Priority**: P2 (Medium)
**Type**: Integration Test

#### Test Setup
- Mock Bacen to timeout on first 2 attempts
- Succeed on 3rd attempt

#### BDD Format
```gherkin
Given Bacen times out on first 2 requests
When CreateEntry workflow executes
Then the workflow retries 3 times with exponential backoff
And succeeds on 3rd attempt
And the entry status becomes ACTIVE
```

#### Expected Result
- ✅ 3 retry attempts executed
- ✅ Backoff intervals: 100ms, 500ms, 2s
- ✅ Entry becomes ACTIVE
- ✅ Total time < 5s

#### Status
⬜ Not Run

---

### TC-001-024: Integration - Bacen Returns Error

**Priority**: P2 (Medium)
**Type**: Integration Test

#### Test Setup
- Mock Bacen to return SOAP Fault (e.g., ENTRY_INVALID)

#### BDD Format
```gherkin
Given Bacen returns SOAP Fault error
When CreateEntry workflow executes
Then the workflow fails after 3 retries
And the entry status remains PENDING
And an alert is triggered
```

#### Expected Result
- ✅ Workflow fails gracefully
- ✅ Entry status = PENDING
- ✅ Alert sent to DevOps

#### Status
⬜ Not Run

---

### TC-001-025: Integration - Redis Cache Hit

**Priority**: P2 (Medium)
**Type**: Integration Test

#### Preconditions
- Entry "CPF:12345678900" exists in Redis cache (TTL 5 min)

#### BDD Format
```gherkin
Given the entry exists in Redis cache
When CreateEntry workflow checks cache
Then the workflow skips Bacen call
And uses cached data
And the workflow completes in < 100ms
```

#### Expected Result
- ✅ Cache hit detected
- ✅ Bacen call skipped
- ✅ Fast completion (< 100ms)

#### Status
⬜ Not Run

---

## Test Execution Summary

### Coverage Matrix

| Category | P0 Tests | P1 Tests | P2 Tests | Total |
|----------|----------|----------|----------|-------|
| Happy Path | 5 | 0 | 0 | 5 |
| Error Handling | 10 | 0 | 0 | 10 |
| Edge Cases | 0 | 5 | 0 | 5 |
| Performance | 0 | 2 | 0 | 2 |
| Integration | 0 | 0 | 3 | 3 |
| **Total** | **15** | **7** | **3** | **25** |

### Priority Distribution
- P0 (Critical): 15 tests (60%)
- P1 (High): 7 tests (28%)
- P2 (Medium): 3 tests (12%)

### Test Types
- E2E Functional: 10 tests
- Negative Tests: 7 tests
- Security Tests: 4 tests
- Performance Tests: 2 tests
- Integration Tests: 2 tests

---

## Automation Scripts

### Jest E2E Test Example
```javascript
// tests/e2e/create-entry.test.js
describe('CreateEntry E2E', () => {
  let accessToken;

  beforeAll(async () => {
    // Authenticate
    const authResponse = await request(AUTH_URL)
      .post('/auth/login')
      .send({
        email: 'test.user@lbpay.com.br',
        password: 'Test@1234'
      });

    accessToken = authResponse.body.access_token;
  });

  describe('TC-001-001: Create CPF Entry (Happy Path)', () => {
    it('should create CPF entry successfully', async () => {
      // Step 1: Create entry
      const createResponse = await request(DICT_API_URL)
        .post('/api/v1/keys')
        .set('Authorization', `Bearer ${accessToken}`)
        .send({
          key_type: 'CPF',
          key_value: '12345678900',
          account: {
            ispb: '12345678',
            account_number: '123456',
            branch: '0001',
            account_type: 'CACC'
          }
        });

      // Verify synchronous response
      expect(createResponse.status).toBe(201);
      expect(createResponse.body.entry_id).toBeDefined();
      expect(createResponse.body.status).toBe('PENDING');

      const entryId = createResponse.body.entry_id;

      // Step 2: Poll for ACTIVE status (max 2s)
      await waitFor(async () => {
        const getResponse = await request(DICT_API_URL)
          .get(`/api/v1/keys/CPF/12345678900`)
          .set('Authorization', `Bearer ${accessToken}`);

        expect(getResponse.status).toBe(200);
        expect(getResponse.body.status).toBe('ACTIVE');
        expect(getResponse.body.external_id).toBeDefined();
      }, { timeout: 2000, interval: 200 });
    });
  });

  describe('TC-001-009: Duplicate Key Error', () => {
    it('should return 409 Conflict for duplicate CPF', async () => {
      // Create first entry
      await request(DICT_API_URL)
        .post('/api/v1/keys')
        .set('Authorization', `Bearer ${accessToken}`)
        .send({
          key_type: 'CPF',
          key_value: '99988877766',
          account: { ... }
        });

      // Attempt duplicate
      const duplicateResponse = await request(DICT_API_URL)
        .post('/api/v1/keys')
        .set('Authorization', `Bearer ${accessToken}`)
        .send({
          key_type: 'CPF',
          key_value: '99988877766',
          account: { ... }
        });

      expect(duplicateResponse.status).toBe(409);
      expect(duplicateResponse.body.error).toBe('KEY_ALREADY_EXISTS');
    });
  });
});
```

### k6 Performance Test
```javascript
// tests/performance/create-entry-load.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '5m', target: 1000 },  // Ramp-up to 1000 TPS
    { duration: '10m', target: 1000 }, // Sustain 1000 TPS
    { duration: '2m', target: 0 },     // Ramp-down
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'], // p95 < 2s
    http_req_failed: ['rate<0.01'],    // Error rate < 1%
  },
};

export default function () {
  const url = 'https://dict-api-test.lbpay.com.br/api/v1/keys';
  const payload = JSON.stringify({
    key_type: 'CPF',
    key_value: generateRandomCPF(),
    account: {
      ispb: '12345678',
      account_number: '123456',
      branch: '0001',
      account_type: 'CACC'
    }
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${__ENV.ACCESS_TOKEN}`,
    },
  };

  const res = http.post(url, payload, params);

  check(res, {
    'status is 201': (r) => r.status === 201,
    'has entry_id': (r) => JSON.parse(r.body).entry_id !== undefined,
  });

  sleep(1);
}

function generateRandomCPF() {
  // Generate valid random CPF for testing
  return Math.floor(Math.random() * 90000000000 + 10000000000).toString();
}
```

---

## Entry/Exit Criteria

### Entry Criteria
- ✅ Core DICT API deployed to test environment
- ✅ Test database populated with test data
- ✅ Bacen mock service available
- ✅ Test users created with proper scopes
- ✅ Test accounts created in Ledger

### Exit Criteria
- ✅ All P0 tests passed (15/15)
- ✅ 90%+ P1 tests passed (6/7 minimum)
- ✅ No P0 defects open
- ✅ Performance benchmarks met (p95 < 300ms, 1000 TPS)
- ✅ Code coverage > 80%

---

## Defect Tracking

### Known Issues
None at document creation time.

### Defect Template
```markdown
**Defect ID**: DEF-001
**Test Case**: TC-001-009
**Severity**: P0
**Status**: Open
**Description**: Duplicate key returns 500 instead of 409
**Steps to Reproduce**: Run TC-001-009
**Expected**: 409 Conflict
**Actual**: 500 Internal Server Error
**Assigned To**: Backend Team
**Created**: 2025-10-25
```

---

## Glossary

- **BDD**: Behavior-Driven Development (Given/When/Then)
- **E2E**: End-to-End test
- **JWT**: JSON Web Token
- **P0/P1/P2**: Priority levels (Critical/High/Medium)
- **TPS**: Transactions Per Second
- **RBAC**: Role-Based Access Control
- **EVP**: Chave Aleatória (Random PIX Key)
- **CACC**: Conta Corrente (Checking Account)
- **SVGS**: Conta Poupança (Savings Account)

---

**Última Revisão**: 2025-10-25
**Aprovado por**: QA Lead
**Próxima Revisão**: Sprint review
