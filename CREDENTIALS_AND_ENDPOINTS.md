# Local Development Credentials & API Endpoints

This document provides all credentials, endpoints, and instructions for developing locally without affecting production users.

## 🔐 Database Credentials

### PostgreSQL (Local - Docker)
```
Host: localhost
Port: 5432
Username: postgres
Password: postgres
Database: maid_recruitment
Connection String: postgres://postgres:postgres@localhost:5432/maid_recruitment?sslmode=disable
```

**CLI Access:**
```bash
docker exec -it maid-recruitment-db psql -U postgres -d maid_recruitment
```

### pgAdmin (Database Management UI)
```
URL: http://localhost:5050
Email: admin@example.com
Password: admin
```

**Steps to Connect:**
1. Open http://localhost:5050
2. Right-click "Servers" → Create → Server
3. **General Tab:**
   - Name: `maid-recruitment-dev`
4. **Connection Tab:**
   - Host: `maid-recruitment-db`
   - Port: `5432`
   - Username: `postgres`
   - Password: `postgres`
   - Database: `maid_recruitment`

---

## 🚀 Backend API Credentials & Endpoints

### API Server
```
Base URL: http://localhost:8080
API Version: v1
Full API Base: http://localhost:8080/api/v1
```

### Authentication
```
JWT Secret: your-local-dev-secret-change-in-production-12345678901234
Header: Authorization: Bearer <jwt_token>
```

---

## 📄 Passport OCR Endpoints

### 1. Parse Passport Preview (Fast, No Database Storage)
**Endpoint:** `POST /api/v1/candidates/passport/parse-preview`

**Purpose:** Extract passport data from an image WITHOUT storing it to the database. Used for form autofill preview.

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/candidates/passport/parse-preview \
  -H "Authorization: Bearer <jwt_token>" \
  -F "file=@/path/to/passport.jpg"
```

**Response (200 OK):**
```json
{
  "passport": {
    "holder_name": "JOHN DOE",
    "passport_number": "A12345678",
    "country_code": "US",
    "nationality": "AMERICAN",
    "date_of_birth": "1990-05-15T00:00:00Z",
    "place_of_birth": "New York",
    "gender": "M",
    "expiry_date": "2030-12-31T00:00:00Z",
    "issue_date": "2020-01-01T00:00:00Z",
    "issuing_authority": "U.S. Department of State",
    "mrz_line_1": "P<USADOE<<JOHN<......",
    "mrz_line_2": "A123456781USA9005155M3012318<<<<<<1",
    "confidence": 0.95,
    "extracted_at": "2026-06-08T20:15:00Z"
  }
}
```

**Error Responses:**
- `401 Unauthorized` - Missing/invalid JWT token
- `400 Bad Request` - No file provided or invalid format
- `503 Service Unavailable` - Tesseract OCR not available
- `413 Payload Too Large` - File exceeds 12MB limit

**Supported Formats:** JPG, PNG

---

### 2. Parse & Store Passport (Full, With Database Storage)
**Endpoint:** `POST /api/v1/candidates/{candidateId}/passport/parse`

**Purpose:** Extract passport data AND store it in the database for the candidate.

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/candidates/123e4567-e89b-12d3-a456-426614174000/passport/parse \
  -H "Authorization: Bearer <jwt_token>" \
  -F "file=@/path/to/passport.jpg"
```

**Response (200 OK):**
```json
{
  "id": "uuid-of-passport-record",
  "candidate_id": "123e4567-e89b-12d3-a456-426614174000",
  "holder_name": "JOHN DOE",
  "passport_number": "A12345678",
  "country_code": "US",
  "nationality": "AMERICAN",
  "date_of_birth": "1990-05-15T00:00:00Z",
  "place_of_birth": "New York",
  "gender": "M",
  "expiry_date": "2030-12-31T00:00:00Z",
  "issue_date": "2020-01-01T00:00:00Z",
  "issuing_authority": "U.S. Department of State",
  "mrz_line_1": "P<USADOE<<JOHN<......",
  "mrz_line_2": "A123456781USA9005155M3012318<<<<<<1",
  "confidence": 0.95,
  "extracted_at": "2026-06-08T20:15:00Z"
}
```

**Error Responses:**
- `404 Not Found` - Candidate doesn't exist
- `401 Unauthorized` - Missing/invalid JWT token
- `400 Bad Request` - No file provided or invalid format
- `503 Service Unavailable` - Tesseract OCR not available

---

## 📋 CV Generation Endpoints

### 1. Generate CV (Creates PDF)
**Endpoint:** `POST /api/v1/candidates/{candidateId}/cv/generate`

**Purpose:** Generate a professional PDF CV from candidate profile and documents.

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/candidates/123e4567-e89b-12d3-a456-426614174000/cv/generate \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "generated_by": "admin@example.com",
    "branding": {
      "company_name": "Maid Recruitment",
      "logo_url": "https://...",
      "primary_color": "#0066cc",
      "font_family": "Arial"
    }
  }'
```

**Response (200 OK):**
```json
{
  "cv_url": "http://localhost:9000/maid-recruitment/cvs/candidate-id-cv-timestamp.pdf",
  "file_name": "candidate-id-cv-2026-06-08.pdf",
  "generated_at": "2026-06-08T20:15:00Z",
  "size_bytes": 245632
}
```

**Error Responses:**
- `404 Not Found` - Candidate doesn't exist
- `400 Bad Request` - Missing required documents (photo/passport)
- `401 Unauthorized` - Missing/invalid JWT token
- `503 Service Unavailable` - S3 storage not available

**Requirements:**
- Candidate must have a photo document
- Candidate must have a passport document
- Both documents must have valid URLs in S3

---

### 2. Download CV
**Endpoint:** `GET /api/v1/candidates/{candidateId}/cv`

**Purpose:** Download the generated CV PDF file.

**Request:**
```bash
curl -X GET http://localhost:8080/api/v1/candidates/123e4567-e89b-12d3-a456-426614174000/cv \
  -H "Authorization: Bearer <jwt_token>" \
  -o candidate-cv.pdf
```

**Response:** Binary PDF file (200 OK)

---

## 🛠️ Setup & Installation

### 1. Install Tesseract OCR (Required for Passport Extraction)

#### Windows (via Installer - Recommended)
1. Download installer from: https://github.com/UB-Mannheim/tesseract/wiki
2. Run installer (e.g., `tesseract-ocr-w64-setup-v5.x.x.exe`)
3. Choose installation path (default: `C:\Program Files (x86)\Tesseract-OCR`)
4. In `.env.local`, set:
   ```
   TESSERACT_PATH=C:\Program Files (x86)\Tesseract-OCR\tesseract.exe
   ```

#### Windows (via Chocolatey)
```powershell
choco install tesseract
# Then set in .env.local:
# TESSERACT_PATH=C:\Program Files\Tesseract-OCR\tesseract.exe
```

#### macOS (via Homebrew)
```bash
brew install tesseract
# Then set in .env.local:
# TESSERACT_PATH=/usr/local/bin/tesseract
```

#### Linux (Ubuntu/Debian)
```bash
sudo apt-get install tesseract-ocr
# Then set in .env.local:
# TESSERACT_PATH=/usr/bin/tesseract
```

### 2. Verify Tesseract Installation
```bash
tesseract --version
```

If this works, Tesseract is properly installed and in your PATH. You can leave `TESSERACT_PATH=` empty in `.env.local`.

---

## 🧪 Testing OCR Services

### Test Passport Extraction (No Database)
```bash
# 1. Get a test passport image (or use any JPG/PNG)
# 2. Call the preview endpoint
curl -X POST http://localhost:8080/api/v1/candidates/passport/parse-preview \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -F "file=@test-passport.jpg"
```

### Test CV Generation
```bash
# 1. Create a candidate first (with documents)
# 2. Call CV generation endpoint
curl -X POST http://localhost:8080/api/v1/candidates/{candidateId}/cv/generate \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{"generated_by":"test@example.com"}'
```

---

## 📊 Service Status & Monitoring

### Check Database Health
```bash
curl -X GET http://localhost:8080/api/v1/health
```

### View Database Tables
```bash
docker exec maid-recruitment-db psql -U postgres -d maid_recruitment -c "\dt"
```

### View Passport Data in Database
```bash
docker exec maid-recruitment-db psql -U postgres -d maid_recruitment -c "SELECT * FROM passport_data LIMIT 5;"
```

### Check S3 Storage (if configured)
```bash
# MinIO UI (if using local MinIO): http://localhost:9001
# Access key: minioadmin
# Secret key: minioadmin
```

---

## 🔄 Workflow: Complete Candidate Intake

### Step 1: Create Candidate
```bash
curl -X POST http://localhost:8080/api/v1/candidates \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "John Doe",
    "nationality": "American",
    "date_of_birth": "1990-05-15",
    "age": 35,
    "place_of_birth": "New York",
    "religion": "Christian",
    "marital_status": "Single",
    "children_count": 0,
    "education_level": "Bachelor",
    "experience_years": 5,
    "country_of_experience": "USA",
    "languages": ["English", "Spanish"],
    "skills": ["Leadership", "Communication"]
  }'
