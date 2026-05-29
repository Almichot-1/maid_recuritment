"use client"

import * as React from "react"
import Link from "next/link"
import { usePathname } from "next/navigation"
import { LayoutDashboard, Users, UserPlus, CheckSquare, Bell, Settings, Search, ChevronLeft, ChevronRight, LogOut, Route, Link2 } from "lucide-react"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Logo } from "@/components/shared/logo"
import { LocaleSwitcher } from "@/components/shared/locale-switcher"
import { useCurrentUser, useLogout } from "@/hooks/use-auth"
import { usePairingContext } from "@/hooks/use-pairings"
import { useLayoutStore } from "@/stores/layout-store"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { useProfileAvatar } from "@/hooks/use-profile-avatar"
import { useI18n } from "@/lib/i18n"
import { getRoleHomePath, isRoleHomePath } from "@/lib/role-home"
import { PartnerSwitcher } from "@/components/pairings/partner-switcher"

export type NavItem = {
  name: string
  href: string
  icon: React.ElementType
}

function matchesNavPath(pathname: string, href: string) {
  if (isRoleHomePath(href)) {
    return isRoleHomePath(pathname)
  }

  if (pathname === href) {
    return true
  }

  return pathname.startsWith(`${href}/`)
}

function getActiveNavHref(pathname: string, links: NavItem[]) {
  const matchingLinks = links
    .filter((link) => matchesNavPath(pathname, link.href))
    .sort((left, right) => right.href.length - left.href.length)

  return matchingLinks[0]?.href || null
}

export function Sidebar() {
  const pathname = usePathname()
  const { user, isEthiopianAgent, isForeignAgent } = useCurrentUser()
  const { avatarDataURL } = useProfileAvatar()
  const { hasActivePairs, isReady } = usePairingContext()
  const logout = useLogout()
  const { isRTL, t } = useI18n()
  const { isSidebarCollapsed, toggleSidebar, setSidebarCollapsed } = useLayoutStore()
  const [mounted, setMounted] = React.useState(false)
  const hasWorkspaceAccess = !isReady || hasActivePairs
  const dashboardHref = hasWorkspaceAccess ? getRoleHomePath(user?.role) : "/waiting"
  const dashboardLabel = hasWorkspaceAccess
    ? (user?.role === "ethiopian_agent" ? t("nav.agencyHome") : t("nav.employerHome"))
    : t("nav.waiting")

  React.useEffect(() => {
    setMounted(true)
    const stored = localStorage.getItem("sidebar-collapsed")
    if (stored === "true") setSidebarCollapsed(true)
  }, [setSidebarCollapsed])

  const ethiopianLinks: NavItem[] = [
    { name: dashboardLabel, href: dashboardHref, icon: LayoutDashboard },
    { name: t("nav.partnerWorkspaces"), href: "/partners", icon: Link2 },
    { name: t("nav.candidates"), href: "/candidates", icon: Users },
    { name: t("nav.addCandidate"), href: "/candidates/new", icon: UserPlus },
    { name: t("nav.selections"), href: "/selections", icon: CheckSquare },
    { name: t("nav.processTracking"), href: "/tracking", icon: Route },
    { name: t("common.notifications"), href: "/notifications", icon: Bell },
    { name: t("common.settings"), href: "/settings", icon: Settings },
  ]

  const foreignLinks: NavItem[] = [
    { name: dashboardLabel, href: dashboardHref, icon: LayoutDashboard },
    { name: t("nav.partnerWorkspaces"), href: "/partners", icon: Link2 },
    { name: t("nav.browseCandidates"), href: "/candidates", icon: Search },
    { name: t("nav.mySelections"), href: "/selections", icon: CheckSquare },
    { name: t("nav.processTracking"), href: "/tracking", icon: Route },
    { name: t("common.notifications"), href: "/notifications", icon: Bell },
    { name: t("common.settings"), href: "/settings", icon: Settings },
  ]

  const waitingLinks: NavItem[] = [
    { name: t("nav.waiting"), href: "/waiting", icon: Route },
  ]

  const links = hasWorkspaceAccess
    ? (isEthiopianAgent ? ethiopianLinks : (isForeignAgent ? foreignLinks : []))
    : waitingLinks
  const activeHref = getActiveNavHref(pathname, links)

  if (!mounted || !user) {
    return <aside className="fixed inset-y-0 hidden w-64 border-r bg-background md:flex" />
  }

  return (
    <aside
      className={cn(
        "fixed inset-y-0 z-50 hidden flex-col border-r border-border bg-card text-foreground transition-all duration-300 ease-in-out md:flex",
        isRTL ? "right-0 border-r-0 border-l" : "left-0",
        isSidebarCollapsed ? "w-16" : "w-64"
      )}
    >
      <div className="border-b border-border px-4 py-4">
        <div className="flex items-start justify-between gap-3">
          {!isSidebarCollapsed ? (
            <div className="space-y-3">
              <Logo href={dashboardHref} showText size="sm" />
              <div className="space-y-1">
                <p className="route-stamp text-[10px]">{t("sidebar.section")}</p>
                <p className="text-xs text-muted-foreground">{t("sidebar.shortNote")}</p>
              </div>
            </div>
          ) : (
            <Logo href={dashboardHref} showText={false} size="sm" />
          )}
          <Button
            variant="ghost"
            size="icon"
            onClick={toggleSidebar}
            className="h-9 w-9 shrink-0"
            aria-label={isSidebarCollapsed ? "Expand sidebar" : "Collapse sidebar"}
          >
            {isSidebarCollapsed ? <ChevronRight className="h-4 w-4" /> : <ChevronLeft className="h-4 w-4" />}
          </Button>
        </div>
        <div className={cn("mt-4", isSidebarCollapsed && "hidden")}>
          <LocaleSwitcher compact />
        </div>
      </div>

      <div className="hide-scrollbar flex-1 space-y-1 overflow-y-auto px-2 py-4">
        {!isSidebarCollapsed && hasWorkspaceAccess && hasActivePairs ? (
          <div className="mb-4 px-1">
            <PartnerSwitcher className="w-full min-w-0" />
          </div>
        ) : null}
        {links.map((link) => {
          const isActive = activeHref === link.href
          return (
            <Link
              key={link.name}
              href={link.href}
              className={cn(
                "group relative flex items-center gap-3 border border-transparent px-3 py-3 text-sm font-bold uppercase tracking-[0.05em] transition-colors duration-200",
                isActive
                  ? "border-primary bg-primary text-primary-foreground"
                  : "text-muted-foreground hover:border-border hover:bg-muted/30 hover:text-foreground"
              )}
            >
              <div className="relative shrink-0">
                <link.icon className="h-5 w-5" />
              </div>
              {!isSidebarCollapsed && <span className="truncate flex-1">{link.name}</span>}
              {isSidebarCollapsed && (
                <div className={cn("absolute top-1/2 hidden -translate-y-1/2 border bg-background px-2 py-1 text-xs text-foreground group-hover:flex", isRTL ? "right-14" : "left-14")}>
                  {link.name}
                </div>
              )}
            </Link>
          )
        })}
      </div>

      <div className="overflow-hidden border-t border-border p-4">
        <div className={cn("flex items-center", isSidebarCollapsed ? "justify-center" : "gap-3")}>
          <Avatar className="h-8 w-8 shrink-0 border border-border">
            <AvatarImage src={avatarDataURL || `https://api.dicebear.com/7.x/initials/svg?seed=${user.full_name}`} alt={user.full_name} />
            <AvatarFallback className="bg-foreground text-xs text-background">{user.full_name?.charAt(0) || "U"}</AvatarFallback>
          </Avatar>
          {!isSidebarCollapsed && (
            <div className="flex flex-col truncate flex-1 min-w-0">
              <span className="truncate text-sm font-bold text-foreground">{user.full_name}</span>
              <span className="truncate text-[10px] uppercase tracking-[0.18em] text-muted-foreground">{user.role?.replace('_', ' ')}</span>
            </div>
          )}
          {!isSidebarCollapsed && (
            <Button variant="ghost" size="icon" onClick={logout} className="shrink-0" aria-label={t("common.logout")}>
              <LogOut className="h-4 w-4" />
            </Button>
          )}
        </div>
        {isSidebarCollapsed && (
          <Button variant="ghost" size="icon" onClick={logout} className="mt-4 flex h-8 w-full items-center justify-center" aria-label={t("common.logout")}>
            <LogOut className="h-5 w-5" />
          </Button>
        )}
      </div>
    </aside>
  )
}
