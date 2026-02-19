---
# Page settings
layout: default
keywords: kat,postgres,database,cli,migrations,sql,dag,graph,concepts
title: Core Concepts
description: |
    Understand the key concepts behind Kat's graph-based migration system.
comments: false
permalink: /concepts/
page_nav:
    prev:
        content: Quick Start
        url: '/quickstart'
    next:
        content: Configuration
        url: '/config'
---

# Core Concepts

Understanding these key concepts will help you get the most out of Kat's graph-based migration system.

## What is a Migration?

A **migration** is a versioned change to your database schema. In Kat, each migration consists of:

- **Timestamp ID**: Unique identifier (Unix timestamp when created)
- **Name**: Human-readable description (e.g., `create_users_table`)
- **Up SQL**: Commands to apply the change
- **Down SQL**: Commands to reverse the change
- **Dependencies**: Optional parent migrations that must run first

## Directed Acyclic Graph (DAG)

Kat organizes migrations as a **Directed Acyclic Graph**:

- **Directed**: Migrations have a clear flow direction (parent → child)
- **Acyclic**: No circular dependencies allowed
- **Graph**: Migrations can have multiple parents and children

### Visual Example

```text
Traditional (Linear):           Kat (Graph):
                               
001_users                      001_users ──┬─→ 003_posts
002_posts                                  │
003_comments                   002_profiles ──┴─→ 004_comments
004_tags                                         │
                                               005_tags
Order: 1→2→3→4                 Order: 1→(2,3)→4→5
(rigid)                        (flexible, dependency-aware)
```

## Key Terms

| Term | Definition | Example |
|------|------------|---------|
| **Migration** | A versioned database schema change | `1679012345_create_users_table` |
| **Parent** | A migration that must run before another | `create_users_table` is parent of `add_email_column` |
| **Child** | A migration that depends on another | `add_email_column` is child of `create_users_table` |
| **Leaf** | A migration with no dependencies | First migration in your project |
| **Topological Sort** | Algorithm that determines safe execution order | Ensures parents run before children |
| **Vertex** | A single migration in the graph | Each migration directory |
| **Edge** | A dependency relationship | Parent-child connection |

## How Kat Differs from Traditional Tools

### Traditional Migration Tools
- **Linear sequence**: Migrations must be numbered sequentially
- **Coordination required**: Developers must coordinate to avoid conflicts
- **Rigid ordering**: Cannot change execution order after creation
- **Branch conflicts**: Feature branches create merge conflicts

### Kat's Graph-Based Approach
- **Explicit dependencies**: Declare what each migration actually needs
- **Parallel development**: Multiple developers work independently
- **Flexible ordering**: Kat computes optimal execution sequence
- **Clean merges**: No renumbering or conflicts when merging branches

## Migration States

Kat tracks migrations in three states:

1. **Pending**: Migration exists in filesystem but not applied to database
2. **Applied**: Migration has been executed and recorded in tracking table  
3. **Rolled back**: Migration was applied but then reversed

## Dependency Rules

When creating migrations with dependencies:

1. **Parents must exist**: Referenced parent migrations must be in your migration directory
2. **No cycles**: Cannot create circular dependencies (A → B → C → A)
3. **Multiple parents**: A migration can depend on multiple parent migrations
4. **Execution order**: Kat uses topological sorting to determine safe order

## Best Practices

### 🟢 Do
- Make migrations **idempotent** (safe to run multiple times)
- Use `IF EXISTS` and `IF NOT EXISTS` clauses
- Declare **explicit dependencies** via parents
- Keep migrations **focused** on a single logical change
- **Test both directions** (up and down)

### 🔴 Don't
- Create **circular dependencies**
- Make **destructive changes** without proper down migrations
- **Modify applied migrations** (create new ones instead)
- **Skip parent declarations** for dependent changes

## Transaction Behavior

- Each migration runs in its **own transaction** by default
- If a migration fails, the transaction **automatically rolls back**
- The database remains in its **previous state**
- **No partial applications** - migrations are all-or-nothing

### Non-Transactional Migrations

Some operations like `CREATE INDEX CONCURRENTLY` cannot run inside a transaction. Set `no_transaction: true` in `metadata.yaml` to opt out:

```yaml
name: add_index_concurrently
timestamp: 1679012345
no_transaction: true
```

> ⚠️ Non-transactional migrations have no automatic rollback on failure. See [Migration docs](/migration/#non-transactional-migrations) for details.

## Migration Tracking

Kat maintains a tracking table (default: `migrations`) with:

| Column | Type | Purpose |
|--------|------|---------|
| `id` | SERIAL | Auto-increment ID |
| `name` | VARCHAR | Migration identifier |
| `migration_time` | TIMESTAMP | When migration was applied |
| `duration` | INTEGER | Execution time in milliseconds |

This table enables Kat to:
- Skip already-applied migrations
- Determine which migrations to roll back
- Provide execution history and timing

## Common Patterns

### Feature Branch Workflow
```bash
# Main branch
git checkout main
kat add create_users_table

# Feature branch
git checkout -b feature/profiles
kat add create_profiles_table  # Kat determines create_users_table as parent

# Another feature branch  
git checkout -b feature/posts
kat add create_posts_table  # Creates parallel branch from users table

# Merge both - Kat handles the dependency resolution automatically
```

### Complex Dependencies
```bash
# Base tables
kat add create_users_table
kat add create_products_table

# For complex dependencies, edit metadata.yaml manually:
kat add create_orders_table
# Then edit migrations/TIMESTAMP_create_orders_table/metadata.yaml:
# parents: [1679012345, 1679012350]  # users and products timestamps
```

## Understanding Error Messages

Common error patterns and solutions:

| Error Type | Cause | Solution |
|------------|-------|----------|
| "Cycle detected" | Circular dependency created | Remove the circular reference |
| "Parent not found" | Referenced parent doesn't exist | Check parent migration ID |
| "SQL error" | Invalid SQL in migration | Fix the SQL syntax |
| "Connection failed" | Database connectivity issue | Check connection details |

## Next Steps

Now that you understand the concepts:

1. **[Configure your database](/config/)** connection details
2. **[Learn migration commands](/migration/)** for day-to-day usage  
3. **[Explore advanced patterns](/migration/#advanced-migration-patterns)** for complex scenarios

---

> 💡 **Tip**: Think of Kat migrations like Git commits - each has dependencies (parents) and forms a graph that can be traversed safely.
