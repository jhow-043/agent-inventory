# Visão Geral do Serviço

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  
> **Status:** Aprovado  
> **Classificação:** Uso Interno  

---

## 1. Objetivo

Documentar a visão geral do serviço "Sistema de Inventário de Ativos de TI", descrevendo sua proposta de valor, escopo, público-alvo e modelo de entrega.

---

## 2. Escopo

Este documento abrange o serviço em sua totalidade na **Fase 1**, limitada à coleta e visualização de inventário de ativos Windows.

---

## 3. Identificação do Serviço

| Atributo | Valor |
|---|---|
| **Nome do Serviço** | Sistema de Inventário de Ativos de TI |
| **ID do Serviço** | SVC-INV-001 |
| **Proprietário do Serviço** | Equipe de Infraestrutura de TI |
| **Fase Atual** | Fase 1 — Inventário de Ativos Windows |
| **Status** | Em Desenvolvimento |
| **Criticidade** | Média |
| **Tipo de Serviço** | Serviço Interno de TI |

---

## 4. Proposta de Valor

### 4.1 Problema

Organizações com parques Windows de 100 a 500 máquinas frequentemente não possuem visibilidade centralizada sobre:
- Quais ativos de hardware existem na rede
- Qual software está instalado em cada máquina
- Qual o status de licenciamento do Windows
- Qual a configuração de rede de cada dispositivo

Essa falta de visibilidade gera:
- Dificuldade em planejamento de capacidade
- Risco de não-conformidade de software
- Tempo excessivo em auditorias manuais
- Impossibilidade de tomar decisões baseadas em dados sobre o parque

### 4.2 Solução

Um sistema automatizado de inventário composto por:
1. **Agente Windows** que coleta dados de hardware, software, rede e licenciamento
2. **API Central** que recebe, armazena e serve os dados
3. **Dashboard Web** que permite visualização e consulta dos ativos

### 4.3 Benefícios Esperados

| Benefício | Descrição |
|---|---|
| **Visibilidade centralizada** | Todos os ativos visíveis em um único painel |
| **Automação da coleta** | Eliminação de inventários manuais (planilhas, scripts ad-hoc) |
| **Dados atualizados** | Atualização periódica automática (configurável) |
| **Tomada de decisão** | Dados concretos para planejamento de compras, upgrades, auditorias |
| **Rastreabilidade** | Identificação de cada dispositivo por serial number |
| **Custo zero de licenciamento** | Stack 100% open-source |

---

## 5. Escopo da Fase 1

### 5.1 Incluído

| Categoria | Dados Coletados |
|---|---|
| **Sistema** | Hostname, versão do Windows, build, serial number, último boot, usuário logado |
| **Hardware** | CPU (modelo, cores), RAM total, discos (modelo, tamanho, tipo), placa-mãe, BIOS |
| **Rede** | Interfaces ativas, endereço IP, endereço MAC |
| **Software** | Programas instalados (nome, versão, fabricante) |
| **Licenciamento** | Status de ativação do Windows |

### 5.2 Excluído (Fases Futuras)

- Suporte multi-tenant
- Execução remota de comandos
- Scanner de rede (discovery ativo)
- Histórico de mudanças por dispositivo
- Agente Linux/macOS
- Integrações externas (GLPI, ServiceNow, etc.)
- Funcionalidades de RMM (Remote Monitoring & Management)
- Monitoramento de desempenho em tempo real

---

## 6. Público-Alvo

### 6.1 Usuários Primários

| Persona | Papel | Uso do Sistema |
|---|---|---|
| **Administrador de TI** | Gerencia o parque de máquinas | Consulta dashboard, gera relatórios informais, identifica problemas |
| **Gestor de TI** | Toma decisões sobre o parque | Visualiza totais, status de ativos, planeja compras |

### 6.2 Usuários Secundários

| Persona | Papel | Uso do Sistema |
|---|---|---|
| **Auditoria** | Verifica conformidade | Consulta software instalado, status de licenças |
| **Suporte Técnico** | Atende chamados | Consulta configuração do dispositivo do usuário |

---

## 7. Modelo de Entrega

### 7.1 Arquitetura de Alto Nível

