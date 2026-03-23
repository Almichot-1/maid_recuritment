ALTER TABLE selections
  ADD COLUMN IF NOT EXISTS employer_contract_url TEXT,
  ADD COLUMN IF NOT EXISTS employer_contract_file_name TEXT,
  ADD COLUMN IF NOT EXISTS employer_contract_uploaded_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS employer_id_url TEXT,
  ADD COLUMN IF NOT EXISTS employer_id_file_name TEXT,
  ADD COLUMN IF NOT EXISTS employer_id_uploaded_at TIMESTAMPTZ;
