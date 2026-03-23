DROP TRIGGER IF EXISTS trg_agency_approval_requests_updated_at ON agency_approval_requests;
DROP TRIGGER IF EXISTS trg_admins_updated_at ON admins;

DROP INDEX IF EXISTS idx_agency_approval_requests_status;
DROP INDEX IF EXISTS idx_audit_logs_created_at;
DROP INDEX IF EXISTS idx_audit_logs_admin_id;
DROP INDEX IF EXISTS idx_admins_role;
DROP INDEX IF EXISTS idx_users_account_status;

DROP TABLE IF EXISTS agency_approval_requests;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS admins;

ALTER TABLE users
DROP COLUMN IF EXISTS account_status;

DROP TYPE IF EXISTS agency_approval_status;
DROP TYPE IF EXISTS account_status;
DROP TYPE IF EXISTS admin_role;
