# DAT-003: Estratégia de Migrations (Database Versioning)

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Componente**: Database Migrations (Core DICT + Connect)
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: ARCHITECT (AI Agent - Technical Architect)
**Revisor**: [Aguardando]
**Aprovador**: Tech Lead, DBA, DevOps Lead

---

## Sumário Executivo

Este documento define a **estratégia de versionamento e migrations** para os bancos de dados do projeto DICT (Core DICT e Connect), resolvendo o gap identificado em [ANA-003](../00_Analises/ANA-003_Analise_Repo_Connect.md) onde migrations não foram encontradas.

**Ferramenta Escolhida**: **Goose** (para projetos Go) ou **Flyway** (agnóstico)

**Baseado em**:
- [ANA-003: Análise Repositório Connect](../00_Analises/ANA-003_Analise_Repo_Connect.md) - "Migrations pendentes"
- [DAT-001: Schema Core DICT](DAT-001_Schema_Database_Core_DICT.md)
- [DAT-002: Schema Connect](DAT-002_Schema_Database_Connect.md)

---

## Controle de Versão

| Versão | Data | Autor | Descrição |
|--------|------|-------|-----------|
| 1.0 | 2025-10-25 | ARCHITECT | Versão inicial - Estratégia de Migrations |

---

## Índice

1. [Escolha da Ferramenta](#1-escolha-da-ferramenta)
2. [Convenções de Nomenclatura](#2-convenções-de-nomenclatura)
3. [Estrutura de Diretórios](#3-estrutura-de-diretórios)
4. [Workflow de Migrations](#4-workflow-de-migrations)
5. [Rollback Strategy](#5-rollback-strategy)
6. [Environments](#6-environments)
7. [CI/CD Integration](#7-cicd-integration)
8. [Best Practices](#8-best-practices)

---

## 1. Escolha da Ferramenta

### 1.1. Comparação: Goose vs Flyway

| Critério | Goose | Flyway | Decisão |
|----------|-------|--------|---------|
| **Linguagem** | Go nativo | Java (agnóstico) | 🏆 **Goose** (projetos Go) |
| **Performance** | Rápido | Médio | 🏆 Goose |
| **Rollback** | Sim (down migrations) | Apenas versão paga | 🏆 Goose |
| **SQL + Code** | SQL + Go code | SQL apenas | 🏆 Goose (flexibilidade) |
| **CI/CD** | Simples | Simples | ⚖️ Empate |
| **Community** | Ativa | Muito ativa | Flyway |
| **Custo** | Free/OSS | Free + Pro | 🏆 Goose |

**Decisão**: **Goose** para Core DICT e Connect (ambos são projetos Go)

### 1.2. Instalação Goose

```bash
# Via Go
go install github.com/pressly/goose/v3/cmd/goose@latest

# Via Homebrew (macOS)
brew install goose

# Verificar instalação
goose --version
```

---

## 2. Convenções de Nomenclatura

### 2.1. Formato de Migration Files

```
{timestamp}_{description}.sql
```

**Componentes**:
- `{timestamp}`: YYYYMMDDHHmmss (14 dígitos)
- `{description}`: Snake_case, descritivo

**Exemplos**:
```
20251025100000_create_schema_dict.sql
20251025100100_create_table_entries.sql
20251025100200_create_table_accounts.sql
20251025100300_create_table_claims.sql
20251025100400_create_indexes_entries.sql
20251025100500_create_triggers_audit.sql
```

### 2.2. Prefixos Semânticos (Opcional)

```
{timestamp}_{tipo}_{descrição}.sql

Tipos:
- create_*    → Criar schema/table/index
- alter_*     → Alterar estrutura
- drop_*      → Remover elementos
- data_*      → Migrations de dados
- fix_*       → Correções
```

**Exemplos**:
```
20251025100000_create_schema_dict.sql
20251025110000_alter_add_column_entries_sync_status.sql
20251025120000_data_populate_default_users.sql
20251025130000_fix_constraint_claims_expires_at.sql
```

---

## 3. Estrutura de Diretórios

### 3.1. Estrutura Core DICT

```
lb-conn/core-dict/
├── db/
│   └── migrations/
│       ├── 20251025100000_create_schema_dict.sql
│       ├── 20251025100100_create_table_entries.sql
│       ├── 20251025100200_create_table_accounts.sql
│       ├── 20251025100300_create_table_claims.sql
│       ├── 20251025100400_create_table_portabilities.sql
│       ├── 20251025100500_create_table_users.sql
│       ├── 20251025100600_create_schema_audit.sql
│       ├── 20251025100700_create_table_entry_events.sql
│       ├── 20251025100800_create_indexes_entries.sql
│       ├── 20251025100900_create_indexes_claims.sql
│       ├── 20251025101000_create_function_update_updated_at.sql
│       ├── 20251025101100_create_triggers_updated_at.sql
│       ├── 20251025101200_create_function_audit_entry_changes.sql
│       ├── 20251025101300_create_triggers_audit.sql
│       ├── 20251025101400_create_function_expire_old_claims.sql
│       └── 20251025101500_data_insert_default_admin_user.sql
│
├── scripts/
│   ├── migrate.sh           # Script helper para migrations
│   └── rollback.sh          # Script helper para rollback
│
└── Makefile                 # Tasks de migrations
```

### 3.2. Estrutura Connect

```
lb-conn/connector-dict/
├── db/
│   └── migrations/
│       ├── 20251025200000_create_schema_connect.sql
│       ├── 20251025200100_create_schema_workflows.sql
│       ├── 20251025200200_create_table_claim_workflows.sql
│       ├── 20251025200300_create_table_bridge_requests.sql
│       ├── 20251025200400_create_schema_audit.sql
│       ├── 20251025200500_create_table_workflow_events.sql
│       ├── 20251025200600_create_indexes_workflows.sql
│       ├── 20251025200700_create_triggers_workflows.sql
│       ├── 20251025200800_create_table_vsync_workflows.sql      # FUTURO
│       └── 20251025200900_create_table_otp_workflows.sql        # FUTURO
│
├── scripts/
│   ├── migrate.sh
│   └── rollback.sh
│
└── Makefile
```

---

## 4. Workflow de Migrations

### 4.1. Criar Nova Migration

```bash
# Core DICT
cd lb-conn/core-dict
goose -dir db/migrations create add_column_entries_external_id sql

# Connect
cd lb-conn/connector-dict
goose -dir db/migrations create add_table_vsync_workflows sql
```

**Output**:
```
Created new file: db/migrations/20251025150000_add_column_entries_external_id.sql
```

### 4.2. Estrutura de Migration File

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS dict.entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key_type VARCHAR(20) NOT NULL,
    key_value VARCHAR(255) NOT NULL,
    -- ... outras colunas
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS dict.entries CASCADE;
-- +goose StatementEnd
```

**Comentários especiais do Goose**:
- `-- +goose Up`: Código para aplicar migration
- `-- +goose Down`: Código para reverter migration
- `-- +goose StatementBegin` / `-- +goose StatementEnd`: Delimitadores para statements complexos

### 4.3. Aplicar Migrations

```bash
# Desenvolvimento (local)
goose -dir db/migrations postgres "user=postgres dbname=lbpay_core_dict sslmode=disable" up

# Staging
goose -dir db/migrations postgres "user=dict_app password=${DB_PASSWORD} host=staging-db.internal dbname=lbpay_core_dict sslmode=require" up

# Produção (com confirmação)
goose -dir db/migrations postgres "${DATABASE_URL}" up
```

### 4.4. Verificar Status

```bash
# Ver status de migrations
goose -dir db/migrations postgres "${DATABASE_URL}" status

# Output
    Applied At                  Migration
    =======================================
    Mon Oct 25 10:00:00 2025 -- 20251025100000_create_schema_dict.sql
    Mon Oct 25 10:01:00 2025 -- 20251025100100_create_table_entries.sql
    Mon Oct 25 10:02:00 2025 -- 20251025100200_create_table_accounts.sql
    Pending                  -- 20251025100300_create_table_claims.sql
    Pending                  -- 20251025100400_create_indexes_entries.sql
```

---

## 5. Rollback Strategy

### 5.1. Rollback de 1 Migration

```bash
# Reverter última migration
goose -dir db/migrations postgres "${DATABASE_URL}" down

# Verificar
goose -dir db/migrations postgres "${DATABASE_URL}" status
```

### 5.2. Rollback para Versão Específica

```bash
# Rollback até versão específica
goose -dir db/migrations postgres "${DATABASE_URL}" down-to 20251025100200

# Isso reverterá todas migrations após 20251025100200
```

### 5.3. Rollback Total (⚠️ PERIGOSO)

```bash
# Reverter TODAS as migrations (apenas dev/test)
goose -dir db/migrations postgres "${DATABASE_URL}" reset

# ⚠️ NUNCA executar em produção sem backup!
```

### 5.4. Re-aplicar (Redo)

```bash
# Reverter e re-aplicar última migration (útil para debug)
goose -dir db/migrations postgres "${DATABASE_URL}" redo
```

---

## 6. Environments

### 6.1. Configuração por Environment

```bash
# Development (local)
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=lbpay_core_dict
export DB_SSLMODE=disable

# Staging
export DB_HOST=staging-db.internal
export DB_PORT=5432
export DB_USER=dict_app
export DB_PASSWORD=$(vault kv get -field=password secret/staging/db/dict)
export DB_NAME=lbpay_core_dict
export DB_SSLMODE=require

# Production
export DB_HOST=prod-db-primary.internal
export DB_PORT=5432
export DB_USER=dict_app
export DB_PASSWORD=$(vault kv get -field=password secret/production/db/dict)
export DB_NAME=lbpay_core_dict
export DB_SSLMODE=require
export DB_SSLCERT=/etc/ssl/certs/db-client.crt
export DB_SSLKEY=/etc/ssl/certs/db-client.key
export DB_SSLROOTCERT=/etc/ssl/certs/db-ca.crt
```

### 6.2. Connection Strings

```bash
# Helper para gerar connection string
function db_url() {
    echo "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"
}

# Uso
goose -dir db/migrations postgres "$(db_url)" up
```

---

## 7. CI/CD Integration

### 7.1. Makefile

```makefile
# Makefile
.PHONY: migrate migrate-down migrate-status migrate-create

# Variables
DB_URL ?= postgres://postgres:postgres@localhost:5432/lbpay_core_dict?sslmode=disable
MIGRATIONS_DIR = db/migrations

migrate:
	@echo "Aplicando migrations..."
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up

migrate-down:
	@echo "Revertendo última migration..."
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" down

migrate-status:
	@echo "Status das migrations:"
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" status

migrate-create:
	@read -p "Nome da migration: " name; \
	goose -dir $(MIGRATIONS_DIR) create $$name sql

migrate-reset:
	@echo "⚠️  ATENÇÃO: Isso reverterá TODAS as migrations!"
	@read -p "Tem certeza? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" reset; \
	else \
		echo "Operação cancelada."; \
	fi
```

**Uso**:
```bash
# Aplicar migrations
make migrate

# Criar nova migration
make migrate-create

# Ver status
make migrate-status

# Rollback
make migrate-down
```

### 7.2. GitHub Actions (CI/CD)

```yaml
# .github/workflows/db-migrations.yml
name: Database Migrations

on:
  push:
    branches: [main, develop]
    paths:
      - 'db/migrations/**'

jobs:
  migrate-staging:
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/develop'

    steps:
      - uses: actions/checkout@v3

      - name: Install Goose
        run: |
          curl -fsSL https://raw.githubusercontent.com/pressly/goose/master/install.sh | sh
          sudo mv ./bin/goose /usr/local/bin/

      - name: Run Migrations (Staging)
        env:
          DB_URL: ${{ secrets.STAGING_DB_URL }}
        run: |
          goose -dir db/migrations postgres "$DB_URL" status
          goose -dir db/migrations postgres "$DB_URL" up

  migrate-production:
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    needs: [migrate-staging]  # Apenas após staging OK
    environment: production   # Require approval

    steps:
      - uses: actions/checkout@v3

      - name: Install Goose
        run: |
          curl -fsSL https://raw.githubusercontent.com/pressly/goose/master/install.sh | sh
          sudo mv ./bin/goose /usr/local/bin/

      - name: Backup Database
        env:
          DB_URL: ${{ secrets.PRODUCTION_DB_URL }}
        run: |
          pg_dump "$DB_URL" > backup_$(date +%Y%m%d_%H%M%S).sql

      - name: Run Migrations (Production)
        env:
          DB_URL: ${{ secrets.PRODUCTION_DB_URL }}
        run: |
          goose -dir db/migrations postgres "$DB_URL" status
          goose -dir db/migrations postgres "$DB_URL" up

      - name: Upload Backup
        uses: actions/upload-artifact@v3
        with:
          name: db-backup
          path: backup_*.sql
          retention-days: 30
```

---

## 8. Best Practices

### 8.1. ✅ DO

1. **Sempre incluir `-- +goose Down`**
   ```sql
   -- Permite rollback
   -- +goose Down
   DROP TABLE IF EXISTS dict.entries CASCADE;
   ```

2. **Usar transações implícitas**
   ```sql
   -- Goose wraps em transaction automaticamente
   -- Se falhar, faz rollback
   ```

3. **Testar em dev/staging antes de prod**
   ```bash
   # Dev
   make migrate DB_URL=dev

   # Staging
   make migrate DB_URL=staging

   # Prod (apenas se staging OK)
   make migrate DB_URL=prod
   ```

4. **Versionamento semântico**
   ```
   20251025_v1.0.0_initial_schema.sql
   20251026_v1.1.0_add_vsync_table.sql
   20251027_v1.1.1_fix_claims_constraint.sql
   ```

5. **Documentar migrations complexas**
   ```sql
   -- +goose Up
   -- Migration: Adiciona coluna external_id
   -- Razão: Armazenar ID retornado pelo Bacen
   -- Jira: DICT-123
   ALTER TABLE dict.entries ADD COLUMN external_id VARCHAR(100);
   ```

### 8.2. ❌ DON'T

1. **Nunca editar migration já aplicada em prod**
   ```bash
   # ❌ Errado
   vim db/migrations/20251025100000_create_table_entries.sql

   # ✅ Correto: criar nova migration
   goose create fix_entries_table sql
   ```

2. **Nunca fazer migrations de dados grandes sem batch**
   ```sql
   -- ❌ Errado (pode travar DB)
   UPDATE dict.entries SET status = 'ACTIVE';

   -- ✅ Correto
   DO $$
   DECLARE
       batch_size INT := 1000;
   BEGIN
       LOOP
           UPDATE dict.entries
           SET status = 'ACTIVE'
           WHERE id IN (
               SELECT id FROM dict.entries
               WHERE status IS NULL
               LIMIT batch_size
           );
           EXIT WHEN NOT FOUND;
           COMMIT;
       END LOOP;
   END $$;
   ```

3. **Nunca usar `DROP ... CASCADE` em produção sem cuidado**
   ```sql
   -- ❌ Muito perigoso
   DROP TABLE dict.entries CASCADE;

   -- ✅ Mais seguro
   DROP TABLE IF EXISTS dict.entries_old;  -- Tabela já renomeada
   ```

4. **Nunca rodar migrations manualmente em prod sem backup**
   ```bash
   # ❌ Errado
   goose up  # direto em prod

   # ✅ Correto
   pg_dump $DB_URL > backup.sql
   goose up
   ```

---

## 9. Troubleshooting

### 9.1. Migration Falha no Meio

**Problema**: Migration falhou, mas metade das queries executaram

**Solução**:
```bash
# 1. Verificar status
goose status

# 2. Se migration está marcada como aplicada mas falhou
# Marcar manualmente como não aplicada
psql $DB_URL -c "DELETE FROM goose_db_version WHERE version_id = 20251025100000;"

# 3. Corrigir migration file
vim db/migrations/20251025100000_*.sql

# 4. Re-aplicar
goose up
```

### 9.2. Migrations Out of Order

**Problema**: Alguém criou migration com timestamp antigo

**Solução**:
```bash
# Goose detecta e alerta
# Renumerar migration manualmente
mv 20251024000000_new.sql 20251025120000_new.sql

# Ou forçar aplicação
goose up-by-one  # Aplica próxima migration pendente
```

---

## Próximas Revisões

**Pendências**:
- [ ] Definir estratégia de backup antes de migrations em prod
- [ ] Implementar validação de migrations em PR (linting)
- [ ] Criar dashboard de status de migrations (Grafana)
- [ ] Definir política de retenção de backups

---

**Referências**:
- [Goose Documentation](https://github.com/pressly/goose)
- [DAT-001: Schema Core DICT](DAT-001_Schema_Database_Core_DICT.md)
- [DAT-002: Schema Connect](DAT-002_Schema_Database_Connect.md)
- [ANA-003: Análise Connect](../00_Analises/ANA-003_Analise_Repo_Connect.md)
