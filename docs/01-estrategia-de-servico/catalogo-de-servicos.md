# Catálogo de Serviços

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  
> **Status:** Aprovado  
> **Classificação:** Uso Interno  

---

## 1. Objetivo

Documentar formalmente todos os serviços oferecidos pelo Sistema de Inventário de Ativos de TI, seus atributos, dependências e nível de suporte.

---

## 2. Escopo

Este catálogo cobre todos os serviços disponíveis na Fase 1 e indica serviços planejados para fases futuras.

---

## 3. Catálogo de Serviços de Negócio

### SVC-INV-001 — Inventário Automatizado de Ativos Windows

| Atributo | Valor |
|---|---|
| **ID** | SVC-INV-001 |
| **Nome** | Inventário Automatizado de Ativos Windows |
| **Descrição** | Coleta automática e periódica de dados de hardware, software, rede e licenciamento de computadores Windows por meio de um agente instalado como serviço |
| **Tipo** | Serviço de negócio |
| **Status** | Em desenvolvimento |
| **Criticidade** | Média |
| **Disponibilidade alvo** | 99.5% (ver [Requisitos de Nível de Serviço](../02-desenho-de-servico/requisitos-de-nivel-de-servico.md)) |
| **Horário de operação** | 24×7 (coleta contínua) |
| **Usuários** | Todas as estações Windows gerenciadas (100–500 dispositivos) |
| **Proprietário** | Equipe de Infraestrutura de TI |
| **Canal de entrega** | Agente Windows instalado como serviço |

#### Sub-serviços

| ID | Sub-serviço | Dados coletados |
|---|---|---|
| SVC-INV-001.1 | Coleta de informações do sistema | Hostname, versão Windows, build, serial, último boot, usuário logado |
| SVC-INV-001.2 | Coleta de hardware | CPU (modelo, cores), RAM, discos (modelo, tamanho, tipo), placa-mãe, BIOS |
| SVC-INV-001.3 | Coleta de rede | Interfaces ativas, endereço IP, endereço MAC |
| SVC-INV-001.4 | Coleta de software | Programas instalados (nome, versão, fabricante) |
| SVC-INV-001.5 | Coleta de licenciamento | Status de ativação do Windows |

---

### SVC-INV-002 — Dashboard de Visualização de Ativos

| Atributo | Valor |
|---|---|
| **ID** | SVC-INV-002 |
| **Nome** | Dashboard de Visualização de Ativos |
| **Descrição** | Interface web para consulta, filtro e visualização dos dados de inventário coletados pelos agentes |
| **Tipo** | Serviço de negócio |
| **Status** | Em desenvolvimento |
| **Criticidade** | Média |
| **Disponibilidade alvo** | 99.5% |
| **Horário de operação** | Horário comercial (uso primário), 24×7 (disponível) |
| **Usuários** | Administradores de TI, gestores, auditores |
| **Proprietário** | Equipe de Infraestrutura de TI |
| **Canal de entrega** | Aplicação web (browser) |

#### Funcionalidades

| ID | Funcionalidade | Descrição |
|---|---|---|
| SVC-INV-002.1 | Login autenticado | Acesso protegido por usuário e senha com JWT |
| SVC-INV-002.2 | Painel resumo (Dashboard) | Total de dispositivos, contagem online/offline |
| SVC-INV-002.3 | Lista de dispositivos | Tabela com filtros por hostname e sistema operacional |
| SVC-INV-002.4 | Detalhe do dispositivo | Visualização completa: sistema, hardware, discos, rede, software |

---

## 4. Catálogo de Serviços de Suporte (Técnicos)

### SVC-SUP-001 — API Central de Inventário

| Atributo | Valor |
|---|---|
| **ID** | SVC-SUP-001 |
| **Nome** | API Central de Inventário |
| **Descrição** | API REST que recebe dados dos agentes, armazena no banco de dados e serve informações ao dashboard |
| **Tipo** | Serviço de suporte técnico |
| **Status** | Em desenvolvimento |
| **Tecnologia** | Go (Gin), PostgreSQL |
| **Protocolo** | HTTP/JSON (Fase 1) |
| **Autenticação** | Token de dispositivo (agentes), JWT (dashboard) |
| **Deploy** | Container Docker |
| **Depende de** | PostgreSQL (SVC-SUP-002) |

#### Endpoints (Fase 1)

| Método | Endpoint | Autenticação | Descrição |
|---|---|---|---|
| POST | `/api/v1/register` | Enrollment Key | Registra novo dispositivo |
| POST | `/api/v1/inventory` | Device Token | Recebe snapshot de inventário |
| POST | `/api/v1/auth/login` | Pública | Login de usuário do dashboard |
| POST | `/api/v1/auth/refresh` | Refresh Token | Renova access token |
| GET | `/api/v1/devices` | JWT | Lista dispositivos (com filtros) |
| GET | `/api/v1/devices/:id` | JWT | Detalhe de um dispositivo |
| GET | `/api/v1/dashboard/stats` | JWT | Estatísticas do dashboard |
| GET | `/healthz` | Pública | Verificação de saúde (liveness) |
| GET | `/readyz` | Pública | Verificação de prontidão (readiness) |

---

### SVC-SUP-002 — Banco de Dados PostgreSQL

| Atributo | Valor |
|---|---|
| **ID** | SVC-SUP-002 |
| **Nome** | Banco de Dados PostgreSQL |
| **Descrição** | Armazenamento persistente e relacional de todos os dados de inventário |
| **Tipo** | Serviço de suporte técnico |
| **Status** | Em desenvolvimento |
| **Tecnologia** | PostgreSQL 16+ |
| **Deploy** | Container Docker com volume persistente |
| **Dependentes** | API Central (SVC-SUP-001) |

---

## 5. Mapa de Dependências entre Serviços

```
SVC-INV-001 (Agent)
       │
       │  HTTP POST (inventory)
       ▼
SVC-SUP-001 (API) ──────→ SVC-SUP-002 (PostgreSQL)
       ▲
       │  HTTP GET (devices, stats)
       │
SVC-INV-002 (Dashboard)
```

---

## 6. Serviços Planejados (Fases Futuras)

| ID | Serviço | Fase | Status |
|---|---|---|---|
| SVC-INV-003 | Inventário de Ativos Linux | Fase 2+ | Planejado |
| SVC-INV-004 | Histórico de Mudanças de Ativos | Fase 2+ | Planejado |
| SVC-INV-005 | Relatórios Exportáveis | Fase 2+ | Planejado |
| SVC-INV-006 | Integração com ITSM (GLPI, ServiceNow) | Fase 3+ | Planejado |
| SVC-SUP-003 | Monitoramento com Prometheus + Grafana | Fase 2+ | Planejado |
| SVC-SUP-004 | Reverse Proxy com TLS (HTTPS) | Fase 2 | Planejado |

---

## 7. Níveis de Suporte

| Nível | Responsável | Escopo |
|---|---|---|
| **N1 — Suporte básico** | Administrador de TI | Verificação de status, reinício de serviços, instalação de agentes |
| **N2 — Suporte técnico** | Desenvolvedor | Troubleshooting da API, análise de logs, correção de bugs |
| **N3 — Suporte especializado** | Desenvolvedor | Alterações de código, redesign de componentes, migrações de schema |

---

## 8. Responsáveis

| Papel | Responsável |
|---|---|
| Proprietário do catálogo | Gestor de TI |
| Mantenedor | Desenvolvedor |
| Revisão periódica | A cada release major |

---

## 9. Referências

- [Visão Geral do Serviço](visao-geral-do-servico.md)
- [Requisitos de Nível de Serviço](../02-desenho-de-servico/requisitos-de-nivel-de-servico.md)
- [Arquitetura da Solução](../02-desenho-de-servico/arquitetura-da-solucao.md)

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial |
