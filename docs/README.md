# Documentação ITIL v4 — Sistema de Inventário de Ativos de TI

> **Versão:** 1.0.0  
> **Data:** Fevereiro de 2026  
> **Status:** Fase 1 — Inventário de Ativos Windows  
> **Classificação:** Uso Interno  

---

## Sobre este Documento

Esta documentação segue o framework **ITIL v4** (Information Technology Infrastructure Library) e cobre todo o ciclo de vida do serviço "Sistema de Inventário de Ativos de TI". Todos os documentos estão em **português brasileiro** e são versionados junto ao código-fonte no repositório Git.

### Componentes do Sistema

| Componente | Tecnologia | Descrição |
|---|---|---|
| **Windows Agent** | Go | Agente de coleta instalado como Windows Service |
| **API Central** | Go (Gin) | API REST que recebe e serve dados de inventário |
| **Banco de Dados** | PostgreSQL | Armazenamento persistente dos dados de inventário |
| **Dashboard Web** | React + TypeScript | Interface web para visualização dos ativos |

### Comunicação (Fase 1)

Na Fase 1, toda a comunicação ocorre via **HTTP** (sem TLS). O roadmap de migração para HTTPS está documentado em [Gestão de Segurança](02-desenho-de-servico/gestao-de-seguranca.md).

---

## Índice Geral

### 01 — Estratégia de Serviço

| # | Documento | Descrição |
|---|---|---|
| 1 | [Visão Geral do Serviço](01-estrategia-de-servico/visao-geral-do-servico.md) | Proposta de valor, escopo e público-alvo |
| 2 | [Catálogo de Serviços](01-estrategia-de-servico/catalogo-de-servicos.md) | Serviços oferecidos e seus atributos |
| 3 | [Análise Financeira](01-estrategia-de-servico/analise-financeira.md) | TCO, custos e comparação com alternativas |
| 4 | [Gestão de Demanda](01-estrategia-de-servico/gestao-de-demanda.md) | Capacidade planejada e crescimento projetado |

### 02 — Desenho de Serviço

| # | Documento | Descrição |
|---|---|---|
| 5 | [Arquitetura da Solução](02-desenho-de-servico/arquitetura-da-solucao.md) | Arquitetura técnica, stack e decisões |
| 6 | [Requisitos de Nível de Serviço](02-desenho-de-servico/requisitos-de-nivel-de-servico.md) | SLI, SLO e SLA definidos |
| 7 | [Gestão de Capacidade](02-desenho-de-servico/gestao-de-capacidade.md) | Sizing, limites e escalabilidade |
| 8 | [Gestão de Disponibilidade](02-desenho-de-servico/gestao-de-disponibilidade.md) | Uptime, SPOFs e recuperação |
| 9 | [Gestão de Continuidade](02-desenho-de-servico/gestao-de-continuidade.md) | Disaster recovery, backup, RPO/RTO |
| 10 | [Gestão de Segurança](02-desenho-de-servico/gestao-de-seguranca.md) | Políticas de segurança e roadmap HTTP→HTTPS |
| 11 | [Gestão de Fornecedores](02-desenho-de-servico/gestao-de-fornecedores.md) | Dependências externas e licenciamento |

### 03 — Transição de Serviço

| # | Documento | Descrição |
|---|---|---|
| 12 | [Gestão de Mudanças](03-transicao-de-servico/gestao-de-mudancas.md) | Processo de change management |
| 13 | [Gestão de Configuração e Ativos](03-transicao-de-servico/gestao-de-configuracao-e-ativos.md) | CMDB e configuration items |
| 14 | [Gestão de Liberação e Implantação](03-transicao-de-servico/gestao-de-liberacao-e-implantacao.md) | Release management e deploy |
| 15 | [Validação e Testes](03-transicao-de-servico/validacao-e-testes.md) | Estratégia de testes completa |
| 16 | [Gestão de Conhecimento](03-transicao-de-servico/gestao-de-conhecimento.md) | Base de conhecimento e onboarding |

### 04 — Operação de Serviço

| # | Documento | Descrição |
|---|---|---|
| 17 | [Gestão de Incidentes](04-operacao-de-servico/gestao-de-incidentes.md) | Incident management e severidades |
| 18 | [Gestão de Problemas](04-operacao-de-servico/gestao-de-problemas.md) | Root cause analysis e known errors |
| 19 | [Gestão de Eventos](04-operacao-de-servico/gestao-de-eventos.md) | Monitoramento, logs e alertas |
| 20 | [Cumprimento de Requisições](04-operacao-de-servico/cumprimento-de-requisicoes.md) | Service requests e procedimentos |
| 21 | [Runbooks Operacionais](04-operacao-de-servico/runbooks-operacionais.md) | Procedimentos operacionais detalhados |

### 05 — Melhoria Contínua

| # | Documento | Descrição |
|---|---|---|
| 22 | [Plano de Melhoria](05-melhoria-continua/plano-de-melhoria.md) | Backlog de melhorias e priorização |
| 23 | [Métricas e KPIs](05-melhoria-continua/metricas-e-kpis.md) | Indicadores de desempenho |

### 06 — Anexos

| # | Documento | Descrição |
|---|---|---|
| 24 | [Glossário](06-anexos/glossario.md) | Termos ITIL e técnicos |
| 25 | [Matriz RACI](06-anexos/matriz-raci.md) | Responsabilidades por processo |
| 26 | [Diagrama — Arquitetura Geral](06-anexos/diagramas/arquitetura-geral.md) | Visão macro dos componentes |
| 27 | [Diagrama — Fluxo de Comunicação](06-anexos/diagramas/fluxo-de-comunicacao.md) | Sequência Agent → API → DB |
| 28 | [Diagrama — Fluxo de Autenticação](06-anexos/diagramas/fluxo-de-autenticacao.md) | Enrollment, token e JWT |
| 29 | [Diagrama — Modelo de Dados](06-anexos/diagramas/modelo-de-dados.md) | Diagrama ER do banco de dados |
| 30 | [Diagrama — Fluxo de Deploy](06-anexos/diagramas/fluxo-de-deploy.md) | Pipeline CI/CD |

---

## Convenções

- **Formato:** Markdown (.md), versionado no Git
- **Idioma:** Português brasileiro
- **Diagramas:** Mermaid (renderizável no GitHub/VS Code)
- **Template:** Cada documento segue a estrutura: Objetivo → Escopo → Definições → Processo/Conteúdo → Responsáveis → Referências
- **Versionamento:** Cada documento possui cabeçalho com versão, data e status
- **Revisão:** Documentos são revisados a cada release major do sistema

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Criação inicial da documentação ITIL v4 completa |

---

> **Referência ITIL v4:** Esta documentação está alinhada com os 5 estágios do ciclo de vida do serviço ITIL: Estratégia, Desenho, Transição, Operação e Melhoria Contínua do Serviço.
