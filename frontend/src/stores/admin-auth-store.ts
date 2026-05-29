import { create } from "zustand"
import { AdminUser } from "@/types"
import { getApiBaseUrl } from "@/lib/api-base-url"
import { clearPersistedAdminUser, persistAdminUser } from "@/lib/auth-storage"
import { clearQueryCache } from "@/lib/query-client"

interface AdminMeResponse {
  admin: AdminUser
}

interface AdminAuthState {
  admin: AdminUser | null
  token: string | null
  isAuthenticated: boolean
  isLoading: boolean
  setAuth: (admin: AdminUser, token: string) => void
  logout: () => void
  loadFromStorage: () => Promise<void>
}

export const useAdminAuthStore = create<AdminAuthState>((set) => ({
  admin: null,
  token: null,
  isAuthenticated: false,
  isLoading: true,

  setAuth: (admin, token) => {
    persistAdminUser(admin)
    set({ admin, token, isAuthenticated: true, isLoading: false })
  },

  logout: () => {
    clearPersistedAdminUser()
    clearQueryCache()
    set({ admin: null, token: null, isAuthenticated: false, isLoading: false })
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

      if (response.status === 401 || response.status === 403) {
        clearPersistedAdminUser()
        set({ admin: null, token: null, isAuthenticated: false, isLoading: false })
        return
      }

      if (!response.ok) {
        set({ isLoading: false })
        return
      }

      const data = (await response.json()) as AdminMeResponse
      persistAdminUser(data.admin)
      set({ admin: data.admin, token: null, isAuthenticated: true, isLoading: false })
    } catch (error) {
      console.error("Failed to load admin auth state", error)
      set({ isLoading: false })
    }
  },
}))
