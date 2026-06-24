ALTER TABLE agency_pairings
  ADD COLUMN IF NOT EXISTS default_country   TEXT,
  ADD COLUMN IF NOT EXISTS default_currency  TEXT,
  ADD COLUMN IF NOT EXISTS partner_logo_url  TEXT;

COMMENT ON COLUMN agency_pairings.default_country  IS 'Default country applied on CVs for this pairing (set by Ethiopian agent)';
COMMENT ON COLUMN agency_pairings.default_currency IS 'Default currency appended to salary on CVs for this pairing (e.g. JOD)';
COMMENT ON COLUMN agency_pairings.partner_logo_url IS 'Logo URL uploaded by Ethiopian agent for the foreign partner brand on CVs';
