# Assignment 2: User Test Cases
## Maid Recruitment Platform

### Document Information
- **Project**: Maid Recruitment Platform
- **Assignment**: Assignment 2 - User Testing
- **Date**: June 2026
- **Version**: 1.0

---

## Table of Contents
1. [Test Environment Setup](#test-environment-setup)
2. [User Roles](#user-roles)
3. [Test Scenarios](#test-scenarios)
4. [Detailed Test Cases](#detailed-test-cases)

---

## Test Environment Setup

### Prerequisites
- **Frontend URL**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **Database**: PostgreSQL running via Docker
- **Storage**: MinIO (S3-compatible) running via Docker
- **Test Browsers**: Chrome, Firefox, Safari, Edge

### Test Credentials
```
Ethiopian Agent:
Email: ethiopian@test.com
Password: password123

Foreign Agent:
Email: foreign@test.com
Password: password123
```

---

## User Roles

### Ethiopian Agent (Local Recruitment Agency)
- Creates and manages candidate profiles
- Uploads candidate documents (passport, photos, videos)
- Publishes candidates to foreign partners
- Manages pairing requests with foreign agencies
- Generates CVs and recruitment materials

### Foreign Agent (Overseas Recruitment Agency)
- Views published candidates from Ethiopian partners
- Creates selection requests for candidates
- Reviews and approves/rejects candidate selections
- Manages partnerships with Ethiopian agencies
- Downloads candidate materials and CVs

---

## Test Scenarios

### Scenario 1: Complete Candidate Lifecycle
**Objective**: Test the full journey from candidate creation to selection approval

**Actors**: Ethiopian Agent, Foreign Agent

**Flow**:
1. Ethiopian agent creates candidate profile
2. Ethiopian agent uploads documents
3. Ethiopian agent publishes candidate
4. Foreign agent views published candidate
5. Foreign agent creates selection request
6. Ethiopian agent reviews and approves selection
7. Both agents communicate via chat

---

### Scenario 2: Multi-Partner Collaboration
**Objective**: Test candidate sharing across multiple foreign partners

**Actors**: Ethiopian Agent, Multiple Foreign Agents

**Flow**:
1. Ethiopian agent creates candidate with partner-specific job details
2. Candidate published to multiple partners simultaneously
3. Multiple partners create competing selection requests
4. Ethiopian agent manages multiple selections
5. First-come-first-served lock mechanism tested

---

### Scenario 3: Document Management & OCR
**Objective**: Test passport OCR and document handling

**Actors**: Ethiopian Agent

**Flow**:
1. Upload passport document
2. Verify OCR auto-fills candidate data
3. Upload photo and video
4. Generate CV with all documents
5. Download and verify CV quality

---

## Detailed Test Cases

### TC-001: User Registration (Ethiopian Agent)

**Priority**: High  
**Type**: Functional  
**Estimated Time**: 5 minutes

#### Preconditions
- Access to registration page
- Valid email address

#### Test Steps
1. Navigate to http://localhost:3000/register
2. Select "Ethiopian Agent" role
3. Fill in the following fields:
   - Full Name: "Ahmed Hassan"
   - Company Name: "Addis Recruitment Services"
   - Email: "ahmed@addisrecruitment.com"
   - Password: "SecurePass123!"
   - Confirm Password: "SecurePass123!"
4. Check "I agree to Terms and Conditions"
5. Click "Create Account"

#### Expected Results
- ✅ Account created successfully
- ✅ Redirect to email verification page
- ✅ Verification email sent
- ✅ Success toast notification displayed

#### Actual Results
- [ ] Pass
- [ ] Fail (Describe issue): _______________

#### Notes
_______________________________________________

---

### TC-002: User Login (Ethiopian Agent)

**Priority**: High  
**Type**: Functional  
**Estimated Time**: 2 minutes

#### Preconditions
- User account exists and is verified
- User is on login page

#### Test Steps
1. Navigate to http://localhost:3000/login
2. Enter email: "ethiopian@test.com"
3. Enter password: "password123"
4. Check "Remember me" (optional)
5. Click "Sign In"

#### Expected Results
- ✅ Login successful
- ✅ Redirect to Ethiopian agent dashboard
- ✅ Welcome message displayed
- ✅ User menu shows correct name and role

#### Actual Results
- [ ] Pass
- [ ] Fail (Describe issue): _______________

#### Notes
_______________________________________________

---

### TC-003: Create Candidate Profile

**Priority**: Critical  
**Type**: Functional  
**Estimated Time**: 10 minutes

#### Preconditions
- Logged in as Ethiopian agent
- Access to candidate creation page

#### Test Steps
1. Navigate to Candidates → Add New
2. Fill Personal Information:
   - Full Name: "Almaz Tadesse"
   - Nationality: "Ethiopian"
   - Date of Birth: "1995-05-15"
   - Place of Birth: "Addis Ababa"
   - Gender: "Female"
   - Religion: "Christian"
   - Marital Status: "Single"
   - Children: "0"

3. Fill Passport Information:
   - Passport Number: "EP1234567"
   - Issue Date: "2023-01-01"
   - Expiry Date: "2028-01-01"
   - Issuing Authority: "Ethiopia Immigration"

4. Fill Education & Experience:
   - Education Level: "High School"
   - Years of Experience: "5"
   - Add Experience Abroad: "Saudi Arabia - 3 years"

5. Select Skills:
   - ✓ Cooking
   - ✓ Cleaning
   - ✓ Childcare
   - ✓ Elderly Care

6. Add Languages:
   - Language: "Arabic", Proficiency: "Fluent"
   - Language: "English", Proficiency: "Intermediate"
   - Language: "Amharic", Proficiency: "Fluent"

7. Click "Save Candidate"

#### Expected Results
- ✅ Candidate created successfully
- ✅ Redirect to candidate detail page
- ✅ Success notification displayed
- ✅ All data saved correctly
- ✅ Candidate appears in candidates list

#### Actual Results
- [ ] Pass
- [ ] Fail (Describe issue): _______________

#### Notes
_______________________________________________

---

### TC-004: Upload Passport Document with OCR

**Priority**: Critical  
**Type**: Functional  
**Estimated Time**: 5 minutes

#### Preconditions
- Candidate profile created
- Valid passport document (PDF/JPG/PNG) available
- On candidate creation/edit page

#### Test Steps
1. In the "Documents and Media" section
2. Click "Upload" under "Passport document"
3. Select a passport file (< 10MB)
4. Wait for upload to complete
5. Observe OCR processing indicator
6. Wait for OCR to complete

#### Expected Results
- ✅ File uploads successfully
- ✅ "Processing passport OCR..." indicator appears
- ✅ OCR completes within 10 seconds
- ✅ Following fields auto-filled:
  - Passport Number
  - Gender
  - Date of Birth
  - Passport Issue Date
  - Passport Expiry Date
- ✅ Preview thumbnail displayed
- ✅ Success notification shown

#### Actual Results
- [ ] Pass
- [ ] Fail (Describe issue): _______________

#### OCR Accuracy
- Full Name: [ ] Correct [ ] Incorrect
- Gender: [ ] Correct [ ] Incorrect
- DOB: [ ] Correct [ ] Incorrect
- Passport #: [ ] Correct [ ] Incorrect

#### Notes
_______________________________________________

---

### TC-005: Upload Photo and Video

**Priority**: High  
**Type**: Functional  
**Estimated Time**: 5 minutes

#### Preconditions
- Candidate profile exists
- Photo file (JPG/PNG < 10MB) available
- Video file (MP4 < 50MB) available

#### Test Steps
1. Navigate to candidate edit page
2. In "Documents and Media" section
3. Upload full-body photo:
   - Click "Upload" under "Full body photo"
   - Select photo file
   - Wait for upload completion
4. Upload video interview:
   - Click "Upload" under "Video interview"
   - Select video file
   - Wait for upload completion

#### Expected Results
- ✅ Photo uploads successfully
- ✅ Photo preview displayed
- ✅ Video uploads successfully
- ✅ Video preview/thumbnail displayed
- ✅ Progress indicators work correctly
- ✅ Success notifications shown
- ✅ Files can be removed and re-uploaded

#### Actual Results
- [ ] Pass
- [ ] Fail (Describe issue): _______________

#### Notes
_______________________________________________

---

### TC-006: Publish Candidate to Foreign Partner

**Priority**: Critical  
**Type**: Functional  
**Estimated Time**: 3 minutes

#### Preconditions
- Candidate profile created with all required fields
- Active pairing exists with foreign partner
- Logged in as Ethiopian agent

#### Test Steps
1. Navigate to candidate detail page
2. Locate "Publish Candidate" button
3. Click "Publish Candidate"
4. If multiple partners exist, select target partner
5. Confirm publication

#### Expected Results
- ✅ Candidate status changes to "published"
- ✅ Success notification displayed
- ✅ Badge shows "Published" status
- ✅ Foreign partner can now view candidate
- ✅ Notification sent to foreign partner
- ✅ Candidate appears in foreign partner's dashboard

#### Actual Results
- [ ] Pass
- [ ] Fail (Describe issue): _______________

#### Notes
_______________________________________________

---

### TC-007: Generate Candidate CV

**Priority**: High  
**Type**: Functional  
**Estimated Time**: 5 minutes

#### Preconditions
- Published candidate with complete profile
- Documents uploaded
- Agency logo configured (optional)

#### Test Steps
1. Navigate to candidate detail page
2. Click "Generate CV" or "Download CV" button
3. Wait for CV generation
4. PDF downloads automatically
5. Open PDF and review

#### Expected Results
- ✅ CV generates within 5 seconds
- ✅ PDF downloads successfully
- ✅ CV contains all candidate information:
  - Personal details
  - Photo displayed correctly
  - Skills listed
  - Languages with proficiency
  - Experience abroad details
  - Education level
- ✅ Agency logo appears (if configured)
- ✅ Professional formatting maintained
- ✅ No data truncation or overflow

#### Actual Results
- [ ] Pass
- [ ] Fail (Describe issue): _______________

#### CV Quality Assessment
- Layout: [ ] Professional [ ] Needs Improvement
- Data Accuracy: [ ] 100% [ ] Has Errors
- Image Quality: [ ] Clear [ ] Blurry

#### Notes
_______________________________________________

---

### TC-008: Foreign Agent Views Published Candidates

**Priority**: Critical  
**Type**: Functional  
**Estimated Time**: 5 minutes

#### Preconditions
- Foreign agent account active
- Ethiopian partner has published candidates
- Active pairing exists

#### Test Steps
1. Login as foreign agent
2. Navigate to "Candidates" page
3. View list of published candidates
4. Apply filters:
   - Skills filter
   - Experience filter
   - Age range filter
5. Click on a candidate to view details

#### Expected Results
- ✅ Only published candidates visible
- ✅ Candidates from paired Ethiopian agencies shown
- ✅ Filters work correctly
- ✅ Candidate cards display:
  - Name
  - Photo
  - Age
  - Experience
  - Skills
  - Availability status
- ✅ Detail page shows complete information
- ✅ Documents are downloadable
- ✅ CV can be generated/downloaded

#### Actual Results
- [ ] Pass
- [ ] Fail (Describe issue): _______________

#### Notes
_______________________________________________

---

### TC-009: Create Selection Request

**Priority**: Critical  
**Type**: Functional  
**Estimated Time**: 5 minutes

#### Preconditions
- Logged in as foreign agent
- Published candidate available
- No existing active selection for this candidate

#### Test Steps
1. Navigate to candidate detail page
2. Click "Select Candidate" button
3. Fill selection form:
   - Destination Country: "Saudi Arabia"
   - Salary Offered: "1500 SAR"
   - Additional Notes: "Urgent requirement for housemaid"
4. Click "Submit Selection Request"

#### Expected Results
- ✅ Selection request created successfully
- ✅ Success notification displayed
- ✅ Selection status badge appears on candidate
- ✅ Ethiopian agent receives notification
- ✅ Selection appears in "Selections" list
- ✅ Lock timer starts (if configured)
- ✅ Other foreign agents see "locked" status

#### Actual Results
- [ ] Pass
- [ ] Fail (Describe issue): _______________

#### Notes
_______________________________________________

---

### TC-010: Ethiopian Agent Approves Selection

**Priority**: Critical  
**Type**: Functional  
**Estimated Time**: 3 minutes

#### Preconditions
- Selection request exists
- Logged in as Ethiopian agent
- Selection is in "pending" status

#### Test Steps
1. Navigate to "Selections" page
2. Locate pending selection
3. Click to view selection details
4. Review selection information
5. Click "Approve" button
6. Confirm approval in dialog

#### Expected Results
- ✅ Selection status changes to "approved"
- ✅ Success notification displayed
- ✅ Foreign agent receives notification
- ✅ Candidate status updates
- ✅ Other pending selections auto-rejected
- ✅ Selection lock released
- ✅ Timeline updated with approval event

#### Actual Results
- [ ] Pass
- [ ] Fail (Describe issue): _______________

#### Notes
_______________________________________________

---

### TC-011: Selection Lock Mechanism

**Priority**: High  
**Type**: Functional  
**Estimated Time**: 10 minutes

#### Preconditions
- Multiple foreign agents paired with same Ethiopian agency
- One published candidate available

#### Test Steps
1. **Foreign Agent 1**: Create selection request
2. Verify lock timer starts
3. **Foreign Agent 2**: Attempt to create selection for same candidate
4. Wait for lock timer to expire
5. **Foreign Agent 2**: Retry selection creation

#### Expected Results
- ✅ Agent 1 successfully creates selection
- ✅ Lock timer displays (e.g., "30:00" countdown)
- ✅ Agent 2 sees "Temporarily locked" message
- ✅ Agent 2 cannot create selection while locked
- ✅ After timer expires, lock releases
- ✅ Agent 2 can now create selection
- ✅ Only one active selection allowed at a time

#### Actual Results
- [ ] Pass
- [ ] Fail (Describe issue): _______________

#### Notes
_______________________________________________

---

### TC-012: Real-time Chat Between Agents

**Priority**: Medium  
**Type**: Functional  
**Estimated Time**: 5 minutes

#### Preconditions
- Active pairing between Ethiopian and foreign agents
- Both agents logged in

#### Test Steps
1. **Ethiopian Agent**: Navigate to chat
2. Select foreign partner
3. Type message: "Hello, I have 3 new candidates available"
4. Send message
5. **Foreign Agent**: Open chat
6. Verify message received
7. Reply: "Great! I'll review them today"
8. **Ethiopian Agent**: Verify reply received

#### Expected Results
- ✅ Messages sent instantly
- ✅ Messages appear in real-time (< 2 seconds)
- ✅ Unread message badges update
- ✅ Message history persists
- ✅ Timestamps displayed correctly
- ✅ Notifications sent for new messages
- ✅ Chat works bidirectionally

#### Actual Results
- [ ] Pass
- [ ] Fail (Describe issue): _______________

#### Notes
_______________________________________________

---

### TC-013: Notifications System

**Priority**: Medium  
**Type**: Functional  
**Estimated Time**: 10 minutes

#### Test Steps
1. Perform various actions that trigger notifications:
   - Create selection request
   - Approve/reject selection
   - Send chat message
   - Publish candidate
2. Click notification bell icon
3. Review notification list
4. Click on a notification
5. Mark notifications as read

#### Expected Results
- ✅ Notification badge shows unread count
- ✅ All relevant events generate notifications:
  - Selection created
  - Selection approved
  - Selection rejected
  - New message received
  - Candidate published
  - Selection expired
- ✅ Notifications contain correct information
- ✅ Clicking notification navigates to relevant page
- ✅ Mark as read functionality works
- ✅ Notifications persist across sessions
- ✅ Real-time updates (< 3 seconds)

#### Actual Results
- [ ] Pass
- [ ] Fail (Describe issue): _______________

#### Notes
_______________________________________________

---

### TC-014: Profile Settings Management

**Priority**: Medium  
**Type**: Functional  
**Estimated Time**: 5 minutes

#### Preconditions
- User logged in

#### Test Steps
1. Navigate to Settings page
2. **Profile Settings**:
   - Update full name
   - Update company name
   - Upload agency logo
3. **Preferences**:
   - Change theme (Light/Dark/System)
   - Toggle email notifications
   - Toggle selection alerts
4. **Security**:
   - Change password
   - View active sessions
   - Logout from other devices
5. Click "Save" for each section

#### Expected Results
- ✅ All fields update successfully
- ✅ Logo upload works (JPG/PNG)
- ✅ Logo appears in generated CVs
- ✅ Theme changes immediately
- ✅ Notification preferences saved
- ✅ Password change requires current password
- ✅ Active sessions list displayed
- ✅ Logout all devices works
- ✅ Settings persist after logout/login

#### Actual Results
- [ ] Pass
- [ ] Fail (Describe issue): _______________

#### Notes
_______________________________________________

---

### TC-015: Responsive Design (Mobile)

**Priority**: Medium  
**Type**: UI/UX  
**Estimated Time**: 15 minutes

#### Test Devices
- [ ] iPhone (Safari)
- [ ] Android (Chrome)
- [ ] iPad (Safari)
- [ ] Responsive mode in browser

#### Test Steps
Test the following pages on mobile:
1. Login page
2. Dashboard
3. Candidates list
4. Candidate detail
5. Create candidate form
6. Selections list
7. Chat interface
8. Settings page

#### Expected Results (Per Page)
- ✅ Layout adapts to screen size
- ✅ All functionality accessible
- ✅ No horizontal scrolling
- ✅ Buttons easily tappable
- ✅ Forms usable on mobile
- ✅ Navigation menu works
- ✅ Images load and display correctly
- ✅ Text readable without zooming

#### Actual Results
- [ ] Pass
- [ ] Fail (Describe issue): _______________

#### Page-Specific Issues
- Login: _______________
- Dashboard: _______________
- Candidates: _______________
- Forms: _______________
- Chat: _______________

#### Notes
_______________________________________________

---

### TC-016: Performance - Page Load Times

**Priority**: Medium  
**Type**: Performance  
**Estimated Time**: 10 minutes

#### Test Steps
Measure page load times (with cache cleared):

1. Login page
2. Dashboard
3. Candidates list (50+ candidates)
4. Candidate detail page
5. Create candidate form
6. CV generation

#### Expected Results
- ✅ Login page: < 2 seconds
- ✅ Dashboard: < 3 seconds
- ✅ Candidates list: < 4 seconds
- ✅ Candidate detail: < 2 seconds
- ✅ Create form: < 3 seconds
- ✅ CV generation: < 5 seconds

#### Actual Results (Record times)
- Login: _____ seconds
- Dashboard: _____ seconds
- Candidates list: _____ seconds
- Candidate detail: _____ seconds
- Create form: _____ seconds
- CV generation: _____ seconds

#### Performance Rating
- [ ] Excellent (all under expected)
- [ ] Good (most under expected)
- [ ] Needs Improvement (many over expected)

#### Notes
_______________________________________________

---

### TC-017: Error Handling & Validation

**Priority**: High  
**Type**: Functional  
**Estimated Time**: 10 minutes

#### Test Steps
Test various error scenarios:

1. **Form Validation**:
   - Submit empty registration form
   - Enter invalid email format
   - Password mismatch
   - Required fields missing in candidate form

2. **Authentication Errors**:
   - Login with wrong password
   - Login with non-existent email
   - Access protected page without login

3. **File Upload Errors**:
   - Upload file exceeding size limit
   - Upload unsupported file type
   - Upload corrupted file

4. **Network Errors**:
   - Simulate slow/failed API requests
   - Test offline behavior

#### Expected Results
- ✅ Clear error messages displayed
- ✅ Field-specific validation errors shown
- ✅ Error messages user-friendly (not technical)
- ✅ Forms don't submit with invalid data
- ✅ Unauthorized access redirects to login
- ✅ File size/type errors shown before upload
- ✅ Loading states shown during requests
- ✅ Network errors handled gracefully
- ✅ Retry mechanisms available

#### Actual Results
- [ ] Pass
- [ ] Fail (Describe issue): _______________

#### Error Message Quality
- [ ] Clear and helpful
- [ ] Too technical
- [ ] Insufficient information

#### Notes
_______________________________________________

---

### TC-018: Data Persistence & Recovery

**Priority**: High  
**Type**: Functional  
**Estimated Time**: 5 minutes

#### Test Steps
1. **Draft Auto-save**:
   - Start creating candidate
   - Fill partial information
   - Close browser tab
   - Reopen and navigate to create page
   - Verify data recovered

2. **Session Persistence**:
   - Login with "Remember me"
   - Close browser
   - Reopen and navigate to app
   - Verify still logged in

3. **Form State**:
   - Start filling long form
   - Navigate away
   - Return to form
   - Verify state preserved

#### Expected Results
- ✅ Draft data auto-saves every few seconds
- ✅ Draft recovered on page reload
- ✅ User notified of draft recovery
- ✅ "Remember me" keeps session active
- ✅ Form state preserved during navigation
- ✅ No data loss on browser refresh

#### Actual Results
- [ ] Pass
- [ ] Fail (Describe issue): _______________

#### Notes
_______________________________________________

---

### TC-019: Multi-Partner Workflow

**Priority**: High  
**Type**: Integration  
**Estimated Time**: 15 minutes

#### Preconditions
- 1 Ethiopian agent account
- 3 Foreign agent accounts
- All paired together

#### Test Steps
1. **Ethiopian Agent**:
   - Create candidate with partner-specific job details:
     - Partner A: Saudi Arabia, 1500 SAR
     - Partner B: UAE, 2000 AED
     - Partner C: Kuwait, 400 KWD
   - Publish to all partners

2. **Partner A**:
   - View candidate (verify Saudi Arabia job details shown)
   - Create selection request

3. **Partner B**:
   - View candidate (verify UAE job details shown)
   - Attempt selection (should be locked)

4. **Ethiopian Agent**:
   - View all selection requests
   - Approve Partner A's selection

5. **Verify**:
   - Partner B and C selections auto-rejected
   - Correct job details shown to each partner

#### Expected Results
- ✅ Candidate published to all partners
- ✅ Each partner sees their specific job details
- ✅ Selection lock prevents conflicts
- ✅ Only one selection can be approved
- ✅ Auto-rejection works correctly
- ✅ All notifications sent appropriately

#### Actual Results
- [ ] Pass
- [ ] Fail (Describe issue): _______________

#### Notes
_______________________________________________

---

### TC-020: Security & Access Control

**Priority**: Critical  
**Type**: Security  
**Estimated Time**: 10 minutes

#### Test Steps
1. **Role-based Access**:
   - Login as Ethiopian agent
   - Verify cannot access foreign agent features
   - Login as foreign agent
   - Verify cannot create candidates

2. **Data Isolation**:
   - Verify agents only see paired partners' data
   - Verify unpublished candidates not visible to foreign agents
   - Verify other agencies' data not accessible

3. **Session Security**:
   - Test concurrent logins on different devices
   - Test session timeout
   - Test "logout all devices"

4. **API Security**:
   - Test direct API access without auth token
   - Test accessing other users' resources

#### Expected Results
- ✅ Role permissions enforced
- ✅ Ethiopian agents cannot select candidates
- ✅ Foreign agents cannot create candidates
- ✅ Data properly isolated between agencies
- ✅ Unpublished candidates hidden from foreign agents
- ✅ Multiple sessions work correctly
- ✅ Sessions timeout after inactivity
- ✅ Logout all devices works
- ✅ API requires authentication
- ✅ Users cannot access others' resources

#### Actual Results
- [ ] Pass
- [ ] Fail (Describe issue): _______________

#### Security Issues Found
_______________________________________________

#### Notes
_______________________________________________

---

## Test Summary Report

### Test Execution Summary
- **Total Test Cases**: 20
- **Passed**: _____
- **Failed**: _____
- **Blocked**: _____
- **Not Executed**: _____

### Pass Rate
- **Percentage**: _____% 

### Critical Issues Found
1. _______________________________________________
2. _______________________________________________
3. _______________________________________________

### Major Issues Found
1. _______________________________________________
2. _______________________________________________
3. _______________________________________________

### Minor Issues Found
1. _______________________________________________
2. _______________________________________________

### Overall Assessment
- [ ] Ready for Production
- [ ] Ready with Minor Fixes
- [ ] Requires Major Fixes
- [ ] Not Ready

### Recommendations
_______________________________________________
_______________________________________________
_______________________________________________

### Tester Information
- **Name**: _____________________
- **Date**: _____________________
- **Environment**: _____________________
- **Browser**: _____________________
- **Signature**: _____________________

---

## Appendix: Test Data

### Sample Candidates
```
Candidate 1:
Name: Almaz Tadesse
DOB: 1995-05-15
Nationality: Ethiopian
Experience: 5 years (3 in Saudi Arabia)
Skills: Cooking, Cleaning, Childcare, Elderly Care
Languages: Arabic (Fluent), English (Intermediate), Amharic (Fluent)

Candidate 2:
Name: Tigist Bekele
DOB: 1992-08-22
Nationality: Ethiopian
Experience: 8 years (5 in UAE)
Skills: Cooking, Cleaning, Ironing, Laundry
Languages: Arabic (Fluent), English (Basic), Amharic (Fluent)
```

### File Samples Required
- Passport document (PDF/JPG)
- Full-body photo (JPG/PNG)
- Video interview (MP4)
- Agency logo (PNG with transparent background recommended)

---

**End of Test Case Document**
 