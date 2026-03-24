"use client"

import * as React from "react"
import Link from "next/link"
import dynamic from "next/dynamic"
import { usePathname } from "next/navigation"
import { Search, LogOut, User as UserIcon, Settings } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
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
import { useCurrentUser, useLogout } from "@/hooks/use-auth"
import { isRoleHomePath } from "@/lib/role-home"

const NotificationBell = dynamic(
  () => import("@/components/notifications/notification-bell").then((module) => module.NotificationBell)
)

export function Header() {
  const pathname = usePathname()
  const { user } = useCurrentUser()
  const logout = useLogout()
  
  // Format pathname roughly to a readable string
  const formatPathname = (path: string) => {
    if (path.startsWith("/dashboard/agency")) return "Agency Home"
    if (path.startsWith("/dashboard/employer")) return "Employer Home"
    if (isRoleHomePath(path)) return "Overview"
    if (path.startsWith("/partners")) return "Partner Workspaces"
    if (path.startsWith("/waiting")) return "Partner Assignment"
    if (path.startsWith("/tracking")) return "Process Tracking"
    const parts = path.split('/').filter(Boolean)
    const page = parts[parts.length - 1]
    return page.charAt(0).toUpperCase() + page.slice(1).replace('-', ' ')
  }

  const title = formatPathname(pathname)

  if (!user) return null

  return (
    <header className="flex h-14 items-center gap-4 border-b bg-background px-4 lg:px-6 sticky top-0 z-40 w-full shadow-sm md:shadow-none min-w-0">
      <div className="flex-1 flex items-center">
        <h1 className="text-lg font-semibold tracking-tight truncate hidden md:block">
          {title}
        </h1>
      </div>
      
      {/* Top right actions */}
      <div className="flex items-center gap-2 md:gap-4 ml-auto">
        <form className="hidden sm:block">
          <div className="relative">
            <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              type="search"
              placeholder="Search..."
              className="w-full sm:w-[200px] md:w-[300px] pl-8 bg-muted/50 rounded-full h-9 border-none focus-visible:ring-1"
            />
          </div>
        </form>
        
        <NotificationBell />
        
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" className="relative h-8 w-8 rounded-full ml-1">
              <Avatar className="h-8 w-8">
                <AvatarImage src={`https://api.dicebear.com/7.x/initials/svg?seed=${user.full_name}`} alt={user.full_name} />
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
                  <span>Profile</span>
                </Link>
              </DropdownMenuItem>
              <DropdownMenuItem asChild className="cursor-pointer">
                <Link href="/settings">
                  <Settings className="mr-2 h-4 w-4" />
                  <span>Settings</span>
                </Link>
              </DropdownMenuItem>
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={logout} className="text-destructive cursor-pointer focus:bg-destructive/10 focus:text-destructive">
              <LogOut className="mr-2 h-4 w-4" />
              <span>Log out</span>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  )
}
