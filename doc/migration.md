---
# Page settings
layout: default
keywords: kat,postgres,database,cli,migrations,sql
title: Migrations
description: |
    Learn how to create, apply, and roll back database migrations with Kat.
comments: false
permalink: /migration/
page_nav:
    prev:
        content: Database Connectivity
        url: '/ping'
    next:
        content: Exporting Migrations
        url: '/export'
---

# Working with Migrations in Kat

Migrations are the core functionality of Kat, allowing you to version control your database schema and make changes in a controlled, repeatable manner. This guide covers everything you need to know about creating, applying, and rolling back migrations.

## Migration Concepts

In Kat, migrations follow these principles:

1. **Versioned**: Each migration has a unique timestamp identifier
2. **Directional**: Migrations can move forward (up) or backward (down)
3. **Dependency-Based**: Dependencies between migrations are automatically calculated
4. **Graph-Based**: Migrations form a directed acyclic graph (DAG) for proper ordering
5. **Tracked**: Applied migrations are recorded in a database table
6. **Idempotent**: Well-written migrations can be run multiple times safely

## Migration Structure

Each migration in Kat consists of three files organized in a directory structure:

```
migrations/
  ├─ 1679012345_create_users/
  │   ├─ up.sql      # SQL commands to apply the migration
  │   ├─ down.sql    # SQL commands to revert the migration
  │   └─ metadata.yaml  # Migration metadata (name, timestamp)
  ├─ 1679023456_add_email_column/
  │   ├─ up.sql
  │   ├─ down.sql
  │   └─ metadata.yaml
  └─ ...
```

### Migration Files

- **up.sql**: Contains SQL statements to apply the migration (create tables, add columns, etc.)
- **down.sql**: Contains SQL statements to reverse the migration (drop tables, remove columns, etc.)
- **metadata.yaml**: Contains metadata about the migration:
  ```yaml
  name: 1679012345_create_users
  timestamp: 1679012345
  ```

## Creating Migrations

To create a new migration, use the `add` command:

```bash
kat add create_users_table
```

This generates a new migration with the following files:

```
migrations/
  └─ 1679012345_create_users_table/
      ├─ up.sql
      ├─ down.sql
      └─ metadata.yaml
```

The timestamp ensures migrations are applied in the correct order. The name is sanitized (lowercase, spaces replaced with underscores, non-alphanumeric characters removed).

### Writing Migration SQL

After creating the migration files, you'll need to edit them with your specific SQL commands:

**up.sql example**:
```sql
-- Create users table
CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  username VARCHAR(100) NOT NULL UNIQUE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Add indexes
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
```

**down.sql example**:
```sql
-- Drop indexes
DROP INDEX IF EXISTS idx_users_username;

-- Drop users table
DROP TABLE IF EXISTS users;
```

### Migration Best Practices

- **Make migrations idempotent**: Use `IF EXISTS` and `IF NOT EXISTS` clauses
- **Use transactions**: Kat automatically wraps migrations in transactions
- **Implement both up and down**: Always provide the reverse operation
- **Reverse order in down migrations**: If your up migration creates A then B, your down migration should drop B then A
- **Keep migrations focused**: Each migration should have a single purpose
- **Test migrations**: Verify both up and down migrations work as expected

## Applying Migrations

To apply pending migrations, use the `up` command:

```bash
kat up
```

### How Up Migrations Work

When you run `kat up`, the following process occurs:

1. Kat scans your migrations directory for all migration folders
2. Kat computes migration definitions, converting files into executable SQL queries
3. Kat builds a directed acyclic graph (DAG) based on migration dependencies
4. Kat sorts migrations using topological ordering to respect dependencies
5. Kat connects to your database using your configuration
6. If needed, Kat creates a tracking table (specified by `tablename` in your config)
7. Kat reads the tracking table to determine which migrations have already been applied
8. For each pending migration:
   - Kat begins a transaction
   - Kat executes the SQL queries from the migration Definition
   - Kat records the migration in the tracking table
   - Kat commits the transaction
9. Kat provides a summary of the applied migrations

### Up Command Options

```bash
# Apply migrations with default config
kat up

# Apply migrations with a specific config file
kat up --config /path/to/config.yaml

# Validate migrations without applying them (dry run)
kat up --dry-run
```

