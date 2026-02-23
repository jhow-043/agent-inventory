# Frontend (Dashboard Web)

**Versão:** 1.2.0

Dashboard web para visualização e gerenciamento do inventário de dispositivos, construído com React 19 e TypeScript.

## Stack

| Tecnologia | Versão | Uso |
|---|---|---|
| **React** | 19.1 | UI library |
| **TypeScript** | 5.8 | Type safety |
| **Vite** | 7.1 | Build tool + dev server + HMR |
| **Tailwind CSS** | 4.1 | Utility-first CSS (via `@tailwindcss/vite`) |
| **TanStack Query** | 5.90 | Server state management (cache, refetch, mutations) |
| **React Router** | 7.13 | Client-side routing |
| **Recharts** | 3.7 | Gráficos (pizza + barras) |
| **ESLint** | 9.36 | Linting |

## Arquitetura

```
frontend/
├── index.html                   # Entry point HTML
├── vite.config.ts               # Vite: host, port, proxy /api → backend
├── tsconfig.app.json            # TypeScript config
├── eslint.config.js             # ESLint flat config
├── package.json
└── src/
    ├── main.tsx                 # React root + QueryClientProvider
    ├── App.tsx                  # BrowserRouter + definição de rotas
    ├── index.css                # Tailwind imports + estilos globais
    ├── vite-env.d.ts
    │
    ├── api/                     # Camada HTTP
    │   ├── client.ts            # Fetch wrapper (credentials, auto-redirect 401, 204 handling)
    │   ├── auth.ts              # login(), logout(), getMe()
    │   ├── devices.ts           # listDevices(), getDevice(), exportCSV(), bulk ops
    │   ├── departments.ts       # CRUD departamentos
    │   ├── dashboard.ts         # getStats()
    │   ├── users.ts             # CRUD usuários
    │   └── audit.ts             # getAuditLogs()
    │
    ├── components/
    │   ├── Layout.tsx           # Sidebar colapsável + header + tema toggle + Outlet
    │   ├── GlobalSearch.tsx     # Busca global por hostname/OS
    │   ├── ProtectedRoute.tsx   # Guard: redireciona para /login se não autenticado
    │   ├── AdminRoute.tsx       # Guard: exige role admin
    │   ├── ErrorBoundary.tsx    # Captura erros React com fallback UI
    │   └── ui/                  # Componentes reutilizáveis
    │       ├── Badge.tsx        # Status badges com cores
    │       ├── Button.tsx       # Button com variantes + loading
    │       ├── Card.tsx         # Card container
    │       ├── Input.tsx        # Input com label
    │       ├── Modal.tsx        # Modal dialog
    │       ├── PageHeader.tsx   # Título + descrição de página
    │       ├── Select.tsx       # Select dropdown
    │       └── index.ts         # Re-exports
    │
    ├── hooks/
    │   ├── useAuth.tsx          # Context de autenticação (login, logout, user, role)
    │   ├── useTheme.tsx         # Context de tema (dark/light) com persistência localStorage
    │   ├── useDebounce.ts       # Debounce de valores (filtros)
    │   ├── useSidebar.tsx       # Context de sidebar (colapsada/expandida)
    │   └── useToast.tsx         # Context de notificações toast
    │
    ├── pages/
    │   ├── Login.tsx            # Tela de login com form e validação
    │   ├── Dashboard.tsx        # KPIs (cards) + gráficos Recharts (pizza OS + barras status)
    │   ├── DeviceList.tsx       # Tabela de dispositivos + filtros + paginação + export CSV
    │   ├── DeviceDetail.tsx     # Detalhe completo: hardware, discos, rede, software, remoto, histórico
    │   ├── Departments.tsx      # CRUD departamentos + contagem de devices
    │   ├── DepartmentDetail.tsx # Devices vinculados a um departamento
    │   └── Settings.tsx         # Tema dark/light + gerenciamento de usuários (admin)
    │
    └── types/
        └── index.ts             # Interfaces TypeScript (Device, User, Department, etc.)
```

## Rotas

| Rota | Componente | Auth | Descrição |
|---|---|---|---|
| `/login` | `Login` | Público | Tela de login |
| `/` | `Dashboard` | JWT | Dashboard com KPIs e gráficos |
| `/devices` | `DeviceList` | JWT | Lista de dispositivos (filtros e paginação) |
| `/devices/:id` | `DeviceDetail` | JWT | Detalhe completo de um dispositivo |
| `/departments` | `Departments` | JWT | Lista e CRUD de departamentos |
| `/departments/:id` | `DepartmentDetail` | JWT | Devices de um departamento |
| `/settings` | `Settings` | Admin | Tema + gerenciamento de usuários |

## Features

- **Tema Dark/Light** — toggle no header, persistido em `localStorage`
- **Sidebar colapsável** — menu lateral que pode ser recolhido
- **Busca global** — pesquisa por hostname e OS
- **Filtros avançados** — hostname, OS, status, departamento com debounce
- **Paginação** — navegação com controle de limite por página (até 200)
- **Export CSV** — exportação da lista de dispositivos
- **Hardware History** — timeline de mudanças de hardware com diff detalhado
- **Activity Log** — histórico de atividades do dispositivo
- **Bulk Operations** — alterar status/departamento ou excluir em lote (admin)
- **Audit Logs** — visualização de logs de auditoria do sistema (admin)
- **RBAC** — rotas admin protegidas no frontend e no backend
- **Auto-redirect 401** — sessão expirada redireciona automaticamente para login

## API Client

O client HTTP centralizado em `src/api/client.ts` oferece:

- Base URL: `/api/v1` (proxy via Vite em dev)
- `credentials: 'include'` — envia cookie JWT automaticamente
- Auto-redirect para `/login` em resposta 401 (sessão expirada)
- Tratamento de 204 No Content (respostas vazias)
- `ApiError` class com `status` para tratamento granular de erros

## Executando

### Desenvolvimento

```bash
cd frontend
npm install
npm run dev
```

O Vite inicia o dev server com HMR. A configuração padrão em `vite.config.ts`:
- **Host**: configurável (padrão: `localhost`)
- **Port**: `5173`
- **Proxy**: `/api` → `http://<IP_DO_SERVIDOR>:8081` (evita CORS em dev)

### Build de Produção

```bash
npm run build     # TypeScript check + Vite build → dist/
npm run preview   # Servir build localmente para teste
```

### Lint

```bash
npm run lint      # ESLint com regras React + TypeScript
```

## Configuração

### `vite.config.ts`

Para usar na rede interna, ajuste:

```typescript
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    host: '192.168.x.x',   // IP da sua máquina
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://192.168.x.x:8081',  // IP + porta da API
        changeOrigin: true,
      },
    },
  },
})
```
