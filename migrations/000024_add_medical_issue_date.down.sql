-- Revert expiry_date to NOT NULL and TIMESTAMPTZ (fill any null values to prevent constraint error)
UPDATE medical_data SET expiry_date = NOW() WHERE expiry_date IS NULL;
ALTER TABLE medical_data ALTER COLUMN expiry_date TYPE TIMESTAMPTZ USING expiry_date::TIMESTAMPTZ;
ALTER TABLE medical_data ALTER COLUMN expiry_date SET NOT NULL;

-- Remove issue_date column if it exists
ALTER TABLE medical_data DROP COLUMN IF EXISTS issue_date;