```

### Step 2: Upload Passport & Extract Data
```bash
curl -X POST http://localhost:8080/api/v1/candidates/{candidateId}/documents/upload \
  -H "Authorization: Bearer <jwt_token>" \
  -F "file=@passport.jpg" \
  -F "document_type=passport"

# Then extract passport data
curl -X POST http://localhost:8080/api/v1/candidates/{candidateId}/passport/parse \
  -H "Authorization: Bearer <jwt_token>" \
  -F "file=@passport.jpg"
```

### Step 3: Upload Photo
```bash
curl -X POST http://localhost:8080/api/v1/candidates/{candidateId}/documents/upload \
  -H "Authorization: Bearer <jwt_token>" \
  -F "file=@photo.jpg" \
  -F "document_type=photo"
```

### Step 4: Generate CV
```bash
curl -X POST http://localhost:8080/api/v1/candidates/{candidateId}/cv/generate \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{"generated_by": "user@example.com"}'
```

### Step 5: Download CV
```bash
curl -X GET http://localhost:8080/api/v1/candidates/{candidateId}/cv \
  -H "Authorization: Bearer <jwt_token>" \
  -o candidate-cv.pdf
```

---

## 📝 Important Notes

### ⚠️ OCR Limitations
- OCR works best with **clear, well-lit, high-resolution images** (300+ DPI recommended)
- **Best formats:** JPG, PNG
- **Maximum file size:** 12MB
- **Confidence score:** Higher is better (0.95+ is excellent)

### 🔒 Data Isolation
- All changes are **only in your local database** (Docker container)
- Production Supabase database is **completely untouched**
- You can safely test, delete, and recreate data without affecting users

### 💾 Persistence
- Database data persists even after `docker-compose down`
- To completely reset: `docker-compose down -v`

### 🚀 Performance Tips
1. **Passport extraction is slow first time** (~3-5 seconds) due to Tesseract initialization
2. **Subsequent extractions are cached** for 15 minutes
3. **CV generation takes ~2-3 seconds** due to PDF rendering
4. Use **preview endpoint** for real-time form validation instead of full storage

### 🔑 API Security (Local Development)
- JWT tokens are required for all endpoints (even locally)
- Use the JWT secret from `.env.local` to generate test tokens
- Don't commit `.env.local` to version control

---

## 🐛 Troubleshooting

### Tesseract Not Found Error
```
Error: passport OCR is unavailable
```
**Solution:**
1. Install Tesseract (see Installation section above)
2. Verify with: `tesseract --version`
3. Set correct path in `.env.local` if needed

### OCR Service Unavailable
```
HTTP 503: passport OCR is not configured
```
**Solution:**
- Ensure Tesseract is installed and in PATH
- Restart the API: `make run-api`
- Check logs for Tesseract initialization errors

### CV Generation Fails - Missing Documents
```
Error: missing required documents for CV generation
```
**Solution:**
- Upload both photo and passport documents
- Ensure they have valid S3 URLs
- Check database: `SELECT * FROM documents WHERE candidate_id = '...';`

### Database Connection Refused
```
Error: failed to connect to postgres://postgres:postgres@localhost:5432
```
**Solution:**
- Start Docker: `docker-compose up -d`
- Verify containers: `docker-compose ps`
- Check logs: `docker-compose logs postgres`

### Image Upload Fails
```
Error: file size exceeds limit
```
**Solution:**
- Maximum file size is 12MB
- Compress images before upload
- Use web image format (JPG with 85% quality)

---

## 🔗 Quick Links

- **API Documentation:** [See README.md](README.md)
- **Database Setup:** [See Migrations](migrations/)
- **Local Dev Guide:** [See LOCAL_DEV_SETUP.md](LOCAL_DEV_SETUP.md)
- **Frontend Setup:** [See frontend/README.md](frontend/README.md)

---

## ✅ Verification Checklist

- [ ] Docker is running (`docker ps`)
- [ ] PostgreSQL container is healthy
- [ ] `.env.local` is configured
- [ ] Tesseract is installed (`tesseract --version`)
- [ ] Backend API starts without errors (`make run-api`)
- [ ] You can access http://localhost:8080 (should show JSON error for no path)
- [ ] You can access pgAdmin at http://localhost:5050

**Once all checks pass, you're ready to develop locally! 🎉**
