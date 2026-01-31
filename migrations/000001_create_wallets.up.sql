-- Migration: Create wallets table
-- Version: 000001
-- Description: Tabel untuk menyimpan data wallet/akun keuangan
--
-- Wallet adalah entitas utama yang merepresentasikan tempat penyimpanan uang.
-- Contoh: Cash, BCA, GoPay, OVO, Mandiri, dll.
--
-- Setiap wallet memiliki:
-- - Saldo (balance) yang di-track secara real-time
-- - Tipe (cash, bank, e-wallet) untuk kategorisasi
-- - Status aktif/non-aktif untuk soft delete

-- Enable UUID extension jika belum ada
-- UUID digunakan sebagai primary key karena:
-- 1. Tidak sequential (lebih aman dari enumeration attack)
-- 2. Bisa di-generate di aplikasi tanpa query ke DB
-- 3. Unique secara global
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Tipe wallet: cash (uang tunai), bank (rekening), ewallet (dompet digital)
-- Menggunakan ENUM untuk validasi di level database
CREATE TYPE wallet_type AS ENUM ('cash', 'bank', 'ewallet');

-- Tabel wallets
CREATE TABLE IF NOT EXISTS wallets (
    -- Primary key menggunakan UUID v4
    -- uuid_generate_v4() generate random UUID
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Nama wallet yang ditampilkan ke user
    -- NOT NULL karena wajib diisi
    -- Contoh: "Cash", "BCA Tabungan", "GoPay"
    name VARCHAR(100) NOT NULL,
    
    -- Tipe wallet untuk kategorisasi
    -- DEFAULT 'cash' untuk kemudahan penggunaan
    type wallet_type NOT NULL DEFAULT 'cash',
    
    -- Saldo wallet
    -- NUMERIC(15, 2) = maksimal 15 digit, 2 desimal
    -- Contoh: 1234567890123.45 (sampai triliun)
    -- DEFAULT 0 untuk wallet baru
    balance NUMERIC(15, 2) NOT NULL DEFAULT 0,
    
    -- Kode mata uang ISO 4217
    -- 3 karakter: IDR, USD, EUR, dll
    -- DEFAULT 'IDR' untuk Indonesia
    currency CHAR(3) NOT NULL DEFAULT 'IDR',
    
    -- Warna untuk UI (hex color)
    -- Contoh: "#7C3AED" (purple)
    color VARCHAR(7),
    
    -- Icon emoji atau nama icon
    -- Contoh: "💰", "🏦", "wallet"
    icon VARCHAR(50),
    
    -- Status aktif untuk soft delete
    -- TRUE = wallet aktif dan ditampilkan
    -- FALSE = wallet non-aktif (tersembunyi tapi data tetap ada)
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- Timestamps untuk audit trail
    -- created_at: kapan record dibuat
    -- updated_at: kapan terakhir diupdate
    -- Menggunakan TIMESTAMPTZ untuk timezone-aware
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index untuk query yang sering digunakan
-- is_active: Filter wallet aktif (paling sering)
CREATE INDEX idx_wallets_is_active ON wallets(is_active);

-- Trigger untuk auto-update updated_at
-- Setiap kali row di-UPDATE, updated_at otomatis diupdate
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_wallets_updated_at
    BEFORE UPDATE ON wallets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Komentar untuk dokumentasi
COMMENT ON TABLE wallets IS 'Menyimpan data wallet/akun keuangan pengguna';
COMMENT ON COLUMN wallets.balance IS 'Saldo wallet dalam mata uang yang ditentukan';
COMMENT ON COLUMN wallets.is_active IS 'FALSE untuk soft delete';
