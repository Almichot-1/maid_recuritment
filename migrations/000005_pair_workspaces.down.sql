DROP TRIGGER IF EXISTS trg_candidate_pair_shares_updated_at ON candidate_pair_shares;
DROP TRIGGER IF EXISTS trg_agency_pairings_updated_at ON agency_pairings;

ALTER TABLE selections DROP CONSTRAINT IF EXISTS fk_selections_pairing_id;
ALTER TABLE selections DROP COLUMN IF EXISTS pairing_id;

DROP TABLE IF EXISTS candidate_pair_shares;
DROP TABLE IF EXISTS agency_pairings;

DROP TYPE IF EXISTS agency_pairing_status;
