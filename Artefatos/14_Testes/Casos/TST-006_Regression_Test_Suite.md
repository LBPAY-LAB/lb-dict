# TST-006: Regression Test Suite

**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: QA Team
**Status**: ✅ Completo

---

## Sumário Executivo

Este documento especifica a **Regression Test Suite** (Suite de Testes de Regressão) para o sistema DICT LBPay, garantindo que novas mudanças não quebrem funcionalidades existentes.

**Objetivo**: Executar suite automatizada de testes de regressão em cada Pull Request e release, cobrindo todos os caminhos críticos do sistema.

**Cobertura**:
- Automated test suite covering all critical paths
- Test data management and test fixtures
- Test environments (dev, staging, production-like)
- CI integration (GitHub Actions, run on every PR)
- Test coverage target: 80% code coverage
- Regression detection and reporting

**Ferramentas**:
- Jest: Unit & Integration Tests
- Supertest: API Integration Tests
- k6: Performance Regression Tests
- Istanbul/nyc: Code Coverage
- GitHub Actions: CI/CD Pipeline
- Allure: Test Reporting

**Referências**:
- [TST-001: Test Cases CreateEntry](./TST-001_Test_Cases_CreateEntry.md)
- [TST-004: Performance Tests](./TST-004_Performance_Tests.md)
- [TST-005: Security Tests](./TST-005_Security_Tests.md)

---

## Test Strategy

### Regression Test Pyramid

```
                    E2E Tests
                 (10% - Critical Paths)
                   /          \
              API Integration Tests
           (30% - Business Flows)
          /                      \
     Unit Tests
(60% - Functions, Classes, Modules)
```

### Test Suite Composition

```yaml
Test Suite Distribution:
  Unit Tests: 60%
    - Total: ~1200 tests
    - Execution Time: 2-3 minutes
    - Coverage: Functions, classes, utilities

  Integration Tests: 30%
    - Total: ~600 tests
    - Execution Time: 5-7 minutes
    - Coverage: API endpoints, database, external services

  E2E Tests: 10%
    - Total: ~200 tests
    - Execution Time: 10-15 minutes
    - Coverage: Critical user flows

Total Tests: ~2000 tests
Total Execution Time: 17-25 minutes
```

### Coverage Targets

```yaml
Code Coverage Requirements:
  Overall: 80% minimum

  By Component:
    API Layer: 90% (critical)
    Business Logic: 85% (critical)
    Data Access Layer: 80%
    Utilities: 75%
    Configuration: 60%

  By Metric:
    Line Coverage: 80%
    Branch Coverage: 75%
    Function Coverage: 85%
    Statement Coverage: 80%
```

---

## Test Environment Setup

### Test Environments

#### Development Environment
```yaml
Environment: dev
Purpose: Developer local testing
API: http://localhost:3000
Database: PostgreSQL (Docker)
External Services: All mocked
CI: Not required
```

#### CI Environment
```yaml
Environment: ci
Purpose: Automated testing in GitHub Actions
API: Started in CI pipeline
Database: PostgreSQL (GitHub Actions service)
External Services: Mocked
CI: Required for all PRs
Coverage: Collected and reported
```

#### Staging Environment
```yaml
Environment: staging
Purpose: Pre-production regression testing
API: https://dict-api-staging.lbpay.com.br
Database: PostgreSQL (staging cluster)
External Services: Staging/mock mix
CI: Nightly regression suite
Schedule: Daily at 2am
```

### Test Dependencies Installation

```bash
# Install test dependencies
npm install --save-dev \
  jest \
  @types/jest \
  supertest \
  @types/supertest \
  ts-jest \
  jest-junit \
  @jest/globals \
  allure-commandline \
  allure-jest

# Install coverage tools
npm install --save-dev \
  nyc \
  @istanbuljs/nyc-config-typescript

# Install test utilities
npm install --save-dev \
  faker \
  nock \
  @testcontainers/postgresql
```

