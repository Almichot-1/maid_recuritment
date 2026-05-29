-- PostgreSQL does not support removing values from an existing ENUM type.
-- To fully revert, recreate the type without 'medical' and update all references.
-- This is intentionally a no-op to avoid destructive schema changes on rollback.
SELECT 1;
