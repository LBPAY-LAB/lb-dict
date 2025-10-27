# Changelog

Todas as mudancas notaveis neste projeto serao documentadas neste arquivo.

O formato e baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.0.0/),
e este projeto adere ao [Semantic Versioning](https://semver.org/lang/pt-BR/).

## [Unreleased]

### Added
- Setup inicial do projeto conn-dict
- Estrutura base de diretorios (cmd, internal, pkg, test)
- Configuracao Docker Compose com Temporal, Pulsar, Redis
- Dockerfile multi-stage para Connect API e Worker
- Makefile com targets de build, test, lint, run
- Configuracao OpenTelemetry Collector
- Configuracao Temporal dynamic config
- Arquivos de configuracao (.env.example, .gitignore, .dockerignore)
- README.md com documentacao inicial
- go.mod com dependencias principais (Temporal v1.36.0, Pulsar v0.16.0)

### Changed
- N/A

### Deprecated
- N/A

### Removed
- N/A

### Fixed
- N/A

### Security
- N/A

---

## [1.0.0] - TBD

### Added
- Implementacao inicial do RSFN Connect
- ClaimWorkflow (30 dias de monitoramento)
- Integracao com Temporal Server
- Integracao com Apache Pulsar (consumer/producer)
- Integracao com Redis (cache)
- Cliente gRPC para Bridge
- Observabilidade com OpenTelemetry

---

**Legenda**:
- `Added` para novas funcionalidades
- `Changed` para mudancas em funcionalidades existentes
- `Deprecated` para funcionalidades que serao removidas
- `Removed` para funcionalidades removidas
- `Fixed` para correcoes de bugs
- `Security` para correcoes de vulnerabilidades
