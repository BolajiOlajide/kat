---
# Page settings
layout: default
keywords: kat,postgres,database,cli,migrations,sql
title: Kat - Configuration
description: |
    Kat is a PostgreSQL database migration tool. It allows you run your migrations with raw SQL files.
comments: false
page_nav:
    prev:
        content: Installation
        url: '/install'
    next: 
        content: Migration
        url: '/migration'
---

Kat requires a configuration file to know where to look for migrations and to know which database to perform migrations on.

```yaml
migration:
    tablename: migration_logs
    directory: /Users/bolaji/Desktop/kat-test/migrations
database:
    url: postgres://bolaji:andela@localhost:5432/katest
```
