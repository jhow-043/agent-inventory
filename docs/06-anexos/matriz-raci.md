# Matriz RACI

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  
> **Status:** Aprovado  
> **Classificação:** Uso Interno  

---

## 1. Objetivo

Definir as responsabilidades de cada papel em cada processo e atividade do Sistema de Inventário de Ativos de TI usando o modelo RACI.

---

## 2. Legenda

| Letra | Significado | Descrição |
|---|---|---|
| **R** | Responsible (Responsável) | Executa a atividade |
| **A** | Accountable (Aprovador) | Responde pelo resultado final. Apenas um por atividade. |
| **C** | Consulted (Consultado) | Contribui com informação ou expertise antes da decisão |
| **I** | Informed (Informado) | É notificado após a decisão ou conclusão |

---

## 3. Papéis

| ID | Papel | Descrição |
|---|---|---|
| **DEV** | Desenvolvedor | Desenvolve, testa e implanta o sistema. Na Fase 1, é o único desenvolvedor. |
| **ADM** | Administrador de TI | Opera o sistema no dia a dia, instala agents, gerencia usuários |
| **GES** | Gestor de TI | Patrocinador do projeto. Aprova mudanças de alto impacto, define prioridades |

> **Nota:** Na Fase 1 (equipe solo), DEV acumula a maioria das funções R e A. À medida que a equipe crescer, as responsabilidades devem ser redistribuídas.

---

## 4. Matriz RACI — Estratégia de Serviço

| Atividade | DEV | ADM | GES |
|---|---|---|---|
| Definir visão e objetivos do serviço | R | C | A |
| Aprovar escopo do projeto | C | — | A |
| Analisar custos (TCO/ROI) | R | — | A |
| Priorizar funcionalidades | R | C | A |
| Definir roadmap de evolução | R/A | I | C |

---

## 5. Matriz RACI — Desenho de Serviço

| Atividade | DEV | ADM | GES |
|---|---|---|---|
| Definir arquitetura da solução | R/A | — | I |
| Escolher stack tecnológica | R/A | — | I |
| Definir SLIs/SLOs | R | C | A |
| Definir requisitos de segurança | R/A | C | I |
| Planejar capacidade de hardware | R | A | I |
| Definir estratégia de backup (RPO/RTO) | R/A | C | I |
| Documentar processos de continuidade | R/A | C | I |
| Avaliar dependências/fornecedores | R/A | — | I |

---

## 6. Matriz RACI — Transição de Serviço

| Atividade | DEV | ADM | GES |
|---|---|---|---|
| Desenvolver código | R/A | — | — |
| Revisar Pull Requests | R/A | — | — |
| Executar migrations de banco | R/A | I | — |
| Build e deploy da API/Dashboard | R/A | I | — |
| Deploy do Agent em estações | C | R/A | I |
| Deploy em massa (GPO) | C | R/A | I |
| Executar testes automatizados | R/A | — | — |
| Executar testes de aceitação | R | R | A |
| Gerenciar configuração (CMDB) | R/A | C | — |
| Documentar mudanças (changelog) | R/A | — | I |
| Aprovar mudanças normais | R | — | A |
| Aprovação mudanças de emergência | R/A | — | I |
| Treinamento de operadores | R | R | I |

---

## 7. Matriz RACI — Operação de Serviço

| Atividade | DEV | ADM | GES |
|---|---|---|---|
| Monitorar health checks | I | R/A | — |
| Analisar logs e métricas | R | C | — |
| Detectar incidentes | C | R | — |
| Classificar incidentes (P1-P4) | R/A | C | — |
| Resolver incidentes P1/P2 | R/A | C | I |
| Resolver incidentes P3/P4 | R/A | C | — |
| Executar post-mortem (P1/P2) | R/A | C | I |
| Análise de causa raiz (RCA) | R/A | C | — |
| Manter KEDB | R/A | C | — |
| Executar backup diário | — | R/A | — |
| Validar integridade do backup | R | A | — |
| Restaurar backup (disaster recovery) | R/A | C | I |
| Registrar novo agent | I | R/A | — |
| Revogar token de device | R | R/A | I |
| Criar/remover usuário do dashboard | I | R | A |
| Resetar senha de usuário | — | R/A | — |
| Rotacionar enrollment key | R | A | — |
| Gerar relatório de inventário | C | R/A | I |
| Verificação de saúde semanal (RB-010) | C | R/A | — |

---

## 8. Matriz RACI — Melhoria Contínua

| Atividade | DEV | ADM | GES |
|---|---|---|---|
| Coletar métricas mensais | R/A | C | — |
| Gerar relatório mensal | R/A | — | I |
| Revisão mensal de métricas | R | C | I |
| Revisão trimestral de SLOs | R | C | A |
| Propor melhorias (CSI Register) | R | C | C |
| Priorizar melhorias | R | C | A |
| Implementar melhorias | R/A | I | — |
| Migração HTTP → HTTPS (futura) | R/A | C | A |

---

## 9. Resumo de Carga por Papel

### DEV (Desenvolvedor)

| Tipo | Exemplos | Frequência |
|---|---|---|
| R/A (executa e responde) | Desenvolvimento, deploy, RCA, KEDB, documentação | Diário |
| R (executa) | Incidentes P1/P2, relatórios, métricas | Semanal |
| C (consultado) | Deploy de agents, requisições operacionais | Sob demanda |

### ADM (Administrador de TI)

| Tipo | Exemplos | Frequência |
|---|---|---|
| R/A (executa e responde) | Install/gerenciar agents, backup, monitoramento, service requests | Diário |
| R (executa) | Testes de aceitação, relatórios | Semanal |
| C (consultado) | Design, incidentes, melhorias | Sob demanda |

### GES (Gestor de TI)

| Tipo | Exemplos | Frequência |
|---|---|---|
| A (aprova) | Escopo, SLOs, mudanças, prioridades, orçamento | Mensal |
| C (consultado) | Prioridades, melhorias | Sob demanda |
| I (informado) | Incidentes P1/P2, deploys, métricas | Sob demanda |

---

## 10. Notas

1. **Equipe solo:** Na Fase 1, o DEV acumula praticamente todas as funções R e A técnicas. Isso é aceitável para o escopo de 100-500 devices, mas deve ser reavaliado quando/se a equipe crescer.

2. **Bus Factor:** Como o DEV é responsável por quase tudo técnico, a documentação e os runbooks são especialmente críticos para mitigar o risco de indisponibilidade do DEV.

3. **Evolução:** Ao adicionar novos membros, redistribuir primeiro:
   - Operação de Serviço (monitoramento, backup, service requests) → Segundo ADM
   - Desenvolvimento (code review, testes, deploy) → Segundo DEV
   - Segurança → Papel dedicado ou consultor

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial |
