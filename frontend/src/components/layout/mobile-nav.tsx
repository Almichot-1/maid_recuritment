"use client"

import * as React from "react"
import Link from "next/link"
import { usePathname } from "next/navigation"
import { 
  LayoutDashboard, 
  Users, 
  UserPlus, 
  CheckSquare, 
  Bell, 
  Settings, 
  Search, 
  Route,
  Link2,
  LogOut,
  Menu
} from "lucide-react"

import { Button } from "@/components/ui/button"
import { Sheet, SheetContent, SheetTrigger, SheetHeader, SheetTitle } from "@/components/ui/sheet"
import { useCurrentUser, useLogout } from "@/hooks/use-auth"
import { useUnreadCount } from "@/hooks/use-notifications"
import { usePairingContext } from "@/hooks/use-pairings"
import { useProfileAvatar } from "@/hooks/use-profile-avatar"
import { NavItem } from "./sidebar"
import { cn } from "@/lib/utils"
import { getRoleHomeLabel, getRoleHomePath, isRoleHomePath } from "@/lib/role-home"
import { PartnerSwitcher } from "@/components/pairings/partner-switcher"

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

export function MobileNav() {
  const [open, setOpen] = React.useState(false)
  const pathname = usePathname()
  const { user, isEthiopianAgent, isForeignAgent } = useCurrentUser()
  const { avatarDataURL } = useProfileAvatar()
  const { count: unreadCount } = useUnreadCount()
  const { hasActivePairs, isReady } = usePairingContext()
  const logout = useLogout()
  const hasWorkspaceAccess = !isReady || hasActivePairs
  const dashboardHref = hasWorkspaceAccess ? getRoleHomePath(user?.role) : "/waiting"
  const dashboardLabel = hasWorkspaceAccess ? getRoleHomeLabel(user?.role) : "Workspace Status"

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

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button variant="ghost" size="icon" className="md:hidden shrink-0 text-muted-foreground hover:text-foreground">
          <Menu className="h-5 w-5" />
          <span className="sr-only">Toggle navigation menu</span>
        </Button>
      </SheetTrigger>
      <SheetContent side="left" className="bg-slate-950 text-slate-300 border-r-slate-800 p-0 flex flex-col w-72">
        <SheetHeader className="p-4 border-b border-slate-800 text-left">
          <SheetTitle className="text-white font-bold">Maid Recruiting</SheetTitle>
        </SheetHeader>
        {hasWorkspaceAccess && hasActivePairs ? (
          <div className="border-b border-slate-800 p-4">
            <PartnerSwitcher compact className="w-full border-slate-700/80 bg-slate-900/70 text-slate-100" />
          </div>
        ) : null}
        <div className="flex-1 overflow-y-auto py-4 space-y-1 px-2 hide-scrollbar">
          {links.map((link) => {
            const isActive = activeHref === link.href
            const isNotification = link.name === "Notifications"
            return (
              <Link
                key={link.name}
                href={link.href}
                onClick={() => setOpen(false)}
                className={cn(
                  "flex items-center gap-3 px-3 py-2.5 rounded-md transition-colors font-medium text-sm",
                  isActive
                    ? "bg-primary text-primary-foreground"
                    : "hover:bg-slate-800 hover:text-white"
                )}
              >
                <link.icon className="h-5 w-5 shrink-0" />
                <span className="truncate flex-1">{link.name}</span>
                {isNotification && unreadCount > 0 && (
                  <span className={cn(
                    "ml-auto text-[10px] font-bold px-1.5 py-0.5 rounded-full",
                    isActive ? "bg-primary-foreground text-primary" : "bg-destructive text-destructive-foreground"
                  )}>
                    {unreadCount}
                  </span>
                )}
              </Link>
            )
          })}
        </div>
        <div className="border-t border-slate-800 p-4">
          {user ? (
            <div className="mb-4 flex items-center gap-3 rounded-2xl border border-slate-800 bg-slate-900/70 p-3">
              <div className="h-10 w-10 overflow-hidden rounded-full border border-slate-700">
                <img
                  src={avatarDataURL || `https://api.dicebear.com/7.x/initials/svg?seed=${user.full_name}`}
                  alt={user.full_name}
                  className="h-full w-full object-cover"
                />
              </div>
              <div className="min-w-0">
                <p className="truncate text-sm font-semibold text-white">{user.full_name}</p>
                <p className="truncate text-xs text-slate-400">{user.company_name || user.email}</p>
              </div>
            </div>
          ) : null}
          <Button 
            variant="ghost" 
            className="w-full justify-start text-slate-400 hover:text-destructive hover:bg-slate-800"
            onClick={() => {
              setOpen(false)
              logout()
            }}
          >
            <LogOut className="mr-2 h-5 w-5" />
            Logout
          </Button>
        </div>
      </SheetContent>
    </Sheet>
  )
}
