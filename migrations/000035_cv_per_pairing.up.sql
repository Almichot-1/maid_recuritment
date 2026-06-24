ALTER TABLE candidate_pair_shares
  ADD COLUMN IF NOT EXISTS cv_pdf_url TEXT;

COMMENT ON COLUMN candidate_pair_shares.cv_pdf_url
  IS 'Per-pairing CV PDF URL. Each share gets its own CV with pairing-specific overrides (country, salary). NULL until first generation.';
