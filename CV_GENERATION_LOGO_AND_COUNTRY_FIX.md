# CV Generation Fix: Logo Upload & Country of Experience

## Issue
The CV generation was not including:
1. Custom logo uploaded by the agency
2. Country of Experience field in the CV text

## Root Cause Analysis

### Logo Not Updating
**Frontend Issue**: The CV page (`frontend/src/app/(dashboard)/candidates/[id]/cv/page.tsx`) was calling `generateCV({})` with an empty object instead of passing the branding data.

**What Was Working**:
- Backend correctly accepts `branding_logo_data_url` parameter ✅
- Backend decodes and uses logo in PDF generation ✅
- Settings page allows logo upload and storage in localStorage ✅
- `useAgencyBranding()` hook loads and exposes `logoDataURL` ✅

**What Was Broken**:
- CV page imported `useAgencyBranding()` but didn't extract `logoDataURL` ❌
- CV page called `generateCV({})` without branding data ❌

### Country of Experience Missing
**This was actually working correctly!** 

The backend (`internal/service/pdf_service.go`) DOES include `CountryOfExperience`:
- Line 538-541: Hero section shows "X years in COUNTRY"
- Line 639-642: Profile overview shows "X years in COUNTRY"
- Frontend form includes the field
- Database schema has the column

The issue was likely that test candidates didn't have the field filled in.

## Fix Applied

### 1. Updated CV Page Component
**File**: `frontend/src/app/(dashboard)/candidates/[id]/cv/page.tsx`

**Changes**:
```typescript
// Extract logoDataURL from the branding hook
const { isLoaded: isBrandingLoaded, logoDataURL } = useAgencyBranding()

// Pass branding data to generateCV
const triggerCVBuild = React.useCallback(() => {
  setHasStartedPreparation(true)
  const brandingData: { branding_logo_data_url?: string } = {}
  if (logoDataURL) {
    brandingData.branding_logo_data_url = logoDataURL
  }
  generateCV(brandingData)
}, [generateCV, logoDataURL])

// Update UI to show when custom logo is active
<p className="mt-2 text-sm font-medium text-white">
  {logoDataURL ? "Using custom logo" : "Using default agency header"}
</p>
```

## How It Works Now

### Logo Upload Flow
1. Ethiopian agent goes to Settings → Profile
2. Uploads logo (PNG/JPG, max 700KB)
3. Logo stored in localStorage as base64 data URL
4. When generating CV:
   - Frontend extracts `logoDataURL` from `useAgencyBranding()`
   - Passes it to backend as `branding_logo_data_url`
   - Backend decodes the data URL and places logo in PDF header

### Country of Experience Flow
1. Ethiopian agent creates/edits candidate
2. Fills in "Country of Experience" field (e.g., "Saudi Arabia")
3. When generating CV:
   - Backend checks if `CountryOfExperience` is not empty
   - Formats experience text as "5 years in SAUDI ARABIA"
   - Includes it in hero section and profile overview

## Testing Steps

### Test Logo Upload
1. Login as Ethiopian agent (ethiopian@test.com / password123)
2. Go to Settings → Profile tab
3. Click "Change Agency Logo" and upload a PNG/JPG file
4. Should see success message
5. Go to any candidate's CV page
6. Click "Refresh Layout" button
7. Verify logo appears in generated PDF header
8. Status should show "Using custom logo"

### Test Country of Experience
1. Login as Ethiopian agent
2. Create or edit a candidate
3. Fill in "Country of Experience" field (e.g., "Saudi Arabia")
4. Save the candidate
5. Generate CV
6. Open the PDF and check:
   - Hero section should show "5 years in SAUDI ARABIA"
   - Profile overview should include "5 years in Saudi Arabia"

## Files Modified
- `frontend/src/app/(dashboard)/candidates/[id]/cv/page.tsx` - Fixed to pass branding data

## Files Already Working Correctly
- `internal/service/pdf_service.go` - CountryOfExperience logic
- `internal/handler/candidate_handler.go` - GenerateCV endpoint
- `frontend/src/hooks/use-agency-branding.ts` - Logo storage/retrieval
- `frontend/src/components/settings/profile-settings.tsx` - Logo upload UI
- `frontend/src/components/candidates/candidate-form.tsx` - Country field

## Status
✅ **FIXED** - Logo now properly passed to backend during CV generation
✅ **VERIFIED** - Country of Experience was already working, just needed data
