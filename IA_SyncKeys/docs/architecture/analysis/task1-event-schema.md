# Task 1: Pulsar Event Schema Analysis

## Overview
Analysis of Pulsar event patterns in connector-dict to understand how to publish Entry events to `persistent://lb-conn/dict/dict-events` topic.

## Current Architecture

### Event Flow
1. **API Request** → Dict API receives Entry operations (create/update/delete)
2. **gRPC Call** → Dict API calls Bridge via gRPC to execute BACEN operations
3. **Response Processing** → Bridge returns BACEN response
4. **Event Publishing** → Orchestration-worker publishes events to dict-events topic

### Topic Configuration
- **Topic**: `persistent://lb-conn/dict/dict-events`
- **Subscription Model**: Shared (multiple consumers can process events)
- **Environment Variable**: `PULSAR_TOPIC_DICT_EVENTS`

## Event Publishing Pattern

### Current Implementation (Claims)
The system publishes events through Temporal workflows:

```go
// Workflow publishes event after successful operation
if err := workflows.ExecuteDictEventsPublishActivity(ctx, correlationID, action, payload); err != nil {
    workflow.GetLogger(ctx).Error("DictEventsPublishActivity failed", "error", err)
    return err
}
```

### Event Properties
Every event includes standard properties:
```go
type MessageProperties struct {
    CorrelationID string    // Unique request identifier
    Action        Action    // Event type (e.g., "key.created", "key.updated")
}
```

### Payload Structure
Events publish the entire BACEN response as payload:
- For Claims: `pkgClaim.CreateClaimResponse` / `pkgClaim.ConfirmClaimResponse`
- For Entry (expected): `pkgDirectory.CreateEntryResponse` / `pkgDirectory.UpdateEntryResponse`

## Entry Events Schema (Proposed)

### Event Types (Actions)
Based on claim patterns, Entry events should use:
- `key.created` - When a new Entry is created
- `key.updated` - When an Entry is updated
- `key.deleted` - When an Entry is deleted

### Event Payload Structure
For Entry events, the payload would be the BACEN response:

```go
// CreateEntryResponse from BACEN
type CreateEntryResponse struct {
    Entry         Entry          // Complete entry data
    CorrelationID *string
    ResponseTime  *time.Time
}

// Entry structure contains all CID-relevant fields
type Entry struct {
    Key              string        // PIX key value
    KeyType          KeyType       // CPF, CNPJ, PHONE, EMAIL, EVP
    Account          Account       // Bank account details
    Owner            Person        // Owner information
    CreationDate     time.Time
    KeyOwnershipDate time.Time
}
```

## Key Findings

### 1. Event Publishing Location
Currently, events are NOT published directly from the Dict API. Instead:
- Dict API makes gRPC calls to Bridge
- Orchestration-worker handles async operations via Temporal
- Events are published from Temporal workflows

### 2. Missing Entry Event Publishing
- No Entry workflows exist in orchestration-worker
- Entry operations are synchronous (direct gRPC calls)
- No events are published for Entry operations currently

### 3. Data Availability for CID
The Entry response contains all necessary fields for CID generation:
- ✅ Key value and type
- ✅ Account information (participant, branch, account number)
- ✅ Owner information (tax ID, name)
- ✅ Timestamps (creation date, ownership date)

## Implementation Considerations

### Option 1: Direct Publishing from Dict API
- Modify Dict API to publish events after successful Entry operations
- Add publisher to Entry application service
- Simpler architecture, but couples API with event publishing

### Option 2: Create Entry Workflows (Like Claims)
- Add Entry workflows to orchestration-worker
- Publish events from workflows
- Consistent with existing pattern, but adds complexity

### Option 3: Dedicated VSync Consumer (Recommended)
- Create new container `apps/dict.vsync/`
- Subscribe to `persistent://lb-conn/dict/dict-events`
- Process Entry events to generate CID
- Separate concern, scalable, follows microservices pattern

## Event Format Example

```json
{
  "properties": {
    "correlation_id": "550e8400-e29b-41d4-a716-446655440000",
    "action": "key.created"
  },
  "payload": {
    "entry": {
      "key": "11122233344",
      "keyType": "CPF",
      "account": {
        "participant": "12345678",
        "branch": "0001",
        "accountNumber": "123456",
        "accountType": "CACC",
        "openingDate": "2024-01-15T10:00:00Z"
      },
      "owner": {
        "type": "NATURAL_PERSON",
        "taxIDNumber": "11122233344",
        "name": "João Silva"
      },
      "creationDate": "2024-10-29T14:30:00Z",
      "keyOwnershipDate": "2024-10-29T14:30:00Z"
    },
    "correlationId": "550e8400-e29b-41d4-a716-446655440000",
    "responseTime": "2024-10-29T14:30:01Z"
  }
}
```

## Conclusion

The event schema is well-defined through the SDK types. The main challenge is that Entry events are NOT currently being published. The VSync system will need to:

1. **Option A**: Modify Dict API to publish Entry events
2. **Option B**: Create Entry workflows in orchestration-worker
3. **Option C**: Implement event publishing in the Bridge response handler

The data structure from BACEN responses contains all necessary fields for CID generation per the BACEN specification.

## Next Steps
1. Confirm with stakeholders which option to implement for Entry event publishing
2. Design the VSync consumer to process these events
3. Map Entry fields to CID structure per BACEN specification