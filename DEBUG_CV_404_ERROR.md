# Debug: 404 Error on CV Generation

## The Error
```
POST https://maid-recruitment-api-frankfurt.onrender.com/api/v1/candidates/{id}/generate-cv
404 (Not Found)
```

## Possible Causes

### 1. Backend Still Deploying
Render deployments take 5-10 minutes. The old backend is running while the new one builds.

**Check:**
1. Go to https://dashboard.render.com
2. Click on your API service
3. Look at "Events" tab - is it still building?
4. Wait for "Deploy live" message

### 2. User Role Issue
The route requires `EthiopianAgent` role:
```go
RequireRole(string(domain.EthiopianAgent))
```

**Check:**
1. Are you logged in as an Ethiopian Agent?
2. Try logging out and back in
3. Check your user role in the database

### 3. API URL Mismatch
The frontend might be pointing to the wrong backend.

**Check frontend .env.local:**
```bash
cd frontend
cat .env.local
```

Should show:
```
NEXT_PUBLIC_API_URL=https://maid-recruitment-api-frankfurt.onrender.com/api/v1
```

### 4. Old Backend Version Running
The deployed backend might not have the latest code.

**Check:**
1. Go to Render dashboard
2. Look at the latest deployment
3. Check the commit hash - does it match your latest push?
4. If not, manually trigger a redeploy

## Quick Tests

### Test 1: Check if backend is responding
```bash
curl https://maid-recruitment-api-frankfurt.onrender.com/health
```
Should return health status.

### Test 2: Check the candidates endpoint
```bash
curl https://maid-recruitment-api-frankfurt.onrender.com/api/v1/candidates
```
If this works, the API is up.

### Test 3: Test CV generation with auth token
Get your auth token from browser DevTools → Application → Local Storage, then:

```bash
curl -X POST \
  https://maid-recruitment-api-frankfurt.onrender.com/api/v1/candidates/YOUR_CANDIDATE_ID/generate-cv \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{}'
```

## Solutions

### Solution 1: Wait for Deployment
Just wait 10 minutes and try again. Render is probably still deploying.

### Solution 2: Force Redeploy on Render
1. Go to Render Dashboard
2. Select your API service  
3. Click "Manual Deploy" → "Clear build cache & deploy"
4. Wait for it to finish

### Solution 3: Check Render Logs
1. Go to Render Dashboard
2. Select your API service
3. Click "Logs" tab
4. Look for errors during startup
5. Look for 404 errors when you try to generate CV

### Solution 4: Verify Database Migration Ran
The new backend code expects `country_of_experience` column to exist.

```sql
-- Connect to your database and check:
\d candidates

-- Should show country_of_experience column
-- If not, run:
ALTER TABLE candidates ADD COLUMN country_of_experience TEXT;
```

## Debug Checklist

- [ ] Render deployment shows "Live" status
- [ ] Latest commit is deployed
- [ ] No errors in Render logs
- [ ] User is logged in as Ethiopian Agent
- [ ] Database migration ran successfully
- [ ] Frontend .env points to correct API URL
- [ ] Auth token is valid and not expired

## Expected Timeline

After pushing to GitHub:

1. **0-2 min:** GitHub receives push
2. **1-3 min:** Render detects changes
3. **2-8 min:** Backend builds (Go compilation)
4. **8-10 min:** Backend deploys and goes live
5. **10+ min:** Ready to test

So if you just pushed, wait at least 10 minutes before testing.

## Still Not Working?

If after 15 minutes you still get 404:

1. **Check Render logs** for routing errors
2. **Verify the route exists** in the deployed code:
   - In Render shell: `grep -r "generate-cv" .`
3. **Check if there's a rollback** - maybe Render reverted to old version
4. **Look at build errors** - maybe the build failed silently

## Working Alternative

If CV generation keeps failing, you can temporarily use the download endpoint instead:

```typescript
// Instead of:
await api.post(`/candidates/${id}/generate-cv`)

// Try:
const response = await api.get(`/candidates/${id}/download-cv`)
```

The download-cv route doesn't require role restrictions and might work.
