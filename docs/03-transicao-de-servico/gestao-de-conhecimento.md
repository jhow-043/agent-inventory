# Gestão de Conhecimento

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  
> **Status:** Aprovado  
> **Classificação:** Uso Interno  

---

## 1. Objetivo

Garantir que o conhecimento sobre o Sistema de Inventário de Ativos de TI seja capturado, organizado e acessível para todos os envolvidos, reduzindo dependência de indivíduos e acelerando onboarding.

---

## 2. Escopo

Conhecimento técnico e operacional de todos os componentes da Fase 1.

---

## 3. Base de Conhecimento

### 3.1 Estrutura

| Localização | Tipo de Conhecimento | Formato |
|---|---|---|
| `docs/` (este diretório) | Documentação ITIL, processos, procedimentos | Markdown |
| `README.md` (raiz) | Visão geral, como rodar, como buildar | Markdown |
| Código-fonte | GoDoc, JSDoc, comentários inline | In-code |
| `docs/06-anexos/diagramas/` | Diagramas de arquitetura, fluxos, ER | Mermaid (Markdown) |
| Git history | Decisões, contexto de mudanças | Commits convencionais |
| GitHub Issues | Bugs, features, decisões técnicas | Issue tracker |

### 3.2 Documentação Técnica no Código

| Padrão | Linguagem | Exemplo |
|---|---|---|
| GoDoc | Go | Comentário acima de cada função/tipo exportado |
| JSDoc | TypeScript | Comentário acima de componentes e funções |
| Inline | Ambos | Comentários explicando "por quê" (não "o quê") |

```go
// SubmitInventory processes an inventory snapshot from an agent.
// It performs an upsert of the device and replaces all related data
// (hardware, disks, network interfaces, installed software) in a
// single database transaction.
func (s *InventoryService) SubmitInventory(ctx context.Context, payload *models.InventoryPayload) error {
```

---

## 4. Onboarding de Novo Desenvolvedor

### 4.1 Checklist de Onboarding

| # | Atividade | Duração Estimada | Recurso |
|---|---|---|---|
| 1 | Ler [README.md](../../README.md) | 15 min | Repositório |
| 2 | Ler [Visão Geral do Serviço](../01-estrategia-de-servico/visao-geral-do-servico.md) | 15 min | docs/ |
| 3 | Ler [Arquitetura da Solução](../02-desenho-de-servico/arquitetura-da-solucao.md) | 30 min | docs/ |
| 4 | Clonar repositório e rodar `make dev` | 15 min | Makefile |
| 5 | Explorar a estrutura de código | 30 min | Editor |
| 6 | Rodar testes: `make test` | 10 min | Terminal |
| 7 | Ler [Gestão de Mudanças](../03-transicao-de-servico/gestao-de-mudancas.md) | 15 min | docs/ |
| 8 | Fazer uma mudança simples e abrir PR | 30 min | Prática |
| 9 | Ler [Runbooks Operacionais](../04-operacao-de-servico/runbooks-operacionais.md) | 20 min | docs/ |
| 10 | Ler diagrama ER e explorar migrations | 15 min | docs/ + código |
| **Total** | | **~3 horas** | |

### 4.2 Pré-requisitos Técnicos

| Ferramenta | Versão | Instalação |
|---|---|---|
| Go | 1.22+ | https://go.dev/dl/ |
| Node.js | 20+ LTS | https://nodejs.org/ |
| Docker Desktop | Latest | https://docker.com/ |
| Git | 2.40+ | https://git-scm.com/ |
| VS Code | Latest | https://code.visualstudio.com/ |
| golangci-lint | Latest | `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` |

### 4.3 Extensões VS Code Recomendadas

| Extensão | Propósito |
|---|---|
| Go (golang.go) | Suporte Go |
| ES7+ React Snippets | Snippets React |
| Tailwind CSS IntelliSense | Autocomplete Tailwind |
| Prettier | Formatação JS/TS |
| Docker | Suporte Docker |
| REST Client | Testar endpoints |
| Mermaid Preview | Renderizar diagramas |

---

## 5. FAQ Técnico

### 5.1 Agent