### Jest Configuration

```javascript
// jest.config.js
module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'node',
  roots: ['<rootDir>/src', '<rootDir>/tests'],

  // Test matching
  testMatch: [
    '**/__tests__/**/*.test.ts',
    '**/*.spec.ts',
  ],

  // Coverage configuration
  collectCoverage: true,
  collectCoverageFrom: [
    'src/**/*.{ts,tsx}',
    '!src/**/*.d.ts',
    '!src/**/index.ts',
    '!src/**/*.interface.ts',
    '!src/**/*.type.ts',
  ],
  coverageThresholds: {
    global: {
      branches: 75,
      functions: 85,
      lines: 80,
      statements: 80,
    },
    './src/api/**/*.ts': {
      branches: 80,
      functions: 90,
      lines: 90,
      statements: 90,
    },
    './src/services/**/*.ts': {
      branches: 80,
      functions: 85,
      lines: 85,
      statements: 85,
    },
  },

  // Reporters
  reporters: [
    'default',
    'jest-junit',
    ['jest-html-reporter', {
      pageTitle: 'DICT LBPay Test Report',
      outputPath: 'coverage/test-report.html',
    }],
  ],

  // Test timeout
  testTimeout: 10000,

  // Setup files
  setupFilesAfterEnv: ['<rootDir>/tests/setup.ts'],

  // Module paths
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/src/$1',
    '^@tests/(.*)$': '<rootDir>/tests/$1',
  },
};
```

---

## Unit Tests (60% of Suite)

### Unit Test Structure

```
tests/
  unit/
    api/
      controllers/
        entry.controller.test.ts
        claim.controller.test.ts
      middleware/
        auth.middleware.test.ts
        validation.middleware.test.ts
    services/
      entry.service.test.ts
      claim.service.test.ts
      bacen-bridge.service.test.ts
    repositories/
      entry.repository.test.ts
      claim.repository.test.ts
    utils/
      cpf-validator.test.ts
      jwt-helper.test.ts
```

### Example Unit Test

```typescript
// tests/unit/services/entry.service.test.ts
import { EntryService } from '@/services/entry.service';
import { EntryRepository } from '@/repositories/entry.repository';
import { BacenBridgeService } from '@/services/bacen-bridge.service';
import { jest } from '@jest/globals';

describe('EntryService', () => {
  let entryService: EntryService;
  let entryRepository: jest.Mocked<EntryRepository>;
  let bacenBridge: jest.Mocked<BacenBridgeService>;

  beforeEach(() => {
    // Create mocks
    entryRepository = {
      findByKey: jest.fn(),
      create: jest.fn(),
      update: jest.fn(),
    } as any;

    bacenBridge = {
      createEntry: jest.fn(),
    } as any;

    entryService = new EntryService(entryRepository, bacenBridge);
  });

  describe('createEntry', () => {
    it('should create entry successfully when key does not exist', async () => {
      // Arrange
      const input = {
        key_type: 'CPF',
        key_value: '12345678900',
        account: {
          ispb: '12345678',
          account_number: '123456',
          branch: '0001',
          account_type: 'CACC',
        },
      };

      entryRepository.findByKey.mockResolvedValue(null);
      entryRepository.create.mockResolvedValue({
        entry_id: 'uuid-123',
        ...input,
        status: 'PENDING',
        created_at: new Date(),
      });

      // Act
      const result = await entryService.createEntry(input);

      // Assert
      expect(entryRepository.findByKey).toHaveBeenCalledWith('CPF', '12345678900');
      expect(entryRepository.create).toHaveBeenCalledWith(input);
      expect(result.status).toBe('PENDING');
      expect(result.entry_id).toBe('uuid-123');
    });

    it('should throw error when key already exists', async () => {
      // Arrange
      const input = {
        key_type: 'CPF',
        key_value: '12345678900',
        account: { /* ... */ },
      };

      entryRepository.findByKey.mockResolvedValue({
        entry_id: 'existing-uuid',
        status: 'ACTIVE',
      } as any);

      // Act & Assert
      await expect(entryService.createEntry(input))
        .rejects
        .toThrow('KEY_ALREADY_EXISTS');

      expect(entryRepository.create).not.toHaveBeenCalled();
    });

    it('should validate CPF format before creating entry', async () => {
      // Arrange
      const input = {
        key_type: 'CPF',
        key_value: '11111111111', // Invalid CPF
        account: { /* ... */ },
      };

      // Act & Assert
      await expect(entryService.createEntry(input))
        .rejects
        .toThrow('INVALID_KEY_FORMAT');
    });
  });

  describe('getEntry', () => {
    it('should return entry when found', async () => {
      // Arrange
      const mockEntry = {
        entry_id: 'uuid-123',
        key_type: 'CPF',
        key_value: '12345678900',
        status: 'ACTIVE',
      };

      entryRepository.findByKey.mockResolvedValue(mockEntry as any);

      // Act
      const result = await entryService.getEntry('CPF', '12345678900');

      // Assert
      expect(result).toEqual(mockEntry);
      expect(entryRepository.findByKey).toHaveBeenCalledWith('CPF', '12345678900');
    });

    it('should return null when entry not found', async () => {
      // Arrange
      entryRepository.findByKey.mockResolvedValue(null);

      // Act
      const result = await entryService.getEntry('CPF', '99999999999');

      // Assert
      expect(result).toBeNull();
    });
  });
});
```

