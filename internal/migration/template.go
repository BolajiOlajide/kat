package migration

const metadataFileTemplate = `name: %s
timestamp: %d
`
const downMigrationFileTemplate = `-- Undo the changes made in the up migration
`

const upMigrationFileTemplate = `-- Perform migration here.
--
--  It's helpful to make migrations idempotent, that way migrations can be executed multiple times
-- and the database structure will be the same.
`

const migrationLogsStmt = `CREATE TABLE IF NOT EXISTS %s (
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL,
	timestamp INTEGER NOT NULL,
	error TEXT,
	success BOOLEAN NOT NULL,
	started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
)`
