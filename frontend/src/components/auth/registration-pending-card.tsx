"use client"

import Link from "next/link"
import { Clock3, Mail, ShieldCheck } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { useI18n } from "@/lib/i18n"

interface RegistrationPendingCardProps {
  companyName: string
  email: string
  roleLabel: string
}

export function RegistrationPendingCard({ companyName, email, roleLabel }: RegistrationPendingCardProps) {
  const { t } = useI18n()

  return (
    <Card className="w-full max-w-2xl border-border">
      <CardHeader className="space-y-4 pb-4 text-center">
        <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-primary/10 text-primary">
          <Clock3 className="h-8 w-8" aria-hidden />
        </div>
        <div className="space-y-2">
          <CardTitle className="text-3xl font-semibold tracking-tight">{t("auth.pendingTitle")}</CardTitle>
          <CardDescription className="text-base">
            {t("auth.pendingBody", { company: companyName, role: roleLabel })}
          </CardDescription>
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="grid gap-4 md:grid-cols-3">
          <div className="rounded-2xl border border-border bg-muted/30 p-4">
            <ShieldCheck className="mb-3 h-5 w-5 text-primary" aria-hidden />
            <h2 className="font-medium text-foreground">{t("auth.pendingStep1Title")}</h2>
            <p className="mt-2 text-sm text-muted-foreground">{t("auth.pendingStep1Body")}</p>
          </div>
          <div className="rounded-2xl border border-border bg-muted/30 p-4">
            <Mail className="mb-3 h-5 w-5 text-primary" aria-hidden />
            <h2 className="font-medium text-foreground">{t("auth.pendingStep2Title")}</h2>
            <p className="mt-2 text-sm text-muted-foreground">{t("auth.pendingStep2Body", { email })}</p>
          </div>
          <div className="rounded-2xl border border-border bg-muted/30 p-4">
            <Clock3 className="mb-3 h-5 w-5 text-primary" aria-hidden />
            <h2 className="font-medium text-foreground">{t("auth.pendingStep3Title")}</h2>
            <p className="mt-2 text-sm text-muted-foreground">{t("auth.pendingStep3Body")}</p>
          </div>
        </div>

        <p className="rounded-2xl border border-dashed border-border bg-card p-5 text-sm text-muted-foreground">
          {t("auth.pendingSupport")}
        </p>

        <div className="flex flex-col gap-3 sm:flex-row">
          <Button className="flex-1" asChild>
            <Link href="/login">{t("auth.pendingLogin")}</Link>
          </Button>
          <Button variant="outline" className="flex-1 border-border bg-background hover:border-primary hover:bg-primary/5" asChild>
            <Link href="/">{t("auth.pendingHome")}</Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
