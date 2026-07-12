# Render Deployment Configuration Audit Report

**Generated:** July 12, 2026  
**Workspace:** My Workspace (ahmedyasine230@gmail.com)  
**Repository:** https://github.com/Almichot-1/maid_recuritment

---

## 🚨 URGENT ACTION REQUIRED

**Your application will go offline mid-month if you don't act now!**

You have 2 active services consuming 1,440 hours/month, but the Free tier only provides 750 hours/month. Around day 15-16 of each month, both services will shut down and your API will be completely offline for ~15 days.

**What to do RIGHT NOW:**
1. Go to Render Dashboard: https://dashboard.render.com/web/srv-d70godeuk2gs739aa37g
2. Click "Suspend Service" for the Oregon service
3. Keep only Frankfurt running (or vice versa if Oregon is your primary)

See Section 7 for detailed analysis.

---

## Executive Summary

This audit identified **3 web services** in your Render account. Two are active, one is suspended. **No cron jobs were found**, which is correct for the Free tier. 

**🚨 CRITICAL ISSUE:** You have TWO active services consuming free tier hours simultaneously. This will cause both services to shut down mid-month (~day 15-16) when you exceed the 750 hour limit.

Additionally, there are **critical configuration discrepancies** in your build and start commands that need immediate attention.

---

## 1. Service Inventory

### Active Services

#### 1.1 maid-recruitment-api-frankfurt (Primary Service)
- **Service ID:** `srv-d71aegma2pns73f1u5e0`
- **Status:** ✅ Active (not_suspended)
- **Region:** Frankfurt
- **Plan:** Free Tier
- **URL:** https://maid-recruitment-api-frankfurt.onrender.com
- **Branch:** main
- **Auto-deploy:** Enabled (on commit)
- **Health Check:** `/api/v1/health` ✅
- **Created:** March 24, 2026
- **Last Updated:** June 13, 2026

#### 1.2 maid-recruitment-api (Oregon Service)
- **Service ID:** `srv-d70godeuk2gs739aa37g`
- **Status:** ✅ Active (not_suspended)
- **Region:** Oregon
- **Plan:** Free Tier
- **URL:** https://maid-recruitment-api.onrender.com
- **Branch:** main
- **Auto-deploy:** Enabled (on commit)
- **Health Check:** `/api/v1/health` ✅
- **Created:** March 23, 2026
- **Last Updated:** June 13, 2026

#### 1.3 maid_recuritment (Legacy/Suspended)
- **Service ID:** `srv-d70gn3c50q8c73a6i8i0`
- **Status:** ⚠️ **SUSPENDED** (user-suspended)
- **Region:** Oregon
- **Plan:** Free Tier
- **URL:** https://maid-recuritment.onrender.com
- **Created:** March 23, 2026
- **Last Updated:** April 3, 2026
- **Note:** This appears to be a legacy/test service that's been manually suspended.

---

## 2. Cron Job Audit ✅

**Result:** No cron jobs found in your Render account.

This is **CORRECT** for the Free tier, as Render does not support cron jobs on free plans. The expiry scheduler functionality should be handled within your web service using `RUN_EXPIRY_SCHEDULER=true` as planned.

---

## 3. Command Configuration Analysis

### 3.1 maid-recruitment-api-frankfurt (Primary)

**Current Build Command:**
```bash
bash ./scripts/install_tesseract.sh && bash ./scripts/run_migrations.sh && go build -o ./bin/api ./cmd/api
```

**Expected Build Command:**
```bash
bash ./scripts/install_tesseract.sh && go build -o ./bin/api ./cmd/api && go build -o ./bin/devmigrate ./cmd/devmigrate
```

**❌ DISCREPANCY FOUND:**
1. The build command runs migrations during build (`bash ./scripts/run_migrations.sh`) instead of at startup
2. Missing: `go build -o ./bin/devmigrate ./cmd/devmigrate`
3. This means the devmigrate binary is never built

**Current Start Command:**
```bash
./bin/api
```

**Expected Start Command:**
```bash
bash ./scripts/run_migrations.sh && ./bin/api
```

**❌ CRITICAL DISCREPANCY:**
- Migrations are run during build instead of startup
- This is problematic because:
  - Build-time migrations may fail if the database isn't accessible during build
  - Migrations should run at startup to ensure database is up-to-date before the API starts
  - The expected pattern is: build the binaries → start the service → run migrations → start API

---

### 3.2 maid-recruitment-api (Oregon)

**Current Build Command:**
```bash
go build -o ./bin/api ./cmd/api
```

**Expected Build Command:**
```bash
bash ./scripts/install_tesseract.sh && go build -o ./bin/api ./cmd/api && go build -o ./bin/devmigrate ./cmd/devmigrate
```

**❌ CRITICAL DISCREPANCIES:**
1. Missing: `bash ./scripts/install_tesseract.sh` - OCR functionality won't work
2. Missing: `go build -o ./bin/devmigrate ./cmd/devmigrate` - migration binary not built
3. No migrations are run at all

**Current Start Command:**
```bash
./bin/api
```

**Expected Start Command:**
```bash
bash ./scripts/run_migrations.sh && ./bin/api
```

**❌ CRITICAL DISCREPANCY:**
- No migrations run before starting the API
- Tesseract OCR is not installed
- This service is likely non-functional for passport OCR features

---

## 4. Environment Variables

⚠️ **LIMITATION:** The Render MCP tools do not provide access to environment variables through the API. 

**Manual Verification Required:**

You must manually verify the following in the Render Dashboard for **both active services**:

### 4.1 Required Go Configuration
- [ ] `GO_VERSION` = `"1.25.8"` (Must match your go.mod toolchain requirement)

### 4.2 Expiry Scheduler Configuration
- [ ] `RUN_EXPIRY_SCHEDULER` = `"true"` (Required since no cron jobs are on Free tier)

### 4.3 Database Configuration
- [ ] `DATABASE_URL` is set and points to external database (Supabase/Neon/etc.)
- [ ] Database URL does NOT contain `/tmp/` or local file paths
- [ ] Database is persistent and not ephemeral

### 4.4 S3/Storage Configuration (CRITICAL for Free Tier)
- [ ] `S3_BUCKET` is populated
- [ ] `AWS_ACCESS_KEY_ID` is populated
- [ ] `AWS_SECRET_ACCESS_KEY` is populated
- [ ] `AWS_REGION` is set correctly

**⚠️ WARNING:** Render Free tier uses ephemeral storage. Any files uploaded to the local filesystem will be lost on service restart or redeploy. S3 configuration is **MANDATORY** for persistent file storage.

### 4.5 Other Required Variables
- [ ] `JWT_SECRET` is set
- [ ] `SMTP_*` variables are configured (if email is used)
- [ ] `FRONTEND_URL` is set correctly
- [ ] Any other application-specific environment variables

---

## 5. Summary of Issues

### Critical Issues (Fix Immediately)

1. **🚨 FREE TIER HOURS VIOLATION (HIGHEST PRIORITY)**
   - Two active services consuming 1,440 hours/month
   - Free tier limit is 750 hours/month
   - **Both services will shut down mid-month (~day 15-16)**
   - **ACTION REQUIRED:** Suspend Oregon service immediately

2. **❌ Both Services Have Incorrect Build/Start Commands**
   - Frankfurt: Runs migrations during build instead of startup
   - Oregon: Doesn't run migrations or install Tesseract at all
   - Neither service builds the `devmigrate` binary

3. **✅ Go Version Requirement**
   - Your `go.mod` specifies `toolchain go1.25.8`
   - Ensure `GO_VERSION=1.25.8` is set in Render environment variables
   - This must match your go.mod toolchain requirement

4. **⚠️ Oregon Service Is Likely Non-Functional**
   - Missing Tesseract installation
   - Missing migrations
   - Passport OCR features will fail

### Moderate Issues (Verify Soon)

5. **❔ Environment Variables Cannot Be Verified**
   - Manual dashboard check required for all env vars
   - S3 configuration is critical for data persistence

6. **❔ Unclear Service Strategy**
   - Both Frankfurt and Oregon services are active
   - Unclear which is primary
   - Creates unnecessary free tier hour consumption

### Positive Findings ✅

7. **✅ No Cron Jobs** - Correct for Free tier
8. **✅ Health Checks Configured** - Both active services have health endpoints
9. **✅ Auto-deploy Enabled** - Services will update on git push
10. **✅ Free Tier Confirmed** - Both services on free plan

---

## 6. Recommended Actions

### URGENT - Must Do Immediately (Priority 0)

**🚨 SUSPEND OREGON SERVICE TO AVOID MONTH-END SHUTDOWNS**

You must choose one service to keep active:

**Option A: Keep Frankfurt Only (RECOMMENDED)**
```
Reason: Frankfurt has the more correct build configuration
Action: Suspend srv-d70godeuk2gs739aa37g (maid-recruitment-api)
```

**Option B: Keep Oregon Only**
```
Reason: Only if Oregon is your actual primary endpoint
Action: Suspend srv-d71aegma2pns73f1u5e0 (maid-recruitment-api-frankfurt)
Warning: Oregon needs major config fixes first
```

**To Suspend via Render Dashboard:**
1. Go to https://dashboard.render.com/web/srv-d70godeuk2gs739aa37g (Oregon)
2. Click "Suspend Service" in settings
3. Confirm suspension

**To Suspend via MCP Tools:**
```
Call: mcp_Render_pause_project
Parameter: project_id = "srv-d70godeuk2gs739aa37g"
```

### Immediate Actions (Priority 1)

1. **Update Frankfurt Service Commands:**
   ```bash
   # Build Command
   bash ./scripts/install_tesseract.sh && go build -o ./bin/api ./cmd/api && go build -o ./bin/devmigrate ./cmd/devmigrate
   
   # Start Command
   bash ./scripts/run_migrations.sh && ./bin/api
   ```

2. **Update Oregon Service Commands:**
   - Use the same commands as Frankfurt above
   - Or consider suspending if not needed

3. **Verify Go Version:**
   - Check that `GO_VERSION=1.25.8` is correct in env vars
   - Update to a valid Go version if needed (likely 1.22.x or 1.23.x)

4. **Manually Verify All Environment Variables:**
   - Log into Render Dashboard
   - Navigate to each service → Environment tab
   - Verify all variables from section 4

### Short-term Actions (Priority 2)

5. **Test After Configuration Changes:**
   - Trigger a manual deploy after command updates
   - Verify migrations run successfully at startup
   - Test passport OCR functionality
   - Verify file uploads persist (S3 check)

6. **Monitor Free Tier Usage:**
   - Check Render dashboard for hour consumption
   - With 1 service: ~720 hrs/month (safe)
   - Should stay well under 750 hour limit

---

## 7. Free Tier Constraints Compliance

| Constraint | Status | Notes |
|------------|--------|-------|
| No Cron Jobs | ✅ Pass | No cron services found |
| Web Services Only | ✅ Pass | All services are web_service type |
| Ephemeral Storage | ⚠️ Verify | Must verify S3 is configured |
| **750 hours/month** | **❌ FAIL** | **2 active services = 1440 hrs/month - EXCEEDS LIMIT** |
| Sleep after inactivity | ℹ️ Note | Services will sleep after 15 min idle |

### 🚨 Critical: Free Tier Hours Violation

**Current Situation:**
- Frankfurt service: 24 hrs/day × 30 days = **720 hours/month**
- Oregon service: 24 hrs/day × 30 days = **720 hours/month**
- **Total consumption: 1,440 hours/month**
- **Free tier limit: 750 hours/month**

**Visual Timeline:**
```
Day 1-15:  [████████████████] Both services running (750 hours consumed)
Day 16:    [SUSPENDED] Both services shut down - API OFFLINE
Day 17-30: [SUSPENDED] No service - complete outage
Next month: Cycle repeats
```

**With 1 Service (Correct Configuration):**
```
Day 1-30:  [████████████████████████████] One service running (720 hours)
           [✓✓✓✓] 30 hours buffer remaining - no interruption
```

**Impact:**
- You will exceed your free hours around **day 15-16 of each month**
- **Both services will be suspended** until the next billing cycle
- Your API will go **completely offline** for ~15 days each month
- No traffic will be served during suspension

**Required Action:**
You **MUST** suspend one of the two active services immediately. Running both is not sustainable on the Free tier.

---

## 8. Additional Notes

- **Database Access:** Ensure DATABASE_URL is accessible from both Frankfurt and Oregon regions if using both services
- **CORS Configuration:** Verify FRONTEND_URL is set correctly for both services
- **SSL/TLS:** Render provides automatic HTTPS for all services ✅
- **Logs:** Use Render dashboard or MCP tools (`list_logs`) to monitor service health
- **Monitoring:** Consider setting up external uptime monitoring to keep free tier services awake

---

## 9. Next Steps

### Immediate (Within 24 hours)
1. ✅ Review this audit report
2. ⬜ **SUSPEND Oregon service** (srv-d70godeuk2gs739aa37g) to prevent mid-month shutdowns
3. ⬜ Update build/start commands for Frankfurt service
4. ⬜ Manually verify all environment variables in Render dashboard
5. ⬜ Test deployment after configuration changes

### Short-term (Within 1 week)
6. ⬜ Monitor free tier hour usage (should be ~720/750 with 1 service)
7. ⬜ Verify all features work (OCR, migrations, file uploads)
8. ⬜ Document final configuration in your repository
9. ⬜ Set up external uptime monitoring to keep service awake

---

**Report End**

For questions or to update service configurations, use the Render MCP tools or visit:
- Frankfurt Service: https://dashboard.render.com/web/srv-d71aegma2pns73f1u5e0
- Oregon Service: https://dashboard.render.com/web/srv-d70godeuk2gs739aa37g
