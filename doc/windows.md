---
# Page settings
layout: default
keywords: kat,postgres,database,cli,migrations,sql,windows
title: Windows Guide
description: |
    Complete guide for using Kat on Windows with PowerShell, WSL, and common troubleshooting.
comments: false
permalink: /windows/
---

# Using Kat on Windows

This guide covers everything Windows users need to know to get Kat working smoothly on their systems.

## Installation Options

### PowerShell (Recommended)

The easiest installation method for Windows:

```powershell
# Install latest version
iex (iwr https://kat.bolaji.de/install.ps1).Content

# Install specific version
$env:VERSION="v1.0.0"; iex (iwr https://kat.bolaji.de/install.ps1).Content

# Install to custom directory
$env:INSTALL_DIR="$env:USERPROFILE\tools"; iex (iwr https://kat.bolaji.de/install.ps1).Content
```

### Package Managers

**Scoop (Recommended for developers):**
```powershell
scoop bucket add bolaji https://github.com/BolajiOlajide/scoop-bucket
scoop install bolaji/kat
```

**Chocolatey:**
```powershell
choco install kat
```

### Manual Installation

1. Download from [GitHub Releases](https://github.com/BolajiOlajide/kat/releases)
2. Extract `kat_windows_amd64.zip`
3. Move `kat.exe` to a directory in your PATH
4. Verify: `kat version`

## PostgreSQL Setup on Windows

### Using PostgreSQL Installer

1. Download from [postgresql.org](https://www.postgresql.org/download/windows/)
2. Install with default settings
3. Note your password and port (usually 5432)

### Using Docker Desktop

```powershell
# Start PostgreSQL container
docker run --name kat-postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=myapp -p 5432:5432 -d postgres:16

# Verify it's running
docker ps
```

## Configuration Examples

### Basic Windows Configuration

```yaml
# kat.conf.yaml
migration:
  tablename: migrations
  directory: migrations
database:
  user: postgres
  password: your_password_here
  host: localhost
  port: 5432
  name: myapp
  sslmode: disable
```

### Using Environment Variables (Windows)

```powershell
# Set environment variables in PowerShell
$env:DB_USER="postgres"
$env:DB_PASSWORD="your_password"
$env:DB_NAME="myapp"

# Configure Kat to use them
```

```yaml
# kat.conf.yaml
database:
  user: ${DB_USER}
  password: ${DB_PASSWORD}
  name: ${DB_NAME}
  host: localhost
  port: 5432
  sslmode: disable
```

### WSL (Windows Subsystem for Linux)

If you're using WSL, you can follow the Linux installation instructions:

```bash
# In WSL terminal
curl -sSL https://kat.bolaji.de/install | bash

# Connect to Windows PostgreSQL from WSL
# Use host: localhost or your Windows IP
```

## Common Windows Issues & Solutions

### PowerShell Execution Policy

If you get execution policy errors:

```powershell
# Check current policy
Get-ExecutionPolicy

# Set policy for current user (if needed)
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### PATH Issues

If `kat` command is not found after installation:

```powershell
# Check if kat.exe is in PATH
where kat

# Add to PATH manually if needed
$env:PATH += ";C:\Users\YourUser\AppData\Local\kat\bin"

# Make it permanent
[Environment]::SetEnvironmentVariable("Path", $env:PATH, "User")
```

### PostgreSQL Connection Issues

**Issue**: "pq: password authentication failed"
```yaml
# Solution: Update your configuration with correct credentials
database:
  user: postgres
  password: your_actual_password
```

**Issue**: "pq: database does not exist"
```powershell
# Solution: Create the database first
createdb myapp

# Or using psql
psql -U postgres -c "CREATE DATABASE myapp;"
```

**Issue**: Connection refused
- Ensure PostgreSQL service is running: `Services.msc` → PostgreSQL
- Check if port 5432 is accessible: `netstat -an | findstr 5432`

### File Permission Issues

Windows doesn't use Unix permissions, but Kat handles this automatically. If you encounter permission errors:

1. Run PowerShell as Administrator
2. Ensure antivirus isn't blocking kat.exe
3. Check that installation directory is writable

## Development on Windows

### Building from Source

```powershell
# Clone and build
git clone https://github.com/BolajiOlajide/kat.git
cd kat

# Build using Go (works on Windows)
go build ./cmd/kat

# Or using mise (if installed)
mise run build
```

### Testing

```powershell
# Run all tests
go test ./...

# Run tests with Docker (requires Docker Desktop)
docker run --rm -v ${PWD}:C:\app -w C:\app golang:1.23 go test ./...
```

## Docker Integration

### Using with Docker Compose

```yaml
# docker-compose.yml for Windows
version: '3.8'
services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  kat:
    image: mcr.microsoft.com/powershell:latest
    depends_on:
      - postgres
    volumes:
      - .:/app:ro
    working_dir: /app
    command: |
      pwsh -c "
        # Install Kat
        iex (iwr https://kat.bolaji.de/install.ps1).Content
        
        # Wait for database and apply migrations
        kat ping --retry-count 10
        kat up
      "

volumes:
  postgres_data:
```

## Best Practices for Windows

1. **Use PowerShell Core** (pwsh) instead of Windows PowerShell for better compatibility
2. **Set UTF-8 encoding**: `$OutputEncoding = [System.Text.UTF8Encoding]::new()`
3. **Use forward slashes** in config file paths when possible: `directory: "migrations"`
4. **Escape backslashes** in YAML strings: `directory: "C:\\Users\\YourUser\\migrations"`
5. **Use environment variables** for sensitive data instead of hardcoding in config files

## VS Code Integration

Add these tasks to `.vscode/tasks.json`:

```json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "Kat: Apply Migrations",
            "type": "shell",
            "command": "kat",
            "args": ["up"],
            "group": "build",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            }
        },
        {
            "label": "Kat: Create Migration",
            "type": "shell",
            "command": "kat",
            "args": ["add", "${input:migrationName}"],
            "group": "build"
        }
    ],
    "inputs": [
        {
            "id": "migrationName",
            "description": "Migration name",
            "default": "new_migration",
            "type": "promptString"
        }
    ]
}
```

## Troubleshooting

### Windows-Specific Error Messages

| Error | Cause | Solution |
|-------|-------|----------|
| "Access is denied" | Permission issue | Run PowerShell as Administrator |
| "The system cannot find the path specified" | PATH not updated | Restart terminal or add to PATH manually |
| "pq: SSL is not enabled on the server" | SSL mismatch | Set `sslmode: disable` in config |

### Getting Help

- 📖 [General Documentation](https://kat.bolaji.de/)
- 💬 [GitHub Discussions](https://github.com/BolajiOlajide/kat/discussions) - Tag posts with `windows`
- 🐛 [Report Windows Issues](https://github.com/BolajiOlajide/kat/issues/new?labels=windows)

---

> 💡 **Tip**: Most Windows issues stem from PATH or PostgreSQL connection configuration. Check these first before reporting bugs.
