-- Migration: Create categories table
-- Version: 000002
-- Description: Tabel untuk menyimpan kategori transaksi
--
-- Category digunakan untuk mengkategorikan transaksi.
-- Ada 2 tipe: income (pemasukan) dan expense (pengeluaran)
--
-- Contoh kategori income:
-- - Salary (Gaji)
-- - Freelance
-- - Investment Returns
--
-- Contoh kategori expense:
-- - Food & Dining
-- - Transportation
-- - Shopping
--
-- Category mendukung hierarki (parent-child) untuk sub-kategori:
-- - Food & Dining (parent)
--   - Lunch
--   - Dinner
--   - Coffee

-- Tipe kategori: income atau expense
CREATE TYPE category_type AS ENUM ('income', 'expense');

-- Tabel categories
CREATE TABLE IF NOT EXISTS categories (
    -- Primary key UUID
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Nama kategori
    -- Contoh: "Food & Dining", "Salary", "Transportation"
    name VARCHAR(100) NOT NULL,
    
    -- Tipe kategori: income atau expense
    -- Ini menentukan kategori muncul di form income atau expense
    type category_type NOT NULL,
    
    -- Warna untuk UI visualization
    -- Contoh: "#EF4444" untuk expense (red), "#10B981" untuk income (green)
    color VARCHAR(7),
    
    -- Icon emoji atau nama icon
    -- Contoh: "🍔", "💰", "🚗"
    icon VARCHAR(50),
    
    -- Parent category untuk hierarki (self-referencing)
    -- NULL = ini adalah top-level category
    -- NOT NULL = ini adalah sub-category
    --
    -- Contoh:
    -- Food & Dining (parent_id = NULL)
    --   └── Lunch (parent_id = Food & Dining ID)
    --   └── Dinner (parent_id = Food & Dining ID)
    parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    
    -- Urutan tampilan (untuk custom ordering)
    -- Lower number = tampil lebih dulu
    sort_order INTEGER DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index untuk query yang sering
-- type: Filter kategori berdasarkan income/expense
CREATE INDEX idx_categories_type ON categories(type);

-- parent_id: Query sub-categories
CREATE INDEX idx_categories_parent_id ON categories(parent_id);

-- Komentar dokumentasi
COMMENT ON TABLE categories IS 'Kategori untuk transaksi income dan expense';
COMMENT ON COLUMN categories.parent_id IS 'Self-reference untuk hierarki sub-kategori';
