---
# Page settings
layout: default
keywords: kat,postgres,database,cli,migrations,sql
title: Installation
description: |
    Kat is a PostgreSQL database migration tool. It allows you run your migrations with raw SQL files.
comments: false
permalink: /installation/
page_nav:
    next:
        content: Initialization
        url: '/init'
---

# Installing Kat

Kat is a CLI tool for performing PostgreSQL database migrations. This guide covers different methods to install Kat on your system.

## Prerequisites

Before installing Kat, ensure your system meets the following requirements:

- **PostgreSQL**: Kat is designed to work with PostgreSQL databases
- **Go**: Version 1.20 or higher (only required for building from source)

## Installation Methods

### Using the Install Script (Recommended)

For macOS and Linux, the easiest way to install Kat is using the install script:

```bash
# Install latest version (automatically fetches the latest release)
curl -sSL https://kat.bolaji.de/install | bash

# Install specific version
curl -sSL https://kat.bolaji.de/install | VERSION=v1.0.0 bash
```

This will:
1. Detect your operating system (macOS or Linux)
2. Fetch the latest release version from GitHub if no version is specified
3. Download the appropriate pre-compiled binary from GitHub Releases
4. Install it to `/usr/local/bin`, making it available in your PATH

### Manual Installation from Pre-compiled Binaries

You can also download and install the binary manually:

1. Visit the [GitHub Releases page](https://github.com/BolajiOlajide/kat/releases)
2. Download the appropriate archive for your operating system (replace `[VERSION]` with the version you want, e.g. `v1.0.0`):
   - macOS: `kat_[VERSION]_darwin_amd64.tar.gz`
   - Linux: `kat_[VERSION]_linux_amd64.tar.gz`
3. Extract the binary:
   ```bash
   tar -xzf kat_[VERSION]_[os]_amd64.tar.gz
   ```
4. Move the binary to a location in your PATH:
   ```bash
   sudo mv kat /usr/local/bin/
   ```
5. Make it executable:
   ```bash
   sudo chmod +x /usr/local/bin/kat
   ```

### Installing from Source

If you prefer to build from source or need to customize the installation:

1. Clone the repository:
   ```bash
   git clone https://github.com/BolajiOlajide/kat.git
   cd kat
   ```

2. Install using make:
   ```bash
   make install
   ```

   This runs `go install ./...`, which compiles and installs the binary to your Go bin directory.

3. Alternatively, you can run:
   ```bash
   go install github.com/BolajiOlajide/kat/cmd/kat@latest
   ```

   This will download, compile, and install the latest version directly.

## Verifying the Installation

To verify that Kat was installed correctly, run:

```bash
kat version
```

You should see output showing the version of Kat that you installed.

## Troubleshooting

### Common Issues

1. **Command not found**
   - Ensure the installation directory is in your PATH
   - For Go installations, make sure `$GOPATH/bin` is in your PATH

2. **Permission denied**
   - Make sure the binary is executable: `chmod +x /path/to/kat`
   - You might need to use `sudo` for some installation steps

3. **Installation fails**
   - Check your Go version: `go version`
   - Ensure you have internet access to download dependencies

### Getting Help

If you encounter any issues during installation:
- Check the [GitHub Issues](https://github.com/BolajiOlajide/kat/issues) to see if others have faced similar problems
- Open a new issue with details about your environment and the error message

## Next Steps

After successfully installing Kat, the next step is to [initialize](/init/) your project with Kat's configuration and directory structure.