"use client"

import * as React from "react"
import { Bell, CheckCheck, ChevronRight, Home, Inbox, Loader2 } from "lucide-react"
import Link from "next/link"

import { useNotifications, useMarkAllAsRead } from "@/hooks/use-notifications"
import { NotificationItem } from "@/components/notifications/notification-item"
import { PageHeader } from "@/components/layout/page-header"
import { Button } from "@/components/ui/button"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Badge } from "@/components/ui/badge"

export default function NotificationsPage() {
  const { data: allNotifications = [], isLoading } = useNotifications()
  const { mutate: markAllAsRead, isPending: isMarkingAll } = useMarkAllAsRead()

  // Filter notifications
  const unreadNotifications = React.useMemo(
    () => allNotifications.filter((n) => !n.is_read),
    [allNotifications]
  )

  const selectionNotifications = React.useMemo(
    () => allNotifications.filter((n) => n.type === "selection"),
    [allNotifications]
  )

  const approvalNotifications = React.useMemo(
    () => allNotifications.filter((n) => n.type === "approval" || n.type === "rejection"),
    [allNotifications]
  )

  const statusNotifications = React.useMemo(
    () => allNotifications.filter((n) => n.type === "status_update"),
    [allNotifications]
  )

  const breadcrumbs = (
    <nav className="flex items-center text-sm font-medium text-muted-foreground mb-6">
      <Link href="/dashboard" className="transition-all hover:text-primary flex items-center">
        <Home className="mr-1.5 h-4 w-4" />
        Dashboard
      </Link>
      <ChevronRight className="h-4 w-4 mx-1 opacity-50" />
      <span className="text-foreground font-semibold">Notifications</span>
    </nav>
  )

  const emptyState = (
    <div className="flex flex-col items-center justify-center py-16 space-y-4">
      <div className="flex h-20 w-20 items-center justify-center rounded-full bg-muted">
        <Inbox className="h-10 w-10 text-muted-foreground" />
      </div>
      <div className="text-center space-y-2">
        <h3 className="text-lg font-semibold">No notifications yet</h3>
        <p className="text-muted-foreground max-w-md">
          You will be notified about selections, approvals, and status updates.
        </p>
      </div>
    </div>
  )

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500 pb-10">
      {breadcrumbs}

      <div className="flex items-center justify-between">
        <PageHeader
          heading="Notifications"
          text="Stay updated on selections, approvals, and recruitment progress."
        />
        {unreadNotifications.length > 0 && (
          <Button
            onClick={() => markAllAsRead()}
            disabled={isMarkingAll}
            variant="outline"
          >
            {isMarkingAll ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <CheckCheck className="mr-2 h-4 w-4" />
            )}
            Mark All as Read
          </Button>
        )}
      </div>

      <Tabs defaultValue="all" className="space-y-6">
        <TabsList className="grid w-full grid-cols-5 h-auto p-1">
          <TabsTrigger value="all" className="relative">
            All
            {allNotifications.length > 0 && (
              <Badge variant="secondary" className="ml-2">
                {allNotifications.length}
              </Badge>
            )}
          </TabsTrigger>
          <TabsTrigger value="unread" className="relative">
            Unread
            {unreadNotifications.length > 0 && (
              <Badge
                variant="secondary"
                className="ml-2 bg-blue-500 text-white hover:bg-blue-600"
              >
                {unreadNotifications.length}
              </Badge>
            )}
          </TabsTrigger>
          <TabsTrigger value="selection" className="relative">
            Selection
            {selectionNotifications.length > 0 && (
              <Badge variant="secondary" className="ml-2">
                {selectionNotifications.length}
              </Badge>
            )}
          </TabsTrigger>
          <TabsTrigger value="approval" className="relative">
            Approval
            {approvalNotifications.length > 0 && (
              <Badge variant="secondary" className="ml-2">
                {approvalNotifications.length}
              </Badge>
            )}
          </TabsTrigger>
          <TabsTrigger value="status" className="relative">
            Status Updates
            {statusNotifications.length > 0 && (
              <Badge variant="secondary" className="ml-2">
                {statusNotifications.length}
              </Badge>
            )}
          </TabsTrigger>
        </TabsList>

        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : (
          <>
            <TabsContent value="all" className="space-y-3">
              {allNotifications.length > 0 ? (
                allNotifications.map((notification) => (
                  <NotificationItem key={notification.id} notification={notification} />
                ))
              ) : (
                emptyState
              )}
            </TabsContent>

            <TabsContent value="unread" className="space-y-3">
              {unreadNotifications.length > 0 ? (
                unreadNotifications.map((notification) => (
                  <NotificationItem key={notification.id} notification={notification} />
                ))
              ) : (
                <div className="text-center py-12 text-muted-foreground">
                  <Bell className="h-12 w-12 mx-auto mb-3 opacity-20" />
                  <p>No unread notifications</p>
                </div>
              )}
            </TabsContent>

            <TabsContent value="selection" className="space-y-3">
              {selectionNotifications.length > 0 ? (
                selectionNotifications.map((notification) => (
                  <NotificationItem key={notification.id} notification={notification} />
                ))
              ) : (
                <div className="text-center py-12 text-muted-foreground">
                  <p>No selection notifications</p>
                </div>
              )}
            </TabsContent>

            <TabsContent value="approval" className="space-y-3">
              {approvalNotifications.length > 0 ? (
                approvalNotifications.map((notification) => (
                  <NotificationItem key={notification.id} notification={notification} />
                ))
              ) : (
                <div className="text-center py-12 text-muted-foreground">
                  <p>No approval notifications</p>
                </div>
              )}
            </TabsContent>

            <TabsContent value="status" className="space-y-3">
              {statusNotifications.length > 0 ? (
                statusNotifications.map((notification) => (
                  <NotificationItem key={notification.id} notification={notification} />
                ))
              ) : (
                <div className="text-center py-12 text-muted-foreground">
                  <p>No status update notifications</p>
                </div>
              )}
            </TabsContent>
          </>
        )}
      </Tabs>
    </div>
  )
}
