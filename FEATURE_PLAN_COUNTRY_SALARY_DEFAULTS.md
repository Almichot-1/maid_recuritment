# Feature Plan: Country Applied + Salary Defaults & Dual Agency Branding

## Current Issues to Fix First
1. **Backend Connection Reset** - CV generation endpoint sometimes gets connection reset from frontend
2. **Need dual agency branding** - Currently only shows one logo/company name

## NEW FEATURES REQUESTED

### Feature 1: Country Applied & Salary with Smart Defaults

#### Problem
Ethiopian agents fill the same country and salary for many candidates (e.g., 100 candidates to KSA with 1000 SR), then switch to a different country/salary batch (e.g., Beirut with 200 USD).

#### Solution: Form Defaults with Memory

**1. Add New Fields to Candidate**
- `country_applied` (string) - e.g., "Saudi Arabia", "Lebanon", "UAE"
- `salary_offered` (string) - e.g., "1000 SR", "200 USD", "500 KWD"
- `country_of_experience` (string) - **ALREADY EXISTS** - e.g., "Saudi Arabia", "Ethiopia"
  - ✅ Already in database schema
  - ✅ Already in candidate form
  - ✅ Already shown in generated CV (in experience text)
  - 🔧 Needs to be added to Application Profile Sheet in PDF

**2. Smart Default System**
- Store last used values in **localStorage** (per user)
- Pre-fill form with last used values
- User can change anytime
- Values persist until changed

**3. UI Design**
```
┌─────────────────────────────────────┐
│ Job Details                         │
├─────────────────────────────────────┤
│ Country Applied: [Saudi Arabia ▼]  │ <- Dropdown with common countries
│ Salary Offered:  [1000 SR      ]   │ <- Text input (free form)
│                                      │
│ ℹ️ These values are remembered and   │
│   auto-filled for your next         │
│   candidate to save time.           │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ Experience Details (Already Exists) │
├─────────────────────────────────────┤
│ Years of Experience: [5]            │
│ Country of Experience: [Ethiopia ▼] │ <- Already in form
│                                      │
│ ✅ Shows in CV as "5 years in       │
│    ETHIOPIA" in profile sections    │
└─────────────────────────────────────┘
```

**4. Country Dropdown Options**
- Saudi Arabia
- United Arab Emirates
- Kuwait
- Qatar
- Bahrain
- Oman
- Lebanon
- Jordan
- (Allow custom input too)

**5. Storage Keys**
```javascript
localStorage keys:
- `candidate_defaults:${userId}:country_applied`
- `candidate_defaults:${userId}:salary_offered`
```

**6. Form Behavior**
- On form load: Read from localStorage and set as default values
- User can override anytime
- On form submit: Save new values to localStorage for next candidate

---

### Feature 2: Dual Agency Branding in CV Header

#### Current State
CV header shows only one agency (using uploaded logo or default text)

#### New Design
```
┌─────────────────────────────────────────────────────────────┐
│  [Foreign Agency Logo/Name]        [Ethiopian Agency Logo]  │
│   (Left Side)                           (Right Side)        │
└─────────────────────────────────────────────────────────────┘
```

#### Implementation Plan

**1. Data Sources**

**Foreign Agency (Left Side):**
- Comes from the **pairing** between Ethiopian and Foreign agents
- Foreign agent uploads logo in their Settings
- Foreign agent's company name from their profile

**Ethiopian Agency (Right Side):**
- Current logo upload system (already works)
- Ethiopian agent's company name from their profile

**2. Backend Changes**

**A. Update GenerateCV Request**
```go
type GenerateCVRequest struct {
    // Ethiopian agency (right side) - current fields
    BrandingLogoDataURL *string `json:"branding_logo_data_url"`
    CompanyName         *string `json:"company_name"`
    
    // Foreign agency (left side) - NEW fields
    ForeignAgencyLogoDataURL *string `json:"foreign_agency_logo_data_url"`
    ForeignAgencyName        *string `json:"foreign_agency_name"`
}
```

**B. PDF Service Header Update**
- Draw two logos/names side by side
- Left side: Foreign agency (if available)
- Right side: Ethiopian agency (if available)
- Fallback to single logo if only one provided
- Show text names if no logos

**3. Frontend Changes**

**A. Fetch Foreign Agency Data**
- Need to get the pairing info when generating CV
- Extract foreign agent's company name and logo
- Foreign agent needs logo upload feature (same as Ethiopian)

**B. CV Generation Call**
```typescript
generateCV({
  // Ethiopian agency (right)
  branding_logo_data_url: ethiopianLogoDataURL,
  company_name: ethiopianAgentProfile.company_name,
  
  // Foreign agency (left)
  foreign_agency_logo_data_url: foreignLogoDataURL,
  foreign_agency_name: foreignAgencyName,
})
```

**4. Settings Page Enhancement**
- Both Ethiopian AND Foreign agents can upload logos
- Add "Company Name" field in profile settings
- Logo stored in localStorage per user

---

## IMPLEMENTATION PHASES

### Phase 1: Fix Current Issues (URGENT)
- [ ] Fix backend connection reset on CV generation
- [ ] Verify logo upload size limits (currently 700KB)

