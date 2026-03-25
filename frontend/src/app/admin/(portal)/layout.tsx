"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { Loader2 } from "lucide-react"

import { AdminShell } from "@/components/admin/admin-shell"
import { useAdminAuthStore } from "@/stores/admin-auth-store"

export default function AdminPortalLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter()
  const { isAuthenticated, isLoading, loadFromStorage } = useAdminAuthStore()

  React.useEffect(() => {
    loadFromStorage()
  }, [loadFromStorage])

  React.useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.replace("/admin/login")
    }
  }, [isAuthenticated, isLoading, router])

  if (isLoading || !isAuthenticated) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background text-foreground">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    )
  }

  return <AdminShell>{children}</AdminShell>
}
