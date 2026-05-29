"use client"

import * as React from "react"
import { Suspense } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import { Loader2 } from "lucide-react"
import { toast } from "sonner"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { getApiBaseUrl } from "@/lib/api-base-url"

interface AdminSetupPreviewResponse {
  admin: {
    email: string
    full_name: string
  }
  provisioning_url: string
}

function AdminSetupContent() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const token = searchParams.get("token") || ""

  const [isLoading, setIsLoading] = React.useState(true)
  const [isSubmitting, setIsSubmitting] = React.useState(false)
  const [preview, setPreview] = React.useState<AdminSetupPreviewResponse | null>(null)
  const [newPassword, setNewPassword] = React.useState("")
  const [confirmPassword, setConfirmPassword] = React.useState("")
  const [otpCode, setOTPCode] = React.useState("")

  React.useEffect(() => {
    if (!token) {
      setIsLoading(false)
      return
    }

    let cancelled = false

    const loadPreview = async () => {
      try {
        const response = await fetch(`${getApiBaseUrl()}/admin/setup/preview`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ token }),
        })
        if (!response.ok) {
          throw new Error("Invalid or expired setup link")
        }
        const data = (await response.json()) as AdminSetupPreviewResponse
        if (!cancelled) {
          setPreview(data)
        }
      } catch (error) {
        if (!cancelled) {
          toast.error(error instanceof Error ? error.message : "Failed to load setup details")
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false)
        }
      }
    }

    void loadPreview()
    return () => {
      cancelled = true
    }
  }, [token])

  const handleSubmit = async () => {
    if (!token) {
      toast.error("Missing setup token")
      return
    }
    if (newPassword !== confirmPassword) {
      toast.error("Passwords do not match")
      return
    }

    try {
      setIsSubmitting(true)
      const response = await fetch(`${getApiBaseUrl()}/admin/setup/complete`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          token,
          new_password: newPassword,
          otp_code: otpCode,
        }),
      })
      const data = await response.json().catch(() => ({}))
      if (!response.ok) {
        throw new Error(data?.error || "Failed to complete admin setup")
      }
      toast.success("Admin setup completed. You can sign in now.")
      router.replace("/admin/login")
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Failed to complete admin setup")
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="mx-auto flex min-h-screen max-w-3xl items-center justify-center px-4 py-10">
      <Card className="w-full">
        <CardHeader>
          <CardTitle>Admin account setup</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          {isLoading ? (
            <div className="flex items-center justify-center py-10">
              <Loader2 className="h-6 w-6 animate-spin" />
            </div>
          ) : !preview ? (
            <div className="rounded-xl border border-destructive/20 bg-destructive/5 p-4 text-sm text-destructive">
              This setup link is invalid or has expired.
            </div>
          ) : (
            <>
              <div className="space-y-2">
                <p className="text-sm text-muted-foreground">Setting up admin access for</p>
                <p className="font-medium">{preview.admin.full_name} ({preview.admin.email})</p>
              </div>

              <div className="grid gap-6 md:grid-cols-[220px_minmax(0,1fr)]">
                <div className="overflow-hidden rounded-xl border bg-white p-3">
                  <img
                    src={`https://api.qrserver.com/v1/create-qr-code/?size=220x220&data=${encodeURIComponent(preview.provisioning_url)}`}
                    alt="MFA QR code"
                    className="h-full w-full object-contain"
                  />
                </div>
                <div className="space-y-3">
                  <p className="text-sm text-muted-foreground">
                    Scan this QR code with your authenticator app, then choose your permanent password and enter the 6-digit code from the app to activate the account.
                  </p>
                  <div className="rounded-xl border bg-muted/30 p-3 text-xs break-all">
                    {preview.provisioning_url}
                  </div>
                </div>
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <Input
                  type="password"
                  placeholder="New password"
                  value={newPassword}
                  onChange={(event) => setNewPassword(event.target.value)}
                />
                <Input
                  type="password"
                  placeholder="Confirm password"
                  value={confirmPassword}
                  onChange={(event) => setConfirmPassword(event.target.value)}
                />
              </div>

              <Input
                inputMode="numeric"
                placeholder="Authenticator OTP code"
                value={otpCode}
                onChange={(event) => setOTPCode(event.target.value)}
                maxLength={6}
              />

              <Button className="w-full" disabled={isSubmitting} onClick={handleSubmit}>
                {isSubmitting ? "Completing setup..." : "Complete admin setup"}
              </Button>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

export default function AdminSetupPage() {
  return (
    <Suspense
      fallback={
        <div className="mx-auto flex min-h-screen max-w-3xl items-center justify-center px-4 py-10">
          <Loader2 className="h-6 w-6 animate-spin" />
        </div>
      }
    >
      <AdminSetupContent />
    </Suspense>
  )
}
