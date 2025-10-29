---
description: Comprehensive code review by architect and security auditor
---

Perform comprehensive code review for implemented CID/VSync code.

## Review Process

### Phase 1: Architect Review (`think hard`)

**Invoke**: `ultra-architect-planner`

**Review Checklist**:
```markdown
# Architecture & Patterns Review

## Clean Architecture Compliance
- [ ] Domain layer has NO infrastructure dependencies
- [ ] Application layer depends only on domain
- [ ] Infrastructure implements domain interfaces
- [ ] Dependency direction: Infrastructure → Application → Domain

## Pattern Alignment with connector-dict
- [ ] Domain entities match claim module patterns
- [ ] Repository interfaces in domain layer
- [ ] Use cases in application layer
- [ ] Event handlers follow dict patterns

## Temporal Workflow Patterns
- [ ] Continue-As-New implemented for cron workflow
- [ ] Child workflows with ParentClosePolicy: ABANDON
- [ ] Workflow IDs based on requestID (idempotency)
- [ ] Activities with proper retry policies
- [ ] Error classification (retryable vs non-retryable)

## Event-Driven Patterns
- [ ] Fingerprint-based idempotency
- [ ] Redis caching (24h TTL)
- [ ] Pulsar Nack on processing errors
- [ ] Event schema aligned with Dict API

## Code Quality
- [ ] Go idiomaticity (golangci-lint clean)
- [ ] Proper error handling (wrapped with context)
- [ ] No naked returns in functions >10 lines
- [ ] Exported functions have godoc comments
```

**Output**: `docs/reviews/architecture-review.md`

### Phase 2: Security Audit (`ultrathink`)

**Invoke**: `security-compliance-auditor`

**Audit Checklist**:
```markdown
# Security & Compliance Audit

## BACEN Compliance
- [ ] CID generation follows Chapter 9 spec (SHA-256, field order)
- [ ] VSync calculation correct (XOR cumulative)
- [ ] All operations audited (audit_logs table)
- [ ] Timestamps in UTC ISO-8601
- [ ] Immutable audit trail (append-only)

## LGPD Compliance
- [ ] No PII in logs (CPF, phone, email masked)
- [ ] Personal data encrypted at rest
- [ ] Access control for sensitive data
- [ ] Data retention policy defined

## Security Vulnerabilities
- [ ] SQL injection prevented (parameterized queries)
- [ ] Input validation on all inputs
- [ ] No secrets in code/config
- [ ] Constant-time comparison for sensitive data
- [ ] Approved cryptography only (SHA-256, not MD5)

## Authentication & Authorization
- [ ] gRPC connections authenticated
- [ ] Database connections use least privilege
- [ ] Redis connections authenticated
- [ ] Temporal connections secured
```

**Output**: `docs/security/security-audit.md`

### Phase 3: Code Quality Review (`think`)

**Invoke**: `qa-lead-test-architect`

**Quality Checklist**:
```markdown
# Code Quality & Testing Review

## Test Coverage
- [ ] Overall coverage >80%
- [ ] Domain layer coverage >90%
- [ ] Application layer coverage >85%
- [ ] Infrastructure layer coverage >75%

## Test Quality
- [ ] Unit tests use table-driven approach
- [ ] Integration tests use testcontainers
- [ ] Workflow tests mock activities properly
- [ ] E2E tests cover critical flows
- [ ] No flaky tests (run 10x passes)

## Code Metrics
- [ ] Cyclomatic complexity <10 per function
- [ ] Max function length <50 lines
- [ ] No code duplication >10 lines
- [ ] golangci-lint score A
```

**Output**: `docs/reviews/qa-review.md`

## Review Execution

```bash
# Step 1: Run automated checks
make lint
make test
make coverage-check

# Step 2: Invoke architect review
"ultra-architect-planner, think hard about reviewing the CID/VSync implementation:

Review files:
- internal/domain/cid/*
- internal/application/usecases/cid/*
- internal/infrastructure/temporal/workflows/*
- internal/infrastructure/pulsar/handlers/cid/*

Check:
1. Clean Architecture compliance
2. Pattern alignment with connector-dict
3. Temporal workflow best practices
4. Event-driven patterns

Output: docs/reviews/architecture-review.md"

# Step 3: Invoke security audit
"security-compliance-auditor, ultrathink about security audit:

Audit all code for:
1. BACEN Chapter 9 compliance
2. LGPD compliance
3. Security vulnerabilities (OWASP Top 10)
4. Secrets management
5. Input validation

Output: docs/security/security-audit.md"

# Step 4: Invoke QA review
"qa-lead-test-architect, think about code quality review:

Review:
1. Test coverage metrics
2. Test quality and patterns
3. Code complexity metrics
4. golangci-lint results

Output: docs/reviews/qa-review.md"
```

## Review Report Format

```markdown
# Code Review Report

**Date**: 2024-01-15
**Reviewer**: ultra-architect-planner + security-compliance-auditor + qa-lead-test-architect
**Status**: APPROVED / CHANGES_REQUESTED

## Summary
Brief summary of review findings.

## Critical Issues ❌
Issues that MUST be fixed before approval:
1. [Issue description]
   - Location: file.go:123
   - Impact: [Impact]
   - Fix: [Recommendation]

## Major Issues 🟡
Issues that should be fixed before deployment:
1. [Issue description]

## Minor Issues 🔵
Improvements for future:
1. [Issue description]

## Positive Highlights ✅
Excellent patterns to propagate:
1. [Highlight]

## Approval
- [ ] Architecture Review: APPROVED
- [ ] Security Audit: APPROVED
- [ ] QA Review: APPROVED

**Final Verdict**: [APPROVED / CHANGES_REQUESTED]
```

## Post-Review Actions

If **APPROVED**:
```bash
# Merge to main
git checkout main
git merge feature/cid-vsync
git push origin main

# Tag release
git tag -a v1.0.0 -m "CID/VSync implementation"
git push origin v1.0.0
```

If **CHANGES_REQUESTED**:
```bash
# Fix issues
# Re-run review process
# Repeat until approved
```
