-- Migration: Create transfers table
-- Version: 000004
-- Description: Tabel untuk menyimpan transfer antar wallet
--
-- Transfer adalah operasi khusus yang memindahkan uang dari satu wallet
-- ke wallet lain. Berbeda dengan transaksi biasa, transfer melibatkan
-- 2 wallet sekaligus.
--
-- Contoh:
-- - Transfer dari BCA ke GoPay Rp 500.000
-- - Tarik tunai dari ATM (Bank → Cash)
-- - Top up e-wallet (Bank → E-wallet)
--
-- Transfer bisa memiliki fee (biaya transfer)

CREATE TABLE IF NOT EXISTS transfers (
    -- Primary key UUID
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Wallet sumber (uang keluar dari sini)
    from_wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    
    -- Wallet tujuan (uang masuk ke sini)
    to_wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    
    -- Jumlah yang ditransfer
    amount NUMERIC(15, 2) NOT NULL CHECK (amount > 0),
    
    -- Biaya transfer (opsional)
    -- Contoh: biaya transfer antar bank Rp 6.500
    fee NUMERIC(15, 2) DEFAULT 0 CHECK (fee >= 0),
    
    -- Catatan transfer
    note TEXT,
    
    -- Timestamp transfer
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Constraint: tidak boleh transfer ke wallet yang sama
    CONSTRAINT different_wallets CHECK (from_wallet_id != to_wallet_id)
);

-- Index untuk query
CREATE INDEX idx_transfers_from_wallet ON transfers(from_wallet_id);
CREATE INDEX idx_transfers_to_wallet ON transfers(to_wallet_id);
CREATE INDEX idx_transfers_created_at ON transfers(created_at);

-- Komentar dokumentasi
COMMENT ON TABLE transfers IS 'Menyimpan transfer antar wallet';
COMMENT ON COLUMN transfers.fee IS 'Biaya transfer (dibebankan ke wallet sumber)';
