import { useEffect, useRef, useCallback, useState } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { buildWebSocketUrl } from "@/lib/api-base-url"
import { Notification } from "@/types"
import { useAuthStore } from "@/stores/auth-store"

interface NotificationPushEvent {
  notification: Notification
}

export function useNotificationRealtime() {
  const wsRef = useRef<WebSocket | null>(null)
  const queryClient = useQueryClient()
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const isConnectingRef = useRef(false)
  const [isConnected, setIsConnected] = useState(false)
  const retryDelayRef = useRef(1000)

  const MAX_RETRIES = 5
  const retryCountRef = useRef(0)

  const connect = useCallback(() => {
    if (isConnectingRef.current || wsRef.current?.readyState === WebSocket.OPEN) {
      return
    }

    if (retryCountRef.current >= MAX_RETRIES) {
      return
    }

    isConnectingRef.current = true
    retryCountRef.current += 1

    try {
      const authToken = useAuthStore.getState().authToken
      const params: Record<string, string> = {}
      if (authToken) params.auth_token = authToken
      const wsUrl = buildWebSocketUrl("/ws/notifications", params)
      const ws = new WebSocket(wsUrl)

      ws.onopen = () => {
        isConnectingRef.current = false
        retryCountRef.current = 0
        setIsConnected(true)
        retryDelayRef.current = 1000
        if (reconnectTimeoutRef.current) {
          clearTimeout(reconnectTimeoutRef.current)
          reconnectTimeoutRef.current = null
        }
      }

      ws.onmessage = (event) => {
        try {
          const data: NotificationPushEvent = JSON.parse(event.data)
          if (!data.notification) return

          const incoming = data.notification

          queryClient.setQueryData<Notification[]>(["notifications"], (old) => {
            if (!old) return [incoming]
            return [incoming, ...old]
          })

          queryClient.invalidateQueries({ queryKey: ["notifications", "summary"], exact: false })

          // Invalidate progress queries on progress-related notifications
          const progressTypes = ["status_update", "flight_booked", "arrived"];
          if (progressTypes.includes(incoming.type)) {
            queryClient.invalidateQueries({ queryKey: ["selection-progress"], exact: false });
          }
        } catch {
          // ignore parse errors
        }
      }

      ws.onerror = () => {
        isConnectingRef.current = false
      }

      ws.onclose = () => {
        isConnectingRef.current = false
        setIsConnected(false)

        if (retryCountRef.current < MAX_RETRIES) {
          reconnectTimeoutRef.current = setTimeout(() => {
            retryDelayRef.current = Math.min(retryDelayRef.current * 2, 30000)
            connect()
          }, retryDelayRef.current)
        }
      }

      wsRef.current = ws
    } catch {
      isConnectingRef.current = false
      if (retryCountRef.current < MAX_RETRIES) {
        reconnectTimeoutRef.current = setTimeout(() => {
          retryDelayRef.current = Math.min(retryDelayRef.current * 2, 30000)
          connect()
        }, retryDelayRef.current)
      }
    }
  }, [queryClient])

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
      reconnectTimeoutRef.current = null
    }
    if (wsRef.current && wsRef.current.readyState !== WebSocket.CLOSED && wsRef.current.readyState !== WebSocket.CLOSING) {
      wsRef.current.close()
      wsRef.current = null
    }
    isConnectingRef.current = false
  }, [])

  useEffect(() => {
    connect()
    return () => {
      disconnect()
    }
  }, [connect, disconnect])

  return {
    isConnected,
    disconnect,
    reconnect: connect,
  }
}
