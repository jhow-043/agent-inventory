# Script para testar o envio de inventário

$apiUrl = "http://localhost:8080/api/v1"
$enrollmentKey = "8a3f9c2e7b4d1f6a"

# 1. Enroll do dispositivo
Write-Host "Enrollando dispositivo..." -ForegroundColor Cyan

$enrollPayload = @{
    hostname = "TEST-PC-001"
    serial_number = "SN123456789"
} | ConvertTo-Json

$enrollResponse = Invoke-RestMethod -Uri "$apiUrl/enroll" `
    -Method POST `
    -Headers @{"X-Enrollment-Key" = $enrollmentKey; "Content-Type" = "application/json"} `
    -Body $enrollPayload

Write-Host "Device ID: $($enrollResponse.device_id)" -ForegroundColor Green
Write-Host "Token: $($enrollResponse.token)" -ForegroundColor Green

# 2. Enviar inventário
Write-Host "`nEnviando inventário..." -ForegroundColor Cyan

$inventoryPayload = @{
    hostname = "TEST-PC-001"
    serial_number = "SN123456789"
    os_name = "Microsoft Windows 11 Pro"
    os_version = "10.0.22631"
    os_build = "22631"
    os_arch = "64-bit"
    last_boot_time = (Get-Date).AddDays(-2).ToString("o")
    logged_in_user = "teste.usuario"
    agent_version = "1.0.0"
    license_status = "Licensed"
    hardware = @{
        cpu_model = "Intel(R) Core(TM) i7-10700K CPU @ 3.80GHz"
        cpu_cores = 8
        cpu_threads = 16
        ram_total_bytes = 34359738368  # 32 GB
        motherboard_manufacturer = "ASUS"
        motherboard_product = "ROG STRIX Z490-E GAMING"
        motherboard_serial = "MB123456789"
        bios_vendor = "American Megatrends Inc."
        bios_version = "2.14.1245"
    }
    disks = @(
        @{
            model = "Samsung SSD 970 EVO Plus 1TB"
            size_bytes = 1000204886016  # 1 TB
            media_type = "SSD"
            serial_number = "S4EWNX0N123456"
            interface_type = "NVMe"
            drive_letter = "C:"
            partition_size_bytes = 1000204886016
            free_space_bytes = 450000000000
        },
        @{
            model = "WD Blue 2TB"
            size_bytes = 2000398934016  # 2 TB
            media_type = "HDD"
            serial_number = "WD-WCAZA1234567"
            interface_type = "SATA"
            drive_letter = "D:"
            partition_size_bytes = 2000398934016
            free_space_bytes = 1500000000000
        }
    )
    network_interfaces = @(
        @{
            name = "Ethernet"
            mac_address = "00:D8:61:AA:BB:CC"
            ipv4_address = "192.168.1.100"
            ipv6_address = "fe80::1234:5678:abcd:ef01"
            speed_mbps = 1000
            is_physical = $true
        },
        @{
            name = "Wi-Fi"
            mac_address = "A4:B1:C2:D3:E4:F5"
            ipv4_address = "192.168.1.101"
            ipv6_address = ""
            speed_mbps = 867
            is_physical = $true
        }
    )
    installed_software = @(
        @{
            name = "Microsoft Office Professional Plus 2021"
            version = "16.0.14332.20481"
            vendor = "Microsoft Corporation"
            install_date = "2024-01-15"
        },
        @{
            name = "Google Chrome"
            version = "120.0.6099.129"
            vendor = "Google LLC"
            install_date = "2024-02-01"
        },
        @{
            name = "Adobe Acrobat Reader DC"
            version = "23.008.20470"
            vendor = "Adobe Inc."
            install_date = "2024-01-20"
        },
        @{
            name = "Visual Studio Code"
            version = "1.85.2"
            vendor = "Microsoft Corporation"
            install_date = "2024-02-05"
        }
    )
    remote_tools = @(
        @{
            tool_name = "AnyDesk"
            remote_id = "987654321"
            version = "7.1.14"
        },
        @{
            tool_name = "TeamViewer"
            remote_id = "1234567890"
            version = "15.47.5"
        }
    )
} | ConvertTo-Json -Depth 10

$inventoryResponse = Invoke-RestMethod -Uri "$apiUrl/inventory" `
    -Method POST `
    -Headers @{"Authorization" = "Bearer $($enrollResponse.token)"; "Content-Type" = "application/json"} `
    -Body $inventoryPayload

Write-Host "Inventário enviado com sucesso!" -ForegroundColor Green
Write-Host "Mensagem: $($inventoryResponse.message)" -ForegroundColor Green
