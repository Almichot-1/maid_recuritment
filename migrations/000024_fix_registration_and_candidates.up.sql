-- Create ENUM if it doesn't exist
DO $$ BEGIN
    CREATE TYPE account_status AS ENUM ('pending_approval', 'active', 'rejected', 'suspended');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Add missing columns to users
ALTER TABLE users 
    ADD COLUMN IF NOT EXISTS account_status account_status NOT NULL DEFAULT 'active',
    ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT FALSE;

-- Add missing columns to candidates
ALTER TABLE candidates 
    ADD COLUMN IF NOT EXISTS experience_country JSONB NOT NULL DEFAULT '[]'::jsonb;
