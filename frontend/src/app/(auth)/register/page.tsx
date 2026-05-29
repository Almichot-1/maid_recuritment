"use client"

import * as React from "react"
import Link from "next/link"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { AlertCircle, Eye, EyeOff, Globe, Loader2, Users } from "lucide-react"

import { registerSchema, type RegisterFormInput } from "@/lib/validations"
import { getRegisterErrorMessage, useRegister } from "@/hooks/use-auth"
import { UserRole } from "@/types"
import { AgencyAuthSessionGate } from "@/components/auth/agency-auth-session-gate"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Checkbox } from "@/components/ui/checkbox"
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form"
import { cn } from "@/lib/utils"
import { useI18n } from "@/lib/i18n"

function calculateStrength(password: string): number {
  if (!password) return 0
  let strength = 0
  if (password.length >= 8) strength += 1
  if (/\d/.test(password)) strength += 1
  if (/[!@#$%^&*(),.?":{}|<>]/.test(password)) strength += 1
  return strength
}

export default function RegisterPage() {
  const { mutate: register, isPending } = useRegister()
  const { t } = useI18n()

  const [showPassword, setShowPassword] = React.useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = React.useState(false)

  const form = useForm<RegisterFormInput>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      email: "",
      password: "",
      confirmPassword: "",
      full_name: "",
      role: UserRole.ETHIOPIAN_AGENT,
      company_name: "",
      acceptTerms: false,
    },
  })

  const passwordValue = form.watch("password")
  const watchedEmail = form.watch("email")
  const strength = calculateStrength(passwordValue)

  React.useEffect(() => {
    if (form.formState.errors.root?.message) {
      form.clearErrors("root")
    }
  }, [watchedEmail, form])

  function onSubmit(data: RegisterFormInput) {
    form.clearErrors("root")
    register(data, {
      onError: (error) => {
        form.setError("root", {
          type: "server",
          message: getRegisterErrorMessage(error),
        })
      },
    })
  }

  const strengthLabel =
    strength === 1
      ? t("auth.passwordWeak")
      : strength === 2
        ? t("auth.passwordMedium")
        : strength === 3
          ? t("auth.passwordStrong")
          : ""

  return (
    <AgencyAuthSessionGate>
      <Card className="my-8 w-full max-w-3xl">
        <CardHeader className="space-y-2 border-b border-border pb-5">
          <CardTitle className="font-display text-4xl">{t("auth.registerTitle")}</CardTitle>
          <CardDescription>{t("auth.registerBody")}</CardDescription>
        </CardHeader>
        <CardContent className="pt-6">
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-5">
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <FormField
                  control={form.control}
                  name="full_name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t("auth.fullName")}</FormLabel>
                      <FormControl>
                        <Input placeholder={t("auth.fullNamePlaceholder")} {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="company_name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t("auth.companyName")}</FormLabel>
                      <FormControl>
                        <Input placeholder={t("auth.companyNamePlaceholder")} {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>

              <FormField
                control={form.control}
                name="email"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("auth.email")}</FormLabel>
                    <FormControl>
                      <Input placeholder={t("auth.emailPlaceholder")} type="email" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              {form.formState.errors.root?.message ? (
                <div className="border border-destructive bg-destructive/10 px-4 py-3 text-sm text-destructive" role="alert">
                  <div className="flex items-start gap-3">
                    <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
                    <span>{form.formState.errors.root.message}</span>
                  </div>
                </div>
              ) : null}

              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
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
                          </Button>
                        </div>
                      </FormControl>
                      <div className="mt-2 flex h-1.5 w-full gap-1 overflow-hidden bg-muted">
                        <div className={cn("h-full flex-1", strength >= 1 ? "bg-destructive" : "bg-transparent")} />
                        <div className={cn("h-full flex-1", strength >= 2 ? "bg-[color:var(--color-warning)]" : "bg-transparent")} />
                        <div className={cn("h-full flex-1", strength >= 3 ? "bg-[color:var(--color-success)]" : "bg-transparent")} />
                      </div>
                      <p className="min-h-[16px] text-xs text-muted-foreground">{strengthLabel}</p>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="confirmPassword"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t("auth.confirmPassword")}</FormLabel>
                      <FormControl>
                        <div className="relative">
                          <Input
                            type={showConfirmPassword ? "text" : "password"}
                            placeholder={t("auth.passwordPlaceholder")}
                            className="pr-10"
                            {...field}
                          />
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            className="absolute right-0 top-0 h-full px-3 py-2"
                            onClick={() => setShowConfirmPassword((value) => !value)}
                          >
                            {showConfirmPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                          </Button>
                        </div>
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>

              <FormField
                control={form.control}
                name="role"
                render={({ field }) => (
                  <FormItem className="space-y-3 pt-2">
                    <FormLabel>{t("auth.accountType")}</FormLabel>
                    <FormControl>
                      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                        <button
                          type="button"
                          aria-pressed={field.value === UserRole.ETHIOPIAN_AGENT}
                          aria-label={t("auth.ethiopianAgency")}
                          className={cn(
                            "border p-4 text-left transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2",
                            field.value === UserRole.ETHIOPIAN_AGENT ? "border-primary bg-primary/5" : "border-border bg-card"
                          )}
                          onClick={() => field.onChange(UserRole.ETHIOPIAN_AGENT)}
                        >
                          <div className="flex items-start gap-3">
                            <Users className="mt-0.5 h-5 w-5 text-primary" />
                            <div className="space-y-1">
                              <p className="text-sm font-bold uppercase tracking-[0.06em]">{t("auth.ethiopianAgency")}</p>
                              <p className="text-xs text-muted-foreground">{t("auth.ethiopianAgencyBody")}</p>
                            </div>
                          </div>
                        </button>

                        <button
                          type="button"
                          aria-pressed={field.value === UserRole.FOREIGN_AGENT}
                          aria-label={t("auth.foreignAgency")}
                          className={cn(
                            "border p-4 text-left transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2",
                            field.value === UserRole.FOREIGN_AGENT ? "border-primary bg-primary/5" : "border-border bg-card"
                          )}
                          onClick={() => field.onChange(UserRole.FOREIGN_AGENT)}
                        >
                          <div className="flex items-start gap-3">
                            <Globe className="mt-0.5 h-5 w-5 text-primary" />
                            <div className="space-y-1">
                              <p className="text-sm font-bold uppercase tracking-[0.06em]">{t("auth.foreignAgency")}</p>
                              <p className="text-xs text-muted-foreground">{t("auth.foreignAgencyBody")}</p>
                            </div>
                          </div>
                        </button>
                      </div>
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="acceptTerms"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-start space-x-3 space-y-0 pt-2">
                    <FormControl>
                      <Checkbox checked={field.value} onCheckedChange={field.onChange} />
                    </FormControl>
                    <div className="space-y-1 leading-none">
                      <FormLabel className="cursor-pointer font-normal">{t("auth.acceptTerms")}</FormLabel>
                    </div>
                  </FormItem>
                )}
              />

              <Button type="submit" className="w-full" disabled={isPending}>
                {isPending ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin" />
                    {t("auth.submittingRegistration")}
                  </>
                ) : (
                  t("auth.submitRegistration")
                )}
              </Button>
            </form>
          </Form>
        </CardContent>
        <CardFooter className="border-t border-border pt-4 text-sm text-muted-foreground">
          <div>
            {t("auth.haveAccount")}{" "}
            <Link href="/login" className="text-primary hover:underline">
              {t("common.login")}
            </Link>
          </div>
        </CardFooter>
      </Card>
    </AgencyAuthSessionGate>
  )
}
