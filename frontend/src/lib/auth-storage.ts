import { AdminUser, User } from "@/types"

/** Session-scoped cache only; cookie session is validated via /auth/me. Email is omitted to reduce persisted PII. */
type StoredAgencyUser = Pick<User, "id" | "role" | "account_status" | "full_name" | "avatar_url" | "company_name">

/** Admin snapshot kept minimal for the same reason. */
type StoredAdminUser = Pick<AdminUser, "id" | "role" | "full_name">

const AGENCY_USER_KEY = "auth_user"
const ADMIN_USER_KEY = "admin_auth_user"

function canUseSessionStorage() {
  return typeof window !== "undefined"
}

export function persistAgencyUser(user: User) {
  if (!canUseSessionStorage()) {
    return
  }

  const snapshot: StoredAgencyUser = {
    id: user.id,
    role: user.role,
    account_status: user.account_status,
    full_name: user.full_name,
    avatar_url: user.avatar_url,
    company_name: user.company_name,
  }

  sessionStorage.setItem(AGENCY_USER_KEY, JSON.stringify(snapshot))
}

export function persistAdminUser(admin: AdminUser) {
  if (!canUseSessionStorage()) {
    return
  }

  const snapshot: StoredAdminUser = {
    id: admin.id,
    role: admin.role,
    full_name: admin.full_name,
  }

  sessionStorage.setItem(ADMIN_USER_KEY, JSON.stringify(snapshot))
}

export function clearPersistedAgencyUser() {
  if (!canUseSessionStorage()) {
    return
  }

  sessionStorage.removeItem(AGENCY_USER_KEY)
  localStorage.removeItem(AGENCY_USER_KEY)
  localStorage.removeItem("auth_token")
}

export function clearPersistedAdminUser() {
  if (!canUseSessionStorage()) {
    return
  }

  sessionStorage.removeItem(ADMIN_USER_KEY)
  localStorage.removeItem(ADMIN_USER_KEY)
  localStorage.removeItem("admin_auth_token")
}