### Phase 2: Country Applied & Salary Defaults
- [ ] Migration: Add `country_applied` and `salary_offered` columns
- [ ] Backend: Update candidate model and validation
- [ ] Frontend: Add fields to candidate form
- [ ] Frontend: Create localStorage hook for defaults
- [ ] Frontend: Add country dropdown with common options
- [ ] UI: Add info message about remembered defaults

### Phase 3: Dual Agency Branding
- [ ] Backend: Add foreign agency fields to GenerateCV request
- [ ] Backend: Update PDF service to draw dual headers
- [ ] Frontend: Add "Company Name" field to profile settings
- [ ] Frontend: Enable logo upload for Foreign agents too
- [ ] Frontend: Fetch pairing data to get foreign agency info
- [ ] Frontend: Pass both agencies to generateCV call
- [ ] PDF: Design dual logo layout (left/right split)

---

## DATABASE CHANGES NEEDED

### Migration: Add Country Applied & Salary Fields

```sql
-- Migration up
ALTER TABLE candidates 
ADD COLUMN country_applied VARCHAR(100),
ADD COLUMN salary_offered VARCHAR(100);

CREATE INDEX idx_candidates_country_applied ON candidates(country_applied);

-- Migration down
ALTER TABLE candidates 
DROP COLUMN country_applied,
DROP COLUMN salary_offered;

DROP INDEX IF EXISTS idx_candidates_country_applied;
```

### Migration: Add Company Name to Users

```sql
-- Migration up
ALTER TABLE users 
ADD COLUMN company_name VARCHAR(255);

-- Migration down
ALTER TABLE users 
DROP COLUMN company_name;
```

---

## STORAGE STRATEGY

### Ethiopian Agent's Logo & Company
- Logo: `localStorage` key: `agency_branding:${ethiopianUserId}`
- Company Name: Database `users.company_name`

### Foreign Agent's Logo & Company
- Logo: `localStorage` key: `agency_branding:${foreignUserId}`
- Company Name: Database `users.company_name`

### Form Defaults (Country & Salary)
- `localStorage` key: `candidate_defaults:${userId}:country_applied`
- `localStorage` key: `candidate_defaults:${userId}:salary_offered`

---

## PDF HEADER LAYOUT OPTIONS

### Option A: Side by Side Logos
```
┌─────────────────────────────────────────┐
│  [Foreign Logo]    [Ethiopian Logo]     │
│  Company Name      Company Name         │
└─────────────────────────────────────────┘
```

### Option B: Text + Logos
```
┌─────────────────────────────────────────┐
│  Presented by:                          │
│  [Foreign Logo] Company Name            │
│  & [Ethiopian Logo] Company Name        │
└─────────────────────────────────────────┘
```

### Option C: Banner Style (RECOMMENDED)
```
┌─────────────────────────────────────────┐
│ [Foreign Logo]  |  CANDIDATE CV  | [Eth Logo] │
│ Foreign Agency  |                | Eth Agency │
└─────────────────────────────────────────┘
```

---

## QUESTIONS FOR APPROVAL

1. **Country/Salary Defaults** - Should these be optional or required fields?
2. **Company Name** - Where should users edit this? Profile settings or separate branding section?
3. **Foreign Agent Access** - Should foreign agents ALSO be able to upload logos and company names?
4. **PDF Header Layout** - Which option do you prefer? (A, B, or C above)
5. **Salary Format** - Free text or structured (amount + currency dropdown)?

---

## ESTIMATED IMPACT

### Phase 1 (Fix Issues): 1-2 hours
- Debug connection reset
- Test with different logo sizes

### Phase 2 (Country/Salary): 3-4 hours
- Database migration
- Backend validation
- Form fields + localStorage hook
- UI polish

### Phase 3 (Dual Branding): 5-6 hours
- Database migration for company_name
- Extend logo system to foreign agents
- Pairing data fetch logic
- PDF dual header layout
- Testing across different scenarios

**Total: ~10 hours development time**

---

## APPROVED ✅

1. ✅ Add `country_applied` and `salary_offered` fields with localStorage defaults
2. ✅ Add `company_name` field to user profiles
3. ✅ Enable logo upload for BOTH Ethiopian and Foreign agents
4. ✅ PDF header shows BOTH agencies side by side
5. ✅ PDF header layout: **Option C (Banner Style)** - Logos on both sides
6. ✅ Foreign agents CAN upload logos (for dual branding)
7. ✅ Country/Salary: **OPTIONAL** fields with smart defaults
8. ✅ Company name editable in **Profile Settings**
9. ✅ Country of Experience already in CV, add to Application Profile Sheet

---

## IMPLEMENTATION ORDER

### Phase 1: Database & Backend (30 min)
- Migration for `country_applied`, `salary_offered` columns
- Migration for `company_name` in users table
- Update candidate model validation
- Update GenerateCV request to accept foreign agency data

### Phase 2: Country/Salary Defaults (1 hour)
- Create `useCandidateDefaults` hook for localStorage
- Add fields to candidate form
- Add country dropdown with common options
- Wire up auto-fill logic

### Phase 3: Dual Agency Branding (2 hours)
- Add Company Name field to profile settings
- Enable logo upload for foreign agents
- Update CV page to fetch pairing data
- Pass both agencies to generateCV call
- Update PDF service dual header layout

### Phase 4: CV Enhancements (1 hour)
- Add Country of Experience to Application Profile Sheet
- Show Country Applied and Salary in CV
- Test all scenarios

**STARTING IMPLEMENTATION NOW!**
