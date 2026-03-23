DROP TRIGGER IF EXISTS trg_status_steps_updated_at ON status_steps;
DROP TRIGGER IF EXISTS trg_selections_updated_at ON selections;
DROP TRIGGER IF EXISTS trg_candidates_updated_at ON candidates;
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;

DROP FUNCTION IF EXISTS set_updated_at();

DROP INDEX IF EXISTS idx_selections_one_pending_per_candidate;

DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS status_steps;
DROP TABLE IF EXISTS approvals;
DROP TABLE IF EXISTS selections;
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS candidates;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS notification_type;
DROP TYPE IF EXISTS step_status;
DROP TYPE IF EXISTS approval_decision;
DROP TYPE IF EXISTS selection_status;
DROP TYPE IF EXISTS document_type;
DROP TYPE IF EXISTS candidate_status;
DROP TYPE IF EXISTS user_role;
