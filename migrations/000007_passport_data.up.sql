CREATE TABLE IF NOT EXISTS passport_data (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	candidate_id UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
	holder_name TEXT NOT NULL DEFAULT '',
	passport_number TEXT NOT NULL DEFAULT '',
	nationality TEXT NOT NULL DEFAULT '',
	date_of_birth DATE,
	place_of_birth TEXT NOT NULL DEFAULT '',
	gender TEXT NOT NULL DEFAULT '',
	issue_date DATE,
	expiry_date DATE,
	place_of_issue TEXT NOT NULL DEFAULT '',
	issuing_authority TEXT NOT NULL DEFAULT '',
	mrz_line1 TEXT NOT NULL DEFAULT '',
	mrz_line2 TEXT NOT NULL DEFAULT '',
	country_code TEXT NOT NULL DEFAULT '',
	confidence NUMERIC(5,2) NOT NULL DEFAULT 0,
	passport_warning_sent_flags INTEGER NOT NULL DEFAULT 0,
	extracted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_passport_data_candidate_id
	ON passport_data(candidate_id);

CREATE INDEX IF NOT EXISTS idx_passport_data_expiry_date
	ON passport_data(expiry_date)
	WHERE expiry_date IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_passport_data_passport_number
	ON passport_data(passport_number)
	WHERE passport_number <> '';

DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_passport_data_updated_at') THEN
		CREATE TRIGGER trg_passport_data_updated_at
		BEFORE UPDATE ON passport_data
		FOR EACH ROW
		EXECUTE FUNCTION set_updated_at();
	END IF;
END
$$;
