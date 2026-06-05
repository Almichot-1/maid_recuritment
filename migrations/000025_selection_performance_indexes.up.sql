-- Add composite index for approval decision queries
CREATE INDEX IF NOT EXISTS idx_approvals_selection_id_decision
ON approvals(selection_id, decision);

-- Add index for candidate owner queries (to optimize Ethiopian agent view)
CREATE INDEX IF NOT EXISTS idx_candidates_created_by_pairing
ON candidates(created_by, pairing_id)
WHERE status != 'archived';

-- Add composite index for selection queries by candidate
CREATE INDEX IF NOT EXISTS idx_selections_candidate_pairing_status
ON selections(candidate_id, pairing_id, status);

-- Add index for expiry queries in background job
CREATE INDEX IF NOT EXISTS idx_selections_status_expires_at
ON selections(status, expires_at)
WHERE status = 'pending';
