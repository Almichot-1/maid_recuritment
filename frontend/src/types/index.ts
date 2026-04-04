export enum UserRole {
  ETHIOPIAN_AGENT = 'ethiopian_agent',
  FOREIGN_AGENT = 'foreign_agent'
}

export enum AccountStatus {
  PENDING_APPROVAL = 'pending_approval',
  ACTIVE = 'active',
  REJECTED = 'rejected',
  SUSPENDED = 'suspended',
}

export interface User {
  id: string;
  email: string;
  email_verified: boolean;
  full_name: string;
  role: UserRole;
  company_name?: string;
  avatar_url?: string;
  auto_share_candidates?: boolean;
  default_foreign_pairing_id?: string | null;
  account_status: AccountStatus;
  current_session_id?: string;
}

export interface UserSession {
  id: string;
  device_label: string;
  browser_name: string;
  os_name: string;
  ip_address?: string;
  last_seen_at: string;
  expires_at: string;
}

export enum AdminRole {
  SUPER_ADMIN = 'super_admin',
  SUPPORT_ADMIN = 'support_admin',
}

export interface AdminUser {
  id: string;
  email: string;
  full_name: string;
  role: AdminRole;
  last_login?: string | null;
  force_password_change?: boolean;
}

export enum CandidateStatus {
  DRAFT = 'draft',
  AVAILABLE = 'available',
  LOCKED = 'locked',
  UNDER_REVIEW = 'under_review',
  APPROVED = 'approved',
  IN_PROGRESS = 'in_progress',
  COMPLETED = 'completed',
  REJECTED = 'rejected'
}

export interface Candidate {
  id: string;
  full_name: string;
  nationality?: string;
  date_of_birth?: string;
  age?: number;
  place_of_birth?: string;
  religion?: string;
  marital_status?: string;
  children_count?: number;
  education_level?: string;
  experience_years?: number;
  languages: string[];
  skills: string[];
  status: CandidateStatus;
  created_by: string;
  cv_pdf_url?: string;
  locked_by?: string;
  locked_at?: string;
  lock_expires_at?: string;
  documents: Document[];
  created_at: string;
  updated_at: string;
}

export interface Document {
  id: string;
  candidate_id?: string;
  document_type: string;
  file_url: string;
  file_name: string;
  file_size?: number;
  uploaded_at?: string;
}

export enum SelectionStatus {
  PENDING = 'pending',
  APPROVED = 'approved',
  REJECTED = 'rejected',
  EXPIRED = 'expired'
}

export interface SelectionCandidateSummary {
  id: string;
  full_name: string;
  status: CandidateStatus;
  created_by: string;
  age?: number;
  experience_years?: number;
  photo_url?: string;
}

export interface Selection {
  id: string;
  candidate_id: string;
  pairing_id?: string;
  selected_by: string;
  status: SelectionStatus;
  expires_at: string;
  created_at: string;
  updated_at?: string;
  time_remaining?: string;
  candidate?: SelectionCandidateSummary;
  ethiopian_approved?: boolean;
  foreign_approved?: boolean;
  employer_contract?: SelectionSupportingDocument;
  employer_id?: SelectionSupportingDocument;
  selected_by_name?: string;
}

export interface SelectionSupportingDocument {
  file_url: string;
  file_name: string;
  uploaded_at?: string;
}

export interface Approval {
  id: string;
  selection_id: string;
  user_id: string;
  decision: 'approved' | 'rejected';
  decided_at: string;
}

export interface StatusStep {
  id: string;
  candidate_id: string;
  step_name: string;
  step_status: 'pending' | 'in_progress' | 'completed' | 'failed';
  completed_at?: string;
  notes?: string;
  medical_document_url?: string;
  updated_at: string;
  updated_by: {
    id: string;
    name: string;
  };
}

export interface CandidateProgress {
  candidate_id: string;
  overall_status: CandidateStatus;
  steps: StatusStep[];
  progress_percentage: number;
  last_updated_at: string;
}

export interface PassportData {
  id: string;
  candidate_id: string;
  holder_name: string;
  passport_number: string;
  country_code?: string;
  nationality: string;
  date_of_birth: string;
  place_of_birth?: string;
  gender: string;
  issue_date?: string;
  expiry_date: string;
  issuing_authority?: string;
  mrz_line_1: string;
  mrz_line_2: string;
  confidence: number;
  extracted_at: string;
}

export interface Notification {
  id: string;
  title: string;
  message: string;
  type: string;
  is_read: boolean;
  related_entity_type?: string;
  related_entity_id?: string;
  created_at: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  meta: {
    page: number;
    page_size: number;
    total: number;
  };
}

export interface AdminDashboardStats {
  total_agencies: number;
  ethiopian_agencies: number;
  foreign_agencies: number;
  pending_approvals: number;
  total_candidates: number;
  active_selections: number;
  completed_recruitments: number;
  success_rate: number;
}

export interface AdminAgencySummary {
  id: string;
  company_name: string;
  contact_person: string;
  email: string;
  role: UserRole;
  account_status: AccountStatus;
  registration_date: string;
  total_candidates: number;
  total_selections: number;
}

export interface AdminAgencyActivitySummary {
  total_candidates: number;
  active_candidates: number;
  completed_recruitments: number;
  total_selections: number;
  approved_selections: number;
  active_recruitments: number;
}

export interface AdminRecentActivityItem {
  id: string;
  type: string;
  title: string;
  status: string;
  occurred_at: string;
}

export interface AdminAgencyDetail {
  agency: AdminAgencySummary;
  approval_status: string;
  rejection_reason?: string;
  admin_notes?: string;
  submitted_documents: Array<Record<string, string>>;
  activity_summary: AdminAgencyActivitySummary;
  recent_activity: AdminRecentActivityItem[];
}

export interface AdminCandidateOverview {
  id: string;
  full_name: string;
  age?: number | null;
  status: string;
  agency_id: string;
  agency_name: string;
  company_name: string;
  created_at: string;
}

export interface AdminSelectionOverview {
  id: string;
  candidate_id: string;
  candidate_name: string;
  ethiopian_agency: string;
  foreign_agency: string;
  status: string;
  selected_date: string;
  approval_status: string;
}

export interface AdminAuditLogOverview {
  id: string;
  admin_id: string;
  admin_name: string;
  action: string;
  target_type: string;
  target_id?: string;
  ip_address?: string;
  details: unknown;
  created_at: string;
}

export interface AdminAgencyLoginOverview {
  session_id: string;
  user_id: string;
  agency_name: string;
  contact_name: string;
  email: string;
  role: string;
  device_label: string;
  browser_name: string;
  os_name: string;
  ip_address?: string;
  logged_in_at: string;
  last_seen_at: string;
  is_active: boolean;
}

export interface AdminAgencyLoginSummary {
  total_login_events: number;
  active_sessions: number;
  ethiopian_login_events: number;
  foreign_login_events: number;
}

export interface AdminManagementRecord {
  id: string;
  email: string;
  full_name: string;
  role: AdminRole;
  is_active: boolean;
  failed_login_attempts: number;
  force_password_change: boolean;
  last_login?: string | null;
  locked_until?: string | null;
  created_at: string;
}

export interface PlatformSettings {
  id: string;
  selection_lock_duration_hours: number;
  require_both_approvals: boolean;
  auto_approve_agencies: boolean;
  auto_expire_selections: boolean;
  email_notifications_enabled: boolean;
  maintenance_mode: boolean;
  maintenance_message: string;
  agency_approval_email_template: string;
  agency_rejection_email_template: string;
  selection_notification_email_template: string;
  expiry_notification_email_template: string;
}

export enum AgencyPairingStatus {
  ACTIVE = "active",
  SUSPENDED = "suspended",
  ENDED = "ended",
}

export interface PairingAgencySummary {
  id: string;
  full_name: string;
  company_name: string;
  email: string;
  role: UserRole;
}

export interface WorkspaceSummary {
  id: string;
  status: AgencyPairingStatus | string;
  ethiopian_agency: PairingAgencySummary;
  foreign_agency: PairingAgencySummary;
  partner_agency: PairingAgencySummary;
  approved_at?: string;
  notes?: string;
}

export interface PairingContext {
  current_user_role: UserRole | string;
  has_active_pairs: boolean;
  active_pairing_id?: string;
  workspaces: WorkspaceSummary[];
}

export interface CandidatePairShare {
  id: string;
  pairing_id: string;
  shared_at: string;
  is_active: boolean;
  partner_agency: PairingAgencySummary;
  workspace: WorkspaceSummary;
}

export interface AdminPairing {
  id: string;
  status: AgencyPairingStatus | string;
  ethiopian_agency: PairingAgencySummary;
  foreign_agency: PairingAgencySummary;
  approved_at?: string;
  notes?: string;
}

export interface SmartAlertSelection {
  selection_id: string;
  candidate_id: string;
  candidate_name: string;
  expires_at: string;
  warning_level: string;
  remaining_label: string;
}

export interface SmartAlertPassport {
  candidate_id: string;
  candidate_name: string;
  passport_number: string;
  expiry_date: string;
  warning_level: string;
}

export interface SmartAlertMedical {
  candidate_id: string;
  candidate_name: string;
  expiry_date: string;
  warning_level: string;
}

export interface SmartAlertFlightUpdate {
  candidate_id: string;
  candidate_name: string;
  stage: string;
  updated_at: string;
  status: string;
}

export interface DashboardSmartAlerts {
  expiring_selections: SmartAlertSelection[];
  expiring_passports: SmartAlertPassport[];
  expiring_medicals: SmartAlertMedical[];
  flight_updates: SmartAlertFlightUpdate[];
  recently_arrived: SmartAlertFlightUpdate[];
}
