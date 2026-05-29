"use client"

import * as React from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import axios from "axios"
import { Eye, EyeOff, KeyRound, Loader2, Mail, ShieldCheck } from "lucide-react"

import { useAdminLogin } from "@/hooks/use-admin-auth"
import { useAdminAuthStore } from "@/stores/admin-auth-store"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

export default function AdminLoginPage() {
  const router = useRouter()
  const { mutate: login, isPending, error, reset } = useAdminLogin()
  const isAuthenticated = useAdminAuthStore((state) => state.isAuthenticated)

  const [email, setEmail] = React.useState("")
  const [password, setPassword] = React.useState("")
  const [otpCode, setOtpCode] = React.useState("")
  const [showPassword, setShowPassword] = React.useState(false)

  React.useEffect(() => {
    if (isAuthenticated) {
      router.replace("/admin/dashboard")
    }
  }, [isAuthenticated, router])

  const feedbackMessage = React.useMemo(() => {
    if (!error) {
      return ""
    }

    if (axios.isAxiosError<{ error?: string }>(error)) {
      const status = error.response?.status
      if (status === 423) {
        return "This admin account is temporarily locked after repeated failed attempts. Please wait 15 minutes, then try again."
      }

      if (status === 401) {
        return "The email, password, or MFA code is not correct. Your entries were kept so you can fix only the wrong part."
      }

      return error.response?.data?.error || "We could not sign you in right now."
    }

    return "We could not sign you in right now."
  }, [error])

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    reset()
    login({ email, password, otp_code: otpCode })
  }

  if (isAuthenticated) {
    return null
  }

  return (
    <Card className="w-full max-w-xl border-border bg-card shadow-xl">
      <CardHeader className="space-y-4 pb-4">
        <div className="flex items-center gap-3">
          <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-primary/15 text-primary">
            <ShieldCheck className="h-6 w-6" />
          </div>
          <div>
            <CardTitle className="text-3xl tracking-tight">Admin Login</CardTitle>
            <CardDescription>
              Sign in with your operator credentials and six-digit MFA code.
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <form className="space-y-5" onSubmit={handleSubmit}>
          {feedbackMessage ? (
            <div className="rounded-2xl border border-destructive/35 bg-destructive/10 px-4 py-3 text-sm text-destructive">
              {feedbackMessage}
            </div>
          ) : null}

          <div className="space-y-2">
            <Label htmlFor="admin-email">Email</Label>
            <div className="relative">
              <Mail className="absolute left-3 top-3.5 h-4 w-4 text-muted-foreground" />
              <Input
                id="admin-email"
                type="email"
                autoComplete="email"
                className="pl-9"
                placeholder="admin@platform.test"
                value={email}
                onChange={(event) => {
                  if (error) {
                    reset()
                  }
                  setEmail(event.target.value)
                }}
                required
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="admin-password">Password</Label>
            <div className="relative">
              <Input
                id="admin-password"
                type={showPassword ? "text" : "password"}
                autoComplete="current-password"
                className="pr-10"
                placeholder="Enter your admin password"
                value={password}
                onChange={(event) => {
                  if (error) {
                    reset()
                  }
                  setPassword(event.target.value)
                }}
                required
              />
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="absolute right-0 top-0 h-full text-muted-foreground hover:bg-transparent hover:text-foreground"
                onClick={() => setShowPassword((value) => !value)}
              >
                {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </Button>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="admin-otp">MFA Code</Label>
            <div className="relative">
              <KeyRound className="absolute left-3 top-3.5 h-4 w-4 text-muted-foreground" />
              <Input
                id="admin-otp"
                inputMode="numeric"
                pattern="[0-9]{6}"
                maxLength={6}
                className="pl-9"
                placeholder="123456"
                value={otpCode}
                onChange={(event) => {
                  if (error) {
                    reset()
                  }
                  setOtpCode(event.target.value.replace(/\D/g, "").slice(0, 6))
                }}
                required
              />
            </div>
            <p className="text-xs text-muted-foreground">
              Enter the current six-digit code from your authenticator app. Previously shared codes expire every 30 seconds.
            </p>
          </div>

          <Button type="submit" className="w-full" disabled={isPending}>
            {isPending ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Authenticating...
              </>
            ) : (
              "Enter Admin Portal"
            )}
          </Button>
        </form>

        <div className="mt-6 rounded-2xl border border-border bg-muted/40 p-4 text-sm text-muted-foreground">
          Admin accounts are created manually by platform operators. Agency registrations cannot be used here.
        </div>

        <div className="mt-4 text-sm text-muted-foreground">
          Need the agency portal instead?{" "}
          <Link href="/login" className="font-medium text-primary hover:text-brand-strong">
            Go to agency login
          </Link>
        </div>
      </CardContent>
    </Card>
  )
}
