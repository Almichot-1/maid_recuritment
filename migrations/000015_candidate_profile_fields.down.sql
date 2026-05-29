ALTER TABLE candidates
	DROP COLUMN IF EXISTS education_level,
	DROP COLUMN IF EXISTS children_count,
	DROP COLUMN IF EXISTS marital_status,
	DROP COLUMN IF EXISTS religion,
	DROP COLUMN IF EXISTS place_of_birth,
	DROP COLUMN IF EXISTS date_of_birth,
	DROP COLUMN IF EXISTS nationality;
