const DEV_API_BASE_URL = "http://localhost:8080/api/v1"

export function getApiBaseUrl() {
  const configured = process.env.NEXT_PUBLIC_API_URL?.trim()

  if (configured) {
    return configured.replace(/\/+$/, "")
  }

  if (process.env.NODE_ENV === "production") {
    throw new Error(
      "NEXT_PUBLIC_API_URL must be set for production builds and deployments.",
    )
  }

  return DEV_API_BASE_URL
}
