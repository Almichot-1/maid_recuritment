-- Migration rollback: Recreate status_steps table (if needed for rollback)
-- Note: This is a basic recreation - data will not be restored

CREATE TABLE IF NOT EXISTS status_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    selection_id UUID NOT NULL REFERENCES selections(id) ON DELETE CASCADE,
    step_name TEXT NOT NULL,
    status TEXT NOT NULL,
    completed_at TIMESTAMPTZ,
    notes TEXT,
    coc_status TEXT,
    arrival_city TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_status_steps_selection_id ON status_steps(selection_id);
