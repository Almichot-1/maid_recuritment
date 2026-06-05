import { useEffect, useRef, useCallback } from 'react'
import { useQueryClient } from '@tanstack/react-query'

interface SelectionUpdateMessage {
  selection_id: string
  status: string
  updated_at: string
  action: string
  pairing_id: string
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

  const connect = useCallback(() => {
    // Don't reconnect if already connecting/connected
    if (isConnectingRef.current || wsRef.current?.readyState === WebSocket.OPEN) {
      return
    }

    isConnectingRef.current = true

    try {
      // Determine WebSocket URL based on environment
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const host = window.location.host
      const wsUrl = `${protocol}//${host}/api/selections/updates`

      const ws = new WebSocket(wsUrl)

      ws.onopen = () => {
        console.log('[SelectionUpdates] WebSocket connected')
        isConnectingRef.current = false
        // Clear any pending reconnect attempts
        if (reconnectTimeoutRef.current) {
          clearTimeout(reconnectTimeoutRef.current)
          reconnectTimeoutRef.current = null
        }
      }

      ws.onmessage = (event) => {
        try {
          const update: SelectionUpdateMessage = JSON.parse(event.data)

          // If pairing filter is set, only process updates for that pairing
          if (pairingId && update.pairing_id !== pairingId) {
            return
          }

          console.log('[SelectionUpdates] Received update:', update)

          // Update React Query cache instead of refetching
          // This updates the selection in cache without making a network request
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

          // Also invalidate the selections list to reflect status changes
          // The invalidation is silent (no refetch) if the data is still fresh
          queryClient.invalidateQueries({
            queryKey: ['my-selections'],
            exact: false,
          })
        } catch (error) {
          console.error('[SelectionUpdates] Failed to parse message:', error)
        }
      }

      ws.onerror = (error) => {
        console.error('[SelectionUpdates] WebSocket error:', error)
        isConnectingRef.current = false
      }

      ws.onclose = () => {
        console.log('[SelectionUpdates] WebSocket disconnected')
        isConnectingRef.current = false

        // Attempt to reconnect after 3 seconds
        reconnectTimeoutRef.current = setTimeout(() => {
          console.log('[SelectionUpdates] Attempting to reconnect...')
          connect()
        }, 3000)
      }

      wsRef.current = ws
    } catch (error) {
      console.error('[SelectionUpdates] Failed to create WebSocket:', error)
      isConnectingRef.current = false

      // Attempt to reconnect
      reconnectTimeoutRef.current = setTimeout(connect, 3000)
    }
  }, [queryClient, pairingId])

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
      reconnectTimeoutRef.current = null
    }

    if (wsRef.current) {
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
    isConnected: wsRef.current?.readyState === WebSocket.OPEN,
    disconnect,
    reconnect: connect,
  }
}
