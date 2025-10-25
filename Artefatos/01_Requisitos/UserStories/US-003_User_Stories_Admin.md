# US-003: User Stories - Admin Operations

**Documento**: US-003_User_Stories_Admin.md
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: Product Owner - DICT Team
**Epic**: EP-007 - Requirements & Business
**Priority**: Should Have
**Status**: Ready for Development

---

## Sumário Executivo

Este documento contém as user stories para operações administrativas do sistema DICT, incluindo visualização de workflows, cancelamento de workflows travados, e auditoria de operações. Estas funcionalidades são destinadas a usuários com perfil de administrador e operadores do sistema.

---

## User Stories

### US-003.1: View All Workflows

**Story ID**: US-003.1
**Priority**: Should Have
**Story Points**: 5
**Sprint**: Sprint 9

#### User Story

As an **Admin or System Operator**,
I want **to view all running, completed, and failed Temporal workflows in the DICT system**,
So that **I can monitor system health and identify issues proactively**.

#### Acceptance Criteria

**AC-003.1.1: List All Workflows with Filters**
- **Given** I am logged in as an admin user
- **When** I access the workflow monitoring dashboard
- **Then** the system displays a paginated list of all Temporal workflows
- **And** each workflow shows: workflow ID, type, status, start time, duration, current activity
- **And** I can filter by:
  - Workflow type (CreateEntry, DeleteEntry, ClaimWorkflow, VSYNCDaily)
  - Status (RUNNING, COMPLETED, FAILED, TIMED_OUT, CANCELLED)
  - Date range (start date, end date)
  - Account ID or Key ID
- **And** default view shows last 24 hours, sorted by start time (newest first)
- **And** pagination supports 20, 50, 100 items per page

**AC-003.1.2: Real-Time Workflow Status**
- **Given** I am viewing the workflow list
- **When** a workflow status changes (e.g., RUNNING to COMPLETED)
- **Then** the dashboard updates in real-time (WebSocket or polling every 5 seconds)
- **And** new workflows appear at the top of the list
- **And** completed/failed workflows show completion timestamp

**AC-003.1.3: View Workflow Details**
- **Given** I select a specific workflow from the list
- **When** I click to view details
- **Then** the system displays:
  - Complete workflow execution history
  - All activities executed (with timestamps and durations)
  - Activity inputs and outputs (JSON format)
  - Current workflow state
  - Retry attempts (if any)
  - Error messages and stack traces (for failed activities)
  - Scheduled timers (for ClaimWorkflow: shows countdown to day 30)
  - Event history (complete Temporal event log)

**AC-003.1.4: Search Workflows**
- **Given** I am on the workflow dashboard
- **When** I search by workflow ID, key value, or account number
- **Then** the system returns matching workflows
- **And** highlights the search term in results
- **And** supports partial matching

**AC-003.1.5: Workflow Statistics Summary**
- **Given** I access the workflow dashboard
- **When** the page loads
- **Then** the system displays summary statistics:
  - Total workflows today
  - Running workflows (count)
  - Failed workflows in last 24h (count and percentage)
  - Average workflow duration by type
  - Success rate (last 7 days)
  - Workflows requiring attention (stuck > 1 hour)

**AC-003.1.6: Export Workflow Data**
- **Given** I have filtered workflows
- **When** I click "Export" button
- **Then** the system generates a CSV or JSON file with:
  - All filtered workflows
  - Key fields: ID, type, status, timestamps, account ID, error messages
- **And** downloads the file to my device
- **And** logs the export action for audit

#### Business Rules

**BR-003.1.1: Access Control**
- Only users with role ADMIN or OPERATOR can access workflow dashboard
- Read-only access for OPERATOR role
- Full access (including cancellation) for ADMIN role

**BR-003.1.2: Workflow Types**
- CreateEntryWorkflow: Create DICT key
- DeleteEntryWorkflow: Delete DICT key
- ClaimWorkflow: Portability/Ownership claim (30-day timer)
- VSYNCDailyWorkflow: Daily synchronization with BACEN
- RetryFailedWorkflow: Retry previously failed operations

