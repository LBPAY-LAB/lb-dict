# US-001: User Stories - DICT Key Management

**Documento**: US-001_User_Stories_DICT_Keys.md
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: Product Owner - DICT Team
**Epic**: EP-007 - Requirements & Business
**Priority**: Must Have
**Status**: Ready for Development

---

## Sumário Executivo

Este documento contém as user stories para gerenciamento de chaves DICT (PIX), incluindo operações de criação, consulta e exclusão de chaves. As histórias estão escritas no formato padrão "As a... I want... So that..." com critérios de aceitação detalhados em formato Given/When/Then.

---

## User Stories

### US-001.1: Create DICT Key

**Story ID**: US-001.1
**Priority**: Must Have
**Story Points**: 8
**Sprint**: Sprint 8

#### User Story

As a **PIX user**,
I want **to register a new DICT key (CPF, email, phone, or random key)**,
So that **I can receive PIX transfers using this key instead of sharing my bank account details**.

#### Acceptance Criteria

**AC-001.1.1: Valid CPF Key Creation**
- **Given** I am an authenticated user with a valid account
- **When** I request to create a DICT key with type "CPF" and value matching my CPF
- **Then** the system validates the CPF format (11 digits, valid check digits)
- **And** creates the key linked to my default account
- **And** sends a synchronization request to BACEN RSFN
- **And** returns a confirmation with key status "PENDING"
- **And** sends me a notification when BACEN confirms the registration

**AC-001.1.2: Valid Email Key Creation**
- **Given** I am an authenticated user with a valid account
- **When** I request to create a DICT key with type "EMAIL" and a valid email address
- **Then** the system validates the email format (RFC 5322)
- **And** sends a verification code to the email address
- **And** waits for email confirmation within 24 hours
- **And** only after confirmation, synchronizes with BACEN RSFN
- **And** returns key status "PENDING_CONFIRMATION" initially, then "ACTIVE" after confirmation

**AC-001.1.3: Valid Phone Key Creation**
- **Given** I am an authenticated user with a valid account
- **When** I request to create a DICT key with type "PHONE" and a valid phone number in format "+5511999999999"
- **Then** the system validates the phone format (E.164)
- **And** sends a verification SMS with a 6-digit code
- **And** waits for SMS confirmation within 10 minutes
- **And** after confirmation, synchronizes with BACEN RSFN
- **And** returns key status "PENDING_CONFIRMATION" initially, then "ACTIVE" after confirmation

**AC-001.1.4: Random Key Creation (EVP)**
- **Given** I am an authenticated user with a valid account
- **When** I request to create a DICT key with type "EVP" (random key)
- **Then** the system generates a UUID v4 in format "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx"
- **And** validates the key is unique in local database
- **And** creates the key linked to my specified account
- **And** synchronizes with BACEN RSFN
- **And** returns the generated random key to me
- **And** sets key status to "PENDING" until BACEN confirms

**AC-001.1.5: Duplicate Key Rejection**
- **Given** I try to create a DICT key
- **When** the key already exists in my account or another account
- **Then** the system returns HTTP 409 Conflict
- **And** provides an error message "This key is already registered. If you believe this is an error, you can initiate a portability claim."
- **And** does not synchronize with BACEN
- **And** logs the attempt for audit purposes

**AC-001.1.6: Account Ownership Validation**
- **Given** I try to create a DICT key
- **When** I specify an account number
- **Then** the system validates that I am the legal owner of the account
- **And** validates that the account is active (not blocked or closed)
- **And** validates that the account type allows DICT registration (current or savings account)
- **And** rejects the request if any validation fails with appropriate error message

**AC-001.1.7: Maximum Keys Per Account**
- **Given** I try to create a DICT key
- **When** I have already registered the maximum number of keys allowed (20 keys for individuals, 50 for legal entities)
- **Then** the system returns HTTP 422 Unprocessable Entity
- **And** provides error message "Maximum number of keys reached. Please delete an existing key before creating a new one."
- **And** does not synchronize with BACEN

#### Business Rules

**BR-001.1.1: Key Types Allowed**
- CPF: Only for individual accounts, must match account holder CPF
- CNPJ: Only for legal entity accounts, must match account holder CNPJ
- EMAIL: Any valid email, requires email verification
- PHONE: Brazilian phone numbers only (+55), requires SMS verification
- EVP: System-generated UUID v4, no verification required

**BR-001.1.2: Key Limits**
- Individual (CPF): Maximum 20 keys
- Legal Entity (CNPJ): Maximum 50 keys
- One CPF/CNPJ key per account holder (unique)
- Multiple EMAIL/PHONE/EVP keys allowed

**BR-001.1.3: Verification Timeouts**
- Email verification: 24 hours
- SMS verification: 10 minutes
- After timeout, key creation request is cancelled

**BR-001.1.4: BACEN Synchronization**
- All key creations must be synchronized with BACEN RSFN
- Synchronization timeout: 30 seconds
- Maximum retry attempts: 3 (exponential backoff)
- If BACEN rejects, key is deleted from local database

**BR-001.1.5: Key Ownership**
- Keys can only be created for accounts owned by the authenticated user
- Account must be active and not blocked
- Account type must support DICT (current or savings account)

#### Dependencies

- API-001: Core DICT REST API must be implemented
- TEC-002: Bridge integration with BACEN RSFN must be operational
- TSP-001: Temporal Workflow for CreateEntry must be deployed
- DAT-001: Database schema for entries table must exist

#### Technical Notes

- **API Endpoint**: POST /api/v1/entries
- **Request Payload**:
  ```json
  {
    "type": "CPF" | "CNPJ" | "EMAIL" | "PHONE" | "EVP",
    "value": "string (optional for EVP)",
    "accountNumber": "string",
    "accountType": "CACC" | "SVGS"
  }
  ```
- **Response Payload**:
  ```json
  {
    "keyId": "uuid",
    "type": "string",
    "value": "string",
    "status": "PENDING" | "PENDING_CONFIRMATION" | "ACTIVE",
    "createdAt": "ISO8601 timestamp"
  }
  ```

---

### US-001.2: View DICT Keys

**Story ID**: US-001.2
**Priority**: Must Have
**Story Points**: 5
**Sprint**: Sprint 8

#### User Story

As a **PIX user**,
I want **to view all my registered DICT keys and their current status**,
So that **I can manage my keys and know which ones are active or pending**.

#### Acceptance Criteria

**AC-001.2.1: List All Keys**
- **Given** I am an authenticated user
- **When** I request to list all my DICT keys
- **Then** the system returns a paginated list of all keys linked to my accounts
- **And** each key includes: type, value, status, account number, creation date, last update date
- **And** keys are sorted by creation date (newest first)
- **And** pagination supports page size of 10, 20, 50 items

**AC-001.2.2: Filter Keys by Status**
- **Given** I am viewing my DICT keys
- **When** I apply a filter by status (ACTIVE, PENDING, BLOCKED, PENDING_CONFIRMATION)
- **Then** the system returns only keys matching the selected status
- **And** maintains pagination
- **And** displays total count of filtered keys

**AC-001.2.3: Filter Keys by Type**
- **Given** I am viewing my DICT keys
- **When** I apply a filter by type (CPF, CNPJ, EMAIL, PHONE, EVP)
- **Then** the system returns only keys matching the selected type
- **And** maintains pagination

**AC-001.2.4: View Key Details**
- **Given** I am viewing my DICT keys list
- **When** I select a specific key to view details
- **Then** the system displays complete key information:
  - Key ID (UUID)
  - Type and value
  - Status and status history
  - Linked account (number, branch, bank)
  - Creation timestamp
  - Last update timestamp
  - BACEN synchronization status
  - Active claims (if any)
  - Audit log (last 10 events)

**AC-001.2.5: Empty State**
- **Given** I am a new user with no DICT keys
- **When** I request to view my keys
- **Then** the system displays an empty state message
- **And** provides a call-to-action button "Create your first PIX key"
- **And** explains the benefits of having DICT keys

**AC-001.2.6: Search Keys**
- **Given** I have multiple DICT keys registered
- **When** I search by key value (partial or complete)
- **Then** the system returns keys matching the search term
- **And** highlights the matching portion of the key value
- **And** supports case-insensitive search

#### Business Rules

**BR-001.2.1: Visibility**
- Users can only view keys linked to their own accounts
- Admin users can view keys from any account (with proper authorization)

**BR-001.2.2: Status Values**
- ACTIVE: Key is registered and operational
- PENDING: Waiting BACEN confirmation
- PENDING_CONFIRMATION: Waiting user verification (email/SMS)
- BLOCKED: Temporarily blocked (by user or admin)
- DELETED: Soft-deleted (hidden from list by default)
- CLAIM_PENDING: A portability/ownership claim is in progress

**BR-001.2.3: Pagination**
- Default page size: 20 items
- Maximum page size: 100 items
- Include total count in response headers

#### Dependencies

- API-001: Core DICT REST API
- DAT-001: Database schema with entries table

#### Technical Notes

- **API Endpoint**: GET /api/v1/entries
- **Query Parameters**:
  - `page`: integer (default: 1)
  - `pageSize`: integer (default: 20, max: 100)
  - `status`: enum (optional filter)
  - `type`: enum (optional filter)
  - `search`: string (optional search term)
- **Response Payload**:
  ```json
  {
    "entries": [
      {
        "keyId": "uuid",
        "type": "string",
        "value": "string",
        "status": "string",
        "accountNumber": "string",
        "createdAt": "ISO8601",
        "updatedAt": "ISO8601"
      }
    ],
    "pagination": {
      "page": 1,
      "pageSize": 20,
      "totalItems": 45,
      "totalPages": 3
    }
  }
  ```

---

### US-001.3: Delete DICT Key

**Story ID**: US-001.3
**Priority**: Must Have
**Story Points**: 5
**Sprint**: Sprint 8

#### User Story

As a **PIX user**,
I want **to delete a DICT key that I no longer want to use**,
So that **I stop receiving PIX transfers through that key and improve my privacy**.

#### Acceptance Criteria

