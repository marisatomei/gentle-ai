#Requires -Version 5.1
<#
.SYNOPSIS
    gentle-ai — Local Build & Install Script (Windows)
    Builds the binary from the current source tree and installs it locally.

.DESCRIPTION
    Use this when you are working on a local branch and want to test your
    changes with a real versioned binary — without publishing a GitHub Release.

    The script:
      1. Resolves a version string (argument, git tag, or fallback)
      2. Runs `go build` with that version injected via ldflags
      3. Installs the binary to $InstallDir (default: %LOCALAPPDATA%\gentle-ai\bin)
      4. Verifies the installed binary reports the expected version

.EXAMPLE
    # Build + install with auto-detected version
    .\scripts\build-local.ps1

    # Specify an explicit version
    .\scripts\build-local.ps1 -Version 1.2.0-copilot

    # Install to a custom directory
    .\scripts\build-local.ps1 -InstallDir "$env:USERPROFILE\.local\bin"
#>

[CmdletBinding()]
param(
    [string]$Version   = "",
    [string]$InstallDir = ""
)

$ErrorActionPreference = "Stop"

$BINARY_NAME = "gentle-ai"
$MAIN_PKG    = "./cmd/gentle-ai"
$VERSION_VAR = "main.version"

# ============================================================================
# Logging helpers (same style as install.ps1)
# ============================================================================

function Write-Info    { param([string]$Message) Write-Host "[info]    $Message" -ForegroundColor Blue }
function Write-Success { param([string]$Message) Write-Host "[ok]      $Message" -ForegroundColor Green }
function Write-Warn    { param([string]$Message) Write-Host "[warn]    $Message" -ForegroundColor Yellow }
function Write-Err     { param([string]$Message) Write-Host "[error]   $Message" -ForegroundColor Red }
function Write-Step    { param([string]$Message) Write-Host "`n==> $Message" -ForegroundColor Cyan }

function Stop-WithError {
    param([string]$Message)
    Write-Err $Message
    exit 1
}

# ============================================================================
# Banner
# ============================================================================

function Show-Banner {
    Write-Host ""
    Write-Host "   ____            _   _              _    ___ " -ForegroundColor Cyan
    Write-Host "  / ___| ___ _ __ | |_| | ___        / \  |_ _|" -ForegroundColor Cyan
    Write-Host " | |  _ / _ \ '_ \| __| |/ _ \_____ / _ \  | | " -ForegroundColor Cyan
    Write-Host " | |_| |  __/ | | | |_| |  __/_____/ ___ \ | | " -ForegroundColor Cyan
    Write-Host "  \____|\___|_| |_|\__|_|\___|    /_/   \_\___|" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "  Local Build & Install" -ForegroundColor DarkGray
    Write-Host ""
}

# ============================================================================
# Resolve version
# ============================================================================

function Resolve-BuildVersion {
    param([string]$Explicit)

    if ($Explicit -ne "") {
        # Strip leading 'v' for consistency
        $v = $Explicit.TrimStart("v")
        Write-Info "Using explicit version: $v"
        return $v
    }

    # Try the nearest git tag on the current branch
    if (Get-Command "git" -ErrorAction SilentlyContinue) {
        $tag = git describe --tags --abbrev=0 2>$null
        if ($LASTEXITCODE -eq 0 -and $tag) {
            $commitsSince = git rev-list "$tag..HEAD" --count 2>$null
            $shortHash    = git rev-parse --short HEAD 2>$null
            $v = $tag.TrimStart("v")
            if ($commitsSince -gt 0) {
                $v = "$v-dev.$commitsSince+$shortHash"
            }
            Write-Info "Resolved version from git: $v"
            return $v
        }
    }

    # Fallback
    $fallback = "0.0.0-local"
    Write-Warn "Could not resolve git version — using fallback: $fallback"
    return $fallback
}

# ============================================================================
# Prerequisites
# ============================================================================

function Test-Prerequisites {
    Write-Step "Checking prerequisites"

    if (-not (Get-Command "go" -ErrorAction SilentlyContinue)) {
        Stop-WithError "Go is not installed or not in PATH. Install it from https://golang.org/dl/"
    }

    $goVersion = (go version) -replace "go version ", ""
    Write-Success "Go found: $goVersion"
}

# ============================================================================
# Build
# ============================================================================

function Build-Binary {
    param([string]$BuildVersion, [string]$OutPath)

    Write-Step "Building $BINARY_NAME"
    Write-Info "Version : $BuildVersion"
    Write-Info "Output  : $OutPath"

    $ldflags = "-s -w -X ${VERSION_VAR}=${BuildVersion}"

    & go build -ldflags $ldflags -o $OutPath $MAIN_PKG
    if ($LASTEXITCODE -ne 0) {
        Stop-WithError "go build failed"
    }

    $size = [math]::Round((Get-Item $OutPath).Length / 1KB, 0)
    Write-Success "Built successfully (${size} KB)"
}

# ============================================================================
# Install
# ============================================================================

function Install-Binary {
    param([string]$SourcePath, [string]$Dir)

    Write-Step "Installing binary"

    if (-not $Dir) {
        $Dir = Join-Path $env:LOCALAPPDATA "gentle-ai\bin"
    }

    if (-not (Test-Path $Dir)) {
        New-Item -ItemType Directory -Path $Dir -Force | Out-Null
        Write-Info "Created directory: $Dir"
    }

    $destPath = Join-Path $Dir "$BINARY_NAME.exe"

    # If the binary is currently running, the copy will fail — warn clearly
    Copy-Item -Path $SourcePath -Destination $destPath -Force
    Write-Success "Installed to: $destPath"

    # Advise on PATH if not already present
    $userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($userPath -notlike "*$Dir*") {
        Write-Warn "$Dir is not in your user PATH."
        Write-Host ""
        Write-Host "  Run this once to add it permanently:" -ForegroundColor DarkGray
        Write-Host "  [Environment]::SetEnvironmentVariable('PATH', `$env:PATH + ';$Dir', 'User')" -ForegroundColor DarkGray
        Write-Host ""
    }

    return $destPath
}

# ============================================================================
# Verify
# ============================================================================

function Test-Installation {
    param([string]$BinaryPath, [string]$ExpectedVersion)

    Write-Step "Verifying installation"

    if (-not (Test-Path $BinaryPath)) {
        Stop-WithError "Binary not found at $BinaryPath"
    }

    $output = & $BinaryPath version 2>&1
    Write-Success "$BINARY_NAME version output: $output"

    if ($output -notlike "*$ExpectedVersion*") {
        Write-Warn "Reported version does not match expected '$ExpectedVersion' — check ldflags"
    }
}

# ============================================================================
# Main
# ============================================================================

function Main {
    Show-Banner

    Test-Prerequisites

    Write-Step "Resolving version"
    $buildVersion = Resolve-BuildVersion -Explicit $Version

    $tmpBinary = Join-Path $env:TEMP "$BINARY_NAME-local-build.exe"

    Build-Binary -BuildVersion $buildVersion -OutPath $tmpBinary

    $installedPath = Install-Binary -SourcePath $tmpBinary -Dir $InstallDir

    Remove-Item $tmpBinary -Force -ErrorAction SilentlyContinue

    Test-Installation -BinaryPath $installedPath -ExpectedVersion $buildVersion

    Write-Host ""
    Write-Host "Done! Local build $buildVersion is installed." -ForegroundColor Green
    Write-Host "Run: $BINARY_NAME" -ForegroundColor Cyan
    Write-Host ""
}

Main
