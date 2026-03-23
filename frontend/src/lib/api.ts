import { useAuthStore } from "@/stores/auth-store"
import { createApiClient } from "@/lib/create-api-client"

const api = createApiClient({
  tokenKey: "auth_token",
  onUnauthorized: () => {
    useAuthStore.getState().logout()
    window.location.href = "/login"
  },
})

export default api
