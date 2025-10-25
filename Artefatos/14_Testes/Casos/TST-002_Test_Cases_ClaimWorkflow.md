# TST-002: Test Cases - ClaimWorkflow (30 days)

**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: QA Team
**Status**: ✅ Completo

---

## Sumário Executivo

Este documento apresenta os **test cases completos** para o **ClaimWorkflow de 30 dias**, o fluxo mais crítico e complexo do sistema DICT LBPay.

**Objetivo**: Garantir cobertura de testes abrangente para todos os cenários de reivindicação de chave PIX, incluindo criação, confirmação, cancelamento e expiração automática.

**Cobertura**:
- Create Claim (P0)
- Confirm Claim (P0)
- Cancel Claim (P0)
- Expire Claim - Auto-confirm (P0)
- ISPB Validation (P0)
- Period Validation (P0)
- Concurrent Claims (P0)
- Edge Cases (P1)

**Referências**:
- [DIA-006: Sequence Diagram - ClaimWorkflow](../../02_Arquitetura/Diagramas/DIA-006_Sequence_Claim_Workflow.md)
- [TEC-003: RSFN Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)

---

## Test Environment Setup

### Prerequisites
```yaml
Environment: test
Core DICT API: https://dict-api-test.lbpay.com.br
Temporal Server: test.temporal.lbpay.com.br
Temporal UI: https://temporal-ui-test.lbpay.com.br
Bridge Mock: Bacen mock enabled
Time Acceleration: Enabled (30 days → 30 seconds for testing)
```

### Test Users
```yaml
Claimer User (Banco B):
  email: claimer@bancob.com.br
  password: Test@1234
  ispb: 87654321
  scopes: [dict:read, dict:write, dict:claim]

Owner User (Banco A):
  email: owner@bancoa.com.br
  password: Test@1234
  ispb: 12345678
  scopes: [dict:read, dict:write, dict:claim_respond]
```

### Test Entries
```json
{
  "entry_1": {
    "entry_id": "550e8400-e29b-41d4-a716-446655440000",
    "key_type": "CPF",
    "key_value": "12345678900",
    "owner_ispb": "12345678",
    "owner_account": "123456",
    "status": "ACTIVE"
  },
  "entry_2": {
    "entry_id": "660e8400-e29b-41d4-a716-446655440111",
    "key_type": "EMAIL",
    "key_value": "user@example.com",
    "owner_ispb": "12345678",
    "owner_account": "123456",
    "status": "ACTIVE"
  }
}
```

---

## P0 Test Cases (Critical Path - Create Claim)

### TC-002-001: Create Claim - Happy Path

**Priority**: P0 (Critical)
**Type**: E2E Functional
**Estimated Duration**: 200-500ms (synchronous)

#### Preconditions
- Entry "CPF:12345678900" exists and is ACTIVE
- Entry belongs to Owner (ISPB 12345678)
- Claimer ISPB (87654321) is different from Owner ISPB
- No OPEN claim exists for this entry

#### Test Data
```json
{
  "entry_id": "550e8400-e29b-41d4-a716-446655440000",
  "claimer_account": {
    "ispb": "87654321",
    "account_number": "654321",
    "branch": "0002",
    "account_type": "CACC",
    "holder_document": "98765432100",
    "holder_name": "João Silva"
  },
  "completion_period_days": 30
}
```

#### BDD Format (Given/When/Then)
```gherkin
Given the claimer is authenticated with valid JWT
And the claimer has scope "dict:claim"
And the entry "CPF:12345678900" exists and is ACTIVE
And the entry belongs to Owner ISPB 12345678
And the claimer ISPB 87654321 is different from owner
And no OPEN claim exists for this entry
When the claimer sends POST /api/v1/claims
Then the API returns 201 Created
And the claim status is "OPEN"
And the expires_at is NOW() + 30 days
And a Temporal workflow "ClaimWorkflow" is started
And the Bacen is notified via Bridge
And the owner receives email notification
And the claim is persisted with bacen_claim_id
```

#### Steps
1. Authenticate as claimer user:
   ```bash
   POST /auth/login
   {
     "email": "claimer@bancob.com.br",
     "password": "Test@1234"
   }
   ```
   Expected: 200 OK with `access_token`

2. Create claim:
   ```bash
   POST /api/v1/claims
   Authorization: Bearer {claimer_access_token}
   Content-Type: application/json

   {
     "entry_id": "550e8400-e29b-41d4-a716-446655440000",
     "claimer_account": {
       "ispb": "87654321",
       "account_number": "654321",
       "branch": "0002",
       "account_type": "CACC",
       "holder_document": "98765432100",
       "holder_name": "João Silva"
     },
     "completion_period_days": 30
   }
   ```

3. Verify synchronous response:
   Expected: 201 Created
   ```json
   {
     "claim_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
     "entry_id": "550e8400-e29b-41d4-a716-446655440000",
     "claimer_ispb": "87654321",
     "owner_ispb": "12345678",
     "status": "OPEN",
     "completion_period_days": 30,
     "expires_at": "2025-11-24T10:00:00Z",
     "created_at": "2025-10-25T10:00:00Z"
   }
   ```

4. Verify Temporal workflow started:
   ```bash
   # Check Temporal UI or query Temporal API
   GET https://temporal-ui-test.lbpay.com.br/workflows?workflow_id=claim-7c9e6679
   ```
   Expected: Workflow "ClaimWorkflow" in RUNNING state

5. Verify database state:
   ```sql
   SELECT * FROM dict.claims
   WHERE claim_id = '7c9e6679-7425-40de-944b-e07fc1f90ae7';
   ```
   Expected:
   - status = 'OPEN'
   - bacen_claim_id IS NOT NULL
   - expires_at = NOW() + 30 days

6. Verify owner received notification:
   ```bash
   # Check email logs or notification service
   GET /notifications/logs?user_id={owner_user_id}&template=claim_received
   ```
   Expected: 1 notification sent

#### Expected Result
- ✅ Claim created with status OPEN
- ✅ expires_at = NOW() + 30 days
- ✅ Temporal workflow started
- ✅ Bacen notified (bacen_claim_id exists)
- ✅ Owner notified via email
- ✅ Total time < 500ms

#### Actual Result
[To be filled during execution]

#### Status
⬜ Not Run | 🟡 In Progress | ✅ Pass | ❌ Fail | 🚫 Blocked

---

### TC-002-002: Create Claim - Invalid ISPB (Same as Owner)

**Priority**: P0 (Critical)
**Type**: Negative Test

#### Preconditions
- Entry exists with owner_ispb = 12345678
- Claimer attempts to create claim with same ISPB 12345678

#### Test Data
```json
{
  "entry_id": "550e8400-e29b-41d4-a716-446655440000",
  "claimer_account": {
    "ispb": "12345678",
    "account_number": "999999",
    "branch": "0099",
    "account_type": "CACC"
  },
  "completion_period_days": 30
}
```

#### BDD Format
```gherkin
Given the entry belongs to Owner ISPB 12345678
When the claimer attempts to create claim with same ISPB 12345678
Then the API returns 403 Forbidden
And the error code is "INVALID_CLAIMER_ISPB"
And the error message is "Claimer ISPB must be different from Owner ISPB"
And no claim is created
```

#### Steps
1. Authenticate as user with ISPB 12345678
2. Attempt to create claim for entry owned by same ISPB
3. Verify response:
   ```json
   {
     "error": "INVALID_CLAIMER_ISPB",
     "message": "Claimer ISPB must be different from Owner ISPB",
     "claimer_ispb": "12345678",
     "owner_ispb": "12345678",
     "timestamp": "2025-10-25T10:00:00Z"
   }
   ```

#### Expected Result
- ✅ 403 Forbidden returned
- ✅ Clear error message
- ✅ No claim created in database
- ✅ No Temporal workflow started

#### Status
⬜ Not Run

---

### TC-002-003: Create Claim - Invalid Period (Not 30 days)

**Priority**: P0 (Critical)
**Type**: Negative Test

#### Test Data
```json
{
  "entry_id": "550e8400-e29b-41d4-a716-446655440000",
  "claimer_account": { ... },
  "completion_period_days": 15
}
```

#### BDD Format
```gherkin
Given the claimer is authenticated
When the claimer sends completion_period_days = 15 (not 30)
Then the API returns 400 Bad Request
And the error code is "INVALID_COMPLETION_PERIOD"
And the error message is "Completion period must be exactly 30 days (Bacen regulation)"
```

#### Expected Result
- ✅ 400 Bad Request
- ✅ Error: Period must be 30 days
- ✅ Reference to Bacen regulation

#### Status
⬜ Not Run

---

### TC-002-004: Create Claim - Entry Not ACTIVE

**Priority**: P0 (Critical)
**Type**: Negative Test

#### Preconditions
- Entry exists but has status PENDING or DELETED

#### BDD Format
```gherkin
Given the entry status is PENDING (not ACTIVE)
When the claimer attempts to create claim
Then the API returns 400 Bad Request
And the error code is "ENTRY_NOT_ACTIVE"
And the error message is "Only ACTIVE entries can be claimed"
```

#### Expected Result
- ✅ 400 Bad Request
- ✅ Error: Entry must be ACTIVE
- ✅ No claim created

#### Status
⬜ Not Run

---

### TC-002-005: Create Claim - Duplicate Claim (Already OPEN)

**Priority**: P0 (Critical)
**Type**: Negative Test

#### Preconditions
- Entry already has an OPEN claim

#### BDD Format
```gherkin
Given the entry has an OPEN claim (claim_id: abc123)
When another claimer attempts to create a new claim for same entry
Then the API returns 409 Conflict
And the error code is "CLAIM_ALREADY_EXISTS"
And the error message contains the existing claim_id
```

#### Expected Result
- ✅ 409 Conflict
- ✅ Error: Claim already exists
- ✅ Existing claim_id returned
- ✅ No new claim created

#### Status
⬜ Not Run

---

## P0 Test Cases (Confirm Claim)

### TC-002-006: Confirm Claim - Owner Confirms Before 30 Days

**Priority**: P0 (Critical)
**Type**: E2E Functional
**Estimated Duration**: 2-3 seconds (with Bacen sync)

#### Preconditions
- Claim exists with status OPEN
- Claim was created < 30 days ago (e.g., 5 days ago)
- User is the owner of the entry

#### BDD Format
```gherkin
Given a claim exists with status OPEN
And the claim was created 5 days ago
And the owner is authenticated
And the owner has scope "dict:claim_respond"
When the owner sends POST /api/v1/claims/{claim_id}/confirm
Then the API returns 200 OK
And the claim status changes to CONFIRMED
And the Temporal workflow receives "confirm" signal
And the workflow cancels the 30-day timer
And the Bacen is notified via Bridge (CompleteClaim)
And the entry is transferred to claimer account
And both owner and claimer receive notifications
And the Temporal workflow completes successfully
```

#### Steps
1. Setup: Create claim (use TC-002-001)
   - Claim ID: 7c9e6679-7425-40de-944b-e07fc1f90ae7
   - Status: OPEN
   - Expires in 25 days

2. Authenticate as owner:
   ```bash
   POST /auth/login
   {
     "email": "owner@bancoa.com.br",
     "password": "Test@1234"
   }
   ```

3. Confirm claim:
   ```bash
   POST /api/v1/claims/7c9e6679-7425-40de-944b-e07fc1f90ae7/confirm
   Authorization: Bearer {owner_access_token}
   ```

4. Verify synchronous response:
   Expected: 200 OK
   ```json
   {
     "claim_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
     "status": "CONFIRMED",
     "confirmed_at": "2025-10-25T10:00:00Z",
     "confirmed_by": "{owner_user_id}"
   }
   ```

5. Verify database state (within 2 seconds):
   ```sql
   SELECT * FROM dict.claims
   WHERE claim_id = '7c9e6679-7425-40de-944b-e07fc1f90ae7';
   ```
   Expected:
   - status = 'COMPLETED' (after workflow finishes)
   - bacen_status = 'COMPLETED'
   - completed_at IS NOT NULL

6. Verify entry transferred:
   ```sql
   SELECT * FROM dict.entries
   WHERE entry_id = '550e8400-e29b-41d4-a716-446655440000';
   ```
   Expected:
   - account_id = {claimer_account_id}
   - owner_ispb = '87654321' (changed to claimer)
   - updated_at IS NOT NULL

7. Verify Temporal workflow completed:
   ```bash
   GET https://temporal-ui-test.lbpay.com.br/workflows?workflow_id=claim-7c9e6679
   ```
   Expected: Workflow status = COMPLETED

8. Verify notifications sent:
   - Owner: "Reivindicação confirmada. Chave transferida."
   - Claimer: "Reivindicação concluída. Chave agora é sua!"

#### Expected Result
- ✅ Claim status = CONFIRMED → COMPLETED
- ✅ Entry transferred to claimer
- ✅ Bacen notified (CompleteClaimRequest sent)
- ✅ Both users notified
- ✅ Temporal workflow completed
- ✅ Total time < 3 seconds

#### Status
⬜ Not Run

---

### TC-002-007: Confirm Claim - Unauthorized User

**Priority**: P0 (Critical)
**Type**: Security Test

#### Preconditions
- Claim exists with status OPEN
- User is NOT the owner (e.g., a random user or the claimer)

#### BDD Format
```gherkin
Given a claim exists owned by user A
And user B (not the owner) is authenticated
When user B attempts to confirm the claim
Then the API returns 403 Forbidden
And the error code is "UNAUTHORIZED_CLAIM_ACTION"
And the claim status remains OPEN
```

#### Expected Result
- ✅ 403 Forbidden
- ✅ Error: Only owner can confirm
- ✅ Claim status unchanged

#### Status
⬜ Not Run

---

### TC-002-008: Confirm Claim - Already Confirmed

**Priority**: P0 (Critical)
**Type**: Edge Case

#### Preconditions
- Claim already has status CONFIRMED or COMPLETED

#### BDD Format
```gherkin
Given a claim has status CONFIRMED
When the owner attempts to confirm again
Then the API returns 409 Conflict
And the error code is "CLAIM_ALREADY_RESOLVED"
And the error message is "Claim has already been confirmed"
```

#### Expected Result
- ✅ 409 Conflict
- ✅ Error: Claim already resolved
- ✅ No state change

#### Status
⬜ Not Run

---

## P0 Test Cases (Cancel Claim)

### TC-002-009: Cancel Claim - Owner Cancels Before 30 Days

**Priority**: P0 (Critical)
**Type**: E2E Functional
**Estimated Duration**: 2-3 seconds

#### Preconditions
- Claim exists with status OPEN
- Claim was created < 30 days ago
- User is the owner

#### Test Data
```json
{
  "reason": "Cliente não autorizou a transferência da chave"
}
```

#### BDD Format
```gherkin
Given a claim exists with status OPEN
And the owner is authenticated
When the owner sends POST /api/v1/claims/{claim_id}/cancel with reason
Then the API returns 200 OK
And the claim status changes to CANCELLED
And the Temporal workflow receives "cancel" signal
And the workflow cancels the 30-day timer
And the Bacen is notified via Bridge (CancelClaim)
And the entry remains with the owner (NOT transferred)
And both owner and claimer receive notifications
And the Temporal workflow completes
```

#### Steps
1. Setup: Create claim (use TC-002-001)

2. Authenticate as owner

3. Cancel claim:
   ```bash
   POST /api/v1/claims/7c9e6679-7425-40de-944b-e07fc1f90ae7/cancel
   Authorization: Bearer {owner_access_token}
   Content-Type: application/json

   {
     "reason": "Cliente não autorizou a transferência da chave"
   }
   ```

4. Verify response:
   Expected: 200 OK
   ```json
   {
     "claim_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
     "status": "CANCELLED",
     "reason": "Cliente não autorizou a transferência da chave",
     "cancelled_at": "2025-10-25T10:00:00Z",
     "cancelled_by": "{owner_user_id}"
   }
   ```

5. Verify database state:
   ```sql
   SELECT * FROM dict.claims
   WHERE claim_id = '7c9e6679-7425-40de-944b-e07fc1f90ae7';
   ```
   Expected:
   - status = 'CANCELLED'
   - reason = 'Cliente não autorizou...'
   - cancelled_at IS NOT NULL

6. Verify entry NOT transferred:
   ```sql
   SELECT * FROM dict.entries
   WHERE entry_id = '550e8400-e29b-41d4-a716-446655440000';
   ```
   Expected:
   - owner_ispb = '12345678' (unchanged - still owner)
   - account_id = {original_owner_account_id}

7. Verify Temporal workflow completed

8. Verify notifications:
   - Owner: "Reivindicação cancelada com sucesso."
   - Claimer: "Reivindicação foi cancelada pelo owner."

#### Expected Result
- ✅ Claim status = CANCELLED
- ✅ Entry remains with owner (NOT transferred)
- ✅ Bacen notified (CancelClaimRequest sent)
- ✅ Both users notified
- ✅ Temporal workflow completed
- ✅ Reason persisted in database

#### Status
⬜ Not Run

---

### TC-002-010: Cancel Claim - Missing Reason

**Priority**: P0 (Critical)
**Type**: Negative Test

#### BDD Format
```gherkin
Given a claim exists with status OPEN
When the owner sends cancel request WITHOUT reason
Then the API returns 400 Bad Request
And the error code is "MISSING_REASON"
And the error message is "Reason is required when cancelling a claim"
```

#### Expected Result
- ✅ 400 Bad Request
- ✅ Error: Reason required
- ✅ Claim status unchanged

#### Status
⬜ Not Run

---

## P0 Test Cases (Expire Claim - Auto-confirm)

### TC-002-011: Expire Claim - 30 Days Pass Without Response

**Priority**: P0 (Critical)
**Type**: E2E Functional (Time-based)
**Estimated Duration**: 30 seconds (with time acceleration)

#### Preconditions
- Claim exists with status OPEN
- 30 days have passed since creation (accelerated to 30 seconds in test)
- Owner did NOT confirm or cancel

#### BDD Format
```gherkin
Given a claim exists with status OPEN
And 30 days have passed since creation
And the owner did NOT respond (no confirm or cancel)
When the Temporal workflow timer fires (30 days expired)
Then the workflow auto-confirms the claim (Bacen regulation)
And the claim status changes to EXPIRED
And the Bacen is notified via Bridge (CompleteClaim with auto_confirmed flag)
And the entry is transferred to claimer account
And both owner and claimer receive notifications
And the Temporal workflow completes
```

#### Steps
1. Setup: Create claim (use TC-002-001)
   - Note: Enable time acceleration in Temporal test environment
   - 30 days → 30 seconds

2. Wait for 30 seconds (simulating 30 days)

3. Verify database state after 30 seconds:
   ```sql
   SELECT * FROM dict.claims
   WHERE claim_id = '7c9e6679-7425-40de-944b-e07fc1f90ae7';
   ```
   Expected:
   - status = 'EXPIRED'
   - auto_confirmed = TRUE
   - completed_at IS NOT NULL

4. Verify entry transferred:
   ```sql
   SELECT * FROM dict.entries
   WHERE entry_id = '550e8400-e29b-41d4-a716-446655440000';
   ```
   Expected:
   - owner_ispb = '87654321' (transferred to claimer)
   - account_id = {claimer_account_id}

5. Verify Temporal workflow:
   - Workflow status = COMPLETED
   - Workflow history shows TimerFired event
   - Workflow history shows auto-confirm logic executed

6. Verify Bacen notified:
   - Bridge logs show CompleteClaim request sent
   - Request contains auto_confirmed: true

7. Verify notifications:
   - Owner: "Reivindicação expirou (30 dias). Chave transferida automaticamente."
   - Claimer: "Reivindicação concluída por expiração. Chave agora é sua!"

#### Expected Result
- ✅ Claim status = EXPIRED (auto-confirmed)
- ✅ Entry transferred to claimer after 30 days
- ✅ Bacen notified with auto_confirmed flag
- ✅ Both users notified
- ✅ Temporal workflow completed
- ✅ Total time = ~30 seconds (with acceleration)

#### Status
⬜ Not Run

---

### TC-002-012: Expire Claim - Temporal Workflow Survives Restart

**Priority**: P0 (Critical)
**Type**: Durability Test

#### Preconditions
- Claim exists with status OPEN
- Claim was created 15 days ago
- Temporal worker is running

#### BDD Format
```gherkin
Given a claim exists with 15 days remaining
And the Temporal workflow is RUNNING
When the Temporal worker is restarted
Then the workflow continues from where it left off
And the 30-day timer is preserved
And after 30 total days the claim auto-confirms
```

#### Steps
1. Setup: Create claim
2. Wait 15 days (accelerated time)
3. Verify workflow is RUNNING
4. Restart Temporal worker (kill pod or service)
5. Wait for worker to come back
6. Verify workflow still exists and is RUNNING
7. Wait remaining 15 days (accelerated)
8. Verify claim auto-confirms

#### Expected Result
- ✅ Workflow survives restart
- ✅ Timer preserved across restart
- ✅ Claim auto-confirms after full 30 days
- ✅ No data loss

#### Status
⬜ Not Run

---

## P1 Test Cases (Edge Cases)

### TC-002-013: Concurrent Confirm and Cancel

**Priority**: P1 (High)
**Type**: Race Condition Test

#### BDD Format
```gherkin
Given a claim exists with status OPEN
When the owner clicks CONFIRM
And simultaneously clicks CANCEL
Then only one action succeeds (either confirm OR cancel)
And the other action returns 409 Conflict
And the claim has consistent final state
```

#### Expected Result
- ✅ One action succeeds (200 OK)
- ✅ Other action fails (409 Conflict)
- ✅ Database has consistent state
- ✅ No duplicate Bacen calls

#### Status
⬜ Not Run

---

### TC-002-014: Confirm Claim on Day 29 (Just Before Expiration)

**Priority**: P1 (High)
**Type**: Edge Case

#### BDD Format
```gherkin
Given a claim was created 29 days ago
And the claim will expire in 1 day
When the owner confirms the claim
Then the confirmation succeeds
And the 30-day timer is cancelled
And the claim status is CONFIRMED (not EXPIRED)
```

#### Expected Result
- ✅ Confirmation succeeds on day 29
- ✅ Timer cancelled
- ✅ Status = CONFIRMED (not EXPIRED)

#### Status
⬜ Not Run

---

### TC-002-015: Multiple Claims by Same Claimer

**Priority**: P1 (High)
**Type**: Business Logic Test

#### BDD Format
```gherkin
Given claimer has 3 OPEN claims for different entries
When the claimer creates a 4th claim
Then the 4th claim is created successfully
And all 4 claims coexist independently
And each has its own 30-day timer
```

#### Expected Result
- ✅ Multiple claims allowed per claimer
- ✅ Each claim independent
- ✅ Each has own timer

#### Status
⬜ Not Run

---

### TC-002-016: Claim After Previous Claim Expired

**Priority**: P1 (High)
**Type**: Sequential Test

#### BDD Format
```gherkin
Given entry had a claim that EXPIRED 5 days ago
And the entry is now owned by the previous claimer
When a new claimer creates a new claim for same entry
Then the new claim is created successfully
And the old claim does not interfere
```

#### Expected Result
- ✅ New claim allowed after previous expired
- ✅ Old claim does not block
- ✅ New claim has fresh 30-day timer

#### Status
⬜ Not Run

---

### TC-002-017: Bacen Returns Error on CreateClaim

**Priority**: P1 (High)
**Type**: Integration Error Test

#### BDD Format
```gherkin
Given Bacen is mocked to return SOAP Fault on CreateClaim
When the claimer creates a claim
Then the API returns 201 Created (synchronous)
But the Temporal workflow retries 3 times
And after 3 failures the workflow fails
And the claim status remains PENDING
And an alert is triggered
```

#### Expected Result
- ✅ Synchronous response = 201
- ✅ Workflow retries 3 times
- ✅ Workflow fails after 3 attempts
- ✅ Claim status = PENDING
- ✅ Alert sent to DevOps

#### Status
⬜ Not Run

---

### TC-002-018: Owner Confirms After Claim Expired

**Priority**: P1 (High)
**Type**: Edge Case

#### BDD Format
```gherkin
Given a claim expired 1 hour ago (auto-confirmed)
And the entry is already transferred to claimer
When the owner attempts to confirm the claim
Then the API returns 409 Conflict
And the error code is "CLAIM_ALREADY_RESOLVED"
And the error message is "Claim has already expired and auto-confirmed"
```

