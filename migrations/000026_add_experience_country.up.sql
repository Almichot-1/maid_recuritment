-- Add country_of_experience field to candidates table
ALTER TABLE candidates
ADD COLUMN country_of_experience TEXT;

COMMENT ON COLUMN candidates.country_of_experience IS 'Country where the candidate gained work experience (only relevant if experience_years > 0)';
