"use client"

import * as React from "react"
import dynamic from "next/dynamic"
import { usePathname } from "next/navigation"
import { MobileNav } from "@/components/layout/mobile-nav"
import { isRoleHomePath } from "@/lib/role-home"

const NotificationBell = dynamic(
  () => import("@/components/notifications/notification-bell").then((module) => module.NotificationBell)
)

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
    <header className="sticky top-0 z-40 border-b border-border/70 bg-background/90 px-4 py-3 backdrop-blur md:hidden">
      <div className="flex items-center gap-3">
        <MobileNav />

        <div className="min-w-0 flex-1">
          <p className="truncate text-[11px] font-semibold uppercase tracking-[0.24em] text-primary/80">
            Workspace
          </p>
          <h1 className="truncate text-base font-semibold">{title}</h1>
        </div>

        <NotificationBell />
      </div>
    </header>
  )
}
