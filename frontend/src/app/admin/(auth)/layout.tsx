"use client"

import * as React from "react"

import { LocaleSwitcher } from "@/components/shared/locale-switcher"
import { Logo } from "@/components/shared/logo"
import { useI18n } from "@/lib/i18n"

const adminPoints = [
  { titleKey: "admin.authPoint1Title", bodyKey: "admin.authPoint1Body" },
  { titleKey: "admin.authPoint2Title", bodyKey: "admin.authPoint2Body" },
  { titleKey: "admin.authPoint3Title", bodyKey: "admin.authPoint3Body" },
]

export default function AdminAuthLayout({ children }: { children: React.ReactNode }) {
  const { t } = useI18n()

  return (
    <div className="min-h-screen bg-background">
      <div className="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between border-b border-border pb-4">
          <Logo size="sm" />
          <LocaleSwitcher compact />
        </div>

        <div className="grid gap-10 py-10 lg:grid-cols-[minmax(0,1fr)_560px] lg:py-16">
          <div className="space-y-6">
            <div className="space-y-4">
              <p className="section-kicker">{t("admin.authLabel")}</p>
              <h1 className="font-display text-5xl leading-none text-foreground sm:text-6xl">{t("admin.authTitle")}</h1>
              <p className="max-w-2xl text-base text-muted-foreground sm:text-lg">{t("admin.authBody")}</p>
            </div>

            <div className="border-t border-border">
              {adminPoints.map((point, index) => (
                <div key={point.titleKey} className="grid gap-4 border-b border-border py-5 md:grid-cols-[96px_minmax(0,1fr)]">
                  <div className="font-display text-4xl leading-none text-primary">0{index + 1}</div>
                  <div className="space-y-2">
                    <h2 className="text-sm font-bold uppercase tracking-[0.06em] text-foreground">{t(point.titleKey)}</h2>
                    <p className="text-sm text-muted-foreground">{t(point.bodyKey)}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="flex justify-center lg:justify-end">{children}</div>
        </div>
      </div>
    </div>
  )
}
