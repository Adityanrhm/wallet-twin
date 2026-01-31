-- Migration: Seed default categories
-- Version: 000008
-- Description: Insert default kategori income dan expense
--
-- Kategori default ini memberikan starting point untuk user baru.
-- User bisa edit, hapus, atau tambah kategori sesuai kebutuhan.

-- =========================================
-- INCOME CATEGORIES
-- =========================================

INSERT INTO categories (name, type, icon, color, sort_order) VALUES
    ('Salary', 'income', '💰', '#10B981', 1),
    ('Freelance', 'income', '💼', '#3B82F6', 2),
    ('Investment', 'income', '📈', '#8B5CF6', 3),
    ('Gift', 'income', '🎁', '#EC4899', 4),
    ('Refund', 'income', '↩️', '#6366F1', 5),
    ('Other Income', 'income', '💵', '#14B8A6', 99);

-- =========================================
-- EXPENSE CATEGORIES
-- =========================================

INSERT INTO categories (name, type, icon, color, sort_order) VALUES
    -- Kebutuhan rutin
    ('Food & Dining', 'expense', '🍔', '#EF4444', 1),
    ('Transportation', 'expense', '🚗', '#F59E0B', 2),
    ('Shopping', 'expense', '🛒', '#EC4899', 3),
    ('Bills & Utilities', 'expense', '🏠', '#8B5CF6', 4),
    
    -- Lifestyle
    ('Entertainment', 'expense', '🎮', '#3B82F6', 5),
    ('Health', 'expense', '💊', '#10B981', 6),
    ('Education', 'expense', '📚', '#6366F1', 7),
    ('Travel', 'expense', '✈️', '#14B8A6', 8),
    
    -- Lainnya
    ('Personal Care', 'expense', '💇', '#F472B6', 9),
    ('Gifts & Donations', 'expense', '🎁', '#A855F7', 10),
    ('Insurance', 'expense', '🛡️', '#64748B', 11),
    ('Other Expense', 'expense', '💳', '#94A3B8', 99);

-- =========================================
-- SUB-CATEGORIES (contoh hierarki)
-- =========================================

-- Sub-kategori untuk Food & Dining
INSERT INTO categories (name, type, icon, color, parent_id, sort_order)
SELECT 'Groceries', 'expense', '🥬', '#22C55E', id, 1
FROM categories WHERE name = 'Food & Dining';

INSERT INTO categories (name, type, icon, color, parent_id, sort_order)
SELECT 'Restaurant', 'expense', '🍽️', '#F97316', id, 2
FROM categories WHERE name = 'Food & Dining';

INSERT INTO categories (name, type, icon, color, parent_id, sort_order)
SELECT 'Coffee', 'expense', '☕', '#92400E', id, 3
FROM categories WHERE name = 'Food & Dining';

-- Sub-kategori untuk Transportation
INSERT INTO categories (name, type, icon, color, parent_id, sort_order)
SELECT 'Fuel', 'expense', '⛽', '#EAB308', id, 1
FROM categories WHERE name = 'Transportation';

INSERT INTO categories (name, type, icon, color, parent_id, sort_order)
SELECT 'Public Transport', 'expense', '🚌', '#0EA5E9', id, 2
FROM categories WHERE name = 'Transportation';

INSERT INTO categories (name, type, icon, color, parent_id, sort_order)
SELECT 'Ride Sharing', 'expense', '🚕', '#22D3EE', id, 3
FROM categories WHERE name = 'Transportation';

INSERT INTO categories (name, type, icon, color, parent_id, sort_order)
SELECT 'Parking', 'expense', '🅿️', '#64748B', id, 4
FROM categories WHERE name = 'Transportation';
