"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { Bell, CheckCheck, Loader2 } from "lucide-react"
import { formatDistanceToNow } from "date-fns"

import { useNotifications, useUnreadCount, useMarkAsRead, useMarkAllAsRead } from "@/hooks/use-notifications"
import { Notification } from "@/types"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Separator } from "@/components/ui/separator"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import { getNotificationIcon, getNotificationColor } from "@/lib/notification-utils"

export function NotificationBell() {
  const router = useRouter()
  const { count } = useUnreadCount()
  const { data: recentNotifications = [], isLoading } = useNotifications()
  const { mutate: markAsRead } = useMarkAsRead()
  const { mutate: markAllAsRead, isPending: isMarkingAll } = useMarkAllAsRead()

  const lastFive = recentNotifications.slice(0, 5)

  const handleNotificationClick = (notification: Notification) => {
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

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" className="relative">
          <Bell className="h-5 w-5" />
          {count > 0 && (
            <Badge
              className="absolute -top-1 -right-1 h-5 w-5 flex items-center justify-center p-0 bg-red-500 hover:bg-red-600 text-white text-xs"
            >
              {count > 9 ? "9+" : count}
            </Badge>
          )}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-[380px] p-0">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b">
          <h3 className="font-semibold text-base">Notifications</h3>
          {count > 0 && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => markAllAsRead()}
              disabled={isMarkingAll}
              className="h-8 text-xs"
            >
              {isMarkingAll ? (
                <Loader2 className="h-3 w-3 animate-spin mr-1" />
              ) : (
                <CheckCheck className="h-3 w-3 mr-1" />
              )}
              Mark all read
            </Button>
          )}
        </div>

        {/* Notifications List */}
        <div className="max-h-[400px] overflow-y-auto">
          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          ) : lastFive.length > 0 ? (
            <div className="divide-y">
              {lastFive.map((notification) => {
                const Icon = getNotificationIcon(notification.type)
                const color = getNotificationColor(notification.type)

                return (
                  <button
                    key={notification.id}
                    onClick={() => handleNotificationClick(notification)}
                    className={cn(
                      "w-full text-left p-4 hover:bg-muted/50 transition-colors relative",
                      !notification.is_read && "bg-blue-50/50 dark:bg-blue-950/10"
                    )}
                  >
                    <div className="flex gap-3">
                      <div className={cn("flex h-8 w-8 shrink-0 items-center justify-center rounded-full", color)}>
                        <Icon className="h-4 w-4" />
                      </div>
                      <div className="flex-1 min-w-0 space-y-1">
                        <p className={cn("text-sm leading-tight", !notification.is_read && "font-semibold")}>
                          {notification.title}
                        </p>
                        <p className="text-xs text-muted-foreground line-clamp-2">
                          {notification.message}
                        </p>
                        <p className="text-xs text-muted-foreground">
                          {formatDistanceToNow(new Date(notification.created_at), { addSuffix: true })}
                        </p>
                      </div>
                      {!notification.is_read && (
                        <div className="h-2 w-2 rounded-full bg-blue-500 shrink-0 mt-1" />
                      )}
                    </div>
                  </button>
                )
              })}
            </div>
          ) : (
            <div className="py-12 text-center text-sm text-muted-foreground">
              <Bell className="h-12 w-12 mx-auto mb-3 opacity-20" />
              <p>No notifications yet</p>
            </div>
          )}
        </div>

        {/* Footer */}
        {lastFive.length > 0 && (
          <>
            <Separator />
            <div className="p-2">
              <Button
                variant="ghost"
                className="w-full justify-center text-sm"
                onClick={() => router.push("/notifications")}
              >
                View All Notifications
              </Button>
            </div>
          </>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
