CREATE TABLE passport_data (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    candidate_id UUID NOT NULL UNIQUE REFERENCES candidates(id) ON DELETE CASCADE,
    holder_name TEXT NOT NULL,
    passport_number TEXT NOT NULL,
    country_code TEXT,
    nationality TEXT NOT NULL,
    date_of_birth TIMESTAMPTZ NOT NULL,
    place_of_birth TEXT,
    gender TEXT NOT NULL,
    issue_date TIMESTAMPTZ,
    expiry_date TIMESTAMPTZ NOT NULL,
    issuing_authority TEXT,
    mrz_line_1 TEXT NOT NULL,
    mrz_line_2 TEXT NOT NULL,
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0,
    extracted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    passport_warning_sent_flags INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_passport_data_expiry_date ON passport_data(expiry_date);
CREATE INDEX idx_passport_data_candidate_id ON passport_data(candidate_id);

CREATE TRIGGER trg_passport_data_updated_at
BEFORE UPDATE ON passport_data
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
