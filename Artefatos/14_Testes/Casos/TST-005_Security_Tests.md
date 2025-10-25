# TST-005: Security Tests

**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: QA Team / Security Team
**Status**: ✅ Completo

---

## Sumário Executivo

Este documento apresenta os **test cases de segurança** para o sistema DICT LBPay, cobrindo autenticação, autorização, validação de entrada, criptografia e conformidade com OWASP Top 10.

**Objetivo**: Garantir que o sistema DICT é seguro contra vulnerabilidades comuns e atende aos requisitos de segurança bancária.

**Cobertura**:
- Autenticação: JWT validation, token expiration, refresh token
- Autorização: RBAC (Role-Based Access Control), scope validation
- Input Validation: SQL injection, XSS, command injection, path traversal
- mTLS: Certificate validation, expired certificates, untrusted CA
- OWASP Top 10 (2021) compliance
- Data Protection: Encryption at rest and in transit
- API Security: Rate limiting, CORS, CSP

**Ferramentas**:
- OWASP ZAP (Zed Attack Proxy)
- Burp Suite
- SQLMap
- JWT.io
- OpenSSL
- Custom security test scripts

**Referências**:
- [API-002: Core DICT REST API](../../04_APIs/REST/API-002_Core_DICT_REST_API.md)
- [TST-001: Test Cases CreateEntry](./TST-001_Test_Cases_CreateEntry.md)
- OWASP Top 10: https://owasp.org/Top10/

---

## Test Environment Setup

### Environment Configuration
```yaml
Environment: security-test
Core DICT API: https://dict-api-sec.lbpay.com.br
Auth Service: https://auth-sec.lbpay.com.br
Database: PostgreSQL 16.4 (isolated security test DB)
WAF: AWS WAF (enabled with logging)
mTLS: Enabled (client certificate required for bridge)
```

### Test Tools Installation

#### OWASP ZAP
```bash
# macOS
brew install --cask owasp-zap

# Ubuntu
sudo snap install zaproxy --classic

# Docker
docker pull owasp/zap2docker-stable
```

#### Burp Suite
```bash
# Download from https://portswigger.net/burp/communitydownload
# Community Edition is free

# Configure proxy: localhost:8080
```

#### SQLMap
```bash
# macOS
brew install sqlmap

# Ubuntu
sudo apt-get install sqlmap

# Docker
docker pull paoloo/sqlmap
```

### Test Users
```yaml
Admin User:
  email: admin.security@lbpay.com.br
  password: AdminSec@1234
  scopes: [dict:read, dict:write, dict:admin]
  role: admin

Regular User:
  email: user.security@lbpay.com.br
  password: UserSec@1234
  scopes: [dict:read, dict:write]
  role: customer

Read-Only User:
  email: readonly.security@lbpay.com.br
  password: ReadSec@1234
  scopes: [dict:read]
  role: customer

Attacker User (for testing):
  email: attacker@lbpay.com.br
  password: Attacker@1234
  scopes: [dict:read]
  role: customer
```

---

## SEC-001: Authentication - JWT Token Validation

**Priority**: P0 (Critical)
**Type**: Security Test - Authentication
**OWASP**: A07:2021 - Identification and Authentication Failures

### Test Objective
Verify JWT token validation is implemented correctly and cannot be bypassed.

### Test Cases

#### SEC-001-001: Valid JWT Token
```gherkin
Given the user has a valid JWT token
When the user sends a request with the token
Then the request is authenticated successfully
And the user's identity is correctly extracted from token
```

**Steps**:
1. Login and obtain valid JWT token
2. Send GET /api/v1/keys with valid token
3. Verify 200 OK response

**Expected**: ✅ Request authenticated

#### SEC-001-002: Invalid JWT Token Signature
```gherkin
Given the attacker has a JWT token with modified signature
When the attacker sends a request with the tampered token
Then the API returns 401 Unauthorized
And the error is "Invalid token signature"
```

