-- Migration rollback: Remove arrival/destination enhancements from selection_progress

ALTER TABLE selection_progress DROP CONSTRAINT IF EXISTS selection_progress_arrival_status_check;

ALTER TABLE selection_progress
  DROP COLUMN IF EXISTS destination_country,
  DROP COLUMN IF EXISTS departure_date,
  DROP COLUMN IF EXISTS arrival_document_url,
  DROP COLUMN IF EXISTS arrival_document_name,
  DROP COLUMN IF EXISTS arrival_uploaded_at;

ALTER TABLE selection_progress
  ADD CONSTRAINT selection_progress_arrival_status_check
  CHECK (arrival_status IN ('not_arrived', 'arrived'));
