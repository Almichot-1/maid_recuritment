# Migration Fix Summary

## Problem
The candidate publish feature was returning 404 "candidate not found" errors after deploying the `country_of_experience` feature.

## Root Cause
When Render tried to run database migrations during the build process, it failed with permission errors:
```
ERROR: must be owner of table candidates (SQLSTATE 42501)
```

The migration script at `scripts/run_migrations.sh` was using the standard Supabase connection string, which doesn't have ALTER TABLE permissions. This caused:
1. Migration 000026 (add country_of_experience) to fail
2. Backend code expected the column but it didn't exist in production
3. GORM queries failed due to schema mismatch
4. Publish endpoint returned "candidate not found"

## Solution Applied
Used Supabase MCP tools to apply the migration directly:
```sql
ALTER TABLE candidates
ADD COLUMN IF NOT EXISTS country_of_experience TEXT;

COMMENT ON COLUMN candidates.country_of_experience IS 'Country where the candidate gained work experience (only relevant if experience_years > 0)';
```

✅ Migration successfully applied via `mcp_Supabase_apply_migration`

## Testing Needed
1. Try publishing candidate ID: `183a97c1-3e16-4726-88a9-4d8f5bb310c2`
2. Test creating a new candidate with country_of_experience
3. Test CV generation with country field
4. Verify the field shows in generated CVs

## Status
- ✅ Database schema updated
- ✅ Backend code deployed
- ✅ Frontend code deployed
- ⏳ Pending user verification

## Next Steps
All code changes are complete and deployed. The `country_of_experience` feature should now work end-to-end. Test the publish functionality to confirm the fix.
