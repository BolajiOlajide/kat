-- Undo the changes made in the up migration

-- Drop the trigger first
DROP TRIGGER IF EXISTS update_transactions_updated_at ON transactions;

-- Drop the indexes
DROP INDEX IF EXISTS idx_transactions_user_id;
DROP INDEX IF EXISTS idx_transactions_status;
DROP INDEX IF EXISTS idx_transactions_created_at;
DROP INDEX IF EXISTS idx_transactions_reference;

-- Drop the table
DROP TABLE IF EXISTS transactions;

-- Drop the enums
DROP TYPE IF EXISTS transaction_status;
DROP TYPE IF EXISTS transaction_type;
