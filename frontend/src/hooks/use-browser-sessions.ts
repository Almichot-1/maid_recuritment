"use client"

import * as React from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import axios from "axios"
import { toast } from "sonner"

import api from "@/lib/api"
import { useCurrentUser } from "@/hooks/use-auth"
import { UserSession } from "@/types"

interface UserSessionsResponse {
  sessions: UserSession[]
  current_session_id?: string
}

interface ApiErrorResponse {
  error?: string
  message?: string
}

export function useBrowserSessions() {
  const { user } = useCurrentUser()
  const queryClient = useQueryClient()

  const { data } = useQuery({
    queryKey: ["active-sessions", user?.id],
    queryFn: async () => {
      const response = await api.get<UserSessionsResponse>("/users/sessions")
      return response.data
    },
    enabled: Boolean(user?.id),
    staleTime: 30_000,
  })

  const revokeSession = useMutation({
    mutationFn: async (sessionID: string) => {
      await api.delete(`/users/sessions/${sessionID}`)
      return sessionID
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["active-sessions", user?.id] })
      toast.success("Session removed successfully")
    },
    onError: (error: unknown) => {
      const message = axios.isAxiosError<ApiErrorResponse>(error)
        ? error.response?.data?.error || error.response?.data?.message || error.message
        : error instanceof Error
          ? error.message
          : "Failed to remove session"
      toast.error(message)
    },
  })

  const orderedSessions = React.useMemo(() => {
    const sessions = data?.sessions || []
    const currentSessionID = data?.current_session_id
    return [...sessions].sort((left, right) => {
      if (left.id === currentSessionID) return -1
      if (right.id === currentSessionID) return 1
      return new Date(right.last_seen_at).getTime() - new Date(left.last_seen_at).getTime()
    })
  }, [data?.current_session_id, data?.sessions])

  return {
    sessions: orderedSessions,
    currentSessionID: data?.current_session_id || user?.current_session_id || null,
    removeSession: (sessionID: string) => revokeSession.mutate(sessionID),
  }
}
