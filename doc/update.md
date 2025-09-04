---
# Page settings
layout: default
keywords: kat,postgres,database,cli,migrations,sql,update,upgrade
title: Updating Kat
description: |
    How to update Kat to the latest version safely.
comments: false
permalink: /update/
page_nav:
    prev:
        content: Graph Visualization
        url: '/export'
    next:
        content: Custom Logging
        url: '/logger'
---

# Updating Kat

Keep your Kat installation up to date to get the latest features, bug fixes, and security updates.

## Check Current Version

First, check which version you currently have:

```bash
kat version
```

## Update Methods

### Using the Install Script (Recommended)

The easiest way to update Kat is to re-run the install script:

```bash
curl -sSL https://kat.bolaji.de/install | sudo bash
```

This will download and install the latest version, replacing your current installation.

### Manual Update

1. Visit the [GitHub Releases page](https://github.com/BolajiOlajide/kat/releases)
2. Download the latest release for your platform
3. Replace your existing `kat` binary with the new one

### Update from Source

If you installed from source:

```bash
cd /path/to/kat-source
git pull origin main
make install
```

## Version Compatibility

Kat follows [Semantic Versioning](https://semver.org/):

- **Patch releases** (1.0.1 → 1.0.2): Bug fixes, fully backward compatible
- **Minor releases** (1.0.x → 1.1.0): New features, backward compatible
- **Major releases** (1.x.x → 2.0.0): Breaking changes, migration guide provided

## Migration Compatibility

Your existing migrations will continue to work across all Kat updates. The migration file format and database schema are stable APIs that won't change without a major version bump.

## Backup Recommendations

Before updating in production environments:

1. **Test the update** in a non-production environment first
2. **Backup your migration tracking table**: 
   ```sql
   pg_dump -t migrations your_database > migrations_backup.sql
   ```
3. **Dry run your migrations** after updating:
   ```bash
   kat up --dry-run
   ```

## Troubleshooting Updates

If you encounter issues after updating:

1. **Check the changelog** for breaking changes
2. **Verify your configuration** is still valid
3. **Test database connectivity**: `kat ping`
4. **Revert to previous version** if needed and report the issue

## Getting Help

- 📋 [Changelog](https://github.com/BolajiOlajide/kat/releases) - See what's new
- 💬 [GitHub Discussions](https://github.com/BolajiOlajide/kat/discussions) - Ask questions  
- 🐛 [GitHub Issues](https://github.com/BolajiOlajide/kat/issues) - Report problems

---

> 💡 **Tip**: Pin Kat to a specific version in CI/CD environments to ensure reproducible builds.
