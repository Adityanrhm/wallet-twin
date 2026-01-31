-- Migration: Create goals table
-- Version: 000007
-- Description: Tabel untuk target tabungan (savings goals)
--
-- Goals membantu user menabung untuk tujuan tertentu.
-- User set target dan deadline, aplikasi track progress.
--
-- Contoh:
-- - Emergency Fund: Rp 10.000.000 (deadline: 6 bulan)
-- - Holiday Trip: Rp 5.000.000 (deadline: Desember)
-- - New Laptop: Rp 15.000.000 (deadline: 1 tahun)
--
-- User bisa menambah kontribusi ke goal dari wallet manapun.
-- Kontribusi dicatat terpisah untuk tracking history.

-- Status goal
CREATE TYPE goal_status AS ENUM ('active', 'completed', 'cancelled');

-- Tabel goals
CREATE TABLE IF NOT EXISTS goals (
    -- Primary key UUID
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Nama goal
    -- Contoh: "Emergency Fund", "Holiday Trip"
    name VARCHAR(100) NOT NULL,
    
    -- Deskripsi goal (opsional)
    description TEXT,
    
    -- Target jumlah yang ingin dicapai
    target_amount NUMERIC(15, 2) NOT NULL CHECK (target_amount > 0),
    
    -- Jumlah yang sudah terkumpul
    -- Di-update setiap ada kontribusi
    current_amount NUMERIC(15, 2) NOT NULL DEFAULT 0,
    
    -- Deadline (opsional)
    -- NULL = tidak ada deadline
    deadline DATE,
    
    -- Status goal
    status goal_status NOT NULL DEFAULT 'active',
    
    -- Warna untuk UI
    color VARCHAR(7),
    
    -- Icon
    icon VARCHAR(50),
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tabel goal_contributions
-- Mencatat setiap kontribusi ke goal
CREATE TABLE IF NOT EXISTS goal_contributions (
    -- Primary key UUID
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Goal yang dikontribusi
    goal_id UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    
    -- Jumlah kontribusi
    amount NUMERIC(15, 2) NOT NULL CHECK (amount > 0),
    
    -- Catatan (opsional)
    note TEXT,
    
    -- Timestamp kontribusi
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index untuk goals
CREATE INDEX idx_goals_status ON goals(status);

-- Partial index: hanya goal aktif
CREATE INDEX idx_goals_active ON goals(deadline) WHERE status = 'active';

-- Index untuk contributions
CREATE INDEX idx_goal_contributions_goal_id ON goal_contributions(goal_id);

-- Trigger untuk update goals.updated_at
CREATE TRIGGER update_goals_updated_at
    BEFORE UPDATE ON goals
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Komentar dokumentasi
COMMENT ON TABLE goals IS 'Target tabungan/savings goals';
COMMENT ON TABLE goal_contributions IS 'History kontribusi ke goal';
COMMENT ON COLUMN goals.current_amount IS 'Jumlah terkumpul (sum of contributions)';
