"use client"

export const SESSION_EVENT = "browser-sessions-updated"

export type BrowserSession = {
  id: string
  user_id: string
  browser_name: string
  os_name: string
  device_label: string
  created_at: string
  last_active_at: string
}

export function getSessionsStorageKey(userID?: string) {
  return userID ? `browser_sessions:${userID}` : null
}

export function getCurrentSessionStorageKey(userID?: string) {
  return userID ? `browser_session_current:${userID}` : null
}

export function emitSessionsUpdate() {
  if (typeof window !== "undefined") {
    window.dispatchEvent(new Event(SESSION_EVENT))
  }
}

export function readSessions(userID?: string): BrowserSession[] {
  const key = getSessionsStorageKey(userID)
  if (!key || typeof window === "undefined") {
    return []
  }

  try {
    const stored = localStorage.getItem(key)
    return stored ? (JSON.parse(stored) as BrowserSession[]) : []
  } catch {
    return []
  }
}

export function writeSessions(userID: string, sessions: BrowserSession[]) {
  const key = getSessionsStorageKey(userID)
  if (!key || typeof window === "undefined") {
    return
  }

  localStorage.setItem(key, JSON.stringify(sessions))
  emitSessionsUpdate()
}

export function getCurrentSessionID(userID?: string) {
  const key = getCurrentSessionStorageKey(userID)
  if (!key || typeof window === "undefined") {
    return null
  }

  return sessionStorage.getItem(key)
}

export function setCurrentSessionID(userID: string, sessionID: string) {
  const key = getCurrentSessionStorageKey(userID)
  if (!key || typeof window === "undefined") {
    return
  }

  sessionStorage.setItem(key, sessionID)
}

function detectBrowser(userAgent: string) {
  const ua = userAgent.toLowerCase()
  if (ua.includes("edg")) return "Microsoft Edge"
  if (ua.includes("chrome")) return "Google Chrome"
  if (ua.includes("firefox")) return "Firefox"
  if (ua.includes("safari")) return "Safari"
  return "Browser"
}

function detectOS(userAgent: string) {
  const ua = userAgent.toLowerCase()
  if (ua.includes("windows")) return "Windows"
  if (ua.includes("android")) return "Android"
  if (ua.includes("iphone") || ua.includes("ipad") || ua.includes("mac os")) {
    return "Apple"
  }
  if (ua.includes("linux")) return "Linux"
  return "Device"
}

function buildSessionMeta(userID: string): BrowserSession {
  const userAgent = typeof window !== "undefined" ? window.navigator.userAgent : ""
  const browser = detectBrowser(userAgent)
  const os = detectOS(userAgent)
  const now = new Date().toISOString()

  return {
    id: crypto.randomUUID(),
    user_id: userID,
    browser_name: browser,
    os_name: os,
    device_label: `${browser} on ${os}`,
    created_at: now,
    last_active_at: now,
  }
}

export function ensureBrowserSession(userID?: string) {
  if (!userID || typeof window === "undefined") {
    return
  }

  const sessions = readSessions(userID)
  const currentSessionID = getCurrentSessionID(userID)
  const now = new Date().toISOString()

  if (currentSessionID) {
    const updated = sessions.map((session) =>
      session.id === currentSessionID
        ? { ...session, last_active_at: now }
        : session,
    )

    if (updated.some((session) => session.id === currentSessionID)) {
      writeSessions(userID, updated)
      return
    }
  }

  const nextSession = buildSessionMeta(userID)
  setCurrentSessionID(userID, nextSession.id)
  writeSessions(userID, [nextSession, ...sessions].slice(0, 10))
}

export function removeCurrentBrowserSession(userID?: string) {
  if (!userID || typeof window === "undefined") {
    return
  }

  const currentSessionID = getCurrentSessionID(userID)
  if (!currentSessionID) {
    return
  }

  const remaining = readSessions(userID).filter(
    (session) => session.id !== currentSessionID,
  )
  writeSessions(userID, remaining)

  const key = getCurrentSessionStorageKey(userID)
  if (key) {
    sessionStorage.removeItem(key)
  }
}

export function clearBrowserSessions(userID?: string) {
  const key = getSessionsStorageKey(userID)
  const currentKey = getCurrentSessionStorageKey(userID)
  if (!key || typeof window === "undefined") {
    return
  }

  localStorage.removeItem(key)
  if (currentKey) {
    sessionStorage.removeItem(currentKey)
  }
  emitSessionsUpdate()
}

export function removeBrowserSession(userID: string, sessionID: string) {
  const remaining = readSessions(userID).filter((session) => session.id !== sessionID)
  writeSessions(userID, remaining)
}
