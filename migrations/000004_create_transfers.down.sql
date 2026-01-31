-- Rollback: Drop transfers table

DROP INDEX IF EXISTS idx_transfers_created_at;
DROP INDEX IF EXISTS idx_transfers_to_wallet;
DROP INDEX IF EXISTS idx_transfers_from_wallet;
DROP TABLE IF EXISTS transfers CASCADE;
