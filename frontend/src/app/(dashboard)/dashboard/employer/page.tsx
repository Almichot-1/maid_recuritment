"use client"

import { useRouter } from "next/navigation"
import * as React from "react"
import { Loader2 } from "lucide-react"

import { ForeignDashboard } from "@/components/dashboard/foreign-dashboard"
import { useCurrentUser } from "@/hooks/use-auth"
import { getRoleHomePath } from "@/lib/role-home"

export default function EmployerHomePage() {
  const router = useRouter()
  const { user, isForeignAgent } = useCurrentUser()

  React.useEffect(() => {
    if (user && !isForeignAgent) {
      router.replace(getRoleHomePath(user.role))
    }
  }, [isForeignAgent, router, user])

  if (!user || !isForeignAgent) {
    return (
      <div className="flex h-[50vh] w-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return <ForeignDashboard />
}
