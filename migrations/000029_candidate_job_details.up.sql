-- Add country_applied and salary_offered columns to candidates table
ALTER TABLE candidates 
ADD COLUMN IF NOT EXISTS country_applied VARCHAR(100),
ADD COLUMN IF NOT EXISTS salary_offered VARCHAR(100);

-- Add company_name to users table for agency branding
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS company_name VARCHAR(255);

-- Add index for filtering by country_applied
CREATE INDEX IF NOT EXISTS idx_candidates_country_applied ON candidates(country_applied);

-- Add comments for clarity
COMMENT ON COLUMN candidates.country_applied IS 'The country where the candidate is being recruited to work';
COMMENT ON COLUMN candidates.salary_offered IS 'The salary being offered to the candidate (free text format, e.g., "1000 SR")';
COMMENT ON COLUMN users.company_name IS 'Agency company name for branding in CVs and documents';
