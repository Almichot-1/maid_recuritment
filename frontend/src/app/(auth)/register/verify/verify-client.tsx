"use client"

import * as React from "react"
import Link from "next/link"
import { CheckCircle2, Loader2, MailWarning } from "lucide-react"

import { useVerifyEmail } from "@/hooks/use-auth"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

interface VerifyClientProps {
  token: string
}

export function VerifyClient({ token }: VerifyClientProps) {
  const verifyEmail = useVerifyEmail()
  const [hasTriggered, setHasTriggered] = React.useState(false)

  React.useEffect(() => {
    if (!token || hasTriggered) {
      return
    }
    setHasTriggered(true)
    verifyEmail.mutate(token)
  }, [hasTriggered, token, verifyEmail])

  const isLoading = verifyEmail.isPending || (!hasTriggered && !!token)
  const isSuccess = verifyEmail.isSuccess

  return (
    <Card className="w-full max-w-xl border-slate-200 shadow-xl">
      <CardHeader className="space-y-4 pb-4 text-center">
        <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-emerald-100 text-emerald-700">
          {isLoading ? <Loader2 className="h-8 w-8 animate-spin" /> : isSuccess ? <CheckCircle2 className="h-8 w-8" /> : <MailWarning className="h-8 w-8" />}
        </div>
        <div className="space-y-2">
          <CardTitle className="text-3xl font-semibold tracking-tight">
            {isLoading ? "Verifying your email" : isSuccess ? "Email verified" : "Verification link issue"}
          </CardTitle>
          <CardDescription className="text-base">
            {isLoading
              ? "Please wait while we confirm your email address."
              : isSuccess
                ? "Your email is confirmed. The agency can now move into admin review."
                : "That verification link is missing, invalid, or expired."}
          </CardDescription>
        </div>
      </CardHeader>
      <CardContent className="flex flex-col gap-3 sm:flex-row">
        <Link href="/login" className="flex-1">
          <Button className="w-full">Go To Login</Button>
        </Link>
        <Link href="/register" className="flex-1">
          <Button variant="outline" className="w-full">Back To Register</Button>
        </Link>
      </CardContent>
    </Card>
  )
}

