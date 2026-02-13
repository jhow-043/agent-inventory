# Gestão de Mudanças

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  
> **Status:** Aprovado  
> **Classificação:** Uso Interno  

---

## 1. Objetivo

Definir o processo de gestão de mudanças para o Sistema de Inventário de Ativos de TI, garantindo que todas as alterações sejam controladas, documentadas e minimizem riscos ao serviço.

---

## 2. Escopo

Todas as mudanças no ambiente de produção: código, configuração, infraestrutura, banco de dados e agentes Windows.

---

## 3. Definições

| Termo | Definição |
|---|---|
| **Mudança** | Qualquer adição, modificação ou remoção que possa afetar o serviço de TI |
| **RFC** | Request for Change — solicitação formal de mudança |
| **CAB** | Change Advisory Board — corpo consultivo (adaptado para dev solo: self-review com checklist) |
| **PIR** | Post-Implementation Review — revisão pós-implementação |

---

## 4. Classificação de Mudanças

### 4.1 Tipos

| Tipo | Descrição | Aprovação | Exemplo |
|---|---|---|---|
| **Padrão** | Mudança de baixo risco, pré-aprovada, repetitiva | Pré-aprovada (checklist) | Atualização de dependência minor, deploy de patch |
| **Normal** | Mudança planejada com risco moderado | Self-review + checklist | Nova feature, migration de schema, upgrade major |
| **Emergencial** | Correção urgente de incidente em produção | Imediata (post-review) | Hotfix de bug crítico, patch de segurança |

### 4.2 Matriz de Impacto e Urgência

| | Urgência Alta | Urgência Média | Urgência Baixa |
|---|---|---|---|
| **Impacto Alto** | Emergencial | Normal (prioridade alta) | Normal (prioridade média) |
| **Impacto Médio** | Normal (prioridade alta) | Normal (prioridade média) | Padrão |
| **Impacto Baixo** | Normal (prioridade média) | Padrão | Padrão |

---

## 5. Processo de Mudança

### 5.1 Fluxo — Mudança Normal

```
1. Identificar necessidade
      │
2. Criar RFC (issue ou PR description)
      │
3. Avaliar impacto e risco (checklist)
      │
4. Planejar implementação + rollback
      │
5. Implementar em branch feature/*
      │
6. Testes (CI pipeline: lint + test + build)
      │
7. Self-review (PR checklist)
      │
8. Merge para develop
      │
9. Deploy em staging/teste (se aplicável)
      │
10. Deploy em produção
      │
11. Verificação pós-deploy (health check)
      │
12. PIR (se mudança de alto impacto)
```

### 5.2 Fluxo — Mudança Emergencial

```
1. Incidente detectado (P1/P2)
      │
2. Diagnóstico e identificação da correção
      │
3. Implementar hotfix em branch hotfix/*
      │
4. Testes mínimos (compilação + teste unitário afetado)
      │
5. Deploy direto em produção
      │
6. Verificação pós-deploy
      │
7. Post-mortem + retrospectiva
      │
8. Backport para develop
```

### 5.3 Fluxo — Mudança Padrão

```
1. Executar conforme procedimento pré-aprovado
2. Registrar a mudança (commit com mensagem convencional)
3. Verificar resultado
```

---

## 6. Template de RFC (Request for Change)

```markdown
## RFC: [Título da Mudança]

**Data:** YYYY-MM-DD
**Solicitante:** [Nome]
**Tipo:** Normal / Emergencial
**Prioridade:** Alta / Média / Baixa

### Descrição
[O que será mudado e por quê]

### Impacto
- **Componentes afetados:** [Agent / API / Dashboard / Banco / Infra]
- **Downtime estimado:** [Nenhum / X minutos]
- **Usuários afetados:** [Agents / Dashboard users / Ambos]

### Plano de Implementação
1. [Passo 1]
2. [Passo 2]
3. ...

### Plano de Rollback
1. [Passo para reverter se algo der errado]
2. ...

### Testes
- [ ] Testes unitários passam
- [ ] Testes de integração passam
- [ ] Teste manual em ambiente de desenvolvimento

### Riscos
| Risco | Probabilidade | Impacto | Mitigação |
|---|---|---|---|
| [Risco 1] | [Alta/Média/Baixa] | [Alto/Médio/Baixo] | [Ação] |

### Janela de Execução
- **Data planejada:** YYYY-MM-DD HH:MM
- **Duração estimada:** X minutos
- **Responsável pela execução:** [Nome]
```

---

## 7. Checklist de Self-Review (PR Review)

### 7.1 Qualidade de Código

- [ ] Código segue os padrões do projeto (golangci-lint / eslint sem erros)
- [ ] Nomes de variáveis, funções e arquivos em inglês
- [ ] Comentários em inglês
- [ ] Sem secrets ou credenciais hardcoded
- [ ] Sem `TODO` ou `FIXME` não rastreados

### 7.2 Testes

- [ ] Testes unitários adicionados/atualizados para o código novo
- [ ] CI pipeline passa (lint + test + build)
- [ ] Cobertura de testes não diminuiu

### 7.3 Segurança

- [ ] Input validation em novos endpoints
- [ ] Autenticação/autorização verificada em novos endpoints
- [ ] Sem logging de dados sensíveis (tokens, senhas)

### 7.4 Banco de Dados

- [ ] Migration up E down criadas (se aplicável)
- [ ] Migration testada em ambiente limpo
- [ ] Índices adicionados para queries novas (se necessário)
- [ ] Sem breaking changes no schema sem migration

### 7.5 Deploy

- [ ] docker-compose.yml atualizado (se necessário)
- [ ] Variáveis de ambiente documentadas (se novas)
- [ ] Rollback plan definido
- [ ] Documentação atualizada (se comportamento mudou)

---

## 8. Controle de Mudanças de Banco de Dados

### 8.1 Regras para Migrations

| Regra | Descrição |
|---|---|
| Sempre criar UP e DOWN | Toda migration deve ter reversão |
| Nunca editar migration aplicada | Criar nova migration para correção |
| Testar em ambiente limpo | Rodar todas as migrations do zero |
| Uma mudança lógica por migration | Não misturar concerns |
| Nomenclatura sequencial | `000001_create_users.up.sql` |
| Idempotência | Usar `IF NOT EXISTS` quando possível |

### 8.2 Mudanças de Alto Risco no Banco

| Mudança | Risco | Procedimento |
|---|---|---|
| ALTER TABLE (campo NOT NULL) | Alto | Adicionar com default → popular → remover default |
| DROP TABLE / DROP COLUMN | Crítico | Backup antes; migration down funcional |
| Mudança de tipo de campo | Alto | Nova coluna → migrar dados → remover antiga |
| Adição de índice em tabela grande | Médio | `CREATE INDEX CONCURRENTLY` |

---

## 9. Versionamento e Branches

### 9.1 Estratégia de Branching (Git Flow Simplificado)

```
main ─────────────────────────────────→  (sempre deployable)
  │                                  ▲
  └── develop ──────────────────────▶│  (integração)
        │         │         │        │
        └─ feature/xyz     │        │
                  └─ feature/abc    │
                           └─ hotfix/fix-123
```

| Branch | Propósito | Merge para | Proteção |
|---|---|---|---|
| `main` | Código em produção | — | Protected, merge via PR |
| `develop` | Integração de features | `main` (via PR) | — |
| `feature/*` | Desenvolvimento de feature | `develop` | — |
| `hotfix/*` | Correção urgente | `main` + `develop` | — |

### 9.2 Versionamento Semântico

```
MAJOR.MINOR.PATCH

Exemplos:
v1.0.0 — Primeira release de produção
v1.1.0 — Nova feature (ex: filtro adicional no dashboard)
v1.1.1 — Bugfix (ex: correção de query SQL)
v2.0.0 — Breaking change (ex: mudança de schema incompatível)
```

### 9.3 Commits Convencionais

```
feat: add device filtering by OS version
fix: correct disk size calculation for NVMe drives
docs: update API endpoint documentation
refactor: extract inventory upsert to separate service
test: add integration tests for device repository
chore: update Go dependencies
ci: add govulncheck to CI pipeline
```

---

## 10. Changelog

Changelog gerado automaticamente a partir dos commits convencionais. Formato:

```markdown
# Changelog

## [1.1.0] - 2026-XX-XX
### Added
- Device filtering by OS version (#12)
- Export device list as CSV (#15)

### Fixed
- Disk size calculation for NVMe drives (#14)

## [1.0.0] - 2026-XX-XX
### Added
- Initial release: Agent, API, Dashboard
```

---

## 11. Responsáveis

| Papel | Responsável | Atribuição |
|---|---|---|
| Solicitante da mudança | Desenvolvedor / Gestor de TI | Criar RFC |
| Avaliador | Desenvolvedor (self-review) | Avaliar impacto, aprovar |
| Implementador | Desenvolvedor | Executar a mudança |
| Validador pós-deploy | Administrador de TI | Confirmar que serviço está operacional |

---

## 12. Referências

- [Gestão de Liberação e Implantação](gestao-de-liberacao-e-implantacao.md)
- [Gestão de Configuração e Ativos](gestao-de-configuracao-e-ativos.md)
- [Gestão de Incidentes](../04-operacao-de-servico/gestao-de-incidentes.md)

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial |
