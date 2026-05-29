-- Add issue_date column to medical_data table if not exists
ALTER TABLE medical_data ADD COLUMN IF NOT EXISTS issue_date DATE;

-- Alter expiry_date column to be nullable and type DATE
ALTER TABLE medical_data ALTER COLUMN expiry_date TYPE DATE USING expiry_date::DATE;
ALTER TABLE medical_data ALTER COLUMN expiry_date DROP NOT NULL;
