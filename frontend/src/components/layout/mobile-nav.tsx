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
import { LocaleSwitcher } from "@/components/shared/locale-switcher"
import { Sheet, SheetContent, SheetTrigger, SheetHeader, SheetTitle } from "@/components/ui/sheet"
import { useCurrentUser, useLogout } from "@/hooks/use-auth"
import { usePairingContext } from "@/hooks/use-pairings"
import { useProfileAvatar } from "@/hooks/use-profile-avatar"
import { useI18n } from "@/lib/i18n"
import { NavItem } from "./sidebar"
import { cn } from "@/lib/utils"
import { getRoleHomePath, isRoleHomePath } from "@/lib/role-home"
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
  const { hasActivePairs, isReady } = usePairingContext()
  const logout = useLogout()
  const { isRTL, t } = useI18n()
  const hasWorkspaceAccess = !isReady || hasActivePairs
  const dashboardHref = hasWorkspaceAccess ? getRoleHomePath(user?.role) : "/waiting"
  const dashboardLabel = hasWorkspaceAccess
    ? (user?.role === "ethiopian_agent" ? t("nav.agencyHome") : t("nav.employerHome"))
    : t("nav.waiting")

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

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button variant="ghost" size="icon" className="shrink-0 md:hidden">
          <Menu className="h-5 w-5" />
          <span className="sr-only">Toggle navigation menu</span>
        </Button>
      </SheetTrigger>
      <SheetContent
        side={isRTL ? "right" : "left"}
        className="flex h-full w-full max-w-none flex-col border-0 bg-background p-0"
      >
        <SheetHeader className="border-b border-border p-4 text-left">
          <SheetTitle className="font-display text-3xl text-foreground">RecruitMatch</SheetTitle>
          <p className="route-stamp text-[10px] text-muted-foreground">{t("sidebar.serial")}</p>
        </SheetHeader>
        {hasWorkspaceAccess && hasActivePairs ? (
          <div className="border-b border-border p-4">
            <PartnerSwitcher compact className="w-full" />
          </div>
        ) : null}
        <div className="border-b border-border p-4">
          <LocaleSwitcher />
        </div>
        <div className="hide-scrollbar flex-1 space-y-1 overflow-y-auto px-2 py-4">
          {links.map((link) => {
            const isActive = activeHref === link.href
            return (
              <Link
                key={link.name}
                href={link.href}
                onClick={() => setOpen(false)}
                className={cn(
                  "flex items-center gap-3 border border-transparent px-3 py-3 text-sm font-bold uppercase tracking-[0.05em] transition-colors",
                  isActive
                    ? "border-primary bg-primary text-primary-foreground"
                    : "text-muted-foreground hover:border-border hover:bg-muted/30 hover:text-foreground"
                )}
              >
                <link.icon className="h-5 w-5 shrink-0" />
                <span className="truncate flex-1">{link.name}</span>
              </Link>
            )
          })}
        </div>
        <div className="border-t border-border p-4">
          {user ? (
            <div className="mb-4 flex items-center gap-3 border border-border bg-muted/20 p-3">
              <div className="h-10 w-10 overflow-hidden rounded-full border border-border">
                <img
                  src={avatarDataURL || `https://api.dicebear.com/7.x/initials/svg?seed=${user.full_name}`}
                  alt={user.full_name}
                  className="h-full w-full object-cover"
                />
              </div>
              <div className="min-w-0">
                <p className="truncate text-sm font-bold text-foreground">{user.full_name}</p>
                <p className="truncate text-xs text-muted-foreground">{user.company_name || user.email}</p>
              </div>
            </div>
          ) : null}
          <Button 
            variant="ghost" 
            className="w-full justify-start"
            onClick={() => {
              setOpen(false)
              logout()
            }}
          >
            <LogOut className="mr-2 h-5 w-5" />
            {t("common.logout")}
          </Button>
        </div>
      </SheetContent>
    </Sheet>
  )
}
