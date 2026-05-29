CREATE TABLE IF NOT EXISTS medical_data (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    candidate_id UUID NOT NULL UNIQUE REFERENCES candidates(id) ON DELETE CASCADE,
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    expiry_date TIMESTAMPTZ NOT NULL,
    raw_text TEXT NOT NULL DEFAULT '',
    extracted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    warning_sent_flags INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_medical_data_expiry_date ON medical_data (expiry_date);
