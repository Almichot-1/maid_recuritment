ALTER TABLE public.users
  DROP COLUMN IF EXISTS default_foreign_pairing_id,
  DROP COLUMN IF EXISTS auto_share_candidates;