#### Expected Result
- ✅ 409 Conflict returned
- ✅ Error: Claim already resolved
- ✅ No state change

#### Status
⬜ Not Run

---

## Test Execution Summary

### Coverage Matrix

| Category | P0 Tests | P1 Tests | Total |
|----------|----------|----------|-------|
| Create Claim | 5 | 0 | 5 |
| Confirm Claim | 3 | 2 | 5 |
| Cancel Claim | 2 | 0 | 2 |
| Expire Claim | 2 | 2 | 4 |
| Edge Cases | 0 | 4 | 4 |
| **Total** | **12** | **8** | **20** |

### Priority Distribution
- P0 (Critical): 12 tests (60%)
- P1 (High): 8 tests (40%)

### Test Types
- E2E Functional: 8 tests
- Negative Tests: 5 tests
- Security Tests: 2 tests
- Edge Cases: 4 tests
- Durability Tests: 1 test

---

## Automation Scripts

### Jest E2E Test Example
```javascript
// tests/e2e/claim-workflow.test.js
describe('ClaimWorkflow E2E', () => {
  let claimerToken, ownerToken, entryId, claimId;

  beforeAll(async () => {
    // Authenticate both users
    const claimerAuth = await request(AUTH_URL)
      .post('/auth/login')
      .send({
        email: 'claimer@bancob.com.br',
        password: 'Test@1234'
      });
    claimerToken = claimerAuth.body.access_token;

    const ownerAuth = await request(AUTH_URL)
      .post('/auth/login')
      .send({
        email: 'owner@bancoa.com.br',
        password: 'Test@1234'
      });
    ownerToken = ownerAuth.body.access_token;

    // Create test entry (owned by owner)
    const entryResponse = await request(DICT_API_URL)
      .post('/api/v1/keys')
      .set('Authorization', `Bearer ${ownerToken}`)
      .send({
        key_type: 'CPF',
        key_value: '11122233344',
        account: { ... }
      });
    entryId = entryResponse.body.entry_id;
  });

  describe('TC-002-001: Create Claim', () => {
    it('should create claim successfully', async () => {
      const response = await request(DICT_API_URL)
        .post('/api/v1/claims')
        .set('Authorization', `Bearer ${claimerToken}`)
        .send({
          entry_id: entryId,
          claimer_account: {
            ispb: '87654321',
            account_number: '654321',
            branch: '0002',
            account_type: 'CACC'
          },
          completion_period_days: 30
        });

      expect(response.status).toBe(201);
      expect(response.body.claim_id).toBeDefined();
      expect(response.body.status).toBe('OPEN');
      expect(response.body.expires_at).toBeDefined();

      claimId = response.body.claim_id;

      // Verify workflow started
      await waitFor(async () => {
        const workflow = await temporalClient.getWorkflowHandle(
          `claim-${claimId}`
        );
        expect(workflow).toBeDefined();
      }, { timeout: 1000 });
    });
  });

  describe('TC-002-006: Confirm Claim', () => {
    it('should confirm claim and transfer entry', async () => {
      const response = await request(DICT_API_URL)
        .post(`/api/v1/claims/${claimId}/confirm`)
        .set('Authorization', `Bearer ${ownerToken}`);

      expect(response.status).toBe(200);
      expect(response.body.status).toBe('CONFIRMED');

      // Wait for async processing
      await waitFor(async () => {
        const claimStatus = await request(DICT_API_URL)
          .get(`/api/v1/claims/${claimId}`)
          .set('Authorization', `Bearer ${ownerToken}`);

        expect(claimStatus.body.status).toBe('COMPLETED');

        // Verify entry transferred
        const entry = await request(DICT_API_URL)
          .get(`/api/v1/keys/CPF/11122233344`)
          .set('Authorization', `Bearer ${claimerToken}`);

        expect(entry.body.owner_ispb).toBe('87654321');
      }, { timeout: 5000 });
    });
  });

  describe('TC-002-011: Expire Claim (Time Acceleration)', () => {
    it('should auto-confirm claim after 30 days', async () => {
      // Create new claim
      const createResponse = await request(DICT_API_URL)
        .post('/api/v1/claims')
        .set('Authorization', `Bearer ${claimerToken}`)
        .send({ ... });

      const newClaimId = createResponse.body.claim_id;

      // Enable time acceleration in Temporal test env
      // (30 days → 30 seconds)

      // Wait 30 seconds
      await sleep(30000);

      // Verify claim auto-confirmed
      const claimStatus = await request(DICT_API_URL)
        .get(`/api/v1/claims/${newClaimId}`)
        .set('Authorization', `Bearer ${claimerToken}`);

      expect(claimStatus.body.status).toBe('EXPIRED');
      expect(claimStatus.body.auto_confirmed).toBe(true);

      // Verify entry transferred
      const entry = await request(DICT_API_URL)
        .get(`/api/v1/keys/CPF/...`)
        .set('Authorization', `Bearer ${claimerToken}`);

      expect(entry.body.owner_ispb).toBe('87654321');
    });
  });
});
```

