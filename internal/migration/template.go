package migration

const metadataFileTemplate = `name: %s
timestamp: %d
description: |
  <add comment about your migration in here>
parents: %s
`

const downMigrationFileTemplate = `-- Undo the changes made in the up migration
--
-- Note: All migrations in kat are automatically wrapped in a transaction.
-- You don't need to add BEGIN/COMMIT statements manually.
`

const upMigrationFileTemplate = `-- Perform migration here.
--
--  It's helpful to make migrations idempotent, that way migrations can be executed multiple times
-- and the database structure will be the same.
--
-- Note: All migrations in kat are automatically wrapped in a transaction.
-- You don't need to add BEGIN/COMMIT statements manually.
`
