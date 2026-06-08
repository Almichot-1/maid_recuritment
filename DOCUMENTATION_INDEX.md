# 📚 Local Development Documentation Index

Welcome! This is your complete guide to running the Maid Recruitment Platform locally with all features working.

---

## 🚀 **START HERE: Quick Links**

### **For the Impatient (5 Minutes)**
→ **[QUICK_CREDENTIALS.md](QUICK_CREDENTIALS.md)** - Copy-paste all credentials

### **Step-by-Step Setup (20 Minutes)**
→ **[COMPLETE_SETUP_GUIDE.md](COMPLETE_SETUP_GUIDE.md)** - Full walkthrough from scratch

### **Authentication Details (Reference)**
→ **[SIGN_IN_AND_AUTH.md](SIGN_IN_AND_AUTH.md)** - Login flows and JWT tokens

### **API Endpoints & Testing (Reference)**
→ **[CREDENTIALS_AND_ENDPOINTS.md](CREDENTIALS_AND_ENDPOINTS.md)** - All APIs and examples

---

## 📋 **Documentation Overview**

### **Setup & Infrastructure**

| File | Purpose | Time |
|------|---------|------|
| **[COMPLETE_SETUP_GUIDE.md](COMPLETE_SETUP_GUIDE.md)** | 📖 Full step-by-step setup with all phases | 20 min |
| **[LOCAL_DEV_SETUP.md](LOCAL_DEV_SETUP.md)** | 🛠️ Local development configuration & troubleshooting | Reference |
| **[TESSERACT_INSTALLATION.md](TESSERACT_INSTALLATION.md)** | 📦 Install OCR for passport extraction | 10 min |
| **[docker-compose.yml](docker-compose.yml)** | 🐳 Docker configuration (PostgreSQL, pgAdmin) | Setup |
| **[.env.local](.env.local)** | ⚙️ Environment variables | Setup |

### **Authentication & Credentials**

| File | Purpose | Time |
|------|---------|------|
| **[QUICK_CREDENTIALS.md](QUICK_CREDENTIALS.md)** | 🔑 All credentials - copy & paste | 1 min |
| **[SIGN_IN_AND_AUTH.md](SIGN_IN_AND_AUTH.md)** | 🔐 Complete authentication guide (admin, users, JWT) | Reference |
| **Admin Account** | Email: `admin@test.com` | See QUICK_CREDENTIALS |
| **Admin Password** | `AdminPassword123!` | See QUICK_CREDENTIALS |
| **MFA Setup** | Google Authenticator | See SIGN_IN_AND_AUTH |

### **API Documentation**

| File | Purpose | Time |
|------|---------|------|
| **[CREDENTIALS_AND_ENDPOINTS.md](CREDENTIALS_AND_ENDPOINTS.md)** | 📡 All API endpoints with examples | Reference |
| **Passport OCR** | Extract data from passport images | See CREDENTIALS_AND_ENDPOINTS |
| **CV Generation** | Generate professional PDFs | See CREDENTIALS_AND_ENDPOINTS |
| **Database Queries** | Direct PostgreSQL access | See CREDENTIALS_AND_ENDPOINTS |

---

## ✅ **Quick Start Checklist**

### **Phase 1: Initial Setup (Already Done ✓)**
- [x] Docker containers running (PostgreSQL + pgAdmin)
- [x] Backend .env.local configured
- [x] Frontend .env.local configured
- [x] All documentation created

### **Phase 2: Admin Account (Do This Next)**
```powershell
cd c:\Users\NOOR AL MUSABAH\Documents\PROJECT_2
go run ./cmd/adminseed/main.go -email="admin@test.com" -password="AdminPassword123!" -name="Platform Super Admin" -role="super_admin"
```
- [ ] Admin seeded
- [ ] MFA secret saved
- [ ] Google Authenticator configured

### **Phase 3: Start Backend**
```powershell
cd c:\Users\NOOR AL MUSABAH\Documents\PROJECT_2
make run-api
```
- [ ] Backend running on http://localhost:8080

### **Phase 4: Admin Login**
Use credentials from QUICK_CREDENTIALS.md
- [ ] Admin logged in successfully
- [ ] Can make authenticated API requests

### **Phase 5: Optional - Start Frontend**
```powershell
cd c:\Users\NOOR AL MUSABAH\Documents\PROJECT_2\frontend
npm run dev
```
- [ ] Frontend running on http://localhost:3000

---

