# PowerShell installation script for Kat on Windows
param(
    [string]$Version = "",
    [string]$InstallDir = "$env:LOCALAPPDATA\kat\bin"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# Ensure TLS 1.2+ for older PowerShell versions
$tls12 = [Net.SecurityProtocolType]::Tls12
try {
    # TLS 1.3 support (only available in newer .NET versions)
    $tls13 = [Net.SecurityProtocolType]::Tls13
    [Net.ServicePointManager]::SecurityProtocol = $tls12 -bor $tls13
}
catch {
    # Fallback to TLS 1.2 only for older PowerShell
    [Net.ServicePointManager]::SecurityProtocol = $tls12
}

function Write-Info {
    param([string]$Message)
    Write-Host $Message -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host $Message -ForegroundColor Green
}

function Write-Error {
    param([string]$Message)
    Write-Host $Message -ForegroundColor Red
    exit 1
}

# Check if running as Administrator for system-wide install
function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# Get latest version if not specified
if (-not $Version) {
    Write-Info "No version specified, fetching latest release..."
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/BolajiOlajide/kat/releases/latest"
        $Version = $response.tag_name
        Write-Info "Latest version: $Version"
    }
    catch {
        Write-Error "Failed to fetch latest version: $($_.Exception.Message)"
    }
}

# Detect architecture
$arch = "amd64"  # Default to amd64
if ($env:PROCESSOR_ARCHITEW6432 -eq "ARM64" -or $env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
    $arch = "arm64"
} elseif (-not [Environment]::Is64BitOperatingSystem) {
    $arch = "386"
}
Write-Info "Detected architecture: $arch"

# Set download URL
$downloadUrl = "https://github.com/BolajiOlajide/kat/releases/download/$Version/kat_windows_$arch.zip"

# Test if download URL exists
try {
    $response = Invoke-WebRequest -Uri $downloadUrl -Method Head
}
catch {
    Write-Error "Binary for windows on $arch architecture not found for version $Version. Please check available releases at: https://github.com/BolajiOlajide/kat/releases"
}

# Create installation directory
if (-not (Test-Path $InstallDir)) {
    Write-Info "Creating installation directory: $InstallDir"
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# Download and extract
$tempFile = [System.IO.Path]::GetTempFileName() + ".zip"
try {
    Write-Info "Downloading kat $Version for Windows ($arch)..."
    Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile
    
    Write-Info "Extracting to $InstallDir..."
    Expand-Archive -Path $tempFile -DestinationPath $InstallDir -Force
    
    # Verify installation
    $katPath = Join-Path $InstallDir "kat.exe"
    if (Test-Path $katPath) {
        Write-Success "kat $Version has been installed to '$InstallDir'"
        
        # Add to PATH if not already there
        $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
        $normalizedInstallDir = $InstallDir.TrimEnd('\', '/')
        $pathEntries = $userPath -split ';' | ForEach-Object { $_.TrimEnd('\', '/') }
        
        if ($pathEntries -notcontains $normalizedInstallDir) {
            Write-Info "Adding $InstallDir to your PATH..."
            $newPath = if ($userPath) { "$userPath;$InstallDir" } else { $InstallDir }
            [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
            Write-Info "Please restart your terminal to use the updated PATH"
        }
        
        # Verify installation
        try {
            $versionOutput = & $katPath --version 2>$null
            Write-Success "Installation verified: $versionOutput"
        }
        catch {
            Write-Info "Installation complete but verification failed. You may need to restart your terminal."
        }
        
        Write-Success "Installation complete! Run 'kat --help' to get started."
    }
    else {
        Write-Error "Installation failed: kat.exe not found in extracted files"
    }
}
finally {
    # Clean up
    if (Test-Path $tempFile) {
        Remove-Item $tempFile
    }
}
