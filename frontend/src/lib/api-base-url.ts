const defaultApiBaseUrl = "http://localhost:8080/api/v1"

export function getApiBaseUrl() {
  return (process.env.NEXT_PUBLIC_API_URL || defaultApiBaseUrl).replace(/\/+$/, "")
}
