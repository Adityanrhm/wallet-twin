-- Migration: Create recurring_transactions table
-- Version: 000006
-- Description: Tabel untuk transaksi berulang/terjadwal
--
-- Recurring transaction adalah transaksi yang terjadi secara berkala.
-- Sistem akan otomatis generate transaksi saat jatuh tempo.
--
-- Contoh:
-- - Gaji bulanan (income, monthly, tanggal 25)
-- - Langganan Netflix (expense, monthly)
-- - Bayar listrik (expense, monthly)
-- - Uang saku mingguan (expense, weekly)
--
-- Workflow:
-- 1. User setup recurring transaction
-- 2. Sistem check setiap hari untuk next_due
-- 3. Jika next_due <= today, generate transaksi
-- 4. Update next_due ke periode berikutnya

-- Frekuensi recurring
CREATE TYPE recurring_frequency AS ENUM ('daily', 'weekly', 'monthly', 'yearly');

CREATE TABLE IF NOT EXISTS recurring_transactions (
    -- Primary key UUID
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Wallet untuk transaksi
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    
    -- Kategori transaksi
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    
    -- Tipe: income atau expense
    type transaction_type NOT NULL,
    
    -- Jumlah transaksi
    amount NUMERIC(15, 2) NOT NULL CHECK (amount > 0),
    
    -- Deskripsi transaksi
    description TEXT,
    
    -- Frekuensi recurring
    frequency recurring_frequency NOT NULL,
    
    -- Tanggal jatuh tempo berikutnya
    -- Ini yang di-check oleh scheduler
    next_due DATE NOT NULL,
    
    -- Tanggal akhir recurring (opsional)
    -- NULL = recurring selamanya
    -- NOT NULL = berhenti setelah tanggal ini
    end_date DATE,
    
    -- Status aktif
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Constraint: end_date harus di masa depan dari next_due
    CONSTRAINT valid_end_date CHECK (end_date IS NULL OR end_date >= next_due)
);

-- Index untuk query
CREATE INDEX idx_recurring_wallet_id ON recurring_transactions(wallet_id);
CREATE INDEX idx_recurring_next_due ON recurring_transactions(next_due);
CREATE INDEX idx_recurring_is_active ON recurring_transactions(is_active);

-- Partial index: recurring aktif yang jatuh tempo (untuk scheduler)
-- Query: SELECT * FROM recurring_transactions 
--        WHERE is_active = TRUE AND next_due <= CURRENT_DATE
CREATE INDEX idx_recurring_due_active ON recurring_transactions(next_due) 
    WHERE is_active = TRUE;

-- Komentar dokumentasi
COMMENT ON TABLE recurring_transactions IS 'Transaksi berulang/terjadwal';
COMMENT ON COLUMN recurring_transactions.next_due IS 'Tanggal jatuh tempo berikutnya';