**BR-003.1.3: Status Classification**
- RUNNING: Workflow in progress
- COMPLETED: Successfully finished
- FAILED: Unrecoverable error
- TIMED_OUT: Exceeded maximum duration
- CANCELLED: Manually cancelled by admin

**BR-003.1.4: Data Retention**
- Workflow data retained for 90 days
- After 90 days, archived to cold storage
- Audit logs retained for 7 years (compliance requirement)

#### Dependencies

- TSP-001: Temporal Workflow Engine TechSpec
- API-001: Core DICT REST API
- SEC-002: Role-Based Access Control (RBAC) policy

#### Technical Notes

- **API Endpoint**: GET /api/v1/admin/workflows
- **Query Parameters**:
  - `type`: enum (workflow type)
  - `status`: enum (workflow status)
  - `startDate`: ISO8601
  - `endDate`: ISO8601
  - `search`: string (workflow ID, key, account)
  - `page`: integer
  - `pageSize`: integer
- **Response Payload**:
  ```json
  {
    "workflows": [
      {
        "workflowId": "uuid",
        "workflowType": "CreateEntryWorkflow",
        "status": "RUNNING",
        "startTime": "ISO8601",
        "duration": 1234 (milliseconds),
        "currentActivity": "CallBackenRSFN",
        "accountId": "uuid",
        "keyId": "uuid",
        "keyValue": "user@example.com"
      }
    ],
    "statistics": {
      "totalToday": 1523,
      "running": 45,
      "failed": 12,
      "avgDuration": 3456,
      "successRate": 98.5
    },
    "pagination": {...}
  }
  ```
- **Integration**: Temporal UI embedded or custom dashboard using Temporal SDK

---

### US-003.2: Cancel Stuck Workflows

**Story ID**: US-003.2
**Priority**: Should Have
**Story Points**: 5
**Sprint**: Sprint 9

#### User Story

As an **Admin**,
I want **to manually cancel stuck or problematic Temporal workflows**,
So that **I can resolve blocking issues and prevent system degradation**.

#### Acceptance Criteria

**AC-003.2.1: Identify Stuck Workflows**
- **Given** I am viewing the workflow dashboard
- **When** a workflow has been running for longer than expected duration
- **Then** the system highlights it with a warning indicator (yellow for > 1 hour, red for > 4 hours)
- **And** provides a "stuck workflows" filter/tab
- **And** shows estimated expected duration vs actual duration

**AC-003.2.2: Cancel Single Workflow**
- **Given** I identify a stuck or problematic workflow
- **When** I click "Cancel Workflow" button
- **Then** the system displays a confirmation dialog with:
  - Workflow ID and type
  - Current activity
  - Warning about potential side effects
  - Reason text field (mandatory)
  - "Cancel" and "Confirm Cancellation" buttons
- **And** after confirmation, sends cancellation request to Temporal
- **And** updates workflow status to "CANCELLED"
- **And** logs the cancellation action with reason and admin user ID

**AC-003.2.3: Bulk Cancel Workflows**
- **Given** I have multiple stuck workflows of the same type
- **When** I select multiple workflows (checkbox) and click "Bulk Cancel"
- **Then** the system displays confirmation dialog with count of selected workflows
- **And** requires a reason (single reason for all)
- **And** after confirmation, cancels all selected workflows
- **And** shows progress indicator during bulk operation
- **And** displays summary report (success count, failure count)

**AC-003.2.4: Cancel with Cleanup Option**
- **Given** I am cancelling a workflow
- **When** I select "Cleanup partial data" checkbox
- **Then** the system executes a cleanup activity before cancelling:
  - Rollback database changes (if possible)
  - Remove orphaned records
  - Update related entities status
- **And** logs all cleanup actions

**AC-003.2.5: Prevent Cancellation of Critical Workflows**
- **Given** I try to cancel a workflow
- **When** the workflow is in a critical state (e.g., ClaimWorkflow on day 29)
- **Then** the system displays a warning message
- **And** requires additional confirmation with "I understand the consequences" checkbox
- **And** sends alert to senior admin (notification)

