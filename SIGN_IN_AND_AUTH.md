# Complete Sign-In & Authentication Guide

This guide provides all credentials, endpoints, and step-by-step instructions for authentication.

---

## 🔑 **Quick Credentials Summary**

### **Admin Panel**
```
Email:    admin@test.com
Password: AdminPassword123!
OTP/MFA:  (Generated during setup - see below)
```

### **Backend API**
```
Base URL:   http://localhost:8080
API Version: /api/v1
JWT Secret: your-local-dev-secret-change-in-production-12345678901234
```

### **pgAdmin Database UI**
```
URL:      http://localhost:5050
Email:    admin@example.com
Password: admin
```

### **PostgreSQL Database**
```
Host:     localhost:5432
User:     postgres
Password: postgres
Database: maid_recruitment
```

---

## 🚀 **Step 1: Seed Admin Account (First Time Only)**

Before you can sign in, you need to create an admin account in the local database.

### **Command to Seed Admin**

```powershell
cd c:\Users\NOOR AL MUSABAH\Documents\PROJECT_2
go run ./cmd/adminseed/main.go -email="admin@test.com" -password="AdminPassword123!" -name="Platform Super Admin" -role="super_admin"
```

### **Expected Output**

After running the command, you'll see JSON output like:

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

**Save this output!** You'll need the `mfa_secret` and `current_otp` for login.

---

## 🔐 **Step 2: Set Up MFA (Multi-Factor Authentication)**

The admin account requires MFA/OTP (One-Time Password) for security.

### **Option A: Using Google Authenticator (Recommended)**

1. **Download Google Authenticator:**
   - iOS: App Store
   - Android: Google Play Store

2. **Add Account:**
   - Open Google Authenticator
   - Tap "+" → "Scan QR code"
   - Use the `provisioning_url` from the seed output
   - Alternatively, manually enter the `mfa_secret`

3. **Get OTP Code:**
   - Google Authenticator will show a 6-digit code
   - This code refreshes every 30 seconds

### **Option B: Using Microsoft Authenticator**

1. Download and open Microsoft Authenticator
2. Add account
3. Scan QR code from `provisioning_url`

### **Option C: For Local Development (No App Needed)**

Use the `current_otp` from the seed output directly:

```json
"current_otp": "123456"
```

You can generate OTPs programmatically or use online generators with the `mfa_secret`.

---

## 🔑 **Step 3: Admin Login via API**

### **Admin Login Endpoint**

```
POST /api/v1/admin/login
```

### **Request Format**

```json
{
  "email": "admin@test.com",
  "password": "AdminPassword123!",
  "otp_code": "123456"
}
```

### **Using cURL**

```bash
curl -X POST http://localhost:8080/api/v1/admin/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@test.com",
    "password": "AdminPassword123!",
    "otp_code": "123456"
  }' \
  -c cookies.txt
```

### **Using PowerShell**

```powershell
$body = @{
    email = "admin@test.com"
    password = "AdminPassword123!"
    otp_code = "123456"
} | ConvertTo-Json

$response = Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/admin/login" `
  -Method Post `
  -Headers @{"Content-Type" = "application/json"} `
  -Body $body `
  -SessionVariable "adminSession"

$response | ConvertTo-Json
```

### **Response (200 OK)**

```json
{
  "admin": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "admin@test.com",
    "full_name": "Platform Super Admin",
    "role": "super_admin",
    "last_login": "2026-06-08T20:30:00Z",
    "force_password_change": false
  },
  "expires_at": "2026-06-08T21:30:00Z"
}
```

### **JWT Token in Response**

The token is returned as an HTTP-only session cookie (automatically managed):
- **Cookie Name:** `admin_session`
- **Duration:** 1 hour
- **Usage:** Include in subsequent requests automatically

---

## 👥 **Step 4: User Registration (Optional)**

If you want to create regular users (agencies, etc.), use the registration endpoint.

### **User Registration Endpoint**

```
POST /api/v1/auth/register
```

### **Request Format**

```json
{
  "email": "agency@example.com",
  "password": "SecurePassword123!",
  "full_name": "Ethiopian Agency",
  "role": "ethiopian_agent",
  "company_name": "ABC Recruitment Agency"
}
```

### **Using cURL**

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "agency@example.com",
    "password": "SecurePassword123!",
    "full_name": "Ethiopian Agency",
    "role": "ethiopian_agent",
    "company_name": "ABC Recruitment Agency"
  }' \
  -c agency-cookies.txt
