# Comprehensive Use Case Scenarios
## Maid Recruitment Platform — Ethiopian & Foreign Agent Perspective
### Full Lifecycle Testing: Candidate Formation → Selection → Tracking → Navigation → System Reactions

---

## Table of Contents
1. [Document Overview](#1-document-overview)
2. [Test Environment & Roles](#2-test-environment--roles)
3. [Module Navigation Map](#3-module-navigation-map)
4. [Phase 1: Agency Onboarding & Workspace Setup](#4-phase-1-agency-onboarding--workspace-setup)
5. [Phase 2: Candidate Formation & Management](#5-phase-2-candidate-formation--management)
6. [Phase 3: Partner Workspace & Candidate Sharing](#6-phase-3-partner-workspace--candidate-sharing)
7. [Phase 4: CV Generation & Document Management](#7-phase-4-cv-generation--document-management)
8. [Phase 5: Candidate Selection & Approval — Foreign Agent Deep Dive](#8-phase-5-candidate-selection--approval--foreign-agent-deep-dive)
9. [Phase 6: Post-Selection Process Tracking](#9-phase-6-post-selection-process-tracking)
10. [Phase 7: Communication & Notifications](#10-phase-7-communication--notifications)
11. [Phase 8: Admin Oversight & Governance](#11-phase-8-admin-oversight--governance)
12. [Phase 9: Edge Cases & Error Scenarios](#12-phase-9-edge-cases--error-scenarios)
13. [Phase 10: Cross-Cutting Concerns](#13-phase-10-cross-cutting-concerns)
14. [End-to-End User Journey Maps](#14-end-to-end-user-journey-maps)

---

## 1. Document Overview

### Purpose
This document defines comprehensive use case scenarios for testing the Maid Recruitment Platform from both the Ethiopian sending agency perspective and the Foreign receiving agency perspective. It covers the entire lifecycle from agency onboarding through candidate formation, partner collaboration, selection approval, process tracking, and system administration — with special emphasis on how every action ripples through the system via notifications, status changes, UI updates, real-time events, and data persistence.

### Scope
- **Primary Focus**: Ethiopian sending agencies (ethiopian_agent role) — candidate creation, management, publishing, CV generation, approval, tracking
- **Secondary Focus**: Foreign receiving agencies (foreign_agent role) — browsing, selection, approval, progress viewing, chat
- **Tertiary Focus**: Super Admin / Support Admin — governance, oversight, configuration
- **System Reaction Coverage**: Every use case documents what happens across the entire system — notifications sent, status changes, UI updates on both sides, real-time events, data state changes
- **Accessibility Testing**: Explicitly **out of scope** for this test cycle. Keyboard navigation, screen reader compatibility (ARIA labels, roles, focus management), colour contrast, and zoom/resize behaviour are not covered. A dedicated accessibility audit should be planned as a separate exercise. This decision is stated here so the gap is conscious rather than overlooked.

### How to Use This Document
- Each use case is self-contained with preconditions, step-by-step flow, expected results, and test data
- Use cases are organized in logical order — execute them sequentially for full end-to-end testing
- Each use case includes "System Reactions" showing what happens across all modules
- Edge cases and negative test scenarios are included throughout

---

## 2. Test Environment & Roles

### Test Credentials

**Pre-Seeded Admin:**
| Field | Value |
|---|---|
| Email | admin@maidrecruitment.com |
| Password | Admin@123456 |
| Role | super_admin |

**Ethiopian Agency 1 — Addis Talent Solutions PLC:**
| Field | Value |
|---|---|
| Full Name | Biruk Alemu |
| Email | addis@talent.et |
| Password | Test@1234 |
| Role | ethiopian_agent |
| Company | Addis Talent Solutions PLC |

**Ethiopian Agency 2 — Nile Workforce Services PLC:**
| Field | Value |
|---|---|
| Full Name | Selam Tesfaye |
| Email | nile@workforce.et |
| Password | Test@1234 |
| Role | ethiopian_agent |
| Company | Nile Workforce Services PLC |

**Foreign Agency 1 — Gulf Home Staff Recruitment (Saudi Arabia):**
| Field | Value |
|---|---|
| Full Name | Khalid Al-Rashid |
| Email | gulf@homestaff.sa |
| Password | Test@1234 |
| Role | foreign_agent |
| Company | Gulf Home Staff Recruitment |

**Foreign Agency 2 — Dubai Elite Maids (UAE):**
| Field | Value |
|---|---|
| Full Name | Fatima Al-Mansouri |
| Email | dubai@elitemaids.ae |
| Password | Test@1234 |
| Role | foreign_agent |
| Company | Dubai Elite Maids |

**Foreign Agency 3 — Kuwait Family Care Services (Kuwait):**
| Field | Value |
|---|---|
| Full Name | Ahmed Al-Sabah |
| Email | kuwait@familycare.kw |
| Password | Test@1234 |
| Role | foreign_agent |
| Company | Kuwait Family Care Services |

### Sample Candidate Test Data

**Candidate A — Almaz Tadesse:**
| Field | Value |
|---|---|
| Full Name | Almaz Tadesse |
| Nationality | Ethiopian |
| Date of Birth | 1995-05-15 |
| Age | 31 |
| Place of Birth | Addis Ababa |
| Gender | Female |
| Passport Number | EP1234567 |
| Passport Issue | 2023-01-01 |
| Passport Expiry | 2028-01-01 |
| Religion | Christian |
| Marital Status | Single |
| Children | 0 |
| Education | High School |
| Experience Years | 5 |
| Countries Abroad | Saudi Arabia (3 yr) |
| Languages | Arabic (Fluent), English (Intermediate), Amharic (Fluent) |
| Skills | Cooking, Cleaning, Childcare, Elderly Care |

**Candidate B — Tigist Bekele:**
| Field | Value |
|---|---|
| Full Name | Tigist Bekele |
| Nationality | Ethiopian |
| Date of Birth | 1992-08-22 |
| Age | 33 |
| Place of Birth | Bahir Dar |
| Gender | Female |
| Passport Number | EP7654321 |
| Passport Issue | 2022-06-15 |
| Passport Expiry | 2027-06-15 |
| Religion | Muslim |
| Marital Status | Married |
| Children | 2 |
| Education | Diploma |
| Experience Years | 8 |
| Countries Abroad | UAE (5 yr), Qatar (3 yr) |
| Languages | Arabic (Fluent), English (Basic), Amharic (Fluent) |
| Skills | Cooking, Cleaning, Ironing, Laundry, Baby Care |

**Candidate C — Hiwot Alemu:**
| Field | Value |
|---|---|
| Full Name | Hiwot Alemu |
| Nationality | Ethiopian |
| Date of Birth | 1998-11-03 |
| Age | 27 |
| Place of Birth | Adama |
| Gender | Female |
| Passport Number | EP9988776 |
| Passport Issue | 2024-03-10 |
| Passport Expiry | 2029-03-10 |
| Religion | Christian |
| Marital Status | Single |
| Children | 0 |
| Education | Bachelor'\''s Degree |
| Experience Years | 3 |
| Countries Abroad | Kuwait (3 yr) |
| Languages | Arabic (Fluent), English (Advanced), Amharic (Fluent) |
| Skills | Cooking, Childcare, Elderly Care, First Aid |

**Candidate D — Muluwork Desta:**
| Field | Value |
|---|---|
| Full Name | Muluwork Desta |
| Nationality | Ethiopian |
| Date of Birth | 1997-02-28 |
| Age | 29 |
| Place of Birth | Hawassa |
| Gender | Female |
| Passport Number | EP5544332 |
| Passport Issue | 2023-09-01 |
| Passport Expiry | 2028-09-01 |
| Religion | Christian |
| Marital Status | Divorced |
| Children | 1 |
| Education | High School |
| Experience Years | 4 |
| Countries Abroad | Saudi Arabia (2 yr), Lebanon (2 yr) |
| Languages | Arabic (Fluent), English (Basic), Amharic (Fluent) |
| Skills | Cooking, Cleaning, Childcare |

### System Prerequisites
- Frontend: `http://localhost:3000` (or deployed URL)
- Backend API: `http://localhost:8080` (or deployed URL)
- PostgreSQL database running with migrations applied
- S3-compatible storage (MinIO for local, Cloudflare R2 for production)
- Tesseract OCR installed (in Docker container)
- SMTP service configured (or email service disabled fallback) — **note the active mode** before running; UC-000 verifies this and all downstream UCs that expect email must be interpreted accordingly
- WebSocket endpoint accessible at `ws://localhost:8080/api/v1/ws/notifications`
- Test browsers: Chrome, Firefox, Edge, Safari

---

## 3. Module Navigation Map

### 3.1 Role-Based Home Paths

| Role | Home Path |
|---|---|
| ethiopian_agent | `/dashboard/agency` |
| foreign_agent | `/dashboard/employer` |
| super_admin / support_admin | `/admin/dashboard` |

### 3.2 Ethiopian Agent — Full Navigation Tree

```
Landing Page (/)
  └── Login (/login)
        └── Ethiopian Agency Dashboard (/dashboard/agency)
              │
              ├── Candidate Library (/candidates)
              │     ├── Add New Candidate (/candidates/new)
              │     ├── Candidate Detail (/candidates/[id])
              │     │     ├── Edit Candidate (/candidates/[id]/edit)
              │     │     ├── CV Preview/Download (/candidates/[id]/cv)
              │     │     └── Process Tracking (/candidates/[id]/tracking)
              │     └── [Batch Operations] (via selection mode)
              │           ├── Batch Share
              │           ├── Bulk Publish
              │           ├── Bulk CV Download
              │           ├── Bulk CV Regenerate
              │           └── Bulk Override
              │
              ├── My Selections (/selections)
              │     └── Selection Detail (/selections/[id])
              │           ├── [Approve/Reject Actions]
              │           ├── [Upload Employer Docs]
              │           ├── [Progress Tracking]
              │           ├── [Candidate Link] (/candidates/[candidate_id])
              │           ├── [Chat about Candidate] (/partners/chat?candidate_id=...)
              │           └── [Tracking Link] (/candidates/[candidate_id]/tracking)
              │
              ├── Partner Workspaces (/partners)
              │     ├── [Partner Switcher]
              │     ├── Workspace Chat (/partners/chat)
              │     │     └── [Candidate-scoped Threads]
              │     └── [CV Defaults Editor]
              │
              ├── Process Tracking Hub (/tracking)
              │
              ├── Notifications (/notifications)
              │     └── [Notification items link to candidates/selections/chats]
              │
              └── Settings (/settings)
                    ├── Profile
                    ├── Security
                    └── Preferences
```

### 3.3 Foreign Agent — Full Navigation Tree

```
Landing Page (/)
  └── Login (/login)
        └── Foreign Employer Dashboard (/dashboard/employer)
              │
              ├── Browse Candidates (/candidates)
              │     ├── Candidate Detail (/candidates/[id])
              │     │     ├── CV Download (/candidates/[id]/cv)
              │     │     ├── Select Candidate [action]
              │     │     └── View Documents
              │     └── [Filters: Skills, Age, Experience, Languages]
              │
              ├── My Selections (/selections)
              │     └── Selection Detail (/selections/[id])
              │           ├── [Approve/Reject Actions (dual approval)]
              │           ├── [Upload Employer Contract + ID]
              │           ├── [Progress Timeline — Read Only]
              │           ├── [Candidate Link] (/candidates/[candidate_id])
              │           ├── [Chat about Candidate] (/partners/chat?candidate_id=...)
              │           └── [Tracking Link] (/candidates/[candidate_id]/tracking)
              │
              ├── Partner Workspaces (/partners)
              │     ├── [View Partner Details]
              │     └── [Workspace Activity]
              │
              ├── Workspace Chat (/partners/chat)
              │     └── [Workspace + Candidate Threads]
              │
              ├── Process Tracking Hub (/tracking)
              │
              ├── Notifications (/notifications)
              │
              └── Settings (/settings)
                    ├── Profile
                    ├── Security
                    └── Preferences
```

### 3.4 Cross-Module Navigation Paths

| From Module | Action | Navigates To |
|---|---|---|
| Dashboard → Smart Alert (passport expiry) | Click item | `/candidates/[id]` |
| Dashboard → Smart Alert (selection expiry) | Click item | `/selections/[id]` |
| Dashboard → Smart Alert (flight update) | Click item | `/candidates/[id]/tracking` |
| Dashboard → Pending Actions | "Review selections" | `/selections` |
| Dashboard → Quick Actions | "Add a new candidate" | `/candidates/new` |
| Dashboard → Recent Activity | Click candidate name | `/candidates/[id]` |
| Candidates List → Candidate Card | Click name | `/candidates/[id]` |
| Candidate Detail → Edit | Click "Edit Candidate" | `/candidates/[id]/edit` |
| Candidate Detail → CV | Click "Download CV" | `/candidates/[id]/cv` |
| Candidate Detail → Tracking | Click "Open Process Tracking" | `/candidates/[id]/tracking` |
| Candidate Detail → Publish | Click "Publish" | Stays on page, opens dialog |
| Candidate Detail → Share | Click "Share With Partners" | Opens dialog |
| Candidate Detail → Select (Foreign) | Click "Select Candidate" | Opens dialog → `/selections/[new_id]` |
| Candidate Detail → Delete (Ethiopian) | Click "Delete" | Confirmation → `/candidates` |
| Selections List → Selection Item | Click item | `/selections/[id]` |
| Selection Detail → Candidate | Click "Open Candidate" | `/candidates/[candidate_id]` |
| Selection Detail → Chat | Click "Open chat about this candidate" | `/partners/chat?candidate_id=...` |
| Selection Detail → Tracking | Click "Open Tracking" | `/candidates/[candidate_id]/tracking` |
| Selection Detail → Approve | Click "Approve" | Confirmation → status update |
| Selection Detail → Reject | Click "Reject" | Dialog → status update |
| Selection Detail → Upload Docs (Foreign) | Click upload area | File picker → upload |
| Partners → Switch Workspace | Click agency button | Switches active pairing context |
| Partners → Chat | Click "Open Chat" | `/partners/chat` |
| Partners → Candidates | Click "Go to full library" | `/candidates` |
| Partners → Selections | Click "Open selections" | `/selections` |
| Chat → Notification | Click notification | `/partners/chat` or respective entity page |
| Notifications → Item | Click notification | `/candidates/[id]` or `/selections/[id]` |
| Notifications → Bell dropdown item | Click | `/candidates/[id]` or `/selections/[id]` |
| Sidebar → Any Module | Click nav link | Respective module page |
| Header → Notification Bell | Click bell | Dropdown → `/notifications` |
| Header → User Avatar | Click dropdown | `/settings` |

### 3.5 Admin Navigation Tree

```
Admin Login (/admin/login)
  └── Admin Dashboard (/admin/dashboard)
        │
        ├── Pending Approvals (/admin/agencies/pending)
        │     └── Agency Detail (/admin/agencies/[id])
        │
        ├── All Agencies (/admin/agencies)
        │     └── Agency Detail (/admin/agencies/[id])
        │
        ├── Pair Workspaces (/admin/pairings)
        │
        ├── Candidates Overview (/admin/candidates) [read-only]
        │
        ├── Selections Overview (/admin/selections) [read-only]
        │
        ├── Admin Management (/admin/admins) [super_admin only]
        │
        ├── Audit Logs (/admin/audit-logs)
        │
        └── Platform Settings (/admin/settings) [super_admin only]
```

---

## 3.6 Pre-Flight Environment Verification (Use Case 000)

### UC-000: Pre-Flight Environment Verification

**Priority**: Critical
**Type**: Setup
**Est. Time**: 5 min

**User Story**: Before running any functional test case, the tester must confirm the environment is alive, configured correctly, and all baseline services are reachable.

**Preconditions**:
- Backend and frontend URLs known
- No authentication required for health endpoints
- Tester has access to API documentation or is prepared to inspect response bodies

**Scenario Flow**:
| Step | Action | Expected Result |
|---|---|---|
| 1 | GET `http://localhost:8080/api/v1/health` | Returns `200 OK` with body `{"status":"up","timestamp":"2026-07-08T14:30:00Z","version":"1.0.0","db":"connected","s3":"reachable","ocr":"available","email":"enabled"}` |
| 2 | GET `http://localhost:3000` | Returns `200 OK` — home/login page loads without JS errors (check browser console) |
| 3 | GET `http://localhost:8080/api/v1/health/db` | Returns `200 OK` — `{"db":"connected","migrations_current":true}` |
| 4 | GET `http://localhost:8080/api/v1/health/storage` | Returns `200 OK` — `{"s3":"reachable","bucket_count":3}` |
| 5 | GET `http://localhost:8080/api/v1/health/ocr` | Returns `200 OK` — `{"tesseract":"available","version":"5.3.3"}` |
| 6 | GET `http://localhost:8080/api/v1/health/email` | If SMTP configured: `{"email":"enabled","provider":"smtp","mode":"live"}`. If SMTP not configured: `{"email":"disabled","mode":"simulated","note":"Emails are logged to stdout in dev mode"}`. **Record this value** — all subsequent UCs that include "email sent" expectations must be reinterpreted when mode is `simulated`. |
| 7 | Open browser devtools Network tab, verify WebSocket handshake | `101 Switching Protocols` on `ws://localhost:8080/api/v1/ws/notifications` |
| 8 | Verify timezone discrepancy | Compare server response timestamp (Step 1's `timestamp` field) with local machine time. If offset > 5 seconds, note the skew — date-sensitive assertions in later UCs may fail. |

**Expected Results**:
- All health endpoints return `200 OK`
- Email mode recorded for the session
- Timezone skew noted (if any)
- System is ready for UC-001 onwards

**Negative Scenarios**:
| Variation | Expected Behavior |
|---|---|
| DB migration not run | Health step 3 returns `{"db":"unhealthy","migrations_current":false}` — **blocking**: do not proceed until fixed |
| S3 bucket missing | Health step 4 returns `{"s3":"reachable","error":"bucket 'uploads' not found"}` — **blocking**: do not proceed |
| Tesseract not installed | Health step 5 returns `{"tesseract":"unavailable","error":"executable not found"}` — **blocking**: CV generation UCs will fail |
| WebSocket fails | Step 7 returns non-101 — **non-blocking warning**: real-time notifications will not arrive; refresh-based fallback still works |

---

## 4. Phase 1: Agency Onboarding & Workspace Setup

### UC-001: Ethiopian Agency Self-Registration

**Priority**: Critical
**Type**: Functional
**Est. Time**: 10 min

**User Story**: As a new Ethiopian recruitment agency, I want to register on the platform so that I can start recruiting candidates for foreign partners.

**Preconditions**:
- No existing session
- Valid email not already registered
- Access to registration page

**Test Data**:
| Field | Value |
|---|---|
| Full Name | Biruk Alemu |
| Company Name | Addis Talent Solutions PLC |
| Email | addis@talent.et |
| Password | Test@1234 |
| Confirm Password | Test@1234 |
| Role | Ethiopian Agency |

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Navigate to landing page `/` | Landing page loads with hero, features, CTA buttons | `/` |
| 2 | Click "Get Started" or "Register as Ethiopian Agency" | Registration form displays | `/register` |
| 3 | Enter Full Name: "Biruk Alemu" | Field accepts input | — |
| 4 | Enter Company Name: "Addis Talent Solutions PLC" | Field accepts input | — |
| 5 | Enter Email: "addis@talent.et" | Field accepts input | — |
| 6 | Enter Password: "Test@1234" | Password strength indicator shows (Medium/Strong) | — |
| 7 | Enter Confirm Password: "Test@1234" | Confirm matches password | — |
| 8 | Select "Ethiopian Agency" account type card | Card highlights as selected | — |
| 9 | Check "I agree to Terms and Conditions" | Checkbox checked | — |
| 10 | Click "Submit For Review" | Form submits, loading spinner shown | — |
| 11 | — | Success: Account created, redirect to pending page | `/register/pending?email=addis@talent.et&company_name=Addis%20Talent%20Solutions%20PLC&role=ethiopian_agent` |

**Expected Results**:
- ✅ Account created with `account_status = pending_approval`
- ✅ Redirected to pending approval page
- ✅ Email verification token generated and sent (or shown if email disabled)
- ✅ Success toast notification displayed
- ✅ User cannot access dashboard until email verified AND admin approved
- ✅ Backend: `users` table row created with hashed password, UUID generated

**System Reactions**:
- Users table: New row with `role = ethiopian_agent`, `account_status = pending_approval`, `email_verified = false`
- Email verification: Token stored in `email_verification_tokens` table
- No notification sent (account not yet active)

**Negative Scenarios**:
| Variation | Expected Behavior |
|---|---|
| Register with existing email | Error: "Email already registered" |
| Password too short (< 8 chars) | Validation error on field |
| Passwords don'\''t match | Validation error on confirm field |
| Empty required fields | Field-level validation errors |
| Uncheck terms checkbox | Submit button disabled |
| Invalid email format | Validation error on email field |
| Network timeout during submit | Error toast, retry option |

---

### UC-002: Email Verification Flow

**Priority**: Critical
**Type**: Functional
**Est. Time**: 5 min

**User Story**: As a newly registered Ethiopian agent, I want to verify my email address so that I can proceed with platform access.

**Preconditions**:
- Account created (UC-001)
- Verification email sent (or token visible in URL/local storage)

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | From pending page, locate verification link/button | Pending page shows status info | `/register/pending` |
| 2 | Click verification link from email (or navigate with token) | Token validated, email marked verified | `/register/verify?token=...` |
| 3 | — | Success message: "Email verified successfully" | — |
| 4 | Navigate to login page | Login form displays | `/login` |
| 5 | Attempt to login | Should fail with "Account pending admin approval" | — |

**Expected Results**:
- ✅ `email_verified` set to `true` in users table
- ✅ EmailVerificationToken marked as used
- ✅ User still blocked by `account_status = pending_approval`
- ✅ Cannot access dashboard

**System Reactions**:
- Users table: `email_verified` → `true`
- Email verification tokens: Token marked as `used_at` timestamped
- Agency approval requests: A new `agency_approval_requests` row created with `status = pending`
- Admin side: Agency now appears in admin pending approvals list

**Negative Scenarios**:
| Variation | Expected Behavior |
|---|---|
| Expired verification token (24h+) | Error: "Token expired, request new one" |
| Already used token | Error: "Token already used" |
| Invalid/malformed token | Error: "Invalid verification token" |
| Resend verification (rate limited) | Cooldown message, 5/10min limit |
| Attempt login before verification | Error: "Please verify your email first" |

---

### UC-003: Admin Approval of Ethiopian Agency

**Priority**: Critical
**Type**: Functional
**Est. Time**: 5 min

**User Story**: As a super admin, I want to review and approve pending Ethiopian agencies so that they can access the platform.

**Preconditions**:
- Admin account exists and logged in
- At least one pending Ethiopian agency (from UC-001/UC-002)

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Login as super_admin | Redirected to admin dashboard | `/admin/login` → `/admin/dashboard` |
| 2 | View dashboard metrics | "Pending Approvals" count > 0, queue listed | — |
| 3 | Click "Review queue" or pending agency card | Pending approvals list shown | `/admin/agencies/pending` |
| 4 | Filter tab: "Ethiopian Agencies" | Only Ethiopian agencies shown | — |
| 5 | Click "Review" on "Addis Talent Solutions PLC" | Agency detail page loads | `/admin/agencies/[id]` |
| 6 | Review agency profile: Company, Contact, Email, Type, Status | All data visible and correct | — |
| 7 | Click "Approve Agency" | Confirmation dialog appears | — |
| 8 | Confirm approval | Status changes to "active" | — |
| 9 | — | Success toast, agency status now "Active" | — |

**Expected Results**:
- ✅ Agency `account_status` changed to `active`
- ✅ `agency_approval_requests` record created with `approved` status
- ✅ Audit log entry: "Approved agency [name]"
- ✅ Notification sent to agency (if email enabled)
- ✅ Agency can now login and access dashboard

**System Reactions**:
- Users table: `account_status` → `active`
- Agency approval requests: Status → `approved`, `reviewed_by` → admin ID, timestamped
- Audit logs: New entry with `action = "Approved agency"`, target = agency ID
- Agency side: If logged in, would now see dashboard (but still needs pairing)
- Email notification: Sent to agency (if SMTP configured)

**Negative Scenarios**:
| Variation | Expected Behavior |
|---|---|
| Reject agency | Status changes to "rejected", reason required |
| Approve already-approved agency | Error or no-op |
| Support admin tries to approve | Allowed (support_admin can approve/reject) |
| Approve with no pairing created | User can login but sees `/waiting` |

---

### UC-004: Foreign Agency Self-Registration

**Priority**: Critical
**Type**: Functional
**Est. Time**: 8 min

**User Story**: As a new foreign recruitment agency, I want to register on the platform so that I can find candidates from Ethiopian partners.

**Preconditions**:
- No existing session
- Valid email not already registered

**Step-by-Step Flow**:
| Step | Action | Expected Result |
|---|---|---|
| 1 | Navigate to `/register` | Registration form loads |
| 2 | Select "Foreign Agency (Jordan)" account type | Card highlights |
| 3 | Enter Full Name: "Khalid Al-Rashid" | Field accepts |
| 4 | Enter Company Name: "Gulf Home Staff Recruitment" | Field accepts |
| 5 | Enter Email: "gulf@homestaff.sa" | Field accepts |
| 6 | Enter Password: "Test@1234", Confirm same | Passwords match |
| 7 | Accept terms, click "Submit For Review" | Account created |
| 8 | Complete email verification | Email verified |
| 9 | Admin approves foreign agency | Same approval flow as UC-003 |

**Repeat for**:
- Dubai Elite Maids: "dubai@elitemaids.ae"
- Kuwait Family Care: "kuwait@familycare.kw"

**Expected Results**:
- ✅ Foreign agent accounts created with `role = foreign_agent`
- ✅ Admin approval required
- ✅ Email verification required
- ✅ Same flows as Ethiopian agent

---

### UC-005: Admin Creates Agency Pairings (Workspaces)

**Priority**: Critical
**Type**: Functional
**Est. Time**: 10 min

**User Story**: As a super admin, I want to create workspace pairings between Ethiopian and foreign agencies so they can collaborate on candidates.

**Preconditions**:
- At least one approved Ethiopian agency (from UC-003)
- At least three approved Foreign agencies (from UC-004)

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Navigate to Pair Workspaces | Pairings page loads | `/admin/pairings` |
| 2 | View left panel "Create a private workspace" | Form with dropdowns visible | — |
| 3 | Select Ethiopian agency: "Addis Talent Solutions PLC" | Dropdown populates | — |
| 4 | Select Foreign agency: "Gulf Home Staff Recruitment" | Dropdown populates | — |
| 5 | Enter internal note: "Primary partnership for Saudi housemaid recruitment" | Text area accepts input | — |
| 6 | Click "Create private workspace" | Form submits, spinner shown | — |
| 7 | — | Success toast, pairing appears in right panel | — |
| 8 | Create second pairing: Addis Talent + Dubai Elite Maids | Second pairing created | — |
| 9 | Enter note: "UAE market partnership" | — | — |
| 10 | Create third pairing: Addis Talent + Kuwait Family Care | Third pairing created | — |
| 11 | Enter note: "Kuwait market expansion" | — | — |

**Expected Results**:
- ✅ `agency_pairings` rows created with `status = active`
- ✅ Both Ethiopian and Foreign agencies linked
- ✅ Admin internal note saved
- ✅ Pairing visible in right panel with status badge "Active"
- ✅ Agency users will now see dashboard (not waiting page)

**System Reactions**:
- Database: Three `agency_pairings` rows created
- Ethiopian agent side: On next login/refresh, will see dashboard instead of waiting page
- Foreign agent side: On next login/refresh, will see employer dashboard
- PartnerSwitcher: Will show 3 workspace options for Ethiopian agent
- Sidebar: Full navigation becomes available for both roles

**Negative Scenarios**:
| Variation | Expected Behavior |
|---|---|
| Select same agency twice | Error or dropdown removes already-selected |
| Create pairing with unapproved agency | Not available in dropdown (only active agencies) |
| Create duplicate pairing (same Ethiopian + same Foreign) | **Blocked**: Backend returns `409 Conflict` — "A pairing between these agencies already exists." The `agency_pairings` table enforces a unique constraint on `(ethiopian_user_id, foreign_user_id)` for active pairings. The UI must disable the Create button or show an inline error. |
| Suspend an active pairing | Status changes, both agencies affected immediately |

---

### UC-006: Ethiopian Agent — First Login & Dashboard Landing

**Priority**: Critical
**Type**: Functional
**Est. Time**: 5 min

**User Story**: As an approved Ethiopian agent with active workspace, I want to log in and see my dashboard so I can manage my recruitment operations.

**Preconditions**:
- Account approved (UC-003)
- At least one active pairing (UC-005)
- Email verified

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Navigate to login page | Login form displays | `/login` |
| 2 | Enter email: "addis@talent.et" | Field accepts input | — |
| 3 | Enter password: "Test@1234" | Password field masked | — |
| 4 | Check "Remember me" | Email saved to localStorage | — |
| 5 | Click "Sign In" | Loading spinner, API call | — |
| 6 | — | Redirected to Ethiopian agency dashboard | `/dashboard/agency` |

**Dashboard Verification**:
| Element | Expected Content |
|---|---|
| Hero Card | Welcome message, "Addis Talent Solutions PLC", partner badge showing active workspaces |
| Today at a Glance | "Shared in workspace: 0", "Profiles in tracking: 0", "Approval queue: 0" |
| Stat Cards | Total Candidates: 0, Available: 0, Selected: 0, In Progress: 0 |
| Recent Activity | Empty state: "No activity yet" |
| Pending Actions | "No pending actions" or empty state |
| Quick Actions | "Add New Candidate", "View All Selections", "Open candidate workspace" buttons |
| Smart Alerts | "All clear" or empty state |

**Sidebar Verification**:
| Nav Link | Expected |
|---|---|
| Agency Home | Active (current page) |
| Partner Workspaces | Link to `/partners` |
| Candidates | Link to `/candidates` |
| Add Candidate | Link to `/candidates/new` |
| Selections | Link to `/selections` |
| Chat | Link to `/partners/chat` |
| Process Tracking | Link to `/tracking` |
| Notifications | Link to `/notifications` (with badge count = 0) |
| Settings | Link to `/settings` |
| PartnerSwitcher | Visible, shows 3 active workspaces (Gulf Home, Dubai Elite, Kuwait Family) |
| User Avatar | Shows name "Biruk Alemu", company, logout |

**Expected Results**:
- ✅ JWT token set in `auth_session` cookie
- ✅ User session created in `user_sessions` table
- ✅ Dashboard loads with all sections
- ✅ Sidebar shows role-appropriate nav links
- ✅ PartnerSwitcher lists all 3 active pairings
- ✅ Notification bell shows 0 unread

**System Reactions**:
- Auth: JWT generated, session stored in `user_sessions`
- Cookie: `auth_session` set (HttpOnly, Secure, SameSite)
- Dashboard API: Returns counts from database (all zero initially)
- Smart Alerts API: No alerts returned

**Negative Scenarios**:
| Variation | Expected Behavior |
|---|---|
| Wrong password (8 attempts/15min) | "Invalid credentials" → account lockout after threshold |
| Non-existent email | "Invalid credentials" (no user enumeration) |
| Unverified email | "Please verify your email first" |
| Unapproved account | "Account pending admin approval" |
| Suspended account | "Account suspended. Contact admin." |
| Maintenance mode on | Non-admin users see maintenance page |
| Concurrent login from different device | Both sessions active (configurable) |

### UC-007: Foreign Agent — First Login & Dashboard Landing

**Priority**: Critical
**Type**: Functional
**Est. Time**: 5 min

**User Story**: As an approved foreign agent with active workspace, I want to log in and see my employer dashboard so I can start browsing candidates.

**Preconditions**:
- Account approved
- Active pairing with Ethiopian agency (UC-005)
- Ethiopian agent has NOT yet published any candidates

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Navigate to login page | Login form displays | `/login` |
| 2 | Enter email: "gulf@homestaff.sa" | Field accepts | — |
| 3 | Enter password: "Test@1234" | Field masked | — |
| 4 | Click "Sign In" | Processing | — |
| 5 | — | Redirected to employer dashboard | `/dashboard/employer` |

**Dashboard Verification**:
| Element | Expected Content |
|---|---|
| Hero Card | "Browse & select candidates to build your team", active partner badge |
| Quick Pulse | "Available Now: 0", "Pending Approvals: 0", "Approved Profiles: 0" |
| Select from live candidates | Empty state: "No candidates available yet" |
| My live selections | Empty state: "No active selections" |
| Recently approved | Empty state: "No approved selections yet" |

**Sidebar Verification**:
| Nav Link | Expected |
|---|---|
| Employer Home | Active (current page) |
| Browse Candidates | Link to `/candidates` |
| My Selections | Link to `/selections` |
| Partner Workspaces | Link to `/partners` |
| Chat | Link to `/partners/chat` |
| Process Tracking | Link to `/tracking` |
| Notifications | Link to `/notifications` |
| Settings | Link to `/settings` |

**Expected Results**:
- ✅ Foreign agent dashboard loads with all sections
- ✅ Empty states shown correctly (no candidates published yet)
- ✅ Sidebar shows foreign agent-specific nav links
- ✅ No "Add Candidate" or edit capabilities visible

---

## 5. Phase 2: Candidate Formation & Management

### UC-008: Create Candidate Profile — Full Flow

**Priority**: Critical
**Type**: Functional
**Est. Time**: 15 min

**User Story**: As an Ethiopian agent, I want to create a comprehensive candidate profile so that I can present her qualifications to foreign partners.

**Preconditions**:
- Logged in as Ethiopian agent (UC-006)
- Access to candidate creation page

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Click "Add Candidate" from sidebar or Quick Actions | Create candidate form loads | `/candidates/new` |
| 2 | Verify breadcrumbs: Dashboard > Candidates > Add New | Breadcrumbs correct | — |
| 3 | Enter Full Name: "Almaz Tadesse" | Field accepts input | — |
| 4 | Enter Nationality: "Ethiopian" | Field accepts input | — |
| 5 | Select Date of Birth: 1995-05-15 | Date picker works, Age auto-calculates to 31 | — |
| 6 | Enter Place of Birth: "Addis Ababa" | Field accepts input | — |
| 7 | Select Gender: "Female" | Radio/select works | — |
| 8 | Enter Passport Number: "EP1234567" | Field accepts input | — |
| 9 | Select Passport Issue Date: 2023-01-01 | Date picker works | — |
| 10 | Select Passport Expiry Date: 2028-01-01 | Date validation: must be after issue date | — |
| 11 | Select Religion: "Christian" | Dropdown works | — |
| 12 | Select Marital Status: "Single" | Dropdown works | — |
| 13 | Enter Children: 0 | Numeric input accepts | — |
| 14 | Select Education Level: "High School" | Dropdown works | — |
| 15 | Enter Experience Years: 5 | Numeric input accepts | — |
| 16 | Add Experience Abroad: Country "Saudi Arabia", Years "3" | Dynamic entry added, can add more | — |
| 17 | Add Languages: Amharic (Fluent), Arabic (Fluent), English (Intermediate) | Tags/chips created | — |
| 18 | Add Skills: Cooking, Cleaning, Childcare, Elderly Care | Tags/chips created | — |
| 19 | Enter Remark: "Experienced housemaid with excellent references" | Text area accepts input | — |
| 20 | Click "Save Candidate" | Form submits, progress overlay shows | — |
| 21 | — | Candidate created, redirect to detail page | `/candidates/[id]` |

**Expected Results**:
- ✅ Candidate created with `status = draft`
- ✅ All 40+ fields saved correctly
- ✅ Age auto-calculated from DOB
- ✅ Redirected to candidate detail page
- ✅ Success toast notification
- ✅ Candidate appears in candidates list

**System Reactions**:
- Database: `candidates` row created with all fields, `status = draft`, `created_by` = Ethiopian agent ID
- Candidates list: New candidate appears with "Draft" badge
- Dashboard stats: Total Candidates increments by 1
- No notifications sent (candidate is draft, not yet published)

**Form Validation Tests**:
| Scenario | Expected Behavior |
|---|---|
| Submit empty form | All required fields show validation errors |
| Future date of birth | Validation error: "Date of birth cannot be in the future" |
| Passport expiry before issue | Validation error: "Expiry must be after issue date" |
| Negative experience years | Validation error or min value constraint |
| Invalid email in fields | Validation error where applicable |
| Very long name (>100 chars) | Truncation or validation error |
| Special characters in name | Should accept (Ethiopian names may contain) |
| Age > 100 | Validation error: implausible age |

**Create additional candidates for testing**:
- Repeat UC-008 to create Candidate B (Tigist Bekele)
- Repeat UC-008 to create Candidate C (Hiwot Alemu)
- Repeat UC-008 to create Candidate D (Muluwork Desta)

---

### UC-009: Draft Auto-Save & Recovery

**Priority**: High
**Type**: Functional
**Est. Time**: 8 min

**User Story**: As an Ethiopian agent filling a long candidate form, I want my progress auto-saved so I don't lose data if I navigate away.

**Preconditions**:
- Logged in as Ethiopian agent
- On candidate creation page

**Step-by-Step Flow**:
| Step | Action | Expected Result |
|---|---|---|
| 1 | Navigate to `/candidates/new` | Fresh form loads |
| 2 | Partially fill form (name, nationality, DOB, 2 skills) | Form accepts partial data |
| 3 | Wait 5+ seconds (auto-save fires every 500ms) | Data saved to localStorage |
| 4 | Close browser tab | — |
| 5 | Reopen browser, navigate to `/candidates/new` | — |
| 6 | — | Draft restored banner appears: "We found a saved draft. [Restore] [Clear]" |
| 7 | Click "Restore" | Previously entered data populates form |
| 8 | Clear a field, wait for auto-save | Updated draft saved |
| 9 | Click "Clear draft" | Draft cleared, fresh form shown |

**Expected Results**:
- ✅ Form values persist in localStorage
- ✅ Document references persisted in IndexedDB (if any)
- ✅ Draft recovery banner appears on return
- ✅ Clear draft removes saved data
- ✅ Draft does NOT persist after successful submission

---

### UC-010: Passport Upload with OCR Processing

**Priority**: Critical
**Type**: Functional
**Est. Time**: 10 min

**User Story**: As an Ethiopian agent, I want to upload a candidate's passport and have the data automatically extracted via OCR so I don't have to type it manually.

**Preconditions**:
- Candidate created (UC-008)
- Valid passport image file (JPG/PNG/PDF, < 10MB)
- Tesseract OCR engine running on backend

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Navigate to candidate detail page | Candidate info displayed | `/candidates/[id]` |
| 2 | Scroll to "Documents" section | Upload areas visible: Passport, Photo, Video | — |
| 3 | Click Upload under "Passport" | File picker opens | — |
| 4 | Select passport image file | Upload starts, progress shown | — |
| 5 | — | "Processing passport OCR..." indicator appears | — |
| 6 | Wait for OCR completion (≤ 10 seconds) | — | — |
| 7 | — | OCR completes, fields auto-populated: Passport Number, Gender, DOB, Issue Date, Expiry Date, Issuing Authority, Nationality | — |
| 8 | Verify auto-filled data against actual passport | Data matches | — |

**Expected Results**:
- ✅ File uploads to S3/MinIO successfully
- ✅ Document record created in `documents` table with `document_type = passport`
- ✅ `passport_data` record created with parsed fields and confidence score
- ✅ Candidate fields updated with OCR-extracted data
- ✅ Passport preview thumbnail displayed
- ✅ Success notification shown

**System Reactions**:
- S3/MinIO: File stored with candidate_id prefix
- Documents table: New row with `document_type = passport`, `file_url` pointing to S3
- Passport_data table: Parsed fields, confidence score, MRZ lines stored
- Candidates table: Relevant fields updated with OCR data (passport_number, gender, dob, etc.)
- Candidate detail page: Passport thumbnail visible, OCR fields auto-filled

**OCR Accuracy Checklist**:
| Field | Check Against Source |
|---|---|
| Full Name (holder_name) | Matches passport MRZ/visual zone |
| Passport Number | Matches exactly |
| Nationality | "ETHIOPIAN" or "ETH" |
| Date of Birth | Correct format (YYYY-MM-DD) |
| Gender | M/F match |
| Issue Date | Matches passport |
| Expiry Date | Matches passport |
| Issuing Authority | Matches passport |
| MRZ Lines | Both lines parsed correctly |

**Negative Scenarios**:
| Variation | Expected Behavior |
|---|---|
| Upload blurry/unclear image | Lower confidence score, warning to user |
| Upload non-passport image | OCR returns low confidence, fields may be incorrect |
| Upload > 10MB file | Rejected before upload with size error |
| Upload unsupported format (HEIC, BMP) | Rejected with format error |
| Network failure during upload | Retry option, error toast |
| OCR service timeout | Error: "OCR processing failed, please try again" |
| Passport already uploaded for candidate | Replaces or warns about duplicate |

---

### UC-011: OCR Preview (Parse Without Save)

**Priority**: Medium
**Type**: Functional
**Est. Time**: 5 min

**User Story**: As an Ethiopian agent, I want to preview OCR parsing results before saving to ensure accuracy.

**Preconditions**:
- Logged in as Ethiopian agent
- On candidate creation page OR existing candidate detail page

**Step-by-Step Flow**:
| Step | Action | Expected Result |
|---|---|---|
| 1 | Click passport upload area | File picker opens |
| 2 | Look for "Parse Preview" option | Button visible (rate limited: 12/min) |
| 3 | Select passport image, click "Parse Preview" | OCR processes without saving |
| 4 | — | Preview dialog shows extracted fields with confidence scores |
| 5 | Review each field | Fields highlighted green (high confidence), yellow (medium), red (low) |
| 6 | Accept or manually correct fields | User can edit before final save |
| 7 | Click "Apply to Profile" | Data written to candidate record |

---

### UC-012: Upload Photo & Video Documents

**Priority**: High
**Type**: Functional
**Est. Time**: 8 min

**User Story**: As an Ethiopian agent, I want to upload a candidate's photo and video interview so foreign partners can visually assess the candidate.

**Preconditions**:
- Candidate exists (UC-008)
- Photo file (JPG/PNG, < 10MB)
- Video file (MP4, < 50MB)

**Step-by-Step Flow**:
| Step | Action | Expected Result |
|---|---|---|
| 1 | Navigate to candidate detail page | `/candidates/[id]` |
| 2 | In Documents section, click Upload under "Full Body Photo" | File picker opens, filtered for images |
| 3 | Select photo file | Upload progress bar shown |
| 4 | Wait for upload completion | Photo thumbnail preview displayed |
| 5 | Click Upload under "Video Interview" | File picker opens, filtered for video |
| 6 | Select video file | Upload progress shown (larger file, longer time) |
| 7 | Wait for upload completion | Video thumbnail/player shown |
| 8 | Click on photo thumbnail | Lightbox/modal with full-size photo |
| 9 | Click on video thumbnail | Video player opens, can play interview |

**Expected Results**:
- ✅ Photo uploaded to S3, `document_type = photo`
- ✅ Video uploaded to S3, `document_type = video`
- ✅ Thumbnails/previews displayed correctly
- ✅ Download/view links work
- ✅ Files can be removed (delete button with confirmation)
- ✅ Re-upload replaces existing document

**System Reactions**:
- S3/MinIO: Photo and video files stored
- Documents table: Two new rows for photo and video
- Candidate detail page: Thumbnails visible, documents section updated
- Foreign agent side: When candidate is published, foreign agent will see these documents

**Negative Scenarios**:
| Variation | Expected Behavior |
|---|---|
| Upload photo > 10MB | Rejected with size error |
| Upload video > 50MB | Rejected with size error |
| Upload photo with wrong orientation | Should handle EXIF rotation (or display as-is) |
| Upload corrupt image file | Upload fails, error message |
| Upload non-video file to video field | Rejected by type filter |
| Network timeout during video upload (large file) | Retry option, partial upload handled |

---

### UC-013: Edit Candidate Profile

**Priority**: High
**Type**: Functional
**Est. Time**: 10 min

**User Story**: As an Ethiopian agent, I want to edit an existing candidate's profile so I can update information or correct errors.

**Preconditions**:
- Candidate exists in `draft` or `available` status
- Logged in as the creating Ethiopian agent

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Navigate to candidate detail page | Profile displayed | `/candidates/[id]` |
| 2 | Click "Edit Candidate" button | Edit form loads with pre-filled data | `/candidates/[id]/edit` |
| 3 | Verify "Safe edit mode" banner | Banner visible: "Editing this profile..." | — |
| 4 | Change Full Name: "Almaz Tadesse Beyene" | Field updates | — |
| 5 | Change Experience Years: 5 → 6 | Field updates | — |
| 6 | Add new Skill: "Pet Care" | Tag added | — |
| 7 | Add new Language: "French (Basic)" | Language entry added | — |
| 8 | Update Marital Status: "Single" → "Married" | Dropdown updates | — |
| 9 | Replace passport document | Old removed, new uploaded | — |
| 10 | Click "Save Changes" | Form submits, loading spinner | — |
| 11 | — | Redirected to detail page with updated data | `/candidates/[id]` |

**Expected Results**:
- ✅ All changed fields updated in database
- ✅ New document replaces old (or marks old as inactive)
- ✅ Success toast notification
- ✅ Candidate detail page shows updated values
- ✅ Edit does NOT change candidate status (remains same)

**System Reactions**:
- Candidates table: Updated fields
- Documents table: Old passport document may be soft-deleted or replaced
- If candidate was already published/available: Foreign agents will see updated data on next refresh
- If CV was generated: CV becomes stale (should regenerate)

**Edit Permissions By Candidate Status**:
| Candidate Status | Editable Fields | Frozen Fields | Behaviour |
|---|---|---|---|
| `draft` | All fields, documents, metadata | None | Full edit |
| `available` | All fields, documents, metadata | None | Full edit (foreign agents see stale data until refresh; regenerate CV after edits) |
| `locked` | **Only**: Remark/notes, Language entries, Skill entries | **Frozen**: Full name, DOB, Nationality, Passport fields, Gender, Photo, Video | Edit button remains enabled but frozen fields are `disabled`/grayed in the form. A banner reads: "Candidate is under selection — some fields are locked." |
| `under_review` | Same as `locked` | Same as `locked` | Same banner |
| `approved` | **Only**: Remark/notes | **Frozen**: All profile fields, documents, languages, skills | Edit button shows warning: "Candidate is approved — only notes can be changed." API rejects non-remark field updates with `422 Unprocessable`. |
| `in_progress` | **Only**: Remark/notes | **Frozen**: All profile fields, documents | Same as `approved` |
| `completed` | **None** | All fields | Edit button hidden/disabled entirely |

**Additional Negative Scenarios**:
| Variation | Expected Behavior |
|---|---|
| Another Ethiopian agent tries to edit | `403 Forbidden` — `created_by` check enforced |
| Replace passport on a `locked` candidate | Button disabled; API returns `409 Conflict: "Cannot replace passport while candidate is locked."` |
| Replace photo on an `approved` candidate | Button disabled; API returns `422 Unprocessable: "Photo is frozen once candidate is approved."` |

---

### UC-014: Candidate Status Lifecycle — Draft to Available

**Priority**: Critical
**Type**: Functional
**Est. Time**: 8 min

**User Story**: As an Ethiopian agent, I want to publish a candidate from draft to available so foreign partners can view and select her.

**Preconditions**:
- Candidate created with complete profile (UC-008)
- Documents uploaded (passport, photo) (UC-010, UC-012)
- Active pairing exists (UC-005)

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Navigate to candidate detail page | Status: "Draft" badge visible | `/candidates/[id]` |
| 2 | Click "Publish" button | Publish dialog opens | — |
| 3 | Dialog shows partner workspaces | Lists all active pairings (Gulf Home, Dubai Elite, Kuwait Family) | — |
| 4 | Select "Publish to all partners" | All three checked | — |
| 5 | Click "Confirm Publish" | Loading spinner, API call | — |
| 6 | — | Status changes to "available" | — |
| 7 | — | Success toast: "Candidate published successfully" | — |
| 8 | Status badge now shows "Available" | Badge color changes (green/blue) | — |
| 9 | Verify in candidates list | Status column shows "Available" | `/candidates` |

**Expected Results**:
- ✅ Candidate `status` updated to `available`
- ✅ Notification sent to all paired foreign agencies
- ✅ Candidate appears in foreign agent candidate lists
- ✅ Foreign agents can view full profile and documents
- ✅ Status badge updates throughout UI

**System Reactions**:
- Candidates table: `status` → `available`
- Candidate pair shares: `candidate_pair_shares` rows created for each selected partner
- Notifications: One notification sent to each paired foreign agency (per candidate)
- Foreign agent dashboard: Candidate appears in "Select from live candidates" section
- Foreign agent candidates list: Candidate visible with "Available" badge
- Foreign agent notification bell: Badge increments, dropdown shows notification
- WebSocket: Real-time notification pushed to foreign agent if connected
- Dashboard (Ethiopian): Stats update — Total Available increments

**Status Transition Matrix**:
| From → To | Allowed? | Trigger |
|---|---|---|
| `draft` → `available` | ✅ | Publish action |
| `available` → `draft` | ✅ | Unpublish action (if no active selections) |
| `available` → `locked` | ✅ | Foreign agent creates selection |
| `draft` → `locked` | ❌ | Must be published first |
| `locked` → `under_review` | ✅ | Ethiopian agent reviews selection |
| `under_review` → `approved` | ✅ | Dual approval completes |
| `under_review` → `rejected` | ✅ | Either party rejects |
| `approved` → `in_progress` | ✅ | Post-selection tracking starts |
| `in_progress` → `completed` | ✅ | Final step (arrival) marked done |
| `available` → `deleted` | ✅ | Soft delete |
| `locked` → `available` | ✅ | Selection expires or is released |

---

### UC-015: Batch Publish Multiple Candidates

**Priority**: High
**Type**: Functional
**Est. Time**: 10 min

**User Story**: As an Ethiopian agent, I want to publish multiple candidates at once so I can save time.

**Preconditions**:
- Multiple candidates in `draft` status (Candidates B, C, D created)
- All have complete profiles and documents

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Navigate to candidates list | All candidates shown | `/candidates` |
| 2 | Click "Select" button to enter selection mode | Checkboxes appear on each candidate card | — |
| 3 | Select 3 candidates (check boxes) | Selection count banner appears at bottom | — |
| 4 | Click "Bulk Publish" from BatchActionBar | `BulkPublishDialog` opens | — |
| 5 | Select target partner(s): "All Partners" | All active pairings selected | — |
| 6 | Click "Confirm" | Processing shown for each candidate | — |
| 7 | — | Success: "3 candidates published" | — |
| 8 | Verify candidate statuses | All changed to "available" | `/candidates` |

**Expected Results**:
- ✅ Bulk publish processes all selected candidates
- ✅ Statuses updated to available
- ✅ Notifications sent for each candidate to each partner
- ✅ Partial success handled (if one fails, others still published)

**System Reactions (per published candidate)**:
- Candidates table: `status` → `available`
- Candidate pair shares: Rows created for each pairing × each candidate
- Notifications: 3 candidates × 3 partners = 9 notifications total
- Foreign agent dashboards: New candidates appear in each partner's workspace
- Dashboard (Ethiopian): Stat counts update

---

### UC-016: Partner-Specific Job Details (Pair Override)

**Priority**: High
**Type**: Functional
**Est. Time**: 10 min

**User Story**: As an Ethiopian agent, I want to set different job details (country, salary) per partner for the same candidate so each partner sees their relevant information.

**Preconditions**:
- Candidate in `available` status
- Multiple active pairings (Gulf Home SA, Dubai Elite UAE, Kuwait Family KW)

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Navigate to candidate detail page | Partner overrides section visible | `/candidates/[id]` |
| 2 | Click "Edit Override" for Gulf Home (Saudi) | Dialog opens | — |
| 3 | Set Country: "Saudi Arabia", Salary: "1500 SAR" | Fields saved | — |
| 4 | Click "Edit Override" for Dubai Elite (UAE) | Dialog opens | — |
| 5 | Set Country: "UAE", Salary: "2000 AED" | Fields saved | — |
| 6 | Click "Edit Override" for Kuwait Family | Dialog opens | — |
| 7 | Set Country: "Kuwait", Salary: "400 KWD" | Fields saved | — |
| 8 | Return to candidate detail | Each override shown per partner | — |

**Expected Results**:
- ✅ `candidate_pair_overrides` records created for each pairing
- ✅ Each foreign partner sees their specific job details when viewing candidate
- ✅ CV generation uses per-partner overrides
- ✅ Overrides editable and deletable

**System Reactions**:
- Database: Three `candidate_pair_overrides` rows created
- Foreign agent candidate detail: Shows country/salary specific to their workspace
- CV generation: Will use override values when generating per-partner CV

**Verify as Foreign Agent**:
| Step | Action | Expected Result |
|---|---|---|
| 1 | Login as Gulf Home (SA) | Dashboard loads |
| 2 | View candidate Almaz | Country: "Saudi Arabia", Salary: "1500 SAR" |
| 3 | Login as Dubai Elite (UAE) | Dashboard loads |
| 4 | View same candidate | Country: "UAE", Salary: "2000 AED" |
| 5 | Login as Kuwait Family | Dashboard loads |
| 6 | View same candidate | Country: "Kuwait", Salary: "400 KWD" |

---

## 6. Phase 3: Partner Workspace & Candidate Sharing

### UC-017: Workspace Switching via PartnerSwitcher

**Priority**: High
**Type**: Functional
**Est. Time**: 5 min

**User Story**: As an Ethiopian agent with multiple partners, I want to switch between workspaces so I can manage each partnership separately.

**Preconditions**:
- Ethiopian agent with 3+ active pairings (UC-005)
- Logged in

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Observe sidebar PartnerSwitcher | Shows active workspaces list | — |
| 2 | Current workspace: "Gulf Home Staff Recruitment" | Active pairing context = Gulf Home | — |
| 3 | Switch to "Dubai Elite Maids" in PartnerSwitcher | Workspace context changes | — |
| 4 | Navigate to candidates list | Only candidates shared with Dubai Elite shown | `/candidates` |
| 5 | Navigate to selections | Only selections involving Dubai Elite shown | `/selections` |
| 6 | Switch back to "Gulf Home Staff Recruitment" | Context reverts | — |
| 7 | Navigate to partners page | Partner workspace shows Gulf Home details | `/partners` |

**Expected Results**:
- ✅ PartnerSwitcher updates active pairing context
- ✅ All views are scoped to selected workspace
- ✅ Switching is persisted (stored in state/context)
- ✅ Dashboard adjusts metrics based on workspace context

---

### UC-018: Workspace Sharing Preferences

**Priority**: Medium
**Type**: Functional
**Est. Time**: 5 min

**User Story**: As an Ethiopian agent, I want to configure auto-sharing preferences so new candidates are automatically shared with my preferred partner.

**Preconditions**:
- Logged in as Ethiopian agent
- Active pairings

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Navigate to Partner Workspaces | Workspace page loads | `/partners` |
| 2 | Locate "Sharing Preferences" card | Auto-share toggle, default partner dropdown | — |
| 3 | Toggle "Auto-share new candidates" ON | Toggle turns on | — |
| 4 | Select Default Foreign Partner: "Gulf Home Staff Recruitment" | Dropdown updates | — |
| 5 | Click "Save Preferences" | Success toast | — |
| 6 | Navigate to create new candidate | New candidate form loads | `/candidates/new` |
| 7 | Create candidate | Candidate created | — |
| 8 | Publish candidate | Auto-shared to Gulf Home (default partner) | — |

**Expected Results**:
- ✅ `auto_share_candidates` flag updated in user profile
- ✅ `default_foreign_pairing_id` saved
- ✅ New candidates auto-shared to default partner
- ✅ Manual override still possible per candidate

---

### UC-019: Candidate Share/Unshare in Workspace

**Priority**: High
**Type**: Functional
**Est. Time**: 8 min

**User Story**: As an Ethiopian agent, I want to control which candidates are visible in each partner workspace so I can manage visibility per partnership.

**Preconditions**:
- Candidates in `available` status
- Active pairings
- Logged in as Ethiopian agent

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Navigate to candidate detail | Share section visible | `/candidates/[id]` |
| 2 | Click "Share With Partners" | `CandidateShareDialog` opens | — |
| 3 | Dialog shows all active pairings with toggle switches | Gulf Home (on), Dubai Elite (on), Kuwait Family (off) | — |
| 4 | Toggle Kuwait Family to ON | Checked | — |
| 5 | Click "Save" | Shares created | — |
| 6 | Login as Kuwait Family foreign agent | Dashboard loads | — |
| 7 | Navigate to browse candidates | Candidate now visible | `/candidates` |
| 8 | Switch back to Ethiopian agent side | — | — |
| 9 | Click "Remove from This Partner" on candidate detail | Confirmation dialog | — |
| 10 | Confirm removal | Candidate unshared from current workspace | — |
| 11 | Login as Kuwait Family again | Candidate no longer visible | — |

**Expected Results**:
- ✅ `candidate_pair_shares` records created/removed
- ✅ Foreign agents see only shared candidates
- ✅ Unshare removes visibility immediately
- ✅ Notification sent on share (optional config)

---

### UC-020: Partners Page — Workspace Overview

**Priority**: Medium
**Type**: Functional
**Est. Time**: 8 min

**User Story**: As an Ethiopian agent, I want to see a summary of each partner workspace so I can quickly understand the state of each partnership.

**Preconditions**:
- Candidates shared across multiple workspaces
- Some selections created
- Logged in as Ethiopian agent

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Navigate to Partner Workspaces | Workspace overview loads | `/partners` |
| 2 | Left panel shows partner list | All 3 partners listed | — |
| 3 | Click "Gulf Home Staff Recruitment" | Right panel updates | — |
| 4 | Hero card shows: Partner name, candidate count, selection count, tracking count | Metrics correct | — |
| 5 | Summary cards: "Visible Candidates: 2", "Selections: 1", "Tracking: 0" | Numbers match actual data | — |
| 6 | CV Defaults card shows: Country "—", Salary "—" | No defaults set yet | — |
| 7 | Candidates in workspace table lists shared candidates | Names, status, created date | — |
| 8 | Recent activity shows selections with links | Clickable selection links | — |
| 9 | Click "Edit Defaults" | `PartnerDefaultsDialog` opens | — |
| 10 | Set Country: "Saudi Arabia", Currency: "SAR", Salary: "1500" | Saved | — |
| 11 | Verify default now shows in CV Defaults card | Updated values shown | — |

**Expected Results**:
- ✅ Each workspace shows accurate, scoped data
- ✅ Partner defaults saved and reflected in CV generation
- ✅ Navigation to candidates, selections, chat works from workspace page
- ✅ Metrics update as data changes

---

## 7. Phase 4: CV Generation & Document Management

### UC-021: Generate Candidate CV

**Priority**: Critical
**Type**: Functional
**Est. Time**: 10 min

**User Story**: As an Ethiopian agent, I want to generate a professional PDF CV for a candidate so that foreign partners can review qualifications.

**Preconditions**:
- Candidate with complete profile and documents (UC-008, UC-010, UC-012)
- Partner logo uploaded (optional)
- Candidate in `available` status

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Navigate to candidate detail page | Profile displayed | `/candidates/[id]` |
| 2 | Locate CV Section | "Generate CV" button visible | — |
| 3 | Click "Generate CV" | Processing indicator, API call | — |
| 4 | Wait for generation (≤ 5 seconds) | — | — |
| 5 | — | Success: "CV generated successfully" | — |
| 6 | "Download CV" button now active | Click to download | — |
| 7 | Click "Download CV" | PDF downloads | — |

**CV Quality Verification**:
| Element | Expected |
|---|---|
| Personal Details | Full name, age, nationality, DOB, place of birth |
| Photo | Candidate photo displayed correctly |
| Passport Info | Passport number, issue/expiry dates |
| Languages | Listed with proficiency levels |
| Skills | Listed as tags/bullets |
| Experience | Years, countries, duration per country |
| Education | Level displayed |
| Agency Logo | Appears if configured (partner logo) |
| Partner Overrides | Country/salary shown if set |
| Layout | Professional, no overflow, no truncation |
| Amharic Support | Amharic text renders correctly (embedded font) |
| QR Code | QR code present (if enabled) |

**System Reactions**:
- PDF service: Generates PDF using gofpdf with embedded fonts
- S3/MinIO: CV PDF stored
- Candidates table: `cv_pdf_url` updated with S3 URL
- Candidate pair shares: `cv_pdf_url` updated per share
- Foreign agent side: CV now available for download

**Negative Scenarios**:
| Variation | Expected Behavior |
|---|---|
| Generate CV without photo | CV generates without photo (graceful) |
| Generate CV without partner logo | CV omits logo section |
| Generate CV for incomplete profile | Error: "Missing required fields" |
| Rate limit exceeded (8/min) | Error: "Too many requests" |
| Network timeout during generation | Retry option |
| Concurrent generation requests | Queued or rejected |

---

### UC-022: Per-Partner CV with Custom Overrides

**Priority**: High
**Type**: Functional
**Est. Time**: 10 min

**User Story**: As an Ethiopian agent, I want each partner to receive a CV with their specific country/salary details so candidates are presented appropriately.

**Preconditions**:
- Pair overrides set for each partner (UC-016)
- CV generated (UC-021)

**Step-by-Step Flow**:
| Step | Action | Expected Result |
|---|---|---|
| 1 | Navigate to candidate CV page | `/candidates/[id]/cv` |
| 2 | CV Download options card shows per-partner downloads | "Download CV (Gulf Home)", "Download CV (Dubai Elite)", "Download CV (Kuwait Family)" |
| 3 | Click "Download CV (Gulf Home)" | PDF downloads with Saudi Arabia, 1500 SAR |
| 4 | Open PDF and verify | Country: Saudi Arabia, Salary: 1500 SAR in CV |
| 5 | Click "Download CV (Dubai Elite)" | PDF downloads with UAE, 2000 AED |
| 6 | Open PDF and verify | Country: UAE, Salary: 2000 AED in CV |
| 7 | Click "Download CV (Kuwait Family)" | PDF downloads with Kuwait, 400 KWD |
| 8 | Open PDF and verify | Country: Kuwait, Salary: 400 KWD in CV |

**Expected Results**:
- ✅ Each partner's CV contains their specific overrides
- ✅ CV PDFs are distinct (different file names)
- ✅ `candidate_pair_shares` record stores per-partner CV URL
- ✅ Default CV (no override) also available

---

### UC-023: Bulk CV ZIP Download

**Priority**: Medium
**Type**: Functional
**Est. Time**: 8 min

**User Story**: As a foreign agent or Ethiopian agent, I want to download CVs for multiple candidates at once as a ZIP file.

**Preconditions**:
- Multiple candidates with CVs generated
- Logged in as either role

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Navigate to candidates list | Grid/table view | `/candidates` |
| 2 | Click "Select" for batch mode | Checkboxes appear | — |
| 3 | Select 2+ candidates | BatchActionBar appears | — |
| 4 | Click "Bulk CV Download" | Processing dialog appears | — |
| 5 | Wait for ZIP generation | Download starts automatically | — |
| 6 | Open ZIP file | Contains individual CV PDFs per selected candidate | — |

**Expected Results**:
- ✅ ZIP file generates with all selected candidate CVs
- ✅ File names are meaningful (e.g., `Almaz_Tadesse_CV.pdf`)
- ✅ ZIP downloads within reasonable time
- ✅ Works for both roles (Ethiopian and foreign agent)

---

### UC-024: CV Regeneration

**Priority**: Medium
**Type**: Functional
**Est. Time**: 5 min

**User Story**: As an Ethiopian agent, I want to regenerate a CV after updating candidate data so the CV reflects the latest information.

**Preconditions**:
- CV already generated
- Candidate data updated (UC-013)
- Logged in as Ethiopian agent

**Step-by-Step Flow**:
| Step | Action | Expected Result |
|---|---|---|
| 1 | Update candidate data (e.g., add new skill) | Saved successfully |
| 2 | Navigate to CV page | `/candidates/[id]/cv` |
| 3 | Click "Refresh Layout" or "Regenerate CV" | Processing |
| 4 | Wait for regeneration | New PDF generated |
| 5 | Download and verify | New skill appears in CV |
| 6 | Old CV replaced | New CV is current version |

**Expected Results**:
- ✅ CV regenerated with latest data
- ✅ Old CV URL replaced or versioned
- ✅ Success notification


---

## 8. Phase 5: Candidate Selection & Approval — Foreign Agent Deep Dive

**This phase is the most critical section of the entire platform. Every action a foreign agent takes triggers a cascade of system reactions across notifications, selection pages, candidate statuses, tracking, chat, dashboards, and admin views. Each use case in this section documents the FULL SYSTEM REACTION on both sides.**

---

### UC-025: Foreign Agent Browses Available Candidates — Full Experience

**Priority**: Critical
**Type**: Functional
**Est. Time**: 12 min

**User Story**: As a foreign agent, I want to browse available candidates shared by my Ethiopian partner so I can find suitable workers.

**Preconditions**:
- Foreign agent account active and approved (UC-004/UC-007)
- Active pairing with Ethiopian agency (UC-005)
- Ethiopian agent has published 4 candidates to all partners (UC-014, UC-015)
- Partner overrides configured (UC-016)

**Step-by-Step Flow**:
| Step | Actor | Action | Expected Result | Navigation |
|---|---|---|---|---|
| 1 | Gulf Home (SA) | Login as foreign agent | Dashboard loads | `/login` → `/dashboard/employer` |
| 2 | Gulf Home (SA) | View dashboard "Select from live candidates" grid | Shows 4 candidate cards (Almaz, Tigist, Hiwot, Muluwork) | — |
| 3 | Gulf Home (SA) | Each card shows: photo, name, age, experience, skills, "Available" badge | All fields present | — |
| 4 | Gulf Home (SA) | Click "Browse Candidates" | Full candidate list with filters | `/candidates` |
| 5 | Gulf Home (SA) | Apply filter: Skills = "Childcare" | List filters to Almaz, Hiwot (have Childcare skill) | — |
| 6 | Gulf Home (SA) | Apply filter: Age range 25-35 | Filters further | — |
| 7 | Gulf Home (SA) | Apply filter: Languages = "Arabic (Fluent)" | All 4 have Arabic, so all shown | — |
| 8 | Gulf Home (SA) | Sort by: "Most Experienced" | Order: Tigist (8yr), Almaz (5yr), Muluwork (4yr), Hiwot (3yr) | — |
| 9 | Gulf Home (SA) | Click candidate "Almaz Tadesse" | Detail page loads | `/candidates/[id]` |

**Candidate Detail Verification (Foreign Agent View)**:
| Section | Expected Content |
|---|---|
| Profile Header | Photo (clickable), Name, "Available" badge, Age 31, 5yr experience |
| Personal Details | Full Name, DOB, Nationality, Place of Birth, Passport Number, Gender, Religion, Marital Status, Children, Education |
| Languages | Arabic (Fluent), English (Intermediate), Amharic (Fluent) |
| Skills | Cooking, Cleaning, Childcare, Elderly Care |
| Partner Job Details | Country: Saudi Arabia, Salary: 1500 SAR (from override UC-016) |
| Documents | Passport (view), Photo (view), Video (play) |
| CV Section | "Download CV" or "Preview CV" button |
| Actions Card | "Select Candidate" button (prominent) |
| No Edit/Delete | No edit/delete buttons visible |

**Expected Results**:
- ✅ Foreign agent sees only shared/published candidates (scoped to workspace)
- ✅ Partner-specific overrides shown (Saudi Arabia, 1500 SAR)
- ✅ Filters and sorting work correctly
- ✅ Documents and CV viewable/downloadable
- ✅ Cannot edit or delete candidate
- ✅ "Select Candidate" action available

**System Check — Workspace Data Isolation**:
| Step | Actor | Action | Expected Result |
|---|---|---|---|
| 10 | Dubai Elite (UAE) | Login, browse candidates | Same 4 candidates visible, but with UAE partner overrides |
| 11 | Dubai Elite (UAE) | View Almaz detail | Country: UAE, Salary: 2000 AED (different from Gulf Home) |
| 12 | Kuwait Family (KW) | Login, browse candidates | Same 4 candidates visible, Kuwait overrides |
| 13 | Kuwait Family (KW) | View Almaz detail | Country: Kuwait, Salary: 400 KWD |

✅ **Data Isolation Verified**: Each foreign agent sees their workspace-scoped data with correct partner overrides.

---

### UC-026: Foreign Agent Creates Selection Request — Full System Reaction

**Priority**: Critical
**Type**: Functional
**Est. Time**: 15 min

**User Story**: As a foreign agent, I want to select a candidate so that the Ethiopian agency can begin the recruitment process.

**This is the MOST IMPORTANT use case. It triggers the most extensive system reaction in the entire platform.**

**Preconditions**:
- Candidate in `available` status
- Logged in as foreign agent (Gulf Home SA)
- No existing active selection for this candidate
- Employer contract file ready (PDF)
- Employer ID file ready (JPG/PNG)

**Step-by-Step Flow**:
| Step | Actor | Action | Expected Result | Navigation |
|---|---|---|---|---|
| 1 | Gulf Home (SA) | Navigate to candidate detail (Almaz) | "Select Candidate" button visible, enabled | `/candidates/[id]` |
| 2 | Gulf Home (SA) | Click "Select Candidate" | `SelectCandidateDialog` opens | — |
| 3 | Gulf Home (SA) | Dialog shows: Candidate name "Almaz Tadesse", Partner "Gulf Home Staff Recruitment" | Info displayed correctly | — |
| 4 | Gulf Home (SA) | Upload Employer Contract (PDF) | File uploads with progress | — |
| 5 | Gulf Home (SA) | Upload Employer ID (JPG) | File uploads with progress | — |
| 6 | Gulf Home (SA) | Enter notes: "Urgent requirement — need housemaid within 2 months" | Text area accepts | — |
| 7 | Gulf Home (SA) | Click "Submit Selection Request" | Button shows loading, API call | — |

**FULL SYSTEM REACTIONS — WHAT HAPPENS ACROSS THE ENTIRE PLATFORM**:

| # | Component | System Reaction |
|---|---|---|
| **Database** | `selections` table | New row created with: `candidate_id`, `pairing_id`, `selected_by` = Gulf Home agent ID, `status = pending`, `employer_contract_url` = S3 URL, `employer_id_url` = S3 URL, `expires_at` = now + lock duration (e.g., 24h) |
| **Database** | `candidates` table | `status` → `locked`, `locked_by` = selection ID, `locked_at` = now, `lock_expires_at` = now + duration |
| **Database** | `documents` table | Two new rows: employer contract + employer ID |
| **Database** | `notifications` table | New notification for Ethiopian agent: "Gulf Home Staff Recruitment has selected Almaz Tadesse" with `type = selection`, `related_entity_type = selection`, `related_entity_id` = selection ID |
| **Ethiopian UI** | Notification bell | Badge increments (+1 unread), dropdown shows new notification |
| **Ethiopian UI** | Dashboard Pending Actions | "Approval queue" count increments to 1 |
| **Ethiopian UI** | Selections list | New selection appears with status "Pending" |
| **Ethiopian UI** | WebSocket | Real-time notification pushed if connected |
| **Gulf Home UI** | Selection detail | Redirected to `/selections/[id]` with status "Pending" |
| **Gulf Home UI** | Dashboard | "My live selections" now shows 1 pending selection |
| **Gulf Home UI** | Candidate detail | Status now shows "Locked — Awaiting approval" |
| **Dubai Elite UI** | Candidate detail | Status now shows "Locked — Selected by another agency", "Select Candidate" disabled |
| **Dubai Elite UI** | Candidates list | Candidate shows "Locked" badge instead of "Available" |
| **Kuwait Family UI** | Candidate detail | Same locked view as Dubai Elite |
| **Admin UI** | Selections overview | New selection visible in read-only view |

| 8 | Gulf Home (SA) | — | Success toast: "Selection created successfully" | — |
| 9 | Gulf Home (SA) | — | Redirected to selection detail page | `/selections/[id]` |

**Gulf Home Selection Detail Page Verification**:
| Element | Expected Content |
|---|---|
| Profile Header | Almaz's photo, Name, Status "Pending" badge (yellow), Age, Experience |
| Selection Status | "Pending — Awaiting Ethiopian Partner Approval" |
| Timer | Shows countdown: "Lock expires in 23:59:59" |
| Employer Contract | File listed, downloadable |
| Employer ID | File listed, downloadable |
| Decision Log | Empty (no decisions yet) |
| Actions | No approve/reject (foreign agent cannot self-approve) |
| Candidate Link | "Open Candidate" → `/candidates/[candidate_id]` |
| Chat Link | "Open chat about this candidate" → `/partners/chat?candidate_id=...` |

**Ethiopian Agent Side Verification (Login as Biruk)**:
| Element | Expected Content |
|---|---|
| Dashboard Hero | "Approval queue: 1" in Pending Actions |
| Dashboard Smart Alerts | Selection pending alert |
| Notification Bell | Unread count = 1 |
| Notifications Dropdown | "Gulf Home Staff Recruitment has selected Almaz Tadesse" — click to navigate |
| Selections List | New entry: Almaz Tadesse — Gulf Home — Pending — Created [timestamp] |
| Selection Detail | Full details: Employer contract viewable, employer ID viewable, candidate info |
| Approve/Reject | Both buttons visible and enabled |

**Expected Results Summary**:
- ✅ Selection created with `status = pending`
- ✅ Employer contract and ID uploaded to S3
- ✅ Candidate locked (other foreign agents blocked)
- ✅ Lock timer started (configurable duration)
- ✅ Ethiopian agent notified in real-time
- ✅ Notification bell badge updated
- ✅ Dashboard pending actions updated
- ✅ Dubai Elite and Kuwait Family see candidate as locked
- ✅ Selection detail page renders correctly for both sides

---

### UC-027: Lock Mechanism — Foreign Agent Sees Locked State (System Reaction Deep Dive)

**Priority**: High
**Type**: Functional
**Est. Time**: 10 min

**User Story**: As a foreign agent who did NOT create the first selection, I want to see that the candidate is locked so I know not to attempt selection.

**Preconditions**:
- Gulf Home has an active pending selection on Almaz (UC-026)
- Candidate status = locked

**Step-by-Step Flow**:
| Step | Actor | Action | Expected Result | Navigation |
|---|---|---|---|---|
| 1 | Dubai Elite (UAE) | Login as foreign agent | Dashboard loads | `/login` → `/dashboard/employer` |
| 2 | Dubai Elite (UAE) | View dashboard "Select from live candidates" grid | Almaz shows "Locked" badge | — |
| 3 | Dubai Elite (UAE) | Click on Almaz | Detail page loads | `/candidates/[id]` |
| 4 | Dubai Elite (UAE) | Observe status badge | "Locked — Selected by another agency" | — |
| 5 | Dubai Elite (UAE) | "Select Candidate" button | Disabled/grayed out | — |
| 6 | Dubai Elite (UAE) | Hover over disabled button | Tooltip: "This candidate is currently locked. Selection in progress by another agency." | — |
| 7 | Dubai Elite (UAE) | Check candidates list | Almaz has "Locked" badge in list view | `/candidates` |
| 8 | Dubai Elite (UAE) | Try to access selection API directly | 400/403: "Candidate is locked" | — |

**System Reactions**:
| # | Component | State |
|---|---|---|
| Candidates | `status` | `locked` |
| Candidates | `locked_by` | Selection ID (Gulf Home's selection) |
| Candidates | `lock_expires_at` | Timestamp = now + lock duration |
| Foreign Agent UI (non-selector) | "Select Candidate" | Disabled with message |
| Foreign Agent UI (non-selector) | Status badge | "Locked — Selected by another agency" |
| API (selection create) | POST `/candidates/{id}/select` | Returns 400: "Candidate is currently locked" |

**Lock Expiry Verification**:
| Step | Action | Expected Result |
|---|---|---|
| 9 | Wait for lock to expire (or admin reduces duration) | — |
| 10 | Dubai Elite refreshes candidate page | "Select Candidate" re-enabled, status back to "Available" |
| 11 | Dubai Elite can now create selection | Selection created successfully |

---

### UC-028: Ethiopian Agent Reviews & Approves Selection — Full System Reaction

**Priority**: Critical
**Type**: Functional
**Est. Time**: 12 min

**User Story**: As an Ethiopian agent, I want to review a selection request and approve it so the recruitment process can begin.

**Preconditions**:
- Selection request exists with `status = pending` (UC-026)
- Logged in as Ethiopian agent (Biruk)

**Step-by-Step Flow**:
| Step | Actor | Action | Expected Result | Navigation |
|---|---|---|---|---|
| 1 | Biruk (ET) | Dashboard shows "Approval queue: 1" + notification | Both indicators visible | `/dashboard/agency` |
| 2 | Biruk (ET) | Click notification bell | Dropdown shows: "Gulf Home has selected Almaz Tadesse" | — |
| 3 | Biruk (ET) | Click notification | Navigates to selection detail | `/selections/[selection_id]` |
| 4 | Biruk (ET) | OR: Click "Review selections" from dashboard | Selections list | `/selections` |
| 5 | Biruk (ET) | Selection list shows: Almaz Tadesse — Gulf Home — Pending | Entry visible | — |
| 6 | Biruk (ET) | Click selection | Detail page loads | `/selections/[selection_id]` |

**Selection Detail Review**:
| Section | Expected Content |
|---|---|
| Candidate Profile | Almaz's photo, name, age, experience |
| Selection Status | "Pending — Waiting for your decision" |
| Employer Contract | Downloadable file (Gulf Home's contract) |
| Employer ID | Downloadable file (Gulf Home's ID) |
| Decision Log | Empty (no decisions yet) |
| Partner Info | Gulf Home Staff Recruitment (Saudi Arabia) |
| Partner Override | Country: Saudi Arabia, Salary: 1500 SAR |
| Approve Button | Green/Primary: "Approve Selection" |
| Reject Button | Red/Danger: "Reject / Unlock Candidate" |

| 7 | Biruk (ET) | Review employer contract — click to view/download | Document opens/downloads | — |
|---|---|---|---|---|
| 8 | Biruk (ET) | Review employer ID — click to view | ID document shown | — |
| 9 | Biruk (ET) | Click "Approve Selection" | Confirmation dialog appears | — |
| 10 | Biruk (ET) | Dialog text: "Approve this selection? This will lock the candidate and start recruitment tracking." | Informational warning | — |
| 11 | Biruk (ET) | Click "Confirm" | Loading state, API call | — |

**FULL SYSTEM REACTION ON APPROVAL**:

| # | Component | System Reaction |
|---|---|---|
| **Database** | `selections` table | `status` → `approved` |
| **Database** | `approvals` table | New row: `selection_id`, `user_id` = Biruk, `decision` = approved, timestamp |
| **Database** | `candidates` table | `status` → `approved` (permanent lock) |
| **Database** | `selection_progress` table | New row created with all steps at "pending" |
| **Database** | `notifications` table | New notification for Gulf Home: "Your selection of Almaz Tadesse has been approved by Addis Talent Solutions" with `type = approval` |
| **Database** | `notifications` table | New notification for Dubai Elite (if they had pending): "Almaz Tadesse has been approved by another agency" (auto-rejection) |
| **Database** | `notifications` table | New notification for Kuwait Family (same auto-rejection) |
| **Gulf Home UI** | Notification bell | Badge increments, new notification |
| **Gulf Home UI** | Dashboard | "My live selections" shows "Approved" status |
| **Gulf Home UI** | Selection detail | Status → "Approved" (green), "Recruitment Tracking" section appears with progress bar at 0% |
| **Gulf Home UI** | WebSocket | Real-time notification pushed |
| **Gulf Home UI** | Candidate detail | Status → "Approved" (permanent — cannot be selected) |
| **Biruk (ET) UI** | Selection detail | Status → "Approved" (green), "Recruitment Tracking" section now visible |
| **Biruk (ET) UI** | Dashboard | Approval queue decrements, In Progress count increments |
| **Biruk (ET) UI** | Notification bell | (May receive confirmation notification) |
| **Dubai Elite UI** | Candidate detail | Status → "Approved — Recruitment in progress" (permanently locked) |
| **Dubai Elite UI** | Notifications | "Almaz Tadesse has been approved — no longer available" |
| **Kuwait Family UI** | Candidate detail | Same as Dubai Elite |
| **Admin UI** | Selections overview | Selection status updated to "Approved" |
| **Admin UI** | Dashboard metrics | Success rate may update |
| **All** | Decision log | Shows: "Approved by Biruk Alemu (Addis Talent Solutions) — [timestamp]" |

| 12 | Biruk (ET) | — | Success toast: "Selection approved successfully" | — |
|---|---|---|---|---|
| 13 | Biruk (ET) | — | Selection detail shows "Approved" status, tracking section visible | `/selections/[id]` |

**Expected Results**:
- ✅ Selection status → `approved`
- ✅ Approval recorded in approvals table
- ✅ Candidate permanently locked (status → `approved`)
- ✅ Selection progress record created (all steps pending)
- ✅ Gulf Home notified in real-time
- ✅ Other foreign agents auto-rejected and notified
- ✅ Recruitment tracking section now visible to both parties
- ✅ Dashboard stats updated

---

### UC-029: Dual Approval Flow (Both Parties Must Approve)

**Priority**: Medium
**Type**: Functional
**Est. Time**: 12 min

**User Story**: When the platform requires both parties to approve, I want both the Ethiopian and foreign agents to approve before the selection is finalized.

**Preconditions**:
- Admin has set "Require both parties to approve" = ON in platform settings (UC-044)
- Selection request created by foreign agent (UC-026)
- Selection is in `pending` status

**Full Dual Approval Flow**:
| Step | Actor | Action | Expected Result |
|---|---|---|---|
| 1 | Admin | Enable "Require both parties to approve" in settings | Setting saved |
| 2 | Gulf Home (FA) | Create selection (as in UC-026) | Status: "pending" |
| 3 | Biruk (ET) | Navigate to selection detail | Both "Approve" and "Reject" visible |
| 4 | Biruk (ET) | Click "Approve" | Confirmation dialog |
| 5 | Biruk (ET) | Confirm | Status still "pending" (NOT yet approved) |
| 6 | — | Decision log shows: "Approved by Biruk Alemu (Addis Talent Solutions) — Awaiting second approval" | First approval recorded |
| 7 | Biruk (ET) | Selection detail page | Status shows: "Pending — Awaiting Foreign Partner Approval", progress shows 1/2 approvals |
| 8 | Gulf Home (FA) | Navigate to selection detail | Status: "Pending", Decision log shows Ethiopian approved |
| 9 | Gulf Home (FA) | "Approve" button now visible (needs second approval) | Button enabled |
| 10 | Gulf Home (FA) | Click "Approve" | Confirmation dialog |
| 11 | Gulf Home (FA) | Confirm | Status → "approved" |
| 12 | — | Decision log: "Approved by Khalid Al-Rashid (Gulf Home Staff Recruitment)" | Second approval recorded |
| 13 | — | Full system reaction triggers (same as UC-028 approval) | Candidate locked, tracking starts, notifications sent |

**System Reactions**:
- Selections: Status remains `pending` until both approvals received
- Approvals table: Two rows when complete (one per party)
- Decision log: Shows each approval with timestamp and approver details
- Notifications: Ethiopian notified when foreign agent gives second approval
- Foreign agent sees: "Awaiting your approval" after Ethiopian approves

**Rejection During Dual Approval**:
| Step | Actor | Action | Expected Result |
|---|---|---|---|
| 14 | Biruk (ET) | (On a different selection) Click "Reject" | Selection → "rejected" immediately |
| 15 | — | Candidate unlocked, both parties notified | Even if only one party rejected |

---

### UC-030: Foreign Agent Approves (Dual Approval Second Step)

**Priority**: Medium
**Type**: Functional
**Est. Time**: 8 min

**User Story**: As a foreign agent in a dual-approval setup, I want to give my final approval so the selection is complete.

**Preconditions**:
- "Require both parties to approve" = ON
- Ethiopian agent already approved (UC-028 variation)
- Logged in as Gulf Home (foreign agent)

**Step-by-Step Flow**:
| Step | Actor | Action | Expected Result | Navigation |
|---|---|---|---|---|
| 1 | Gulf Home (FA) | Notification received: "Addis Talent Solutions has approved your selection of Almaz Tadesse — awaiting your approval" | Click notification → selection detail | `/selections/[id]` |
| 2 | Gulf Home (FA) | Selection detail shows: "Pending — Awaiting Your Approval" | Status banner, progress 1/2 | — |
| 3 | Gulf Home (FA) | Decision log shows Ethiopian approval | Biruk's approval with timestamp | — |
| 4 | Gulf Home (FA) | "Approve" button visible | Enabled | — |
| 5 | Gulf Home (FA) | Click "Approve" | Confirm dialog | — |
| 6 | Gulf Home (FA) | Confirm | Status → "approved" | — |

**System Reactions**:
- Same as UC-028 approval cascade (notifications, tracking enabled, candidate locked)
- Additional: Decision log shows both approvals
- Ethiopian agent notified: "Your selection has been fully approved"

---

### UC-031: Ethiopian Agent Rejects Selection

**Priority**: High
**Type**: Functional
**Est. Time**: 8 min

**User Story**: As an Ethiopian agent, I want to reject a selection request if I disagree with the terms or have concerns.

**Preconditions**:
- Selection request exists with `status = pending` (UC-026)
- Logged in as Ethiopian agent

**Step-by-Step Flow**:
| Step | Actor | Action | Expected Result |
|---|---|---|---|
| 1 | Biruk (ET) | Navigate to selection detail | Pending selection displayed | `/selections/[id]` |
| 2 | Biruk (ET) | Click "Unlock Candidate" or "Reject" | Rejection dialog opens | — |
| 3 | Biruk (ET) | Enter reason: "Salary too low for candidate's experience level" | Text area (required) | — |
| 4 | Biruk (ET) | Click "Confirm Rejection" | Processing | — |
| 5 | — | Selection status → "rejected" | — |
| 6 | — | Candidate status → "available" (unlocked) | — |

**System Reactions**:
| # | Component | Reaction |
|---|---|---|
| Selections | `status` | `rejected` |
| Approvals | New row | `decision = rejected`, `reason` saved |
| Candidates | `status` | `available` (unlocked) |
| Notifications (Gulf Home) | New notification | "Your selection of Almaz Tadesse has been rejected. Reason: Salary too low" |
| Notifications (Dubai Elite) | New notification | "Almaz Tadesse is now available" |
| Gulf Home UI | Selection detail | Shows "Rejected" status, reason visible |
| Dubai Elite UI | Candidate detail | "Select Candidate" re-enabled |
| Ethiopian dashboard | Approval queue | Decrements |

| 7 | Biruk (ET) | — | Success toast | — |
|---|---|---|---|---|
| 8 | Gulf Home (FA) | Check selections | Status "Rejected", reason visible | `/selections/[id]` |
| 9 | Dubai Elite (FA) | Check candidate | "Select Candidate" now available | `/candidates/[id]` |

**Expected Results**:
- ✅ Selection → `rejected`
- ✅ Approval record with decision and reason
- ✅ Candidate unlocked → `available`
- ✅ Rejection reason visible to both parties
- ✅ Gulf Home notified with reason
- ✅ Dubai Elite notified of availability
- ✅ Candidate can be re-selected

---

### UC-032: Foreign Agent Experiences Rejection — Full System Reaction

**Priority**: High
**Type**: Functional
**Est. Time**: 8 min

**User Story**: As a foreign agent whose selection was rejected, I want to see why it was rejected and what my options are.

**Preconditions**:
- Gulf Home's selection was rejected (UC-031)

**Step-by-Step Flow**:
| Step | Actor | Action | Expected Result | Navigation |
|---|---|---|---|---|
| 1 | Gulf Home (FA) | Login | Dashboard loads | `/login` → `/dashboard/employer` |
| 2 | Gulf Home (FA) | Notification bell shows unread | Badge count = 1+ | — |
| 3 | Gulf Home (FA) | Click notification bell | Dropdown shows rejection notification | — |
| 4 | Gulf Home (FA) | Click notification | Navigates to selection detail | `/selections/[id]` |
| 5 | Gulf Home (FA) | Status badge | "Rejected" (red) | — |
| 6 | Gulf Home (FA) | Rejection reason displayed | "Salary too low for candidate's experience level" | — |
| 7 | Gulf Home (FA) | Decision log | Shows: "Rejected by Biruk Alemu (Addis Talent Solutions) — [timestamp]" | — |
| 8 | Gulf Home (FA) | Candidate link | Click "Open Candidate" | `/candidates/[id]` |
| 9 | Gulf Home (FA) | Candidate status | "Available" (no longer locked) | — |
| 10 | Gulf Home (FA) | "Select Candidate" button | Re-enabled (can try again with better offer) | — |

**Expected Results**:
- ✅ Rejection visible with reason
- ✅ Decision log shows who rejected and when
- ✅ Candidate is available for re-selection
- ✅ Notification persists until marked read

---

### UC-033: Selection Auto-Expiry — Full System Reaction

**Priority**: High
**Type**: Functional
**Est. Time**: 15 min (simulate or wait)

**User Story**: I want pending selections to auto-expire after the configured duration so stale selections don't block candidates.

**Preconditions**:
- Platform setting: "Auto-expire pending selections" = ON
- Selection lock duration: Set to 10 minutes (for testing) or 24h (production)
- Pending selection exists (UC-026)

**Step-by-Step Flow**:
| Step | Actor | Action | Expected Result |
|---|---|---|---|
| 1 | Admin | Set lock duration to 10 minutes | `/admin/settings` |
| 2 | Gulf Home (FA) | Create selection on Almaz (UC-026) | Selection pending, lock timer: 09:59... |
| 3 | Biruk (ET) | View selection detail | Timer shows countdown |
| 4 | Dubai Elite (FA) | View candidate | "Locked — Selection in progress (expires in 09:55)" |
| 5 | — | Wait 10 minutes (or trigger expiry worker) | — |
| 6 | — | Selection → "expired" automatically | — |
| 7 | — | Candidate → "available" | — |

**System Reactions on Expiry**:
| # | Component | Reaction |
|---|---|---|
| Selections | `status` | `expired` |
| Candidates | `status` | `available`, `locked_by` → null, `lock_expires_at` → null |
| Notifications (Gulf Home) | "Your selection of Almaz Tadesse has expired" | `type = expiry` |
| Notifications (Biruk) | "Selection by Gulf Home for Almaz Tadesse has expired" | `type = expiry` |
| Notifications (Dubai Elite) | "Almaz Tadesse is now available for selection" | `type = status_update` |
| Gulf Home UI | Selection detail | "Expired" badge, message: "Selection period ended" |
| Dubai Elite UI | Candidate detail | "Select Candidate" re-enabled |
| Dashboard alerts | Smart alerts | Expiry notifications shown |

| 8 | Gulf Home (FA) | View selection | Status: "Expired", message shown | `/selections/[id]` |
|---|---|---|---|---|
| 9 | Dubai Elite (FA) | View candidate | Can now select | `/candidates/[id]` |

**Expected Results**:
- ✅ Expiry worker fires (cron job or via API process)
- ✅ Selections beyond duration auto-expire
- ✅ Warning notification sent before expiry (configurable threshold)
- ✅ Candidate released back to available pool
- ✅ Both parties notified
- ✅ Other foreign agents notified of availability

---

### UC-034: Foreign Agent Re-Selects After Rejection/Expiry

**Priority**: High
**Type**: Functional
**Est. Time**: 10 min

**User Story**: As a foreign agent, I want to re-select a candidate after my previous selection was rejected or expired, possibly with better terms.

**Preconditions**:
- Previous selection for Almaz was rejected (UC-031) or expired (UC-033)
- Candidate is now `available`
- Logged in as Gulf Home (FA)

**Step-by-Step Flow**:
| Step | Actor | Action | Expected Result |
|---|---|---|---|
| 1 | Gulf Home (FA) | Navigate to candidate detail | "Select Candidate" re-enabled | `/candidates/[id]` |
| 2 | Gulf Home (FA) | Click "Select Candidate" | Dialog opens | — |
| 3 | Gulf Home (FA) | Upload new/updated employer contract | Can be same or different | — |
| 4 | Gulf Home (FA) | Enter notes: "Revised offer — increased salary to 1800 SAR" | Shows improvement | — |
| 5 | Gulf Home (FA) | Submit | New selection created | — |
| 6 | — | Full system reaction triggers (same as UC-026) | Candidate locked, Ethiopian notified | — |
| 7 | Biruk (ET) | Review new selection | Can see notes about improved offer | `/selections/[id]` |
| 8 | Biruk (ET) | Approve (if terms acceptable) | Process continues normally | — |

**Expected Results**:
- ✅ Re-selection after rejection/expiry works normally
- ✅ New selection gets fresh lock timer
- ✅ Previous selection history preserved (audit trail)
- ✅ Ethiopian agent can see rejection history and new offer

---

### UC-035: Foreign Agent Uploads/Replaces Employer Documents

**Priority**: Medium
**Type**: Functional
**Est. Time**: 5 min

**User Story**: As a foreign agent, I want to upload or replace my employer contract and ID documents so the Ethiopian agent has accurate information.

**Preconditions**:
- Selection in `pending` status
- Logged in as Gulf Home (FA)

**Step-by-Step Flow**:
| Step | Actor | Action | Expected Result |
|---|---|---|---|
| 1 | Gulf Home (FA) | Navigate to selection detail | Employer documents section visible | `/selections/[id]` |
| 2 | Gulf Home (FA) | Click "Replace Contract" | File picker opens | — |
| 3 | Gulf Home (FA) | Select new contract PDF | Uploads with progress | — |
| 4 | Gulf Home (FA) | — | Old contract replaced, new one available | — |
| 5 | Gulf Home (FA) | Click "Replace Employer ID" | File picker opens | — |
| 6 | Gulf Home (FA) | Select new ID image | Uploads | — |
| 7 | Gulf Home (FA) | — | Old ID replaced, new one available | — |
| 8 | Biruk (ET) | View selection detail | New documents visible, can download | `/selections/[id]` |

**Expected Results**:
- ✅ Documents replaced successfully
- ✅ Old documents may be soft-deleted or versioned
- ✅ Ethiopian agent sees updated documents
- ✅ Notification may or may not be sent (check behavior)

---

### UC-036: Selection Unlock / Release by Ethiopian Agent

**Priority**: Medium
**Type**: Functional
**Est. Time**: 8 min

**User Story**: As an Ethiopian agent, I want to unlock a selection (release candidate) if the foreign agent is not proceeding.

**Preconditions**:
- Selection in `pending` status
- Logged in as Ethiopian agent

**Step-by-Step Flow**:
| Step | Actor | Action | Expected Result |
|---|---|---|---|
| 1 | Biruk (ET) | Navigate to selection detail | Pending selection | `/selections/[id]` |
| 2 | Biruk (ET) | Click "Unlock Candidate" | Confirmation dialog | — |
| 3 | Biruk (ET) | Enter reason: "Foreign agent not responding to communications" | Text area | — |
| 4 | Biruk (ET) | Confirm unlock | Processing | — |
| 5 | — | Selection status → "released" | — |
| 6 | — | Candidate status → "available" | — |

**System Reactions**:
| # | Component | Reaction |
|---|---|---|
| Selections | `status` | `released` |
| Candidates | `status` | `available` |
| Notifications (Gulf Home) | "Your selection of Almaz Tadesse has been released" | `type = rejection` with reason |
| Notifications (other FAs) | "Almaz Tadesse is now available" | `type = status_update` |
| Gulf Home UI | Selection detail | "Released" badge, reason visible |

| 7 | Gulf Home (FA) | View selection | "Released" status, reason visible |
|---|---|---|---|
| 8 | Dubai Elite (FA) | View candidate | "Select Candidate" available again |

---

### UC-037: Foreign Agent Views Selection History & Past Selections

**Priority**: Low
**Type**: Functional
**Est. Time**: 5 min

**User Story**: As a foreign agent, I want to see my complete selection history so I can track past activity.

**Preconditions**:
- Multiple selections in various statuses (pending, approved, rejected, expired)
- Logged in as Gulf Home (FA)

**Step-by-Step Flow**:
| Step | Actor | Action | Expected Result | Navigation |
|---|---|---|---|---|
| 1 | Gulf Home (FA) | Navigate to My Selections | List of all selections | `/selections` |
| 2 | Gulf Home (FA) | Tab: "Active" | Shows pending selections | — |
| 3 | Gulf Home (FA) | Tab: "Approved" | Shows approved selections | — |
| 4 | Gulf Home (FA) | Tab: "Rejected" | Shows rejected selections | — |
| 5 | Gulf Home (FA) | Tab: "Expired" | Shows expired selections | — |
| 6 | Gulf Home (FA) | Tab: "All" | Shows every selection across all statuses | — |
| 7 | Gulf Home (FA) | Sort by: Newest | Order correct | — |
| 8 | Gulf Home (FA) | Pagination | 25 per page, navigate | — |
| 9 | Gulf Home (FA) | Click a past rejected selection | Detail shows current status, past decisions | `/selections/[id]` |

**Expected Results**:
- ✅ All selection status tabs work correctly
- ✅ Sort and pagination function
- ✅ Past selection details preserved with full history
- ✅ Decision log shows complete approval/rejection history


---

## 9. Phase 6: Post-Selection Process Tracking

### UC-038: Ethiopian Agent Updates Selection Progress — Medical Step

**Priority**: Critical
**Type**: Functional
**Est. Time**: 10 min

**User Story**: As an Ethiopian agent, I want to update the medical examination step so the foreign partner can see progress.

**Preconditions**:
- Selection approved (UC-028)
- Logged in as Ethiopian agent

**Step-by-Step Flow**:
| Step | Actor | Action | Expected Result | Navigation |
|---|---|---|---|---|
| 1 | Biruk (ET) | Navigate to selection detail | Status timeline visible | `/selections/[id]` |
| 2 | Biruk (ET) | Progress section shows steps: Medical, CoC, Visa, Ticket, Arrival | All steps show "Pending" (gray), progress bar 0% | — |
| 3 | Biruk (ET) | Click "Update" on Medical step | Form opens (inline or modal) | — |
| 4 | Biruk (ET) | Set status: "Completed" | Dropdown/select | — |
| 5 | Biruk (ET) | Upload medical document (PDF) | File upload with progress | — |
| 6 | Biruk (ET) | Enter notes: "Medical exam passed at Addis General Hospital. All clear." | Text area | — |
| 7 | Biruk (ET) | Click "Save" | Processing, API call | — |

**Full System Reaction on Medical Update**:
| # | Component | Reaction |
|---|---|---|
| `selection_progress` | `medical_status` | `completed` |
| `selection_progress` | `medical_document_url` | S3 URL of uploaded document |
| `selection_progress` | `notes` | "Medical exam passed at Addis General Hospital" |
| Notifications (Gulf Home) | New notification | "Medical step completed for Almaz Tadesse" with `type = status_update` |
| Gulf Home UI | Selection detail | Medical step shows green "Completed" checkmark |
| Gulf Home UI | Progress bar | Advances from 0% to 20% (1/5 steps) |
| Gulf Home UI | Notification bell | Badge increments |
| Gulf Home UI | WebSocket | Real-time push if connected |
| Dashboard (Foreign) | Smart Alerts | May show progress update |
| Dashboard (Ethiopian) | Stats | In Progress tracking updates |

| 8 | Biruk (ET) | — | Success toast: "Medical step updated" | — |
|---|---|---|---|---|
| 9 | Gulf Home (FA) | View selection detail (may be simultaneous) | Medical: "Completed", document viewable | `/selections/[id]` |
| 10 | Gulf Home (FA) | Click medical document link | Downloads/opens medical document | — |

**Expected Results**:
- ✅ Medical step → `completed` with green checkmark
- ✅ Medical document stored and accessible
- ✅ Progress bar advances
- ✅ Gulf Home notified in real-time
- ✅ Notes visible to both parties

---

### UC-039: Update Certificate of Clearance (CoC) Step

**Priority**: High
**Type**: Functional
**Est. Time**: 8 min

**User Story**: As an Ethiopian agent, I want to update the Certificate of Clearance step so the partner knows the candidate has passed background checks.

**Preconditions**:
- Selection approved
- Medical step completed (UC-038)
- Logged in as Ethiopian agent

**Step-by-Step Flow**:
| Step | Actor | Action | Expected Result |
|---|---|---|---|
| 1 | Biruk (ET) | Navigate to selection detail | CoC step shows "Pending" | `/selections/[id]` |
| 2 | Biruk (ET) | Click "Update" on CoC step | Form opens with CoC type selector | — |
| 3 | Biruk (ET) | Select CoC type: "Online" | Radio/select | — |
| 4 | Biruk (ET) | Set status: "Completed" | Dropdown | — |
| 5 | Biruk (ET) | Enter notes: "Online background check completed, Certificate #12345" | Text area | — |
| 6 | Biruk (ET) | Click "Save" | Processing | — |
| 7 | — | CoC → "Completed" (green checkmark) | Progress: 2/5 = 40% | — |

**System Reactions**:
- `selection_progress`: `coc_status` → `completed`, `coc_type` → `online`
- Gulf Home notified with `status_update` notification
- Gulf Home sees CoC completed with type and notes
- Progress bar advances to 40%

---

### UC-040: Update Visa & Ticket Steps

**Priority**: High
**Type**: Functional
**Est. Time**: 8 min

**User Story**: As an Ethiopian agent, I want to update visa processing and ticket booking so the foreign partner can track the candidate's departure readiness.

**Preconditions**:
- Medical and CoC completed (40% progress)
- Logged in as Ethiopian agent

**Step-by-Step Flow**:
| Step | Actor | Action | Expected Result |
|---|---|---|---|
| 1 | Biruk (ET) | Navigate to selection detail | Visa step shows "Pending" | `/selections/[id]` |
| 2 | Biruk (ET) | Click "Update" on Visa step | Form opens | — |
| 3 | Biruk (ET) | Status: "In Progress" | Dropdown | — |
| 4 | Biruk (ET) | Notes: "Visa application submitted to Saudi embassy on 2026-07-20" | Text area | — |
| 5 | Biruk (ET) | Save | Visa → "In Progress" (yellow) | — |
| 6 | Biruk (ET) | Later, update Visa: "Completed" | Dropdown | — |
| 7 | Biruk (ET) | Upload visa document (if applicable) | File upload | — |
| 8 | Biruk (ET) | Save | Visa → "Completed" (green) | — |
| 9 | Biruk (ET) | Update Ticket step: "Completed" | Dropdown | — |
| 10 | Biruk (ET) | Notes: "Flight ET-450 on 2026-08-15 Addis to Riyadh" | Text area | — |
| 11 | Biruk (ET) | Save | Ticket → "Completed" (green) | — |
| 12 | — | Progress: 4/5 = 80% | Bar shows 80% | — |

**System Reactions per Step Update**:
- Gulf Home receives `status_update` notification each time
- Progress bar advances incrementally
- Documents uploaded are accessible to foreign agent
- Notes visible to both parties

---

### UC-041: Update Arrival Step — Complete the Process

**Priority**: High
**Type**: Functional
**Est. Time**: 8 min

**User Story**: As an Ethiopian agent or foreign agent, I want to record the candidate's arrival so the recruitment process is complete.

**Preconditions**:
- All previous steps completed (80% progress)
- Logged in as Ethiopian agent (or foreign agent, depending on who handles arrival)

**Step-by-Step Flow**:
| Step | Actor | Action | Expected Result |
|---|---|---|---|
| 1 | Biruk (ET) | Navigate to selection detail | Arrival step shows "Pending" | `/selections/[id]` |
| 2 | Biruk (ET) | Click "Update" on Arrival step | Form opens | — |
| 3 | Biruk (ET) | Status: "Completed" | Dropdown | — |
| 4 | Biruk (ET) | Arrival Date: 2026-08-15 | Date picker | — |
| 5 | Biruk (ET) | Arrival City: "Riyadh" | Text field | — |
| 6 | Biruk (ET) | Destination Country: "Saudi Arabia" | Text field | — |
| 7 | Biruk (ET) | Departure Date: 2026-08-14 | Date picker | — |
| 8 | Biruk (ET) | Notes: "Candidate arrived safely, picked up by employer at airport" | Text area | — |
| 9 | Biruk (ET) | Click "Save" | Processing | — |

**Full System Reaction on Arrival (Process Complete)**:
| # | Component | Reaction |
|---|---|---|
| `selection_progress` | `arrival_status` | `completed`, all arrival fields saved |
| Candidate | `status` | `completed` (lifecycle finished) |
| Notifications (Gulf Home) | "Almaz Tadesse has arrived in Saudi Arabia" | `type = arrived` |
| Notifications (Biruk) | "Arrival recorded for Almaz Tadesse" | `type = status_update` |
| Dashboard (Foreign) | Approved profiles | Count updates |
| Dashboard (Ethiopian) | In Progress stats | Decrements, Completed increments |
| All | Progress bar | 5/5 = 100%, "Complete" shown |

| 10 | — | Progress → 100%, Candidate → "Completed" | Lifecycle finished |

**Expected Results**:
- ✅ All arrival fields saved correctly
- ✅ Progress shows 100%
- ✅ Candidate status → `completed`
- ✅ Notification sent: "Candidate arrived in destination country"
- ✅ Audit log entry for completion
- ✅ Both parties see completed status

---

### UC-042: Foreign Agent Views Progress Timeline

**Priority**: Medium
**Type**: Functional
**Est. Time**: 5 min

**User Story**: As a foreign agent, I want to view the progress tracking for my selected candidates so I know the status of recruitment.

**Preconditions**:
- Selection approved with progress updates (UC-038 through UC-041)
- Logged in as Gulf Home (FA)

**Step-by-Step Flow**:
| Step | Actor | Action | Expected Result | Navigation |
|---|---|---|---|---|
| 1 | Gulf Home (FA) | Navigate to selections list | Selection shows status "Approved — In Progress" | `/selections` |
| 2 | Gulf Home (FA) | Click on approved selection | Detail page loads | `/selections/[id]` |
| 3 | Gulf Home (FA) | Scroll to "Recruitment Tracking" | Progress bar shows e.g. 80% | — |
| 4 | Gulf Home (FA) | Timeline: Medical (Completed), CoC (Completed), Visa (Completed), Ticket (Completed), Arrival (Pending) | Each step status visible | — |
| 5 | Gulf Home (FA) | Click "Medical" to expand | Shows: Status, Notes, Document link, Date | — |
| 6 | Gulf Home (FA) | Click medical document link | Opens/downloads document | — |
| 7 | Gulf Home (FA) | Click "Ticket" to expand | Notes: "Flight ET-450 on 2026-08-15 Addis to Riyadh" | — |
| 8 | Gulf Home (FA) | All steps are read-only | No edit buttons visible | — |

**Expected Results**:
- ✅ Foreign agent sees accurate, real-time progress
- ✅ All documents downloadable
- ✅ Notes and dates visible
- ✅ Progress percentage correct
- ✅ No edit capability (read-only view)

---

### UC-043: Batch Progress Update

**Priority**: Medium
**Type**: Functional
**Est. Time**: 8 min

**User Story**: As an Ethiopian agent, I want to batch update progress for multiple selections at once.

**Preconditions**:
- Multiple selections in progress
- Logged in as Ethiopian agent

**Step-by-Step Flow**:
| Step | Actor | Action | Expected Result | Navigation |
|---|---|---|---|---|
| 1 | Biruk (ET) | Navigate to Process Tracking hub | List of in-progress selections | `/tracking` |
| 2 | Biruk (ET) | Select multiple selections | Checkboxes | — |
| 3 | Biruk (ET) | Choose step to update: "Medical" | Dropdown | — |
| 4 | Biruk (ET) | Set status: "Completed" | Selected | — |
| 5 | Biruk (ET) | Submit batch update | Processing for each selection | — |
| 6 | — | Each selection's medical step updated | — |
| 7 | Biruk (ET) | Verify individual selections | Each shows updated status | `/selections/[id]` |

**Expected Results**:
- ✅ Batch progress endpoint processes all selections
- ✅ Partial success handled gracefully
- ✅ Notifications sent for each updated selection

---

### UC-044: Foreign Agent Views Tracking Hub

**Priority**: Low
**Type**: Functional
**Est. Time**: 5 min

**User Story**: As a foreign agent, I want to view a centralized tracking page showing all my in-progress selections.

**Preconditions**:
- Multiple selections in various tracking stages
- Logged in as Gulf Home (FA)

**Step-by-Step Flow**:
| Step | Actor | Action | Expected Result | Navigation |
|---|---|---|---|---|
| 1 | Gulf Home (FA) | Navigate to Process Tracking | Hub loads | `/tracking` |
| 2 | Gulf Home (FA) | Shows all in-progress selections | List with candidate names, progress % | — |
| 3 | Gulf Home (FA) | Click a selection | Navigates to selection detail | `/selections/[id]` |
| 4 | Gulf Home (FA) | All data is read-only | No edit capability | — |

**Expected Results**:
- ✅ Tracking hub shows all selections with progress
- ✅ Links navigate to selection detail
- ✅ Read-only access enforced

---

## 10. Phase 7: Communication & Notifications

### UC-045: Real-Time Chat Between Workspace Partners

**Priority**: Medium
**Type**: Functional
**Est. Time**: 10 min

**User Story**: As an Ethiopian agent, I want to chat with my foreign partner in real-time so we can coordinate on candidates and selections.

**Preconditions**:
- Active pairing exists
- Both agents logged in (two browsers/devices)
- Chat WebSocket connected

**Step-by-Step Flow**:
| Step | Actor | Action | Expected Result | Navigation |
|---|---|---|---|---|
| 1 | Biruk (ET) | Navigate to Chat | Chat page loads | `/partners/chat` |
| 2 | — | Workspace thread auto-created | "Gulf Home Staff Recruitment" thread in left panel | — |
| 3 | Biruk (ET) | Click workspace thread | Chat area opens, messages load | — |
| 4 | Biruk (ET) | Type: "Hello! We have 3 new candidates ready for your review" | Text appears in composer | — |
| 5 | Biruk (ET) | Press Enter (or click Send) | Message sent instantly | — |
| 6 | Gulf Home (FA) | Open Chat (already on page) | Message appears in real-time (< 2 sec) | — |
| 7 | Gulf Home (FA) | Reply: "Perfect! I'll review them today. Can you share their CVs?" | Message sent | — |
| 8 | Biruk (ET) | Message appears in real-time | Real-time delivery confirmed | — |
| 9 | Both | Verify messages show: sender name, timestamp, company | All metadata correct | — |

**Expected Results**:
- ✅ Real-time message delivery via WebSocket
- ✅ Messages persisted in `chat_messages` table
- ✅ Timestamps displayed correctly
- ✅ Sender info (name, company) shown
- ✅ Unread badges update
- ✅ Messages grouped by day
- ✅ Scroll to latest message on send
- ✅ Connection status indicator (Live/Reconnecting/Offline)

---

### UC-046: Candidate-Scoped Chat Thread (from Selection Detail)

**Priority**: Medium
**Type**: Functional
**Est. Time**: 5 min

**User Story**: As either agent, I want to open a chat about a specific candidate so discussions are organized by context.

**Preconditions**:
- Selection exists for a candidate
- Both agents logged in

**Step-by-Step Flow**:
| Step | Actor | Action | Expected Result | Navigation |
|---|---|---|---|---|
| 1 | Biruk (ET) | Navigate to selection detail | Detail page loads | `/selections/[id]` |
| 2 | Biruk (ET) | Click "Open chat about this candidate" | Navigates to chat with candidate context | `/partners/chat?candidate_id=...` |
| 3 | — | Candidate-specific thread auto-resolved/created | Shown in left panel with candidate name | — |
| 4 | Biruk (ET) | Type: "Almaz's medical exam is scheduled for next week" | Message sent | — |
| 5 | Gulf Home (FA) | Opens chat | Sees both workspace and candidate threads | — |
| 6 | Gulf Home (FA) | Clicks candidate thread | Messages visible, replies accordingly | — |

**Expected Results**:
- ✅ Candidate-specific thread created (`scope_type = candidate`)
- ✅ Both parties see the thread
- ✅ Messages scoped to candidate discussion
- ✅ Navigation from selection detail to chat is seamless

---

### UC-047: All Notification Types — Complete Verification

**Priority**: Medium
**Type**: Functional
**Est. Time**: 20 min

**User Story**: As both agents, I want to receive notifications for important events so I don't miss updates.

**Trigger All Notification Types and Verify System Reactions**:
| # | Trigger Action | Notification Type | Recipient | Visible In |
|---|---|---|---|---|
| 1 | Ethiopian agent publishes candidate | `candidate_published` | Foreign Agent | Bell, Notifications page, WebSocket |
| 2 | Foreign agent creates selection | `selection` | Ethiopian Agent | Bell, Notifications page, WebSocket |
| 3 | Ethiopian agent approves selection | `approval` | Foreign Agent | Bell, Notifications page, WebSocket |
| 4 | Ethiopian agent rejects selection | `rejection` | Foreign Agent | Bell, Notifications page, WebSocket |
| 5 | Foreign agent approves (dual) | `approval` | Ethiopian Agent | Bell, Notifications page, WebSocket |
| 6 | Ethiopian agent updates tracking step | `status_update` | Foreign Agent | Bell, Notifications page, WebSocket |
| 7 | Ethiopian agent marks arrival | `arrived` | Foreign Agent | Bell, Notifications page, WebSocket |
| 8 | Selection expires | `expiry` | Both | Bell, Notifications page, WebSocket |
| 9 | Passport expiry warning | `passport_expiry_warning` | Ethiopian Agent | Bell, Dashboard Smart Alerts |
| 10 | Medical document expiry | `expiry_warning` | Ethiopian Agent | Bell, Dashboard Smart Alerts |
| 11 | Flight booked (arrival step) | `flight_booked` | Foreign Agent | Bell, Notifications page |
| 12 | New chat message | (chat unread count) | Other party | Bell badge, sidebar badge |

**Step-by-Step Verification**:
| Step | Action | Expected Result |
|---|---|---|
| 1 | Perform each trigger action from table | Each generates a notification |
| 2 | Observe notification bell on header | Badge count increases with each unread |
| 3 | Click bell icon | Dropdown shows last 5 notifications |
| 4 | Each notification shows: icon, title, message, time ago | All fields correct |
| 5 | Unread notifications have blue background + blue dot | Visual distinction |
| 6 | Click "View All Notifications" | Full notification center loads | `/notifications` |
| 7 | Tabs: All / Unread / Selection / Approval / Status Updates | Each tab filters correctly |
| 8 | Click a notification | Navigates to related entity |
| 9 | Verify navigation destinations | Correct entity per type |
| 10 | Click "Mark All as Read" | All notifications marked read, badge cleared |
| 11 | Navigate back to notification center | All show as read | `/notifications` |
| 12 | Logout and login again | Notifications persist, read status persists |

**Expected Results**:
- ✅ All notification types trigger correctly
- ✅ Real-time delivery via WebSocket
- ✅ Bell badge reflects unread count (max "9+" for 10+)
- ✅ Click navigates to correct entity
- ✅ Mark as read works individually and bulk
- ✅ Notifications persist across sessions
- ✅ Types are distinguishable visually

---

### UC-048: WebSocket Connection & Reconnection

**Priority**: Medium
**Type**: Functional
**Est. Time**: 8 min

**User Story**: I want WebSocket connections to be resilient so I don't miss real-time updates during network interruptions.

**Step-by-Step Flow**:
| Step | Actor | Action | Expected Result |
|---|---|---|---|
| 1 | Biruk (ET) | Login | WebSocket connects, status: "Live" |
| 2 | Biruk (ET) | Navigate to chat | Connection indicator shows green |
| 3 | Biruk (ET) | Disconnect network (WiFi off) | Status: "Offline" or "Reconnecting..." |
| 4 | Gulf Home (FA) | Create selection on a candidate (while ET disconnected) | Notification queued server-side |
| 5 | Biruk (ET) | Reconnect network | WebSocket auto-reconnects |
| 6 | Biruk (ET) | — | Pending notifications delivered |
| 7 | Biruk (ET) | Chat messages sent during offline | Received after reconnect |
| 8 | Gulf Home (FA) | Navigate to selections page | Falls back to 5-min polling if WS disconnected |

**Expected Results**:
- ✅ WebSocket connects on login
- ✅ Auto-reconnection attempts on disconnect
- ✅ Connection status indicator visible
- ✅ Messages/notifications queued and delivered on reconnect
- ✅ Graceful fallback to polling



---

## 11. Phase 8: Admin Oversight & Governance

### UC-049: Admin Dashboard Analytics

**Priority**: Medium
**Type**: Functional
**Est. Time**: 8 min

**User Story**: As a super admin, I want to see platform-wide analytics so I can monitor overall operations.

**Preconditions**:
- Super admin logged in
- Multiple agencies, candidates, selections exist from previous use cases

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Navigate to admin dashboard | Analytics loaded | `/admin/dashboard` |
| 2 | Hero card: "Platform Overview" | Pending queue count > 0, candidate states, selection stream | â€” |
| 3 | Metric cards: Total Agencies, Pending Approvals, Total Candidates, Success Rate | Numbers match actual data | â€” |
| 4 | Pending approval queue | Lists pending agencies with "Review" links | â€” |
| 5 | Candidate status distribution | Bar chart with counts per status | â€” |
| 6 | Agency registrations trend | 7-day mini bar chart | â€” |
| 7 | Selections trend | 7-day mini bar chart | â€” |
| 8 | Top Ethiopian agencies (by candidate count) | Ranked list (Addis Talent should be #1) | â€” |
| 9 | Top foreign agencies (by selection count) | Ranked list (Gulf Home should be #1) | â€” |

**Expected Results**:
- âœ… All metrics accurate and up-to-date
- âœ… Charts render correctly
- âœ… Links navigate to relevant admin pages
- âœ… Data scoped to entire platform (not workspace-specific)

---

### UC-050: Admin Agency Management â€” Suspend/Activate

**Priority**: High
**Type**: Functional
**Est. Time**: 8 min

**User Story**: As an admin, I want to manage agency accounts by suspending or activating them.

**Preconditions**:
- Admin logged in
- Active agencies exist

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Navigate to Agencies list | All agencies displayed | `/admin/agencies` |
| 2 | Filter by: Ethiopian, Active | Filtered list | â€” |
| 3 | Click "View" on "Addis Talent Solutions PLC" | Agency detail loads | `/admin/agencies/[id]` |
| 4 | View agency profile: company, contact, email, status, role | All fields correct | â€” |
| 5 | Account controls section: Status dropdown | Current: "Active" | â€” |
| 6 | Select "Suspended" | Reason textarea appears | â€” |
| 7 | Enter reason: "Failed to respond to compliance inquiries" | Text area | â€” |
| 8 | Click "Apply Status Change" | Confirmation dialog | â€” |
| 9 | Confirm | Status changes to "Suspended" | â€” |
| 10 | â€” | Success toast, audit log entry | â€” |
| 11 | Biruk (ET) attempts login | Error: "Account suspended. Contact admin." | â€” |
| 12 | Return to admin, change status back to "Active" | Agency re-activated | â€” |
| 13 | Biruk (ET) can login again | Access restored | â€” |

**System Reactions**:
- Users table: `account_status` â†’ `suspended`
- Active sessions: User sessions may be revoked
- Foreign agent partners: Selections in progress affected? (check behavior)
- Audit logs: Entry created with admin ID, action, reason

---

### UC-051: Admin Read-Only Oversight â€” Candidates & Selections

**Priority**: Low
**Type**: Functional
**Est. Time**: 5 min

**User Story**: As an admin, I want read-only access to candidates and selections so I can audit without interfering.

**Preconditions**:
- Admin logged in
- Data exists

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Navigate to Admin Candidates | Table with all candidates (across all agencies) | `/admin/candidates` |
| 2 | Verify: No edit/delete buttons | Read-only confirmed | â€” |
| 3 | Filter by status, search by name | Filters work | â€” |
| 4 | Pagination works (20 per page) | Navigate pages | â€” |
| 5 | Navigate to Admin Selections | Table with all selections (across all pairings) | `/admin/selections` |
| 6 | Filter by status, search | Works | â€” |

**Expected Results**:
- âœ… Admin sees all candidates/selections across all agencies
- âœ… No edit/delete capability
- âœ… Filters and pagination work

---

### UC-052: Admin Audit Logs

**Priority**: Medium
**Type**: Functional
**Est. Time**: 8 min

**User Story**: As an admin, I want to view audit logs so I can track all admin actions for accountability.

**Preconditions**:
- Admin logged in
- Admin actions performed (approvals, status changes, etc.)

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Navigate to Audit Logs | "Operator Activity" tab selected | `/admin/audit-logs` |
| 2 | Metric cards: Matched entries, Admin logins, Agency actions, Recent activity | Counts displayed | â€” |
| 3 | Table shows: Admin name, Action, Target type, Target ID, IP, Timestamp | Each entry visible | â€” |
| 4 | Filter by action type dropdown | Table filters | â€” |
| 5 | Filter by target type dropdown | Further filtered | â€” |
| 6 | Switch to "Agency Sign-ins" tab | Table: Agency, Role, Device, IP, Login time, Last seen, Session status | â€” |
| 7 | Metric cards: Total login events, Active sessions, Ethiopian logins, Foreign logins | Counts correct | â€” |

---

### UC-053: Admin Platform Settings Management

**Priority**: High
**Type**: Functional
**Est. Time**: 10 min

**User Story**: As a super admin, I want to configure platform settings so the recruitment process follows our rules.

**Preconditions**:
- Super admin logged in

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Navigate to Settings | Settings page loads | `/admin/settings` |
| 2 | Section: "Recruitment Rules" | â€” | â€” |
| 3 | Change "Selection lock duration" to 48 hours | Dropdown updates | â€” |
| 4 | Toggle "Require both parties to approve" ON | Toggle updates | â€” |
| 5 | Toggle "Auto-expire pending selections" ON | Toggle updates | â€” |
| 6 | Section: "Agency Onboarding" | â€” | â€” |
| 7 | Toggle "Auto-approve agencies" ON | Toggle updates | â€” |
| 8 | Section: "Maintenance mode" | â€” | â€” |
| 9 | Enable maintenance mode, enter message: "System upgrade in progress" | Fields accept input | â€” |
| 10 | Click "Save changes" | All settings saved | â€” |
| 11 | â€” | Success toast, settings persisted | â€” |
| 12 | Logout, try to login as Biruk (ET) | Maintenance page with message shown | â€” |
| 13 | Login as admin, disable maintenance mode | Normal access restored | â€” |
| 14 | Revert selection lock to 24h, toggle both approvals OFF | Settings updated and saved | â€” |

**Expected Results**:
- âœ… All settings persisted to `platform_settings` table
- âœ… Maintenance mode blocks non-admin users
- âœ… Settings cached with 5-min TTL
- âœ… Only `super_admin` can access this page
- âœ… Support admin sees restriction notice

---

### UC-054: Admin Account Management (Super Admin Only)

**Priority**: Medium
**Type**: Functional
**Est. Time**: 8 min

**User Story**: As a super admin, I want to manage other admin accounts so I can control platform operator access.

**Preconditions**:
- Super admin logged in

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Navigate to Admin Management | Admin list displayed | `/admin/admins` |
| 2 | Table shows: Name, Role, Status, Last Login, Created, Actions | Current admin(s) listed | â€” |
| 3 | Click "Add Admin" | Dialog opens | â€” |
| 4 | Enter Full Name: "Test Support", Email: "support@maidrecruitment.com", Role: "Support Admin" | Form fills | â€” |
| 5 | Click "Create" | Admin account created | â€” |
| 6 | â€” | Setup URL displayed: "admin/setup?token=..." | â€” |
| 7 | Verify in admin list | New admin appears | â€” |
| 8 | Click "Suspend" on the new admin | Confirmation dialog | â€” |
| 9 | Confirm | Status changes to "Suspended" | â€” |
| 10 | Attempt login as support admin | Blocked: account suspended | â€” |
| 11 | Click "Activate" | Status restored | â€” |
| 12 | Support admin can login | Access restored | â€” |

---

## 12. Phase 9: Edge Cases & Error Scenarios

### UC-055: Concurrent Operations & Race Conditions

**Priority**: High
**Type**: Functional/Concurrency
**Est. Time**: 15 min

**Test concurrent scenarios that could cause data conflicts:**

| # | Scenario | Steps | Expected Behavior |
|---|---|---|---|
| 1 | Two foreign agents select same candidate simultaneously | Both click "Select" at nearly same time | **Lock mechanism**: The backend uses a **PostgreSQL row-level lock (`SELECT ... FOR UPDATE`) on the `candidates` row** inside a serializable transaction. The first request to commit acquires the lock and transitions the candidate to `locked`. The second request's transaction detects the row has already changed and rolls back. The losing request receives `HTTP 409 Conflict` with body `{"error":"candidate_already_locked","message":"This candidate is already being selected by another agency."}`. The UI shows: "Candidate is no longer available — another agency selected them just now." |
| 2 | Ethiopian agent approves while foreign agent rejects simultaneously (dual approval) | Both actions at same time | First action processed, second gets conflict error |
| 3 | Ethiopian agent edits candidate while foreign agent views | Edit saves while view is open | View shows stale data until refresh; no corruption |
| 4 | Two admins approve the same agency | Both click approve | First succeeds, second gets "Agency already approved" |
| 5 | Ethiopian agent publishes while admin suspends agency | Both actions | Suspension takes precedence; publish may fail |
| 6 | Ethiopian agent updates tracking while foreign agent reads | Simultaneous read/write | No corruption; read sees old data until refresh |
| 7 | Both agents send chat message at exactly same time | Both press Enter | Both messages delivered, order may vary |

---

### UC-056: Data Isolation & Role-Based Access Control

**Priority**: Critical
**Type**: Security
**Est. Time**: 15 min

**Verify that users can only access their own data:**

| # | Scenario | Steps | Expected Behavior |
|---|---|---|---|
| 1 | Ethiopian agent A tries to view candidate of Ethiopian agent B | Direct URL access | 403 Forbidden or 404 |
| 2 | Foreign agent tries to create a candidate | Access create page or API | Button hidden, API returns 403 |
| 3 | Ethiopian agent tries to create a selection | Access selection endpoint | 403 Forbidden |
| 4 | Unauthenticated user accesses any protected page | Direct URL | Redirect to login |
| 5 | Agent accesses another pairing's data via API | Tamper X-Pairing-ID header | 403 or data scoped to their pairings |
| 6 | Foreign agent accesses unpublished candidate | Direct URL | 404 or "Not available" |
| 7 | Agent accesses admin routes | Direct URL | Redirect to admin login or 403 |
| 8 | Support admin accesses platform settings | Direct URL | 403 Forbidden (restriction notice) |
| 9 | Foreign agent A views selections belonging to foreign agent B | IDOR attempt | 403 or 404 |
| 10 | Ethiopian agent views selections from a different pairing | Workspace switch | Only sees own pairing's selections |

---

### UC-057: Input Validation & Error Handling

**Priority**: High
**Type**: Functional/Negative
**Est. Time**: 12 min

**Test form validation across the platform:**

| # | Form | Invalid Input | Expected Error |
|---|---|---|---|
| 1 | Registration | Invalid email "notanemail" | "Please enter a valid email" |
| 2 | Registration | Password "123" (too short) | "Password must be at least 8 characters" |
| 3 | Registration | Password "password123", Confirm "different" | "Passwords do not match" |
| 4 | Registration | Empty Full Name | "Full name is required" |
| 5 | Login | Wrong password | "Invalid email or password" |
| 6 | Login | Empty fields | "Email is required" / "Password is required" |
| 7 | Candidate Create | Empty Full Name | Field-level validation error |
| 8 | Candidate Create | Future date of birth | "Date of birth cannot be in the future" |
| 9 | Candidate Create | Passport expiry before issue | "Expiry must be after issue date" |
| 10 | Candidate Create | Negative age | Validation error |
| 11 | File Upload | File > 10MB (photo) | "File too large. Maximum size: 10MB" |
| 12 | File Upload | File > 50MB (video) | "File too large. Maximum size: 50MB" |
| 13 | File Upload | .exe file for passport | "Invalid file type. Accepted: PDF, JPG, PNG" |
| 14 | Password Change | Wrong current password | "Current password is incorrect" |
| 15 | Password Change | New password same as current | "New password must differ from current" |
| 16 | Forgot Password | Non-existent email | "If this email exists, a reset code has been sent" (no enumeration) |
| 17 | Selection Create | No contract uploaded | "Employer contract is required" |
| 18 | Selection Create | No employer ID uploaded | "Employer ID is required" |
| 19 | Progress Update | Invalid status value | Validation error |
| 20 | Chat Message | Empty message | Send button disabled or validation |

---

### UC-058: File Upload Edge Cases

**Priority**: High
**Type**: Functional/Negative
**Est. Time**: 10 min

| # | Scenario | Steps | Expected Behavior |
|---|---|---|---|
| 1 | Upload corrupt image file | Select corrupt JPG | Upload fails, error message |
| 2 | Upload very large file (> 100MB) | Attempt video upload > 100MB | Rejected by client-side or server-side limit |
| 3 | Upload during network interruption | Start upload, disconnect | Error toast, retry option |
| 4 | Upload same file twice | Upload passport, upload same file again | Duplicate detection or overwrite |
| 5 | Upload file with special characters in name | "candidato's passport (2024).pdf" | File saved with sanitized name |
| 6 | Upload very long filename | 200+ character filename | Truncated or accepted |
| 7 | Upload no file (empty selection) | Click upload, cancel in file picker | No action, no error |
| 8 | Rapid consecutive uploads | Upload 5 files quickly | Processed in sequence or parallel |

---

### UC-059: Session & Authentication Edge Cases

**Priority**: Critical
**Type**: Security
**Est. Time**: 12 min

| # | Scenario | Steps | Expected Behavior |
|---|---|---|---|
| 1 | Expired JWT token | Wait 24h, make request | 401, redirected to login |
| 2 | Revoked session | Logout from settings, use old token | 401, "Session revoked" |
| 3 | Concurrent sessions | Login from two browsers | Both active, listed in active sessions |
| 4 | Logout all devices | Use "Logout from all devices" | All sessions revoked |
| 5 | Remember me | Check "Remember me", close browser, reopen | Still logged in (cookie persists) |
| 6 | Session after password change | Change password, old session | Old session remains or gets invalidated (check behavior) |
| 7 | Rate limit exceeded on login | Attempt login 8+ times with wrong password | "Too many attempts. Try again in 15 minutes." |
| 8 | Account lockout | Exceed failed attempts | Account locked, admin must unlock |

---

### UC-060: Data Deletion & Recovery

**Priority**: High
**Type**: Functional
**Est. Time**: 10 min

| # | Scenario | Steps | Expected Behavior |
|---|---|---|---|
| 1 | Soft delete candidate | Delete candidate from detail page | Candidate hidden from lists, soft-deleted in DB |
| 2 | View deleted candidate directly | Access `/candidates/[deleted_id]` | 404 or "Not found" |
| 3 | Foreign agent sees deleted candidate | Candidate deleted while foreign agent views | May show "Not available" on next load |
| 4 | Delete candidate with active selection | Attempt delete while selection exists | **Blocked**: Backend returns `409 Conflict` — "Cannot delete candidate with an active selection. End the selection first." |
| 5 | Delete document from candidate | Remove passport document | Document removed from S3 and DB |
| 6 | Admin ends (suspends) a pairing | Admin sets pairing status to `ended` | **Historic data preserved — read-only for both agents**: Existing selections remain visible in both agents' selection lists but show an informational banner: "This workspace is no longer active. Data preserved for reference." Chat threads become read-only (no new messages can be sent). Tracking records are viewable but not updatable. In-progress tracking steps are frozen at their current state. If admin later resumes the pairing (`active`), full functionality is restored and edit capability resumes. |
| 7 | Admin suspends an agency | Admin sets agency `account_status` to `suspended` | **All pairings involving this agency are immediately set to `suspended`**. Historic selections, tracking records, and chat messages remain in the database but are **inaccessible to both the suspended agency and their partners** (the UI for the affected pairing shows: "Partner agency is no longer active — contact platform admin"). Audit trail preserved. Admin can reinstate (`active`) to restore access. |
| 8 | Admin deletes an agency account (soft-delete / reject) | Admin rejects or soft-deletes agency | **Data archived**: All candidates, selections, tracking records, pairings, and chats belonging to this agency are soft-deleted (or orphaned with a `deleted_at` timestamp). The partner agency's UI hides all data from this pairing as if the pairing never existed. Admin audit logs retain the full history. This action is **irreversible** (no restore path). |

---

### UC-061: Performance & Load Testing (Quick Check)

**Priority**: Medium
**Type**: Performance
**Est. Time**: 10 min

| # | Page/Operation | Expected Time | Measure |
|---|---|---|---|
| 1 | Login page load (cache cleared) | < 2 seconds | ____ sec |
| 2 | Ethiopian dashboard load | < 3 seconds | ____ sec |
| 3 | Candidates list (50+ candidates) | < 4 seconds | ____ sec |
| 4 | Candidate detail page | < 2 seconds | ____ sec |
| 5 | Candidate creation form load | < 3 seconds | ____ sec |
| 6 | CV generation | < 5 seconds | ____ sec |
| 7 | Selection list load (25+ selections) | < 3 seconds | ____ sec |
| 8 | Chat page load with history | < 3 seconds | ____ sec |
| 9 | Partners page load | < 3 seconds | ____ sec |
| 10 | Admin dashboard load | < 4 seconds | ____ sec |

---

### UC-073: Mid-Flow Workspace Unpairing

**Priority**: High
**Type**: Functional / Edge Case
**Est. Time**: 15 min

**User Story**: When an admin unpairs an Ethiopian and Foreign agent while a selection is mid-flow (locked, under_review, or in_progress), the system must preserve existing work while preventing new actions.

**Preconditions**:
- 1 Ethiopian agent (Biruk), 1 Foreign agent (Gulf Home SA)
- Pairing active
- 1 candidate (Almaz) with a selection in `under_review` status
- 1 different candidate (Hana) with a selection in `in_progress` status
- 1 different candidate (Selam) with a selection that is `draft` (not yet shared)

**Scenario Flow**:
| Step | Actor | Action | Expected Result |
|---|---|---|---|
| 1 | Admin | Open pairing management page | Both pairings visible with "End Workspace" button |
| 2 | Admin | Click "End Workspace" for Biruk ↔ Gulf Home | Confirmation dialog: "This will end the workspace. In-progress data will be preserved but locked. Are you sure?" |
| 3 | Admin | Confirm unpairing | Pairing status changes to `ended`. System timestamp recorded. |
| 4 | Biruk (Ethiopian) | View Partners page | Gulf Home SA still visible but shows banner: "This workspace is no longer active." "Share Candidate" button disabled. |
| 5 | Gulf Home (Foreign) | View candidate library | All candidates still listed but with icon: "Partner workspace ended — read only" |
| 6 | Gulf Home (Foreign) | Open Almaz's `under_review` selection | Full read-only view. Approve/Reject buttons disabled. Banner: "This workspace ended on 2026-07-08. No further actions permitted." |
| 7 | Gulf Home (Foreign) | Open Hana's `in_progress` tracking | Tracking data visible but all action buttons disabled. Status frozen. |
| 8 | Biruk (Ethiopian) | Try to edit Selam's profile | Profile opens but "Share with Gulf Home" option removed from share dialog (partner not in active list). Can still edit and share with other active partners. |
| 9 | Gulf Home (Foreign) | Try to chat | Chat opens in read-only mode. Input field disabled. Banner: "This conversation is archived." |
| 10 | Admin | Re-activate the pairing | Admin sets pairing status back to `active`. After re-activation, full functionality is restored for all existing data. The `under_review` selection now shows Approve/Reject buttons again. Tracking can resume. Chat input re-enabled. |

**Expected Results**:
- ✅ All pre-existing data preserved read-only post-unpairing
- ✅ No new actions (select, approve, track, chat) permitted while `ended`
- ✅ Clear, user-facing banners indicate workspace status at all relevant views
- ✅ Full restoration on re-activation — no data loss
- ✅ Other active pairings for each agency unaffected

**Negative Scenarios**:
| Variation | Expected Behavior |
|---|---|
| Admin unpairs while a selection is `locked` | Selection stays `locked`, Ethiopian cannot approve/reject, Foreign sees lock message with workspace-ended banner. When re-activated, lock timer may have expired — candidate returns to `available`. |
| Admin unpairs while WebSocket notification is mid-flight | Real-time message may be dropped. Next page refresh shows correct state. System should log the event. |
| Foreign agent is actively filling a new selection form when pairing ends | Form submission fails with `403 Forbidden — "Your workspace with this agency is no longer active"`. Unsaved form data may be lost (browser warns). |

---

## 13. Phase 10: Cross-Cutting Concerns

### UC-062: Mobile Responsiveness

**Priority**: Medium
**Type**: UI/UX
**Est. Time**: 20 min

**Test all major pages on mobile viewports (375px - 428px width):**

| # | Page | Mobile Elements to Verify |
|---|---|---|
| 1 | `/` Landing | Stacked layout, CTA buttons tappable, no horizontal scroll |
| 2 | `/login` | Form fills viewport, keyboard doesn't obscure fields |
| 3 | `/dashboard/agency` | Stat cards stack, sidebar hidden (hamburger), bottom nav visible |
| 4 | `/candidates` | Grid to single column, filters collapsible, batch bar responsive |
| 5 | `/candidates/new` | Long form scrolls, date pickers usable, file upload touch-friendly |
| 6 | `/candidates/[id]` | Two-column to single column, all sections readable |
| 7 | `/selections` | Tabs tappable, list items touch-friendly |
| 8 | `/selections/[id]` | Timeline responsive, action buttons full-width |
| 9 | `/partners/chat` | Left panel (threads) slides in/out, composer accessible |
| 10 | `/settings` | Tabs tappable, forms scrollable |
| 11 | `/notifications` | List items tappable, tabs scrollable |
| 12 | Admin pages | Tables to card view on mobile, filters accessible |

**Mobile Navigation Elements**:
| Element | Expected |
|---|---|
| Bottom Nav | 5 tabs: Home, Candidates, Selections, Chat, More |
| Mobile Nav (hamburger) | Slide-in sheet with all nav links |
| Header | Sticky, notification bell accessible |
| PartnerSwitcher | Accessible from mobile nav |
| Notification Bell | Badge visible, dropdown usable |

---

### UC-063: Dark Mode / Theme

**Priority**: Low
**Type**: UI
**Est. Time**: 8 min

| # | Step | Expected Result |
|---|---|---|
| 1 | Toggle theme from "Light" to "Dark" in settings | All pages render in dark mode |
| 2 | Verify sidebar | Dark theme colors correct (slate-950) |
| 3 | Verify forms | Input fields, labels, buttons visible |
| 4 | Verify tables | Alternating row colors visible |
| 5 | Verify status badges | Colors distinguishable in dark mode |
| 6 | Verify cards, dialogs, modals | All UI elements visible |
| 7 | Verify charts (admin dashboard) | Colors work in dark mode |
| 8 | Toggle to "System" | Follows OS preference |
| 9 | Verify persistence | Theme persists after logout/login |

---

### UC-064: Waiting State â€” No Active Pairings

**Priority**: Medium
**Type**: Functional
**Est. Time**: 5 min

**User Story**: As an Ethiopian agent without any pairings, I want to see a waiting page so I understand my account is active but not yet connected.

**Preconditions**:
- Ethiopian agent with approved account
- No active pairings assigned by admin

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Login as Ethiopian agent with no pairings | Redirected to waiting page | `/waiting` (not dashboard) |
| 2 | Verify PageHeader: "Waiting for Partner Assignment" | Title displayed | â€” |
| 3 | Hero card: Explains private workspaces | Informational text | â€” |
| 4 | Status tiles: "Account approved" (done), "Workspace pending" (current), "Next step" | Visual progress indicator | â€” |
| 5 | Sidebar shows only "Workspace Status" link | Limited navigation | â€” |
| 6 | Admin creates pairing (UC-005) | â€” | â€” |
| 7 | Refresh page or wait for auto-redirect | Redirected to dashboard | `/dashboard/agency` |
| 8 | Full navigation becomes available | All sidebar links active | â€” |

**Expected Results**:
- âœ… Waiting page displays correct status
- âœ… No access to other modules until paired
- âœ… After pairing created, access restored
- âœ… Auto-redirect or refresh detects pairing status change

---

### UC-065: Platform Maintenance Mode

**Priority**: Medium
**Type**: Functional
**Est. Time**: 8 min

**User Story**: As an admin, I want to enable maintenance mode so users cannot access the platform during updates.

**Preconditions**:
- Super admin logged in

**Step-by-Step Flow**:
| Step | Actor | Action | Expected Result |
|---|---|---|---|
| 1 | Super Admin | Navigate to Settings, Maintenance | `/admin/settings` |
| 2 | Super Admin | Enable maintenance mode, message: "System upgrade in progress. Expected downtime: 2 hours" | Saved |
| 3 | Ethiopian Agent | Try to access any page | Maintenance page with message shown |
| 4 | Foreign Agent | Try to login | Maintenance page shown |
| 5 | Super Admin | Try to access admin pages | Admin portal works normally |
| 6 | Super Admin | Disable maintenance mode | Normal access restored |
| 7 | Ethiopian Agent | Refresh and login | Access restored |

**Expected Results**:
- âœ… Non-admin users blocked during maintenance
- âœ… Custom maintenance message displayed
- âœ… Admin users unaffected
- âœ… Setting persists (survives restart)
- âœ… API middleware blocks non-admin requests

---

### UC-066: Password Reset Flow

**Priority**: High
**Type**: Functional
**Est. Time**: 10 min

**User Story**: As an Ethiopian agent who forgot my password, I want to reset it so I can regain access.

**Preconditions**:
- Registered user exists
- Not logged in

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Navigate to login page | Login form | `/login` |
| 2 | Click "Forgot password?" | Reset request form | `/forgot-password` |
| 3 | Enter email: "addis@talent.et" | Field accepts | â€” |
| 4 | Click "Send reset code" | Processing... | â€” |
| 5 | â€” | Success panel: "Reset code sent" | â€” |
| 6 | â€” | Link appears: "Go to reset page" | `/reset-password?email=addis@talent.et` |
| 7 | Click the link (or navigate) | Reset form loads with email pre-filled | â€” |
| 8 | Enter 6-digit code from email/console | Code field | â€” |
| 9 | Enter New Password: "NewPass@789" | Password with strength indicator | â€” |
| 10 | Enter Confirm Password: "NewPass@789" | Matches | â€” |
| 11 | Click "Reset Password" | Processing... | â€” |
| 12 | â€” | Success: "Password reset successfully" | â€” |
| 13 | â€” | Redirect to login | `/login` |
| 14 | Login with new password | Success | `/dashboard/agency` |

**Expected Results**:
- âœ… Password reset code generated (6-digit, 15-min expiry)
- âœ… Code sent via email
- âœ… Password updated successfully
- âœ… Old password no longer works
- âœ… Rate limited (5 requests / 10 min for forgot, 10/10 min for reset)

**Negative Scenarios**:
| Variation | Expected Behavior |
|---|---|
| Wrong reset code | "Invalid code" (5 attempts then expires) |
| Expired code (> 15 min) | "Code expired, request new one" |
| Weak new password | Validation error |
| Cooldown (resend < 60s) | "Please wait before requesting another code" |

---

### UC-067: Email Verification Resend & Cooldown

**Priority**: Medium
**Type**: Functional
**Est. Time**: 8 min

**User Story**: As a registered but unverified user, I want to resend the verification email if I didn't receive it.

**Preconditions**:
- User registered but not verified

**Step-by-Step Flow**:
| Step | Action | Expected Result |
|---|---|---|
| 1 | Navigate to pending/verify page | Status: "Verification email sent" |
| 2 | Click "Resend verification email" | Processing... |
| 3 | â€” | "Verification email resent" |
| 4 | Click "Resend" again within 60 seconds | Cooldown: "Please wait before requesting again" |
| 5 | Wait 60+ seconds, resend | Works again |
| 6 | Attempt 6th resend within 10 minutes | "Too many requests. Try again later." (5/10min limit) |

---

### UC-068: Partner Logo Upload & CV Branding

**Priority**: Low
**Type**: Functional
**Est. Time**: 5 min

**User Story**: As an Ethiopian agent, I want to upload my agency's logo so it appears on generated CVs.

**Preconditions**:
- Logged in as Ethiopian agent

**Step-by-Step Flow**:
| Step | Action | Expected Result |
|---|---|---|
| 1 | Navigate to Partner Workspaces | Workspace page | `/partners` |
| 2 | Select a workspace (e.g., Gulf Home) | Workspace details | â€” |
| 3 | Click "Edit Defaults" | `PartnerDefaultsDialog` opens | â€” |
| 4 | Look for logo upload option | Upload button | â€” |
| 5 | Upload logo (PNG with transparent background) | Upload progress, preview | â€” |
| 6 | Click "Save" | Logo saved | â€” |
| 7 | Navigate to candidate CV | Generate new CV | `/candidates/[id]/cv` |
| 8 | Download CV | PDF shows agency logo | â€” |

**Expected Results**:
- âœ… Logo uploads to S3
- âœ… `partner_logo_url` saved in pairing
- âœ… CV generation includes logo
- âœ… Logo appears in correct position on CV

---

### UC-069: Account Settings â€” Profile & Security

**Priority**: Medium
**Type**: Functional
**Est. Time**: 10 min

**User Story**: As an Ethiopian agent, I want to manage my account settings so my profile is up to date.

**Step-by-Step Flow**:
| Step | Action | Expected Result | Navigation |
|---|---|---|---|
| 1 | Navigate to Settings | Settings page with 3 tabs | `/settings` |
| 2 | Tab: "Profile" | Profile form loads | â€” |
| 3 | Update Full Name: "Biruk Alemu" to "Biruk H. Alemu" | Field updates | â€” |
| 4 | Update Company Name: "Addis Talent Solutions PLC" | Field updates | â€” |
| 5 | Upload Avatar/Logo | File picker, upload | â€” |
| 6 | Click "Save" | Success toast | â€” |
| 7 | Tab: "Security" | Password form, sessions list | â€” |
| 8 | Change password: Current "Test@1234", New "NewPass@789", Confirm "NewPass@789" | All fields filled | â€” |
| 9 | Click "Change Password" | Success: "Password changed" | â€” |
| 10 | View Active Sessions | Lists current session(s) | â€” |
| 11 | Click "Logout from all other devices" | Other sessions revoked | â€” |
| 12 | Tab: "Preferences" | Toggle switches | â€” |
| 13 | Toggle notification preferences | Saved | â€” |

**Expected Results**:
- âœ… Profile updated successfully
- âœ… Avatar/logo upload works
- âœ… Password changed, old password no longer works
- âœ… Sessions listed and revocable
- âœ… Preferences saved and persisted

---

### UC-070: Multi-Partner Selection Competition (Complete Scenario)

**Priority**: High
**Type**: Integration
**Est. Time**: 20 min

**User Story**: When multiple foreign partners want the same candidate, the system should manage this fairly with locks, approvals, and auto-rejections.

**Preconditions**:
- 1 Ethiopian agent (Addis Talent)
- 3 Foreign agents (Gulf Home SA, Dubai Elite UAE, Kuwait Family KW)
- 1 available candidate (Almaz Tadesse)
- All three pairings active

**Scenario Flow**:
| Step | Actor | Action | Expected Result |
|---|---|---|---|
| 1 | Gulf Home (SA) | Create selection for Almaz | Candidate locked, pending |
| 2 | Dubai Elite (UAE) | View candidate | "Locked â€” Selection in progress by another agency" |
| 3 | Kuwait Family (KW) | View candidate | Same locked message |
| 4 | Ethiopian Agent | Review Gulf Home's selection | Contract + ID visible |
| 5 | Ethiopian Agent | Reject Gulf Home's selection | Selection "rejected", candidate "available" |
| 6 | Dubai Elite (UAE) | Refresh candidate page | "Select Candidate" re-enabled |
| 7 | Dubai Elite (UAE) | Create selection immediately | Candidate locked again |
| 8 | Kuwait Family (KW) | View candidate | Locked again |
| 9 | Ethiopian Agent | Approve Dubai Elite's selection | Selection "approved", candidate "approved" |
| 10 | Gulf Home (SA) | View rejected selection | Shows "Rejected" with reason |
| 11 | Kuwait Family (KW) | Try to select | Cannot: candidate permanently locked |

**Expected Results**:
- ✅ Sequential selection works correctly
- ✅ Rejection â†’ immediate availability
- ✅ First to select after rejection wins
- ✅ Once approved, candidate permanently locked
- ✅ All notifications sent appropriately
- ✅ **Concurrency note**: The losing simultaneous-select request (Step 7 if Dubai Elite and Kuwait Family both click at the exact same instant) receives `HTTP 409 Conflict` — see UC-055 #1 for the lock mechanism details.

---

### UC-071: Cross-Timezone Date & Timestamp Handling

**Priority**: Medium
**Type**: Cross-Cutting
**Est. Time**: 10 min

**User Story**: The platform serves Ethiopian agencies (UTC+3) and foreign partners across Gulf/Southeast Asia timezones (UTC+3 to UTC+8). All date-sensitive operations must display and compute correctly regardless of the user's timezone setting.

**Preconditions**:
- Tester identity profile has `timezone` field set to Africa/Addis_Ababa (UTC+3)
- Foreign partner in the pairing has `timezone` set to Asia/Riyadh (UTC+3) or Asia/Dubai (UTC+4) or Asia/Kuala_Lumpur (UTC+8)
- At least one candidate with a known `created_at` timestamp
- One selection with a known `locked_at` timestamp
- The lock expiry window is configured (default 24h)

**Scenario Flow**:
| Step | Actor | Action | Expected Result |
|---|---|---|---|
| 1 | Tester | Inspect candidate creation timestamp on Ethiopian UI | Timestamp displayed in local time (Africa/Addis_Ababa, UTC+3), e.g. "2026-07-08 09:00 AM EAT" |
| 2 | Tester | Switch context to Foreign partner view (or ask partner to check) | Same candidate's `created_at` shown in partner's timezone, e.g. Asia/Riyadh UTC+3 → same local time; Asia/Dubai UTC+4 → "2026-07-08 10:00 AM GST" |
| 3 | Tester | Note a candidate locked at 2026-07-07 15:00 EAT. Lock expires in 24h. | Expiry shown as "2026-07-08 15:00 EAT" for Ethiopian, "2026-07-08 16:00 GST" for Dubai partner |
| 4 | Tester | Wait past expiry (or ask dev to fast-forward cron) | Lock expired; candidate returns to `available`. The cron worker uses UTC internally — expiry calculation is timezone-agnostic. |
| 5 | Tester | Open the selection approval audit log | Timestamps recorded in UTC with timezone offset appended. E.g. `2026-07-08T06:00:00Z (+03:00)` |
| 6 | Tester | Send a chat message at known time | Message timestamp on sender: local time. On receiver (different tz): converted to receiver's local time. |
| 7 | Tester | Change own profile timezone from Africa/Addis_Ababa to Asia/Dubai | All dates on next page load use new timezone. Chat message timestamps update retroactively. |

**Expected Results**:
- ✅ Each user sees timestamps in their own timezone
- ✅ Lock expiry calculation is consistent across all timezones (UTC-based)
- ✅ Audit logs record UTC with offset — unambiguous
- ✅ Chat timestamps are correctly converted per-recipient
- ✅ Timezone change is reflected immediately on next page load

---

### UC-072: Email Service Unavailable — Fallback Path

**Priority**: Medium
**Type**: Cross-Cutting / Failure Mode
**Est. Time**: 10 min

**User Story**: When the SMTP service is down or misconfigured, every action that would normally send a notification email must gracefully degrade without breaking the user flow.

**Preconditions**:
- SMTP service explicitly disabled (or health step 1 from UC-000 reports `{"email":"disabled","mode":"simulated"}`)
- 1 Ethiopian agent, 1 Foreign agent, active pairing, 1 candidate
- Tester confirmed the email mode from UC-000 step 6

**Scenario Flow**:
| Step | Actor | Action | Expected Result |
|---|---|---|---|
| 1 | Ethiopian Agent | Create candidate and share with Foreign partner | Candidate shared successfully. In-platform notification sent. **No email sent.** Backend logs: `[EMAIL SKIPPED] simulated mode — would send to partner@agency.com re: candidate Almaz Tadesse`. UI does NOT show an error. |
| 2 | Ethiopian Agent | Submit selection for approval | Selection transitions to `under_review`. No email sent to Foreign agent. In-app notification bell still shows the event. |
| 3 | Foreign Agent | Check inbox | No email received. Refresh page — candidate appears in selections list (the flow completed without email). |
| 4 | Admin | Trigger "welcome" email via admin panel (if applicable) | Backend returns success to UI but logs the skip. Admin sees a warning banner: "Email service is in simulated mode — no actual emails are being sent." |
| 5 | Ethiopian Agent | Initiate password reset | Reset token generated and displayed on-screen: "In production this would be emailed. Your reset token: abc-123-def" (because email is unavailable, the token is shown directly). Token works for next 15 minutes. |
| 6 | Tester | Re-enable SMTP, repeat step 1 | Email is sent. Backend logs: `[EMAIL SENT] to partner@agency.com`. UI remains identical — no change from the user's perspective. |

**Expected Results**:
- ✅ All functional flows complete without email
- ✅ In-app notifications are the fallback channel — they always work regardless of email state
- ✅ Logged skipped emails make debugging possible
- ✅ Admin is explicitly warned about simulated mode
- ✅ Password reset provides on-screen token when email unavailable (security: token expires, logged)
- ✅ When SMTP is restored, behaviour returns to normal transparently

---

## 14. End-to-End User Journey Maps

### Journey 1: Complete Ethiopian Agency Lifecycle

**Duration**: ~60 min (all use cases sequentially)
**Actors**: Ethiopian Agent (Biruk), Foreign Agent (Gulf Home), Super Admin

```
Pre-Flight
  [UC-000] Verify environment health and email mode

Phase 1: Onboarding
  [UC-001] Register as Ethiopian Agent
  [UC-002] Verify Email
  [UC-003] Admin Approves Agency
  [UC-005] Admin Creates Pairing (Addis + Gulf Home, Dubai Elite, Kuwait Family)
  [UC-006] First Login & Dashboard

Phase 2: Candidate Formation
  [UC-008] Create Candidate (Almaz) â€” Full Profile
  [UC-009] Draft Auto-Save (create Tigist with partial data)
  [UC-010] Passport Upload + OCR (Almaz's passport)
  [UC-011] OCR Preview (test without saving)
  [UC-012] Upload Photo & Video (Almaz)
  [UC-013] Edit Candidate Profile (fix/add details)
  [UC-014] Publish Candidate (draft to available)
  [UC-015] Batch Publish (Tigist + Hiwot + Muluwork)
  [UC-016] Partner-Specific Overrides (set per-partner)

Phase 3: Workspace Management
  [UC-017] Switch Workspace via PartnerSwitcher
  [UC-018] Configure Sharing Preferences
  [UC-019] Share/Unshare Candidates in Workspace
  [UC-020] View Workspace Overview

Phase 4: CV Generation
  [UC-021] Generate CV for Almaz
  [UC-022] Per-Partner CV with Overrides
  [UC-023] Bulk CV ZIP Download
  [UC-024] CV Regeneration after update

Phase 5: Selection & Approval
  [UC-025] Foreign Agent Browses Candidates
  [UC-026] Foreign Agent Creates Selection (Almaz via Gulf Home)
  [UC-027] Lock Mechanism (prevent Dubai Elite)
  [UC-028] Ethiopian Agent Approves Selection
  [UC-031] (Optional) Reject Scenario
  [UC-029] (Optional) Dual Approval
  [UC-033] (Optional) Selection Auto-Expiry

Phase 6: Post-Selection Tracking
  [UC-038] Update Medical Step
  [UC-039] Update CoC Step
  [UC-040] Update Visa & Ticket Steps
  [UC-041] Update Arrival Step (complete!)
  [UC-043] Batch Progress Update
  [UC-042] Foreign Agent Views Progress

Phase 7: Communication
  [UC-045] Real-Time Chat
  [UC-046] Candidate-Scoped Chat Thread
  [UC-047] All Notification Types
  [UC-048] WebSocket Resilience

Cross-Cutting
  [UC-071] Timezone Handling (verify all timestamps)
  [UC-072] Email-Disabled Fallback (re-run if email is simulated)
  [UC-073] Mid-Flow Unpairing (end workspace mid-selection)
  [UC-055] Concurrent Operations
  [UC-056] Data Isolation & RBAC
  [UC-057] Input Validation
  [UC-058] File Upload Edge Cases
  [UC-059] Session Edge Cases
  [UC-062] Mobile Responsiveness
  [UC-064] Waiting State
  [UC-065] Maintenance Mode
  [UC-066] Password Reset
  [UC-069] Account Settings
  [UC-070] Multi-Partner Competition

Admin Governance
  [UC-049] Admin Dashboard Analytics
  [UC-050] Admin Suspend/Activate Agency
  [UC-051] Read-Only Oversight
  [UC-052] Audit Logs
  [UC-053] Platform Settings
  [UC-054] Admin Account Management
```

---

### Journey 2: Foreign Agent Complete Experience (System Reaction Focus)

**Duration**: ~30 min
**Actors**: Foreign Agent (Gulf Home), Ethiopian Agent (Biruk), Other Foreign Agents

```
Pre-Flight
  [UC-000] Verify environment health (especially email mode for simulated notifications)

Phase 1: Foreign Onboarding
  [UC-004] Register as Foreign Agency
  [UC-002] Verify Email
  [UC-003] Admin Approval
  [UC-005] Pairing Created with Ethiopian Agency
  [UC-007] First Login & Employer Dashboard

Phase 2: Browse & Select
  [UC-025] Browse available candidates (filters, sorting)
  [UC-016] View partner-specific job details (overrides)
  [UC-021] Download CV
  [UC-010] View candidate documents (passport, photo, video)

Phase 3: Create Selection (THE BIG ONE)
  [UC-026] Create Selection â€” FULL SYSTEM REACTION:
            - Selection record created (pending)
            - Contract + ID uploaded to S3
            - Candidate locked globally
            - Ethiopian agent notified (bell, WS, dashboard)
            - Other foreign agents see "Locked" state
            - Lock timer starts
            - Selection page appears for both sides
            - Dashboard stats update for both roles
            - Admin sees new selection

Phase 4: Await Decision
  [UC-027] Other foreign agents experience locked state
  [UC-035] Upload/replace employer documents
  [UC-045] Chat with Ethiopian agent about selection
  [UC-046] Candidate-scoped chat thread

Phase 5: Approval/Rejection
  [UC-028] Ethiopian approves â†’ FULL SYSTEM REACTION:
            - Selection "approved"
            - Candidate permanently locked
            - Tracking section appears (0% progress)
            - Notification received in real-time
            - Dashboard updates
            - Other agents notified of rejection

  [UC-031] OR Ethiopian rejects â†’ FULL SYSTEM REACTION:
            - Selection "rejected"
            - Candidate unlocked
            - Reason visible
            - Other agents notified of availability
  [UC-032] Foreign agent experiences rejection
  [UC-034] Foreign agent re-selects with better offer

Phase 6: Track Progress
  [UC-042] View progress timeline (read-only)
  [UC-038] Medical step â†’ 20% notification
  [UC-039] CoC step â†’ 40% notification
  [UC-040] Visa step â†’ 60% notification
  [UC-040] Ticket step â†’ 80% notification
  [UC-041] Arrival step â†’ 100% Complete! ðŸŽ‰

Phase 7: Completion
  [UC-041] "Candidate arrived" notification received
  [UC-047] View all notification types
  [UC-037] View selection history

Cross-Cutting
  [UC-071] Timezone Handling (verify timestamps display in your timezone)
  [UC-072] Email-Disabled Fallback (verify in-app fallback when email is off)
```

---

### Journey 3: Admin Full Governance Flow

**Duration**: ~25 min
**Actor**: Super Admin

```
Pre-Flight
  [UC-000] Verify environment health and email mode
  [UC-071] Timezone Handling (verify admin audit timestamps)

Admin Operations
  [UC-003] Approve Pending Agency
  [UC-005] Create Pairing Workspace
  [UC-049] View Dashboard Analytics
  [UC-050] Suspend & Reactivate Agency
  [UC-051] View Candidates & Selections (read-only)
  [UC-052] View Audit Logs
  [UC-053] Configure Platform Settings
            - Set lock duration
            - Toggle dual approval
            - Toggle auto-expire
            - Enable maintenance mode
   [UC-065] Verify maintenance mode blocks users
  [UC-073] Mid-Flow Unpairing (end workspace while selection is active)
  [UC-054] Create Support Admin
```

---

### Journey 4: Multi-Partner Competition (Stress Test)

**Duration**: ~25 min
**Actors**: 1 Ethiopian Agent, 3 Foreign Agents

```
Pre-Flight
  [UC-000] Verify environment health
  [UC-055] Concurrency: note lock mechanism active (PostgreSQL row-level lock)

Competition Flow
  [UC-008] Ethiopian creates one candidate
  [UC-016] Set overrides for each partner
  [UC-014] Publish to all partners
  [UC-025] All 3 foreign agents browse
  [UC-026] Foreign Agent 1 selects â†’ locked
  [UC-027] Foreign Agents 2 & 3 see locked
  [UC-031] Ethiopian rejects Foreign Agent 1
  [UC-034] Foreign Agent 2 selects (re-selection)
  [UC-028] Ethiopian approves Foreign Agent 2
  [UC-027] Foreign Agent 3 blocked permanently
  [UC-070] Verify all notifications received
```

---

### Journey 5: Edge Case & Error Exploration

**Duration**: ~35 min
**Actor**: Ethiopian Agent + Foreign Agent

```
Pre-Flight
  [UC-000] Verify environment health, record email mode

Edge & Error Coverage
  [UC-057] Test all form validation errors
  [UC-058] Test file upload edge cases
  [UC-059] Test session scenarios
  [UC-055] Test concurrent operations
  [UC-056] Test data isolation
  [UC-064] Test waiting state (no pairing)
  [UC-065] Test maintenance mode
  [UC-066] Test password reset
  [UC-036] Test selection unlock
  [UC-060] Test data deletion
  [UC-061] Test performance benchmarks
  [UC-062] Test mobile responsiveness
  [UC-063] Test dark mode
  [UC-071] Test timezone display (Ethiopian + Foreign agent views)
  [UC-072] Test email-disabled fallback paths
  [UC-073] Test mid-flow workspace unpairing (admin ends pairing during selection)
```

### Journey 6: Selection â€” Complete Lifecycle of a Single Candidate

**Duration**: ~30 min
**Actors**: Biruk (ET), Gulf Home (FA), Dubai Elite (UAE), Admin

```
Pre-Flight
  [UC-000] Verify environment health and email mode

This journey tracks ONE candidate (Almaz) through the ENTIRE system,
documenting every component that touches her:

  Candidate State:     draft â†’ available â†’ locked â†’ approved â†’ in_progress â†’ completed
  Selection State:     (none) â†’ pending â†’ approved â†’ (tracking starts) â†’ (arrived)
  Foreign Agents:      Gulf Home browses â†’ selects â†’ approved â†’ tracks completion
                       Dubai Elite browses â†’ blocked â†’ sees locked â†’ sees approved
  Notifications:       (published) â†’ (selection created) â†’ (approved) â†’ (each step) â†’ (arrived)
  Ethiopian Dashboard: 0 candidates â†’ 1 available â†’ 1 pending â†’ 1 in progress â†’ 1 completed
  Gulf Home Dashboard: 0 â†’ 1 available â†’ 1 pending â†’ 1 approved â†’ tracking % â†’ done
  Admin View:          Candidate appears â†’ selection appears â†’ status changes â†’ completed
  Chat:                Workspace chat â†’ candidate-specific thread
  Documents:           Passport, Photo, Video, CV, Contract, ID, Medical, Visa, Ticket
```

---

## Appendix A: Use Case Summary Table

| UC# | Name | Priority | Est. Time | Module |
|---|---|---|---|---|---|---|
| UC-000 | Pre-Flight Environment Verification | Critical | 5 min | Setup |
| UC-001 | Ethiopian Agency Self-Registration | Critical | 10 min | Auth |
| UC-002 | Email Verification Flow | Critical | 5 min | Auth |
| UC-003 | Admin Approval of Ethiopian Agency | Critical | 5 min | Admin |
| UC-004 | Foreign Agency Self-Registration | Critical | 8 min | Auth |
| UC-005 | Admin Creates Agency Pairings (Workspaces) | Critical | 10 min | Admin |
| UC-006 | Ethiopian Agent â€” First Login & Dashboard | Critical | 5 min | Dashboard |
| UC-007 | Foreign Agent â€” First Login & Dashboard | Critical | 5 min | Dashboard |
| UC-008 | Create Candidate Profile â€” Full Flow | Critical | 15 min | Candidates |
| UC-009 | Draft Auto-Save & Recovery | High | 8 min | Candidates |
| UC-010 | Passport Upload with OCR Processing | Critical | 10 min | Candidates/OCR |
| UC-011 | OCR Preview (Parse Without Save) | Medium | 5 min | Candidates/OCR |
| UC-012 | Upload Photo & Video Documents | High | 8 min | Candidates |
| UC-013 | Edit Candidate Profile | High | 10 min | Candidates |
| UC-014 | Candidate Status Lifecycle (Draft to Available) | Critical | 8 min | Candidates |
| UC-015 | Batch Publish Multiple Candidates | High | 10 min | Candidates |
| UC-016 | Partner-Specific Job Details (Pair Override) | High | 10 min | Candidates |
| UC-017 | Workspace Switching via PartnerSwitcher | High | 5 min | Partners |
| UC-018 | Workspace Sharing Preferences | Medium | 5 min | Partners |
| UC-019 | Candidate Share/Unshare in Workspace | High | 8 min | Partners |
| UC-020 | Partners Page â€” Workspace Overview | Medium | 8 min | Partners |
| UC-021 | Generate Candidate CV | Critical | 10 min | CV |
| UC-022 | Per-Partner CV with Custom Overrides | High | 10 min | CV |
| UC-023 | Bulk CV ZIP Download | Medium | 8 min | CV |
| UC-024 | CV Regeneration | Medium | 5 min | CV |
| UC-025 | Foreign Agent Browses Available Candidates | Critical | 12 min | Candidates/FA |
| UC-026 | Foreign Agent Creates Selection Request | Critical | 15 min | Selections |
| UC-027 | Lock Mechanism â€” Foreign Agent Sees Locked State | High | 10 min | Selections |
| UC-028 | Ethiopian Agent Approves Selection | Critical | 12 min | Selections |
| UC-029 | Dual Approval Flow (Both Parties) | Medium | 12 min | Selections |
| UC-030 | Foreign Agent Approves (Dual Approval Second Step) | Medium | 8 min | Selections |
| UC-031 | Ethiopian Agent Rejects Selection | High | 8 min | Selections |
| UC-032 | Foreign Agent Experiences Rejection | High | 8 min | Selections |
| UC-033 | Selection Auto-Expiry | High | 15 min | Selections |
| UC-034 | Foreign Agent Re-Selects After Rejection/Expiry | High | 10 min | Selections |
| UC-035 | Foreign Agent Uploads/Replaces Employer Documents | Medium | 5 min | Selections |
| UC-036 | Selection Unlock / Release by Ethiopian Agent | Medium | 8 min | Selections |
| UC-037 | Foreign Agent Views Selection History | Low | 5 min | Selections |
| UC-038 | Update Selection Progress â€” Medical Step | Critical | 10 min | Tracking |
| UC-039 | Update Certificate of Clearance (CoC) | High | 8 min | Tracking |
| UC-040 | Update Visa & Ticket Steps | High | 8 min | Tracking |
| UC-041 | Update Arrival Step â€” Complete the Process | High | 8 min | Tracking |
| UC-042 | Foreign Agent Views Progress Timeline | Medium | 5 min | Tracking |
| UC-043 | Batch Progress Update | Medium | 8 min | Tracking |
| UC-044 | Foreign Agent Views Tracking Hub | Low | 5 min | Tracking |
| UC-045 | Real-Time Chat Between Workspace Partners | Medium | 10 min | Chat |
| UC-046 | Candidate-Scoped Chat Thread | Medium | 5 min | Chat |
| UC-047 | All Notification Types â€” Complete Verification | Medium | 20 min | Notifications |
| UC-048 | WebSocket Connection & Reconnection | Medium | 8 min | Notifications |
| UC-049 | Admin Dashboard Analytics | Medium | 8 min | Admin |
| UC-050 | Admin Agency Management â€” Suspend/Activate | High | 8 min | Admin |
| UC-051 | Admin Read-Only Oversight | Low | 5 min | Admin |
| UC-052 | Admin Audit Logs | Medium | 8 min | Admin |
| UC-053 | Admin Platform Settings | High | 10 min | Admin |
| UC-054 | Admin Account Management | Medium | 8 min | Admin |
| UC-055 | Concurrent Operations & Race Conditions | High | 15 min | Cross |
| UC-056 | Data Isolation & RBAC | Critical | 15 min | Cross |
| UC-057 | Input Validation & Error Handling | High | 12 min | Cross |
| UC-058 | File Upload Edge Cases | High | 10 min | Cross |
| UC-059 | Session & Authentication Edge Cases | Critical | 12 min | Cross |
| UC-060 | Data Deletion & Recovery | High | 10 min | Cross |
| UC-061 | Performance & Load Testing | Medium | 10 min | Cross |
| UC-062 | Mobile Responsiveness | Medium | 20 min | UI/UX |
| UC-063 | Dark Mode / Theme | Low | 8 min | UI/UX |
| UC-064 | Waiting State â€” No Active Pairings | Medium | 5 min | Dashboard |
| UC-065 | Platform Maintenance Mode | Medium | 8 min | Admin |
| UC-066 | Password Reset Flow | High | 10 min | Auth |
| UC-067 | Email Verification Resend & Cooldown | Medium | 8 min | Auth |
| UC-068 | Partner Logo Upload & CV Branding | Low | 5 min | Partners/CV |
| UC-069 | Account Settings â€” Profile & Security | Medium | 10 min | Settings |
| UC-070 | Multi-Partner Selection Competition | High | 20 min | Selections |
| UC-071 | Cross-Timezone Date & Timestamp Handling | Medium | 10 min | Cross |
| UC-072 | Email Service Unavailable Fallback Path | Medium | 10 min | Cross |
| UC-073 | Mid-Flow Workspace Unpairing | High | 15 min | Cross/Edge |

**Total: 74 Use Cases** | **Estimated Full Suite Time: ~10-12 hours**

---

*End of Document*

