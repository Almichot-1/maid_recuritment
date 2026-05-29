"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { Bell, CheckCheck, Loader2 } from "lucide-react"
import { format, formatDistanceToNow, isToday, isYesterday } from "date-fns"

import { useNotifications, useUnreadCount, useMarkAsRead, useMarkAllAsRead } from "@/hooks/use-notifications"
import { usePairingContext } from "@/hooks/use-pairings"
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

function formatNotificationTime(createdAt: string) {
  const date = new Date(createdAt)
  if (isToday(date)) {
    return format(date, "h:mm a")
  }
  if (isYesterday(date)) {
    return `Yesterday ${format(date, "h:mm a")}`
  }
  return formatDistanceToNow(date, { addSuffix: true })
}

function workspaceLabel(companyName?: string, fullName?: string) {
  const name = (companyName || fullName || "").trim()
  return name || "Workspace"
}

export function NotificationBell() {
  const router = useRouter()
  const { activeWorkspace, hasActivePairs } = usePairingContext()
  const [open, setOpen] = React.useState(false)
  const workspaceName = activeWorkspace
    ? workspaceLabel(activeWorkspace.partner_agency?.company_name, activeWorkspace.partner_agency?.full_name)
    : null
  const { count } = useUnreadCount()
  const { data: notificationData, isLoading } = useNotifications(false, {
    enabled: open,
    pageSize: 8,
    refetchInterval: open ? 60000 : false,
  })
  const { mutate: markAsRead } = useMarkAsRead()
  const { mutate: markAllAsRead, isPending: isMarkingAll } = useMarkAllAsRead()

  const items = notificationData?.notifications || []

  const handleNotificationClick = (notification: Notification) => {
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
    setOpen(false)
  }

  return (
    <DropdownMenu open={open} onOpenChange={setOpen}>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" className="relative">
          <Bell className="h-5 w-5" />
          {count > 0 ? (
            <Badge className="absolute -right-1 -top-1 flex h-5 w-5 items-center justify-center bg-destructive p-0 text-[10px] text-destructive-foreground hover:bg-destructive">
              {count > 9 ? "9+" : count}
            </Badge>
          ) : null}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-[min(100vw-2rem,400px)] p-0">
        <div className="flex items-center justify-between border-b border-border px-4 py-3">
          <div className="min-w-0">
            <h3 className="font-semibold text-foreground">Activity</h3>
            {hasActivePairs && workspaceName ? (
              <p className="truncate text-xs text-muted-foreground">{workspaceName}</p>
            ) : null}
          </div>
          {count > 0 ? (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => markAllAsRead()}
              disabled={isMarkingAll}
              className="h-8 text-xs"
            >
              {isMarkingAll ? (
                <Loader2 className="mr-1 h-3 w-3 animate-spin" />
              ) : (
                <CheckCheck className="mr-1 h-3 w-3" />
              )}
              Mark all read
            </Button>
          ) : null}
        </div>

        <div className="max-h-[min(70vh,420px)] overflow-y-auto bg-muted/20 px-3 py-3">
          {isLoading ? (
            <div className="flex items-center justify-center py-10">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          ) : items.length > 0 ? (
            <div className="space-y-2">
              {items.map((notification) => {
                const Icon = getNotificationIcon(notification.type)
                const color = getNotificationColor(notification.type)
                return (
                  <button
                    key={notification.id}
                    type="button"
                    onClick={() => handleNotificationClick(notification)}
                    className={cn(
                      "flex w-full gap-2 rounded-2xl px-1 py-1 text-left transition-colors hover:bg-muted/60",
                      !notification.is_read && "bg-card shadow-sm",
                    )}
                  >
                    <div className="mt-1 flex max-w-[85%] flex-col gap-1 rounded-2xl border border-border bg-card px-3 py-2.5">
                      <div className="flex items-start gap-2">
                        <div className={cn("flex h-7 w-7 shrink-0 items-center justify-center rounded-full", color)}>
                          <Icon className="h-3.5 w-3.5" />
                        </div>
                        <div className="min-w-0 flex-1">
                          <p
                            className={cn(
                              "text-sm leading-snug text-foreground",
                              !notification.is_read && "font-semibold",
                            )}
                          >
                            {notification.title}
                          </p>
                          <p className="mt-0.5 line-clamp-2 text-xs text-muted-foreground">
                            {notification.message}
                          </p>
                        </div>
                      </div>
                      <p className="text-[10px] text-muted-foreground">
                        {formatNotificationTime(notification.created_at)}
                      </p>
                    </div>
                    {!notification.is_read ? (
                      <span className="mt-3 h-2 w-2 shrink-0 rounded-full bg-primary" aria-hidden />
                    ) : null}
                  </button>
                )
              })}
            </div>
          ) : (
            <div className="py-12 text-center text-sm text-muted-foreground">
              <Bell className="mx-auto mb-3 h-10 w-10 opacity-30" />
              <p>No activity yet</p>
            </div>
          )}
        </div>

        {items.length > 0 ? (
          <>
            <Separator />
            <div className="p-2">
              <Button
                variant="ghost"
                className="w-full justify-center text-sm"
                onClick={() => {
                  setOpen(false)
                  router.push("/notifications")
                }}
              >
                View all
              </Button>
            </div>
          </>
        ) : null}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
