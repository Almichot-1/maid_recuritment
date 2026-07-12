"use client"

import { useNotificationRealtime } from "@/hooks/use-notification-realtime"

function NotificationRealtimeConnector() {
  useNotificationRealtime()
  return null
}

export function RealtimeProviders() {
  return (
    <>
      <NotificationRealtimeConnector />
    </>
  )
}