### Unit Test Coverage

```yaml
Unit Test Categories:
  - Service Layer: 400 tests
    - Business logic validation
    - Error handling
    - Data transformations

  - Repository Layer: 300 tests
    - Database queries
    - Transaction handling
    - Error handling

  - Validators: 200 tests
    - CPF validation
    - Email validation
    - Phone validation (E.164)
    - CNPJ validation

  - Utilities: 200 tests
    - JWT helpers
    - Date formatters
    - String utilities

  - Middleware: 100 tests
    - Authentication
    - Authorization
    - Error handling
    - Request validation
```

---

## Integration Tests (30% of Suite)

### Integration Test Structure

```
tests/
  integration/
    api/
      entries/
        create-entry.test.ts
        get-entry.test.ts
        delete-entry.test.ts
      claims/
        create-claim.test.ts
        approve-claim.test.ts
    database/
      transactions.test.ts
      migrations.test.ts
    external/
      bacen-bridge.test.ts
      temporal.test.ts
```

### Example Integration Test

```typescript
// tests/integration/api/entries/create-entry.test.ts
import request from 'supertest';
import { app } from '@/app';
import { Database } from '@/database';
import { generateTestToken } from '@tests/helpers/auth.helper';
import { setupTestDatabase, cleanupTestDatabase } from '@tests/helpers/database.helper';

describe('POST /api/v1/keys - CreateEntry Integration', () => {
  let authToken: string;
  let db: Database;

  beforeAll(async () => {
    db = await setupTestDatabase();
    authToken = generateTestToken({
      user_id: 'test-user-123',
      scopes: ['dict:read', 'dict:write'],
    });
  });

  afterAll(async () => {
    await cleanupTestDatabase(db);
  });

  afterEach(async () => {
    // Clean up test data after each test
    await db.query('DELETE FROM dict.entries WHERE key_value LIKE \'test-%\'');
  });

  describe('Happy Path', () => {
    it('should create CPF entry successfully', async () => {
      // Arrange
      const payload = {
        key_type: 'CPF',
        key_value: '12345678900',
        account: {
          ispb: '12345678',
          account_number: '123456',
          branch: '0001',
          account_type: 'CACC',
        },
      };

      // Act
      const response = await request(app)
        .post('/api/v1/keys')
        .set('Authorization', `Bearer ${authToken}`)
        .send(payload);

      // Assert
      expect(response.status).toBe(201);
      expect(response.body).toMatchObject({
        entry_id: expect.any(String),
        key_type: 'CPF',
        key_value: '12345678900',
        status: 'PENDING',
        created_at: expect.any(String),
      });

      // Verify database
      const dbEntry = await db.query(
        'SELECT * FROM dict.entries WHERE key_value = $1',
        ['12345678900']
      );
      expect(dbEntry.rows).toHaveLength(1);
      expect(dbEntry.rows[0].status).toBe('PENDING');
    });

    it('should create EMAIL entry successfully', async () => {
      const response = await request(app)
        .post('/api/v1/keys')
        .set('Authorization', `Bearer ${authToken}`)
        .send({
          key_type: 'EMAIL',
          key_value: 'test@example.com',
          account: {
            ispb: '12345678',
            account_number: '123456',
            branch: '0001',
            account_type: 'CACC',
          },
        });

      expect(response.status).toBe(201);
      expect(response.body.key_type).toBe('EMAIL');
    });
  });

  describe('Error Handling', () => {
    it('should return 400 for invalid CPF', async () => {
      const response = await request(app)
        .post('/api/v1/keys')
        .set('Authorization', `Bearer ${authToken}`)
        .send({
          key_type: 'CPF',
          key_value: '11111111111', // Invalid
          account: {
            ispb: '12345678',
            account_number: '123456',
            branch: '0001',
            account_type: 'CACC',
          },
        });

      expect(response.status).toBe(400);
      expect(response.body.error).toBe('INVALID_KEY_FORMAT');
    });

    it('should return 409 for duplicate key', async () => {
      // Create first entry
      await request(app)
        .post('/api/v1/keys')
        .set('Authorization', `Bearer ${authToken}`)
        .send({
          key_type: 'CPF',
          key_value: '98765432100',
          account: { /* ... */ },
        });

      // Attempt duplicate
      const response = await request(app)
        .post('/api/v1/keys')
        .set('Authorization', `Bearer ${authToken}`)
        .send({
          key_type: 'CPF',
          key_value: '98765432100',
          account: { /* ... */ },
        });

      expect(response.status).toBe(409);
      expect(response.body.error).toBe('KEY_ALREADY_EXISTS');
    });

    it('should return 401 for missing auth token', async () => {
      const response = await request(app)
        .post('/api/v1/keys')
        .send({
          key_type: 'CPF',
          key_value: '12345678900',
          account: { /* ... */ },
        });

      expect(response.status).toBe(401);
    });

    it('should return 403 for insufficient scopes', async () => {
      const readOnlyToken = generateTestToken({
        user_id: 'test-user-456',
        scopes: ['dict:read'], // No write scope
      });

      const response = await request(app)
        .post('/api/v1/keys')
        .set('Authorization', `Bearer ${readOnlyToken}`)
        .send({
          key_type: 'CPF',
          key_value: '12345678900',
          account: { /* ... */ },
        });

      expect(response.status).toBe(403);
    });
  });

  describe('Database Transactions', () => {
    it('should rollback on error', async () => {
      // Mock database to throw error
      jest.spyOn(db, 'query').mockRejectedValueOnce(new Error('Database error'));

      const response = await request(app)
        .post('/api/v1/keys')
        .set('Authorization', `Bearer ${authToken}`)
        .send({
          key_type: 'CPF',
          key_value: 'test-rollback-cpf',
          account: { /* ... */ },
        });

      expect(response.status).toBe(500);

      // Verify rollback - entry should NOT exist
      const dbEntry = await db.query(
        'SELECT * FROM dict.entries WHERE key_value = $1',
        ['test-rollback-cpf']
      );
      expect(dbEntry.rows).toHaveLength(0);
    });
  });
});
```

