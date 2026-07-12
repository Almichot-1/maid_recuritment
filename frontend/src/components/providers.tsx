"use client"

import * as React from "react"
import { AxiosError } from "axios"
import { ThemeProvider as NextThemesProvider, type ThemeProviderProps } from "next-themes"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { STALE_TIMES, GC_TIMES, RETRY_STRATEGY } from "@/lib/query-constants"

export function ThemeProvider({ children, ...props }: ThemeProviderProps) {
  return <NextThemesProvider {...props}>{children}</NextThemesProvider>
}

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: STALE_TIMES.MEDIUM,
      gcTime: GC_TIMES.DEFAULT,
      refetchOnWindowFocus: false,
      refetchOnReconnect: true,
      networkMode: "online",
      retry: (failureCount, error) => {
        const status = (error as AxiosError)?.response?.status
        if (status === 401 || status === 403 || status === 404 || status === 422) {
          return false
        }
        if ((error as AxiosError)?.code === "ECONNABORTED") {
          return failureCount < RETRY_STRATEGY.MAX_NETWORK_RETRIES
        }
        if (!(error as AxiosError)?.response) {
          return failureCount < RETRY_STRATEGY.MAX_NETWORK_RETRIES
        }
        return failureCount < RETRY_STRATEGY.MAX_SERVER_RETRIES
      },
      retryDelay: (attemptIndex) => Math.min(RETRY_STRATEGY.NETWORK_BACKOFF_MS * 2 ** attemptIndex, RETRY_STRATEGY.MAX_BACKOFF_MS),
    },
    mutations: {
      retry: 1,
      retryDelay: RETRY_STRATEGY.NETWORK_BACKOFF_MS,
    },
  },
})

export function QueryProvider({ children }: { children: React.ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  )
}
