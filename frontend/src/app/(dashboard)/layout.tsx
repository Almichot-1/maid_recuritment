"use client"

import * as React from "react"
import { usePathname, useRouter } from "next/navigation"
import { Sidebar } from "@/components/layout/sidebar"
import { Header } from "@/components/layout/header"
import { MobileHeader } from "@/components/layout/mobile-header"
import { BottomNav } from "@/components/layout/bottom-nav"
import { usePairingContext } from "@/hooks/use-pairings"
import { useAuthStore } from "@/stores/auth-store"
import { useLayoutStore } from "@/stores/layout-store"
import { Loader2 } from "lucide-react"
import { cn } from "@/lib/utils"
import { getRoleHomePath } from "@/lib/role-home"
import { UserRole } from "@/types"

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter()
  const pathname = usePathname()
  const { user, isAuthenticated, isLoading, loadFromStorage } = useAuthStore()
  const { isSidebarCollapsed } = useLayoutStore()
  const { hasActivePairs, isLoading: pairingLoading, isReady } = usePairingContext()

  // Hydrate user store
  React.useEffect(() => {
    loadFromStorage()
  }, [loadFromStorage])

  // Redirect if hydrated and not authenticated
  React.useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push("/login")
    }
  }, [isLoading, isAuthenticated, router])

  React.useEffect(() => {
    if (isLoading || !isAuthenticated || !isReady || pairingLoading) {
      return
    }

    const role = user?.role
    const isAgencyUser = role === UserRole.ETHIOPIAN_AGENT || role === UserRole.FOREIGN_AGENT
    if (!isAgencyUser) {
      return
    }

    if (!hasActivePairs && pathname !== "/waiting") {
      router.replace("/waiting")
      return
    }

    if (hasActivePairs && pathname === "/waiting") {
      router.replace(getRoleHomePath(role))
    }
  }, [hasActivePairs, isAuthenticated, isLoading, isReady, pairingLoading, pathname, router, user?.role])

  const role = user?.role
  const isAgencyUser = role === UserRole.ETHIOPIAN_AGENT || role === UserRole.FOREIGN_AGENT
  const isRoutingToWaiting = isAgencyUser && isReady && !pairingLoading && !hasActivePairs && pathname !== "/waiting"
  const isRoutingAwayFromWaiting = isAgencyUser && isReady && !pairingLoading && hasActivePairs && pathname === "/waiting"

  if (isLoading || !isAuthenticated || (isAgencyUser && (!isReady || pairingLoading)) || isRoutingToWaiting || isRoutingAwayFromWaiting) {
    return (
      <div className="flex h-screen w-full items-center justify-center bg-background">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className="flex min-h-screen w-full bg-background">
      <Sidebar />
      <div 
        className={cn(
          "flex flex-1 flex-col transition-all duration-300 ease-in-out min-w-0",
          isSidebarCollapsed ? "md:pl-16" : "md:pl-64"
        )}
      >
        {/* Desktop Header */}
        <div className="hidden md:block">
          <Header />
        </div>
        
        {/* Mobile Header */}
        <MobileHeader />
        
        <main className="flex-1 p-4 md:p-6 lg:p-8 overflow-x-hidden pb-20 md:pb-8">
          {children}
        </main>
      </div>
      
      {/* Mobile Bottom Navigation */}
      <BottomNav />
    </div>
  )
}