| Pergunta | Resposta |
|---|---|
| Como o agent se registra na API? | Envia `POST /api/v1/register` com enrollment key. A API retorna um device token único. |
| Onde o agent armazena o token? | Em `C:\ProgramData\InventoryAgent\` com permissões restritas (SYSTEM + Administrators). |
| O que acontece se a API estiver offline? | Agent faz retry com backoff exponencial: 2s → 4s → 8s → ... → max 5min. |
| O agent precisa de internet? | Não. Apenas acesso de rede ao servidor da API (rede interna). |
| Como desinstalar o agent? | `agent.exe stop && agent.exe uninstall`. Remover pasta manualmente. |

### 5.2 API

| Pergunta | Resposta |
|---|---|
| Como adicionar um novo endpoint? | Criar handler → service → repository (se necessário) → registrar rota no Gin router. |
| Como criar uma migration? | Criar arquivo `NNNNNN_description.up.sql` e `.down.sql` em `server/migrations/`. |
| Como rodar migrations manualmente? | `make migrate` ou `migrate -path server/migrations -database "$DATABASE_URL" up`. |
| O que é o enrollment key? | Chave compartilhada que permite novos agents se registrarem. Configurada via env var. |
| Como revogar um device token? | DELETE na tabela `device_tokens` filtrando por device_id. |

### 5.3 Dashboard

| Pergunta | Resposta |
|---|---|
| Como o dashboard se autentica? | `POST /auth/login` retorna JWT em httpOnly cookie. Renovado automaticamente. |
| Como adicionar uma nova página? | Criar componente em `pages/`, adicionar rota no React Router, proteger com AuthGuard. |
| Como mudar estilos? | Tailwind CSS utility classes. Temas em `tailwind.config.js`. |
| Como testar componentes? | `npm run test` (Vitest + Testing Library). |

### 5.4 Infraestrutura

| Pergunta | Resposta |
|---|---|
| Como acessar o banco em dev? | `docker exec -it inventory-postgres psql -U inventory -d inventory` |
| Como ver logs da API? | `docker compose logs -f api` |
| Como fazer backup do banco? | `docker exec inventory-postgres pg_dump -U inventory -Fc inventory > backup.dump` |
| Como restaurar um backup? | Ver [Runbook RB-005](../04-operacao-de-servico/runbooks-operacionais.md). |

---

## 6. Lições Aprendidas

### 6.1 Template de Lição Aprendida

```markdown
## Lição: [Título]

**Data:** YYYY-MM-DD
**Contexto:** [O que aconteceu]
**Problema:** [O que deu errado ou o que foi aprendido]
**Aprendizado:** [O que fazer diferente no futuro]
**Ação:** [Mudança implementada, se houver]
```

### 6.2 Registro de Lições Aprendidas

| # | Data | Título | Resumo |
|---|---|---|---|
| — | — | *(será preenchido durante o desenvolvimento)* | — |

---

## 7. Treinamento

### 7.1 Treinamento para Operadores (Admin TI)

| Módulo | Conteúdo | Duração |
|---|---|---|
| 1. Visão geral | O que é o sistema, componentes, como funciona | 30 min |
| 2. Dashboard | Login, navegação, interpretação dos dados | 30 min |
| 3. Operação | Verificar saúde, reiniciar serviços, verificar logs | 30 min |
| 4. Instalação do agent | Instalar, configurar, verificar funcionamento | 30 min |
| 5. Troubleshooting | Problemas comuns e soluções (runbooks) | 30 min |
| **Total** | | **2.5 horas** |

### 7.2 Material de Treinamento

| Tipo | Localização |
|---|---|
| Documentação escrita | Esta pasta `docs/` |
| Runbooks passo-a-passo | [Runbooks Operacionais](../04-operacao-de-servico/runbooks-operacionais.md) |
| FAQ | Seção 5 deste documento |

---

## 8. Política de Manutenção do Conhecimento

| Atividade | Frequência | Responsável |
|---|---|---|
| Atualizar documentação quando código muda | A cada PR relevante | Desenvolvedor |
| Revisão do FAQ | Trimestral | Desenvolvedor |
| Atualizar lições aprendidas | Após cada incidente/post-mortem | Desenvolvedor |
| Revisão de onboarding | A cada nova pessoa | Desenvolvedor |
| Verificar links internos nos docs | Semestralmente | Desenvolvedor |

---

## 9. Referências

- [Visão Geral do Serviço](../01-estrategia-de-servico/visao-geral-do-servico.md)
- [Arquitetura da Solução](../02-desenho-de-servico/arquitetura-da-solucao.md)
- [Runbooks Operacionais](../04-operacao-de-servico/runbooks-operacionais.md)
- [Glossário](../06-anexos/glossario.md)

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial |
