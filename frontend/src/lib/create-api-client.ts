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
    timeout: 30_000,
  })

  client.interceptors.request.use(
    (config) => {
      if (config.data instanceof FormData) {
        config.timeout = 120_000
      }
      if (config.responseType === "blob") {
        config.timeout = 120_000
      }
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

  let isRetrying = false

  client.interceptors.response.use(
    (response) => response,
    async (error) => {
      const requestUrl = String(error.config?.url || "")
      const isLoginAttempt = requestUrl.includes("/auth/login") || requestUrl.includes("/admin/login")

      if (error.response?.status === 401 && typeof window !== "undefined" && !isLoginAttempt) {
        onUnauthorized?.()
        return Promise.reject(error)
      }

      if (error.response?.data?.account_status) {
        return Promise.reject(error)
      }

      if (!error.response && !isRetrying && error.config && !error.config._retryCount) {
        isRetrying = true
        const maxRetries = error.config.data instanceof FormData ? 1 : 2
        error.config._retryCount = (error.config._retryCount || 0) + 1
        if (error.config._retryCount <= maxRetries) {
          const delay = Math.min(1000 * 2 ** (error.config._retryCount - 1), 10_000)
          await new Promise((resolve) => setTimeout(resolve, delay))
          isRetrying = false
          return client(error.config)
        }
        isRetrying = false
      }

      let message = error.message || "An unexpected error occurred"

      if (error.code === "ECONNABORTED") {
        message = "Request timed out. Please check your connection and try again."
      } else if (!error.response) {
        message = "Network error. Please check your internet connection."
      }

      // Blob responses (e.g. file downloads) need special handling — the
      // response body is a Blob, not parsed JSON, so we must read it.
      if (error.config?.responseType === "blob" && error.response?.status >= 400) {
        const data = error.response.data
        if (data && typeof data === "object" && typeof data.text === "function") {
          try {
            const text = await data.text()
            const parsed = JSON.parse(text)
            if (parsed.error) message = parsed.error
            else if (parsed.message) message = parsed.message
            else if (parsed.detail) message = parsed.detail
          } catch {
            // Not JSON — keep the default message
          }
        }
      } else if (error.response?.data?.error) {
        message = error.response.data.error
      } else if (error.response?.data?.message) {
        message = error.response.data.message
      }

      // Skip global toast for known application-level errors that the
      // calling code handles itself (e.g. partner selection required).
      const isKnownAppError =
        error.response?.status === 409 &&
        error.response?.data?.requires_pairing_selection

      if (typeof window !== "undefined" && !isLoginAttempt && !isKnownAppError) {
        toast.error(message)
      }
      return Promise.reject(error)
    }
  )

  return client
}
