import axios from "axios"
import { toast } from "sonner"
import { usePairingStore } from "@/stores/pairing-store"
import { getApiBaseUrl } from "@/lib/api-base-url"

interface CreateApiClientOptions {
  includePairingHeader?: boolean
  onUnauthorized?: () => void
}

export function createApiClient({ includePairingHeader = false, onUnauthorized }: CreateApiClientOptions) {
  const client = axios.create({
    baseURL: getApiBaseUrl(),
    withCredentials: true,
  })

  client.interceptors.request.use(
    (config) => {
      if (typeof window !== "undefined" && includePairingHeader && config.headers) {
        const pairingState = usePairingStore.getState()
        if (pairingState.isReady && pairingState.activePairingId) {
          const pairingId = pairingState.activePairingId
          config.headers["X-Pairing-ID"] = pairingId
        } else if ("X-Pairing-ID" in config.headers) {
          delete config.headers["X-Pairing-ID"]
        }
      }
      return config
    },
    (error) => Promise.reject(error)
  )

  client.interceptors.response.use(
    (response) => response,
    (error) => {
      const requestUrl = String(error.config?.url || "")
      const isLoginAttempt = requestUrl.includes("/auth/login") || requestUrl.includes("/admin/login")

      if (error.response?.status === 401 && typeof window !== "undefined" && !isLoginAttempt) {
        onUnauthorized?.()
      } else if (error.response?.data?.account_status) {
        // Let the calling hook present a more contextual account-state message.
      } else {
        const message =
          error.response?.data?.error ||
          error.response?.data?.message ||
          error.message ||
          "An unexpected error occurred"
        if (typeof window !== "undefined") {
          toast.error(message)
        }
      }
      return Promise.reject(error)
    }
  )

  return client
}