**AC-003.2.6: Cancellation Audit Trail**
- **Given** I cancel one or more workflows
- **When** the cancellation is completed
- **Then** the system creates audit log entries with:
  - Admin user ID and name
  - Timestamp
  - Workflow IDs cancelled
  - Reason provided
  - Cleanup actions taken
  - IP address and session ID
- **And** audit log is immutable and retained for 7 years

#### Business Rules

**BR-003.2.1: Cancellation Authority**
- Only ADMIN role can cancel workflows
- OPERATOR role can view but not cancel
- Cancellation requires mandatory reason (min 10 characters)

**BR-003.2.2: Stuck Workflow Thresholds**
- CreateEntryWorkflow: > 5 minutes = warning, > 30 minutes = critical
- DeleteEntryWorkflow: > 5 minutes = warning, > 30 minutes = critical
- ClaimWorkflow: > 24 hours without activity = warning
- VSYNCDailyWorkflow: > 2 hours = warning, > 6 hours = critical

**BR-003.2.3: Cleanup Actions**
- Delete pending entries not yet synced with BACEN
- Rollback claim status to previous state
- Notify affected users of cancellation
- Create manual intervention ticket for review

**BR-003.2.4: Post-Cancellation**
- Cancelled workflows can be retried manually (separate action)
- System creates incident report for investigation
- Metrics updated (failure rate, cancellation rate)

#### Dependencies

- TSP-001: Temporal Workflow Engine TechSpec
- API-001: Core DICT REST API
- SEC-002: RBAC policy
- SEC-005: Audit logging specification

#### Technical Notes

- **API Endpoint**: POST /api/v1/admin/workflows/{workflowId}/cancel
- **Request Payload**:
  ```json
  {
    "reason": "string (required, min 10 chars)",
    "cleanup": true | false,
    "acknowledgeConsequences": true (required for critical workflows)
  }
  ```
- **Response Payload**:
  ```json
  {
    "workflowId": "uuid",
    "status": "CANCELLED",
    "cancelledAt": "ISO8601",
    "cancelledBy": "adminUserId",
    "cleanupActions": ["action1", "action2"],
    "affectedEntities": {
      "entries": ["keyId1"],
      "claims": [],
      "accounts": []
    }
  }
  ```
- **Temporal API**: Uses `WorkflowClient.CancelWorkflow()` method

---

### US-003.3: Audit Operations

**Story ID**: US-003.3
**Priority**: Should Have
**Story Points**: 8
**Sprint**: Sprint 9

#### User Story

As an **Admin or Compliance Officer**,
I want **to search and review audit logs of all DICT operations**,
So that **I can ensure compliance, investigate incidents, and detect suspicious activities**.

#### Acceptance Criteria

**AC-003.3.1: Search Audit Logs**
- **Given** I am logged in as admin or compliance officer
- **When** I access the audit log dashboard
- **Then** the system displays a search interface with filters:
  - Date range (from, to)
  - Operation type (CREATE_KEY, DELETE_KEY, CREATE_CLAIM, RESPOND_CLAIM, etc.)
  - User ID or account ID
  - Entity ID (key ID, claim ID)
  - Result (SUCCESS, FAILURE)
  - IP address range
  - Admin user (for admin actions)
- **And** executes search and returns paginated results
- **And** default shows last 7 days, sorted by timestamp (newest first)

**AC-003.3.2: View Audit Log Entry Details**
- **Given** I have search results
- **When** I click on a specific audit log entry
- **Then** the system displays complete details:
  - Timestamp (ISO8601 with milliseconds)
  - Operation type and description
  - User ID, name, CPF/CNPJ (masked for privacy)
  - Account ID and account number
  - Entity ID (key ID, claim ID, etc.)
  - Before state (JSON)
  - After state (JSON)
  - Result (success or failure)
  - Error message (if failed)
  - IP address and geolocation
  - User agent (browser/device)
  - Session ID
  - Request ID (for tracing)
  - Related workflow ID (if applicable)

