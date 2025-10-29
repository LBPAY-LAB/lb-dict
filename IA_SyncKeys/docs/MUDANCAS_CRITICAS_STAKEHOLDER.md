# 🔴 MUDANÇAS CRÍTICAS - Validadas com Stakeholder

**Data**: 2024-10-28
**Status**: ✅ VALIDADAS E PRONTAS PARA IMPLEMENTAÇÃO

---

## 🚨 MUDANÇA #1: Container Separado `dict.vsync`

### ❌ ANTES (Incorreto)
Implementar no `orchestration-worker` existente

### ✅ AGORA (Correto)
Criar **NOVO container separado** `apps/dict.vsync/`

**Razão**: Isolamento de responsabilidades - CID/VSync merece container dedicado

**Impacto**:
- Nova estrutura de diretórios
- Novo Dockerfile
- Novos manifestos Kubernetes
- Novo go.mod independente

---

## 🚨 MUDANÇA #2: Pulsar Topic Existente

### ❌ ANTES (Incorreto)
Criar novos topics `key.created` e `key.updated`

### ✅ AGORA (Correto)
Usar topic **EXISTENTE**: `persistent://lb-conn/dict/dict-events`

**Validação do Stakeholder**:
> "O Connector-Dict publica eventos de alteração de estado no tópico `PULSAR_TOPIC_DICT_EVENTS=persistent://lb-conn/dict/dict-events` e o conteúdo deste evento deve conter informações suficientes para conseguir calcular o CID."

**Impacto**:
- Consumir topic existente (não criar novo)
- Filtrar eventos relevantes (Entry state changes)
- Event schema já definido no Dict API

---

## 🚨 MUDANÇA #3: Timestamps SEM DEFAULT

### ❌ ANTES (Incorreto)
```sql
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
```

### ✅ AGORA (Correto)
```sql
created_at TIMESTAMP NOT NULL  -- SEM DEFAULT!
updated_at TIMESTAMP NOT NULL  -- SEM DEFAULT!
-- NO TRIGGERS
```

**Razão do Stakeholder**:
> "A intenção em ter as colunas NOT NULL e sem valor default é para que o sistema seja obrigado a informar estes valores no momento em que estiver executando. Desta maneira garantimos que os horários, que são extremamente importantes para consistência, que aparecem em logs e trilhas de auditoria são exatamente os mesmos."

**Sincronização**:
- Pods/containers sincronizados com Kubernetes cluster
- Kubernetes cluster sincronizado com Banco Central
- Garante correlação perfeita com logs/audit/traces

**Impacto**:
- Aplicação DEVE fornecer timestamps explícitos
- Usar `time.Now().UTC()` no código Go
- NUNCA confiar em DEFAULT do banco

---

## 🚨 MUDANÇA #4: Dados Já Normalizados

### ❌ ANTES (Assumido)
Precisamos normalizar dados dos eventos

### ✅ AGORA (Validado)
Dados **JÁ estão normalizados** quando eventos são publicados

**Validação do Stakeholder**:
> "Uma alteração de estado não aceita pelo Banco Central se não estiver com os dados normalizados. Como os eventos só serão disparados quando a alteração de estado for com sucesso, isso garante que os dados contidos no evento estão de acordo com as regras."

**Impacto**:
- NÃO precisamos adicionar lógica de normalização
- Podemos confiar nos dados do evento
- CID pode ser gerado diretamente dos dados do evento

---

## 🚨 MUDANÇA #5: Sem REST Endpoints Novos

### ❌ ANTES (Confuso)
Criar endpoint POST `/api/v2/sync-verifications/`

### ✅ AGORA (Clarificado)
**NÃO criar endpoints REST**. Usar:
1. **Pulsar Consumer** para eventos de Entry
2. **Temporal Cron Workflow** para verificação periódica

**Validação do Stakeholder**:
> "O consumer que será adicionado deve estar dentro do app dict.vsync, que é um container separado, exclusivo para as atividades de cálculo de CID e verificação de sincronismo. Este dict.vsync pode ter workflows Temporal para as atividades recorrentes de verificação de sincronismo, como o POST /api/v2/sync-verifications/"

**Interpretação Correta**:
- POST /api/v2/sync-verifications/ NÃO é endpoint REST
- É referência ao workflow Temporal que faz verificação
- Chamado via Temporal Cron, não HTTP

**Impacto**:
- ZERO modificações no Dict API (apps/dict)
- ZERO novos endpoints HTTP
- Tudo via Pulsar Events + Temporal Workflows

---

## 📋 Checklist de Conformidade

Antes de iniciar implementação, validar:

- [ ] Nova estrutura `apps/dict.vsync/` criada
- [ ] Consumir topic `persistent://lb-conn/dict/dict-events`
- [ ] Migrations SEM DEFAULT em `created_at` e `updated_at`
- [ ] NÃO adicionar lógica de normalização (dados já normalizados)
- [ ] NÃO criar novos endpoints REST
- [ ] Temporal Cron Workflow configurado
- [ ] Timestamps UTC fornecidos explicitamente pela aplicação

---

## 🎯 Estrutura Correta (Resumo)

```
connector-dict/
├── apps/
│   ├── dict/                    # NÃO MODIFICAR
│   │   └── (publica eventos para dict-events)
│   │
│   ├── dict.vsync/              # 🆕 NOVO CONTAINER SEPARADO
│   │   ├── cmd/
│   │   │   └── worker/main.go   # Entry point
│   │   ├── internal/
│   │   │   ├── domain/          # CID, VSync entities
│   │   │   ├── application/     # Use cases
│   │   │   └── infrastructure/
│   │   │       ├── database/    # Migrations + Repositories
│   │   │       ├── temporal/    # Workflows + Activities
│   │   │       ├── pulsar/      # Consumer (dict-events)
│   │   │       └── grpc/        # Bridge client
│   │   ├── Dockerfile
│   │   └── go.mod
│   │
│   └── orchestration-worker/   # NÃO MODIFICAR
│
└── k8s/
    └── dict.vsync/              # 🆕 Kubernetes manifests
```

---

## 🚦 Ordem de Execução

### Fase 0: Validação Técnica (AGORA)
1. Analisar event schema do topic `dict-events`
2. Verificar campos disponíveis para CID
3. Coordenar com Bridge: endpoints VSync disponíveis?
4. Criar estrutura `apps/dict.vsync/`

### Fase 1: Database Layer
1. Migrations com timestamps NOT NULL (sem default)
2. Repositories
3. Testes de integração

### Fase 2: Domain & Application
1. CID Generator (SHA-256)
2. VSync Calculator (XOR)
3. Use cases
4. Unit tests

### Fase 3: Temporal Workflows
1. VSyncVerificationWorkflow (Cron)
2. ReconciliationWorkflow (Child)
3. Activities (10+)
4. Workflow tests

### Fase 4: Pulsar Integration
1. Consumer para `dict-events`
2. Filtrar eventos de Entry
3. Handler para processar eventos
4. Integration tests

### Fase 5: QA & Deploy
1. E2E tests
2. Security audit
3. Documentation
4. K8s deployment

---

**✅ Mudanças validadas e documentadas. Pronto para iniciar Fase 0.**

**Próximo passo**: Executar `/orchestrate-implementation` com estas mudanças críticas.
