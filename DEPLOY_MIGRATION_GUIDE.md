# Deploy Country of Experience Migration to Production

## Current Issue
The production API is returning 500 errors because the database doesn't have the `country_of_experience` column yet, but the code is trying to use it.

## Solution Options

### Option 1: Quick Fix - Rollback Frontend (Temporary)
If you need the site working immediately while preparing the migration:

1. Comment out the `country_of_experience` field in the frontend temporarily
2. Deploy the frontend
3. Then follow Option 2 to properly deploy the full feature

### Option 2: Proper Deployment (Recommended)

#### Step 1: Connect to Production Database
```bash
# Using Render's PostgreSQL connection string
# Get this from: Render Dashboard → Your Database → Connection String

psql "postgresql://user:password@host/database"
```

#### Step 2: Run the Migration Manually
```sql
-- Add country_of_experience field to candidates table
ALTER TABLE candidates
ADD COLUMN country_of_experience TEXT;

COMMENT ON COLUMN candidates.country_of_experience IS 'Country where the candidate gained work experience (only relevant if experience_years > 0)';
```

#### Step 3: Verify Migration
```sql
-- Check that the column exists
\d candidates

-- Or
SELECT column_name, data_type, is_nullable 
FROM information_schema.columns 
WHERE table_name = 'candidates' 
AND column_name = 'country_of_experience';
```

#### Step 4: Deploy Updated Backend
```bash
# Push your code to Git
git add .
git commit -m "Add country of experience field"
git push

# Render will automatically rebuild and deploy
# Or manually trigger deployment from Render Dashboard
```

#### Step 5: Deploy Updated Frontend
```bash
cd frontend
npm run build
# Deploy to Vercel or your frontend host
```

### Option 3: Using Render's Shell (If Available)

1. Go to Render Dashboard
2. Select your web service
3. Click "Shell" tab
4. Run migration command:
```bash
./bin/devmigrate
# Or if you have migrate CLI:
migrate -database "$DATABASE_URL" -path ./migrations up 1
```

## Migration File Location
- **Up Migration:** `migrations/000026_add_experience_country.up.sql`
- **Down Migration:** `migrations/000026_add_experience_country.down.sql`

## Rollback Instructions
If you need to rollback:

```sql
-- Remove country_of_experience field from candidates table
ALTER TABLE candidates
DROP COLUMN IF EXISTS country_of_experience;
```

## Testing After Deployment

1. Try creating a new candidate with experience > 0
2. Fill in the country of experience field
3. Save and verify it's stored
4. Generate a CV and check if country appears
5. Try editing and updating the field

## Troubleshooting

### Error: "column does not exist"
- The migration hasn't run yet
- Run the ALTER TABLE command manually

### Error: "column already exists"
- Migration already ran
- Check if there are code issues instead

### Frontend shows field but backend fails
- Migration hasn't been applied to production database
- Apply migration first, then deploy backend

## Important Notes

- ✅ The column is nullable (TEXT) - safe to add
- ✅ No data migration needed for existing records
- ✅ Backward compatible - old code will ignore the field
- ✅ New code handles NULL gracefully
- ⚠️ **Must run migration BEFORE deploying new backend code**

## Recommended Deployment Order

1. **First:** Run database migration
2. **Second:** Deploy backend with new code
3. **Third:** Deploy frontend with new field
4. **Finally:** Test end-to-end

This order ensures zero downtime and no errors!
