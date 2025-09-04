---
# Page settings
layout: default
keywords: kat,postgres,database,cli,migrations,sql,quick-start,tutorial
title: Quick Start Guide
description: |
    Get up and running with Kat in 5 minutes. Learn the basics of graph-based PostgreSQL migrations.
comments: false
permalink: /quickstart/
page_nav:
    prev:
        content: Installation
        url: '/installation'
    next:
        content: Concepts
        url: '/concepts'
---

# Quick Start Guide

Get up and running with Kat in 5 minutes. This guide walks you through your first migration using Kat's graph-based approach.

## Prerequisites

- PostgreSQL database running locally or accessible remotely
- Kat installed ([Installation Guide](/installation/))

## Step 1: Set Up Database

Create a PostgreSQL database for this tutorial:

```bash
# Create the database (requires PostgreSQL installed locally)
createdb myapp

# Or using SQL:
psql -c "CREATE DATABASE myapp;"
```

## Step 2: Initialize Your Project

Create a new directory and initialize Kat:

```bash
mkdir my-app && cd my-app
kat init
```

This creates a `kat.conf.yaml` configuration file:
```yaml
migration:
  tablename: migrations
  directory: migrations
database:
  user: postgres
  password: postgres
  host: localhost
  port: 5432
  name: myapp
  sslmode: disable
```

> 💡 **Edit this file** if your database credentials differ from the defaults.

## Step 3: Test Database Connection

Verify Kat can connect to your database:

```bash
kat ping
```

Expected output:
```
Attempting to ping database
Successfully connected to database!
```

## Step 4: Create Your First Migration

Create a migration to set up a users table:

```bash
kat add create_users_table
```

This generates:
```
migrations/
└── 1747834567_create_users_table/
    ├── up.sql
    ├── down.sql
    └── metadata.yaml
```

## Step 5: Write Your Migration SQL

Edit the generated SQL files:

**up.sql:**
```sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
```

**down.sql:**
```sql
DROP INDEX IF EXISTS idx_users_username;
DROP TABLE IF EXISTS users;
```

## Step 6: Apply Your Migration

Run the migration:

```bash
kat up
```

Expected output:
```
Attempting to ping database
Successfully connected to database!
1747834567_create_users_table
Successfully applied 1 migrations

Migration Summary
Successful migrations:
  ✓ 1747834567_create_users_table (up) - 12.345ms

Total: 1 migration(s) applied
```

## Step 7: Create a Dependent Migration

Now create a migration that depends on the users table:

```bash
kat add add_user_profiles --parent 1747834567
```

Edit the new migration files:

**up.sql:**
```sql
CREATE TABLE IF NOT EXISTS user_profiles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bio TEXT,
    avatar_url VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**down.sql:**
```sql
DROP TABLE IF EXISTS user_profiles;
```

## Step 8: Apply All Migrations

```bash
kat up
```

Kat automatically determines the correct order based on dependencies!

## Step 9: Visualize Your Migration Graph (Optional)

```bash
kat export --file graph.dot
```

If you have Graphviz installed, generate a visual:
```bash
# Install Graphviz first (if needed)
# macOS: brew install graphviz
# Ubuntu: apt-get install graphviz  
# Windows: choco install graphviz

dot -Tpng graph.dot -o migrations.png
open migrations.png  # macOS
```

## Step 10: Test Rollback

Roll back the most recent migration:

```bash
kat down --count 1
```

Expected output:
```
Attempting to ping database
Successfully connected to database!
1747834589_add_user_profiles
Successfully rolled back 1 migrations

Migration Summary
Successful migrations:
  ✓ 1747834589_add_user_profiles (down) - 8.123ms

Total: 1 migration(s) rolled back
```

## What Makes This Different?

Unlike traditional migration tools, Kat uses a **Directed Acyclic Graph (DAG)** to manage dependencies:

- **Traditional tools**: Linear sequence (001 → 002 → 003)
- **Kat**: Graph-based with explicit dependencies

This means:
- Multiple developers can create migrations in parallel
- Kat automatically determines the correct execution order
- Complex dependencies are handled safely
- Branch merges don't require renumbering

## Next Steps

Now that you've created your first migrations:

1. **[Learn the concepts](/concepts/)** behind Kat's graph-based approach
2. **[Explore migration patterns](/migration/)** for common database tasks
3. **[Set up CI/CD integration](/migration/#integration-with-cicd-pipelines)** for automated deployments

## Docker Compose Example

Want to try Kat without installing PostgreSQL locally? Here's a complete setup:

```yaml
# docker-compose.yml
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
    
  kat:
    image: alpine:latest
    depends_on:
      - postgres
    volumes:
      - .:/app
    working_dir: /app
    command: |
      sh -c "
        # Install Kat in container
        curl -sSL https://kat.bolaji.de/install | KAT_INSTALL_DIR=/usr/local/bin sh
        
        # Wait for DB and run migrations
        kat ping --retry-count 10 --retry-delay 2000
        kat up
      "
```

Run with: `docker-compose up`

## Need Help?

- 📖 [Full Documentation](https://kat.bolaji.de/)
- 💬 [GitHub Discussions](https://github.com/BolajiOlajide/kat/discussions)
- 🐛 [Report Issues](https://github.com/BolajiOlajide/kat/issues)
