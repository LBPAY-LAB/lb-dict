---
description: Orchestrate complete CID/VSync implementation with adaptive thinking and parallel execution
---

# 🎯 CID/VSync Implementation Orchestrator

You are the **Master Orchestrator** coordinating the specialized agent squad for CID/VSync system implementation.

## 🧠 Thinking Level Determination

Analyze the current task and determine appropriate thinking levels:

```
Complexity Assessment:
- Simple feature (CRUD, basic logic): think
- Integration heavy (Pulsar, gRPC, Temporal): think hard
- Security related (BACEN compliance, audit): ultrathink
- Performance critical (VSync calculation, CID generation): think harder
- Architecture changes (new workflows, domain redesign): ultrathink
```

## Orchestration Plan

### Phase 0: Technical Analysis (CURRENT)
**Duration**: 2-3 days
**Thinking Level**: `think hard`

**Deliverables**:
- [ ] Analyze existing Pulsar events for Entry (key.created, key.updated)
- [ ] Analyze Entry/Key/Account domain structure
- [ ] Verify PostgreSQL connection patterns
- [ ] Coordinate with Bridge team: VSync/CIDList gRPC endpoints
- [ ] Coordinate with Core-Dict team: core-events consumer

**Execute with**:
```
1. ultra-architect-planner (think hard):
   - Analyze connector-dict patterns
   - Map Entry fields needed for CID
   - Design PostgreSQL schema

2. integration-specialist (think):
   - Verify Pulsar event schemas
   - Check Bridge gRPC proto definitions
   - Test Core-Dict event publishing

3. technical-writer (think):
   - Document findings in /docs/architecture/analysis/
   - Create initial architecture diagrams
```

### Phase 1: Domain & Application Layer
**Duration**: 3-4 days
**Thinking Level**: `think hard` (domain modeling)

**Execute in PARALLEL**:

```
Agent 1: go-backend-specialist (think hard)
Task: Implement Domain Layer
- Create domain/cid/ entities (CID, VSync, EntryData)
- Create domain/cid/repository.go interfaces
- Write unit tests (>90% coverage)
Output: internal/domain/cid/*

Agent 2: go-backend-specialist (think)
Task: Implement Application Layer
- Create usecases/cid/create_cid.go
- Create usecases/cid/calculate_vsync.go
- Write unit tests (>85% coverage)
Output: internal/application/usecases/cid/*

Agent 3: qa-lead-test-architect (think)
Task: Create Test Suite
- Design test strategy
- Create test data generators
- Setup testcontainers config
Output: tests/unit/domain/cid/*_test.go
```

**Review with**:
```
ultra-architect-planner (think hard):
- Validate Clean Architecture compliance
- Check pattern alignment with connector-dict
- Approve before next phase

security-compliance-auditor (ultrathink):
- Validate BACEN compliance
- Check PII handling
- Verify audit trail design
```

### Phase 2: PostgreSQL Layer
**Duration**: 2-3 days
**Thinking Level**: `think hard`

**Execute in PARALLEL**:

```
Agent 1: go-backend-specialist (think)
Task: Implement Repositories
- Create postgres/cid_repository.go
- Implement all repository interfaces
- Connection pooling
Output: internal/infrastructure/postgres/cid/

Agent 2: devops-engineer (think)
Task: Database Migrations
- Create 001_create_dict_cids.sql
- Create 002_create_dict_vsyncs.sql
- Create 003_create_dict_sync_verifications.sql
- Create 004_create_dict_reconciliations.sql
Output: migrations/*.sql

Agent 3: qa-lead-test-architect (think hard)
Task: Integration Tests
- PostgreSQL integration tests with testcontainers
- Test all repository methods
- Test transaction handling
Output: tests/integration/postgres/*_test.go
```

### Phase 3: Temporal Workflows
**Duration**: 4-5 days
**Thinking Level**: `think harder` (complex orchestration)

**Execute in PARALLEL**:

```
Agent 1: temporal-workflow-engineer (think harder)
Task: VSyncVerificationWorkflow
- Implement cron workflow (Continue-As-New)
- Handle all 5 key types
- Trigger reconciliation on mismatch
Output: internal/infrastructure/temporal/workflows/vsync_verification.go

Agent 2: temporal-workflow-engineer (think hard)
Task: ReconciliationWorkflow
- Implement child workflow
- ParentClosePolicy: ABANDON
- Multi-activity orchestration
Output: internal/infrastructure/temporal/workflows/reconciliation.go

Agent 3: temporal-workflow-engineer (think)
Task: Activities
- VerifyVSyncActivity
- GetCIDListFromBridgeActivity
- GetLocalCIDsActivity
- ReconstructEntriesActivity
- NotifyCoreDictActivity
Output: internal/infrastructure/temporal/activities/vsync/*

Agent 4: qa-lead-test-architect (think harder)
Task: Workflow Tests
- Test VSyncVerificationWorkflow with mocks
- Test ReconciliationWorkflow
- Test all activities
Output: internal/infrastructure/temporal/workflows/*_test.go
```

**Review with**:
```
ultra-architect-planner (ultrathink):
- Validate workflow design
- Check Continue-As-New implementation
- Verify idempotency patterns

temporal-workflow-engineer (think harder):
- Cross-review workflow implementations
- Validate retry policies
- Check error classification
```

### Phase 4: Integration Layer
**Duration**: 3-4 days
**Thinking Level**: `think hard`

**Execute in PARALLEL**:

```
Agent 1: integration-specialist (think hard)
Task: Pulsar Event Handlers
- HandleKeyCreated
- HandleKeyUpdated
- Idempotency with Redis
- Start Temporal workflows
Output: internal/infrastructure/pulsar/handlers/cid/*

Agent 2: integration-specialist (think)
Task: Bridge gRPC Client
- Implement BridgeClient
- GetVSync(keyType)
- GetCIDList(keyType) with pagination
- Connection pooling
Output: internal/infrastructure/grpc/bridge/*

Agent 3: integration-specialist (think)
Task: Redis Caching
- Implement CIDCache
- CheckProcessed / MarkProcessed
- TTL management (24h)
Output: internal/infrastructure/cache/*

Agent 4: integration-specialist (think)
Task: Core-Dict Notifications
- Pulsar producer for core-events
- Event schema definition
- Error handling
Output: internal/infrastructure/pulsar/producer/*

Agent 5: qa-lead-test-architect (think)
Task: Integration Tests
- Test Pulsar handlers with test broker
- Test gRPC client with mock server
- Test Redis caching
Output: tests/integration/*
```

### Phase 5: Quality Assurance
**Duration**: 2-3 days
**Thinking Level**: `think harder` (comprehensive testing)

**Execute in PARALLEL**:

```
Agent 1: qa-lead-test-architect (think harder)
Task: E2E Tests
- Full flow: Entry → CID creation → VSync calculation
- Daily verification simulation
- Reconciliation flow
Output: tests/e2e/*

Agent 2: security-compliance-auditor (ultrathink)
Task: Security Audit
- BACEN compliance validation
- LGPD compliance check
- Vulnerability scanning
- Audit trail verification
Output: docs/security/audit-report.md

Agent 3: qa-lead-test-architect (think)
Task: Coverage Analysis
- Verify >80% overall coverage
- Identify gaps
- Add missing tests
Output: coverage-report.html
```

### Phase 6: DevOps & Documentation
**Duration**: 2-3 days
**Thinking Level**: `think`

**Execute in PARALLEL**:

```
Agent 1: devops-engineer (think hard)
Task: Docker & Kubernetes
- Create Dockerfile (multi-stage)
- Create k8s manifests
- Setup CI/CD pipeline
Output: Dockerfile, k8s/*, .github/workflows/*

Agent 2: technical-writer (think hard)
Task: Architecture Documentation
- Architecture diagrams (C4 Model)
- Sequence diagrams
- Data flow diagrams
Output: docs/architecture/*

Agent 3: technical-writer (think)
Task: API & Deployment Docs
- API documentation (OpenAPI)
- Deployment guide
- Troubleshooting guide
Output: docs/api/*, docs/deployment/*, docs/troubleshooting/*

Agent 4: devops-engineer (think)
Task: Observability
- Metrics implementation
- Logging setup
- Health checks
- Grafana dashboards
Output: Metrics code, dashboards/*
```

### Phase 7: Production Readiness
**Duration**: 2-3 days
**Thinking Level**: `ultrathink` (critical validation)

**Execute in SEQUENCE**:

```
1. ultra-architect-planner (ultrathink):
   - Final architecture review
   - Validate all patterns
   - Check BACEN compliance
   Output: Architecture sign-off

2. security-compliance-auditor (ultrathink):
   - Final security audit
   - LGPD compliance final check
   - Penetration testing review
   Output: Security sign-off

3. qa-lead-test-architect (think harder):
   - Final test execution
   - Performance benchmarks
   - Load testing
   Output: QA sign-off

4. devops-engineer (ultrathink):
   - Production deployment plan
   - Rollback procedures
   - Monitoring validation
   Output: Deployment sign-off
```

## Execution Commands

### Start Phase
```
/implement-phase <phase-number>
```

### Check Phase Progress
```
/phase-status <phase-number>
```

### Agent Coordination
```
"Think hard about coordinating agents for Phase X:

Context:
- Current phase: [X]
- Dependencies: [List]
- Parallel vs Sequential: [Decision]

Agents to invoke:
1. [Agent name] - [Task] - [Thinking level]
2. [Agent name] - [Task] - [Thinking level]

Execute in [PARALLEL/SEQUENCE]"
```

## Quality Gates

Each phase must pass:
- [ ] Code review by architect (pattern compliance)
- [ ] Security review (BACEN/LGPD compliance)
- [ ] Test coverage >80%
- [ ] All agents completed their tasks
- [ ] Documentation updated

## Adaptive Thinking Protocol

```
IF phase involves BACEN compliance:
    thinking_level = ultrathink

IF phase involves Temporal workflows:
    thinking_level = think harder

IF phase involves integration:
    thinking_level = think hard

IF phase involves simple implementation:
    thinking_level = think

IF uncertain about complexity:
    thinking_level++ (escalate)
```

## Communication Between Agents

```json
{
  "from": "ultra-architect-planner",
  "to": "go-backend-specialist",
  "thinking_level_applied": "think hard",
  "task": "Implement domain/cid/cid.go",
  "requirements": {
    "fields": ["participant", "account", "key", "key_type", ...],
    "methods": ["GenerateCID()", "Validate()"],
    "patterns": "Study domain/claim/claim.go for reference"
  },
  "acceptance_criteria": [
    "Unit tests >90% coverage",
    "golangci-lint passes",
    "Follows connector-dict patterns"
  ]
}
```

---

## Current Status

**Current Phase**: 0 - Technical Analysis
**Next Action**: Execute Phase 0 with ultra-architect-planner

**Command to start**:
```
Think hard about Phase 0 - Technical Analysis:

Analyze connector-dict existing patterns and answer:
1. What Pulsar events exist for Entry operations?
2. What Entry fields are needed for CID generation?
3. What PostgreSQL patterns does connector-dict use?
4. Is Bridge gRPC VSync endpoint ready?
5. Does Core-Dict consume core-events topic?

Document findings in /docs/architecture/analysis/phase0-findings.md
```
