# US-002: User Stories - Claim Management

**Documento**: US-002_User_Stories_Claims.md
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: Product Owner - DICT Team
**Epic**: EP-007 - Requirements & Business
**Priority**: Must Have
**Status**: Ready for Development

---

## Sumário Executivo

Este documento contém as user stories para gerenciamento de reivindicações (claims) de chaves DICT, incluindo reivindicações de portabilidade e posse. O processo de claim segue regras específicas do BACEN, incluindo o período obrigatório de 30 dias para resposta do proprietário atual da chave.

---

## User Stories

### US-002.1: Create Portability Claim

**Story ID**: US-002.1
**Priority**: Must Have
**Story Points**: 8
**Sprint**: Sprint 8

#### User Story

As a **PIX user**,
I want **to create a portability claim for a DICT key currently registered in another institution**,
So that **I can transfer the key ownership to my account in LBPay**.

#### Acceptance Criteria

**AC-002.1.1: Successful Portability Claim Creation**
- **Given** I am an authenticated user with a valid account
- **And** I provide a DICT key (CPF, email, phone, or EVP) that exists in another institution
- **When** I request to create a portability claim
- **Then** the system validates the key exists in BACEN RSFN
- **And** validates the key is owned by me (CPF/CNPJ match for CPF/CNPJ keys, or I provide proof of ownership for email/phone)
- **And** creates a claim with status "WAITING_RESOLUTION"
- **And** initiates a Temporal workflow with 30-day timer
- **And** synchronizes the claim with BACEN RSFN
- **And** notifies the current key owner (at the other institution) about the claim
- **And** returns claim ID and expected resolution date (30 days from now)

**AC-002.1.2: Email/Phone Ownership Verification**
- **Given** I create a portability claim for an EMAIL or PHONE key
- **When** the key type requires ownership verification
- **Then** the system sends a verification code to the email/phone
- **And** waits for my confirmation within 10 minutes (phone) or 24 hours (email)
- **And** only after verification, proceeds with claim creation
- **And** if verification times out, cancels the claim request

**AC-002.1.3: Duplicate Claim Prevention**
- **Given** I try to create a portability claim
- **When** there is already an active claim (by me or another user) for the same key
- **Then** the system returns HTTP 409 Conflict
- **And** provides error message "An active claim already exists for this key. Claim ID: {claimId}, Status: {status}"
- **And** does not create a new claim

**AC-002.1.4: Invalid Key Validation**
- **Given** I try to create a portability claim
- **When** the key does not exist in BACEN RSFN
- **Then** the system returns HTTP 404 Not Found
- **And** provides error message "Key not found in DICT. Please verify the key value."

**AC-002.1.5: Same Institution Prevention**
- **Given** I try to create a portability claim
- **When** the key is already registered in LBPay (same institution)
- **Then** the system returns HTTP 400 Bad Request
- **And** provides error message "This key is already registered in LBPay. Use ownership claim instead if needed."

**AC-002.1.6: Account Ownership Validation**
- **Given** I create a portability claim for a CPF or CNPJ key
- **When** the key value does not match my CPF/CNPJ
- **Then** the system returns HTTP 403 Forbidden
- **And** provides error message "You are not authorized to claim this key. CPF/CNPJ mismatch."

#### Business Rules

**BR-002.1.1: 30-Day Resolution Period**
- Current owner has exactly 30 calendar days to respond to the claim
- After 30 days without response, claim is automatically confirmed (key is transferred)
- Owner can respond earlier with: CONFIRM (accept) or CANCEL (reject)

**BR-002.1.2: Claim Types**
- **Portability Claim**: Transfer key from another institution to LBPay
- **Ownership Claim**: Claim a key currently held by another user in the same institution

**BR-002.1.3: Notification Requirements**
- Current owner must be notified within 1 hour of claim creation
- Claimant receives updates on claim status changes
- Both parties notified when claim is resolved

**BR-002.1.4: BACEN Synchronization**
- All claims must be synchronized with BACEN RSFN
- BACEN tracks claim status centrally
- Resolution must be reported to BACEN

**BR-002.1.5: Maximum Active Claims**
- User can have maximum 5 active claims simultaneously
- No limit on total claims (active + resolved)

#### Dependencies

