"use client"

import * as React from "react"
import { Clock3, ShieldCheck, UsersRound } from "lucide-react"

import { PageHeader } from "@/components/layout/page-header"
import { Card, CardContent } from "@/components/ui/card"
import { useI18n } from "@/lib/i18n"

export function WaitingForPairingState() {
  const { t } = useI18n()

  return (
    <div className="space-y-6 animate-in">
      <PageHeader heading={t("waiting.heading")} text={t("waiting.body")} />

      <Card>
        <CardContent className="grid gap-6 p-6 lg:grid-cols-[minmax(0,1fr)_320px]">
          <div className="space-y-4">
            <p className="section-kicker">{t("waiting.heroLabel")}</p>
            <h2 className="font-display text-4xl text-foreground">{t("waiting.heroTitle")}</h2>
            <p className="max-w-2xl text-sm text-muted-foreground sm:text-base">{t("waiting.heroBody")}</p>
          </div>

          <div className="space-y-3 border-l border-border pl-0 lg:pl-6">
            <StatusTile icon={<ShieldCheck className="h-5 w-5" />} title={t("waiting.card1Title")} description={t("waiting.card1Body")} />
            <StatusTile icon={<UsersRound className="h-5 w-5" />} title={t("waiting.card2Title")} description={t("waiting.card2Body")} />
            <StatusTile icon={<Clock3 className="h-5 w-5" />} title={t("waiting.card3Title")} description={t("waiting.card3Body")} />
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

function StatusTile({
  icon,
  title,
  description,
}: {
  icon: React.ReactNode
  title: string
  description: string
}) {
  return (
    <div className="grid grid-cols-[44px_minmax(0,1fr)] gap-4 border border-border bg-background p-4">
      <div className="flex h-11 w-11 items-center justify-center border border-border text-primary">{icon}</div>
      <div>
        <p className="text-sm font-bold uppercase tracking-[0.06em] text-foreground">{title}</p>
        <p className="mt-1 text-sm text-muted-foreground">{description}</p>
      </div>
    </div>
  )
}
