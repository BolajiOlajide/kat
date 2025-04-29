---
# Page settings
layout: default
keywords: kat,postgres,database,cli,migrations,sql,update
title: Update
description: |
    Learn how to update Kat to the latest version.
comments: false
permalink: /update/
page_nav:
    prev:
        content: Migrations
        url: '/migration'
    next:
        content: Contributing
        url: '/contributing'
---

# Update

The `kat update` command checks for and installs updates to the Kat CLI tool.

## Usage

```bash
kat update
```

## Description

The update command performs the following actions:

1. Checks GitHub for newer versions of Kat
2. Compares the current version with the latest available version
3. If an update is available, downloads the appropriate binary for your platform
4. Replaces the current binary with the new version

## Options

Currently, the update command does not accept any additional options.

## Examples

```bash
# Check for and install updates
kat update
```

## Notes

- The command requires internet connectivity to check for updates
- The update process will replace your current Kat binary with the newer version
- If no update is available, the command will inform you that you're already using the latest version
- The command automatically detects your operating system and architecture to download the correct binary