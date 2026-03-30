import { create } from "zustand"
import { AdminUser } from "@/types"
import { getApiBaseUrl } from "@/lib/api-base-url"

interface AdminMeResponse {
  admin: AdminUser
}

interface AdminAuthState {
  admin: AdminUser | null
  isAuthenticated: boolean
  isLoading: boolean
  setAuth: (admin: AdminUser) => void
  logout: () => void
  loadFromStorage: () => Promise<void>
}

export const useAdminAuthStore = create<AdminAuthState>((set) => ({
  admin: null,
  isAuthenticated: false,
  isLoading: true,

  setAuth: (admin) => {
    localStorage.setItem("admin_auth_user", JSON.stringify(admin))
    localStorage.removeItem("admin_auth_token")
    set({ admin, isAuthenticated: true, isLoading: false })
  },

  logout: () => {
    localStorage.removeItem("admin_auth_token")
    localStorage.removeItem("admin_auth_user")
    set({ admin: null, isAuthenticated: false, isLoading: false })
  },

  loadFromStorage: async () => {
    try {
      if (typeof window === "undefined") {
        return
      }

      set({ isLoading: true })

      const response = await fetch(`${getApiBaseUrl()}/admin/me`, {
        credentials: "include",
      })
      if (!response.ok) {
        throw new Error(`admin session check failed with status ${response.status}`)
      }

      const data = (await response.json()) as AdminMeResponse
      localStorage.setItem("admin_auth_user", JSON.stringify(data.admin))
      localStorage.removeItem("admin_auth_token")
      set({ admin: data.admin, isAuthenticated: true, isLoading: false })
      return
    } catch (error) {
      console.error("Failed to load admin auth state", error)
      localStorage.removeItem("admin_auth_token")
      localStorage.removeItem("admin_auth_user")
    }

    set({ admin: null, isAuthenticated: false, isLoading: false })
  },
}))
