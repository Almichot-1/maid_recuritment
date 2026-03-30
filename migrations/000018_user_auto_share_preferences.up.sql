ALTER TABLE public.users
  ADD COLUMN IF NOT EXISTS auto_share_candidates BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS default_foreign_pairing_id UUID NULL;
