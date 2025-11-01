# Questões para Esclarecimento - DICT Rate Limit Monitoring

## 🎯 Objetivo

Documento para registrar dúvidas técnicas, decisões arquiteturais e coordenação com times externos que precisam ser resolvidas ANTES de iniciar a implementação.

**Responsável**: Tech Lead
**Data Criação**: 2025-10-31
**Status**: 🔴 PENDENTE RESPOSTAS

---

## 🔴 QUESTÕES CRÍTICAS (Bloqueantes)

### 1. Integração com Bridge gRPC ⚠️ **BLOQUEANTE**

**Contexto**: O sistema precisa consultar `/policies` e `/policies/{policy}` do DICT BACEN via Bridge.

**Questões**:

1.1. **O Bridge JÁ possui endpoints implementados para consulta de policies?**
   - [x] ✅ SIM - Endpoints existem
   - [ ] ❌ NÃO - Precisa implementar no Bridge
   - [ ] ⚠️ PARCIAL - Alguns endpoints existem

1.2. **Se SIM, quais são os proto definitions existentes?**
   - Path: `github.com/lb-conn/rsfn-connect-bacen-bridge/proto/...`
   - Methods: `ListPolicies()`, `GetPolicy()`?
   - Request/Response types: `ListPoliciesRequest`, `PolicyResponse`?

1.3. **Se NÃO, qual o timeline para implementação no Bridge?**
   - Estimativa de prazo: ______ dias/semanas
   - Responsável no time Bridge: ____________
   - Prioridade: Alta / Média / Baixa

1.4. **Mappers Bacen ↔ gRPC já existem no SDK?**
   - Location: `github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/mappers/ratelimit`
   - [x] Existem e podem ser reutilizados
   - [ ] Precisam ser criados do zero

**Ação Requerida**:
- [ ] Coordenar reunião com time Bridge (urgente)
- [ ] Validar proto definitions
- [ ] Alinhar timeline de implementação
- [ ] Definir contratos de interface

**Impacto se não resolvido**: ⛔ **BLOQUEIO TOTAL** - Não é possível implementar sem comunicação com DICT

---

### 2. Integração com Core-Dict

**Contexto**: Alertas precisam ser publicados no Pulsar para consumo do Core-Dict.

**Questões**:

2.1. **Core-Dict já consome eventos do topic `core-events`?**
   - [ ] SIM - Consumer ativo
   - [ ] NÃO - Precisa implementar consumer
   - [x] NÃO - Outro time está implementando consumer. Pubicar no tópico é suficiente.

2.2. **Qual schema de evento usar para alertas de rate limit?**
   - Action sugerido: `pkg.ActionRateLimitAlert`
   - Payload format: JSON com `{policy, severity, utilization, message}`
   - [x] Schema aprovado
   - [ ] Precisa ajustar schema

2.3. **Core-Dict precisa tomar alguma ação automatizada ao receber alertas?**
   - [ ] SIM - Qual? ___________________________________
   - [ ] NÃO - Apenas logging/dashboard
   - [x] FUTURO - Planejar mas não implementar agora

**Ação Requerida**:
- [ ] Coordenar com time Core-Dict
- [ ] Validar schema de eventos
- [ ] Definir ações esperadas

**Impacto se não resolvido**: ⚠️ **MÉDIO** - Sistema funciona, mas Core-Dict não recebe notificações

---

### 3. Thresholds de Alerta

**Contexto**: Definir níveis de WARNING e CRITICAL para disparo de alertas.

**Questões**:

3.1. **Thresholds propostos estão adequados?**
   - WARNING: 20% restante (utilização >80%)
   - CRITICAL: 10% restante (utilização >90%)
   - [x] APROVADO
   - [ ] AJUSTAR para: WARNING=____%, CRITICAL=____%

3.2. **Thresholds devem ser globais ou configuráveis por política?**
   - [x] GLOBAIS - Mesmos valores para todas as políticas
   - [ ] CONFIGURÁVEIS - Cada política pode ter seus thresholds
   - Justificativa: ___________________________________________

3.3. **Qual a ação esperada quando CRITICAL threshold é atingido?**
   - [x] Apenas alerta (sem throttling)
   - [ ] Alerta + notificação PagerDuty/Slack
   - [ ] Alerta + throttling automático (out of scope inicial?)
   - [ ] Outro: _______________________________________________

**Ação Requerida**:
- [ ] Validar com stakeholder/produto
- [ ] Definir SLAs de resposta a alertas
- [x] Documentar playbooks de ação
  
