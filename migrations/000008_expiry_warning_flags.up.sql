ALTER TABLE selections
    ADD COLUMN IF NOT EXISTS warning_sent_flags INTEGER NOT NULL DEFAULT 0;

COMMENT ON COLUMN selections.warning_sent_flags IS
    'Bitmask tracking which expiry warnings have been sent. Bit 0 = 24h, Bit 1 = 6h, Bit 2 = 1h';

CREATE INDEX IF NOT EXISTS idx_selections_expires_at_pending
    ON selections(expires_at)
    WHERE status = 'pending';
