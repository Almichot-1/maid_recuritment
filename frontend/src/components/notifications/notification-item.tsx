"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { format, isToday, isYesterday } from "date-fns"
import { Check, Eye } from "lucide-react"

import { Notification } from "@/types"
import { useMarkAsRead } from "@/hooks/use-notifications"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { getNotificationIcon, getNotificationColor } from "@/lib/notification-utils"

interface NotificationItemProps {
  notification: Notification
}

function formatNotificationTime(createdAt: string) {
  const date = new Date(createdAt)
  if (isToday(date)) return format(date, "h:mm a")
  if (isYesterday(date)) return `Yesterday · ${format(date, "h:mm a")}`
  return format(date, "MMM d, h:mm a")
}

export function NotificationItem({ notification }: NotificationItemProps) {
  const router = useRouter()
  const { mutate: markAsRead } = useMarkAsRead()

  const Icon = getNotificationIcon(notification.type)
  const color = getNotificationColor(notification.type)

  const handleClick = () => {
    if (!notification.is_read) {
      markAsRead(notification.id)
    }

    if (notification.related_entity_type && notification.related_entity_id) {
      if (notification.related_entity_type === "candidate") {
        router.push(`/candidates/${notification.related_entity_id}`)
      } else if (notification.related_entity_type === "selection") {
        router.push(`/selections/${notification.related_entity_id}`)
      }
    }
  }

  const handleMarkAsRead = (e: React.MouseEvent) => {
    e.stopPropagation()
    markAsRead(notification.id)
  }

  const relatedLink =
    notification.related_entity_type === "candidate" ? (
      <Button
        variant="link"
        size="sm"
        className="h-auto p-0 text-xs"
        onClick={(e) => {
          e.stopPropagation()
          router.push(`/candidates/${notification.related_entity_id}`)
        }}
      >
        <Eye className="mr-1 h-3 w-3" />
        View candidate
      </Button>
    ) : notification.related_entity_type === "selection" ? (
      <Button
        variant="link"
        size="sm"
        className="h-auto p-0 text-xs"
        onClick={(e) => {
          e.stopPropagation()
          router.push(`/selections/${notification.related_entity_id}`)
        }}
      >
        <Eye className="mr-1 h-3 w-3" />
        View selection
      </Button>
    ) : null

  return (
    <button
      type="button"
      onClick={handleClick}
      className={cn(
        "w-full rounded-2xl border border-border bg-card p-4 text-left transition-colors hover:bg-muted/40",
        !notification.is_read && "border-primary/30 bg-primary/5",
      )}
    >
      <div className="flex gap-3">
        <div className={cn("flex h-9 w-9 shrink-0 items-center justify-center rounded-full", color)}>
          <Icon className="h-4 w-4" />
        </div>
        <div className="min-w-0 flex-1 space-y-1.5">
          <div className="flex items-start justify-between gap-2">
            <p className={cn("text-sm text-foreground", !notification.is_read && "font-semibold")}>
              {notification.title}
            </p>
            <time className="shrink-0 text-[10px] text-muted-foreground">
              {formatNotificationTime(notification.created_at)}
            </time>
          </div>
          <p className="text-sm text-muted-foreground">{notification.message}</p>
          <div className="flex items-center justify-between gap-2">
            {relatedLink}
            {!notification.is_read ? (
              <Button variant="ghost" size="sm" onClick={handleMarkAsRead} className="h-7 text-xs">
                <Check className="mr-1 h-3 w-3" />
                Mark read
              </Button>
            ) : null}
          </div>
        </div>
      </div>
    </button>
  )
}
