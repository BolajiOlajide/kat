package migration

const metadataFileTemplate = `name: %s
timestamp: %d
`
const downMigrationFileTemplate = `-- Undo the changes made in the up migration
`

const upMigrationFileTemplate = `-- Perform migration here.
--
--  * Make migrations idempotent (use IF EXISTS)
--  * Make migrations backwards-compatible (old readers/writers must continue to work)
--  * If you are using CREATE INDEX CONCURRENTLY, then make sure that only one statement
--    is defined per file, and that each such statement is NOT wrapped in a transaction.
--    Each such migration must also declare "createIndexConcurrently: true" in their
--    associated metadata.yaml file.
--  * If you are modifying Postgres extensions, you must also declare "privileged: true"
--    in the associated metadata.yaml file.
`