**Steps**:
1. Obtain valid JWT token
2. Modify token signature:
   ```bash
   # Original token
   eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c

   # Modified signature (last part)
   eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.MODIFIED_SIGNATURE_HERE
   ```
3. Send request with tampered token
4. Verify 401 Unauthorized

**Expected**: ✅ 401 Unauthorized

#### SEC-001-003: Expired JWT Token
```gherkin
Given the user has an expired JWT token
When the user sends a request with expired token
Then the API returns 401 Unauthorized
And the error is "Token expired"
```

**Steps**:
1. Use token that expired 1 hour ago (or modify exp claim)
2. Send GET /api/v1/keys
3. Verify response:
   ```json
   {
     "error": "Unauthorized",
     "message": "Token expired",
     "expired_at": "2025-10-25T10:00:00Z"
   }
   ```

**Expected**: ✅ 401 Unauthorized with clear message

#### SEC-001-004: JWT Token with "none" Algorithm
```gherkin
Given an attacker crafts a JWT with algorithm "none"
When the attacker sends request with this token
Then the API rejects the token
And returns 401 Unauthorized
```

**Steps**:
1. Craft JWT with "none" algorithm:
   ```json
   {
     "alg": "none",
     "typ": "JWT"
   }
   ```
2. Send request with this token
3. Verify rejected

**Expected**: ✅ Token rejected (algorithm "none" not allowed)

#### SEC-001-005: JWT Token Replay Attack
```gherkin
Given a valid JWT token was logged out
When an attacker replays the same token
Then the API rejects the token
And returns 401 Unauthorized
```

**Steps**:
1. Login and obtain token
2. Logout (token should be blacklisted)
3. Attempt to use same token
4. Verify rejected

**Expected**: ✅ Token blacklisted and rejected

#### SEC-001-006: Missing Authorization Header
```gherkin
Given the client does not send Authorization header
When the client sends a protected request
Then the API returns 401 Unauthorized
```

**Steps**:
```bash
curl -X GET https://dict-api-sec.lbpay.com.br/api/v1/keys \
  -H "Content-Type: application/json"
```

**Expected**: ✅ 401 Unauthorized

#### SEC-001-007: Malformed Authorization Header
```gherkin
Given the Authorization header is malformed
When the client sends request
Then the API returns 401 Unauthorized
```

**Test Data**:
```bash
# Missing "Bearer" prefix
Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

# Extra spaces
Authorization: Bearer  eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

# Wrong scheme
Authorization: Basic eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Expected**: ✅ All rejected with 401

**Status**: ⬜ Not Run

---

## SEC-002: Authentication - Token Expiration & Refresh

**Priority**: P0 (Critical)
**Type**: Security Test - Authentication
**OWASP**: A07:2021 - Identification and Authentication Failures

### Test Objective
Verify token expiration and refresh token mechanism work securely.

#### SEC-002-001: Access Token Expires After 15 Minutes
```gherkin
Given an access token was issued 15 minutes ago
When the user attempts to use the token
Then the API returns 401 Unauthorized
And suggests using refresh token
```

**Steps**:
1. Login and obtain access token
2. Wait 15 minutes (or modify token exp claim to simulate)
3. Attempt to use token
4. Verify expired

**Expected**: ✅ Token expires after 15 minutes

#### SEC-002-002: Refresh Token to Obtain New Access Token
```gherkin
Given the user has a valid refresh token
When the user sends refresh token to /auth/refresh
Then a new access token is issued
And the new token is valid for 15 minutes
```

**Steps**:
```bash
POST /auth/refresh
Content-Type: application/json

