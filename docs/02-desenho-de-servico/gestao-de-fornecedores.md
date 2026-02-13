# Gestão de Fornecedores

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  
> **Status:** Aprovado  
> **Classificação:** Uso Interno  

---

## 1. Objetivo

Documentar todas as dependências externas (tecnologias, bibliotecas, serviços) do Sistema de Inventário de Ativos de TI, suas licenças, riscos e alternativas.

---

## 2. Escopo

Todas as dependências utilizadas na Fase 1, incluindo runtime, bibliotecas, ferramentas de desenvolvimento e infraestrutura.

---

## 3. Dependências de Infraestrutura

| Fornecedor/Tecnologia | Tipo | Licença | Custo | Criticidade | Substituível por |
|---|---|---|---|---|---|
| **Docker CE** | Container runtime | Apache 2.0 | Gratuito | Alta | Podman |
| **Docker Compose** | Orquestração local | Apache 2.0 | Gratuito | Alta | Podman Compose |
| **PostgreSQL 16** | Banco de dados | PostgreSQL License | Gratuito | Crítica | MySQL (com adaptações) |
| **Linux (host)** | Sistema operacional | GPL | Gratuito | Alta | Windows Server + Docker |
| **GitHub** | Repositório + CI/CD | SaaS (Free tier) | Gratuito | Média | GitLab, Gitea |
| **GitHub Actions** | CI/CD | SaaS (Free tier) | Gratuito | Média | GitLab CI, Jenkins |

---

## 4. Dependências Go — Backend (Server)

| Biblioteca | Versão | Licença | Propósito | Criticidade | Alternativa |
|---|---|---|---|---|---|
| **gin-gonic/gin** | v1.9+ | MIT | Framework HTTP | Alta | Echo, Chi, stdlib |
| **jmoiron/sqlx** | v1.3+ | MIT | DB access wrapper | Alta | GORM, Bun, stdlib |
| **jackc/pgx** | v5 | MIT | Driver PostgreSQL nativo | Alta | lib/pq |
| **golang-migrate/migrate** | v4 | MIT | Migrations SQL | Média | goose, atlas |
| **golang-jwt/jwt** | v5 | MIT | Geração/validação JWT | Alta | lestrrat-go/jwx |
| **go-playground/validator** | v10 | MIT | Validação de structs | Média | ozzo-validation |
| **caarlos0/env** | v11 | MIT | Parsing de env vars | Baixa | viper, envconfig |
| **google/uuid** | v1 | BSD-3-Clause | Geração de UUIDs | Baixa | gofrs/uuid |

---

## 5. Dependências Go — Agent

| Biblioteca | Versão | Licença | Propósito | Criticidade | Alternativa |
|---|---|---|---|---|---|
| **kardianos/service** | v1 | Zlib | Windows Service wrapper | Alta | golang.org/x/sys (manual) |
| **yusufpapurcu/wmi** | latest | MIT | Coleta WMI no Windows | Alta | go-ole (manual) |
| **go-ole/go-ole** | v1 | MIT | COM/OLE interface (dep wmi) | Alta | — (fundamental) |
| **caarlos0/env** | v11 | MIT | Parsing de config | Baixa | viper |
| **gopkg.in/yaml.v3** | v3 | Apache 2.0 | Parsing de YAML config | Baixa | encoding/json |

---

## 6. Dependências JavaScript — Frontend

| Biblioteca | Versão | Licença | Propósito | Criticidade | Alternativa |
|---|---|---|---|---|---|
| **react** | 18+ | MIT | Framework UI | Crítica | Vue.js, Svelte |
| **react-dom** | 18+ | MIT | React DOM renderer | Crítica | — |
| **react-router-dom** | v6+ | MIT | Roteamento SPA | Alta | TanStack Router |
| **@tanstack/react-query** | v5 | MIT | Server state management | Alta | SWR, RTK Query |
| **typescript** | 5+ | Apache 2.0 | Type safety | Alta | — (JavaScript puro) |
| **tailwindcss** | v3+ | MIT | Utility CSS framework | Média | styled-components, CSS Modules |
| **vite** | v5+ | MIT | Build tool | Alta | webpack, esbuild |
| **vitest** | latest | MIT | Test runner | Média | Jest |
| **@testing-library/react** | latest | MIT | Teste de componentes | Média | Enzyme |

### Dependências de Desenvolvimento (Frontend)

| Biblioteca | Licença | Propósito |
|---|---|---|
| **eslint** | MIT | Linting |
| **prettier** | MIT | Formatação |
| **postcss** | MIT | Processamento CSS |
| **autoprefixer** | MIT | Vendor prefixes |

