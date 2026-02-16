<#
.SYNOPSIS
    Build the Inventory Agent installer.

.DESCRIPTION
    1. Compiles the Go agent for Windows amd64
    2. Runs Inno Setup to produce the installer EXE with embedded enrollment key

.EXAMPLE
    .\build-installer.ps1 -EnrollmentKey "my-secret-key"
    .\build-installer.ps1 -EnrollmentKey "my-secret-key" -SkipGoBuild
#>
param(
    [Parameter(Mandatory)]
    [string]$EnrollmentKey,

    [switch]$SkipGoBuild
)

$ErrorActionPreference = "Stop"
$agentRoot = Split-Path $PSScriptRoot -Parent

# ── Step 1: Build Go binary ────────────────────────────────────────────────
$buildDir = Join-Path $agentRoot "build"
$exePath  = Join-Path $buildDir "inventory-agent.exe"

if (-not $SkipGoBuild) {
    Write-Host ">> Building Go agent..." -ForegroundColor Cyan
    Push-Location $agentRoot
    try {
        $env:CGO_ENABLED = "0"
        $env:GOOS = "windows"
        $env:GOARCH = "amd64"
        go build -ldflags="-s -w" -o $exePath ./cmd/agent
        if ($LASTEXITCODE -ne 0) { throw "Go build failed" }
        Write-Host "   Built: $exePath" -ForegroundColor Green
    } finally {
        Pop-Location
    }
} else {
    if (-not (Test-Path $exePath)) {
        throw "No pre-built binary found at $exePath. Run without -SkipGoBuild."
    }
    Write-Host ">> Skipping Go build, using existing: $exePath" -ForegroundColor Yellow
}

# ── Step 2: Check for icon — generate a placeholder if missing ─────────────
$iconPath = Join-Path $agentRoot "assets\icon.ico"
if (-not (Test-Path $iconPath)) {
    Write-Host ">> No icon.ico found; installer will use default icon." -ForegroundColor Yellow
}

# ── Step 3: Run Inno Setup Compiler ────────────────────────────────────────
$issFile = Join-Path $agentRoot "installer\setup.iss"
$iscc    = Join-Path $env:LOCALAPPDATA "Programs\Inno Setup 6\ISCC.exe"

if (-not (Test-Path $iscc)) {
    # Try common paths
    $altPaths = @(
        "C:\Program Files (x86)\Inno Setup 6\ISCC.exe",
        "C:\Program Files\Inno Setup 6\ISCC.exe"
    )
    foreach ($p in $altPaths) {
        if (Test-Path $p) { $iscc = $p; break }
    }
}

if (-not (Test-Path $iscc)) {
    throw "Inno Setup compiler (ISCC.exe) not found. Install Inno Setup 6: https://jrsoftware.org/isdl.php"
}

Write-Host ">> Compiling installer with Inno Setup..." -ForegroundColor Cyan
Write-Host "   Enrollment key embedded: $($EnrollmentKey.Substring(0, [Math]::Min(4, $EnrollmentKey.Length)))****" -ForegroundColor DarkGray
& $iscc "/DEnrollmentKey=$EnrollmentKey" $issFile
if ($LASTEXITCODE -ne 0) { throw "Inno Setup compilation failed" }

$outputDir  = Join-Path $agentRoot "installer\output"
$installers = Get-ChildItem $outputDir -Filter "*.exe" | Sort-Object LastWriteTime -Descending | Select-Object -First 1
if ($installers) {
    Write-Host ""
    Write-Host ">> Installer ready: $($installers.FullName)" -ForegroundColor Green
    Write-Host "   Size: $([math]::Round($installers.Length / 1MB, 2)) MB" -ForegroundColor Green
} else {
    Write-Host ">> Build completed but installer not found in $outputDir" -ForegroundColor Yellow
}