{
  "refresh_token": "abc123..."
}
```

**Expected Response**:
```json
{
  "access_token": "new-token...",
  "refresh_token": "new-refresh-token...",
  "expires_in": 900
}
```

#### SEC-002-003: Refresh Token Single Use (Rotation)
```gherkin
Given the user used a refresh token to get new access token
When the user attempts to reuse the same refresh token
Then the API rejects the request
And invalidates all tokens for that user (security measure)
```

**Steps**:
1. Use refresh token to get new access token
2. Attempt to use same refresh token again
3. Verify rejection and token invalidation

**Expected**: ✅ Refresh token rotation enforced

#### SEC-002-004: Refresh Token Expires After 7 Days
```gherkin
Given a refresh token was issued 7 days ago
When the user attempts to refresh
Then the API returns 401 Unauthorized
And requires re-authentication
```

**Expected**: ✅ Refresh token expires after 7 days

**Status**: ⬜ Not Run

---

## SEC-003: Authorization - RBAC & Scope Validation

**Priority**: P0 (Critical)
**Type**: Security Test - Authorization
**OWASP**: A01:2021 - Broken Access Control

### Test Objective
Verify Role-Based Access Control and scope validation prevent unauthorized actions.

#### SEC-003-001: Read-Only User Cannot Create Entry
```gherkin
Given the user has scope "dict:read" only
When the user attempts to POST /api/v1/keys
Then the API returns 403 Forbidden
And the error is "Required scope: dict:write"
```

**Steps**:
1. Login as readonly user
2. Attempt to create entry:
   ```bash
   POST /api/v1/keys
   Authorization: Bearer {readonly_token}

   {
     "key_type": "CPF",
     "key_value": "12345678900",
     "account": { ... }
   }
   ```
3. Verify 403 Forbidden

**Expected**: ✅ 403 Forbidden

#### SEC-003-002: Regular User Cannot Access Admin Endpoint
```gherkin
Given the user does not have "dict:admin" scope
When the user attempts to access admin endpoint
Then the API returns 403 Forbidden
```

**Test Data**:
```bash
# Attempt to delete any key (admin only)
DELETE /api/v1/admin/keys/CPF/12345678900
Authorization: Bearer {regular_user_token}
```

**Expected**: ✅ 403 Forbidden

#### SEC-003-003: User Cannot Access Another User's Data
```gherkin
Given User A has entry with entry_id "123"
When User B attempts to access entry "123"
Then the API returns 404 Not Found (hiding existence)
```

**Steps**:
1. User A creates entry, gets entry_id "123"
2. User B attempts:
   ```bash
   GET /api/v1/entries/123
   Authorization: Bearer {user_b_token}
   ```
3. Verify 404 Not Found (not 403, to avoid data leak)

**Expected**: ✅ 404 Not Found

#### SEC-003-004: Scope Tampering in JWT
```gherkin
Given an attacker modifies JWT payload to add "dict:admin" scope
When the attacker sends request with tampered token
Then the API rejects the token (signature invalid)
```

**Steps**:
1. Decode JWT payload
2. Modify scopes: `["dict:read", "dict:write", "dict:admin"]`
3. Re-encode (signature will be invalid)
4. Send request
5. Verify rejected due to signature mismatch

**Expected**: ✅ Token rejected

#### SEC-003-005: Horizontal Privilege Escalation
```gherkin
Given User A owns CPF key "12345678900"
When User B attempts to delete User A's key
Then the API returns 403 Forbidden
```

**Expected**: ✅ 403 Forbidden (cannot delete other user's keys)

#### SEC-003-006: Vertical Privilege Escalation
```gherkin
Given regular user attempts to promote themselves to admin
When they call admin-only endpoints
Then all requests return 403 Forbidden
```

**Expected**: ✅ Privilege escalation prevented

**Status**: ⬜ Not Run

---

## SEC-004: Input Validation - SQL Injection

**Priority**: P0 (Critical)
**Type**: Security Test - Input Validation
**OWASP**: A03:2021 - Injection

### Test Objective
Verify system is protected against SQL injection attacks.

#### SEC-004-001: SQL Injection in Key Value Parameter
```gherkin
Given an attacker sends SQL injection payload in key_value
When the API processes the request
Then the payload is treated as string (not executed)
And no SQL injection occurs
```

**Test Payloads**:
```json
{
  "key_type": "CPF",
  "key_value": "12345678900' OR '1'='1",
  "account": { ... }
}

{
  "key_type": "EMAIL",
  "key_value": "user@example.com'; DROP TABLE entries; --",
  "account": { ... }
}

{
  "key_type": "CPF",
  "key_value": "12345678900' UNION SELECT * FROM users --",
  "account": { ... }
}
```

**Steps**:
1. Send each payload
2. Verify API returns 400 Bad Request (invalid format)
3. Verify database is not affected
4. Check logs for SQL errors (should be none)

**Expected**: ✅ All payloads rejected, no SQL injection

#### SEC-004-002: SQL Injection in GET Parameter
```gherkin
Given an attacker sends SQL injection in URL parameter
When the API processes the GET request
Then the payload is sanitized
And no SQL injection occurs
```

**Test URLs**:
```bash
GET /api/v1/keys/CPF/12345678900'%20OR%20'1'='1
GET /api/v1/keys/EMAIL/user@example.com';%20DROP%20TABLE%20entries;%20--
```

**Expected**: ✅ Payloads sanitized or rejected

#### SEC-004-003: Blind SQL Injection Time-Based
```gherkin
Given an attacker attempts time-based blind SQL injection
When the attacker sends payload with SLEEP/WAITFOR
Then the request completes normally (< 1 second)
And no SQL delay occurs
```

**Test Payloads**:
```bash
GET /api/v1/keys/CPF/12345678900'%20AND%20SLEEP(10)%20--
GET /api/v1/keys/CPF/12345678900'%20WAITFOR%20DELAY%20'00:00:10'%20--
```

**Verification**:
- Response time should be < 1 second
- No database sleep/delay executed

**Expected**: ✅ No time-based injection possible

#### SEC-004-004: Second-Order SQL Injection
```gherkin
Given an attacker stores SQL injection payload in database
When the data is retrieved and used in another query
Then the payload is still treated as string
And no SQL injection occurs
```

**Steps**:
1. Create entry with email: `attacker+'; DROP TABLE entries; --@example.com`
2. Retrieve entry and use in another operation
3. Verify no SQL injection

**Expected**: ✅ Stored payloads do not execute

#### SEC-004-005: Automated SQL Injection Scan with SQLMap
```bash
# Run SQLMap against API
sqlmap -u "https://dict-api-sec.lbpay.com.br/api/v1/keys/CPF/12345678900" \
       --headers="Authorization: Bearer {token}" \
       --batch \
       --level=5 \
       --risk=3

# Check for vulnerabilities
```

**Expected**: ✅ No SQL injection vulnerabilities found

**Status**: ⬜ Not Run

---

## SEC-005: Input Validation - XSS (Cross-Site Scripting)

**Priority**: P0 (Critical)
**Type**: Security Test - Input Validation
**OWASP**: A03:2021 - Injection

### Test Objective
Verify system is protected against XSS attacks.

#### SEC-005-001: Stored XSS in Key Value
```gherkin
Given an attacker sends XSS payload in key_value
When the data is stored and retrieved
Then the payload is HTML-encoded
And no script execution occurs
```

**Test Payloads**:
```json
{
  "key_type": "EMAIL",
  "key_value": "user+<script>alert('XSS')</script>@example.com",
  "account": { ... }
}

{
  "key_type": "EMAIL",
  "key_value": "user+<img src=x onerror=alert('XSS')>@example.com",
  "account": { ... }
}
```

**Verification**:
1. Send payloads
2. Retrieve data via API
3. Verify response is HTML-encoded:
   ```json
   {
     "key_value": "user+&lt;script&gt;alert('XSS')&lt;/script&gt;@example.com"
   }
   ```
4. Verify no script execution in browser

**Expected**: ✅ Payloads HTML-encoded

#### SEC-005-002: Reflected XSS in Error Messages
```gherkin
Given an attacker sends XSS payload that triggers error
When the API returns error message
Then the payload is sanitized in error response
```

**Test**:
```bash
GET /api/v1/keys/CPF/<script>alert('XSS')</script>
```