```
┌─────────────────┐       HTTP/JSON        ┌──────────────────┐
│  Windows Agent   │ ────────────────────→  │    API Central   │
│  (Go, Service)   │   Token de Dispositivo │    (Go, Gin)     │
└─────────────────┘                         │                  │
                                            │  ┌────────────┐  │
┌─────────────────┐       HTTP/JSON         │  │  Handlers   │  │
│  Dashboard Web   │ ────────────────────→  │  │  Services   │  │
│  (React + TS)    │   JWT                  │  │  Repos      │  │
└─────────────────┘                         │  └──────┬─────┘  │
                                            └─────────┼────────┘
                                                      │
                                               ┌──────▼──────┐
                                               │  PostgreSQL  │
                                               └─────────────┘
```

### 7.2 Modelo de Deploy

| Aspecto | Detalhe |
|---|---|
| **Tipo** | On-premises |
| **Containerização** | Docker Compose (API + PostgreSQL + Dashboard) |
| **Agent** | Binário .exe instalado como Windows Service |
| **Comunicação** | HTTP (Fase 1), HTTPS planejado para Fase 2 |
| **Infraestrutura** | Servidor único (não requer cluster) |

### 7.3 Modelo de Comunicação

- O **agente** sempre inicia a comunicação (pull model)
- O agente **nunca** aceita conexões de entrada
- O agente envia **snapshots completos** a cada intervalo configurado
- Estratégia de **retry com backoff exponencial** em caso de falha de comunicação

---

## 8. Premissas

1. Os dispositivos Windows estão na mesma rede (ou têm acesso de rede) ao servidor da API
2. O servidor da API possui recursos suficientes para o volume de dispositivos planejado (ver [Gestão de Capacidade](../02-desenho-de-servico/gestao-de-capacidade.md))
3. O administrador de TI possui acesso administrativo para instalar o agente nas estações Windows
4. Na Fase 1, a comunicação HTTP é aceitável por estar em rede interna controlada

---

## 9. Restrições

1. **Fase 1 apenas** — funcionalidades limitadas ao escopo definido na seção 5
2. **Somente Windows** — agente incompatível com Linux/macOS nesta fase
3. **HTTP** — sem criptografia de transporte (risco aceito, ver [Gestão de Segurança](../02-desenho-de-servico/gestao-de-seguranca.md))
4. **Sem alta disponibilidade** — servidor único, sem clustering
5. **Snapshot completo** — cada envio substitui o anterior (sem sync diferencial)

---

## 10. Critérios de Sucesso

| Critério | Meta |
|---|---|
| Agente coleta dados corretos | 100% dos campos definidos preenchidos para cada dispositivo |
| Agente instalável como serviço | `agent.exe install && agent.exe start` funcional |
| API processa inventário | Recebe, valida e armazena snapshot em <500ms |
| Dashboard exibe dados | Todas as seções (sistema, hardware, discos, rede, software) renderizando |
| Cobertura de dispositivos | 100% dos dispositivos-alvo com agent instalado |
| Dados atualizados | last_seen < 2× intervalo configurado para todos os dispositivos |

---

## 11. Dependências

| Dependência | Tipo | Descrição |
|---|---|---|
| Docker + Docker Compose | Infraestrutura | Necessário para deploy da API, banco e dashboard |
| Servidor on-premises | Infraestrutura | Hardware para hospedar os containers |
| Rede interna | Infraestrutura | Conectividade entre agentes e servidor |
| Go toolchain | Desenvolvimento | Compilação do agente e da API |
| Node.js + npm | Desenvolvimento | Build do dashboard React |
| Git | Desenvolvimento | Controle de versão |
| GitHub | Plataforma | Repositório, CI/CD, releases |

---

## 12. Responsáveis

| Papel | Responsável | Atribuição |
|---|---|---|
| Proprietário do Serviço | Gestor de TI | Decisões estratégicas, priorização |
| Desenvolvedor | Equipe de Dev | Design, implementação, testes, documentação |
| Operador | Administrador de TI | Deploy, operação diária, suporte N1 |

---

## 13. Referências

- [Catálogo de Serviços](catalogo-de-servicos.md)
- [Análise Financeira](analise-financeira.md)
- [Gestão de Demanda](gestao-de-demanda.md)
- [Arquitetura da Solução](../02-desenho-de-servico/arquitetura-da-solucao.md)
- [Gestão de Segurança](../02-desenho-de-servico/gestao-de-seguranca.md)

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial |
