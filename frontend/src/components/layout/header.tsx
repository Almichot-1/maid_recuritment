"use client"

import * as React from "react"
import Link from "next/link"
import dynamic from "next/dynamic"
import { usePathname } from "next/navigation"
import { LogOut, User as UserIcon, Settings } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { LocaleSwitcher } from "@/components/shared/locale-switcher"
import { PartnerSwitcher } from "@/components/pairings/partner-switcher"
import { useCurrentUser, useLogout } from "@/hooks/use-auth"
import { useProfileAvatar } from "@/hooks/use-profile-avatar"
import { useI18n } from "@/lib/i18n"
import { isRoleHomePath } from "@/lib/role-home"

const NotificationBell = dynamic(
  () => import("@/components/notifications/notification-bell").then((module) => module.NotificationBell)
)

export function Header() {
  const pathname = usePathname()
  const { user } = useCurrentUser()
  const { avatarDataURL } = useProfileAvatar()
  const logout = useLogout()
  const { t } = useI18n()
  
  // Format pathname roughly to a readable string
  const formatPathname = (path: string) => {
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
    const parts = path.split('/').filter(Boolean)
    const page = parts[parts.length - 1]
    return page.charAt(0).toUpperCase() + page.slice(1).replace('-', ' ')
  }

  const title = formatPathname(pathname)

  if (!user) return null

  return (
    <header className="sticky top-0 z-40 flex min-w-0 items-center gap-4 border-b bg-background px-4 py-3 lg:px-6">
      <div className="flex flex-1 items-center">
        <div className="min-w-0">
          <p className="route-stamp text-[10px] text-muted-foreground">{t("header.workspaceStamp")}</p>
          <h1 className="truncate font-display text-2xl text-foreground">{title}</h1>
        </div>
      </div>

      <div className="ml-auto flex items-center gap-2 sm:gap-3">
        <div className="hidden md:block">
          <PartnerSwitcher compact />
        </div>
        <div className="hidden lg:block">
          <LocaleSwitcher compact />
        </div>
        <NotificationBell />

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" className="relative h-10 w-10">
              <Avatar className="h-8 w-8">
                <AvatarImage src={avatarDataURL || `https://api.dicebear.com/7.x/initials/svg?seed=${user.full_name}`} alt={user.full_name} />
                <AvatarFallback>{user.full_name?.charAt(0) || "U"}</AvatarFallback>
              </Avatar>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent className="w-56" align="end" forceMount>
            <DropdownMenuLabel className="font-normal">
              <div className="flex flex-col space-y-1">
                <p className="text-sm font-medium leading-none">{user.full_name}</p>
                <p className="text-xs leading-none text-muted-foreground">
                  {user.email}
                </p>
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              <DropdownMenuItem asChild className="cursor-pointer">
                <Link href="/settings">
                  <UserIcon className="mr-2 h-4 w-4" />
                  <span>{t("common.profile")}</span>
                </Link>
              </DropdownMenuItem>
              <DropdownMenuItem asChild className="cursor-pointer">
                <Link href="/settings">
                  <Settings className="mr-2 h-4 w-4" />
                  <span>{t("common.settings")}</span>
                </Link>
              </DropdownMenuItem>
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={logout} className="text-destructive cursor-pointer focus:bg-destructive/10 focus:text-destructive">
              <LogOut className="mr-2 h-4 w-4" />
              <span>{t("common.logout")}</span>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  )
}
