ALTER TABLE agency_pairings
  ADD COLUMN IF NOT EXISTS default_salary TEXT;

COMMENT ON COLUMN agency_pairings.default_salary
  IS 'Default salary amount (e.g. "2000"). Combined with default_currency => "2000 KWD" on CVs. Set per-partner by the Ethiopian agent.';
