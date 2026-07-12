import { useEffect, useRef, useCallback, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { buildWebSocketUrl } from '@/lib/api-base-url'
import { useAuthStore } from '@/stores/auth-store'

interface SelectionUpdateMessage {
  selection_id: string
  status: string
  updated_at: string
  action: string
  pairing_id: string
}

interface ProgressUpdateMessage {
  selection_id: string
  step_name: string
  status: string
  updated_at: string
}

/**
 * Hook for subscribing to real-time selection updates via WebSocket
 * Automatically updates React Query cache when selections change
 * 
 * Usage:
 * const { isConnected } = useSelectionUpdates(activePairingId);
 */
export function useSelectionUpdates(pairingId?: string) {
  const wsRef = useRef<WebSocket | null>(null)
  const queryClient = useQueryClient()
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const isConnectingRef = useRef(false)
  const [isConnected, setIsConnected] = useState(false)
  const retryDelayRef = useRef(1000)

  const MAX_RETRIES = 5
  const retryCountRef = useRef(0)

  const connect = useCallback(() => {
    // Don't reconnect if already connecting/connected
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
      if (pairingId) params.pairing_id = pairingId
      if (authToken) params.auth_token = authToken
      const wsUrl = buildWebSocketUrl("/selections/updates", params)

      const ws = new WebSocket(wsUrl)

      ws.onopen = () => {
        isConnectingRef.current = false
        retryCountRef.current = 0
        setIsConnected(true)
        retryDelayRef.current = 1000
        // Clear any pending reconnect attempts
        if (reconnectTimeoutRef.current) {
          clearTimeout(reconnectTimeoutRef.current)
          reconnectTimeoutRef.current = null
        }
      }

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)

          // Detect if this is a progress update (has step_name) or a selection update (has action)
          if (data.step_name) {
            const update: ProgressUpdateMessage = data;

            // Invalidate the progress query for this selection
            queryClient.invalidateQueries({
              queryKey: ['selection-progress', update.selection_id],
              exact: false,
            });
            queryClient.invalidateQueries({
              queryKey: ['my-selections'],
              exact: false,
            });
            return;
          }

          const update: SelectionUpdateMessage = data

          // If pairing filter is set, only process updates for that pairing
          if (pairingId && update.pairing_id !== pairingId) {
            return
          }

          // Update React Query cache instead of refetching
          queryClient.setQueryData(
            ['selection', update.selection_id],
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            (oldData: any) => {
              if (!oldData) return oldData
              return {
                ...oldData,
                status: update.status,
                updated_at: new Date(update.updated_at),
              }
            }
          )

          // Also invalidate the selections list
          queryClient.invalidateQueries({
            queryKey: ['my-selections'],
            exact: false,
          })
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
  }, [pairingId, queryClient])

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

  // Auto-connect when component mounts
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