**AC-003.3.3: Detect Suspicious Activities**
- **Given** I am reviewing audit logs
- **When** the system detects suspicious patterns
- **Then** it highlights entries with warning icons:
  - Multiple failed login attempts (> 5 in 10 minutes)
  - Rapid key creation (> 10 keys in 1 hour by same user)
  - Bulk deletion attempts
  - Access from unusual IP/location
  - Admin actions outside business hours
- **And** allows filtering to show only suspicious activities

**AC-003.3.4: Export Audit Logs**
- **Given** I have filtered audit logs
- **When** I click "Export for Compliance" button
- **Then** the system generates a comprehensive report:
  - CSV or PDF format
  - All filtered entries
  - Includes digital signature (for non-repudiation)
  - Watermark with export timestamp and admin user
  - Summary statistics
- **And** logs the export action itself (audit of audit)

**AC-003.3.5: Audit Log Integrity Verification**
- **Given** I access audit log dashboard
- **When** I click "Verify Integrity" button
- **Then** the system:
  - Checks cryptographic hash chain of audit logs
  - Verifies no logs were tampered or deleted
  - Displays verification result (pass/fail)
  - Shows timestamp of last verification
  - Alerts if integrity check fails

**AC-003.3.6: Compliance Reports**
- **Given** I am a compliance officer
- **When** I request a compliance report for a date range
- **Then** the system generates a report including:
  - Total operations by type
  - Success/failure rates
  - User activity summary (top 10 active users)
  - Admin actions summary
  - BACEN synchronization status
  - LGPD compliance events (data access, deletion requests)
  - Security incidents
- **And** report is formatted per regulatory requirements (BACEN, LGPD)

**AC-003.3.7: Real-Time Audit Monitoring**
- **Given** I enable real-time monitoring mode
- **When** critical operations occur (key deletion, claim response, admin workflow cancellation)
- **Then** the dashboard shows real-time notifications
- **And** allows immediate drill-down into operation details
- **And** supports filtering real-time stream by severity (INFO, WARNING, CRITICAL)

#### Business Rules

**BR-003.3.1: Audit Log Scope**
- **All** user operations logged (create, read, update, delete)
- **All** admin operations logged
- **All** BACEN API calls logged (request + response)
- **All** authentication events logged (login, logout, failed attempts)
- **All** authorization failures logged

**BR-003.3.2: Data Retention**
- Audit logs retained for **7 years** (regulatory requirement)
- Hot storage: Last 90 days (PostgreSQL)
- Warm storage: 91 days to 2 years (S3/archive)
- Cold storage: 2-7 years (Glacier/tape)

**BR-003.3.3: Privacy and Masking**
- Sensitive data (CPF, account numbers) partially masked in list view
- Full data visible only in detail view (with proper authorization)
- Compliance officers have read-only access to full data

**BR-003.3.4: Immutability**
- Audit logs are **append-only** (cannot be modified or deleted)
- Implemented using cryptographic hash chain
- Each log entry contains hash of previous entry
- Tampering detection through integrity verification

**BR-003.3.5: Performance**
- Audit log queries must return within 5 seconds (P95)
- Support for indexing on: timestamp, userId, operation type, entity ID
- Older logs archived but remain searchable (with slower response time)

#### Dependencies

- SEC-005: Audit Logging Specification
- DAT-004: Audit Log Database Schema
- CMP-002: LGPD Compliance Requirements
- API-001: Core DICT REST API

#### Technical Notes

- **API Endpoint**: GET /api/v1/admin/audit-logs
- **Query Parameters**:
  - `startDate`: ISO8601
  - `endDate`: ISO8601
  - `operationType`: enum
  - `userId`: uuid
  - `entityId`: uuid
  - `result`: SUCCESS | FAILURE
  - `page`: integer
  - `pageSize`: integer (max 100)
- **Response Payload**:
  ```json
  {
    "auditLogs": [
      {
        "logId": "uuid",
        "timestamp": "ISO8601",
        "operationType": "CREATE_KEY",
        "userId": "uuid",
        "userName": "João S.*** (masked)",
        "accountId": "uuid",
        "entityId": "keyId",
        "beforeState": {...},
        "afterState": {...},
        "result": "SUCCESS",
        "ipAddress": "192.168.1.100",
        "geoLocation": "São Paulo, Brazil",
        "requestId": "uuid",
        "workflowId": "uuid (optional)",
        "suspicious": false
      }
    ],
    "pagination": {...}
  }
  ```
- **Database**: Dedicated audit log table with append-only constraint
- **Integrity**: SHA-256 hash chain, verified daily by scheduled job

---

## Summary Table

| Story ID | Title | Priority | Story Points | Dependencies |
|----------|-------|----------|--------------|--------------|
| US-003.1 | View All Workflows | Should Have | 5 | TSP-001, API-001, SEC-002 |
| US-003.2 | Cancel Stuck Workflows | Should Have | 5 | TSP-001, API-001, SEC-002, SEC-005 |
| US-003.3 | Audit Operations | Should Have | 8 | SEC-005, DAT-004, CMP-002, API-001 |

**Total Story Points**: 18

---

## Admin Dashboard Wireframe (Text-Based)

```
╔════════════════════════════════════════════════════════════════════╗
║ DICT Admin Dashboard                          User: admin@lbpay.com ║
╠════════════════════════════════════════════════════════════════════╣
║ [Workflows] [Audit Logs] [System Health] [Reports]                 ║
╠════════════════════════════════════════════════════════════════════╣
║                                                                      ║
║ System Statistics (Last 24h)                                        ║
║ ┌──────────────┬──────────────┬──────────────┬──────────────┐      ║
║ │ Total Wflows │ Running      │ Failed       │ Avg Duration │      ║
║ │ 1,523        │ 45           │ 12 (0.8%)    │ 3.4s         │      ║
║ └──────────────┴──────────────┴──────────────┴──────────────┘      ║
║                                                                      ║
║ Active Workflows                                                     ║
║ [Filter: All Types ▼] [Status: Running ▼] [Search: ________]       ║
║                                                                      ║
║ ┌─────────────────────────────────────────────────────────────┐    ║
║ │ ID         Type           Status    Started   Duration  ⚙️  │    ║
║ ├─────────────────────────────────────────────────────────────┤    ║
║ │ abc123... CreateEntry    RUNNING   10:23:45  00:02:34  [▶] │    ║
║ │ def456... ClaimWorkflow  RUNNING   10:20:12  00:05:47  [▶] │    ║
║ │ ghi789... DeleteEntry    🔴STUCK   09:15:33  01:10:26  [✖] │    ║
║ │ jkl012... VSYNCDaily     RUNNING   02:00:01  08:25:58  [▶] │    ║
║ └─────────────────────────────────────────────────────────────┘    ║
║                                                                      ║
║ [<< Previous] Page 1 of 3 [Next >>]                                 ║
╚════════════════════════════════════════════════════════════════════╝
```

---

## Acceptance Checklist

- [ ] All acceptance criteria reviewed by Product Owner
- [ ] Admin role permissions defined in RBAC policy
- [ ] Temporal Workflow monitoring strategy documented
- [ ] Audit log retention policy validated with compliance team
- [ ] Suspicious activity detection rules defined
- [ ] Export formats validated (CSV, PDF)
- [ ] Integrity verification mechanism implemented
- [ ] Performance requirements specified (query response times)
- [ ] Privacy and data masking requirements clear
- [ ] Dependencies on Temporal SDK documented

---

## References

- **TSP-001**: Temporal Workflow Engine TechSpec
- **SEC-002**: Role-Based Access Control (RBAC) Policy
- **SEC-005**: Audit Logging Specification
- **DAT-004**: Audit Log Database Schema
- **CMP-002**: LGPD Compliance Requirements
- **API-001**: Core DICT REST API Specification
- **Temporal Documentation**: https://docs.temporal.io/

---

**Last Updated**: 2025-10-25
**Next Review**: Sprint 9 Planning Session