---

## 7. Ferramentas de Desenvolvimento

| Ferramenta | Versão | Licença | Propósito | Obrigatória |
|---|---|---|---|---|
| **Go** | 1.22+ | BSD-3-Clause | Compilação agent/server | Sim |
| **Node.js** | 20+ LTS | MIT | Build do dashboard | Sim |
| **npm** | 10+ | Artistic 2.0 | Gerenciador de pacotes JS | Sim |
| **golangci-lint** | latest | GPL-3.0 | Linter Go agregado | Sim (CI) |
| **Make** | — | GPL | Build automation | Recomendado |
| **Git** | 2.40+ | GPL-2.0 | Controle de versão | Sim |

---

## 8. Análise de Risco por Dependência

### 8.1 Matriz de Risco

| Critério | Peso | Explicação |
|---|---|---|
| **Maturidade** | Alto | Projeto estável, >3 anos, releases frequentes |
| **Manutenção** | Alto | Commits recentes (últimos 6 meses), issues respondidas |
| **Comunidade** | Médio | Stars, contributors, usage em produção |
| **Licença** | Médio | Compatível com uso comercial/interno |
| **Alternativas** | Baixo | Existem alternativas viáveis |

### 8.2 Avaliação

| Dependência | Maturidade | Manutenção | Comunidade | Risco de Abandono |
|---|---|---|---|---|
| Gin | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Muito Baixo |
| sqlx | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | Baixo |
| pgx | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | Muito Baixo |
| golang-jwt | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Muito Baixo |
| kardianos/service | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ | Baixo |
| yusufpapurcu/wmi | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ | Médio |
| React | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Muito Baixo |
| Vite | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Muito Baixo |
| TanStack Query | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | Baixo |
| Tailwind CSS | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Muito Baixo |

### 8.3 Dependências de Risco Médio

| Dependência | Risco | Plano de Contingência |
|---|---|---|
| **yusufpapurcu/wmi** | Fork menos popular; maintainer pode abandonar | Migrar para go-ole direto ou StackExchange/wmi se necessário |
| **kardianos/service** | Baixa atividade de commits | Alternativa: implementar Windows Service manualmente com x/sys/windows |

---

## 9. Política de Atualização

### 9.1 Estratégia

| Tipo de Update | Ação | Prazo |
|---|---|---|
| **Patch** (x.x.PATCH) | Atualizar automaticamente | Contínuo |
| **Minor** (x.MINOR.x) | Revisar changelog, atualizar | 1 semana |
| **Major** (MAJOR.x.x) | Avaliar breaking changes, testar em dev | Próxima release |
| **Vulnerabilidade crítica** | Patch emergencial | 24 horas |

### 9.2 Ferramentas de Automação

| Ferramenta | Configuração |
|---|---|
| **Dependabot** (GitHub) | Go + npm, PRs automáticos semanais |
| **go mod verify** | Executado em cada build no CI |
| **npm audit** | Executado em cada build no CI |
| **govulncheck** | Executado semanalmente no CI |

---

## 10. Licenciamento

### 10.1 Compatibilidade

Todas as dependências usam licenças permissivas compatíveis com uso interno:

| Licença | Tipo | Restrições |
|---|---|---|
| **MIT** | Permissiva | Nenhuma (incluir copyright) |
| **BSD-3-Clause** | Permissiva | Nenhuma (incluir copyright) |
| **Apache 2.0** | Permissiva | Incluir NOTICE se presente |
| **PostgreSQL License** | Permissiva | Nenhuma |
| **Zlib** | Permissiva | Nenhuma |

### 10.2 Licenças Ausentes

Nenhuma dependência utiliza licenças restritivas (GPL, AGPL) no código de produção.

> **Nota:** `golangci-lint` é GPL-3.0, mas é ferramenta de desenvolvimento, não distribuída com o produto.

---

## 11. Responsáveis

| Papel | Responsável | Atribuição |
|---|---|---|
| Gestão de dependências | Desenvolvedor | Atualização, avaliação de risco |
| Aprovação de novas deps | Desenvolvedor (self-review) | Avaliar licença, maturidade, necessidade |
| Monitoramento de CVEs | Automático (Dependabot + govulncheck) | Alertas de vulnerabilidade |

---

## 12. Referências

- [Arquitetura da Solução](arquitetura-da-solucao.md)
- [Gestão de Segurança](gestao-de-seguranca.md)
- [Gestão de Mudanças](../03-transicao-de-servico/gestao-de-mudancas.md)

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial |
