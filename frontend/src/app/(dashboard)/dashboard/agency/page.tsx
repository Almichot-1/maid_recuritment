"use client"

import { useRouter } from "next/navigation"
import * as React from "react"
import { Loader2 } from "lucide-react"

import { EthiopianDashboard } from "@/components/dashboard/ethiopian-dashboard"
import { useCurrentUser } from "@/hooks/use-auth"
import { getRoleHomePath } from "@/lib/role-home"

export default function AgencyHomePage() {
  const router = useRouter()
  const { user, isEthiopianAgent } = useCurrentUser()

  React.useEffect(() => {
    if (user && !isEthiopianAgent) {
      router.replace(getRoleHomePath(user.role))
    }
  }, [isEthiopianAgent, router, user])

  if (!user || !isEthiopianAgent) {
    return (
      <div className="flex h-[50vh] w-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return <EthiopianDashboard />
}