### Example Output

```
Attempting to ping database
Successfully connected to database!
1679012345_create_users_table
1679023456_add_email_column
Successfully applied 2 migrations

Migration Summary
Successful migrations:
  ✓ 1679012345_create_users_table (up) - 15.621ms
  ✓ 1679023456_add_email_column (up) - 8.432ms

Total: 2 migration(s) applied
```

## Rolling Back Migrations

To roll back migrations, use the `down` command:

```bash
kat down
```

### How Down Migrations Work

When you run `kat down`, the following process occurs:

1. Kat scans your migrations directory and computes migration definitions
2. Kat connects to your database using your configuration
3. Kat reads the tracking table to identify applied migrations
4. By default, Kat selects the most recent migration for rollback
5. Kat begins a transaction
6. Kat executes the SQL queries from the migration Definition's DownQuery
7. Kat removes the migration record from the tracking table
8. Kat commits the transaction
9. Kat provides a summary of the rolled back migrations

### Down Command Options

```bash
# Roll back the most recent migration
kat down

# Roll back with a specific config file
kat down --config /path/to/config.yaml

# Roll back a specific number of migrations
kat down --count 3

# Validate rollback without applying it (dry run)
kat down --dry-run
```

### Example Output

```
Attempting to ping database
Successfully connected to database!
1679023456_add_email_column
Successfully rolled back 1 migrations

Migration Summary
Successful migrations:
  ✓ 1679023456_add_email_column (down) - 10.124ms

Total: 1 migration(s) rolled back
```

## Migration Tracking

Kat tracks migrations in a database table (default name: `migrations`). This table contains:

- **id**: Auto-incrementing ID
- **name**: Migration name (e.g., `1679012345_create_users_table`)
- **migration_time**: Timestamp when the migration was applied
- **duration**: How long the migration took to apply

You can customize the table name in your configuration:

```yaml
migration:
  tablename: migration_logs
  directory: migrations
```

## Dry Run Mode

Dry run mode allows you to validate migrations without applying them:

```bash
kat up --dry-run
kat down --dry-run
```

In dry run mode:
- SQL statements are not executed
- Database schema remains unchanged
- Migration tracking table is not updated
- Output indicates which migrations would be applied/rolled back

This is useful for:
- Validating migrations before deployment
- Testing migration scripts in CI/CD pipelines
- Reviewing changes before applying them to production

### Example Dry Run Output

```
DRY RUN: Migrations will not be applied
1679012345_create_users_table [DRY RUN] Would execute up migration for 1679012345_create_users_table
1679023456_add_email_column [DRY RUN] Would execute up migration for 1679023456_add_email_column
DRY RUN: Validated 2 migrations without applying them

Migration Summary
Validated migrations:
  ✓ 1679012345_create_users_table (up)
  ✓ 1679023456_add_email_column (up)

Total: 2 migration(s) validated
```

## Graph-Based Migration System

Kat implements a sophisticated graph-based migration system that allows for more complex dependency management between migrations.

### Understanding Migration Dependencies

In Kat, dependencies between migrations are managed using a directed acyclic graph (DAG). Each migration can define its parent migrations in the metadata.yaml file:

```yaml
name: add_user_profiles
timestamp: 1679123456
parents: [1679012345]  # References the timestamp of the parent migration
```

The graph-based system offers several advantages:

1. **Automatic Dependency Detection**: Dependencies are computed based on migration timestamps
2. **Parallel Migrations**: Independent migrations can be applied in parallel when possible
3. **Visualization**: Migration dependencies can be visualized as a graph
4. **Topological Ordering**: Migrations are applied in an order that ensures proper sequencing
5. **Graph-Based Execution**: The system identifies leaf nodes in the graph to determine execution order

### How Dependencies Work

Kat constructs a directed acyclic graph where:
1. Each migration is a vertex in the graph
2. Dependencies between migrations are defined through parent-child relationships
3. The system determines the optimal order of execution by performing topological sorting
4. Migrations are executed in a sequence that respects all dependencies

#### Parent-Child Relationships

In Kat, migrations relate to each other through explicit parent references in the `metadata.yaml` file:

