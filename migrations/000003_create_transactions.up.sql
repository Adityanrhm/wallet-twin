-- Migration: Create transactions table
-- Version: 000003
-- Description: Tabel untuk menyimpan transaksi keuangan
--
-- Transaction adalah inti dari aplikasi - setiap pemasukan dan pengeluaran
-- dicatat sebagai transaction.
--
-- Transaksi selalu terhubung ke:
-- - Wallet: Dari mana/kemana uang mengalir
-- - Category: Untuk apa transaksi ini
--
-- Contoh transaksi:
-- - Income: Gaji bulan Januari +Rp 5.000.000
-- - Expense: Makan siang -Rp 50.000

-- Tipe transaksi
CREATE TYPE transaction_type AS ENUM ('income', 'expense');

-- Tabel transactions
CREATE TABLE IF NOT EXISTS transactions (
    -- Primary key UUID
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Foreign key ke wallet
    -- ON DELETE CASCADE: Jika wallet dihapus, transaksi juga dihapus
    -- PENTING: Ini destructive! Pastikan user konfirmasi sebelum hapus wallet
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    
    -- Foreign key ke category
    -- ON DELETE SET NULL: Jika category dihapus, transaksi tetap ada (uncategorized)
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    
    -- Tipe transaksi: income atau expense
    type transaction_type NOT NULL,
    
    -- Jumlah transaksi (selalu positif)
    -- Tipe menentukan apakah ini menambah atau mengurangi saldo
    -- NUMERIC(15, 2) = sampai triliun dengan 2 desimal
    amount NUMERIC(15, 2) NOT NULL CHECK (amount > 0),
    
    -- Deskripsi/catatan transaksi
    -- Contoh: "Makan siang di warteg", "Gaji bulan Januari"
    description TEXT,
    
    -- Tags untuk filtering tambahan (array of strings)
    -- Contoh: ["work", "lunch"], ["monthly", "salary"]
    tags TEXT[],
    
    -- Tanggal transaksi (bisa berbeda dengan created_at)
    -- User bisa input transaksi untuk tanggal kemarin atau masa lalu
    transaction_date DATE NOT NULL DEFAULT CURRENT_DATE,
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index untuk query yang sering digunakan

-- wallet_id: List transaksi per wallet
CREATE INDEX idx_transactions_wallet_id ON transactions(wallet_id);

-- category_id: Filter/group by category
CREATE INDEX idx_transactions_category_id ON transactions(category_id);

-- transaction_date: Filter by tanggal, sorting, reports
CREATE INDEX idx_transactions_date ON transactions(transaction_date);

-- type: Filter income/expense
CREATE INDEX idx_transactions_type ON transactions(type);

-- Composite index untuk report bulanan (paling sering)
-- Contoh query: SELECT * FROM transactions 
--               WHERE wallet_id = ? AND transaction_date BETWEEN ? AND ?
CREATE INDEX idx_transactions_wallet_date ON transactions(wallet_id, transaction_date);

-- Trigger untuk update updated_at
CREATE TRIGGER update_transactions_updated_at
    BEFORE UPDATE ON transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Komentar dokumentasi
COMMENT ON TABLE transactions IS 'Menyimpan semua transaksi income dan expense';
COMMENT ON COLUMN transactions.amount IS 'Jumlah transaksi (selalu positif)';
COMMENT ON COLUMN transactions.transaction_date IS 'Tanggal transaksi (bisa berbeda dari created_at)';
