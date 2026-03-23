import { create } from "zustand"
import { AdminUser } from "@/types"

interface AdminAuthState {
  admin: AdminUser | null
  token: string | null
  isAuthenticated: boolean
  isLoading: boolean
  setAuth: (admin: AdminUser, token: string) => void
  logout: () => void
  loadFromStorage: () => void
}

export const useAdminAuthStore = create<AdminAuthState>((set) => ({
  admin: null,
  token: null,
  isAuthenticated: false,
  isLoading: true,

  setAuth: (admin, token) => {
    localStorage.setItem("admin_auth_token", token)
    localStorage.setItem("admin_auth_user", JSON.stringify(admin))
    set({ admin, token, isAuthenticated: true, isLoading: false })
  },

  logout: () => {
    localStorage.removeItem("admin_auth_token")
    localStorage.removeItem("admin_auth_user")
    set({ admin: null, token: null, isAuthenticated: false, isLoading: false })
  },

  loadFromStorage: () => {
    try {
      if (typeof window === "undefined") {
        return
      }
      const token = localStorage.getItem("admin_auth_token")
      const adminString = localStorage.getItem("admin_auth_user")
      if (token && adminString) {
        const admin = JSON.parse(adminString) as AdminUser
        set({ admin, token, isAuthenticated: true, isLoading: false })
        return
      }
    } catch (error) {
      console.error("Failed to load admin auth state", error)
    }

    set({ admin: null, token: null, isAuthenticated: false, isLoading: false })
  },
}))
