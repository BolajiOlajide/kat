package migration

const metadataFileTemplate = `name: %s
timestamp: %d
`
const downMigrationFileTemplate = `-- Undo the changes made in the up migration
`

const upMigrationFileTemplate = `-- Perform migration here.
--
--  * Make migrations idempotent (use IF EXISTS)
`
