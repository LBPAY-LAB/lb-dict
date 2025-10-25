# Checklists

**Propósito**: Checklists padronizados para garantir qualidade e completude em entregas

## 📋 Conteúdo

Esta pasta armazenará:

- **Definition of Done (DoD)**: Critérios para considerar uma tarefa completa
- **Definition of Ready (DoR)**: Critérios para uma user story entrar em sprint
- **Code Review Checklist**: Checklist de revisão de código
- **Deployment Checklist**: Passos para deploy seguro em produção
- **Security Checklist**: Verificações de segurança obrigatórias

## 📁 Estrutura Esperada

```
Checklists/
├── DoD_Definition_of_Done.md
├── DoR_Definition_of_Ready.md
├── Code_Review_Checklist.md
├── Deployment_Checklist.md
├── Security_Checklist.md
└── Testing_Checklist.md
```

## 🎯 Exemplos de Checklists

### Definition of Done (DoD)
- [ ] Código desenvolvido e commitado
- [ ] Testes unitários escritos (cobertura > 80%)
- [ ] Code review aprovado por 2+ desenvolvedores
- [ ] Documentação atualizada
- [ ] Pipeline CI/CD passou (build + testes)
- [ ] Deploy em staging validado
- [ ] Aprovação do PO (Product Owner)

### Security Checklist
- [ ] Secrets não commitados em código
- [ ] Input validation implementada
- [ ] SQL injection prevenido (prepared statements)
- [ ] Autenticação e autorização validadas
- [ ] Logs de auditoria implementados

## 📚 Referências

- [Testes](../../14_Testes/)
- [DevOps](../../15_DevOps/)
- [Segurança](../../13_Seguranca/)

---

**Status**: 🔴 Pasta vazia (será preenchida durante setup de processos)
**Fase de Preenchimento**: Fase 2 (início do desenvolvimento)