## 🔑 **All Credentials at a Glance**

```
╔═══════════════════════════════════════════════════════════╗
║              LOCAL DEVELOPMENT CREDENTIALS              ║
╠═══════════════════════════════════════════════════════════╣
║                                                           ║
║  🗄️  DATABASE (PostgreSQL)                              ║
║     Host: localhost:5432                                ║
║     User: postgres / Password: postgres                 ║
║     Database: maid_recruitment                          ║
║                                                           ║
║  🔐 ADMIN ACCOUNT                                        ║
║     Email: admin@test.com                               ║
║     Password: AdminPassword123!                         ║
║     MFA: Google Authenticator (setup required)          ║
║                                                           ║
║  🌐 BACKEND API                                          ║
║     URL: http://localhost:8080                          ║
║     API Base: http://localhost:8080/api/v1              ║
║     JWT Secret: your-local-dev-secret-...               ║
║                                                           ║
║  🎨 DATABASE UI (pgAdmin)                               ║
║     URL: http://localhost:5050                          ║
║     Email: admin@example.com / admin                    ║
║                                                           ║
║  🎯 FRONTEND                                             ║
║     URL: http://localhost:3000                          ║
║     (Run: cd frontend && npm run dev)                   ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
```

---

## 📖 **Reading Guide by Role**

### **Backend Developer**
1. Start: **[COMPLETE_SETUP_GUIDE.md](COMPLETE_SETUP_GUIDE.md)** (Phases 1-5)
2. Reference: **[CREDENTIALS_AND_ENDPOINTS.md](CREDENTIALS_AND_ENDPOINTS.md)** (API docs)
3. When stuck: **[LOCAL_DEV_SETUP.md](LOCAL_DEV_SETUP.md)** (Troubleshooting)

### **Frontend Developer**
1. Start: **[COMPLETE_SETUP_GUIDE.md](COMPLETE_SETUP_GUIDE.md)** (Phase 1-7)
2. Reference: **[QUICK_CREDENTIALS.md](QUICK_CREDENTIALS.md)** (API endpoints)
3. Test: **[SIGN_IN_AND_AUTH.md](SIGN_IN_AND_AUTH.md)** (Login flows)

### **DevOps/Infrastructure**
1. Reference: **[LOCAL_DEV_SETUP.md](LOCAL_DEV_SETUP.md)** (Docker setup)
2. Check: **[docker-compose.yml](docker-compose.yml)** (Configuration)
3. Troubleshoot: **[TESSERACT_INSTALLATION.md](TESSERACT_INSTALLATION.md)** (OCR setup)

### **QA/Tester**
1. Start: **[COMPLETE_SETUP_GUIDE.md](COMPLETE_SETUP_GUIDE.md)** (Full setup)
2. Test: **[CREDENTIALS_AND_ENDPOINTS.md](CREDENTIALS_AND_ENDPOINTS.md)** (API examples)
3. Scenarios: **[SIGN_IN_AND_AUTH.md](SIGN_IN_AND_AUTH.md)** (Login tests)

---

## 🎯 **Common Tasks**

### **"I want to login as admin"**
→ See **[QUICK_CREDENTIALS.md](QUICK_CREDENTIALS.md)** → Admin Account section

### **"How do I generate a CV?"**
→ See **[CREDENTIALS_AND_ENDPOINTS.md](CREDENTIALS_AND_ENDPOINTS.md)** → CV Generation Endpoints

### **"I need to extract passport data"**
→ See **[TESSERACT_INSTALLATION.md](TESSERACT_INSTALLATION.md)** then **[CREDENTIALS_AND_ENDPOINTS.md](CREDENTIALS_AND_ENDPOINTS.md)** → Passport OCR Endpoints

### **"How do I register a test user?"**
→ See **[SIGN_IN_AND_AUTH.md](SIGN_IN_AND_AUTH.md)** → Step 4: User Registration

### **"API is not responding"**
→ See **[LOCAL_DEV_SETUP.md](LOCAL_DEV_SETUP.md)** → Troubleshooting section

### **"MFA code doesn't work"**
→ See **[SIGN_IN_AND_AUTH.md](SIGN_IN_AND_AUTH.md)** → MFA Troubleshooting

---

## 🚀 **Next Steps**

### **Immediate (Right Now)**
```powershell
# 1. Seed admin account
go run ./cmd/adminseed/main.go -email="admin@test.com" -password="AdminPassword123!" -name="Platform Super Admin" -role="super_admin"

# 2. Save the output (especially mfa_secret)
# 3. Setup Google Authenticator with the mfa_secret
# 4. Start backend
make run-api
```

