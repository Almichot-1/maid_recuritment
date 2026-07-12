-- Migration: Add destination_country, departure_date, arrival document fields to selection_progress
-- Also extend arrival_status CHECK to include 'in_transit'

-- First, alter the CHECK constraint on arrival_status
ALTER TABLE selection_progress DROP CONSTRAINT IF EXISTS selection_progress_arrival_status_check;

-- Add new columns
ALTER TABLE selection_progress
  ADD COLUMN IF NOT EXISTS destination_country TEXT,
  ADD COLUMN IF NOT EXISTS departure_date DATE,
  ADD COLUMN IF NOT EXISTS arrival_document_url TEXT,
  ADD COLUMN IF NOT EXISTS arrival_document_name TEXT,
  ADD COLUMN IF NOT EXISTS arrival_uploaded_at TIMESTAMPTZ;

-- Re-add the CHECK constraint with 'in_transit'
ALTER TABLE selection_progress
  ADD CONSTRAINT selection_progress_arrival_status_check
  CHECK (arrival_status IN ('not_arrived', 'in_transit', 'arrived'));

-- Migrate destination_country from legacy status_steps before it's dropped
UPDATE selection_progress sp
SET destination_country = ss.destination_country
FROM selections s
JOIN status_steps ss ON ss.candidate_id = s.candidate_id AND ss.step_name = 'Ticket'
WHERE sp.selection_id = s.id
  AND sp.destination_country IS NULL
  AND ss.destination_country IS NOT NULL;

COMMENT ON COLUMN selection_progress.destination_country IS 'The country the candidate is traveling to';
COMMENT ON COLUMN selection_progress.departure_date IS 'Date the candidate departs (set when in_transit)';
COMMENT ON COLUMN selection_progress.arrival_document_url IS 'Uploaded proof of arrival (e.g. stamped passport, handover form)';
COMMENT ON COLUMN selection_progress.arrival_document_name IS 'Original filename of the arrival document';
COMMENT ON COLUMN selection_progress.arrival_uploaded_at IS 'When the arrival document was uploaded';
