CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE user_role AS ENUM ('ethiopian_agent', 'foreign_agent');
CREATE TYPE candidate_status AS ENUM (
	'draft',
	'available',
	'locked',
	'under_review',
	'approved',
	'rejected',
	'in_progress',
	'completed'
);
CREATE TYPE document_type AS ENUM ('passport', 'photo', 'video');
CREATE TYPE selection_status AS ENUM ('pending', 'approved', 'rejected', 'expired');
CREATE TYPE approval_decision AS ENUM ('approved', 'rejected');
CREATE TYPE step_status AS ENUM ('pending', 'in_progress', 'completed');
CREATE TYPE notification_type AS ENUM ('selection', 'approval', 'rejection', 'status_update', 'expiry');

CREATE TABLE users (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	email TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	full_name TEXT,
	role user_role NOT NULL,
	company_name TEXT,
	is_active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE candidates (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
	full_name TEXT NOT NULL,
	age INTEGER,
	experience_years INTEGER,
	languages JSONB NOT NULL DEFAULT '[]'::jsonb,
	skills JSONB NOT NULL DEFAULT '[]'::jsonb,
	status candidate_status NOT NULL DEFAULT 'draft',
	locked_by UUID REFERENCES users(id) ON DELETE SET NULL,
	locked_at TIMESTAMPTZ,
	lock_expires_at TIMESTAMPTZ,
	cv_pdf_url TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMPTZ
);

CREATE TABLE documents (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	candidate_id UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
	document_type document_type NOT NULL,
	file_url TEXT NOT NULL,
	file_name TEXT,
	file_size BIGINT,
	uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE selections (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	candidate_id UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
	selected_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
	status selection_status NOT NULL DEFAULT 'pending',
	expires_at TIMESTAMPTZ NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE approvals (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	selection_id UUID NOT NULL REFERENCES selections(id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	decision approval_decision NOT NULL,
	decided_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE status_steps (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	candidate_id UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
	step_name TEXT NOT NULL,
	step_status step_status NOT NULL DEFAULT 'pending',
	completed_at TIMESTAMPTZ,
	notes TEXT,
	updated_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE notifications (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	title TEXT NOT NULL,
	message TEXT NOT NULL,
	type notification_type NOT NULL,
	is_read BOOLEAN NOT NULL DEFAULT FALSE,
	related_entity_type TEXT,
	related_entity_id UUID,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_candidates_status ON candidates(status);
CREATE INDEX idx_candidates_locked_by ON candidates(locked_by);
CREATE INDEX idx_selections_candidate_id ON selections(candidate_id);
CREATE UNIQUE INDEX idx_selections_one_pending_per_candidate
ON selections(candidate_id)
WHERE status = 'pending';
CREATE INDEX idx_notifications_user_id ON notifications(user_id);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
	NEW.updated_at = NOW();
	RETURN NEW;
END;
$$;

CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_candidates_updated_at
BEFORE UPDATE ON candidates
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_selections_updated_at
BEFORE UPDATE ON selections
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_status_steps_updated_at
BEFORE UPDATE ON status_steps
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
