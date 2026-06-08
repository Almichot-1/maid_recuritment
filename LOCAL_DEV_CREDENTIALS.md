# Local Development Test Credentials

## Test User Accounts

### Ethiopian Agent (Creates Candidates)
- **Email**: `ethiopian@test.com`
- **Password**: `password123`
- **Role**: Ethiopian Agent
- **Company**: Ethiopian Recruitment Agency
- **Permissions**: 
  - Create and manage candidates
  - Upload documents
  - View own candidates
  - Receive selection notifications

### Foreign Agent (Selects Candidates)
- **Email**: `foreign@test.com`
- **Password**: `password123`
- **Role**: Foreign Agent
- **Company**: International Recruitment Co
- **Permissions**:
  - Browse available candidates
  - Select/lock candidates for review
  - Approve/reject candidates
  - Download CVs

## Quick Start

1. **Start Frontend** (if not already running):
   ```powershell
   cd frontend
   npm run dev
   ```

2. **Open Browser**:
   ```
   http://localhost:3000
   ```

3. **Login** with either account above

## Backend Status

- ✅ Backend API: `http://localhost:8080` (already running)
- ✅ Database: Docker PostgreSQL on port 5432 (already running)
- ✅ Migrations: All applied successfully
- ✅ Agency Pairing: Active between both test users
- ✅ Auto-Share: Enabled for Ethiopian agent (candidates automatically visible to foreign agent)

## Testing Workflow

1. **Login as Ethiopian Agent** → Create candidates with documents
2. **Publish candidates** to make them available
3. **Logout** and **login as Foreign Agent** → Browse and select candidates
4. **Lock and review** candidates
5. **Approve/reject** selections

## Notes

- Both accounts use the same password: `password123`
- These are test accounts for local development only
- Password is intentionally simple for easy testing
- Change passwords before deploying to any environment
