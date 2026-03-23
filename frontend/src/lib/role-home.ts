import { UserRole } from "@/types"

export function getRoleHomePath(role?: UserRole | null) {
  if (role === UserRole.ETHIOPIAN_AGENT) {
    return "/dashboard/agency"
  }

  if (role === UserRole.FOREIGN_AGENT) {
    return "/dashboard/employer"
  }

  return "/dashboard"
}

export function getRoleHomeLabel(role?: UserRole | null) {
  if (role === UserRole.ETHIOPIAN_AGENT) {
    return "Agency Home"
  }

  if (role === UserRole.FOREIGN_AGENT) {
    return "Employer Home"
  }

  return "Dashboard"
}

export function isRoleHomePath(pathname: string) {
  return pathname === "/dashboard" || pathname.startsWith("/dashboard/agency") || pathname.startsWith("/dashboard/employer")
}
