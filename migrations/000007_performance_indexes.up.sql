CREATE INDEX IF NOT EXISTS idx_notifications_user_unread_created_at
  ON notifications (user_id, is_read, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_candidates_created_by_status_created_at
  ON candidates (created_by, status, created_at DESC)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_candidate_pair_shares_pairing_active_candidate
  ON candidate_pair_shares (pairing_id, is_active, candidate_id);

CREATE INDEX IF NOT EXISTS idx_selections_pairing_status_created_at
  ON selections (pairing_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_selections_selected_by_pairing_status_created_at
  ON selections (selected_by, pairing_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_documents_candidate_type
  ON documents (candidate_id, document_type);