**Expected Response**:
```json
{
  "error": "Invalid key format",
  "key_value": "&lt;script&gt;alert('XSS')&lt;/script&gt;"
}
```

**Expected**: ✅ Error messages sanitized

#### SEC-005-003: DOM-Based XSS
```gherkin
Given the frontend renders API data
When XSS payload is in API response
Then the frontend sanitizes before rendering
And no script execution occurs
```

**Expected**: ✅ Frontend sanitizes API responses

#### SEC-005-004: CSP (Content Security Policy) Headers
```gherkin
Given the API returns HTML responses (if any)
When the response is sent
Then CSP headers are present
And prevent inline script execution
```

**Verification**:
```bash
curl -I https://dict-api-sec.lbpay.com.br
```

**Expected Headers**:
```
Content-Security-Policy: default-src 'self'; script-src 'self'; object-src 'none'
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
```

**Expected**: ✅ Security headers present

**Status**: ⬜ Not Run

---

## SEC-006: Input Validation - Command Injection

**Priority**: P0 (Critical)
**Type**: Security Test - Input Validation
**OWASP**: A03:2021 - Injection

### Test Objective
Verify system is protected against command injection.

#### SEC-006-001: Command Injection in Input Fields
```gherkin
Given an attacker sends command injection payload
When the API processes the input
Then the payload is treated as string
And no OS command is executed
```

**Test Payloads**:
```json
{
  "key_type": "EMAIL",
  "key_value": "user@example.com; ls -la",
  "account": { ... }
}

{
  "key_type": "EMAIL",
  "key_value": "user@example.com | cat /etc/passwd",
  "account": { ... }
}

{
  "key_type": "EMAIL",
  "key_value": "user@example.com && rm -rf /",
  "account": { ... }
}
```

**Expected**: ✅ All payloads rejected or treated as strings

#### SEC-006-002: Path Traversal Attack
```gherkin
Given an attacker attempts path traversal
When the attacker sends ../ sequences
Then the API rejects the request
```

**Test Payloads**:
```bash
GET /api/v1/keys/../../etc/passwd
GET /api/v1/files?path=../../../../etc/passwd
```

**Expected**: ✅ Path traversal prevented

**Status**: ⬜ Not Run

---

## SEC-007: mTLS Certificate Validation

**Priority**: P0 (Critical)
**Type**: Security Test - TLS/SSL
**OWASP**: A02:2021 - Cryptographic Failures

### Test Objective
Verify mutual TLS (mTLS) certificate validation works correctly for bridge communication.

#### SEC-007-001: Valid Client Certificate Accepted
```gherkin
Given the client presents a valid certificate signed by trusted CA
When the client connects to bridge
Then the connection is established
And communication proceeds
```

**Steps**:
```bash
curl --cert client.crt --key client.key \
     --cacert ca.crt \
     https://dict-bridge-sec.lbpay.com.br/health
```

**Expected**: ✅ 200 OK

#### SEC-007-002: Missing Client Certificate Rejected
```gherkin
Given the client does not present certificate
When the client attempts to connect
Then the connection is rejected
And TLS handshake fails
```

**Steps**:
```bash
curl https://dict-bridge-sec.lbpay.com.br/health
```

**Expected**: ✅ Connection rejected, SSL handshake error

#### SEC-007-003: Expired Client Certificate Rejected
```gherkin
Given the client certificate has expired
When the client connects
Then the connection is rejected
```

**Steps**:
1. Generate expired certificate:
   ```bash
   openssl req -x509 -newkey rsa:2048 -keyout expired.key -out expired.crt \
               -days -1 -nodes -subj "/CN=expired"
   ```
2. Attempt connection:
   ```bash
   curl --cert expired.crt --key expired.key \
        https://dict-bridge-sec.lbpay.com.br/health
   ```

**Expected**: ✅ Connection rejected (certificate expired)

#### SEC-007-004: Untrusted CA Certificate Rejected
```gherkin
Given the client certificate is signed by untrusted CA
When the client connects
Then the connection is rejected
```

**Steps**:
1. Generate self-signed certificate (not in trusted CA list)
2. Attempt connection

**Expected**: ✅ Connection rejected (untrusted CA)

#### SEC-007-005: Revoked Certificate Rejected (OCSP/CRL)
```gherkin
Given the client certificate is revoked
When the client connects
Then the connection is rejected via OCSP/CRL check
```

**Steps**:
1. Revoke certificate in CA
2. Attempt connection
3. Verify OCSP/CRL check rejects

**Expected**: ✅ Revoked certificate rejected

#### SEC-007-006: Certificate Common Name (CN) Validation
```gherkin
Given the client certificate CN does not match expected
When the client connects
Then the connection may be rejected (depending on config)
```

**Expected**: ✅ CN validation enforced

**Status**: ⬜ Not Run

---

## SEC-008: Data Protection - Encryption at Rest

**Priority**: P1 (High)
**Type**: Security Test - Data Protection
**OWASP**: A02:2021 - Cryptographic Failures

### Test Objective
Verify sensitive data is encrypted at rest in database.

#### SEC-008-001: Sensitive Fields Encrypted in Database
```gherkin
Given sensitive data is stored in database
When querying database directly
Then sensitive fields are encrypted
And cannot be read in plaintext
```

**Verification**:
```sql
-- Check if CPF is encrypted
SELECT key_value FROM dict.entries WHERE key_type = 'CPF' LIMIT 1;
-- Should NOT return plaintext CPF like "12345678900"
-- Should return encrypted value like "AES:abc123..."

-- Check if account numbers are encrypted
SELECT account_number FROM dict.entries LIMIT 1;
-- Should be encrypted
```

**Expected**: ✅ Sensitive fields encrypted

#### SEC-008-002: Encryption Key Rotation
```gherkin
Given encryption keys need rotation
When keys are rotated
Then old data can still be decrypted
And new data uses new key
```

**Expected**: ✅ Key rotation supported

**Status**: ⬜ Not Run

---

## SEC-009: Data Protection - Encryption in Transit

**Priority**: P0 (Critical)
**Type**: Security Test - Data Protection
**OWASP**: A02:2021 - Cryptographic Failures

### Test Objective
Verify all communication uses TLS 1.2+ with strong ciphers.

#### SEC-009-001: TLS Version Enforcement
```gherkin
Given the API requires TLS 1.2+
When client attempts TLS 1.0 or 1.1
Then the connection is rejected
```

**Test**:
```bash
# Attempt TLS 1.0 (should fail)
openssl s_client -connect dict-api-sec.lbpay.com.br:443 -tls1

# Attempt TLS 1.1 (should fail)
openssl s_client -connect dict-api-sec.lbpay.com.br:443 -tls1_1

# Attempt TLS 1.2 (should succeed)
openssl s_client -connect dict-api-sec.lbpay.com.br:443 -tls1_2
```

**Expected**: ✅ Only TLS 1.2+ allowed

#### SEC-009-002: Weak Cipher Suites Disabled
```gherkin
Given the API disables weak cipher suites
When testing cipher suites
Then only strong ciphers are available
```

**Test**:
```bash
nmap --script ssl-enum-ciphers -p 443 dict-api-sec.lbpay.com.br
```

**Expected Ciphers**: ✅ Only strong ciphers (e.g., ECDHE-RSA-AES256-GCM-SHA384)

#### SEC-009-003: HTTP Strict Transport Security (HSTS)
```gherkin
Given the API enforces HTTPS
When the client connects
Then HSTS header is present
```

**Verification**:
```bash
curl -I https://dict-api-sec.lbpay.com.br
```

**Expected Header**:
```
Strict-Transport-Security: max-age=31536000; includeSubDomains
```

**Expected**: ✅ HSTS header present

**Status**: ⬜ Not Run

---

## SEC-010: API Security - Rate Limiting

**Priority**: P1 (High)
**Type**: Security Test - API Security
**OWASP**: A05:2021 - Security Misconfiguration

### Test Objective
Verify rate limiting prevents abuse.

#### SEC-010-001: Rate Limit Enforcement - 100 req/min
```gherkin
Given the API has rate limit of 100 requests per minute
When the user sends 101 requests in 1 minute
Then the 101st request returns 429 Too Many Requests
```

**Test Script**:
```bash
#!/bin/bash
for i in {1..101}; do
  RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Authorization: Bearer ${TOKEN}" \
    https://dict-api-sec.lbpay.com.br/api/v1/keys)

  echo "Request $i: $RESPONSE"

  if [ "$RESPONSE" = "429" ]; then
    echo "✅ Rate limit enforced at request $i"
    break
  fi
done
```

**Expected**: ✅ 429 after 100 requests

#### SEC-010-002: Rate Limit Headers
```gherkin
Given the API enforces rate limits
When the user sends request
Then rate limit headers are returned
```

**Expected Headers**:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 99
X-RateLimit-Reset: 1635177600
Retry-After: 60
```

**Expected**: ✅ Rate limit headers present

#### SEC-010-003: DDoS Protection
```gherkin
Given the API has DDoS protection (WAF)
When massive traffic is detected
Then requests are throttled or blocked
```

**Expected**: ✅ WAF active and blocking malicious traffic

**Status**: ⬜ Not Run

---

## SEC-011: OWASP Top 10 (2021) Compliance

**Priority**: P0 (Critical)
**Type**: Security Compliance
**OWASP**: Full Top 10 Coverage

### OWASP Top 10 Checklist

| # | Vulnerability | Status | Test Case |
|---|---------------|--------|-----------|
| A01:2021 | Broken Access Control | ⬜ | SEC-003 (RBAC) |
| A02:2021 | Cryptographic Failures | ⬜ | SEC-007 (mTLS), SEC-008/009 (Encryption) |
| A03:2021 | Injection | ⬜ | SEC-004 (SQL), SEC-005 (XSS), SEC-006 (Command) |
| A04:2021 | Insecure Design | ⬜ | Architecture review |
| A05:2021 | Security Misconfiguration | ⬜ | SEC-009 (TLS), SEC-010 (Rate Limit) |
| A06:2021 | Vulnerable Components | ⬜ | Dependency scan (Snyk/Dependabot) |
| A07:2021 | Identification & Authentication Failures | ⬜ | SEC-001/002 (JWT) |
| A08:2021 | Software & Data Integrity Failures | ⬜ | Code signing, artifact verification |
| A09:2021 | Security Logging & Monitoring Failures | ⬜ | SEC-012 (Audit Logs) |
| A10:2021 | Server-Side Request Forgery (SSRF) | ⬜ | SEC-013 (SSRF) |

---

## SEC-012: Security Logging & Monitoring

**Priority**: P1 (High)
**Type**: Security Test - Logging
**OWASP**: A09:2021 - Security Logging and Monitoring Failures

### Test Objective
Verify security events are logged and monitored.

#### SEC-012-001: Failed Authentication Logged
```gherkin
Given a user fails authentication
When the failure occurs
Then the event is logged with details
And alerts are triggered after threshold
```

**Verification**:
```bash
# Check audit logs
SELECT * FROM dict.audit_logs
WHERE action = 'LOGIN_FAILED'
AND created_at > NOW() - INTERVAL '1 hour';
```

**Expected**: ✅ Failed login logged

#### SEC-012-002: Unauthorized Access Attempts Logged
```gherkin
Given a user attempts unauthorized action (403)
When the attempt occurs
Then the event is logged
And includes user_id, action, timestamp
```

**Expected**: ✅ 403 attempts logged

#### SEC-012-003: Security Event Alerting
```gherkin
Given 5 failed login attempts in 1 minute
When threshold is exceeded
Then alert is sent to security team
```

**Expected**: ✅ Alerts configured

**Status**: ⬜ Not Run

---

## SEC-013: Server-Side Request Forgery (SSRF)

**Priority**: P1 (High)
**Type**: Security Test - SSRF
**OWASP**: A10:2021 - Server-Side Request Forgery

### Test Objective
Verify system is protected against SSRF attacks.

#### SEC-013-001: SSRF via External URL
```gherkin
Given an attacker provides malicious URL
When the API fetches external resource
Then only whitelisted domains are allowed
```

**Test Payloads**:
```json
{
  "webhook_url": "http://169.254.169.254/latest/meta-data/"
}

{
  "webhook_url": "http://localhost:5432/postgres"
}

{
  "webhook_url": "file:///etc/passwd"
}
```

**Expected**: ✅ All malicious URLs rejected

**Status**: ⬜ Not Run

---

## Security Test Execution Summary

### Test Execution Checklist

#### Pre-Test
- [ ] Security test environment isolated
- [ ] Test tools installed (ZAP, Burp, SQLMap)
- [ ] Test users created
- [ ] Security team notified
- [ ] Backups taken

#### During Test
- [ ] Monitor security logs in real-time
- [ ] Document all findings
- [ ] Capture screenshots/evidence
- [ ] Test in isolated environment only

#### Post-Test
- [ ] Generate security report
- [ ] File security issues in bug tracker
- [ ] Remediate critical/high vulnerabilities
- [ ] Re-test fixes
- [ ] Update security documentation

### Test Results Template

```markdown
# Security Test Report - DICT LBPay

**Test Date**: 2025-10-25
**Tester**: Security QA Team
**Environment**: security-test

## Executive Summary
✅ System passed security testing with 0 critical vulnerabilities

## Vulnerabilities Found

### Critical (P0)
None

### High (P1)
None

### Medium (P2)
- [MED-001] Missing rate limit on /health endpoint (low risk)

### Low (P3)
- [LOW-001] Server version exposed in response headers

## OWASP Top 10 Compliance
✅ All OWASP Top 10 (2021) vulnerabilities tested and mitigated

## Recommendations
1. Add rate limiting to /health endpoint
2. Remove server version from headers
3. Enable additional WAF rules
```

---

## Automated Security Testing

### CI/CD Integration

```yaml
# .github/workflows/security-scan.yml
name: Security Scan

on:
  push:
    branches: [main, develop]
  pull_request:
  schedule:
    - cron: '0 2 * * 1'  # Weekly Monday 2am

jobs:
  sast:
    name: Static Analysis (SAST)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Run Semgrep
        uses: returntocorp/semgrep-action@v1
        with:
          config: >-
            p/security-audit
            p/owasp-top-ten

      - name: Run Snyk
        uses: snyk/actions/node@master
        env:
          SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
        with:
          args: --severity-threshold=high

  dast:
    name: Dynamic Analysis (DAST)
    runs-on: ubuntu-latest
    steps:
      - name: ZAP Scan
        uses: zaproxy/action-baseline@v0.7.0
        with:
          target: 'https://dict-api-sec.lbpay.com.br'
          rules_file_name: '.zap/rules.tsv'
          cmd_options: '-a'

  dependency-check:
    name: Dependency Vulnerability Scan
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Run npm audit
        run: npm audit --audit-level=high

      - name: OWASP Dependency Check
        uses: dependency-check/Dependency-Check_Action@main
        with:
          project: 'dict-lbpay'
          path: '.'
          format: 'HTML'
```

---

## Glossary

- **JWT**: JSON Web Token
- **mTLS**: Mutual TLS (both client and server authenticate)
- **RBAC**: Role-Based Access Control
- **XSS**: Cross-Site Scripting
- **CSRF**: Cross-Site Request Forgery
- **SSRF**: Server-Side Request Forgery
- **OWASP**: Open Web Application Security Project
- **WAF**: Web Application Firewall
- **HSTS**: HTTP Strict Transport Security
- **CSP**: Content Security Policy
- **SAST**: Static Application Security Testing
- **DAST**: Dynamic Application Security Testing

---

**Última Revisão**: 2025-10-25
**Aprovado por**: Security Team Lead
**Próxima Revisão**: Quarterly security assessment
