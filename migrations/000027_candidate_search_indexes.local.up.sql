-- Migration 000027: candidate full-text search index using pg_trgm (LOCAL DEV VERSION)
-- This version removes CONCURRENTLY since it can't run in a transaction block
-- For production, use the regular .up.sql file without the .local suffix

-- Enable trigram extension
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- GIN trigram index on full_name
CREATE INDEX IF NOT EXISTS idx_candidates_full_name_trgm
  ON candidates USING GIN (full_name gin_trgm_ops)
  WHERE deleted_at IS NULL;

-- Partial index on passport_data
CREATE INDEX IF NOT EXISTS idx_passport_data_candidate_id
  ON passport_data (candidate_id);