### **After Setup**
- [ ] Test admin login
- [ ] Create test candidates
- [ ] Register test users
- [ ] Generate test CVs
- [ ] Extract passport data

### **Optional**
- [ ] Start frontend (`cd frontend && npm run dev`)
- [ ] Install Tesseract for OCR
- [ ] Setup more test accounts

---

## 🆘 **Need Help?**

| Issue | Document |
|-------|----------|
| How do I set everything up? | **[COMPLETE_SETUP_GUIDE.md](COMPLETE_SETUP_GUIDE.md)** |
| I forgot a credential | **[QUICK_CREDENTIALS.md](QUICK_CREDENTIALS.md)** |
| How do I login? | **[SIGN_IN_AND_AUTH.md](SIGN_IN_AND_AUTH.md)** |
| How do I call the API? | **[CREDENTIALS_AND_ENDPOINTS.md](CREDENTIALS_AND_ENDPOINTS.md)** |
| Something's broken | **[LOCAL_DEV_SETUP.md](LOCAL_DEV_SETUP.md)** Troubleshooting |
| Tesseract won't install | **[TESSERACT_INSTALLATION.md](TESSERACT_INSTALLATION.md)** |

---

## 📊 **System Status**

### **Currently Running**
- ✅ PostgreSQL (Docker) - localhost:5432
- ✅ pgAdmin (Docker) - http://localhost:5050
- ✅ Git development branch - `development`
- ⏳ Backend API - Ready to start with `make run-api`
- ⏳ Frontend - Ready to start with `npm run dev`

### **Configuration**
- ✅ Docker Compose configured
- ✅ Environment files ready (.env.local)
- ✅ Database migrations available
- ⏳ Admin account - Ready to seed
- ⏳ Tesseract OCR - Optional but recommended

---

## 🎓 **Learning Resources**

### **Understand the Architecture**
→ See [README.md](README.md) for system overview

### **Understand the Database**
→ Check [migrations/](migrations/) folder for schema

### **Understand the Code**
→ Explore [internal/](internal/) for Go code structure
→ Explore [frontend/](frontend/) for React code structure

---

## ✨ **Key Features Ready for Development**

- ✅ **Authentication** - Admin & user accounts with JWT
- ✅ **Multi-Factor Authentication** - MFA/OTP security
- ✅ **Candidate Management** - CRUD operations
- ✅ **Document Upload** - File storage & retrieval
- ✅ **Passport OCR** - Extract data from passport images
- ✅ **CV Generation** - Create professional PDFs
- ✅ **Database** - Full PostgreSQL setup
- ✅ **API** - Complete RESTful endpoints
- ✅ **Frontend** - React with TypeScript
- ✅ **Expiry Scheduler** - Background jobs

---

## 🔒 **Important Reminders**

⚠️ **These are LOCAL development credentials ONLY:**
- Never use these passwords in production
- Never commit .env.local to version control
- All data is isolated to your local Docker containers

✅ **Production users are completely safe:**
- Your local database is 100% separate
- Production Supabase is untouched
- No risk whatsoever to real users

---

## 📞 **Support**

If you get stuck:
1. Check the **relevant documentation file** (see "Need Help?" table)
2. Read the **Troubleshooting section** in that file
3. Verify **System Status** above
4. Re-read **Complete Setup Guide** Phase by Phase

---

**Ready to start? Begin with [COMPLETE_SETUP_GUIDE.md](COMPLETE_SETUP_GUIDE.md)! 🚀**

---

## 📝 **Document Manifest**

- **COMPLETE_SETUP_GUIDE.md** (11KB) - Complete step-by-step setup
- **QUICK_CREDENTIALS.md** (6KB) - Quick copy-paste credentials
- **SIGN_IN_AND_AUTH.md** (15KB) - Authentication detailed guide
- **CREDENTIALS_AND_ENDPOINTS.md** (13KB) - API documentation
- **TESSERACT_INSTALLATION.md** (6KB) - OCR installation
- **LOCAL_DEV_SETUP.md** (5KB) - Local development config
- **docker-compose.yml** - Docker configuration
- **.env.local** - Environment variables
- **.gitignore** - Git ignore rules

**Total Documentation: ~80KB of comprehensive guides**

**Last Updated: 2026-06-08**
**Status: ✅ Ready for Development**
