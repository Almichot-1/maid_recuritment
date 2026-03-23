import { useAdminAuthStore } from "@/stores/admin-auth-store"
import { createApiClient } from "@/lib/create-api-client"

const adminApi = createApiClient({
  tokenKey: "admin_auth_token",
  onUnauthorized: () => {
    useAdminAuthStore.getState().logout()
    window.location.href = "/admin/login"
  },
})

export default adminApi
