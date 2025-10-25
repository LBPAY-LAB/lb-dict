# Casos de Teste

**Propósito**: Casos de teste detalhados para validação funcional do sistema DICT

## 📋 Conteúdo

Esta pasta armazenará:

- **Test Cases**: Casos de teste funcionais (happy path + edge cases)
- **Test Scenarios**: Cenários de teste end-to-end
- **Test Data**: Dados de teste (fixtures, mocks)
- **Test Results**: Relatórios de execução de testes

## 📁 Estrutura Esperada

```
Casos/
├── Functional/
│   ├── TC_Entry/
│   │   ├── TC-001_CreateEntry_CPF_Success.md
│   │   ├── TC-002_CreateEntry_Email_Success.md
│   │   ├── TC-003_CreateEntry_DuplicateKey_Error.md
│   │   └── TC-004_GetEntry_NotFound.md
│   ├── TC_Claim/
│   │   ├── TC-010_CreateClaim_Success.md
│   │   ├── TC-011_AcceptClaim_Success.md
│   │   ├── TC-012_RejectClaim_Success.md
│   │   └── TC-013_ClaimExpiration_30Days.md
│   └── TC_Portability/
│       ├── TC-020_ConfirmPortability_Success.md
│       └── TC-021_CancelPortability_Success.md
├── Integration/
│   ├── TC_E2E_CreateEntry.md
│   ├── TC_E2E_ClaimWorkflow.md
│   └── TC_E2E_Portability.md
├── Performance/
│   ├── TC_PERF_CreateEntry_Load.md
│   └── TC_PERF_ClaimWorkflow_Concurrency.md
├── Security/
│   ├── TC_SEC_mTLS_Authentication.md
│   ├── TC_SEC_XML_Signature_Validation.md
│   └── TC_SEC_SQL_Injection_Prevention.md
└── Test_Data/
    ├── valid_cpf.json
    ├── valid_cnpj.json
    └── test_accounts.json
```

## 🎯 Template de Caso de Teste

```markdown
# TC-001: CreateEntry - CPF Success (Happy Path)

## Informações do Teste

| Campo | Valor |
|-------|-------|
| **ID** | TC-001 |
| **Categoria** | Functional - Entry Operations |
| **Prioridade** | Alta |
| **Tipo** | Positivo (Happy Path) |
| **Automação** | Sim |

## User Story Relacionada
- US-005: Cadastrar Chave CPF

## Pré-condições

1. Sistema DICT está disponível
2. Usuário autenticado no app (token válido)
3. CPF "12345678900" **NÃO** existe no DICT
4. Conta "00000000-0001-12345-6" está ativa

## Dados de Teste

```json
{
  "key": {
    "keyType": "CPF",
    "keyValue": "12345678900"
  },
  "account": {
    "ispb": "00000000",
    "branchCode": "0001",
    "accountNumber": "12345",
    "accountCheckDigit": "6",
    "accountType": "CHECKING",
    "accountHolderName": "João Silva",
    "accountHolderDocument": "12345678900",
    "documentType": "CPF"
  }
}
```

## Passos de Execução

### Step 1: Enviar request CreateEntry
**Ação**: POST /api/v1/entries com dados acima
**Esperado**: Request aceito

### Step 2: Verificar response
**Esperado**:
- HTTP Status: 201 Created
- Response body contém:
  - `entry_id` (UUID válido)
  - `status`: "ACTIVE"
  - `created_at` (timestamp ISO 8601)

### Step 3: Validar criação no banco de dados
**Query**:
```sql
SELECT * FROM dict.entries WHERE key_value = '12345678900';
```
**Esperado**:
- 1 registro encontrado
- `status` = 'ACTIVE'
- `created_at` <= now()

### Step 4: Validar evento Pulsar
**Topic**: `dict.entries.created`
**Esperado**:
- Evento publicado com `entry_id` correto
- `event_type` = "ENTRY_CREATED"

### Step 5: Validar criação no Bacen (via Bridge)
**Verificar logs do Bridge**:
- SOAP request enviado ao Bacen
- mTLS handshake bem-sucedido
- XML assinado digitalmente (ICP-Brasil)
- Response do Bacen: `external_id` recebido

## Resultado Esperado

✅ Chave CPF criada com sucesso no DICT local e no Bacen
✅ Entry persistida no PostgreSQL com status ACTIVE
✅ Evento EntryCreated publicado no Pulsar
✅ Response 201 Created retornado ao cliente

## Pós-condições

- Entry "12345678900" existe no sistema
- Conta "00000000-0001-12345-6" vinculada à entry
- Usuário pode fazer transações PIX usando CPF

## Casos de Falha Relacionados

- TC-003: CreateEntry com CPF duplicado → 409 Conflict
- TC-005: CreateEntry com CPF inválido → 400 Bad Request
- TC-006: CreateEntry sem autenticação → 401 Unauthorized

## Automação (Pseudocódigo)

```javascript
// Exemplo com Jest + Supertest
describe('TC-001: CreateEntry - CPF Success', () => {
  it('should create entry with CPF successfully', async () => {
    // Arrange
    const payload = {
      key: { keyType: 'CPF', keyValue: '12345678900' },
      account: { /* ... */ }
    };

    // Act
    const response = await request(app)
      .post('/api/v1/entries')
      .set('Authorization', `Bearer ${validToken}`)
      .send(payload);

    // Assert
    expect(response.status).toBe(201);
    expect(response.body).toHaveProperty('entry_id');
    expect(response.body.status).toBe('ACTIVE');

    // Verify DB
    const entry = await db.query(
      'SELECT * FROM dict.entries WHERE key_value = $1',
      ['12345678900']
    );
    expect(entry.rows).toHaveLength(1);
    expect(entry.rows[0].status).toBe('ACTIVE');
  });
});
```

## Executado em

| Data | Versão | Resultado | Executor |
|------|--------|-----------|----------|
| 2025-10-25 | v1.0.0 | ✅ PASS | João Silva |
| 2025-10-26 | v1.0.1 | ✅ PASS | QA Bot |
```

## 📊 Matriz de Cobertura de Testes

### Entry Operations

| Operação | Happy Path | Error Cases | Edge Cases | Total |
|----------|-----------|-------------|------------|-------|
| CreateEntry | TC-001, TC-002 | TC-003, TC-004, TC-005 | TC-006 | 6 |
| GetEntry | TC-007 | TC-008 | - | 2 |
| UpdateEntry | TC-009 | - | - | 1 |
| DeleteEntry | TC-010 | TC-011 | - | 2 |
| **TOTAL** | **4** | **4** | **1** | **11** |

### Claim Operations (30 dias)

| Operação | Happy Path | Timeout | Error Cases | Total |
|----------|-----------|---------|-------------|-------|
| CreateClaim | TC-012 | TC-013 | TC-014 | 3 |
| AcceptClaim | TC-015 | - | TC-016 | 2 |
| RejectClaim | TC-017 | - | - | 1 |
| **TOTAL** | **3** | **1** | **2** | **6** |

## 🚀 Estratégia de Automação

### Pirâmide de Testes

```
        /\
       /E2E\       10% - Testes E2E (Cypress, Playwright)
      /------\
     /  API  \     30% - Testes de API (Supertest, REST Assured)
    /----------\
   /   Unit    \  60% - Testes unitários (Jest, Go test)
  /--------------\
```

### Ferramentas

- **Testes Unitários (Go)**: `go test`, `testify/assert`
- **Testes de Integração**: `testcontainers` (PostgreSQL, Redis, Pulsar)
- **Testes E2E**: `Postman`, `Newman`, `k6` (load testing)
- **Testes de Segurança**: `OWASP ZAP`, `SonarQube`

## 📋 Checklist de Testes

### Antes de Deploy

- [ ] Todos os testes unitários passando (100%)
- [ ] Cobertura de código > 80%
- [ ] Testes de integração passando (100%)
- [ ] Testes E2E críticos passando (smoke tests)
- [ ] Testes de performance (load test) passando
- [ ] Testes de segurança executados (sem vulnerabilidades críticas)
- [ ] Testes de regressão executados

## 📚 Referências

- [Estratégia de Testes](../Estrategia_Testes.md)
- [Test Plan](../Test_Plan.md)
- [User Stories](../../01_Requisitos/UserStories/)
- [APIs gRPC](../../04_APIs/gRPC/)

---

**Status**: 🔴 Pasta vazia (será preenchida na Fase 2+)
**Fase de Preenchimento**: Fase 2 (paralelo ao desenvolvimento)
**Responsável**: QA Lead + Desenvolvedores
**Ferramenta**: Jira, TestRail, Xray
