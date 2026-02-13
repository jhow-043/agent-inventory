# Análise Financeira

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  
> **Status:** Aprovado  
> **Classificação:** Uso Interno  

---

## 1. Objetivo

Documentar o Custo Total de Propriedade (TCO) do Sistema de Inventário de Ativos de TI, comparar com alternativas de mercado e justificar economicamente o investimento.

---

## 2. Escopo

Análise financeira da Fase 1, considerando um parque de 100 a 500 dispositivos Windows, com deploy on-premises.

---

## 3. Custo Total de Propriedade (TCO) — Fase 1

### 3.1 Custos de Infraestrutura

| Item | Especificação Mínima | Custo Estimado |
|---|---|---|
| Servidor on-premises | 4 vCPU, 8GB RAM, 100GB SSD | Existente ou ~R$ 0 (reaproveitamento) |
| Rede interna | Switch/router existente | R$ 0 |
| Energia e refrigeração | Proporcional ao servidor | Já incluído nos custos operacionais |

> **Nota:** O sistema foi projetado para rodar em infraestrutura existente. Não requer hardware dedicado para cargas de até 500 dispositivos.

### 3.2 Custos de Licenciamento de Software

| Componente | Tecnologia | Licença | Custo |
|---|---|---|---|
| Agente Windows | Go | BSD-3-Clause | R$ 0 |
| API | Go + Gin | MIT | R$ 0 |
| Banco de dados | PostgreSQL | PostgreSQL License | R$ 0 |
| Dashboard | React + Vite | MIT | R$ 0 |
| Containers | Docker (CE) | Apache 2.0 | R$ 0 |
| SO do servidor | Linux (Docker host) | GPL | R$ 0 |
| **Total de licenciamento** | | | **R$ 0** |

### 3.3 Custos de Desenvolvimento

| Fase | Estimativa (horas) | Descrição |
|---|---|---|
| Fase 0 — Fundação | 16h | Estrutura do projeto, Docker, migrations, CI |
| Fase 1 — API | 40h | Endpoints, autenticação, testes de integração |
| Fase 2 — Agent | 40h | Coletores, Windows Service, retry, testes |
| Fase 3 — Dashboard | 32h | React, páginas, integração com API |
| Fase 4 — Integração | 16h | E2E, segurança, documentação, polish |
| **Total** | **~144h** | |

> O custo de desenvolvimento depende do modelo de contratação. Para desenvolvedor interno, é investimento de tempo. Para consultoria, multiplicar horas pela taxa hora/homem.

### 3.4 Custos Operacionais (Anual)

| Item | Estimativa Mensal | Anual |
|---|---|---|
| Manutenção e atualizações | ~4h/mês | ~48h/ano |
| Suporte N1 (instalação de agentes) | ~2h/mês | ~24h/ano |
| Backup e monitoramento | ~1h/mês | ~12h/ano |
| **Total operacional** | **~7h/mês** | **~84h/ano** |

---

## 4. Resumo TCO — Primeiro Ano

| Categoria | Custo |
|---|---|
| Infraestrutura (hardware) | R$ 0 (reaproveitamento) |
| Licenciamento de software | R$ 0 |
| Desenvolvimento (144h) | Custo conforme contratação |
| Operação (84h/ano) | Custo conforme contratação |
| **Total Fase 1** | **~228 horas-homem** |

---

## 5. Comparação com Soluções de Mercado

### 5.1 Soluções Open-Source

| Solução | Prós | Contras | Custo |
|---|---|---|---|
| **GLPI + FusionInventory** | Maduro, CMDB completo, ITSM integrado | Complexo de instalar/manter, PHP, UI pesada | R$ 0 (licença) + tempo de setup/manutenção |
| **OCS Inventory** | Inventário completo, agentes multiplataforma | Interface defasada, documentação fraca, Perl/PHP | R$ 0 (licença) + tempo de setup |
| **Snipe-IT** | Gestão de ativos limpa, API REST | Foco em gestão manual (não em coleta automática via agente) | R$ 0 (licença) |

### 5.2 Soluções Comerciais

| Solução | Preço Estimado (500 devices) | Funcionalidades |
|---|---|---|
| **Lansweeper** | ~US$ 2-5/device/ano (~R$ 5.000–12.500/ano) | Inventário completo, reports, integração |
| **ManageEngine AssetExplorer** | ~US$ 4-8/device/ano (~R$ 10.000–20.000/ano) | CMDB, inventário, ITSM |
| **PDQ Inventory** | ~US$ 1.500/ano (license) | Inventário Windows, reports, scan ativo |

### 5.3 Análise Comparativa

| Critério | Solução Própria | GLPI | Lansweeper |
|---|---|---|---|
| Custo de licença | R$ 0 | R$ 0 | R$ 5.000–12.500/ano |
| Custo de setup | ~144h dev | ~40h config | ~16h config |
| Customização | Total | Limitada (plugins) | Limitada |
| Manutenção | Própria | Comunidade | Fornecedor |
| Fit com requisitos | 100% | ~70% (over-engineered) | ~80% (features desnecessárias) |
| Dependência externa | Nenhuma | Comunidade PHP | Vendor lock-in |
| Evolução futura | Total controle | Depende de plugins | Depende do roadmap |

---

## 6. Retorno sobre Investimento (ROI)

### 6.1 Economia de Tempo

| Atividade | Sem sistema (manual) | Com sistema (automático) | Economia/mês |
|---|---|---|---|
| Inventário de hardware | ~20h/mês | ~0h (automático) | 20h |
| Inventário de software | ~16h/mês | ~0h (automático) | 16h |
| Verificação de licenças | ~8h/mês | ~1h (consulta dashboard) | 7h |
| Consulta de config para suporte | ~4h/mês | ~0.5h (busca no dashboard) | 3.5h |
| **Total economizado** | | | **~46.5h/mês** |

### 6.2 Payback

- **Investimento:** ~228 horas-homem (1º ano, dev + operação)
- **Economia:** ~46.5 horas/mês = ~558 horas/ano
- **Payback:** ~5 meses (228h ÷ 46.5h/mês)

> Após o payback, o sistema gera economia líquida de ~330 horas/ano (~41 dias úteis).

---

## 7. Riscos Financeiros

| Risco | Probabilidade | Impacto | Mitigação |
|---|---|---|---|
| Subestimação do esforço de desenvolvimento | Média | Médio | Desenvolvimento incremental, escopo fixo na Fase 1 |
| Custo de manutenção maior que previsto | Baixa | Baixo | Código limpo, testes automatizados, documentação |
| Necessidade de hardware dedicado | Baixa | Baixo | Sistema leve, sizing conservador |
| Abandono do projeto | Média | Alto | Entregas incrementais com valor a cada fase |

---

## 8. Recomendação

Considerando:
- Stack 100% open-source (custo zero de licenciamento)
- Reaproveitamento de infraestrutura existente
- Payback estimado em 5 meses
- Controle total sobre customização e evolução
- Independência de fornecedores

**Recomenda-se o desenvolvimento próprio** como a opção de melhor custo-benefício para o cenário de 100-500 dispositivos com deploy on-premises.

---

## 9. Referências

- [Visão Geral do Serviço](visao-geral-do-servico.md)
- [Gestão de Capacidade](../02-desenho-de-servico/gestao-de-capacidade.md)
- [Gestão de Fornecedores](../02-desenho-de-servico/gestao-de-fornecedores.md)

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial |
