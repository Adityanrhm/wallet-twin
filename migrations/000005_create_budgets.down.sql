-- Rollback: Drop budgets table

DROP INDEX IF EXISTS idx_budgets_active_category;
DROP INDEX IF EXISTS idx_budgets_is_active;
DROP INDEX IF EXISTS idx_budgets_category_id;
DROP TABLE IF EXISTS budgets CASCADE;
DROP TYPE IF EXISTS budget_period;
