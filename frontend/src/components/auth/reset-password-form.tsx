"use client"

import * as React from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { ArrowLeft, Loader2, RefreshCcw } from "lucide-react"

import { forgotPasswordRequestSchema, forgotPasswordResetSchema, type ForgotPasswordResetInput } from "@/lib/validations"
import { useRequestPasswordReset, useResetForgottenPassword } from "@/hooks/use-auth"
import { useAuthStore } from "@/stores/auth-store"
import { getRoleHomePath } from "@/lib/role-home"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form"
import { Input } from "@/components/ui/input"

const resendCooldownSeconds = 60

interface ResetPasswordFormProps {
  initialEmail?: string
}

export function ResetPasswordForm({ initialEmail = "" }: ResetPasswordFormProps) {
  const router = useRouter()
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const user = useAuthStore((state) => state.user)
  const { mutate: resetPassword, isPending } = useResetForgottenPassword()
  const { mutate: resendCode, isPending: isResending } = useRequestPasswordReset()
  const [cooldownSeconds, setCooldownSeconds] = React.useState(0)

  React.useEffect(() => {
    if (isAuthenticated) {
      router.replace(getRoleHomePath(user?.role))
    }
  }, [isAuthenticated, router, user?.role])

  React.useEffect(() => {
    if (cooldownSeconds <= 0) {
      return
    }

    const timeoutId = window.setTimeout(() => {
      setCooldownSeconds((current) => Math.max(0, current - 1))
    }, 1000)

    return () => window.clearTimeout(timeoutId)
  }, [cooldownSeconds])

  const form = useForm<ForgotPasswordResetInput>({
    resolver: zodResolver(forgotPasswordResetSchema),
    defaultValues: {
      email: initialEmail,
      code: "",
      new_password: "",
      confirmPassword: "",
    },
  })

  function onSubmit(data: ForgotPasswordResetInput) {
    resetPassword(data)
  }

  const emailValue = form.watch("email")

  function handleResend() {
    const parsed = forgotPasswordRequestSchema.shape.email.safeParse(emailValue)
    if (!parsed.success) {
      form.setError("email", { type: "manual", message: "Enter a valid email before requesting a new code." })
      return
    }

    resendCode(
      { email: emailValue },
      {
        onSuccess: () => {
          setCooldownSeconds(resendCooldownSeconds)
        },
      }
    )
  }

  if (isAuthenticated) {
    return null
  }

  return (
    <Card className="animated-border w-full max-w-md overflow-hidden border-muted shadow-glow">
      <CardHeader className="space-y-2 pb-6 text-center">
        <CardTitle className="text-2xl font-bold tracking-tight">Reset Password</CardTitle>
        <CardDescription>
          Enter the 6-digit code from your email and choose a new password.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-5">
            <FormField
              control={form.control}
              name="email"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Email</FormLabel>
                  <FormControl>
                    <Input placeholder="name@example.com" autoComplete="email" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="code"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Reset code</FormLabel>
                  <FormControl>
                    <Input
                      inputMode="numeric"
                      maxLength={6}
                      placeholder="123456"
                      {...field}
                      onChange={(event) => field.onChange(event.target.value.replace(/\D/g, "").slice(0, 6))}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="new_password"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>New password</FormLabel>
                  <FormControl>
                    <Input type="password" autoComplete="new-password" placeholder="At least 8 characters" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="confirmPassword"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Confirm new password</FormLabel>
                  <FormControl>
                    <Input type="password" autoComplete="new-password" placeholder="Repeat your new password" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="flex flex-col gap-3 sm:flex-row">
              <Button type="submit" className="flex-1" disabled={isPending}>
                {isPending ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Resetting...
                  </>
                ) : (
                  "Reset password"
                )}
              </Button>

              <Button
                type="button"
                variant="outline"
                className="flex-1"
                disabled={cooldownSeconds > 0 || isResending}
                onClick={handleResend}
              >
                {isResending ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Sending...
                  </>
                ) : cooldownSeconds > 0 ? (
                  `Resend in ${cooldownSeconds}s`
                ) : (
                  <>
                    <RefreshCcw className="mr-2 h-4 w-4" />
                    Resend code
                  </>
                )}
              </Button>
            </div>
          </form>
        </Form>
      </CardContent>
      <CardFooter className="justify-center border-t pt-4 text-sm text-muted-foreground">
        <Link href="/login" className="inline-flex items-center gap-2 font-medium text-primary hover:underline">
          <ArrowLeft className="h-4 w-4" />
          Back to login
        </Link>
      </CardFooter>
    </Card>
  )
}
