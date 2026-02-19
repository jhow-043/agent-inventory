# Agent Windows

## O que é

Executável Go que roda como serviço Windows, coleta informações de hardware e software via WMI e Registry, e envia para a API periodicamente.

## Estrutura do Projeto

```
agent/
├── cmd/agent/main.go             # Entry point, CLI, Windows Service, loop principal
├── internal/
│   ├── client/client.go          # HTTP client com retry
│   ├── collector/
│   │   ├── collector.go          # Orquestrador de coleta
│   │   ├── system.go             # Hostname, OS, serial, boot time
│   │   ├── hardware.go           # CPU, RAM, placa-mãe, BIOS
│   │   ├── disk.go               # Discos físicos + partições
│   │   ├── network.go            # Adaptadores de rede
│   │   ├── software.go           # Software instalado (Registry)
│   │   ├── license.go            # Status de ativação Windows
│   │   └── remote.go             # TeamViewer, AnyDesk, RustDesk
│   ├── config/config.go          # Configuração JSON
│   └── token/store.go            # Persistência do device token
└── Dockerfile                    # Cross-compile Windows
```

## Comandos

```bash
inventory-agent.exe <comando>
```

| Comando | Descrição |
|---------|-----------|
| `install` | Instala como serviço Windows (StartAutomatic) |
| `uninstall` | Remove o serviço Windows |
| `start` | Inicia o serviço |
| `stop` | Envia sinal de parada ao serviço |
| `run` | Roda em foreground (modo debug) |
| `run -config C:\path\config.json` | Foreground com config custom |
| `collect` | Coleta inventário e imprime JSON (não precisa de servidor) |
| `version` | Mostra versão |

Se executado sem argumentos, detecta se está rodando como serviço Windows (`svc.IsWindowsService()`). Se sim, roda o agente. Se não, mostra o help.

## Serviço Windows

- **Nome:** `InventoryAgent`
- **Display Name:** `Inventory Agent`
- **Descrição:** `Windows IT Asset Inventory Agent`
- **Tipo de inicio:** Automático
- **Aceita:** Stop e Shutdown

O serviço implementa `svc.Handler` via struct `agentService`. Recebe comandos do SCM (Service Control Manager):

```
svc.Interrogate → reporta status atual
svc.Stop/Shutdown → cancela context → encerra loop → serviço para
```

## Ciclo Principal

```
runAgent(ctx)
    │
    ├── Carrega config.json
    ├── Configura logger JSON (slog)
    ├── Cria token store (data/device.token)
    ├── Cria collector (WMI/Registry)
    ├── Cria HTTP client
    ├── Carrega token salvo (se existir)
    │
    ├── runCycle() ← executa imediatamente
    │
    └── Ticker (interval_hours) ← repete
            │
            ├── runCycle()
            ├── runCycle()
            └── ...
```

### runCycle()

```
1. Coleta inventário completo (WMI + Registry)
   Se falhar → log error → para

2. Se não tem token:
   → POST /api/v1/enroll (hostname + serial_number + enrollment_key)
   → Recebe {device_id, token}
   → Salva token em data/device.token

3. POST /api/v1/inventory com Bearer token
   → SubmitWithRetry: até 5 tentativas
   → Backoff: 2^attempt segundos + jitter (0-999ms)
   → Se recebe 401 ou 403:
       Limpa token (deleta arquivo)
       Próximo ciclo vai fazer enrollment novamente
```

## Configuração

Arquivo `config.json` (por padrão, ao lado do executável):

```json
{
  "server_url": "http://192.168.1.100:8080",
  "enrollment_key": "minha-chave-secreta",
  "interval_hours": 1,
  "data_dir": "data",
  "log_level": "info",
  "insecure_skip_verify": false
}
```

| Campo | Obrigatório | Default | Descrição |
|-------|-------------|---------|-----------|
| `server_url` | **Sim** | — | URL base da API |
| `enrollment_key` | **Sim** | — | Deve ser igual à `ENROLLMENT_KEY` do servidor |
| `interval_hours` | Não | `1` | Intervalo entre coletas (horas) |
| `data_dir` | Não | `data/` (ao lado do .exe) | Diretório para armazenar o token |
| `log_level` | Não | `info` | `debug`, `info`, `warn`, `error` |
| `insecure_skip_verify` | Não | `false` | Pular verificação TLS (usar apenas em desenvolvimento) |

## Token Store

O token do device é salvo em arquivo texto simples: `<data_dir>/device.token`

- Permissões: diretório `0700`, arquivo `0600`
- Contém apenas o token raw (UUID)
- Se o arquivo não existir, o agent faz enrollment
- Se a API retornar 401/403, o arquivo é deletado para forçar re-enrollment

## Collectors — O que Cada Um Coleta

### System (system.go)

Informações básicas do sistema operacional.

| Dado | Fonte WMI/API | Query |
|------|---------------|-------|
| Hostname | `os.Hostname()` | — |
| OS (nome) | `Win32_OperatingSystem` | `SELECT Caption, Version, BuildNumber, OSArchitecture, LastBootUpTime` |
| OS (versão, build, arch) | `Win32_OperatingSystem` | ↑ mesmo query |
| Último boot | `Win32_OperatingSystem` | ↑ campo `LastBootUpTime` |
| Serial Number | `Win32_BIOS` | `SELECT SerialNumber FROM Win32_BIOS` |
| Usuário logado | `Win32_ComputerSystem` | `SELECT UserName FROM Win32_ComputerSystem` |

Se a coleta de Sistema falhar, a coleta inteira é abortada (é a única que causa erro fatal).

### Hardware (hardware.go)

Informações de CPU, memória, placa-mãe e BIOS.

| Dado | Fonte WMI | Query |
|------|-----------|-------|
| CPU modelo, cores, threads | `Win32_Processor` | `SELECT Name, NumberOfCores, NumberOfLogicalProcessors` |
| RAM total | `Win32_PhysicalMemory` | `SELECT Capacity` (soma de todos os pentes) |
| Placa-mãe | `Win32_BaseBoard` | `SELECT Manufacturer, Product, SerialNumber` |
| BIOS vendor, versão | `Win32_BIOS` | `SELECT Manufacturer, SMBIOSBIOSVersion` |

Se falhar, retorna struct vazio (coleta continua normalmente).

### Disk (disk.go)

Discos físicos e partições lógicas.

| Dado | Fonte WMI | Query |
|------|-----------|-------|
| Discos físicos | `Win32_DiskDrive` | `SELECT Model, Size, MediaType, SerialNumber, InterfaceType` |
| Partições | `Win32_LogicalDisk` | `SELECT DeviceID, Size, FreeSpace FROM Win32_LogicalDisk WHERE DriveType = 3` |

**DriveType = 3** filtra apenas discos locais fixos (ignora CD-ROM, rede, etc).

**Classificação de tipo de disco:**
| MediaType (Win32_DiskDrive) | Classificação |
|----------------------------|---------------|
| `Fixed hard disk media` | HDD |
| `Removable Media` | Removable |
| vazio (`""`) | SSD |
| outro | valor original |

Partições são adicionadas como items separados na lista de discos, com `DriveLetter`, `PartitionSizeBytes` e `FreeSpaceBytes`.

### Network (network.go)

Adaptadores de rede físicos com IPs.

**WMI query:**

```sql
SELECT Name, MACAddress, Speed, PhysicalAdapter
FROM Win32_NetworkAdapter
WHERE PhysicalAdapter = TRUE AND MACAddress IS NOT NULL
```

Filtra apenas adaptadores físicos (ignora virtuais como VPN, Hyper-V, etc).

**Resolução de IPs:** O WMI não retorna IPs de forma confiável, então o agent usa `net.Interfaces()` do Go para mapear MAC → IPs:

```
1. Lista todas as interfaces Go (net.Interfaces)
2. Para cada interface com HardwareAddr:
   → Coleta IPv4 e IPv6 (ignora loopback e link-local)
   → Monta mapa: MAC → [IPv4, IPv6]
3. Para cada adaptador WMI:
   → Normaliza MAC (uppercase)
   → Busca no mapa → associa IPs
```

### Software (software.go)

Software instalado via Windows Registry.

**4 caminhos de registro consultados:**

```
HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\*
HKLM\SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\*
HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\*
HKCU\SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\*
```

- `HKLM` = software instalado para todos os usuários
- `HKCU` = software instalado só para o usuário atual
- `WOW6432Node` = software 32-bit em sistema 64-bit

**Filtros:**
- Ignora entradas sem `DisplayName`
- Ignora entradas com `SystemComponent = 1` (componentes do sistema)
- Dedup por `name|version` (lowercase) para evitar duplicatas entre HKLM e HKCU

**Dados coletados por software:**
| Campo | Valor do Registry |
|-------|-------------------|
| Name | `DisplayName` |
| Version | `DisplayVersion` |
| Vendor | `Publisher` |
| InstallDate | `InstallDate` |

### License (license.go)

Status de ativação do Windows.

**WMI query:**

```sql
SELECT LicenseStatus FROM SoftwareLicensingProduct
WHERE ApplicationID = '55c92734-d682-4d71-983e-d6ec3f16059f'
AND PartialProductKey IS NOT NULL
```

O GUID `55c92734-d682-4d71-983e-d6ec3f16059f` é o ApplicationID padrão do Windows em todas as versões.

| LicenseStatus | Retorno |
|---------------|---------|
| 0 | Unlicensed |
| 1 | Licensed |
| 2 | OOBGrace |
| 3 | OOTGrace |
| 4 | NonGenuineGrace |
| 5 | Notification |
| 6 | ExtendedGrace |
| sem resultado | Not Found |
| erro WMI | Unknown |

### Remote Tools (remote.go)

Detecta TeamViewer, AnyDesk e RustDesk com ID remoto e versão.

#### TeamViewer

```
Registry: HKLM\SOFTWARE\TeamViewer (ou WOW6432Node)
  → ClientID (integer) → ID remoto
  → Version (string) → versão
```

#### AnyDesk

```
1. Versão: busca em Uninstall keys do Registry
   Se não achar, verifica se AnyDesk.exe existe em Program Files

2. ID remoto: lê system.conf procurando "ad.anynet.id="
   Paths tentados:
   → %ProgramData%\AnyDesk\system.conf
   → %APPDATA%\AnyDesk\system.conf
   → %USERPROFILE%\AppData\Roaming\AnyDesk\system.conf
```

#### RustDesk

Três estratégias (em ordem de prioridade):

```
1. Registry:
   HKLM\SOFTWARE\RustDesk → ID e Version
   HKLM\SOFTWARE\WOW6432Node\RustDesk
   HKCU\SOFTWARE\RustDesk

2. Config file (TOML):
   %APPDATA%\RustDesk\config\RustDesk.toml
   %ProgramData%\RustDesk\config\RustDesk.toml
   → Procura linha "id = <valor>"

3. CLI fallback (v1.4+):
   Procura rustdesk.exe em:
   → Program Files\RustDesk\
   → Program Files (x86)\RustDesk\
   → Install location do Registry
   Executa: rustdesk.exe --get-id (timeout 10s)
```

## HTTP Client

### Configuração

```go
Timeout:   30 segundos
TLS:       Opcionalmente pode pular verificação (insecure_skip_verify)
```

### Retry com Exponential Backoff

Quando `SubmitInventory` falha, o agent tenta novamente até 5 vezes:

```
Tentativa 1: falhou → espera 1s + jitter
Tentativa 2: falhou → espera 2s + jitter
Tentativa 3: falhou → espera 4s + jitter
Tentativa 4: falhou → espera 8s + jitter
Tentativa 5: falhou → espera 16s + jitter
Tentativa 6: falhou → desiste (retorna erro)
```

Fórmula: $wait = 2^{attempt} + random(0\text{-}999ms)$

Se o contexto for cancelado (serviço parando), o retry é interrompido imediatamente.

### Detecção de Erro de Auth

Se o erro contém "status 401" ou "status 403":
- O token é marcado como inválido
- O arquivo `data/device.token` é deletado
- No próximo ciclo, o agent vai fazer enrollment novamente

## Orquestrador de Coleta

O `collector.go` chama os collectors na ordem:

```
1. System   → FATAL se falhar (retorna erro, ciclo aborta)
2. Hardware → Warn se falhar (usa struct vazio)
3. Disks    → Warn se falhar (lista vazia)
4. Network  → Warn se falhar (lista vazia)
5. Software → Warn se falhar (lista vazia)
6. License  → Warn se falhar (retorna "Unknown")
7. Remote   → Nunca falha (retorna lista vazia)
```

Apenas a coleta de sistema é crítica. Todas as outras são best-effort — se falharem, o inventário é enviado mesmo assim com os dados que conseguiu coletar.

## Modo Collect (Dry Run)

```bash
inventory-agent.exe collect
```

Coleta todos os dados e imprime o JSON formatado no stdout. Não precisa de servidor configurado, não faz enrollment, não envia dados. Útil para debug e verificação do que será coletado.

Exemplo de saída:

```json
{
  "hostname": "DESKTOP-ABC123",
  "serial_number": "5CG1234567",
  "os_name": "Microsoft Windows 11 Pro",
  "os_version": "10.0.22631",
  "os_build": "22631",
  "os_arch": "64-bit",
  "agent_version": "0.1.0",
  "license_status": "Licensed",
  "hardware": {
    "cpu_model": "13th Gen Intel(R) Core(TM) i7-13700K",
    "cpu_cores": 16,
    "cpu_threads": 24,
    "ram_total_bytes": 34359738368
  },
  "disks": [...],
  "network": [...],
  "software": [...],
  "remote_tools": [...]
}
```