### Integration Test Coverage

```yaml
Integration Test Categories:
  - API Endpoints: 300 tests
    - All CRUD operations
    - Error responses
    - Edge cases

  - Database Integration: 150 tests
    - Transaction handling
    - Constraint violations
    - Concurrent access

  - External Service Mocks: 100 tests
    - Bacen Bridge mocked
    - Temporal mocked
    - Auth service mocked

  - Authentication/Authorization: 50 tests
    - JWT validation
    - Scope checks
    - RBAC
```

---

## E2E Tests (10% of Suite)

### E2E Test Structure

```
tests/
  e2e/
    flows/
      create-entry-flow.test.ts
      claim-portability-flow.test.ts
      delete-entry-flow.test.ts
    scenarios/
      happy-path.test.ts
      error-recovery.test.ts
```

### Example E2E Test

```typescript
// tests/e2e/flows/create-entry-flow.test.ts
import request from 'supertest';
import { app } from '@/app';

describe('E2E: CreateEntry Complete Flow', () => {
  it('should complete full CreateEntry workflow from API to Bacen sync', async () => {
    // Step 1: Authenticate
    const authResponse = await request(app)
      .post('/auth/login')
      .send({
        email: 'e2e.user@lbpay.com.br',
        password: 'E2E@1234',
      });

    expect(authResponse.status).toBe(200);
    const { access_token } = authResponse.body;

    // Step 2: Create CPF entry
    const createResponse = await request(app)
      .post('/api/v1/keys')
      .set('Authorization', `Bearer ${access_token}`)
      .send({
        key_type: 'CPF',
        key_value: '12345678900',
        account: {
          ispb: '12345678',
          account_number: '123456',
          branch: '0001',
          account_type: 'CACC',
        },
      });

    expect(createResponse.status).toBe(201);
    expect(createResponse.body.status).toBe('PENDING');
    const { entry_id } = createResponse.body;

    // Step 3: Poll for ACTIVE status (workflow completion)
    let statusResponse;
    let attempts = 0;
    const maxAttempts = 10;

    while (attempts < maxAttempts) {
      await new Promise(resolve => setTimeout(resolve, 200)); // 200ms delay

      statusResponse = await request(app)
        .get('/api/v1/keys/CPF/12345678900')
        .set('Authorization', `Bearer ${access_token}`);

      if (statusResponse.body.status === 'ACTIVE') {
        break;
      }

      attempts++;
    }

    // Step 4: Verify final state
    expect(statusResponse.status).toBe(200);
    expect(statusResponse.body).toMatchObject({
      entry_id,
      key_type: 'CPF',
      key_value: '12345678900',
      status: 'ACTIVE',
      external_id: expect.any(String), // Bacen ID
    });

    // Step 5: Verify audit log
    const auditResponse = await request(app)
      .get(`/api/v1/audit/entries/${entry_id}`)
      .set('Authorization', `Bearer ${access_token}`);

    expect(auditResponse.status).toBe(200);
    expect(auditResponse.body).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          action: 'CREATE',
          entity_id: entry_id,
        }),
      ])
    );

    // Step 6: Cleanup - Delete entry
    const deleteResponse = await request(app)
      .delete(`/api/v1/keys/CPF/12345678900`)
      .set('Authorization', `Bearer ${access_token}`);

    expect(deleteResponse.status).toBe(204);
  });
});
```

---

## Test Data Management

### Test Fixtures

```typescript
// tests/fixtures/entries.fixture.ts
export const validCPFEntry = {
  key_type: 'CPF',
  key_value: '12345678900',
  account: {
    ispb: '12345678',
    account_number: '123456',
    branch: '0001',
    account_type: 'CACC',
  },
};

export const validEmailEntry = {
  key_type: 'EMAIL',
  key_value: 'test@example.com',
  account: {
    ispb: '12345678',
    account_number: '123456',
    branch: '0001',
    account_type: 'CACC',
  },
};

// tests/fixtures/users.fixture.ts
export const testUsers = {
  admin: {
    user_id: 'admin-uuid',
    email: 'admin@lbpay.com.br',
    scopes: ['dict:read', 'dict:write', 'dict:admin'],
  },
  regular: {
    user_id: 'user-uuid',
    email: 'user@lbpay.com.br',
    scopes: ['dict:read', 'dict:write'],
  },
  readonly: {
    user_id: 'readonly-uuid',
    email: 'readonly@lbpay.com.br',
    scopes: ['dict:read'],
  },
};
```

### Test Data Generators

```typescript
// tests/helpers/data-generator.ts
import { faker } from '@faker-js/faker';

export class TestDataGenerator {
  static generateCPF(): string {
    // Generate valid CPF for testing
    const digits = Array.from({ length: 9 }, () => Math.floor(Math.random() * 10));
    // Calculate check digits...
    return digits.join('');
  }

  static generateEmail(): string {
    return faker.internet.email().toLowerCase();
  }

  static generatePhone(): string {
    return `+55${faker.phone.number('11#########')}`;
  }

  static generateCNPJ(): string {
    // Generate valid CNPJ for testing
    return '12345678000190';
  }

  static generateEntry(type: 'CPF' | 'EMAIL' | 'PHONE' | 'CNPJ' | 'EVP') {
    const generators = {
      CPF: this.generateCPF,
      EMAIL: this.generateEmail,
      PHONE: this.generatePhone,
      CNPJ: this.generateCNPJ,
      EVP: () => faker.string.uuid(),
    };

    return {
      key_type: type,
      key_value: generators[type](),
      account: {
        ispb: '12345678',
        account_number: faker.finance.accountNumber(6),
        branch: faker.finance.accountNumber(4),
        account_type: 'CACC',
      },
    };
  }
}
```

### Database Seeding

```typescript
// tests/helpers/database.helper.ts
import { Pool } from 'pg';

export async function setupTestDatabase(): Promise<Pool> {
  const pool = new Pool({
    host: process.env.TEST_DB_HOST || 'localhost',
    port: parseInt(process.env.TEST_DB_PORT || '5432'),
    database: process.env.TEST_DB_NAME || 'dict_test',
    user: process.env.TEST_DB_USER || 'postgres',
    password: process.env.TEST_DB_PASSWORD || 'postgres',
  });

  // Run migrations
  await runMigrations(pool);

  // Seed test data
  await seedTestData(pool);

  return pool;
}

async function seedTestData(pool: Pool): Promise<void> {
  // Insert test accounts
  await pool.query(`
    INSERT INTO dict.accounts (ispb, account_number, branch, account_type, status)
    VALUES
      ('12345678', '123456', '0001', 'CACC', 'ACTIVE'),
      ('12345678', '654321', '0002', 'SVGS', 'ACTIVE'),
      ('12345678', '999999', '0999', 'CACC', 'INACTIVE')
    ON CONFLICT DO NOTHING;
  `);
}

export async function cleanupTestDatabase(pool: Pool): Promise<void> {
  await pool.query('DELETE FROM dict.entries');
  await pool.query('DELETE FROM dict.claims');
  await pool.query('DELETE FROM dict.audit_logs');
  await pool.end();
}
```

---

## CI/CD Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/regression-tests.yml
name: Regression Test Suite

on:
  pull_request:
    branches: [main, develop]
  push:
    branches: [main, develop]
  schedule:
    - cron: '0 2 * * *'  # Nightly at 2am

jobs:
  unit-tests:
    name: Unit Tests
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '20'
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Run unit tests
        run: npm run test:unit

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage/coverage-final.json
          flags: unit

  integration-tests:
    name: Integration Tests
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:16.4
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: dict_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379

    steps:
      - uses: actions/checkout@v3

      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '20'
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Run database migrations
        env:
          DATABASE_URL: postgresql://postgres:postgres@localhost:5432/dict_test
        run: npm run migrate

      - name: Run integration tests
        env:
          DATABASE_URL: postgresql://postgres:postgres@localhost:5432/dict_test
          REDIS_URL: redis://localhost:6379
        run: npm run test:integration

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage/coverage-final.json
          flags: integration

  e2e-tests:
    name: E2E Tests
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:16.4
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: dict_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

    steps:
      - uses: actions/checkout@v3

      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '20'
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Start application
        env:
          DATABASE_URL: postgresql://postgres:postgres@localhost:5432/dict_test
          NODE_ENV: test
        run: |
          npm run build
          npm start &
          sleep 10  # Wait for app to start

      - name: Run E2E tests
        run: npm run test:e2e

      - name: Upload E2E artifacts
        if: failure()
        uses: actions/upload-artifact@v3
        with:
          name: e2e-screenshots
          path: tests/e2e/screenshots/

  coverage-check:
    name: Coverage Check
    runs-on: ubuntu-latest
    needs: [unit-tests, integration-tests, e2e-tests]

    steps:
      - uses: actions/checkout@v3

      - name: Download coverage reports
        uses: actions/download-artifact@v3

      - name: Merge coverage reports
        run: npx nyc merge coverage coverage/merged-coverage.json

      - name: Check coverage thresholds
        run: |
          npx nyc check-coverage \
            --lines 80 \
            --functions 85 \
            --branches 75 \
            --statements 80

      - name: Generate coverage report
        run: npx nyc report --reporter=html --reporter=text

      - name: Comment PR with coverage
        if: github.event_name == 'pull_request'
        uses: romeovs/lcov-reporter-action@v0.3.1
        with:
          lcov-file: ./coverage/lcov.info
          github-token: ${{ secrets.GITHUB_TOKEN }}

  regression-report:
    name: Generate Regression Report
    runs-on: ubuntu-latest
    needs: [unit-tests, integration-tests, e2e-tests]
    if: always()

    steps:
      - uses: actions/checkout@v3

      - name: Download test results
        uses: actions/download-artifact@v3

      - name: Generate Allure report
        run: |
          npm install -g allure-commandline
          allure generate allure-results --clean -o allure-report

      - name: Upload Allure report
        uses: actions/upload-artifact@v3
        with:
          name: allure-report
          path: allure-report/

      - name: Comment PR with test summary
        if: github.event_name == 'pull_request'
        uses: actions/github-script@v6
        with:
          script: |
            const fs = require('fs');
            const summary = fs.readFileSync('test-summary.md', 'utf8');
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: summary
            });
```

### NPM Scripts

```json
{
  "scripts": {
    "test": "jest --coverage",
    "test:unit": "jest --testPathPattern=tests/unit --coverage",
    "test:integration": "jest --testPathPattern=tests/integration --coverage --runInBand",
    "test:e2e": "jest --testPathPattern=tests/e2e --runInBand",
    "test:watch": "jest --watch",
    "test:ci": "jest --ci --coverage --maxWorkers=2",
    "coverage": "jest --coverage && open coverage/lcov-report/index.html",
    "coverage:check": "nyc check-coverage --lines 80 --functions 85 --branches 75"
  }
}
```

---

## Regression Detection

### Performance Regression Detection

```typescript
// tests/helpers/performance-baseline.ts
import fs from 'fs';

interface PerformanceBaseline {
  test: string;
  p50: number;
  p95: number;
  p99: number;
}

export class PerformanceRegressionDetector {
  private baseline: PerformanceBaseline[];

  constructor(baselineFile: string) {
    this.baseline = JSON.parse(fs.readFileSync(baselineFile, 'utf8'));
  }

  check(testName: string, p50: number, p95: number, p99: number): void {
    const baseline = this.baseline.find(b => b.test === testName);

    if (!baseline) {
      console.warn(`No baseline found for test: ${testName}`);
      return;
    }

    const threshold = 0.2; // 20% regression threshold

    if (p95 > baseline.p95 * (1 + threshold)) {
      throw new Error(
        `Performance regression detected in ${testName}:\n` +
        `  Baseline p95: ${baseline.p95}ms\n` +
        `  Current p95: ${p95}ms\n` +
        `  Increase: ${((p95 / baseline.p95 - 1) * 100).toFixed(1)}%`
      );
    }
  }
}
```

### Functional Regression Detection

```typescript
// tests/helpers/snapshot-testing.ts
describe('API Response Snapshot Tests', () => {
  it('should match CreateEntry response structure', async () => {
    const response = await request(app)
      .post('/api/v1/keys')
      .set('Authorization', `Bearer ${token}`)
      .send(validCPFEntry);

    expect(response.body).toMatchSnapshot({
      entry_id: expect.any(String),
      created_at: expect.any(String),
    });
  });
});
```

---

## Test Reporting

### Allure Report Configuration

```javascript
// jest.config.js
module.exports = {
  reporters: [
    'default',
    ['jest-allure', {
      resultsDir: 'allure-results',
      stepName: 'StepName',
    }],
  ],
};
```

### Test Summary Template

```markdown
# Regression Test Summary

**Date**: 2025-10-25
**Commit**: abc123
**PR**: #456

## Summary
✅ All tests passed

## Test Results

| Category | Total | Passed | Failed | Skipped | Duration |
|----------|-------|--------|--------|---------|----------|
| Unit | 1200 | 1200 | 0 | 0 | 2m 34s |
| Integration | 600 | 600 | 0 | 0 | 6m 12s |
| E2E | 200 | 200 | 0 | 0 | 12m 45s |
| **Total** | **2000** | **2000** | **0** | **0** | **21m 31s** |

## Code Coverage

| Metric | Coverage | Threshold | Status |
|--------|----------|-----------|--------|
| Lines | 82.5% | 80% | ✅ Pass |
| Functions | 87.3% | 85% | ✅ Pass |
| Branches | 76.8% | 75% | ✅ Pass |
| Statements | 81.9% | 80% | ✅ Pass |

## Performance Metrics

| Endpoint | p50 | p95 | p99 | Status |
|----------|-----|-----|-----|--------|
| POST /api/v1/keys | 98ms | 287ms | 512ms | ✅ |
| GET /api/v1/keys/{type}/{value} | 23ms | 67ms | 123ms | ✅ |

## Recommendations
- All tests passing
- Coverage targets met
- No performance regressions detected
- Ready to merge
```

---

## Continuous Improvement

### Test Maintenance Strategy

```yaml
Weekly:
  - Review flaky tests (retry > 2 times)
  - Update test fixtures
  - Remove obsolete tests

Monthly:
  - Review code coverage gaps
  - Add tests for new features
  - Refactor slow tests
  - Update baselines

Quarterly:
  - Full regression suite audit
  - Performance baseline update
  - Test tool upgrades
  - Test documentation review
```

### Flaky Test Detection

```typescript
// .github/workflows/flaky-test-detection.yml
# Run tests 10 times to detect flaky tests
name: Flaky Test Detection

on:
  schedule:
    - cron: '0 3 * * 0'  # Weekly Sunday 3am

jobs:
  detect-flaky:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Run tests 10 times
        run: |
          for i in {1..10}; do
            echo "Run $i"
            npm test -- --json --outputFile=results-$i.json || true
          done

      - name: Analyze results
        run: node scripts/detect-flaky-tests.js results-*.json
```

---

## Glossary

- **Regression**: Bug reintroduced after being fixed
- **Flaky Test**: Test that passes/fails non-deterministically
- **Code Coverage**: % of code executed by tests
- **Test Fixture**: Reusable test data
- **Snapshot Testing**: Compare current output to saved snapshot
- **Mocking**: Simulating external dependencies
- **Stubbing**: Replacing function with test double

---

**Última Revisão**: 2025-10-25
**Aprovado por**: QA Lead
**Próxima Execução**: Every PR + Nightly
