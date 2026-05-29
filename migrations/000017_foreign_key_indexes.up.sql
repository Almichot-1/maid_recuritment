CREATE INDEX IF NOT EXISTS idx_agency_approval_requests_reviewed_by
  ON agency_approval_requests (reviewed_by);

CREATE INDEX IF NOT EXISTS idx_agency_pairings_approved_by_admin_id
  ON agency_pairings (approved_by_admin_id);

CREATE INDEX IF NOT EXISTS idx_approvals_selection_id
  ON approvals (selection_id);

CREATE INDEX IF NOT EXISTS idx_approvals_user_id
  ON approvals (user_id);

CREATE INDEX IF NOT EXISTS idx_candidate_pair_shares_shared_by_user_id
  ON candidate_pair_shares (shared_by_user_id);

CREATE INDEX IF NOT EXISTS idx_medical_data_document_id
  ON medical_data (document_id);

CREATE INDEX IF NOT EXISTS idx_status_steps_candidate_id
  ON status_steps (candidate_id);

CREATE INDEX IF NOT EXISTS idx_status_steps_updated_by
  ON status_steps (updated_by);
