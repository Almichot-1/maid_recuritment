ALTER TABLE candidates DROP COLUMN IF EXISTS experience_country;
ALTER TABLE users DROP COLUMN IF EXISTS email_verified;
ALTER TABLE users DROP COLUMN IF EXISTS account_status;
DROP TYPE IF EXISTS account_status;