- API-001: Core DICT REST API
- TEC-002: Bridge integration with BACEN RSFN
- TSP-001: Temporal Workflow for ClaimWorkflow (30 days)
- DAT-002: Database schema for claims table

#### Technical Notes

- **API Endpoint**: POST /api/v1/claims
- **Request Payload**:
  ```json
  {
    "claimType": "PORTABILITY" | "OWNERSHIP",
    "keyType": "CPF" | "CNPJ" | "EMAIL" | "PHONE" | "EVP",
    "keyValue": "string",
    "targetAccountNumber": "string",
    "verificationCode": "string (optional, for email/phone)"
  }
  ```
- **Response Payload**:
  ```json
  {
    "claimId": "uuid",
    "claimType": "PORTABILITY",
    "status": "WAITING_RESOLUTION",
    "keyType": "string",
    "keyValue": "string",
    "createdAt": "ISO8601",
    "resolutionDeadline": "ISO8601 (createdAt + 30 days)"
  }
  ```
- **Temporal Workflow**: ClaimWorkflow
  - Timer: 30 days
  - Activities: ValidateKey, CreateClaim, NotifyOwner, WaitForResponse, ResolveClaim

---

### US-002.2: Respond to Claim (as Current Owner)

**Story ID**: US-002.2
**Priority**: Must Have
**Story Points**: 5
**Sprint**: Sprint 8

#### User Story

As a **PIX user and current owner of a DICT key**,
I want **to respond to a portability or ownership claim initiated by another user**,
So that **I can confirm or reject the transfer of my key**.

#### Acceptance Criteria

**AC-002.2.1: Confirm Claim (Accept Transfer)**
- **Given** I am the current owner of a DICT key with an active claim
- **When** I choose to confirm the claim (accept the transfer)
- **Then** the system validates I am the legitimate owner
- **And** updates claim status to "CONFIRMED"
- **And** immediately initiates the key transfer process
- **And** synchronizes the confirmation with BACEN RSFN
- **And** marks the key as "IN_TRANSFER" in my account
- **And** sends notification to the claimant that the claim was confirmed
- **And** after BACEN processes the transfer, key is removed from my account
- **And** key appears in claimant's account with status "ACTIVE"

**AC-002.2.2: Cancel Claim (Reject Transfer)**
- **Given** I am the current owner of a DICT key with an active claim
- **When** I choose to cancel the claim (reject the transfer)
- **Then** the system validates I am the legitimate owner
- **And** updates claim status to "CANCELLED"
- **And** synchronizes the cancellation with BACEN RSFN
- **And** key remains in my account with status "ACTIVE"
- **And** sends notification to the claimant that the claim was rejected
- **And** provides reason for cancellation (optional text field)

**AC-002.2.3: Response After Deadline**
- **Given** I try to respond to a claim
- **When** more than 30 days have passed since claim creation
- **Then** the system returns HTTP 410 Gone
- **And** provides error message "Claim resolution deadline has passed. The claim was automatically confirmed."
- **And** does not accept my response

**AC-002.2.4: Response to Non-Existent Claim**
- **Given** I try to respond to a claim
- **When** the claim ID does not exist or was already resolved
- **Then** the system returns HTTP 404 Not Found
- **And** provides error message "Claim not found or already resolved"

**AC-002.2.5: Unauthorized Response**
- **Given** I try to respond to a claim
- **When** I am not the current owner of the key
- **Then** the system returns HTTP 403 Forbidden
- **And** provides error message "You are not authorized to respond to this claim"
- **And** logs the attempt for security audit

**AC-002.2.6: Response Confirmation**
- **Given** I choose to confirm or cancel a claim
- **When** the system requests confirmation
- **Then** it displays:
  - Claim details (key type, key value, claimant info)
  - Selected action (confirm or cancel)
  - Consequences of the action
  - Warning that the action is irreversible
  - "Go Back" and "Confirm Action" buttons
- **And** only proceeds after explicit confirmation

#### Business Rules

**BR-002.2.1: Response Window**
- Owner has exactly 30 calendar days to respond
- Response can be submitted any time within the 30-day period
- First response is final; cannot change decision

**BR-002.2.2: Automatic Confirmation**
- If owner does not respond within 30 days, claim is automatically confirmed
- System must process auto-confirmation on day 30 + 1 hour (grace period)

**BR-002.2.3: Response Types**
- **CONFIRM**: Accept the transfer, key moves to claimant
- **CANCEL**: Reject the transfer, key remains with owner

**BR-002.2.4: Notification Requirements**
- Claimant notified immediately when owner responds
- If auto-confirmed, both parties notified on day 31

#### Dependencies

- API-001: Core DICT REST API
- TEC-002: Bridge integration with BACEN RSFN
- TSP-001: Temporal Workflow for ClaimWorkflow
- DAT-002: Database schema for claims table

#### Technical Notes

- **API Endpoint**: PUT /api/v1/claims/{claimId}/respond
- **Request Payload**:
  ```json
  {
    "response": "CONFIRM" | "CANCEL",
    "reason": "string (optional, max 500 chars)"
  }
  ```
- **Response Payload**:
  ```json
  {
    "claimId": "uuid",
    "status": "CONFIRMED" | "CANCELLED",
    "respondedAt": "ISO8601",
    "respondedBy": "userId",
    "reason": "string"
  }
  ```

---

### US-002.3: View Claim Status

**Story ID**: US-002.3
**Priority**: Must Have
**Story Points**: 3
**Sprint**: Sprint 8

#### User Story

As a **PIX user**,
I want **to view the status of all my claims (as claimant or current owner)**,
So that **I can track the progress and know when claims are resolved**.

#### Acceptance Criteria

**AC-002.3.1: List All Claims as Claimant**
- **Given** I am an authenticated user
- **When** I request to view claims where I am the claimant
- **Then** the system returns a paginated list of my claims
- **And** each claim includes: claim ID, claim type, key type, key value, status, creation date, resolution deadline, current owner institution
- **And** claims are sorted by creation date (newest first)
- **And** pagination supports page size of 10, 20, 50 items

**AC-002.3.2: List All Claims as Current Owner**
- **Given** I am an authenticated user
- **When** I request to view claims where I am the current owner
- **Then** the system returns a paginated list of claims against my keys
- **And** each claim includes: claim ID, claim type, key type, key value, status, creation date, resolution deadline, claimant info (masked for privacy)
- **And** claims requiring my action are highlighted
- **And** displays countdown to resolution deadline

**AC-002.3.3: Filter Claims by Status**
- **Given** I am viewing my claims
- **When** I apply a filter by status (WAITING_RESOLUTION, CONFIRMED, CANCELLED, EXPIRED)
- **Then** the system returns only claims matching the selected status
- **And** maintains pagination

**AC-002.3.4: View Claim Details**
- **Given** I am viewing my claims list
- **When** I select a specific claim to view details
- **Then** the system displays complete claim information:
  - Claim ID
  - Claim type (portability or ownership)
  - Key type and value
  - Status and status history
  - Claimant information (name, CPF/CNPJ)
  - Current owner information (name, institution)
  - Creation timestamp
  - Resolution deadline
  - Time remaining (countdown)
  - Response details (if resolved)
  - BACEN synchronization status
  - Audit log

**AC-002.3.5: Claim Timeline Visualization**
- **Given** I am viewing claim details
- **When** the claim is in WAITING_RESOLUTION status
- **Then** the system displays a timeline visualization:
  - Day 0: Claim created
  - Day 30: Resolution deadline
  - Current position (day X of 30)
  - Progress bar showing percentage of time elapsed
  - Estimated remaining days

**AC-002.3.6: Empty State**
- **Given** I have no claims (as claimant or owner)
- **When** I request to view my claims
- **Then** the system displays an empty state message
- **And** provides information about what claims are and how they work

#### Business Rules

**BR-002.3.1: Visibility**
- Users can view claims where they are claimant or current owner
- Admin users can view all claims (with proper authorization)

**BR-002.3.2: Status Values**
- WAITING_RESOLUTION: Waiting for owner response (within 30 days)
- CONFIRMED: Owner accepted the transfer
- CANCELLED: Owner rejected the transfer
- EXPIRED: Automatically confirmed after 30 days with no response
- COMPLETED: Transfer successfully processed by BACEN

**BR-002.3.3: Privacy**
- Claimant's full details visible only to current owner and admins
- Current owner's full details visible only to claimant and admins
- Public view shows masked information