```yaml
name: add_user_profiles
timestamp: 1679123456
parents: [1679012345]  # References the timestamp of the parent migration
```

The `parents` field contains timestamps of migrations that must be applied before the current migration. This creates a clear dependency chain:

1. **Execution Order Guarantee**: Kat ensures parent migrations are always applied before their children
2. **Flexibility**: A migration can have multiple parents, allowing for complex dependency trees
3. **DAG Construction**: These parent-child relationships form the edges in the directed acyclic graph
4. **Topological Sorting**: Kat uses the graph to compute the correct execution sequence

When running migrations, Kat performs a topological sort of the graph to determine the proper execution order. This ensures all dependencies are satisfied before a migration is applied.

### Visualizing the Migration Graph

To visualize your migration dependency graph, you can use the `graph` command:

```bash
kat graph --format dot > migrations.dot
```

This generates a DOT file that can be visualized with tools like Graphviz or online services like [GraphvizOnline](https://dreampuf.github.io/GraphvizOnline).

## Advanced Migration Patterns

### Schema Changes

```sql
-- up.sql
CREATE TABLE products (
  id SERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  price DECIMAL(10,2) NOT NULL
);

-- down.sql
DROP TABLE IF EXISTS products;
```

### Adding or Modifying Columns

```sql
-- up.sql
ALTER TABLE users
ADD COLUMN email VARCHAR(255) UNIQUE,
ADD COLUMN active BOOLEAN DEFAULT true;

-- down.sql
ALTER TABLE users
DROP COLUMN IF EXISTS active,
DROP COLUMN IF EXISTS email;
```

### Working with Constraints and Indexes

```sql
-- up.sql
-- Add constraints
ALTER TABLE orders
ADD CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(id);

-- Add indexes
CREATE INDEX idx_orders_user_id ON orders(user_id);

-- down.sql
-- Drop indexes first
DROP INDEX IF EXISTS idx_orders_user_id;

-- Then drop constraints
ALTER TABLE orders
DROP CONSTRAINT IF EXISTS fk_user_id;
```

### Seeding Data

```sql
-- up.sql
INSERT INTO roles (name) VALUES
('admin'),
('user'),
('guest');

-- down.sql
DELETE FROM roles WHERE name IN ('admin', 'user', 'guest');
```

## Troubleshooting Migrations

### Common Issues

1. **Migration fails to apply**
   - Check your database connection
   - Verify SQL syntax
   - Look for conflicts with existing schema

2. **Rollback fails**
   - Ensure down.sql correctly reverses up.sql
   - Check for dependencies that prevent rollback

3. **Dependency conflicts**
   - Ensure parent migrations exist in your migration directory
   - Check for circular dependencies, which will cause graph build errors

### Recovering from Failed Migrations

If a migration fails during the up or down operation:

1. Kat automatically rolls back the transaction
2. The database remains in its previous state
3. The migration is not recorded in the tracking table
4. Kat displays an error message with details

Example error:
```
Migration failed: 1679012345_create_users_table
Error details: ERROR: syntax error at or near "TABLEE" (SQLSTATE 42601)
Migration process stopped to preserve database integrity
```

To resolve the issue:
1. Fix the SQL in your migration file
2. Run the migration command again

## Environment-Specific Migrations

For environment-specific migrations, consider:

1. Using environment variables in your configuration
2. Creating environment-specific configuration files
3. Using conditional logic in your migrations based on the environment

Example with environment-specific configuration:
```bash
# Development
KAT_DB_NAME=myapp_dev kat up

# Production
KAT_DB_NAME=myapp_prod kat up --config prod-config.yaml
```

## Integration with CI/CD Pipelines

Kat works well in CI/CD pipelines for automated database migrations:

```bash
# Example CI/CD script
#!/bin/bash
set -e

# Test database connection
kat ping --retry-count 5 --retry-delay 1000

# Validate migrations (dry run)
kat up --dry-run

# Apply migrations
kat up
```

## Next Steps

After understanding how to work with migrations, you may want to:

1. Establish migration patterns for your specific database needs
2. Create a workflow for reviewing and testing migrations
3. Set up automated migration application in your deployment pipeline
