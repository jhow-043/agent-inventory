# Validação e Testes

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  
> **Status:** Aprovado  
> **Classificação:** Uso Interno  

---

## 1. Objetivo

Definir a estratégia de testes completa do Sistema de Inventário de Ativos de TI, garantindo que cada release atenda aos requisitos de qualidade antes de ser implantada em produção.

---

## 2. Escopo

Testes para todos os componentes da Fase 1: Agent Windows, API, Dashboard e integração ponta a ponta.

---

## 3. Pirâmide de Testes

```
         ╱╲
        ╱ E2E ╲          Poucos, lentos, alto valor de confiança
       ╱────────╲
      ╱ Integração╲      Quantidade moderada, validam fluxos completos
     ╱──────────────╲
    ╱   Unitários     ╲   Muitos, rápidos, isolados
   ╱────────────────────╲
```

| Tipo | Quantidade | Velocidade | Confiança | Custo de Manutenção |
|---|---|---|---|---|
| **Unitário** | Muitos | Rápido (ms) | Média | Baixo |
| **Integração** | Moderado | Médio (s) | Alta | Médio |
| **E2E** | Poucos | Lento (s-min) | Muito Alta | Alto |

---

## 4. Testes Unitários

### 4.1 Backend Go (Server + Agent)

| Aspecto | Detalhe |
|---|---|
| **Framework** | `testing` (stdlib) + `testify` (assertions e mocks) |
| **Execução** | `go test ./... -race` |
| **Cobertura alvo** | ≥ 80% na camada `service/`, ≥ 70% em `repository/` |
| **O que testar** | Lógica de negócio, validações, transformações de dados |
| **O que NÃO testar** | Handlers HTTP (cobertos em integração), infra/config |

#### Exemplos de Testes Unitários

| Componente | Teste | Técnica |
|---|---|---|
| `service/inventory.go` | Upsert com dados válidos retorna sucesso | Mock do repository interface |
| `service/inventory.go` | Upsert com token inválido retorna erro | Mock do repository |
| `service/auth.go` | Login com senha correta retorna JWT | Mock do user repository |
| `service/auth.go` | Login com senha errada retorna erro | Mock do user repository |
| `service/device.go` | Filtro por hostname funciona | Mock do repository |
| `agent/collector/system.go` | Parse de dados WMI retorna struct correta | Mock da interface WMI |
| `agent/transport/retry.go` | Backoff exponencial calcula delays corretos | Teste de lógica pura |

#### Convenções

```go
// Arquivo: service/inventory_test.go
func TestInventoryService_SubmitInventory_Success(t *testing.T) { ... }
func TestInventoryService_SubmitInventory_InvalidToken(t *testing.T) { ... }
func TestInventoryService_SubmitInventory_DatabaseError(t *testing.T) { ... }
```

- Nomeação: `Test<Struct>_<Method>_<Scenario>`
- Um arquivo de teste por arquivo de código (`inventory.go` → `inventory_test.go`)
- Table-driven tests para variações de input

### 4.2 Frontend React

| Aspecto | Detalhe |
|---|---|
| **Framework** | Vitest + @testing-library/react |
| **Execução** | `npm run test` |
| **Cobertura alvo** | Componentes críticos (login, device list, device detail) |
| **O que testar** | Renderização de componentes, interações, hooks customizados |
| **O que NÃO testar** | Bibliotecas de terceiros, estilos CSS |

#### Exemplos de Testes Unitários (Frontend)

| Componente | Teste |
|---|---|
| `LoginPage` | Exibe formulário com campos username e password |
| `LoginPage` | Chama API de login ao submeter formulário |
| `LoginPage` | Exibe erro ao receber 401 |
| `DeviceTable` | Renderiza lista de devices corretamente |
| `DeviceTable` | Filtra por hostname |
| `DeviceDetail` | Exibe todas as abas (System, Hardware, Disks, Network, Software) |
| `useAuth` | Retorna isAuthenticated=true quando há token válido |