```

### **Using PowerShell**

```powershell
$body = @{
    email = "agency@example.com"
    password = "SecurePassword123!"
    full_name = "Ethiopian Agency"
    role = "ethiopian_agent"
    company_name = "ABC Recruitment Agency"
} | ConvertTo-Json

$response = Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/auth/register" `
  -Method Post `
  -Headers @{"Content-Type" = "application/json"} `
  -Body $body

$response | ConvertTo-Json
```

### **Valid User Roles**

- `ethiopian_agent` - Ethiopian recruitment agency
- `foreign_agent` - Foreign country agent/partner

### **Response (201 Created)**

```json
{
  "message": "User registered successfully. Please verify your email.",
  "user": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "email": "agency@example.com",
    "full_name": "Ethiopian Agency",
    "role": "ethiopian_agent",
    "company_name": "ABC Recruitment Agency",
    "email_verified": false,
    "account_status": "pending_approval"
  }
}
```

---

## 🔑 **Step 5: User Login**

### **User Login Endpoint**

```
POST /api/v1/auth/login
```

### **Request Format**

```json
{
  "email": "agency@example.com",
  "password": "SecurePassword123!"
}
```

### **Using cURL**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "agency@example.com",
    "password": "SecurePassword123!"
  }' \
  -c user-cookies.txt
```

### **Using PowerShell**

```powershell
$body = @{
    email = "agency@example.com"
    password = "SecurePassword123!"
} | ConvertTo-Json

$response = Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/auth/login" `
  -Method Post `
  -Headers @{"Content-Type" = "application/json"} `
  -Body $body `
  -SessionVariable "userSession"

$response | ConvertTo-Json
```

### **Response (200 OK)**

```json
{
  "user": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "email": "agency@example.com",
    "full_name": "Ethiopian Agency",
    "role": "ethiopian_agent",
    "company_name": "ABC Recruitment Agency",
    "email_verified": false,
    "auto_share_candidates": false,
    "account_status": "pending_approval",
    "current_session_id": "session-uuid-here"
  }
}
```

---

## 🎫 **Using JWT Tokens**

### **Get JWT Token**

After login, the token is stored in the session cookie. For API-only usage, you can extract it:

```powershell
# Extract token from response headers/cookies
$headers = @{
    "Authorization" = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    "Content-Type" = "application/json"
}
```

### **Using JWT in Requests**

```bash
curl -X GET http://localhost:8080/api/v1/candidates \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### **JWT Token Format**

Header:
```json
{
  "alg": "HS256",
  "typ": "JWT"
}
```

Payload:
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440001",
  "email": "agency@example.com",
  "role": "ethiopian_agent",
  "exp": 1686321600,
  "iat": 1686318000
}
```

**Token Expiration:** 1 hour from login

---

## 👤 **Check Current User**

### **Endpoint**

```
GET /api/v1/auth/me
```

### **With cURL**

```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -b cookies.txt  # Uses saved session cookie
```

### **Response**

```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "email": "agency@example.com",
    "full_name": "Ethiopian Agency",
    "role": "ethiopian_agent",
    "email_verified": false,
    "auto_share_candidates": false
  }
}
```

---

## 🚪 **Logout**

### **Admin Logout**

```bash
curl -X POST http://localhost:8080/api/v1/admin/logout \
  -b cookies.txt
```

### **User Logout**

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -b cookies.txt
```

---

## 🔄 **Complete Login Flow Example**

Here's a complete example workflow in PowerShell:

```powershell
# 1. Set credentials
$adminEmail = "admin@test.com"
$adminPassword = "AdminPassword123!"
$otpCode = "123456"  # Get this from your authenticator app

# 2. Login
$loginBody = @{
    email = $adminEmail
    password = $adminPassword
    otp_code = $otpCode
} | ConvertTo-Json

$loginResponse = Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/admin/login" `
  -Method Post `
  -Headers @{"Content-Type" = "application/json"} `
  -Body $loginBody `
  -SessionVariable "adminSession"

Write-Host "Admin logged in successfully:"
$loginResponse.admin | Format-Table

# 3. Get admin info
$meResponse = Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/admin/me" `
  -WebSession $adminSession

Write-Host "Current admin:"
$meResponse.admin | Format-Table

# 4. Make authenticated requests
$candidates = Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/candidates" `
  -WebSession $adminSession

Write-Host "Candidates count: $($candidates.length)"

# 5. Logout
Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/admin/logout" `
  -Method Post `
  -WebSession $adminSession

Write-Host "Admin logged out"
```

---

## 📱 **MFA Troubleshooting**

### "OTP Code is Invalid"

**Problem:** Login fails with "invalid admin credentials" when OTP is correct

**Solutions:**
1. **Check time sync:** Ensure your system clock is correct
   - Run: `w32tm /resync` (Windows)
   - Or check: Settings → Time & Language → Date & time

2. **Use current code:** OTP codes refresh every 30 seconds
   - Wait a few seconds after getting code
   - Don't use expired codes

3. **Regenerate MFA:** Run adminseed again with new MFA
   ```powershell
   go run ./cmd/adminseed/main.go -email="admin@test.com" -password="AdminPassword123!"
   ```

### "MFA Secret not working"

**Problem:** Google Authenticator won't scan or doesn't work

**Solutions:**
1. Manually enter the secret
2. Use the `current_otp` from seed output directly (for development)
3. Regenerate seed with different app

---

## 🔒 **Password Requirements**

### **Admin Password**
- Minimum 12 characters
- Should include uppercase, lowercase, numbers, symbols
- Example: `SecurePass123!@`

### **User Password**
- Minimum 8 characters
- Example: `Password123!`

---

## 🌐 **Test Credentials for All Roles**

### **Admin (Super Admin)**
```
Email:    admin@test.com
Password: AdminPassword123!
Role:     super_admin
MFA:      Required (from Google Authenticator)
```

### **Ethiopian Agency Agent**
```
Email:    ethiopian-agency@test.com
Password: AgencyPass123!
Role:     ethiopian_agent
Company:  Ethiopian Recruitment Ltd
```

### **Foreign Country Agent**
```
Email:    foreign-agent@test.com
Password: ForeignPass123!
Role:     foreign_agent
Company:  International Staffing Solutions
```

**To create these test users, use the registration endpoint or adminseed for admin.**

---

## 📝 **API Endpoints Summary**

| Method | Endpoint | Auth Required | Purpose |
|--------|----------|----------------|---------|
| POST | `/api/v1/admin/login` | No | Admin login |
| GET | `/api/v1/admin/me` | Yes (Admin) | Get current admin |
| POST | `/api/v1/admin/logout` | Yes (Admin) | Admin logout |
| POST | `/api/v1/auth/register` | No | Register new user |
| POST | `/api/v1/auth/login` | No | User login |
| GET | `/api/v1/auth/me` | Yes (User) | Get current user |
| POST | `/api/v1/auth/logout` | Yes (User) | User logout |
| POST | `/api/v1/auth/verify-email` | No | Verify email token |
| POST | `/api/v1/auth/forgot-password` | No | Forgot password request |
| POST | `/api/v1/auth/reset-password` | No | Reset password with code |

---

## 🧪 **Quick Test Commands**

### **Test Admin Login**
```powershell
$response = Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/admin/login" `
  -Method Post `
  -ContentType "application/json" `
  -Body '{"email":"admin@test.com","password":"AdminPassword123!","otp_code":"123456"}'

if ($response.admin) {
    Write-Host "✓ Admin login successful"
} else {
    Write-Host "✗ Admin login failed"
}
```

### **Test User Registration**
```powershell
$response = Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/auth/register" `
  -Method Post `
  -ContentType "application/json" `
  -Body '{
    "email":"test@agency.com",
    "password":"TestPass123!",
    "full_name":"Test Agency",
    "role":"ethiopian_agent"
  }'

if ($response.user) {
    Write-Host "✓ User registration successful"
} else {
    Write-Host "✗ User registration failed"
}
```

---

## ✅ **Complete Setup Checklist**

- [ ] Backend API is running (`make run-api`)
- [ ] PostgreSQL is running (`docker-compose ps`)
- [ ] Run adminseed to create admin account
- [ ] Save the `mfa_secret` from seed output
- [ ] Set up Google Authenticator with MFA secret
- [ ] Test admin login with OTP code
- [ ] Test user registration
- [ ] Test user login
- [ ] Access authenticated endpoints

**Once all checks pass, authentication is ready! 🎉**

---

## 🆘 **Common Issues**

### "Email not found"
- Admin account doesn't exist
- **Solution:** Run adminseed first

### "Invalid password"
- Password is incorrect
- **Solution:** Check password spelling, case-sensitive

### "Account locked"
- Too many failed login attempts
- **Solution:** Wait 15 minutes or run adminseed to reset

### "Email verification required"
- User hasn't verified email yet
- **Solution:** Check email or use verify endpoint

### "Account pending approval"
- Admin hasn't approved user registration yet
- **Solution:** Contact admin or approve via admin panel

---

## 🔗 **Related Documentation**

- [API Endpoints](CREDENTIALS_AND_ENDPOINTS.md)
- [Local Development Setup](LOCAL_DEV_SETUP.md)
- [Database Setup](migrations/)
- [Frontend Integration](frontend/README.md)
