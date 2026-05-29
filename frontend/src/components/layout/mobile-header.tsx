"use client"

import * as React from "react"
import dynamic from "next/dynamic"
import { usePathname } from "next/navigation"
import { MobileNav } from "@/components/layout/mobile-nav"
import { useI18n } from "@/lib/i18n"
import { isRoleHomePath } from "@/lib/role-home"

const NotificationBell = dynamic(
  () => import("@/components/notifications/notification-bell").then((module) => module.NotificationBell)
)

const PartnerSwitcher = dynamic(
  () => import("@/components/pairings/partner-switcher").then((module) => module.PartnerSwitcher),
  { ssr: false }
)

export function MobileHeader() {
  const pathname = usePathname()
  const { t } = useI18n()

  const getPageTitle = (path: string) => {
    if (path.startsWith("/dashboard/agency")) return t("nav.agencyHome")
    if (path.startsWith("/dashboard/employer")) return t("nav.employerHome")
    if (isRoleHomePath(path)) return t("nav.dashboard")
    if (path.startsWith("/partners")) return t("nav.partnerWorkspaces")
    if (path.startsWith("/waiting")) return t("nav.waiting")
    if (path.startsWith("/tracking")) return t("nav.processTracking")
    if (path.startsWith("/candidates")) return t("nav.candidates")
    if (path.startsWith("/selections")) return t("nav.selections")
    if (path.startsWith("/notifications")) return t("common.notifications")
    if (path.startsWith("/settings")) return t("common.settings")
    return t("nav.dashboard")
  }

  const title = getPageTitle(pathname)

  return (
    <header className="sticky top-0 z-40 border-b border-border bg-background px-4 py-3 md:hidden">
      <div className="flex items-center gap-3">
        <MobileNav />

        <div className="min-w-0 flex-1">
          <p className="route-stamp truncate text-[10px] text-muted-foreground">{t("header.workspaceStamp")}</p>
          <h1 className="truncate font-display text-2xl text-foreground">{title}</h1>
        </div>

        <NotificationBell />
      </div>

      <div className="mt-2 md:hidden">
        <PartnerSwitcher className="w-full" />
      </div>
    </header>
  )
}
