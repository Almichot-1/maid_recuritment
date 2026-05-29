"use client"

import * as React from "react"
import Link from "next/link"
import { Loader2, LogOut } from "lucide-react"

import api from "@/lib/api"
import { getRoleHomePath } from "@/lib/role-home"
import { useAuthStore } from "@/stores/auth-store"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

interface AgencyAuthSessionGateProps {
  children: React.ReactNode
}

export function AgencyAuthSessionGate({ children }: AgencyAuthSessionGateProps) {
  const { user, isAuthenticated, isLoading, loadFromStorage, logout } = useAuthStore()
  const [isSigningOut, setIsSigningOut] = React.useState(false)

  React.useEffect(() => {
    void loadFromStorage()
  }, [loadFromStorage])

  const handleSignOut = async () => {
    setIsSigningOut(true)
    try {
      await api.post("/auth/logout")
    } catch {
      // Clear the local session even if the server cookie is already gone.
    } finally {
      logout()
      setIsSigningOut(false)
    }
  }

  if (isLoading) {
    return (
      <Card className="w-full max-w-md overflow-hidden border-muted shadow-lg">
        <CardContent className="flex min-h-[220px] items-center justify-center">
          <div className="flex items-center gap-3 text-sm text-muted-foreground">
            <Loader2 className="h-5 w-5 animate-spin" />
            Checking active session...
          </div>
        </CardContent>
      </Card>
    )
  }

  if (isAuthenticated && user) {
    return (
      <Card className="w-full max-w-md overflow-hidden border-muted shadow-lg">
        <CardHeader className="space-y-2 text-center">
          <CardTitle className="text-2xl font-bold tracking-tight">You are already signed in</CardTitle>
          <CardDescription>
            This browser is currently signed in as <span className="font-medium text-foreground">{user.email}</span>.
            Sign out first if you want to continue with a different email.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <Button asChild className="w-full">
            <Link href={getRoleHomePath(user.role)}>Open dashboard</Link>
          </Button>
          <Button variant="outline" className="w-full" onClick={handleSignOut} disabled={isSigningOut}>
            {isSigningOut ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Signing out...
              </>
            ) : (
              <>
                <LogOut className="mr-2 h-4 w-4" />
                Sign out and continue
              </>
            )}
          </Button>
        </CardContent>
      </Card>
    )
  }

  return <>{children}</>
}
