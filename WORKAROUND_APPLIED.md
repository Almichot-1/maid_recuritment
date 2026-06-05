# CV Generation Workaround Applied

## Problem
The `/generate-cv` endpoint returns 404 because Render is stuck running old code that doesn't have this route.

## Root Cause
Render deployments are not picking up the latest code from GitHub, even though:
- ✅ Code exists locally
- ✅ Code pushed to GitHub
- ✅ Render set to auto-deploy
- ❌ Render still running old build

## Workaround Applied

Changed the frontend to use the **`/download-cv`** endpoint instead, which **DOES work**:

### Before (Broken):
```typescript
const response = await api.post(`/candidates/${id}/generate-cv`, payload);
```

### After (Working):
```typescript
const response = await api.get(`/candidates/${id}/download-cv`);
```

## What This Means

✅ **CV generation will work again** once frontend redeploys (5-10 min)

⚠️ **Limitation:** Can't pass custom branding (logo/company name) for now

✅ **Everything else works:** Creating candidates, filling country field, etc.

## Testing

After frontend redeploys (should be done in ~10 minutes):

1. Go to candidate detail page
2. Click "Generate CV"
3. Should work without 404!
4. CV will download/generate successfully

## Long-term Fix Needed

This is a **workaround**, not a permanent solution. You need to fix Render deployments:

### Option 1: Manual Deploy on Render
1. Go to https://dashboard.render.com
2. Select API service
3. Click "Manual Deploy" → "Clear build cache & deploy"
4. This forces a fresh build

### Option 2: Contact Render Support
If manual deploy doesn't work, something is wrong with your Render setup:
- Auto-deploy not triggering
- Build cache preventing updates
- Deployment hooks failing silently

### Option 3: Redeploy from Scratch
1. Delete current Render service
2. Create new one
3. Connect to same GitHub repo
4. Will pull latest code

## Files Changed

**Commit:** `abbd1eb` - "Workaround: use download-cv instead of generate-cv"

**File:** `frontend/src/hooks/use-candidates.ts`

**Change:** Switched from POST `/generate-cv` to GET `/download-cv`

## Deployment Status

- ✅ Code pushed to GitHub
- ⏳ Frontend redeploying (Vercel/your host)
- ⏳ Wait ~10 minutes for frontend
- ✅ Then test CV generation

## Notes

- This workaround uses the simpler download endpoint
- The generate endpoint is more feature-rich (allows branding)
- But download is good enough for now
- Once Render fixes itself, we can switch back

---

**Status:** Workaround deployed, wait 10 min then test  
**Expected:** CV generation works  
**Limitation:** No custom branding support
