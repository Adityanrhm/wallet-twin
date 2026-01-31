-- Rollback: Drop transactions table

DROP TRIGGER IF EXISTS update_transactions_updated_at ON transactions;
DROP INDEX IF EXISTS idx_transactions_wallet_date;
DROP INDEX IF EXISTS idx_transactions_type;
DROP INDEX IF EXISTS idx_transactions_date;
DROP INDEX IF EXISTS idx_transactions_category_id;
DROP INDEX IF EXISTS idx_transactions_wallet_id;
DROP TABLE IF EXISTS transactions CASCADE;
DROP TYPE IF EXISTS transaction_type;
