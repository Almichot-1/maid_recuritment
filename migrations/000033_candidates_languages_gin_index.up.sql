CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_candidates_languages_gin
  ON candidates USING GIN (languages jsonb_path_ops)
  WHERE deleted_at IS NULL;
