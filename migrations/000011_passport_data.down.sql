DROP TRIGGER IF EXISTS trg_passport_data_updated_at ON passport_data;
DROP INDEX IF EXISTS idx_passport_data_expiry_date;
DROP INDEX IF EXISTS idx_passport_data_candidate_id;
DROP TABLE IF EXISTS passport_data;