---

## 5. Testes de Integração

### 5.1 API + Banco de Dados

| Aspecto | Detalhe |
|---|---|
| **Framework** | `testing` + testcontainers-go |
| **Banco** | PostgreSQL real via container Docker (testcontainers) |
| **Execução** | `go test ./server/... -tags=integration` |
| **O que testar** | Handlers → Services → Repositories → PostgreSQL real |
| **Isolamento** | Cada teste de integração usa transação com rollback ou banco limpo |

#### Exemplo: Setup com Testcontainers

```go
func setupTestDB(t *testing.T) *sqlx.DB {
    ctx := context.Background()
    container, _ := postgres.Run(ctx,
        "postgres:16-alpine",
        postgres.WithDatabase("inventory_test"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
    )
    t.Cleanup(func() { container.Terminate(ctx) })

    connStr, _ := container.ConnectionString(ctx, "sslmode=disable")
    db, _ := sqlx.Connect("pgx", connStr)

    // Rodar migrations
    runMigrations(db)

    return db
}
```

#### Cenários de Integração

| Cenário | Fluxo testado |
|---|---|
| Registro de device | POST /register → cria device + token no banco → retorna token |
| Submissão de inventário | POST /inventory (com token) → upsert no banco → dados corretos |
| Login de usuário | POST /auth/login → retorna JWT → JWT válido |
| Lista de devices | GET /devices → retorna devices do banco com filtros |
| Detalhe do device | GET /devices/:id → retorna device com hardware, discos, rede, software |
| Token inválido | POST /inventory (token errado) → 401 Unauthorized |
| Rate limiting | 11 requests em 1 min → 429 Too Many Requests |
| Health checks | GET /healthz → 200; GET /readyz → 200 (com banco) |

### 5.2 Agent — Coletores

| Aspecto | Detalhe |
|---|---|
| **Técnica** | Interface mockada para WMI |
| **Execução** | `go test ./agent/... -race` |
| **Limitação** | Coleta real de WMI só pode ser testada em Windows |

```go
// Interface que permite mock
type WMIQuerier interface {
    Query(query string, dst interface{}) error
}

// Em teste: mock que retorna dados predefinidos
type mockWMI struct{}
func (m *mockWMI) Query(query string, dst interface{}) error {
    // Retorna dados de teste
}
```

---

## 6. Testes End-to-End (E2E)

### 6.1 Ferramentas

| Aspecto | Detalhe |
|---|---|
| **Framework** | Playwright |
| **Browser** | Chromium (headless no CI) |
| **Execução** | `npx playwright test` |
| **Ambiente** | Docker Compose com API + PostgreSQL + Dashboard |
| **Dados** | Seed script com dados de demonstração |

### 6.2 Cenários E2E

| ID | Cenário | Passos |
|---|---|---|
| E2E-01 | **Login bem-sucedido** | Acessar /login → preencher credenciais → clicar Login → redirecionar para /dashboard |
| E2E-02 | **Login falho** | Acessar /login → preencher senha errada → mensagem de erro exibida |
| E2E-03 | **Dashboard carrega** | Login → dashboard mostra total de devices, contagem online/offline |
| E2E-04 | **Lista de devices** | Login → navegar para /devices → tabela exibe devices → filtrar por hostname |
| E2E-05 | **Detalhe do device** | Login → /devices → clicar em device → abas carregam (System, Hardware, Disks, Network, Software) |
| E2E-06 | **Logout** | Login → clicar logout → redirecionar para /login → /dashboard inacessível |
| E2E-07 | **Rota protegida** | Acessar /dashboard sem login → redirecionar para /login |

---

## 7. Testes de Carga

### 7.1 Ferramenta

- **k6** (Grafana k6) — testes de carga em Go, scripts em JavaScript
- Alternativa: `hey` ou script Go customizado

