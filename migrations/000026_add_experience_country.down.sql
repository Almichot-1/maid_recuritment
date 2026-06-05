-- Remove country_of_experience field from candidates table
ALTER TABLE candidates
DROP COLUMN IF EXISTS country_of_experience;
