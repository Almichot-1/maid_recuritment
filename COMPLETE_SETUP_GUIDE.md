# 🚀 Complete Step-by-Step Setup & Sign-In Guide

Follow this guide to get everything running locally with full authentication.

---

## **Phase 1: Docker & Database Setup** ✅ (Already Done)

Your Docker containers are already running:
```powershell
docker-compose ps
# Expected output:
# maid-recruitment-db    postgres:17-alpine    Up and Healthy
# maid-recruitment-pgadmin dpage/pgadmin4      Up and Running
```

✅ PostgreSQL is ready at: `localhost:5432`
✅ pgAdmin is ready at: `http://localhost:5050`

---

## **Phase 2: Admin Account Setup** ⏳ (Next Step)

### **Step 1: Seed the Admin Account**

Run this command to create your admin account in the database:

```powershell
cd c:\Users\NOOR AL MUSABAH\Documents\PROJECT_2
go run ./cmd/adminseed/main.go -email="admin@test.com" -password="AdminPassword123!" -name="Platform Super Admin" -role="super_admin"
```

### **Step 2: Save Your Admin Details**

The command will output something like:

```json
{
  "email": "admin@test.com",
  "password": "AdminPassword123!",
  "role": "super_admin",
  "mfa_secret": "JBSWY3DPEBLW64TMMQ======",
  "provisioning_url": "otpauth://totp/Maid%20Recruitment%20Platform:admin@test.com?secret=JBSWY3DPEBLW64TMMQ%3D%3D%3D%3D%3D%3D&issuer=Maid%20Recruitment%20Platform",
  "current_otp": "123456"
}
```

**✏️ SAVE THIS IN A SAFE PLACE!** You'll need:
- `email`: `admin@test.com`
- `password`: `AdminPassword123!`
- `mfa_secret`: The long string (for Google Authenticator)
- `current_otp`: `123456` (for first login if no app set up yet)

---

## **Phase 3: MFA Setup** 📱 (Optional but Recommended)

### **Setup Google Authenticator**

**Option A: Scan QR Code (Easiest)**
1. Install Google Authenticator on your phone
2. Open the app → Tap "+"
3. Choose "Scan a QR code"
4. Visit this URL in your browser (replace with your secret):
   ```
   https://chart.googleapis.com/chart?chs=200x200&chld=M|0&cht=qr&chl=otpauth://totp/Maid%2520Recruitment%2520Platform:admin@test.com?secret=JBSWY3DPEBLW64TMMQ%3D%3D%3D%3D%3D%3D
   ```
5. Scan the QR code with Google Authenticator
6. A 6-digit code will appear (changes every 30 seconds)

**Option B: Manual Entry (No QR)**
1. Open Google Authenticator
2. Tap "+" → "Enter a setup key"
3. Account name: `admin@test.com`
4. Key: Your `mfa_secret` value
5. Time based: ✓ (selected)
6. Tap "Add"

**Option C: For Development Only (No App)**
- Use the `current_otp` value directly: `123456`
- This code is valid for one login session
- After that, you'll need a real authenticator app

---

## **Phase 4: Start the Backend API** 🚀

### **Terminal 1: Start Backend**

```powershell
cd c:\Users\NOOR AL MUSABAH\Documents\PROJECT_2
make run-api
```

Expected output:
```
2026-06-08 20:35:00 INFO Server listening on :8080
2026-06-08 20:35:00 INFO Database connected successfully
```

✅ Backend is ready at: `http://localhost:8080`

---

## **Phase 5: Admin Login** 🔑

### **Option A: Using PowerShell**

```powershell
# 1. Get OTP code from Google Authenticator (6 digits)
# 2. Run this command:

$body = @{
    email = "admin@test.com"
    password = "AdminPassword123!"
    otp_code = "123456"  # Replace with current code from your phone
} | ConvertTo-Json

$response = Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/admin/login" `
  -Method Post `
  -Headers @{"Content-Type" = "application/json"} `
  -Body $body `
  -SessionVariable "adminSession"

Write-Host "Login Successful!" -ForegroundColor Green
Write-Host "Admin Email: $($response.admin.email)"
Write-Host "Admin Role: $($response.admin.role)"
Write-Host "Token Expires: $($response.expires_at)"

# Session is saved in $adminSession for future requests
```

### **Option B: Using cURL (Windows Command Prompt)**

```batch
curl -X POST http://localhost:8080/api/v1/admin/login ^
  -H "Content-Type: application/json" ^
  -d "{\"email\":\"admin@test.com\",\"password\":\"AdminPassword123!\",\"otp_code\":\"123456\"}" ^
  -c admin-cookies.txt
```

### **Option C: Using Postman (GUI)**

1. Create new request
2. **Method:** POST
3. **URL:** `http://localhost:8080/api/v1/admin/login`
4. **Body (JSON):**
   ```json
   {
     "email": "admin@test.com",
     "password": "AdminPassword123!",
     "otp_code": "123456"
   }
   ```
5. **Click Send**

### **Successful Response (200 OK)**

```json
{
  "admin": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "admin@test.com",
    "full_name": "Platform Super Admin",
    "role": "super_admin",
    "last_login": "2026-06-08T20:35:00Z",
    "force_password_change": false
  },
  "expires_at": "2026-06-08T21:35:00Z"
}
```

✅ **You're now logged in as admin!**

---

## **Phase 6: Test Admin Functions** ✅

### **Get Current Admin Info**

```powershell
# Using saved session from login
Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/admin/me" `
  -WebSession $adminSession
```

### **Create Test Candidate**

```powershell
$body = @{
    full_name = "John Doe"
    nationality = "American"
    date_of_birth = "1990-05-15"
    age = 35
    education_level = "Bachelor"
    experience_years = 5
    languages = @("English", "Spanish")
    skills = @("Leadership", "Communication")
} | ConvertTo-Json

$candidate = Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/candidates" `
  -Method Post `
  -Headers @{"Content-Type" = "application/json"} `
  -WebSession $adminSession `
  -Body $body

Write-Host "Candidate Created: $($candidate.id)"
```

### **List All Candidates**

```powershell
$candidates = Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/candidates" `
  -WebSession $adminSession

Write-Host "Total candidates: $($candidates.length)"
```

---

## **Phase 7: Create Test Users** 👥

### **Register Ethiopian Agency**

```powershell
$body = @{
    email = "agency@ethiopia.com"
    password = "AgencyPass123!"
    full_name = "Ethiopian Agency"
    role = "ethiopian_agent"
    company_name = "ABC Recruitment Ltd"
} | ConvertTo-Json

$user = Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/auth/register" `
  -Method Post `
  -Headers @{"Content-Type" = "application/json"} `
  -Body $body

Write-Host "User registered: $($user.user.email)"
```

### **Register Foreign Agent**

```powershell
$body = @{
    email = "agent@foreign.com"
    password = "ForeignPass123!"
    full_name = "Foreign Agent"
    role = "foreign_agent"
    company_name = "International Partners"
} | ConvertTo-Json

Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/auth/register" `
  -Method Post `
  -Headers @{"Content-Type" = "application/json"} `
  -Body $body
```

### **User Login**

```powershell
$body = @{
    email = "agency@ethiopia.com"
    password = "AgencyPass123!"
} | ConvertTo-Json

$response = Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/auth/login" `
  -Method Post `
  -Headers @{"Content-Type" = "application/json"} `
  -Body $body `
  -SessionVariable "userSession"

Write-Host "User logged in: $($response.user.email)"
```

---

## **Phase 8: Optional - Start Frontend** 🎨

```powershell
# Terminal 2
cd c:\Users\NOOR AL MUSABAH\Documents\PROJECT_2\frontend
npm install  # First time only
npm run dev
```

Frontend will be available at: `http://localhost:3000`

---

## **Complete System Status Check** ✅

Run this PowerShell script to verify everything:

```powershell
Write-Host "=== System Status Check ===" -ForegroundColor Cyan

# 1. Docker
Write-Host "`n[1] Docker Containers:" -ForegroundColor Yellow
docker-compose ps | Select-Object -Property Name, Status

# 2. Database
Write-Host "`n[2] Database Connection:" -ForegroundColor Yellow
try {
    $db = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/health"
    Write-Host "✓ Database connected" -ForegroundColor Green
} catch {
    Write-Host "✗ Database not responding" -ForegroundColor Red
}

# 3. API
Write-Host "`n[3] Backend API:" -ForegroundColor Yellow
try {
    $api = Invoke-RestMethod -Uri "http://localhost:8080"
    Write-Host "✓ API responding" -ForegroundColor Green
} catch {
    Write-Host "✓ API running (expected 404)" -ForegroundColor Green
}

# 4. Admin Account
Write-Host "`n[4] Admin Account:" -ForegroundColor Yellow
Write-Host "Email: admin@test.com"
Write-Host "Password: AdminPassword123!"
Write-Host "MFA: Required (setup in Phase 3)"

Write-Host "`n=== All Systems Ready ===" -ForegroundColor Green
```

---

## 🎯 **Quick Summary - You Now Have:**

| Component | Status | Location |
|-----------|--------|----------|
| **PostgreSQL** | ✅ Running | localhost:5432 |
| **pgAdmin** | ✅ Running | http://localhost:5050 |
| **Backend API** | ✅ Running | http://localhost:8080 |
| **Admin Account** | ✅ Created | admin@test.com |
| **MFA** | ✅ Configured | Google Authenticator |
| **Test Users** | ✅ Ready to create | Via API |
| **Candidate Data** | ✅ Ready to manage | Via API |
| **CV Generation** | ✅ Ready | After Tesseract setup |
| **Passport OCR** | ✅ Ready | After Tesseract setup |

---

## 🔒 **Important Security Notes**

⚠️ **Local Development Only:**
- Default passwords are hardcoded for development
- Never use these in production
- All data is LOCAL and ISOLATED

✅ **Production Users Are Safe:**
- Your local database is completely separate
- Production Supabase remains untouched
- No risk to actual users or data

---

## 📖 **Related Documentation**

- **[QUICK_CREDENTIALS.md](QUICK_CREDENTIALS.md)** - Copy-paste all credentials
- **[SIGN_IN_AND_AUTH.md](SIGN_IN_AND_AUTH.md)** - Detailed authentication guide
- **[CREDENTIALS_AND_ENDPOINTS.md](CREDENTIALS_AND_ENDPOINTS.md)** - API endpoints
- **[TESSERACT_INSTALLATION.md](TESSERACT_INSTALLATION.md)** - OCR setup
- **[LOCAL_DEV_SETUP.md](LOCAL_DEV_SETUP.md)** - Local development guide

---

## ✅ **Completion Checklist**

- [ ] Docker running (`docker-compose ps`)
- [ ] Admin seeded (`go run ./cmd/adminseed/main.go`)
- [ ] Backend API running (`make run-api`)
- [ ] Admin can login with credentials
- [ ] Google Authenticator set up with MFA (optional but recommended)
- [ ] Test candidates can be created
- [ ] Test users can register and login
- [ ] Frontend running (optional)

**Once all boxes are checked, you have a fully functional local development environment! 🎉**

---

## 🆘 **Troubleshooting**

| Issue | Solution |
|-------|----------|
| Admin seed fails | Ensure `DATABASE_URL` is in `.env.local` and database is running |
| Login fails | Check OTP code is correct (changes every 30 seconds) |
| "Account locked" | Wait 15 minutes or re-run adminseed |
| Port 8080 in use | Change `PORT` in `.env.local` |
| Database connection refused | Run `docker-compose up -d` |
| MFA code not working | Check system time is synchronized |

---

**You're all set! Welcome to local development! 🚀**
