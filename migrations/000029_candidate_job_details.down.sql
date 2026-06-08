-- Remove indexes
DROP INDEX IF EXISTS idx_candidates_country_applied;

-- Remove columns
ALTER TABLE candidates 
DROP COLUMN IF EXISTS country_applied,
DROP COLUMN IF EXISTS salary_offered;

ALTER TABLE users 
DROP COLUMN IF EXISTS company_name;
