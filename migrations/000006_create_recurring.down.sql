-- Rollback: Drop recurring_transactions table

DROP INDEX IF EXISTS idx_recurring_due_active;
DROP INDEX IF EXISTS idx_recurring_is_active;
DROP INDEX IF EXISTS idx_recurring_next_due;
DROP INDEX IF EXISTS idx_recurring_wallet_id;
DROP TABLE IF EXISTS recurring_transactions CASCADE;
DROP TYPE IF EXISTS recurring_frequency;
