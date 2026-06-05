# 🚨 URGENT: Fix 500 Error - Deployment Steps

## Current Situation
- ❌ Production is throwing 500 errors when creating candidates
- ❌ New `country_of_experience` field is causing database errors
- ✅ Code has been updated to temporarily remove this field

## Quick Fix (Do This NOW)

### Step 1: Deploy Updated Frontend
The code now **completely removes** the `country_of_experience` field before sending to API.

```bash
cd frontend
npm run build
# Then deploy to Vercel/your hosting
```

**Expected Result:** 500 errors will stop immediately

### Step 2: Verify Fix
1. Try creating a candidate
2. Should work without errors
3. Country field will appear in form but won't be saved (that's OK for now)

## Full Feature Deployment (Do This Later)

### Step 1: Run Database Migration

#### Option A: Via Render Shell
1. Go to Render Dashboard → Your API Service
2. Click "Shell" tab
3. Run:
```bash
psql $DATABASE_URL -c "ALTER TABLE candidates ADD COLUMN country_of_experience TEXT;"
```

#### Option B: Direct Database Connection
```bash
psql "YOUR_DATABASE_URL_HERE"
```
Then:
```sql
ALTER TABLE candidates ADD COLUMN country_of_experience TEXT;
```

### Step 2: Update Frontend Code
Remove the temporary fix in `frontend/src/hooks/use-candidates.ts`:

Change this:
```typescript
// TEMPORARY: Always remove country_of_experience until production DB is migrated
const payload = { ...data };
delete payload.country_of_experience;
```

Back to this:
```typescript
// Remove country_of_experience if it's empty
const payload = { ...data };
if (!payload.country_of_experience) {
  delete payload.country_of_experience;
}
```

### Step 3: Deploy Updated Frontend
```bash
cd frontend
npm run build
# Deploy to hosting
```

### Step 4: Deploy Backend (if needed)
```bash
git push origin main
# Render will auto-deploy
```

## Testing Checklist

### After Quick Fix
- [ ] Can create candidates without errors
- [ ] Can update candidates without errors
- [ ] Form shows country field when experience > 0
- [ ] Country values are not saved (expected)

### After Full Deployment
- [ ] Can create candidates with country
- [ ] Can update candidates with country  
- [ ] Country appears in generated CVs
- [ ] Existing candidates still work

## Rollback Plan

If something goes wrong after full deployment:

1. **Rollback Frontend:**
   - Deploy previous version
   - Or keep the `delete payload.country_of_experience` line

2. **Rollback Database:**
```sql
ALTER TABLE candidates DROP COLUMN IF EXISTS country_of_experience;
```

## Current Code State

✅ **Fixed Issues:**
- Form default value changed from `""` to `undefined` (fixes React warning)
- API calls completely remove the field (fixes 500 error)
- Backend code ready for migration

⏳ **Waiting For:**
- Frontend deployment (fixes 500 error)
- Database migration (enables full feature)

## Support

If you encounter issues:

1. Check Render logs for backend errors
2. Check browser console for frontend errors
3. Verify database migration ran successfully:
```sql
\d candidates
-- Should show country_of_experience column
```

## Timeline

**Right Now:**
- Deploy quick fix frontend → Stops 500 errors

**Within 1-24 hours:**
- Run database migration → Enables feature
- Update frontend code → Enables saving countries
- Deploy full feature → Everything works!

---

**Priority: HIGH**
**Action: Deploy updated frontend ASAP to fix production**
