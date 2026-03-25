"use client"

import * as React from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { ArrowLeft, Loader2, MailCheck } from "lucide-react"

import { forgotPasswordRequestSchema, type ForgotPasswordRequestInput } from "@/lib/validations"
import { useRequestPasswordReset } from "@/hooks/use-auth"
import { useAuthStore } from "@/stores/auth-store"
import { getRoleHomePath } from "@/lib/role-home"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form"
import { Input } from "@/components/ui/input"

export default function ForgotPasswordPage() {
  const router = useRouter()
  const { mutate: requestReset, isPending } = useRequestPasswordReset()
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const user = useAuthStore((state) => state.user)
  const [submittedEmail, setSubmittedEmail] = React.useState("")

  React.useEffect(() => {
    if (isAuthenticated) {
      router.replace(getRoleHomePath(user?.role))
    }
  }, [isAuthenticated, router, user?.role])

  const form = useForm<ForgotPasswordRequestInput>({
    resolver: zodResolver(forgotPasswordRequestSchema),
    defaultValues: {
      email: "",
    },
  })

  function onSubmit(data: ForgotPasswordRequestInput) {
    requestReset(data, {
      onSuccess: () => {
        setSubmittedEmail(data.email)
      },
    })
  }

  if (isAuthenticated) {
    return null
  }

  return (
    <Card className="animated-border w-full max-w-md overflow-hidden border-muted shadow-glow">
      <CardHeader className="space-y-2 pb-6 text-center">
        <CardTitle className="text-2xl font-bold tracking-tight">Forgot Password</CardTitle>
        <CardDescription>
          Enter your agency email and we will send a one-time reset code.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {submittedEmail ? (
          <div className="space-y-4 rounded-3xl border border-primary/20 bg-primary/5 p-5 text-sm">
            <div className="flex items-start gap-3">
              <MailCheck className="mt-0.5 h-5 w-5 text-primary" />
              <div className="space-y-2">
                <p className="font-medium text-foreground">Check your email</p>
                <p className="text-muted-foreground">
                  If an active account exists for <span className="font-medium text-foreground">{submittedEmail}</span>, we sent a 6-digit reset code.
                </p>
              </div>
            </div>
            <Button asChild className="w-full">
              <Link href={`/reset-password?email=${encodeURIComponent(submittedEmail)}`}>Continue to reset password</Link>
            </Button>
          </div>
        ) : (
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

              <Button type="submit" className="w-full" disabled={isPending}>
                {isPending ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Sending code...
                  </>
                ) : (
                  "Send reset code"
                )}
              </Button>
            </form>
          </Form>
        )}
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