### Temporal Workflow Test
```go
// workflow_test.go
func TestClaimWorkflow(t *testing.T) {
    testSuite := &testsuite.WorkflowTestSuite{}
    env := testSuite.NewTestWorkflowEnvironment()

    // Mock activities
    env.OnActivity("CreateClaimInBacen", mock.Anything, mock.Anything).
        Return(&BridgeResponse{BacenClaimID: "bacen_123"}, nil)

    env.OnActivity("CompleteClaimInBacen", mock.Anything, mock.Anything).
        Return(nil, nil)

    // Execute workflow
    env.ExecuteWorkflow(ClaimWorkflow, ClaimWorkflowParams{
        ClaimID: "test-claim-123",
        EntryID: "test-entry-456",
    })

    // Verify workflow is waiting for timer
    require.True(t, env.IsWorkflowCompleted())

    // Fast-forward time by 30 days
    env.RegisterDelayedCallback(func() {
        // Timer should fire
    }, 30*24*time.Hour)

    // Verify workflow completed
    require.NoError(t, env.GetWorkflowError())

    // Verify activities called
    env.AssertExpectations(t)
}
```

---

## Entry/Exit Criteria

### Entry Criteria
- ✅ Core DICT API deployed to test environment
- ✅ Temporal server configured with test namespace
- ✅ Bridge mock available (Bacen simulator)
- ✅ Test users created (claimer and owner)
- ✅ Test entries created
- ✅ Time acceleration enabled in Temporal

### Exit Criteria
- ✅ All P0 tests passed (12/12)
- ✅ 90%+ P1 tests passed (7/8 minimum)
- ✅ No P0 defects open
- ✅ ClaimWorkflow durability verified
- ✅ 30-day timer accuracy verified
- ✅ Bacen integration verified

---

## Monitoring During Tests

### Temporal Metrics
```prometheus
# Workflows started
temporal_workflow_started_total{workflow_type="ClaimWorkflow"}

# Workflows completed
temporal_workflow_completed_total{workflow_type="ClaimWorkflow", status="success"}

# Timer accuracy
temporal_timer_drift_seconds{workflow_type="ClaimWorkflow"}

# Signal received
temporal_signal_received_total{workflow_type="ClaimWorkflow", signal="confirm"}
```

### Business Metrics
```prometheus
# Claims created
dict_claims_created_total{status="OPEN"}

# Claims resolved
dict_claims_resolved_total{resolution="CONFIRMED"}
dict_claims_resolved_total{resolution="CANCELLED"}
dict_claims_resolved_total{resolution="EXPIRED"}

# Average resolution time
dict_claim_resolution_duration_seconds
```

---

## Glossary

- **Claimer**: Institution requesting ownership of a PIX key
- **Owner**: Current owner of the PIX key
- **Auto-confirm**: Automatic confirmation after 30 days without response
- **ISPB**: Identifier of financial institution (8 digits)
- **Temporal**: Workflow orchestration platform
- **Time Acceleration**: Test feature to speed up 30-day timer

---

**Última Revisão**: 2025-10-25
**Aprovado por**: QA Lead
**Próxima Revisão**: Sprint review
