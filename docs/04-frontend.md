# Frontend — Interface Web

## Stack

- React 18 + TypeScript
- Vite (bundler + dev server)
- TailwindCSS v4 (sistema de temas via CSS custom properties)
- React Router v6 (SPA routing)
- TanStack React Query (data fetching + cache)
- Recharts (gráficos)
- Lucide (ícones)

## Estrutura do Projeto

```
frontend/
├── src/
│   ├── main.tsx              # Entry point, providers
│   ├── App.tsx               # Rotas
│   ├── index.css             # Tailwind + tema dark/light
│   ├── api/                  # Chamadas HTTP
│   │   ├── client.ts         # Base fetch wrapper
│   │   ├── auth.ts           # Login, logout, me
│   │   ├── dashboard.ts      # Stats
│   │   ├── devices.ts        # CRUD devices + export CSV
│   │   ├── departments.ts    # CRUD departments
│   │   └── users.ts          # CRUD users
│   ├── hooks/                # Contexts + hooks
│   │   ├── useAuth.tsx        # Autenticação
│   │   ├── useTheme.tsx       # Dark/light mode
│   │   ├── useSidebar.tsx     # Sidebar colapsável
│   │   └── useDebounce.ts     # Debounce genérico
│   ├── pages/                # Páginas
│   │   ├── Login.tsx
│   │   ├── Dashboard.tsx
│   │   ├── DeviceList.tsx
│   │   ├── DeviceDetail.tsx
│   │   ├── Departments.tsx
│   │   ├── DepartmentDetail.tsx
│   │   └── Settings.tsx
│   ├── components/           # Componentes reutilizáveis
│   │   ├── Layout.tsx         # Sidebar + main area
│   │   ├── ProtectedRoute.tsx # Guard de autenticação
│   │   ├── ErrorBoundary.tsx  # Catch de erros React
│   │   └── ui/               # Design system
│   │       ├── Button.tsx
│   │       ├── Input.tsx
│   │       ├── Select.tsx
│   │       ├── Badge.tsx
│   │       ├── Card.tsx
│   │       ├── Modal.tsx
│   │       ├── PageHeader.tsx
│   │       └── index.ts      # Barrel export
│   └── types/
│       └── index.ts           # Interfaces TypeScript
└── vite.config.ts
```

## Hierarquia de Providers

```tsx
QueryClientProvider          // TanStack Query (cache, retry)
  └── ThemeProvider           // Dark/light mode
       └── AuthProvider        // Estado de autenticação
            └── SidebarProvider // Sidebar colapsável
                 └── ErrorBoundary  // Catch de erros React
                      └── App       // Rotas
```

## Rotas

| Path | Página | Protegida | Descrição |
|------|--------|-----------|-----------|
| `/login` | Login | Não | Se já logado, redireciona para `/` |
| `/` | Dashboard | Sim | Estatísticas e gráficos |
| `/devices` | DeviceList | Sim | Lista com filtros, sort, paginação |
| `/devices/:id` | DeviceDetail | Sim | Detalhes completos do device |
| `/departments` | Departments | Sim | CRUD de departamentos |
| `/departments/:id` | DepartmentDetail | Sim | Devices de um departamento |
| `/settings` | Settings | Sim | Tema + gestão de usuários |
| `*` | — | — | Redireciona para `/` |

Rotas protegidas são envolvidas por `ProtectedRoute` que checa `isAuthenticated` e redireciona para `/login`.

## Client HTTP

Todas as chamadas usam `fetch` nativo (sem Axios):

```typescript
async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`/api/v1${path}`, {
    credentials: 'include',   // envia cookie httpOnly
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });

  if (res.status === 401) {
    localStorage.removeItem('authenticated');
    window.location.href = '/login';   // redirect global
  }

  if (!res.ok) {
    throw new ApiError(body.error, res.status);
  }

  return res.json();
}
```

- Base URL: `/api/v1` (Vite proxy em desenvolvimento)
- Cookies: `credentials: 'include'` para enviar o JWT httpOnly
- Erro 401: limpa `localStorage` e redireciona para login automaticamente
- Classe `ApiError` com campo `status` numérico

