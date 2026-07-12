ALTER TABLE candidates
	DROP COLUMN IF EXISTS experience_abroad,
	DROP COLUMN IF EXISTS issuing_authority,
	DROP COLUMN IF EXISTS gender,
	DROP COLUMN IF EXISTS expiry_date,
	DROP COLUMN IF EXISTS issue_date,
	DROP COLUMN IF EXISTS passport_number;
