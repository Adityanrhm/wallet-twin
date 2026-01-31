-- Rollback: Drop categories table

DROP INDEX IF EXISTS idx_categories_parent_id;
DROP INDEX IF EXISTS idx_categories_type;
DROP TABLE IF EXISTS categories CASCADE;
DROP TYPE IF EXISTS category_type;
