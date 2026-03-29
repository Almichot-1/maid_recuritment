"use client"

import * as React from "react"
import Link from "next/link"
import { usePathname } from "next/navigation"
import { LayoutDashboard, Users, UserPlus, CheckSquare, Bell, Settings, Search, ChevronLeft, ChevronRight, LogOut, Route, Link2 } from "lucide-react"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { useCurrentUser, useLogout } from "@/hooks/use-auth"
import { usePairingContext } from "@/hooks/use-pairings"
import { useLayoutStore } from "@/stores/layout-store"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { useProfileAvatar } from "@/hooks/use-profile-avatar"
import { getRoleHomeLabel, getRoleHomePath, isRoleHomePath } from "@/lib/role-home"
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
  const { isSidebarCollapsed, toggleSidebar, setSidebarCollapsed } = useLayoutStore()
  const [mounted, setMounted] = React.useState(false)
  const hasWorkspaceAccess = !isReady || hasActivePairs
  const dashboardHref = hasWorkspaceAccess ? getRoleHomePath(user?.role) : "/waiting"
  const dashboardLabel = hasWorkspaceAccess ? getRoleHomeLabel(user?.role) : "Workspace Status"

  React.useEffect(() => {
    setMounted(true)
    const stored = localStorage.getItem("sidebar-collapsed")
    if (stored === "true") setSidebarCollapsed(true)
  }, [setSidebarCollapsed])

  const ethiopianLinks: NavItem[] = [
    { name: dashboardLabel, href: dashboardHref, icon: LayoutDashboard },
    { name: "Partner Workspaces", href: "/partners", icon: Link2 },
    { name: "Candidates", href: "/candidates", icon: Users },
    { name: "Add Candidate", href: "/candidates/new", icon: UserPlus },
    { name: "Selections", href: "/selections", icon: CheckSquare },
    { name: "Process Tracking", href: "/tracking", icon: Route },
    { name: "Notifications", href: "/notifications", icon: Bell },
    { name: "Settings", href: "/settings", icon: Settings },
  ]

  const foreignLinks: NavItem[] = [
    { name: dashboardLabel, href: dashboardHref, icon: LayoutDashboard },
    { name: "Partner Workspaces", href: "/partners", icon: Link2 },
    { name: "Browse Candidates", href: "/candidates", icon: Search },
    { name: "My Selections", href: "/selections", icon: CheckSquare },
    { name: "Process Tracking", href: "/tracking", icon: Route },
    { name: "Notifications", href: "/notifications", icon: Bell },
    { name: "Settings", href: "/settings", icon: Settings },
  ]

  const waitingLinks: NavItem[] = [
    { name: "Workspace Status", href: "/waiting", icon: Route },
  ]

  const links = hasWorkspaceAccess
    ? (isEthiopianAgent ? ethiopianLinks : (isForeignAgent ? foreignLinks : []))
    : waitingLinks
  const activeHref = getActiveNavHref(pathname, links)

  if (!mounted || !user) return <aside className="hidden md:flex w-64 bg-slate-950 flex-col inset-y-0 fixed z-50 border-r border-slate-800" />

  return (
    <aside
      className={cn(
        "hidden md:flex flex-col fixed inset-y-0 left-0 bg-slate-950 text-slate-300 transition-all duration-300 ease-in-out z-50 border-r border-slate-800",
        isSidebarCollapsed ? "w-16" : "w-64"
      )}
    >
      <div className="flex h-14 items-center justify-between px-4 border-b border-slate-800">
        {!isSidebarCollapsed && (
          <Link href={dashboardHref} className="flex items-center space-x-2 font-bold text-white truncate">
            <span>Maid Recruiting</span>
          </Link>
        )}
        <Button 
          variant="ghost" 
          size="icon" 
          onClick={toggleSidebar} 
          className="text-slate-400 hover:text-white hover:bg-slate-800 ml-auto p-1 h-8 w-8 shrink-0 relative right-[-4px]"
        >
          {isSidebarCollapsed ? <ChevronRight className="h-5 w-5" /> : <ChevronLeft className="h-5 w-5" />}
        </Button>
      </div>

      <div className="flex-1 overflow-y-auto py-4 space-y-1 px-2 hide-scrollbar">
        {!isSidebarCollapsed && hasWorkspaceAccess && hasActivePairs ? (
          <div className="mb-4 px-1">
            <PartnerSwitcher className="w-full min-w-0" />
          </div>
        ) : null}
        {links.map((link) => {
          const isActive = activeHref === link.href
          const isNotification = link.name === "Notifications"
          return (
            <Link
              key={link.name}
              href={link.href}
              className={cn(
                "group relative flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-all duration-200",
                isActive
                  ? "bg-gradient-to-r from-teal-500 to-sky-500 text-white shadow-[0_18px_36px_-26px_rgba(56,189,248,0.85)]"
                  : "hover:-translate-y-0.5 hover:bg-slate-800/90 hover:text-white hover:shadow-[0_12px_26px_-20px_rgba(15,23,42,0.9)]"
              )}
            >
              <div className="relative shrink-0">
                <link.icon className="h-5 w-5" />
                {isNotification && isSidebarCollapsed && (
                  <span className="absolute -top-1 -right-1 h-2.5 w-2.5 rounded-full bg-destructive ring-2 ring-slate-950" />
                )}
              </div>
              {!isSidebarCollapsed && <span className="truncate flex-1">{link.name}</span>}
              {isNotification && !isSidebarCollapsed && (
                <span className={cn(
                  "ml-auto text-[10px] font-bold px-1.5 py-0.5 rounded-full",
                  isActive ? "bg-primary-foreground text-primary" : "bg-destructive text-destructive-foreground"
                )}>3</span>
              )}
              {isSidebarCollapsed && (
                <div className="absolute left-14 hidden group-hover:flex bg-slate-800 text-white text-xs font-semibold px-2 py-1 rounded shadow-md z-50 whitespace-nowrap">
                  {link.name}
                </div>
              )}
            </Link>
          )
        })}
      </div>

      <div className="border-t border-slate-800 p-4 overflow-hidden">
        <div className={cn("flex items-center", isSidebarCollapsed ? "justify-center" : "gap-3")}>
          <Avatar className="h-8 w-8 shrink-0 border border-slate-700">
            <AvatarImage src={avatarDataURL || `https://api.dicebear.com/7.x/initials/svg?seed=${user.full_name}`} alt={user.full_name} />
            <AvatarFallback className="bg-slate-800 text-xs text-white">{user.full_name?.charAt(0) || "U"}</AvatarFallback>
          </Avatar>
          {!isSidebarCollapsed && (
            <div className="flex flex-col truncate flex-1 min-w-0">
              <span className="text-sm font-medium text-white truncate">{user.full_name}</span>
              <span className="text-[10px] text-slate-400 capitalize truncate">{user.role?.replace('_', ' ')}</span>
            </div>
          )}
          {!isSidebarCollapsed && (
            <Button variant="ghost" size="icon" onClick={logout} className="text-slate-400 hover:text-destructive hover:bg-slate-800 shrink-0" title="Logout">
              <LogOut className="h-4 w-4" />
            </Button>
          )}
        </div>
        {isSidebarCollapsed && (
          <Button variant="ghost" size="icon" onClick={logout} className="mt-4 w-full h-8 flex items-center justify-center text-slate-400 hover:text-destructive hover:bg-slate-800" title="Logout">
            <LogOut className="h-5 w-5" />
          </Button>
        )}
      </div>
    </aside>
  )
}