## Autenticação

### Fluxo de Login

```
1. Usuário digita username + password
2. POST /api/v1/auth/login → servidor seta cookie "session" (httpOnly)
3. Frontend seta localStorage('authenticated') = true
4. Frontend chama GET /auth/me → pega username e role
5. Redireciona para /
```

### Fluxo de Logout

```
1. POST /api/v1/auth/logout → servidor limpa cookie
2. Frontend limpa localStorage
3. Redireciona para /login
```

### Persistência de Sessão

O JWT está no cookie httpOnly, então JavaScript não tem acesso direto. O frontend usa `localStorage` apenas como flag:

- `localStorage.authenticated` = saber se deve tentar acessar APIs
- `localStorage.username` = exibir no sidebar

No mount do `AuthProvider`, se `authenticated === true`, chama `GET /auth/me` para validar. Se falhar (cookie expirado), limpa o estado.

## React Query

Configuração global:

```typescript
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,              // 1 retry em falha
      refetchOnWindowFocus: false,  // não refetch ao voltar pra aba
      staleTime: 30_000,     // dados ficam frescos por 30s
    },
  },
});
```

Após mutations (update device, create department, etc), as queries relacionadas são invalidadas manualmente para forçar refetch.

## Páginas

### Login

Card centralizado com efeito glass-morphism e gradientes de fundo animados. Form simples com username/password, exibe erro inline com animação.

### Dashboard

4 cards de estatísticas com gradientes e animação escalonada:
- **Total:** contagem de devices ativos
- **Online:** reportaram na última hora
- **Offline:** ativos sem reportar na última hora
- **Inactive:** desativados

Gráficos (Recharts):
- **PieChart:** distribuição de status (Online/Offline/Inactive)
- **BarChart:** quantidade de devices por departamento

Skeleton loading com `animate-pulse` enquanto carrega.

### DeviceList

**Filtros:**
| Filtro | Tipo | Comportamento |
|--------|------|---------------|
| Hostname | Input text | Debounce 300ms, filtro ILIKE |
| OS | Input text | Debounce 300ms |
| Status | Select dropdown | online/offline/inactive |
| Department | Select dropdown | Lista dinâmica |

**Tabela:**
- Colunas sortáveis: Hostname, OS, Last Seen, Status
- Click na coluna alterna asc/desc
- Status calculado client-side baseado em `last_seen`:
  - `< 1h` = Online (badge verde pulsante)
  - `≥ 1h` = Offline (badge cinza)
  - `status = inactive` = Inactive (badge vermelha)

**Paginação:** 50 items por página com controles prev/next e indicador de total.

**Export CSV:** Botão que dispara download via fetch + Blob + `<a>` temporário.

### DeviceDetail

Mostra todos os dados de um device em seções:

**Remote Access** — Cards para cada ferramenta instalada:
- TeamViewer, AnyDesk, RustDesk
- Dot colorido (verde = instalado com ID, amarelo = instalado sem ID)
- ID remoto com botão copy-to-clipboard
- Versão

**System** — Grid de informações:
- Hostname, Serial Number, OS (nome/versão/build/arch)
- Último boot, usuário logado, versão do agent, status da licença

**Hardware** — Info de CPU, RAM, placa-mãe, BIOS

**Physical Disks** — Tabela com modelo, tamanho, tipo, interface, serial

**Partitions** — Cards com barra de progresso visual:
- Drive letter (C:, D:, etc)
- Espaço usado vs total
- Barra colorida: verde (<70%), amarela (<90%), vermelha (≥90%)

**Network** — Tabela com nome, MAC, IPv4, IPv6, velocidade

**Software** — Tabela scrollável com todos os programas instalados

**Hardware History** — Timeline de mudanças detectadas. Cada snapshot mostra o que mudou (ex: RAM de 8GB → 16GB).

**Controles Admin:**
- Select de departamento (pode atribuir ou remover)
- Botão Activate/Deactivate

### Departments

CRUD completo:
- Form de criação (nome)
- Tabela com: nome, contagem de devices (badge), created_at
- Edição inline (clica Edit, campo vira input)
- Delete com modal de confirmação
- Link para DepartmentDetail

