ALTER TABLE selections
    DROP COLUMN IF EXISTS warning_sent_flags;

DROP INDEX IF EXISTS idx_selections_expires_at_pending;
