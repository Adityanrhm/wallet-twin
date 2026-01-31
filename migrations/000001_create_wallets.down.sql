-- Rollback: Drop wallets table
-- Menghapus semua yang dibuat di 000001_create_wallets.up.sql
-- 
-- URUTAN PENTING: Hapus dari yang paling dependent dulu
-- 1. Trigger (depends on function dan table)
-- 2. Index (depends on table)
-- 3. Table
-- 4. Type (baru bisa dihapus setelah tidak ada yang pakai)

-- Hapus trigger dulu
DROP TRIGGER IF EXISTS update_wallets_updated_at ON wallets;

-- Hapus function (masih dipakai tabel lain, jadi pakai CASCADE dengan hati-hati)
-- Kita tidak hapus function karena mungkin dipakai oleh tabel lain
-- DROP FUNCTION IF EXISTS update_updated_at_column();

-- Hapus table (CASCADE akan hapus semua dependencies)
DROP TABLE IF EXISTS wallets CASCADE;

-- Hapus enum type
DROP TYPE IF EXISTS wallet_type;

-- CATATAN: Kita tidak hapus uuid-ossp extension karena mungkin dipakai sistem lain
-- DROP EXTENSION IF EXISTS "uuid-ossp";