**AC-001.3.1: Successful Key Deletion**
- **Given** I am the owner of an active DICT key
- **When** I request to delete the key
- **Then** the system asks for confirmation with a warning message
- **And** after confirmation, initiates a delete workflow
- **And** synchronizes the deletion with BACEN RSFN
- **And** marks the key status as "PENDING_DELETION" until BACEN confirms
- **And** after BACEN confirmation, changes status to "DELETED"
- **And** sends me a notification confirming the deletion

**AC-001.3.2: Delete Key with Active Claims**
- **Given** I try to delete a DICT key
- **When** the key has an active portability or ownership claim
- **Then** the system returns HTTP 409 Conflict
- **And** displays error message "Cannot delete key with active claims. Please wait for claim resolution or cancel the claim first."
- **And** does not proceed with deletion

**AC-001.3.3: Delete Non-Existent Key**
- **Given** I try to delete a DICT key
- **When** the key does not exist or was already deleted
- **Then** the system returns HTTP 404 Not Found
- **And** provides error message "Key not found"

**AC-001.3.4: Delete Key from Another User**
- **Given** I try to delete a DICT key
- **When** the key belongs to another user's account
- **Then** the system returns HTTP 403 Forbidden
- **And** provides error message "You are not authorized to delete this key"
- **And** logs the attempt for security audit

**AC-001.3.5: BACEN Synchronization Failure**
- **Given** I request to delete a DICT key
- **When** BACEN RSFN is unavailable or rejects the deletion
- **Then** the system retries up to 3 times with exponential backoff
- **And** if all retries fail, marks the key as "DELETION_FAILED"
- **And** schedules a background job to retry later
- **And** notifies me about the temporary failure

**AC-001.3.6: Confirmation Step**
- **Given** I initiate a key deletion
- **When** the system requests confirmation
- **Then** it displays:
  - Key type and value being deleted
  - Warning that the action is irreversible
  - Information that key can be registered again after 7 days (cooling-off period)
  - "Cancel" and "Confirm Deletion" buttons
- **And** only proceeds with deletion after explicit confirmation

#### Business Rules

**BR-001.3.1: Deletion Types**
- **Soft Delete**: Key is marked as DELETED but remains in database for audit
- **Hard Delete**: Not allowed; all deletions are soft deletes
- **Cooling-off Period**: 7 days before the same key can be registered again

**BR-001.3.2: BACEN Synchronization**
- All deletions must be synchronized with BACEN RSFN
- Deletion is only finalized after BACEN confirmation
- If BACEN rejects, key remains ACTIVE with error notification

**BR-001.3.3: Restrictions**
- Cannot delete key with active claims (ownership or portability)
- Cannot delete key that is pending user verification
- Can delete keys in PENDING, ACTIVE, or BLOCKED status

**BR-001.3.4: Audit Trail**
- All deletion attempts (successful or failed) must be logged
- Log includes: user ID, timestamp, key ID, reason (if provided), IP address

#### Dependencies

- API-001: Core DICT REST API
- TEC-002: Bridge integration with BACEN RSFN
- TSP-001: Temporal Workflow for DeleteEntry
- DAT-001: Database schema

#### Technical Notes

- **API Endpoint**: DELETE /api/v1/entries/{keyId}
- **Request Headers**: Authorization token required
- **Response on Success**:
  ```json
  {
    "message": "Key deletion initiated successfully",
    "keyId": "uuid",
    "status": "PENDING_DELETION",
    "estimatedCompletionTime": "ISO8601 timestamp"
  }
  ```
- **Temporal Workflow**: DeleteEntryWorkflow
  - Activity 1: Validate ownership
  - Activity 2: Check active claims
  - Activity 3: Call Bridge to delete from BACEN
  - Activity 4: Update local database
  - Activity 5: Send notification

---

## Summary Table

| Story ID | Title | Priority | Story Points | Dependencies |
|----------|-------|----------|--------------|--------------|
| US-001.1 | Create DICT Key | Must Have | 8 | API-001, TEC-002, TSP-001, DAT-001 |
| US-001.2 | View DICT Keys | Must Have | 5 | API-001, DAT-001 |
| US-001.3 | Delete DICT Key | Must Have | 5 | API-001, TEC-002, TSP-001, DAT-001 |

**Total Story Points**: 18

---

## Acceptance Checklist

- [ ] All acceptance criteria reviewed by Product Owner
- [ ] Business rules validated with compliance team
- [ ] BACEN regulations compliance verified
- [ ] Dependencies identified and documented
- [ ] API contracts defined and reviewed
- [ ] Security requirements included
- [ ] Audit logging requirements specified
- [ ] Error handling scenarios covered
- [ ] User notification requirements defined
- [ ] Performance requirements specified (if applicable)

---

## References

- **BACEN Manual DICT**: [Link to official documentation]
- **PIX Regulations**: Circular 3.985/2020
- **API-001**: Core DICT REST API Specification
- **TEC-002**: Bridge Technical Specification v3.1
- **TSP-001**: Temporal Workflow Engine TechSpec
- **DAT-001**: Database Schema Design

---

**Last Updated**: 2025-10-25
**Next Review**: Sprint 8 Planning Session
