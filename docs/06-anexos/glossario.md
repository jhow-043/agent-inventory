# Glossário

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  

---

## Termos ITIL

| Termo | Definição |
|---|---|
| **CMDB** | Configuration Management Database — banco de dados com todos os itens de configuração |
| **CI** | Configuration Item — qualquer componente que precisa ser gerenciado para entregar um serviço |
| **CSI** | Continual Service Improvement — melhoria contínua de serviço |
| **Incidente** | Interrupção não planejada ou degradação de qualidade de um serviço de TI |
| **ITIL** | Information Technology Infrastructure Library — framework de boas práticas para gestão de serviços de TI |
| **KEDB** | Known Error Database — base de dados de erros conhecidos com workarounds |
| **Known Error** | Problema com causa raiz identificada e workaround documentado |
| **MTBF** | Mean Time Between Failures — tempo médio entre falhas |
| **MTTR** | Mean Time To Restore/Repair — tempo médio para restaurar o serviço |
| **Problema** | Causa raiz desconhecida de um ou mais incidentes |
| **RACI** | Responsible, Accountable, Consulted, Informed — matriz de responsabilidades |
| **RFC** | Request For Change — solicitação formal de mudança |
| **RPO** | Recovery Point Objective — máxima perda de dados aceitável |
| **RTO** | Recovery Time Objective — tempo máximo para restaurar o serviço |
| **SLA** | Service Level Agreement — acordo de nível de serviço |
| **SLI** | Service Level Indicator — indicador mensurável do serviço |
| **SLO** | Service Level Objective — meta para um SLI |
| **Service Request** | Solicitação formal para uma ação pré-definida (não é incidente nem mudança) |
| **Workaround** | Solução temporária que restaura o serviço sem corrigir a causa raiz |

---

## Termos Técnicos

| Termo | Definição |
|---|---|
| **Agent** | Aplicação Go compilada que roda como serviço Windows, coleta dados de inventário via WMI e envia para a API |
| **API** | Application Programming Interface — interface RESTful que recebe dados dos agents e serve o dashboard |
| **Backoff** | Estratégia de espera progressiva entre tentativas de retry (ex: exponential backoff) |
| **bcrypt** | Algoritmo de hash para senhas, com custo computacional configurável |
| **CI/CD** | Continuous Integration / Continuous Delivery — integração e entrega contínua |
| **Clean Architecture** | Padrão arquitetural que separa regras de negócio de detalhes de implementação |
| **Container** | Unidade de software empacotada com código, runtime e dependências (Docker) |
| **CORS** | Cross-Origin Resource Sharing — mecanismo que permite requests entre origens diferentes |
| **Dashboard** | Interface web React para visualização do inventário |
| **Delta Sync** | Envio apenas das mudanças desde a última coleta, em vez do inventário completo |
| **Device Token** | Token único (SHA-256 hashed) que identifica um agent registrado |
| **Docker Compose** | Ferramenta para definir e rodar aplicações Docker multi-container |
| **Enrollment** | Processo de registro inicial de um agent usando a enrollment key |
| **Enrollment Key** | Chave compartilhada usada para o primeiro registro de um agent |
| **Go** | Linguagem de programação compilada, criada pelo Google, usada no backend e agent |
| **Gin** | Framework web HTTP para Go |
| **Health Check** | Endpoint que verifica se um componente está operacional |
| **httpOnly Cookie** | Cookie que não pode ser acessado via JavaScript (proteção XSS) |
| **Jitter** | Variação aleatória adicionada a intervalos para evitar thundering herd |
| **JWT** | JSON Web Token — token de autenticação para usuários do dashboard |
| **Liveness Probe** | Verifica se a aplicação está rodando (responde em `/healthz`) |
| **Migration** | Script de alteração de esquema do banco de dados, versionado |
| **Modular Monolith** | Aplicação implantada como um só binário, mas organizada internamente em módulos independentes |
| **ORM** | Object-Relational Mapping — camada de abstração entre código e banco relacional |
| **pgx** | Driver PostgreSQL nativo para Go, de alta performance |
| **PostgreSQL** | Banco de dados relacional open source |
| **RACI** | Matriz de papéis: Responsável, Aprovador, Consultado, Informado |
| **Rate Limiting** | Limitação de quantidade de requisições por período |
| **React** | Biblioteca JavaScript para construção de interfaces de usuário |
| **Readiness Probe** | Verifica se a aplicação está pronta para receber tráfego (`/readyz`) |
| **Repository Pattern** | Padrão que abstrai acesso a dados por trás de uma interface |
| **REST** | Representational State Transfer — estilo arquitetural para APIs |
| **SemVer** | Semantic Versioning — versionamento MAJOR.MINOR.PATCH |
| **SHA-256** | Algoritmo de hash criptográfico de 256 bits |
| **Shadcn/UI** | Biblioteca de componentes React baseada em Radix UI e Tailwind CSS |
| **slog** | Pacote de logging estruturado da stdlib do Go (1.21+) |
| **sqlx** | Extensão do `database/sql` do Go com features como Named Queries e Struct Scan |
| **Tailwind CSS** | Framework CSS utility-first |
| **TanStack Query** | Biblioteca de data fetching e caching para React |
| **testcontainers-go** | Biblioteca para usar Docker containers em testes de integração Go |
| **Thundering Herd** | Problema quando muitos agents enviam simultaneamente, sobrecarregando a API |
| **TLS/HTTPS** | Transport Layer Security — protocolo de criptografia de comunicação |
| **TypeScript** | Superset tipado de JavaScript |
| **Upsert** | Operação que insere se não existe, ou atualiza se existe (INSERT ON CONFLICT UPDATE) |
| **UUID** | Universally Unique Identifier — identificador único de 128 bits |
| **Vite** | Build tool e dev server para projetos web modernos |
| **WAL** | Write-Ahead Log — mecanismo de journaling do PostgreSQL |
| **Windows Service** | Processo Windows que roda em background, gerenciado pelo SCM |
| **WMI** | Windows Management Instrumentation — interface para coletar informações do sistema Windows |

---

## Abreviações do Projeto

| Abreviação | Significado |
|---|---|
| **API** | Interface programática REST do servidor |
| **CI** | Configuration Item (ITIL) ou Continuous Integration (DevOps) — contexto define |
| **DB** | Database (PostgreSQL) |
| **HW** | Hardware |
| **INC** | Incidente |
| **KE** | Known Error |
| **RB** | Runbook |
| **SR** | Service Request |
| **SW** | Software |
| **UI** | User Interface (Dashboard React) |

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial |
