const defaultApiBaseUrl = "http://localhost:8080/api/v1"

export function getApiBaseUrl() {
  return (process.env.NEXT_PUBLIC_API_URL || defaultApiBaseUrl).replace(/\/+$/, "")
}

export function buildWebSocketUrl(path: string, queryParams?: Record<string, string>): string {
  const apiUrl = new URL(getApiBaseUrl())
  apiUrl.protocol = apiUrl.protocol === "https:" ? "wss:" : "ws:"
  const trimmedPath = apiUrl.pathname.replace(/\/+$/, "")
  const withoutApiPrefix = trimmedPath.replace(/\/api\/v\d+$/i, "")
  apiUrl.pathname = `${withoutApiPrefix}${path}`
  apiUrl.search = ""
  if (queryParams) {
    for (const [key, value] of Object.entries(queryParams)) {
      apiUrl.searchParams.set(key, value)
    }
  }
  return apiUrl.toString()
}
