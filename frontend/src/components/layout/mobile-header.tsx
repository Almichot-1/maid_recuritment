"use client"

import * as React from "react"
import { usePathname } from "next/navigation"
import { NotificationBell } from "@/components/notifications/notification-bell"
import { MobileNav } from "@/components/layout/mobile-nav"
import { isRoleHomePath } from "@/lib/role-home"

export function MobileHeader() {
  const pathname = usePathname()

  const getPageTitle = (path: string) => {
    if (path.startsWith("/dashboard/agency")) return "Agency Home"
    if (path.startsWith("/dashboard/employer")) return "Employer Home"
    if (isRoleHomePath(path)) return "Dashboard"
    if (path.startsWith("/partners")) return "Partner Workspaces"
    if (path.startsWith("/waiting")) return "Partner Assignment"
    if (path.startsWith("/tracking")) return "Process Tracking"
    if (path.startsWith("/candidates")) return "Candidates"
    if (path.startsWith("/selections")) return "Selections"
    if (path.startsWith("/notifications")) return "Notifications"
    if (path.startsWith("/settings")) return "Settings"
    return "Dashboard"
  }

  const title = getPageTitle(pathname)

  return (
    <header className="sticky top-0 z-40 border-b bg-background px-4 py-3 md:hidden">
      <div className="flex items-center gap-4">
        <MobileNav />

        <h1 className="flex-1 truncate text-lg font-semibold">{title}</h1>

        <NotificationBell />
      </div>
    </header>
  )
}
