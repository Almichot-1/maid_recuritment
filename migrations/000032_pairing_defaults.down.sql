ALTER TABLE agency_pairings
  DROP COLUMN IF EXISTS default_country,
  DROP COLUMN IF EXISTS default_currency,
  DROP COLUMN IF EXISTS partner_logo_url;
