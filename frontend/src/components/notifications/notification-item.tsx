"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { formatDistanceToNow } from "date-fns"
import { Check, Eye } from "lucide-react"

import { Notification } from "@/types"
import { useMarkAsRead } from "@/hooks/use-notifications"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { cn } from "@/lib/utils"
import { getNotificationIcon, getNotificationColor } from "@/lib/notification-utils"

interface NotificationItemProps {
  notification: Notification
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

    // Navigate to related entity
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

  const getRelatedEntityLink = () => {
    if (!notification.related_entity_type || !notification.related_entity_id) {
      return null
    }

    if (notification.related_entity_type === "candidate") {
      return (
        <Button
          variant="link"
          size="sm"
          className="h-auto p-0 text-xs"
          onClick={(e) => {
            e.stopPropagation()
            router.push(`/candidates/${notification.related_entity_id}`)
          }}
        >
          <Eye className="h-3 w-3 mr-1" />
          View Candidate
        </Button>
      )
    }

    if (notification.related_entity_type === "selection") {
      return (
        <Button
          variant="link"
          size="sm"
          className="h-auto p-0 text-xs"
          onClick={(e) => {
            e.stopPropagation()
            router.push(`/selections/${notification.related_entity_id}`)
          }}
        >
          <Eye className="h-3 w-3 mr-1" />
          View Selection
        </Button>
      )
    }

    return null
  }

  return (
    <Card
      className={cn(
        "cursor-pointer transition-all hover:shadow-md",
        !notification.is_read && "bg-blue-50/50 dark:bg-blue-950/10 border-blue-200 dark:border-blue-800"
      )}
      onClick={handleClick}
    >
      <CardContent className="p-4">
        <div className="flex gap-4">
          {/* Left: Icon */}
          <div className={cn("flex h-10 w-10 shrink-0 items-center justify-center rounded-full", color)}>
            <Icon className="h-5 w-5" />
          </div>

          {/* Center: Content */}
          <div className="flex-1 min-w-0 space-y-2">
            <div className="flex items-start justify-between gap-2">
              <h4 className={cn("text-sm leading-tight", !notification.is_read && "font-bold")}>
                {notification.title}
              </h4>
              {!notification.is_read && (
                <div className="h-2 w-2 rounded-full bg-blue-500 shrink-0 mt-1" />
              )}
            </div>
            <p className="text-sm text-muted-foreground">{notification.message}</p>
            <div className="flex items-center gap-3">
              {getRelatedEntityLink()}
              <span className="text-xs text-muted-foreground">
                {formatDistanceToNow(new Date(notification.created_at), { addSuffix: true })}
              </span>
            </div>
          </div>

          {/* Right: Actions */}
          <div className="flex flex-col items-end gap-2">
            {!notification.is_read && (
              <Button
                variant="ghost"
                size="sm"
                onClick={handleMarkAsRead}
                className="h-8 text-xs"
              >
                <Check className="h-3 w-3 mr-1" />
                Mark read
              </Button>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
