"use client"

import * as React from "react"
import Link from "next/link"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { AlertCircle, Eye, EyeOff, Loader2, Mail } from "lucide-react"

import { loginSchema, type LoginInput } from "@/lib/validations"
import { getLoginErrorMessage, useLogin } from "@/hooks/use-auth"
import { AgencyAuthSessionGate } from "@/components/auth/agency-auth-session-gate"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Checkbox } from "@/components/ui/checkbox"
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import { useI18n } from "@/lib/i18n"

export default function LoginPage() {
  const { mutate: login, isPending } = useLogin()
  const { t } = useI18n()
  const [showPassword, setShowPassword] = React.useState(false)

  const form = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: "",
      password: "",
    },
  })

  const watchedEmail = form.watch("email")
  const watchedPassword = form.watch("password")

  React.useEffect(() => {
    if (form.formState.errors.root?.message) {
      form.clearErrors("root")
    }
  }, [watchedEmail, watchedPassword, form])

  function onSubmit(data: LoginInput) {
    form.clearErrors("root")
    login(data, {
      onError: (error) => {
        form.setError("root", {
          type: "server",
          message: getLoginErrorMessage(error),
        })
      },
    })
  }

  return (
    <AgencyAuthSessionGate>
      <Card className="w-full max-w-xl">
        <CardHeader className="space-y-2 border-b border-border pb-5">
          <CardTitle className="font-display text-4xl">{t("auth.loginTitle")}</CardTitle>
          <CardDescription>{t("auth.loginBody")}</CardDescription>
        </CardHeader>
        <CardContent className="pt-6">
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-5">
              <FormField
                control={form.control}
                name="email"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("auth.email")}</FormLabel>
                    <FormControl>
                      <div className="relative">
                        <Mail className="absolute left-3 top-3.5 h-4 w-4 text-muted-foreground" />
                        <Input
                          placeholder={t("auth.emailPlaceholder")}
                          className="pl-9"
                          autoComplete="username"
                          {...field}
                        />
                      </div>
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="password"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("auth.password")}</FormLabel>
                    <FormControl>
                      <div className="relative">
                        <Input
                          type={showPassword ? "text" : "password"}
                          placeholder={t("auth.passwordPlaceholder")}
                          className="pr-10"
                          autoComplete="current-password"
                          {...field}
                        />
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          className="absolute right-0 top-0 h-full px-3 py-2"
                          onClick={() => setShowPassword((value) => !value)}
                        >
                          {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                          <span className="sr-only">Toggle password visibility</span>
                        </Button>
                      </div>
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <div className="flex items-center justify-between gap-4">
                <div className="flex items-center gap-2">
                  <Checkbox id="remember" />
                  <label htmlFor="remember" className="text-sm text-muted-foreground">
                    {t("auth.remember")}
                  </label>
                </div>
                <Link href="/forgot-password" aria-label="Open forgot password page" className="text-sm text-primary hover:underline">
                  {t("auth.forgot")}
                </Link>
              </div>

              {form.formState.errors.root?.message ? (
                <div className="border border-destructive bg-destructive/10 px-4 py-3 text-sm text-destructive" role="alert">
                  <div className="flex items-start gap-3">
                    <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
                    <span>{form.formState.errors.root.message}</span>
                  </div>
                </div>
              ) : null}

              <Button type="submit" className="mt-2 w-full" disabled={isPending}>
                {isPending ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin" />
                    {t("auth.signingIn")}
                  </>
                ) : (
                  t("auth.signIn")
                )}
              </Button>
            </form>
          </Form>
        </CardContent>
        <CardFooter className="border-t border-border pt-4 text-sm text-muted-foreground">
          <div>
            {t("auth.needAccount")}{" "}
            <Link href="/register" className="text-primary hover:underline">
              {t("auth.registerLink")}
            </Link>
          </div>
        </CardFooter>
      </Card>
    </AgencyAuthSessionGate>
  )
}
