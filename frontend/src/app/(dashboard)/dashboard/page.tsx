"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { useCurrentUser } from "@/hooks/use-auth"
import { Loader2 } from "lucide-react"
import { getRoleHomePath } from "@/lib/role-home"

export default function DashboardPage() {
  const router = useRouter()
  const { user, isEthiopianAgent, isForeignAgent } = useCurrentUser()

  React.useEffect(() => {
    if (!user) {
      return
    }

    if (isEthiopianAgent || isForeignAgent) {
      router.replace(getRoleHomePath(user.role))
    }
  }, [isEthiopianAgent, isForeignAgent, router, user])

  if (!user) {
    return (
      <div className="flex h-[50vh] w-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (isEthiopianAgent || isForeignAgent) {
    return (
      <div className="flex h-[50vh] w-full flex-col items-center justify-center gap-3 text-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
        <div className="space-y-1">
          <p className="font-medium">Opening your workspace</p>
          <p className="text-sm text-muted-foreground">Taking you to the page designed for your role.</p>
        </div>
      </div>
    )
  }

  // Fallback for an unknown role or admin
  return (
    <div className="p-8 text-center space-y-4 animate-in fade-in">
      <h2 className="text-2xl font-bold tracking-tight">Welcome, {user.full_name}</h2>
      <p className="text-muted-foreground">Your account role ({user.role}) is not officially mapped to a specific dashboard view hierarchy yet.</p>
    </div>
  )
}
