"use client"

import * as React from "react"
import Link from "next/link"
import { usePathname } from "next/navigation"
import { Home, Users, CheckSquare, Menu, MessageSquare } from "lucide-react"
import { cn } from "@/lib/utils"
import { useCurrentUser } from "@/hooks/use-auth"
import { usePairingContext } from "@/hooks/use-pairings"
import { useChatSummary } from "@/hooks/use-chat"
import { getRoleHomePath, isRoleHomePath } from "@/lib/role-home"

function matchesNavPath(pathname: string, href: string) {
  if (isRoleHomePath(href)) {
    return isRoleHomePath(pathname)
  }

  if (pathname === href) {
    return true
  }

  return pathname.startsWith(`${href}/`)
}

export function BottomNav() {
  const pathname = usePathname()
  const { user } = useCurrentUser()
  const { hasActivePairs, isReady } = usePairingContext()
  const { count: chatUnreadCount } = useChatSummary()
  const [isKeyboardOpen, setIsKeyboardOpen] = React.useState(false)
  const navItems = [
    {
      label: "Home",
      href: getRoleHomePath(user?.role),
      icon: Home,
    },
    {
      label: "Candidates",
      href: "/candidates",
      icon: Users,
    },
    {
      label: "Selections",
      href: "/selections",
      icon: CheckSquare,
    },
    {
      label: "Chat",
      href: "/partners/chat",
      icon: MessageSquare,
    },
    {
      label: "More",
      href: "/settings",
      icon: Menu,
    },
  ]

  React.useEffect(() => {
    const handleResize = () => {
      // Detect if keyboard is open by checking if viewport height decreased significantly
      const viewportHeight = window.visualViewport?.height || window.innerHeight
      const windowHeight = window.innerHeight
      setIsKeyboardOpen(viewportHeight < windowHeight * 0.75)
    }

    window.visualViewport?.addEventListener("resize", handleResize)
    window.addEventListener("resize", handleResize)

    return () => {
      window.visualViewport?.removeEventListener("resize", handleResize)
      window.removeEventListener("resize", handleResize)
    }
  }, [])

  if (isKeyboardOpen || (isReady && !hasActivePairs)) {
    return null
  }

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 bg-background border-t md:hidden">
      <div className="grid grid-cols-5 h-16">
        {navItems.map((item) => {
          const Icon = item.icon
          const active = matchesNavPath(pathname, item.href)
          const isChat = item.href === "/partners/chat"

          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex flex-col items-center justify-center gap-1 transition-colors min-h-[44px]",
                active
                  ? "text-primary"
                  : "text-muted-foreground hover:text-foreground"
              )}
            >
              <span className="relative">
                <Icon className="h-5 w-5" />
                {isChat && chatUnreadCount > 0 ? (
                  <span className="absolute -top-1.5 -right-2.5 rounded-full bg-destructive px-1 py-0.5 text-[9px] font-semibold leading-none text-destructive-foreground">
                    {chatUnreadCount > 99 ? "99+" : chatUnreadCount}
                  </span>
                ) : null}
              </span>
              <span className="text-xs font-medium">{item.label}</span>
            </Link>
          )
        })}
      </div>
    </nav>
  )
}
