---
# Page settings
layout: default
keywords: kat,postgres,database,cli,migrations,sql,editor,vscode,zed,yaml,schema
title: Editor Support
description: |
    Get autocompletion and validation for Kat's YAML files in VS Code and Zed.
comments: false
permalink: /editor-support/
page_nav:
    prev:
        content: Configuration
        url: '/config'
    next:
        content: Database Connectivity
        url: '/ping'
---

# Editor Support

Kat ships JSON Schemas for both of its YAML file formats — `kat.conf.yaml` and migration `metadata.yaml`. Editors that understand these schemas will give you autocompletion, inline documentation, and validation as you type.

## What the Schemas Cover

### `kat.conf.yaml`

- `migration.tablename` and `migration.directory` with their default values
- `database` as a choice between a connection `url` or individual credential fields (`host`, `port`, `user`, `password`, `name`, `sslmode`)
- `sslmode` as an enum with all valid PostgreSQL values
- `verbose` flag

### `metadata.yaml`

- Required fields: `name` and `timestamp`
- Optional fields: `description`, `parents` (array of timestamps), and `no_transaction`
- Descriptions for each field explaining their purpose

---

## VS Code

Install the [YAML extension by Red Hat](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml) if you don't have it already. It powers YAML schema support in VS Code.

The project-level `.vscode/settings.json` in this repository already wires the schemas to the right file patterns, so if you open the Kat repository itself everything works out of the box.

For **your own project** that uses Kat, create a `.vscode/settings.json` at the root of your project and add:

```json
{
  "yaml.schemas": {
    "https://raw.githubusercontent.com/BolajiOlajide/kat/main/schemas/kat.conf.schema.json": [
      "kat.conf.yaml"
    ],
    "https://raw.githubusercontent.com/BolajiOlajide/kat/main/schemas/metadata.schema.json": [
      "**/migrations/**/metadata.yaml"
    ]
  }
}
```

---

## Zed

Zed uses `yaml-language-server` under the hood. Add the following to your project's `.zed/settings.json`:

```json
{
  "lsp": {
    "yaml-language-server": {
      "settings": {
        "yaml": {
          "schemas": {
            "https://raw.githubusercontent.com/BolajiOlajide/kat/main/schemas/kat.conf.schema.json": [
              "kat.conf.yaml"
            ],
            "https://raw.githubusercontent.com/BolajiOlajide/kat/main/schemas/metadata.schema.json": [
              "**/migrations/**/metadata.yaml"
            ]
          }
        }
      }
    }
  }
}
```

---

## Any Editor with yaml-language-server Support

If your editor supports `yaml-language-server` (Neovim, Helix, Emacs, etc.), you can add a modeline comment to the top of any Kat YAML file to associate the schema directly:

**`kat.conf.yaml`:**
```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/BolajiOlajide/kat/main/schemas/kat.conf.schema.json
migration:
  tablename: migrations
  directory: migrations
database:
  url: postgres://user:password@localhost/mydb
```

**`metadata.yaml`:**
```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/BolajiOlajide/kat/main/schemas/metadata.schema.json
name: create_users_table
timestamp: 1747578808
```

This approach works regardless of which editor or LSP client you use, and doesn't require any editor configuration.
