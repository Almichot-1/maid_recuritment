import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import adminApi from "@/lib/admin-api"
import {
  AccountStatus,
  AdminAgencyDetail,
  AdminAgencyLoginOverview,
  AdminAgencyLoginSummary,
  AdminAgencySummary,
  AdminAuditLogOverview,
  AdminCandidateOverview,
  AdminDashboardStats,
  AdminManagementRecord,
  AdminRole,
  AdminSelectionOverview,
  PlatformSettings,
  UserRole,
} from "@/types"

interface AdminAgencyFilters {
  status?: AccountStatus | "all"
  role?: UserRole | "all"
  search?: string
}

interface CreateAdminInput {
  email: string
  full_name: string
  role: AdminRole
}

interface UpdateAdminInput {
  id: string
  role?: AdminRole
  is_active?: boolean
  force_password_change?: boolean
}

export function useAdminDashboard() {
  return useQuery({
    queryKey: ["admin-dashboard"],
    queryFn: async () => {
      const response = await adminApi.get<AdminDashboardStats>("/admin/analytics/dashboard")
      return response.data
    },
  })
}

export function usePendingAgencies(role?: UserRole | "all") {
  return useQuery({
    queryKey: ["admin-pending-agencies", role ?? "all"],
    queryFn: async () => {
      const response = await adminApi.get<{ agencies: AdminAgencySummary[] }>("/admin/agencies/pending", {
        params: role && role !== "all" ? { role } : undefined,
      })
      return response.data.agencies
    },
  })
}

export function useAgencies(filters: AdminAgencyFilters) {
  return useQuery({
    queryKey: ["admin-agencies", filters],
    queryFn: async () => {
      const response = await adminApi.get<{ agencies: AdminAgencySummary[] }>("/admin/agencies", {
        params: {
          ...(filters.status && filters.status !== "all" ? { status: filters.status } : {}),
          ...(filters.role && filters.role !== "all" ? { role: filters.role } : {}),
          ...(filters.search ? { search: filters.search } : {}),
        },
      })
      return response.data.agencies
    },
  })
}

export function useAgency(id: string) {
  return useQuery({
    queryKey: ["admin-agency", id],
    queryFn: async () => {
      const response = await adminApi.get<{ agency: AdminAgencyDetail }>(`/admin/agencies/${id}`)
      return response.data.agency
    },
    enabled: Boolean(id),
  })
}

export function useApproveAgency() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (agencyId: string) => {
      await adminApi.post(`/admin/agencies/${agencyId}/approve`)
    },
    onSuccess: async (_, agencyId) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["admin-pending-agencies"] }),
        queryClient.invalidateQueries({ queryKey: ["admin-agencies"] }),
        queryClient.invalidateQueries({ queryKey: ["admin-dashboard"] }),
        queryClient.invalidateQueries({ queryKey: ["admin-agency", agencyId] }),
      ])
    },
  })
}

export function useRejectAgency() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ agencyId, reason, notes }: { agencyId: string; reason: string; notes?: string }) => {
      await adminApi.post(`/admin/agencies/${agencyId}/reject`, { reason, notes })
    },
    onSuccess: async (_, variables) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["admin-pending-agencies"] }),
        queryClient.invalidateQueries({ queryKey: ["admin-agencies"] }),
        queryClient.invalidateQueries({ queryKey: ["admin-dashboard"] }),
        queryClient.invalidateQueries({ queryKey: ["admin-agency", variables.agencyId] }),
      ])
    },
  })
}

export function useUpdateAgencyStatus() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ agencyId, status, reason }: { agencyId: string; status: AccountStatus; reason?: string }) => {
      await adminApi.patch(`/admin/agencies/${agencyId}/status`, { status, reason })
    },
    onSuccess: async (_, variables) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["admin-agencies"] }),
        queryClient.invalidateQueries({ queryKey: ["admin-dashboard"] }),
        queryClient.invalidateQueries({ queryKey: ["admin-agency", variables.agencyId] }),
      ])
    },
  })
}

export function useAdminCandidates(status?: string) {
  return useQuery({
    queryKey: ["admin-candidates", status ?? "all"],
    queryFn: async () => {
      const response = await adminApi.get<{ candidates: AdminCandidateOverview[] }>("/admin/candidates", {
        params: status ? { status } : undefined,
      })
      return response.data.candidates
    },
  })
}

export function useAdminSelections(status?: string) {
  return useQuery({
    queryKey: ["admin-selections", status ?? "all"],
    queryFn: async () => {
      const response = await adminApi.get<{ selections: AdminSelectionOverview[] }>("/admin/selections", {
        params: status ? { status } : undefined,
      })
      return response.data.selections
    },
  })
}

export function useAdminAuditLogs(filters?: { admin_id?: string; action?: string; target_type?: string }) {
  return useQuery({
    queryKey: ["admin-audit-logs", filters ?? {}],
    queryFn: async () => {
      const response = await adminApi.get<{ logs: AdminAuditLogOverview[] }>("/admin/audit-logs", {
        params: filters,
      })
      return response.data.logs
    },
  })
}

export function useAdminAgencyLogins(filters?: { role?: UserRole | "all"; search?: string }) {
  return useQuery({
    queryKey: ["admin-agency-logins", filters ?? {}],
    staleTime: 30_000,
    queryFn: async () => {
      const response = await adminApi.get<{ summary: AdminAgencyLoginSummary; logins: AdminAgencyLoginOverview[] }>(
        "/admin/agency-logins",
        {
          params: {
            ...(filters?.role && filters.role !== "all" ? { role: filters.role } : {}),
            ...(filters?.search ? { search: filters.search } : {}),
          },
        }
      )
      return response.data
    },
  })
}

export function useAdminUsers(enabled = true) {
  return useQuery({
    queryKey: ["admin-users"],
    enabled,
    queryFn: async () => {
      const response = await adminApi.get<{ admins: AdminManagementRecord[] }>("/admin/admins")
      return response.data.admins
    },
  })
}

export function useAdminPlatformSettings(enabled = true) {
  return useQuery({
    queryKey: ["admin-platform-settings"],
    enabled,
    queryFn: async () => {
      const response = await adminApi.get<{ settings: PlatformSettings }>("/admin/settings")
      return response.data.settings
    },
  })
}

export function useUpdateAdminPlatformSettings() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (payload: PlatformSettings) => {
      const response = await adminApi.patch<{ settings: PlatformSettings }>("/admin/settings", payload)
      return response.data.settings
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["admin-platform-settings"] }),
        queryClient.invalidateQueries({ queryKey: ["admin-audit-logs"] }),
      ])
    },
  })
}

export function useCreateAdminUser() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (payload: CreateAdminInput) => {
      const response = await adminApi.post("/admin/admins", payload)
      return response.data as {
        admin: AdminManagementRecord
        temporary_password: string
        mfa_secret: string
        provisioning_url: string
        invitation_warning?: string
      }
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["admin-users"] }),
        queryClient.invalidateQueries({ queryKey: ["admin-audit-logs"] }),
      ])
    },
  })
}

export function useUpdateAdminUser() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, ...payload }: UpdateAdminInput) => {
      const response = await adminApi.patch<{ admin: AdminManagementRecord }>(`/admin/admins/${id}`, payload)
      return response.data.admin
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["admin-users"] }),
        queryClient.invalidateQueries({ queryKey: ["admin-audit-logs"] }),
      ])
    },
  })
}
