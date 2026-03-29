"use client"

import * as React from "react"

import { useCurrentUser } from "@/hooks/use-auth"
import {
  BrowserSession,
  SESSION_EVENT,
  clearBrowserSessions,
  ensureBrowserSession,
  getCurrentSessionID,
  getSessionsStorageKey,
  readSessions,
  removeBrowserSession,
} from "@/lib/browser-sessions"

export function useBrowserSessions() {
  const { user } = useCurrentUser()
  const [sessions, setSessions] = React.useState<BrowserSession[]>([])
  const currentSessionID = user?.id ? getCurrentSessionID(user.id) : null

  React.useEffect(() => {
    if (!user?.id) {
      setSessions([])
      return
    }

    ensureBrowserSession(user.id)
    setSessions(readSessions(user.id))
  }, [user?.id])

  React.useEffect(() => {
    if (typeof window === "undefined" || !user?.id) {
      return
    }

    const syncSessions = () => {
      setSessions(readSessions(user.id))
    }

    const onStorage = (event: StorageEvent) => {
      const key = getSessionsStorageKey(user.id)
      if (!key || event.key === key) {
        syncSessions()
      }
    }

    const touch = () => {
      ensureBrowserSession(user.id)
      syncSessions()
    }

    const interval = window.setInterval(touch, 60_000)
    window.addEventListener("focus", touch)
    window.addEventListener("storage", onStorage)
    window.addEventListener(SESSION_EVENT, syncSessions)

    return () => {
      window.clearInterval(interval)
      window.removeEventListener("focus", touch)
      window.removeEventListener("storage", onStorage)
      window.removeEventListener(SESSION_EVENT, syncSessions)
    }
  }, [user?.id])

  const orderedSessions = React.useMemo(() => {
    return [...sessions].sort((left, right) => {
      if (left.id === currentSessionID) return -1
      if (right.id === currentSessionID) return 1
      return new Date(right.last_active_at).getTime() - new Date(left.last_active_at).getTime()
    })
  }, [currentSessionID, sessions])

  return {
    sessions: orderedSessions,
    currentSessionID,
    clearAllSessions: () => clearBrowserSessions(user?.id),
    removeSession: (sessionID: string) => {
      if (!user?.id) {
        return
      }
      removeBrowserSession(user.id, sessionID)
      setSessions(readSessions(user.id))
    },
  }
}
