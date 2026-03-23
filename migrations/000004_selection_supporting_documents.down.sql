ALTER TABLE selections
  DROP COLUMN IF EXISTS employer_contract_url,
  DROP COLUMN IF EXISTS employer_contract_file_name,
  DROP COLUMN IF EXISTS employer_contract_uploaded_at,
  DROP COLUMN IF EXISTS employer_id_url,
  DROP COLUMN IF EXISTS employer_id_file_name,
  DROP COLUMN IF EXISTS employer_id_uploaded_at;