### 7.2 Cenários de Carga

| Cenário | VUs (Virtual Users) | Duração | Meta |
|---|---|---|---|
| Carga normal | 50 (simulando 500 devices/intervalo) | 5 min | P95 < 500ms, 0% erro |
| Pico | 200 | 2 min | P95 < 1s, < 1% erro |
| Stress | 500 | 2 min | Identificar ponto de quebra |
| Soak | 50 | 30 min | Sem memory leak, latência estável |

### 7.3 Script de Exemplo (k6)

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 50,
  duration: '5m',
};

const payload = JSON.stringify({
  hostname: `device-${__VU}`,
  serial_number: `SN-${__VU}`,
  os_version: 'Windows 11 Pro',
  // ... dados completos de inventário
});

export default function () {
  const res = http.post('http://localhost:8080/api/v1/inventory', payload, {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${__ENV.DEVICE_TOKEN}`,
    },
  });
  check(res, { 'status is 200': (r) => r.status === 200 });
  sleep(1);
}
```

---

## 8. Testes de Segurança

| Teste | Ferramenta | Frequência |
|---|---|---|
| Scan de vulnerabilidades (Go deps) | `govulncheck` | Cada build |
| Scan de vulnerabilidades (npm) | `npm audit` | Cada build |
| Scan de imagens Docker | `trivy` / `docker scout` | Cada build |
| Teste de autenticação | Manual + integração | Cada release |
| Teste de rate limiting | Integração | Cada release |
| Teste de input validation | Unitário + integração | Contínuo |

---

## 9. Critérios de Aceitação

### 9.1 Para merge em develop

- [ ] CI pipeline 100% green
- [ ] Cobertura de testes não diminuiu
- [ ] Nenhuma vulnerabilidade crítica em deps
- [ ] Self-review checklist completo

### 9.2 Para release (tag)

- [ ] Todos os critérios de merge
- [ ] Testes de integração passam com banco real
- [ ] Testes E2E passam (cenários críticos)
- [ ] Teste de carga executado (carga normal sem erros)
- [ ] Changelog atualizado
- [ ] Documentação atualizada

---

## 10. Cobertura de Testes

### 10.1 Metas

| Componente | Camada | Meta de Cobertura |
|---|---|---|
| Server | `service/` | ≥ 80% |
| Server | `repository/` | ≥ 70% (integração) |
| Server | `handler/` | ≥ 60% (via integração) |
| Agent | `collector/` | ≥ 70% |
| Agent | `transport/` | ≥ 80% |
| Web | Componentes críticos | Sim (sem % rígido) |

### 10.2 Medição

- **Go:** `go test -coverprofile=coverage.out; go tool cover -func=coverage.out`
- **React:** `vitest --coverage`
- **Relatório:** Gerado no CI, anexado ao PR

---

## 11. Ambientes de Teste

| Ambiente | Propósito | Banco | Execução |
|---|---|---|---|
| Local | Desenvolvimento e testes manuais | Docker Compose (dev) | Manual |
| CI (GitHub Actions) | Testes automatizados | Testcontainers | Automática |
| E2E | Testes de ponta a ponta | Docker Compose (test) | CI ou manual |

---

## 12. Responsáveis

| Papel | Responsável | Atribuição |
|---|---|---|
| Escrever testes | Desenvolvedor | Unitários, integração, E2E |
| Manter CI | Desenvolvedor | Pipeline GitHub Actions |
| Executar testes de carga | Desenvolvedor | Antes de releases |
| Validação funcional | Administrador de TI | Teste manual pós-deploy |

---

## 13. Referências

- [Gestão de Liberação e Implantação](gestao-de-liberacao-e-implantacao.md)
- [Gestão de Mudanças](gestao-de-mudancas.md)
- [Requisitos de Nível de Serviço](../02-desenho-de-servico/requisitos-de-nivel-de-servico.md)

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial |
