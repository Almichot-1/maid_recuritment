# 🔑 Quick Credentials Reference Card

## Copy-Paste All Credentials Here

### 🗄️ **Database Credentials**
```
Host:               localhost
Port:               5432
Username:           postgres
Password:           postgres
Database:           maid_recruitment
Connection String:  postgres://postgres:postgres@localhost:5432/maid_recruitment?sslmode=disable
```

### 🔐 **Admin Account (Local Dev)**
```
Email:              admin@test.com
Password:           AdminPassword123!
Role:               super_admin
MFA Required:       YES (see MFA Setup below)
```

### 🌐 **Backend API**
```
Base URL:           http://localhost:8080
API Base:           http://localhost:8080/api/v1
JWT Secret:         your-local-dev-secret-change-in-production-12345678901234
Auth Type:          JWT + Session Cookies
```

### 🎨 **Frontend**
```
URL:                http://localhost:3000
API URL:            http://localhost:8080/api/v1
Environment:        development
```

### 📊 **pgAdmin (Database UI)**
```
URL:                http://localhost:5050
Email:              admin@example.com
Password:           admin
```

### 📱 **MFA Setup (Google Authenticator)**
```
Method:             Time-based OTP (TOTP)
App:                Google Authenticator / Microsoft Authenticator
Time Step:          30 seconds
Digits:             6
```

**To Get MFA Secret:**
```bash
go run ./cmd/adminseed/main.go -email="admin@test.com" -password="AdminPassword123!" -name="Platform Super Admin" -role="super_admin"
```

Copy the `provisioning_url` or `mfa_secret` from output and add to your authenticator app.

---

## 🚀 **Startup Commands (Run in Order)**

### **Terminal 1: Start Docker**
```powershell
cd c:\Users\NOOR AL MUSABAH\Documents\PROJECT_2
docker-compose up -d
```

### **Terminal 2: Start Backend API**
```powershell
cd c:\Users\NOOR AL MUSABAH\Documents\PROJECT_2
make run-api
# OR: go run ./cmd/api
```

### **Terminal 3: Start Frontend (Optional)**
```powershell
cd c:\Users\NOOR AL MUSABAH\Documents\PROJECT_2\frontend
npm run dev
```

---

## 🔑 **Admin Login Command (PowerShell)**

```powershell
$otpCode = "123456"  # Get this from Google Authenticator

$body = @{
    email = "admin@test.com"
    password = "AdminPassword123!"
    otp_code = $otpCode
} | ConvertTo-Json

$response = Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/admin/login" `
  -Method Post `
  -Headers @{"Content-Type" = "application/json"} `
  -Body $body `
  -SessionVariable "session"

# Save for later requests
$session  # Contains authentication cookies
```

---

## 👥 **Create Test Users**

### **Registration Command**
```powershell
$body = @{
    email = "agency@test.com"
    password = "Password123!"
    full_name = "Test Agency"
    role = "ethiopian_agent"
    company_name = "Test Company"
} | ConvertTo-Json

Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/auth/register" `
  -Method Post `
  -Headers @{"Content-Type" = "application/json"} `
  -Body $body
```

---

## 🧪 **Test API Endpoints**

### **Passport OCR Preview**
```powershell
# Get JWT token first, then:
Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/candidates/passport/parse-preview" `
  -Method Post `
  -Headers @{"Authorization" = "Bearer <jwt_token>"} `
  -Form @{file = Get-Item "C:\path\to\passport.jpg"}
```

### **Generate CV**
```powershell
$body = '{"generated_by":"admin@test.com"}' | ConvertTo-Json

Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/candidates/{candidateId}/cv/generate" `
  -Method Post `
  -Headers @{
    "Authorization" = "Bearer <jwt_token>"
    "Content-Type" = "application/json"
  } `
  -Body $body
```

---

## 📋 **Service Status Checks**

### **Is Docker Running?**
```powershell
docker ps
# Should show: maid-recruitment-db, maid-recruitment-pgadmin
```

### **Is API Responding?**
```powershell
Invoke-RestMethod http://localhost:8080/api/v1/health
```

### **Is Frontend Running?**
```powershell
Invoke-RestMethod http://localhost:3000
```

### **Database Health**
```powershell
docker exec maid-recruitment-db pg_isready -U postgres
# Response: accepting connections
```

---

## 🔗 **Important URLs**

| Service | URL |
|---------|-----|
| Backend API | http://localhost:8080 |
| Frontend | http://localhost:3000 |
| pgAdmin | http://localhost:5050 |
| Admin Login | POST http://localhost:8080/api/v1/admin/login |
| User Registration | POST http://localhost:8080/api/v1/auth/register |
| User Login | POST http://localhost:8080/api/v1/auth/login |

---

## 📄 **Environment Variables (.env.local)**

```
PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/maid_recruitment?sslmode=disable
JWT_SECRET=your-local-dev-secret-change-in-production-12345678901234
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
RUN_EXPIRY_SCHEDULER=true
ENVIRONMENT=local
TESSERACT_PATH=
OCR_LANGUAGE=eng
```

---

## 🛑 **Stop Everything**

```powershell
# Stop Docker containers
docker-compose down

# Kill backend (Ctrl+C in terminal)
# Kill frontend (Ctrl+C in terminal)
```

---

## ⚡ **Quick Troubleshooting**

| Problem | Solution |
|---------|----------|
| Can't login | Check password and OTP from authenticator |
| Database not running | Run: `docker-compose up -d` |
| API not responding | Check: `make run-api` is running |
| OTP invalid | Check system time sync |
| Port already in use | Kill process or change port in .env.local |

---

## ✅ **Ready to Start!**

1. ✓ Docker running
2. ✓ API running (`make run-api`)
3. ✓ Admin account seeded
4. ✓ MFA configured
5. ✓ Ready to login!

**Get OTP code from authenticator app and login! 🎉**
