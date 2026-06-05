# Country of Experience Field Implementation

## Summary
Added a new "Country of Experience" field that only appears when the candidate has experience years > 0. This field is fully integrated into the CV generation process.

## Changes Made

### Backend Changes

#### 1. Database Migration
- **File:** `migrations/000026_add_experience_country.up.sql`
- Added `country_of_experience` TEXT column to candidates table
- **File:** `migrations/000026_add_experience_country.down.sql`
- Rollback migration to remove the field

#### 2. Domain Model
- **File:** `internal/domain/candidate.go`
- Added `CountryOfExperience string` field to Candidate struct

#### 3. Service Layer
- **File:** `internal/service/candidate_service.go`
- Added `CountryOfExperience string` to CandidateInput struct
- Updated CreateCandidate to save country_of_experience
- Updated UpdateCandidate to update country_of_experience

#### 4. Handler Layer
- **File:** `internal/handler/candidate_handler.go`
- Added `country_of_experience` to CreateCandidateRequest struct
- Added `country_of_experience` to UpdateCandidateRequest struct
- Added `country_of_experience` to CandidateResponse struct
- Updated request-to-service mappings

- **File:** `internal/handler/candidate_response_mapper.go`
- Updated mapCandidateResponse to include country_of_experience

#### 5. Repository Layer
- **File:** `internal/repository/candidate_repository.go`
- Added country_of_experience to the update fields map

#### 6. PDF Generation (CV)
- **File:** `internal/service/pdf_service.go`
- Updated hero card description to show "X years in [Country]" when country is provided
- Updated summary block to show "X years in [Country] of experience" when country is provided
- Falls back to "X years" if country is not specified

### Frontend Changes

#### 1. Types
- **File:** `frontend/src/types/index.ts`
- Added `country_of_experience?` string to Candidate type

#### 2. Validation
- **File:** `frontend/src/lib/validations.ts`
- Added `country_of_experience` as optional trimmed string to candidateSchema

#### 3. Form Component
- **File:** `frontend/src/components/candidates/candidate-form.tsx`
- Added `country_of_experience` to CandidateFormValues type
- Added conditional field that only shows when experience_years > 0
- Field has placeholder: "e.g., Saudi Arabia, UAE, Kuwait"
- Includes helpful description: "Where did she gain this work experience? (Optional)"

#### 4. Hooks
- **File:** `frontend/src/hooks/use-candidates.ts`
- Added `country_of_experience` to CreateCandidateInput and UpdateCandidateInput interfaces
- Updated candidate-to-form mapping

#### 5. Draft System
- **File:** `frontend/src/lib/candidate-draft.ts`
- Added `country_of_experience` to CandidateDraft type

#### 6. Edit Page
- **File:** `frontend/src/app/(dashboard)/candidates/[id]/edit/page.tsx`
- Added country_of_experience to initial data mapping
- Added country_of_experience to update payload

## How It Works

1. **Form Behavior:**
   - The "Country of Experience" field is hidden by default
   - When user enters experience_years > 0, the field appears automatically
   - When user clears or sets experience_years to 0, the field hides
   - Field is optional - user can leave it blank

2. **CV Generation:**
   - If country_of_experience is provided: "5 years in Saudi Arabia"
   - If country_of_experience is empty: "5 years"
   - Appears in both the hero card and profile overview sections

3. **Data Flow:**
   ```
   Frontend Form → API Request → Service Layer → Repository → Database
   Database → Repository → Service → Handler → API Response → Frontend
   ```

## Testing Checklist

- [x] Backend compiles without errors
- [ ] Frontend build completes (pending)
- [ ] Database migration runs successfully
- [ ] Field appears/disappears based on experience_years value
- [ ] Field saves correctly on create
- [ ] Field updates correctly on edit
- [ ] CV PDF shows country when provided
- [ ] CV PDF shows just years when country not provided
- [ ] Field validation works correctly

## Migration Instructions

1. Run the database migration:
   ```bash
   # Development
   go run cmd/devmigrate/main.go

   # Production
   ./migrate -database "postgres://..." -path ./migrations up
   ```

2. Rebuild the backend:
   ```bash
   go build -o ./bin/api ./cmd/api
   ```

3. Rebuild the frontend:
   ```bash
   cd frontend
   npm run build
   ```

4. Restart services to pick up the changes

## Notes

- Field is completely optional
- Backward compatible - existing candidates without this field will work fine
- No data migration needed for existing records
- Field only shows in form when relevant (experience > 0)
