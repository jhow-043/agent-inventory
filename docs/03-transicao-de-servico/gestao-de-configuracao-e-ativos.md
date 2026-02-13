# Gestão de Configuração e Ativos (CMDB)

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  
> **Status:** Aprovado  
> **Classificação:** Uso Interno  

---

## 1. Objetivo

Manter um registro formal de todos os Configuration Items (CIs) do Sistema de Inventário de Ativos de TI, suas relações, baselines e controle de versionamento.

---

## 2. Escopo

Todos os componentes de infraestrutura, software e configuração que compõem o serviço em produção.

---

## 3. Definições

| Termo | Definição |
|---|---|
| **CI** (Configuration Item) | Qualquer componente que precisa ser gerenciado para entregar o serviço |
| **CMDB** (Configuration Management Database) | Registro centralizado de todos os CIs e suas relações |
| **Baseline** | Configuração aprovada de um CI em um ponto específico no tempo |
| **Relacionamento** | Dependência ou associação entre CIs |

---

## 4. Registro de Configuration Items

### CI-001: Windows Agent

| Atributo | Valor |
|---|---|
| **ID** | CI-001 |
| **Nome** | Windows Agent (InventoryAgent) |
| **Tipo** | Software — Aplicação |
| **Proprietário** | Desenvolvedor |
| **Localização** | Estações Windows gerenciadas |
| **Tecnologia** | Go, binário compilado (.exe) |
| **Versão atual** | — (em desenvolvimento) |
| **Versionamento** | SemVer, tag Git |
| **Configuração** | `agent-config.yaml` em `C:\ProgramData\InventoryAgent\` |
| **Dependências** | CI-003 (API) para envio de dados |
| **Criticidade** | Alta |
| **Backup necessário** | Não (binário reproduzível via CI/CD) |
| **Artefatos** | `agent.exe`, `agent-config.yaml` |

### CI-002: API Server

| Atributo | Valor |
|---|---|
| **ID** | CI-002 |
| **Nome** | API Central (inventory-api) |
| **Tipo** | Software — Container Docker |
| **Proprietário** | Desenvolvedor |
| **Localização** | Servidor Docker host |
| **Tecnologia** | Go (Gin), container Docker |
| **Versão atual** | — (em desenvolvimento) |
| **Versionamento** | SemVer, tag Git, image tag |
| **Configuração** | Variáveis de ambiente (`.env`) |
| **Dependências** | CI-004 (PostgreSQL) |
| **Dependentes** | CI-001 (Agent), CI-003 (Dashboard) |
| **Criticidade** | Alta |
| **Backup necessário** | Não (imagem Docker reproduzível) |
| **Porta** | 8080 |

### CI-003: Dashboard Web

| Atributo | Valor |
|---|---|
| **ID** | CI-003 |
| **Nome** | Dashboard Web (inventory-web) |
| **Tipo** | Software — Container Docker |
| **Proprietário** | Desenvolvedor |
| **Localização** | Servidor Docker host |
| **Tecnologia** | React + TypeScript, Nginx, container Docker |
| **Versão atual** | — (em desenvolvimento) |
| **Versionamento** | SemVer, tag Git, image tag |
| **Configuração** | Variáveis de ambiente no build |
| **Dependências** | CI-002 (API) |
| **Criticidade** | Média |
| **Backup necessário** | Não (imagem Docker reproduzível) |
| **Porta** | 3000 |

### CI-004: PostgreSQL

| Atributo | Valor |
|---|---|
| **ID** | CI-004 |
| **Nome** | Banco de Dados PostgreSQL |
| **Tipo** | Software — Container Docker |
| **Proprietário** | Administrador de TI |
| **Localização** | Servidor Docker host |
| **Tecnologia** | PostgreSQL 16, container Docker |
| **Versão atual** | 16.x |
| **Configuração** | Variáveis de ambiente + `postgresql.conf` (se customizado) |
| **Dependentes** | CI-002 (API) |
| **Criticidade** | Crítica |
| **Backup necessário** | **Sim** — contém todos os dados de inventário e tokens |
| **Frequência de backup** | Diário (pg_dump) |
| **Volume Docker** | `inventory-postgres-data` |
| **Porta** | 5432 (interna Docker, não exposta) |

### CI-005: Docker Compose

| Atributo | Valor |
|---|---|
| **ID** | CI-005 |
| **Nome** | Configuração Docker Compose |
| **Tipo** | Configuração — Infraestrutura como Código |
| **Proprietário** | Desenvolvedor |
| **Localização** | Repositório Git |
| **Arquivos** | `docker-compose.yml`, `docker-compose.dev.yml` |
| **Versionamento** | Git |
| **Dependentes** | CI-002, CI-003, CI-004 |
| **Criticidade** | Alta |
| **Backup necessário** | Não (versionado no Git) |

### CI-006: Migrations SQL

| Atributo | Valor |
|---|---|
| **ID** | CI-006 |
| **Nome** | Database Migrations |
| **Tipo** | Configuração — Schema de banco |
| **Proprietário** | Desenvolvedor |
| **Localização** | `server/migrations/` no repositório Git |
| **Versionamento** | Git + sequencial numérico |
| **Dependentes** | CI-004 (PostgreSQL) |
| **Criticidade** | Alta |
| **Regra** | Nunca editar migrations já aplicadas em produção |

### CI-007: Pipeline CI/CD

| Atributo | Valor |
|---|---|
| **ID** | CI-007 |
| **Nome** | GitHub Actions CI Pipeline |
| **Tipo** | Configuração — Automação |
| **Proprietário** | Desenvolvedor |
| **Localização** | `.github/workflows/` no repositório Git |
| **Versionamento** | Git |
| **Criticidade** | Média |

### CI-008: Secrets de Produção

| Atributo | Valor |
|---|---|
| **ID** | CI-008 |
| **Nome** | Secrets e Credenciais de Produção |
| **Tipo** | Configuração — Segurança |
| **Proprietário** | Administrador de TI |
| **Localização** | `.env` no servidor de produção (NÃO no Git) |
| **Conteúdo** | JWT_SECRET, DATABASE_URL, ENROLLMENT_KEY |
| **Criticidade** | Crítica |
| **Backup necessário** | Sim (armazenamento seguro offline) |
| **Rotação** | A cada release major ou após incidente |

---

## 5. Mapa de Relacionamentos entre CIs

```
CI-001 (Agent)
    │
    │  HTTP POST (inventory)
    ▼
