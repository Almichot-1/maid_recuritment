import axios from "axios"
import { useMutation } from "@tanstack/react-query"
import { useRouter } from "next/navigation"
import { toast } from "sonner"
import adminApi from "@/lib/admin-api"
import { useAdminAuthStore } from "@/stores/admin-auth-store"
import { AdminRole, AdminUser } from "@/types"

interface AdminLoginInput {
  email: string
  password: string
  otp_code: string
}

interface AdminLoginResponse {
  admin: AdminUser
  expires_at: string
}

interface AdminChangePasswordInput {
  current_password: string
  new_password: string
}

export function useAdminLogin() {
  const router = useRouter()
  const setAuth = useAdminAuthStore((state) => state.setAuth)

  return useMutation({
    mutationFn: async (data: AdminLoginInput) => {
      const response = await adminApi.post<AdminLoginResponse>("/admin/login", data)
      return response.data
    },
    onSuccess: (data) => {
      setAuth(data.admin)
      toast.success("Admin session started")
      router.push("/admin/dashboard")
    },
    onError: (error) => {
      if (axios.isAxiosError<{ error?: string }>(error)) {
        const message = error.response?.data?.error
        toast.error(
          message === "admin setup required"
            ? "This admin account must finish the one-time setup link before it can sign in."
            : message || "Failed to start admin session"
        )
        return
      }
      toast.error("Failed to start admin session")
    },
  })
}

export function useAdminLogout() {
  const router = useRouter()
  const logout = useAdminAuthStore((state) => state.logout)

  return async () => {
    try {
      await adminApi.post("/admin/logout")
    } catch {
      // handled by interceptor when needed
    } finally {
      logout()
      router.push("/admin/login")
      toast.info("Admin signed out")
    }
  }
}

export function useCurrentAdmin() {
  const admin = useAdminAuthStore((state) => state.admin)

  return {
    admin,
    isSuperAdmin: admin?.role === AdminRole.SUPER_ADMIN,
    isSupportAdmin: admin?.role === AdminRole.SUPPORT_ADMIN,
  }
}

export function useAdminChangePassword() {
  return useMutation({
    mutationFn: async (data: AdminChangePasswordInput) => {
      try {
        const response = await adminApi.post<{ message: string }>("/admin/change-password", data)
        return response.data
      } catch (error) {
        if (axios.isAxiosError<{ error?: string }>(error)) {
          throw new Error(error.response?.data?.error || "Failed to change admin password")
        }
        throw error
      }
    },
  })
}
