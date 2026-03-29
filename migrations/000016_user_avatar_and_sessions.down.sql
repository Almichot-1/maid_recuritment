DROP INDEX IF EXISTS idx_user_sessions_active;
DROP INDEX IF EXISTS idx_user_sessions_user_id_last_seen_at;
DROP TABLE IF EXISTS user_sessions;

ALTER TABLE users
DROP COLUMN IF EXISTS avatar_url;
