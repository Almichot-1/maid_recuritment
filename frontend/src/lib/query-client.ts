import { AxiosError } from "axios"
import { QueryClient } from "@tanstack/react-query"

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      gcTime: 5 * 60_000,
      refetchOnWindowFocus: false,
      retry: (failureCount, error) => {
        const status = (error as AxiosError)?.response?.status
        if (status === 401 || status === 403 || status === 404) {
          return false
        }

        return failureCount < 1
      },
    },
    mutations: {
      retry: 0,
    },
  },
})

export function clearQueryCache() {
  queryClient.clear()
}
