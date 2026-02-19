# Contribuindo

Obrigado pelo interesse em contribuir com o Inventory System!

## Pré-requisitos

- **Go** 1.24+
- **Node.js** 20+
- **Docker** e **Docker Compose**
- **PostgreSQL** 16 (via Docker)

## Setup do Ambiente

```bash
# Clone o repositório
git clone https://github.com/jhow-043/agent-inventory.git
cd agent-inventory

# Suba o banco e a API
docker compose up -d

# Instale as dependências do frontend
cd frontend && npm install

# Rode o frontend em modo dev
npm run dev
```

## Fluxo de Trabalho

### Branches

| Branch | Propósito |
|--------|----------|
| `main` | Branch de produção — sempre estável |
| `develop` | Integração de features (quando aplicável) |
| `feature/<nome>` | Novas funcionalidades |
| `fix/<nome>` | Correção de bugs |
| `hotfix/<nome>` | Correções urgentes em produção |

### Commits

Usamos [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: adiciona filtro por componente no hardware history
fix: corrige comparação de MAC address
docs: atualiza documentação da API
refactor: extrai lógica de comparação de hardware
chore: atualiza dependências do frontend
ci: adiciona step de lint no workflow
```

**Escopos comuns:** `server`, `agent`, `frontend`, `shared`, `ui`, `api`, `db`

### Pull Requests

1. Crie uma branch a partir de `main`
2. Faça seus commits seguindo a convenção
3. Atualize o `CHANGELOG.md`
4. Abra um PR usando o template fornecido
5. Aguarde review

## Estrutura do Projeto

```
├── agent/          # Agent Windows (Go) — coleta inventário via WMI
├── server/         # API Backend (Go + Gin)
├── frontend/       # Dashboard SPA (React + Vite + Tailwind)
├── shared/         # Modelos e DTOs compartilhados (Go)
├── docs/           # Documentação técnica
└── .github/        # CI/CD e templates
```

## Padrões de Código

### Go (server + agent)

- Use `gofmt` / `goimports` para formatação
- Siga as diretrizes do [Effective Go](https://go.dev/doc/effective_go)
- Trate todos os erros — não use `_` para ignorar erros
- Use `slog` para logging estruturado

### TypeScript (frontend)

- Use ESLint com a configuração do projeto
- Componentes React em PascalCase
- Hooks customizados com prefixo `use`
- Types em `src/types/index.ts`

### SQL (migrations)

- Nomeie migrations sequencialmente: `NNN_description.up.sql` / `.down.sql`
- Sempre forneça o script de rollback (down)
- Use `UUID` para PKs e `TIMESTAMPTZ` para datas

## Testes

```bash
# Testes do server
cd server && go test -v ./...

# Lint do frontend
cd frontend && npm run lint

# Type-check do frontend
cd frontend && npx tsc --noEmit
```

## Reportando Bugs

Use o template de [Bug Report](.github/ISSUE_TEMPLATE/bug_report.md) nas Issues.

## Licença

Ao contribuir, você concorda que suas contribuições serão licenciadas sob a [Licença MIT](LICENSE).