CI-002 (API) ────────→ CI-004 (PostgreSQL)
    ▲                       ▲
    │  HTTP GET              │ Schema
    │                        │
CI-003 (Dashboard)     CI-006 (Migrations)

CI-005 (Docker Compose) ──→ CI-002, CI-003, CI-004
CI-007 (CI/CD Pipeline) ──→ CI-001, CI-002, CI-003
CI-008 (Secrets) ──→ CI-002, CI-004
```

### Tabela de Dependências

| CI | Depende de | É dependência de |
|---|---|---|
| CI-001 (Agent) | CI-002 (API) | — |
| CI-002 (API) | CI-004 (PostgreSQL), CI-008 (Secrets) | CI-001 (Agent), CI-003 (Dashboard) |
| CI-003 (Dashboard) | CI-002 (API) | — |
| CI-004 (PostgreSQL) | CI-006 (Migrations) | CI-002 (API) |
| CI-005 (Docker Compose) | — | CI-002, CI-003, CI-004 |
| CI-006 (Migrations) | — | CI-004 (PostgreSQL) |
| CI-007 (CI/CD) | — | CI-001, CI-002, CI-003 |
| CI-008 (Secrets) | — | CI-002, CI-004 |

---

## 6. Baselines de Configuração

### 6.1 Baseline de Desenvolvimento

| CI | Versão | Configuração |
|---|---|---|
| Agent | HEAD de develop | `api_url: http://localhost:8080` |
| API | HEAD de develop | `DATABASE_URL=postgres://inventory:dev@localhost:5432/inventory` |
| Dashboard | HEAD de develop | `VITE_API_URL=http://localhost:8080` |
| PostgreSQL | 16.x | Docker default + compose dev overrides |
| Docker Compose | docker-compose.dev.yml | Volumes de dev, hot-reload |

### 6.2 Baseline de Produção

| CI | Versão | Configuração |
|---|---|---|
| Agent | Tag release (ex: v1.0.0) | `api_url: http://<server-ip>:8080` |
| API | Tag release (ex: v1.0.0) | Env vars securizadas em `.env` |
| Dashboard | Tag release (ex: v1.0.0) | Build com API URL de produção |
| PostgreSQL | 16.x | Tuning conforme [Gestão de Capacidade](../02-desenho-de-servico/gestao-de-capacidade.md) |
| Docker Compose | docker-compose.yml | Restart policies, limits, volumes persistentes |

---

## 7. Processo de Controle de Configuração

### 7.1 Identificação

Todo novo CI deve ser registrado neste documento com:
- ID único
- Tipo e descrição
- Proprietário
- Criticidade
- Dependências

### 7.2 Controle

| Ação | Procedimento |
|---|---|
| **Adicionar CI** | Criar entrada neste documento + PR |
| **Alterar CI** | RFC + atualizar baseline |
| **Remover CI** | RFC + atualizar mapa de dependências |
| **Upgrade de versão** | Processo de mudança normal |

### 7.3 Auditoria

| Atividade | Frequência |
|---|---|
| Verificar que CIs de produção correspondem ao baseline | Mensal |
| Verificar que secrets não estão no Git | A cada PR (CI automático) |
| Verificar integridade dos backups | Semanal |
| Revisão completa do CMDB | Trimestral |

---

## 8. Responsáveis

| Papel | Responsável | Atribuição |
|---|---|---|
| Gerente de configuração | Desenvolvedor | Manter CMDB atualizado |
| Proprietário de CIs de infra | Administrador de TI | CI-004, CI-008 |
| Proprietário de CIs de código | Desenvolvedor | CI-001, CI-002, CI-003, CI-005, CI-006, CI-007 |

---

## 9. Referências

- [Gestão de Mudanças](gestao-de-mudancas.md)
- [Gestão de Liberação e Implantação](gestao-de-liberacao-e-implantacao.md)
- [Gestão de Segurança](../02-desenho-de-servico/gestao-de-seguranca.md)
- [Gestão de Continuidade](../02-desenho-de-servico/gestao-de-continuidade.md)

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial |
