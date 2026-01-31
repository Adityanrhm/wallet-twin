-- Rollback: Drop goals and contributions tables

DROP TRIGGER IF EXISTS update_goals_updated_at ON goals;
DROP INDEX IF EXISTS idx_goal_contributions_goal_id;
DROP INDEX IF EXISTS idx_goals_active;
DROP INDEX IF EXISTS idx_goals_status;
DROP TABLE IF EXISTS goal_contributions CASCADE;
DROP TABLE IF EXISTS goals CASCADE;
DROP TYPE IF EXISTS goal_status;
