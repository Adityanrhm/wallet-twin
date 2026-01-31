-- Migration: Create budgets table
-- Version: 000005
-- Description: Tabel untuk menyimpan budget/anggaran per kategori
--
-- Budget membantu user mengontrol pengeluaran per kategori.
-- User set budget bulanan, dan aplikasi track progress.
--
-- Contoh:
-- - Budget Food & Dining: Rp 2.000.000 per bulan
-- - Budget Transportation: Rp 500.000 per bulan
-- - Budget Entertainment: Rp 300.000 per bulan
--
-- Aplikasi akan alert jika pengeluaran mendekati atau melebihi budget.

-- Periode budget
CREATE TYPE budget_period AS ENUM ('weekly', 'monthly', 'yearly');

CREATE TABLE IF NOT EXISTS budgets (
    -- Primary key UUID
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Kategori yang di-budget
    -- NOT NULL karena budget harus untuk kategori tertentu
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    
    -- Jumlah budget
    amount NUMERIC(15, 2) NOT NULL CHECK (amount > 0),
    
    -- Periode budget
    -- monthly = paling umum
    period budget_period NOT NULL DEFAULT 'monthly',
    
    -- Tanggal mulai budget
    -- Untuk monthly: biasanya tanggal 1 bulan ini
    start_date DATE NOT NULL,
    
    -- Tanggal akhir budget (opsional)
    -- NULL = budget berlaku selamanya (recurring)
    end_date DATE,
    
    -- Status aktif
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Constraint: end_date harus setelah start_date
    CONSTRAINT valid_date_range CHECK (end_date IS NULL OR end_date > start_date)
);

-- Index untuk query
CREATE INDEX idx_budgets_category_id ON budgets(category_id);
CREATE INDEX idx_budgets_is_active ON budgets(is_active);

-- Partial index: hanya budget aktif (paling sering diquery)
CREATE INDEX idx_budgets_active_category ON budgets(category_id) WHERE is_active = TRUE;

-- Komentar dokumentasi
COMMENT ON TABLE budgets IS 'Budget/anggaran per kategori per periode';
COMMENT ON COLUMN budgets.end_date IS 'NULL = budget berlaku selamanya';
