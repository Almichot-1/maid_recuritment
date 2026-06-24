-- candidate_pair_overrides: stores per-pairing country_applied and salary_offered
-- overrides for a candidate. When generating a CV for a specific foreign agency
-- pairing, these values take precedence over the global candidate defaults.
CREATE TABLE IF NOT EXISTS candidate_pair_overrides (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pairing_id   UUID NOT NULL REFERENCES agency_pairings(id)  ON DELETE CASCADE,
    candidate_id UUID NOT NULL REFERENCES candidates(id)        ON DELETE CASCADE,
    country_applied VARCHAR(100),
    salary_offered  VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_pair_override UNIQUE (pairing_id, candidate_id)
);

CREATE INDEX IF NOT EXISTS idx_pair_overrides_pairing_candidate
    ON candidate_pair_overrides(pairing_id, candidate_id);

CREATE INDEX IF NOT EXISTS idx_pair_overrides_candidate
    ON candidate_pair_overrides(candidate_id);

COMMENT ON TABLE  candidate_pair_overrides IS 'Per-pairing overrides for country_applied and salary_offered on a candidate CV.';
COMMENT ON COLUMN candidate_pair_overrides.pairing_id     IS 'FK to agency_pairings – the specific Ethiopian↔Foreign agency relationship.';
COMMENT ON COLUMN candidate_pair_overrides.candidate_id   IS 'FK to candidates – the candidate whose CV is being customised.';
COMMENT ON COLUMN candidate_pair_overrides.country_applied IS 'Override country for this pairing (e.g. Kuwait). Falls back to candidates.country_applied if NULL.';
COMMENT ON COLUMN candidate_pair_overrides.salary_offered  IS 'Override salary for this pairing (e.g. 900 KWD). Falls back to candidates.salary_offered if NULL.';
