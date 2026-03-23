DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'admin_role') THEN
		CREATE TYPE admin_role AS ENUM ('super_admin', 'support_admin');
	END IF;
END
$$;

DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'account_status') THEN
		CREATE TYPE account_status AS ENUM ('pending_approval', 'active', 'rejected', 'suspended');
	END IF;
END
$$;

DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'agency_approval_status') THEN
		CREATE TYPE agency_approval_status AS ENUM ('pending', 'approved', 'rejected');
	END IF;
END
$$;

DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'account_status'
	) THEN
		ALTER TABLE users
		ADD COLUMN account_status account_status NOT NULL DEFAULT 'active';
	END IF;
END
$$;

UPDATE users
SET account_status = 'active'
WHERE account_status IS NULL;

CREATE TABLE IF NOT EXISTS admins (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	email TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	full_name TEXT NOT NULL,
	role admin_role NOT NULL,
	mfa_secret TEXT NOT NULL,
	is_active BOOLEAN NOT NULL DEFAULT TRUE,
	failed_login_attempts INTEGER NOT NULL DEFAULT 0,
	locked_until TIMESTAMPTZ,
	last_login TIMESTAMPTZ,
	force_password_change BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS audit_logs (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	admin_id UUID NOT NULL REFERENCES admins(id) ON DELETE RESTRICT,
	action TEXT NOT NULL,
	target_type TEXT,
	target_id UUID,
	details JSONB NOT NULL DEFAULT '{}'::jsonb,
	ip_address TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS agency_approval_requests (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	agency_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
	status agency_approval_status NOT NULL DEFAULT 'pending',
	reviewed_by UUID REFERENCES admins(id) ON DELETE SET NULL,
	reviewed_at TIMESTAMPTZ,
	rejection_reason TEXT,
	admin_notes TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_account_status ON users(account_status);
CREATE INDEX IF NOT EXISTS idx_admins_role ON admins(role);
CREATE INDEX IF NOT EXISTS idx_audit_logs_admin_id ON audit_logs(admin_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_agency_approval_requests_status ON agency_approval_requests(status);

DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_admins_updated_at') THEN
		CREATE TRIGGER trg_admins_updated_at
		BEFORE UPDATE ON admins
		FOR EACH ROW
		EXECUTE FUNCTION set_updated_at();
	END IF;
END
$$;

DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_agency_approval_requests_updated_at') THEN
		CREATE TRIGGER trg_agency_approval_requests_updated_at
		BEFORE UPDATE ON agency_approval_requests
		FOR EACH ROW
		EXECUTE FUNCTION set_updated_at();
	END IF;
END
$$;