#### Dependencies

- API-001: Core DICT REST API
- DAT-002: Database schema for claims table

#### Technical Notes

- **API Endpoint**: GET /api/v1/claims
- **Query Parameters**:
  - `role`: "claimant" | "owner" (default: both)
  - `page`: integer (default: 1)
  - `pageSize`: integer (default: 20, max: 100)
  - `status`: enum (optional filter)
- **Response Payload**:
  ```json
  {
    "claims": [
      {
        "claimId": "uuid",
        "claimType": "PORTABILITY" | "OWNERSHIP",
        "keyType": "string",
        "keyValue": "string",
        "status": "string",
        "createdAt": "ISO8601",
        "resolutionDeadline": "ISO8601",
        "daysRemaining": 15,
        "role": "claimant" | "owner"
      }
    ],
    "pagination": {
      "page": 1,
      "pageSize": 20,
      "totalItems": 8,
      "totalPages": 1
    }
  }
  ```

---

## Summary Table

| Story ID | Title | Priority | Story Points | Dependencies |
|----------|-------|----------|--------------|--------------|
| US-002.1 | Create Portability Claim | Must Have | 8 | API-001, TEC-002, TSP-001, DAT-002 |
| US-002.2 | Respond to Claim | Must Have | 5 | API-001, TEC-002, TSP-001, DAT-002 |
| US-002.3 | View Claim Status | Must Have | 3 | API-001, DAT-002 |

**Total Story Points**: 16

---

## 30-Day Claim Business Rule Details

### Timeline Breakdown

```
Day 0: Claim Created
├─ Claimant submits portability/ownership claim
├─ System validates key exists in BACEN
├─ System creates claim with status WAITING_RESOLUTION
├─ Temporal Workflow starts with 30-day timer
└─ Current owner notified (email + push notification)

Day 1-29: Response Window
├─ Current owner can respond at any time
├─ If CONFIRM: Key transfer starts immediately
├─ If CANCEL: Claim is rejected, key remains with owner
└─ Claimant can check status but cannot cancel claim

Day 30: Deadline
├─ 23:59:59 - Last moment for owner to respond
└─ Temporal timer fires at midnight

Day 31 (00:00:00 + 1 hour grace):
├─ If no response received: Auto-confirm claim
├─ System updates claim status to EXPIRED
├─ System initiates automatic key transfer
├─ Both parties notified of auto-confirmation
└─ Key transfer process begins

Day 31+: Transfer Processing
├─ BACEN processes the transfer (1-2 hours)
├─ Key removed from current owner's account
├─ Key added to claimant's account
└─ Claim status updated to COMPLETED
```

### Edge Cases

1. **Owner responds on day 30 at 23:59**: Response accepted
2. **Owner responds on day 31 at 00:01**: Response rejected (too late)
3. **BACEN unavailable on day 31**: Auto-confirmation delayed, retries every hour
4. **Claimant closes account during 30-day period**: Claim automatically cancelled
5. **Current owner closes account during 30-day period**: Claim automatically confirmed
6. **Key deleted by owner during claim period**: Claim automatically confirmed, key restored and transferred

---

## Acceptance Checklist

- [ ] All acceptance criteria reviewed by Product Owner
- [ ] 30-day business rule validated with BACEN regulations
- [ ] Notification requirements specified for all scenarios
- [ ] Privacy requirements for claim visibility defined
- [ ] Temporal Workflow for 30-day timer designed
- [ ] Edge cases for deadline handling documented
- [ ] Auto-confirmation logic tested
- [ ] BACEN synchronization requirements clear
- [ ] Audit logging requirements specified
- [ ] Error handling scenarios covered

---

## References

- **BACEN Circular 3.985/2020**: PIX Claim Regulations
- **BACEN Manual DICT - Section 7**: Portability and Ownership Claims
- **API-001**: Core DICT REST API Specification
- **TEC-002**: Bridge Technical Specification v3.1
- **TSP-001**: Temporal Workflow Engine TechSpec
- **DAT-002**: Database Schema - Claims Table
- **BP-002**: Business Process - ClaimWorkflow (30 days)

---

**Last Updated**: 2025-10-25
**Next Review**: Sprint 8 Planning Session
