"use client"

import Link from "next/link"
import { Clock3, ShieldCheck, Mail, Send } from "lucide-react"

import { useResendVerification } from "@/hooks/use-auth"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

interface PendingClientProps {
  email: string
  companyName: string
  roleLabel: string
}

export function PendingClient({ email, companyName, roleLabel }: PendingClientProps) {
  const resendVerification = useResendVerification()

  return (
    <Card className="w-full max-w-2xl border-slate-200 shadow-xl">
      <CardHeader className="space-y-4 pb-4 text-center">
        <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-amber-100 text-amber-700">
          <Clock3 className="h-8 w-8" />
        </div>
        <div className="space-y-2">
          <CardTitle className="text-3xl font-semibold tracking-tight">Verify your email to continue</CardTitle>
          <CardDescription className="text-base">
            We created the registration for <span className="font-medium text-foreground">{companyName}</span>.
            First verify <span className="font-medium text-foreground">{email}</span>, then the admin review for this {roleLabel.toLowerCase()} can continue.
          </CardDescription>
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="grid gap-4 md:grid-cols-3">
          <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
            <Mail className="mb-3 h-5 w-5 text-sky-600" />
            <h2 className="font-medium text-slate-950">Step 1</h2>
            <p className="mt-2 text-sm text-slate-600">
              Open the verification email we just sent and click the secure verification link.
            </p>
          </div>
          <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
            <ShieldCheck className="mb-3 h-5 w-5 text-emerald-600" />
            <h2 className="font-medium text-slate-950">Step 2</h2>
            <p className="mt-2 text-sm text-slate-600">
              Once the email is confirmed, the admin review starts and login stays blocked until approval.
            </p>
          </div>
          <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
            <Clock3 className="mb-3 h-5 w-5 text-amber-600" />
            <h2 className="font-medium text-slate-950">Need another link?</h2>
            <p className="mt-2 text-sm text-slate-600">
              If you do not see the email, check spam first, then request a fresh verification message.
            </p>
          </div>
        </div>

        <div className="rounded-2xl border border-dashed border-slate-300 bg-white p-5 text-sm text-slate-600">
          This keeps agencies tied to real inboxes before they enter the approval workflow.
        </div>

        <div className="flex flex-col gap-3 sm:flex-row">
          <Button
            className="flex-1"
            onClick={() => resendVerification.mutate(email)}
            disabled={resendVerification.isPending}
          >
            <Send className="mr-2 h-4 w-4" />
            {resendVerification.isPending ? "Sending..." : "Resend verification email"}
          </Button>
          <Link href="/login" className="flex-1">
            <Button variant="outline" className="w-full">Go To Login</Button>
          </Link>
        </div>
      </CardContent>
    </Card>
  )
}

