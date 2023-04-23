package runner

const createMigrationTableStmt = `CREATE TABLE IF NOT EXISTS %s (
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL,
	timestamp BIGINT NOT NULL,
	error TEXT,
	success BOOLEAN NOT NULL,
	started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
)

CREATE INDEX idx_migration_logs_timestamp_name ON %s (timestamp, name)`

const dropMigrationTableStmt = `DROP TABLE IF EXISTS %s`