### DepartmentDetail

Lista de devices filtrada pelo departamento. Mesma funcionalidade do DeviceList (filtros, sort, paginação, export CSV), com coluna extra de "User" (logged_in_user).

### Settings

Duas seções:

**Aparência:**
- Toggle slider animado entre Dark e Light
- Ícones Sun/Moon

**Gestão de Usuários:**
- Form de criação: username + senha (min 8 chars)
- Tabela de usuários com username, data de criação, botão delete
- Modal de confirmação de exclusão
- Não permite deletar a si mesmo

## Sistema de Tema

### Como funciona

1. Classe `dark` ou `light` no `<html>`
2. CSS custom properties mudam baseado na classe
3. Componentes usam as variáveis Tailwind (`bg-bg-primary`, `text-text-primary`, etc)

### Cores

```css
:root.dark {
  --theme-bg-primary: #0a0c10;       /* fundo principal */
  --theme-bg-secondary: #141720;     /* cards, sidebar */
  --theme-bg-tertiary: #1c2030;      /* hover, alternating rows */
  /* textos, bordas, etc */
}

:root.light {
  --theme-bg-primary: #fafaf9;
  --theme-bg-secondary: #ffffff;
  --theme-bg-tertiary: #f5f0eb;
  /* textos, bordas, etc */
}
```

**Cor accent:** `#ea580c` (laranja) — usada em botões, links, indicadores de rota ativa.

- Persistência: `localStorage('theme')`
- Default: dark

## Design System (UI Components)

| Componente | Props principais | Variantes |
|------------|-----------------|-----------|
| Button | `variant`, `size`, `icon`, `loading` | primary, secondary, danger, success, ghost × sm, md, lg |
| Input | `icon`, `error` | — |
| Select | `icon` | — |
| Badge | `variant`, `dot`, `pulseDot` | accent, success, danger, warning, info, neutral |
| Card | `animate` | Com/sem animação slide-up |
| Modal | `open`, `onClose`, `title`, `actions` | Usa `<dialog>` nativo |
| PageHeader | `title`, `subtitle`, `actions` | — |

Todos os componentes usam o Tailwind com as variáveis de tema, garantindo que funcionam em dark e light mode.

## Layout

### Sidebar Colapsável

```
Expandida (240px)         Colapsada (60px)
┌──────────────┐          ┌────┐
│ 🏢 Inventory │          │ 🏢 │
│              │          │    │
│ ▪ Dashboard  │          │ 📊 │
│ ▪ Devices    │          │ 💻 │
│ ▪ Departments│          │ 🏢 │
│ ▪ Settings   │          │ ⚙  │
│              │          │    │
│              │          │    │
│ 👤 admin     │          │ 👤 │
│ 🌙 ⏻        │          │ ⏻  │
└──────────────┘          └────┘
```

- Barra laranja à esquerda indica rota ativa
- Estado colapsado/expandido persistido em `localStorage`
- Footer com avatar, toggle de tema, logout

## Animações

| Animação | Uso |
|----------|-----|
| `fade-in` | Entrada suave de opacity |
| `slide-up` | Cards entram de baixo |
| `scale-in` | Modals aparecem com zoom |
| `shimmer` | Loading skeleton |
| `pulse-dot` | Indicador de status Online |

## TypeScript Interfaces

O arquivo `types/index.ts` define todas as interfaces que espelham as respostas da API:

- `Device` — 17 campos (hostname, OS, status, department, timestamps)
- `Hardware` — CPU (model/cores/threads), RAM, mobo, BIOS
- `Disk` — modelo, tamanho, tipo, partição com espaço livre
- `NetworkInterface` — nome, MAC, IPv4, IPv6, velocidade
- `InstalledSoftware` — nome, versão, vendor, data
- `RemoteTool` — ferramenta, ID remoto, versão
- `HardwareHistory` — snapshot JSON, data da mudança
- `Department`, `User`
- Response wrappers: `DeviceListResponse`, `DeviceDetailResponse`, `DashboardStats`, etc