**Impacto se não resolvido**: ⚠️ **MÉDIO** - Pode gerar falsos positivos ou alertas tardios

---

### 4. Frequência de Monitoramento

**Contexto**: Temporal cron workflow para monitoramento contínuo.

**Questões**:

4.1. **Frequência proposta (a cada 5 minutos) está adequada?**
   - Proposta: `*/5 * * * *` (12 checks/hora, 288 checks/dia)
   - [x] APROVADO
   - [ ] AJUSTAR para: _____________ (ex: 1min, 10min, 15min)

4.2. **Impacto de custos (chamadas ao DICT)?**
   - Custo DICT: ~288 calls/dia * 30 dias = ~8.640 calls/mês
   - Política aplicável: `POLICIES_LIST` (limite: 6/min, balde: 20)
   - 12 calls/hora < 6/min? ✅ SIM, está dentro do limite
   - [x] Confirmar que custo/rate limit estão ok

4.3. **Cache Redis deve ser usado para reduzir chamadas ao DICT?**
   - TTL proposto: 60s
   - [ ] SIM, usar cache (reduz chamadas em 1 minuto)
   - [x] NÃO, sempre consultar DICT (dados sempre fresh)

**Ação Requerida**:
- [ ] Validar custo-benefício frequência vs freshness
- [ ] Definir TTL de cache
- [ ] Validar limites de rate do DICT

**Impacto se não resolvido**: ⚠️ **BAIXO** - Pode ajustar após deploy

---

### 5. Categorização do Participante (PSP)

**Contexto**: Algumas políticas variam conforme a **categoria do participante** (A-H), especialmente `ENTRIES_READ_PARTICIPANT_ANTISCAN`.

**Questões**:

5.1. **Em qual categoria (A-H) o LBPay está enquadrado no DICT?**
   - [ ] Categoria A (25K/min refill, 50K capacity)
   - [ ] Categoria B (20K/min refill, 40K capacity)
   - [ ] Categoria C (15K/min refill, 30K capacity)
   - [ ] Categoria D (8K/min refill, 16K capacity)
   - [ ] Categoria E (2.5K/min refill, 5K capacity)
   - [ ] Categoria F (250/min refill, 500 capacity)
   - [ ] Categoria G (25/min refill, 250 capacity)
   - [ ] Categoria H (2/min refill, 50 capacity)
   - [x] NÃO SEI - Precisa consultar DICT

5.2. **Como obter a categoria atual do PSP?**
   - Endpoint DICT: `GET /policies/` retorna campo `<Category>A</Category>`
   - [ ] Confirmar que Bridge mapeia este campo
   - [ ] Armazenar categoria no PostgreSQL?
   - [x] Monitorar mudanças de categoria (PSP pode ser rebaixado/promovido)?

5.3. **Categoria pode mudar durante operação?**
   - Contexto: BACEN pode reclassificar PSP baseado em volume/comportamento
   - [x] SIM - Implementar detecção de mudança de categoria
   - [ ] NÃO - Categoria é fixa
   - [ ] NÃO SEI - Precisa validar com BACEN

**Ação Requerida**:
- [ ] Consultar categoria atual do LBPay no DICT
- [ ] Validar se Bridge retorna campo `Category`
- [x] Precisa alertar quando categoria muda

**Impacto se não resolvido**: 🔴 **CRÍTICO** - Thresholds e alertas podem ser calculados incorretamente para políticas variáveis

---

### 6. Penalidades Anti-Scan (Erro 404)

**Contexto**: Consultas de chaves que retornam 404 consomem **3 fichas** em vez de 1 (anti-scan).

**Questões**:

6.1. **O sistema deve alertar quando há consumo excessivo por erros 404?**
   - Cenário: PSP fazendo varreduras inadvertidamente (bug no código)
   - Consumo: 1,000 erros 404 = 3,000 fichas (vs 1,000 fichas se fossem 200)
   - [x] SIM - Criar métrica específica para taxa de erro 404
   - [ ] NÃO - Não é escopo do monitoramento

6.2. **Como diferenciar consumo normal vs consumo por penalidade?**
   - DICT retorna apenas `AvailableTokens` atual (não detalha histórico de consumo)
   - [ ] Não é possível diferenciar (aceitar limitação)
   - [x] Criar métrica derivada baseada em logs do Dict API
   - [ ] Consultar endpoint adicional do DICT? (se existir)

**Ação Requerida**:
- [ ] Definir se monitoramento de padrão de consumo está em escopo
- [ ] Validar se DICT fornece breakdown de consumo por tipo

**Impacto se não resolvido**: ⚠️ **MÉDIO** - Dificulta diagnóstico de esgotamento rápido de balde

---

### 7. Política CIDS_FILES_WRITE (RefillPeriodSec = 86400s / 1 dia)

**Contexto**: Política `CIDS_FILES_WRITE` tem período de reposição de **1 dia** (vs 60s das outras).

**Questões**:

7.1. **Como monitorar política com refill diário?**
   - Capacity: 200 fichas
   - RefillTokens: 40 fichas/dia
   - Cron a cada 5min pode ser muito frequente?
   - [x] Monitorar normalmente (5min)
   - [ ] Criar cron separado para políticas diárias (ex: 1x/dia)
   - [ ] Monitorar apenas políticas críticas (ignorar CIDS_FILES)

7.2. **Thresholds devem ser diferentes para políticas diárias?**
   - WARNING: 20% de 200 = 50 fichas (consumo de 1-2 dias)
   - CRITICAL: 10% de 200 = 20 fichas (consumo < 1 dia)
   - [x] Usar mesmos thresholds (20%/10%)
   - [ ] Ajustar thresholds: WARNING=___%, CRITICAL=___%

7.3. **Políticas com RefillPeriodSec > 60s devem ter alertas especiais?**
   - Contexto: Recuperação lenta (1 dia vs 1 minuto)
   - [ ] SIM - Alerta antecipado (ex: WARNING em 50% em vez de 25%)
   - [x] NÃO - Manter padrão

**Ação Requerida**:
- [ ] Listar todas as políticas com RefillPeriodSec != 60
- [ ] Definir estratégia de monitoramento diferenciada
- [ ] Validar se essas políticas são críticas para o negócio

**Impacto se não resolvido**: ⚠️ **MÉDIO** - Alertas podem ser tardios para políticas de refill lento

---

### 8. Tempo de Recuperação Total do Balde

**Contexto**: Balde esgotado leva **tempo considerável** para reabastecimento completo.

**Questões**:

8.1. **Sistema deve calcular e exibir ETA (Estimated Time to Recovery)?**
   - Fórmula: `TimeToFull = ((Capacity - AvailableTokens) / RefillTokens) * RefillPeriodSec`
   - Exemplo: ENTRIES_WRITE esgotado = (36,000 / 1,200) * 60s = 1,800s = **30 minutos**
   - [x] SIM - Adicionar campo `time_to_full_recovery` em alertas
   - [ ] NÃO - Não é necessário

8.2. **Dashboard deve mostrar projeção de esgotamento?**
   - Baseado em taxa de consumo recente: "No ritmo atual, balde esgotará em X minutos"
   - Requer calcular taxa de consumo: `(AvailableTokens_t0 - AvailableTokens_t1) / (t1 - t0)`
   - [x] SIM - Feature valiosa para proatividade
   - [ ] NÃO - Complexidade alta, low priority
   - [ ] FUTURO - Fase 2

8.3. **Alertas devem incluir recomendação de ação?**
   - Exemplo: "CRITICAL: ENTRIES_WRITE em 5%. Reduzir criação de chaves pelos próximos 15 minutos"
   - [ ] SIM - Mensagens acionáveis
   - [x] NÃO - Apenas dados brutos

**Ação Requerida**:
- [ ] Definir se cálculos preditivos estão em escopo
- [ ] Validar UX de alertas com time de operações

**Impacto se não resolvido**: ✅ **BAIXO** - Feature nice-to-have, não bloqueante

---

### 9. Sincronização de Clock e Jitter

**Contexto**: Token Bucket depende de timestamps precisos para reposição correta.

**Questões**:

9.1. **Como garantir que timestamps estão sincronizados?**
   - Sistema usa: `checked_at TIMESTAMP WITH TIME ZONE`
   - [ ] Confiar em NTP do servidor PostgreSQL
   - [ ] Confiar em NTP do servidor Temporal
   - [x] Usar timestamp retornado pelo DICT (`<ResponseTime>`)

9.2. **Diferença de clock entre sistemas pode causar cálculos errados?**
   - Cenário: Temporal worker em UTC, PostgreSQL em America/Sao_Paulo
   - [x] SIM - Forçar UTC em todos os componentes
   - [ ] NÃO - PostgreSQL TIMESTAMPTZ é timezone-aware
   - [ ] Validar configuração de timezone

9.3. **Jitter de rede pode afetar precisão de `AvailableTokens`?**
   - Latência Bridge → DICT: ~200-500ms
   - Momento da consulta vs momento do cálculo pelo DICT
   - [ ] Aceitar margem de erro (não crítico)
   - [x] Adicionar timestamp do DICT nos logs (auditoria)

**Ação Requerida**:
- [x] Validar configuração de timezone em todos os componentes
- [ ] Documentar precision esperada vs real
- [ ] Definir se jitter é aceitável

**Impacto se não resolvido**: ⚠️ **MÉDIO** - Pode causar alertas falsos positivos/negativos

---

## 🟡 QUESTÕES IMPORTANTES (Não Bloqueantes)

### 10. Data Retention

**Questões**:

10.1. **Quanto tempo manter histórico de estados (dict_rate_limit_states)?**
   - Proposta: 13 meses (1 ano + mês corrente)
   - Storage estimado: ~300 rows/dia * 365 dias * 24 políticas = ~2.6M rows/ano (~500MB)
   - [x] APROVADO
   - [ ] AJUSTAR para: _______ meses

10.2. **Quanto tempo manter log de alertas (dict_rate_limit_alerts)?**
   - Proposta: Indefinido (audit trail permanente)
   - Storage estimado: ~10 alertas/dia * 365 dias = ~3.650 rows/ano (~5MB)
   - [x] APROVADO
   - [ ] AJUSTAR para: _______ meses

**Impacto se não resolvido**: ✅ **MÍNIMO** - Pode ajustar políticas de retenção depois

---

### 11. Observabilidade & Dashboards

**Questões**:

11.1. **Grafana dashboards devem ser criados como parte do projeto?**
   - [ ] SIM - Incluir JSON templates
   - [x] NÃO - Time de infra cria depois
   - [ ] PARCIAL - Criar templates básicos apenas

11.2. **Alertas Prometheus devem integrar com PagerDuty/Slack?**
   - [ ] SIM - Configurar integrações
   - [x] NÃO - Apenas Prometheus AlertManager local
   - [ ] FUTURO - Planejar mas não implementar agora

**Impacto se não resolvido**: ✅ **MÍNIMO** - Dashboards podem ser criados incrementalmente

---

### 12. Infraestrutura & Deployment

**Questões**:

12.1. **Helm charts devem ser criados?**
   - [ ] SIM - Chart completo (Dict API + Orchestration Worker + DB)
   - [x] NÃO - Usar Kubernetes manifests diretos
   - [ ] REUTILIZAR - Adaptar charts existentes do connector-dict

12.2. **Database migrations - Qual ferramenta?**
   - Opções: Goose / Flyway / golang-migrate / Liquibase
   - [x] Goose (Go-based, usado em projetos Go)
   - [ ] Flyway (Java-based, enterprise-grade)
   - [ ] Outro: _____________

12.3. **Secrets management - Qual solução?**
   - [ ] HashiCorp Vault
   - [x] AWS Secrets Manager (VALIDADO - infraestrutura AWS existente)
   - [ ] Kubernetes Secrets (não recomendado para prod)
   - [ ] Outro: _____________

**Impacto se não resolvido**: ⚠️ **MÉDIO** - Atrasa deployment para produção

---

## 🟢 QUESTÕES OPCIONAIS (Nice to Have)

### 13. Features Futuras

**Questões**:

13.1. **Implementar auto-throttling quando CRITICAL threshold é atingido?**
   - Contexto: Reduzir automaticamente taxa de requisições ao DICT para evitar 429
   - [ ] SIM - Incluir no escopo inicial
   - [ ] NÃO - Out of scope (apenas alerta)
   - [x] FUTURO - Fase 2 do projeto

13.2. **Dashboard público para stakeholders externos?**
   - Contexto: UI web para visualizar estado dos baldes em tempo real
   - [ ] SIM - Criar frontend React
   - [x] NÃO - Apenas APIs backend
   - [ ] FUTURO - Fase 2 do projeto

13.3. **Integração com sistemas de BI/Analytics?**
   - Contexto: Exportar dados para Data Lake / BigQuery / Snowflake
   - [ ] SIM - Implementar CDC (Change Data Capture)
   - [x] NÃO - Out of scope
   - [ ] FUTURO - Planejar arquitetura

**Impacto se não resolvido**: ✅ **NENHUM** - Features opcionais

---

## 📊 Matriz de Priorização

| # | Questão | Prioridade | Status | Responsável | Prazo |
|---|---------|-----------|--------|-------------|-------|
| 1.1-1.4 | Bridge gRPC Integration | 🔴 CRÍTICA | ⏳ Pendente | Tech Lead | 2 dias |
| 2.1-2.3 | Core-Dict Integration | 🟡 ALTA | ⏳ Pendente | Tech Lead | 3 dias |
| 3.1-3.3 | Thresholds de Alerta | 🟡 ALTA | ⏳ Pendente | Produto/Stakeholder | 2 dias |
| 4.1-4.3 | Frequência Monitoramento | 🟡 ALTA | ⏳ Pendente | Tech Lead | 2 dias |
| **5.1-5.3** | **Categorização PSP (A-H)** | **🔴 CRÍTICA** | **⏳ Pendente** | **Tech Lead + BACEN** | **3 dias** |
| **6.1-6.2** | **Penalidades Anti-Scan (404)** | **🟡 ALTA** | **⏳ Pendente** | **Produto** | **1 semana** |
| **7.1-7.3** | **Política CIDS_FILES_WRITE** | **🟡 ALTA** | **⏳ Pendente** | **Tech Lead** | **2 dias** |
| **8.1-8.3** | **Tempo Recuperação (ETA)** | **🟢 MÉDIA** | **⏳ Pendente** | **Produto** | **1 semana** |
| **9.1-9.3** | **Sincronização Clock/Jitter** | **🟡 ALTA** | **⏳ Pendente** | **DevOps/Infra** | **3 dias** |
| 10.1-10.2 | Data Retention | 🟢 MÉDIA | ⏳ Pendente | Infra/DevOps | 1 semana |
| 11.1-11.2 | Observabilidade | 🟢 MÉDIA | ⏳ Pendente | DevOps | 1 semana |
| 12.1-12.3 | Deployment | 🟡 ALTA | ⏳ Pendente | DevOps | 3 dias |
| 13.1-13.3 | Features Futuras | ⚪ BAIXA | ⏳ Pendente | Produto | - |

---

## 📝 Decisões Registradas (Decision Log)

### Decisão 001: Arquitetura de Monitoramento
**Data**: 2025-10-31
**Decisor**: Tech Lead
**Decisão**: Implementar monitoramento via Temporal Cron Workflow (não via Kubernetes CronJob)
**Razão**:
- ✅ Retry automático via Temporal
- ✅ Observabilidade built-in (Temporal UI)
- ✅ Histórico de execuções
- ✅ Continue-As-New para workflows longos
**Status**: ✅ APROVADO

---

### Decisão 002: Partitioning Strategy
**Data**: 2025-10-31
**Decisor**: DB & Domain Engineer
**Decisão**: Particionar `dict_rate_limit_states` por mês (RANGE partitioning)
**Razão**:
- ✅ Performance em queries temporais
- ✅ Drop de partições antigas é trivial (retenção 13 meses)
- ✅ Suporta crescimento indefinido
**Status**: ✅ APROVADO

---

### Decisão 003: [Template para próximas decisões]
**Data**: ____-__-__
**Decisor**: ____________
**Decisão**: ___________________________________________
**Razão**: ___________________________________________
**Status**: ⏳ Pendente / ✅ Aprovado / ❌ Rejeitado

---

## 🔗 Próximos Passos

1. **Enviar este documento para stakeholders** (Produto + Infra + Bridge Team)
2. **Agendar reunião de alinhamento** (2h, envolver todos os times)
3. **Preencher respostas** nas questões marcadas com [ ]
4. **Atualizar Decision Log** com decisões formais
5. **Revisar CLAUDE.md e SPECS** baseado nas respostas
6. **Iniciar Fase 1** após todas as questões CRÍTICAS resolvidas

---

## 📧 Contatos

| Time | Responsável | Email | Slack |
|------|-------------|-------|-------|
| Bridge | ___________ | _____ | #bridge-team |
| Core-Dict | ___________ | _____ | #core-dict-team |
| Infra/DevOps | ___________ | _____ | #devops |
| Produto | ___________ | _____ | #product |
| Security | ___________ | _____ | #security |

---

**Última Atualização**: 2025-11-01
**Versão**: 2.0.0 (+ 5 questões críticas sobre Token Bucket)
**Status**: 🔴 AGUARDANDO RESPOSTAS - **Total: 13 grupos de questões (41 sub-questões)**

## 📈 Resumo de Questões

- 🔴 **CRÍTICAS (Bloqueantes)**: 2 grupos (1-Bridge, 5-Categorização PSP)
- 🟡 **ALTAS (Importantes)**: 5 grupos (2-Core-Dict, 3-Thresholds, 4-Frequência, 6-Anti-Scan, 7-CIDS_FILES, 9-Clock, 12-Deployment)
- 🟢 **MÉDIAS**: 3 grupos (8-ETA, 10-Retention, 11-Observability)
- ⚪ **BAIXAS (Nice-to-have)**: 1 grupo (13-Features Futuras)
