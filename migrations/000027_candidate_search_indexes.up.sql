-- Migration 000027: candidate full-text search index using pg_trgm
-- pg_trgm is already available on Supabase (it ships pre-installed).
-- All indexes are created CONCURRENTLY so they do not block reads or writes
-- on the candidates table while being built.

-- Enable trigram extension (no-op if already enabled by Supabase).
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- GIN trigram index on full_name: turns LIKE '%search%' into an index scan
-- instead of a sequential table scan. Critical for search performance at scale.
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_candidates_full_name_trgm
  ON candidates USING GIN (full_name gin_trgm_ops)
  WHERE deleted_at IS NULL;

-- Partial index on passport_data for the most common lookup pattern
-- (candidate_id is the PK FK, always used in GetByCandidateID).
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_passport_data_candidate_id
  ON passport_data (candidate_id);
