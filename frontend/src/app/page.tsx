"use client"

import Link from "next/link"
import { ArrowRight, CheckCircle2 } from "lucide-react"

import { Logo } from "@/components/shared/logo"
import { LocaleSwitcher } from "@/components/shared/locale-switcher"
import { Button } from "@/components/ui/button"
import { useI18n } from "@/lib/i18n"

const laneRows = [
  { code: "01", titleKey: "landing.feature1Title", bodyKey: "landing.feature1Body" },
  { code: "02", titleKey: "landing.feature2Title", bodyKey: "landing.feature2Body" },
  { code: "03", titleKey: "landing.feature3Title", bodyKey: "landing.feature3Body" },
  { code: "04", titleKey: "landing.feature4Title", bodyKey: "landing.feature4Body" },
]

export default function LandingPage() {
  const { t } = useI18n()

  return (
    <div className="min-h-screen bg-background text-foreground">
      <header className="sticky top-0 z-50 border-b border-border/80 bg-background/95">
        <div className="mx-auto flex max-w-7xl items-center justify-between gap-4 px-4 py-3 sm:px-6 sm:py-4 lg:px-8">
          <Logo size="sm" />
          <nav className="flex items-center gap-2 sm:gap-3" aria-label={t("landing.footerAuth")}>
            <div className="hidden sm:block">
              <LocaleSwitcher compact />
            </div>
            <Button
              variant="outline"
              size="lg"
              className="hidden min-w-[6.5rem] border-border bg-background font-semibold normal-case tracking-normal hover:border-primary hover:bg-primary/5 sm:inline-flex sm:min-w-[7.5rem]"
              asChild
            >
              <Link href="/register">{t("landing.navRegister")}</Link>
            </Button>
            <Button
              size="xl"
              className="min-w-[9.5rem] font-semibold normal-case tracking-normal shadow-sm sm:min-w-[11.5rem]"
              asChild
            >
              <Link href="/login">{t("landing.navLogin")}</Link>
            </Button>
          </nav>
        </div>
      </header>

      <main>
        <section className="border-b border-border">
          <div className="mx-auto grid max-w-7xl gap-12 px-4 py-16 sm:px-6 lg:grid-cols-[minmax(0,1.2fr)_420px] lg:px-8 lg:py-24">
            <div className="space-y-8">
              <div className="space-y-5">
                <p className="section-kicker">{t("landing.heroKicker")}</p>
                <h1 className="headline-display max-w-4xl">{t("landing.heroTitle")}</h1>
                <p className="max-w-2xl text-base text-muted-foreground sm:text-lg">{t("landing.heroBody")}</p>
              </div>

              <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center">
                <Button
                  size="lg"
                  className="h-12 w-full font-semibold normal-case tracking-normal shadow-sm sm:h-14 sm:w-auto sm:min-w-[12rem] sm:px-10 sm:text-lg"
                  asChild
                >
                  <Link href="/login">{t("landing.navLogin")}</Link>
                </Button>
                <Button
                  variant="outline"
                  size="lg"
                  className="h-12 w-full border-border bg-background font-semibold normal-case tracking-normal hover:border-primary hover:bg-primary/5 sm:h-14 sm:w-auto sm:min-w-[12rem] sm:px-10 sm:text-lg"
                  asChild
                >
                  <Link href="/register">
                    {t("landing.heroCta")}
                    <ArrowRight className="h-5 w-5" aria-hidden />
                  </Link>
                </Button>
              </div>
            </div>

            <aside className="relative border border-border bg-card p-6 sm:p-8" aria-labelledby="landing-workflow-heading">
              <div className="space-y-6">
                <div className="border-b border-border pb-4">
                  <p id="landing-workflow-heading" className="section-kicker">
                    {t("landing.heroAsideTitle")}
                  </p>
                  <p className="mt-3 max-w-sm text-sm text-muted-foreground">{t("landing.heroAsideBody")}</p>
                </div>
                <div className="space-y-3">
                  {laneRows.map((row) => (
                    <div key={row.code} className="grid grid-cols-[56px_minmax(0,1fr)] gap-4 border border-border bg-background px-4 py-4">
                      <div className="font-display text-3xl leading-none text-primary">{row.code}</div>
                      <div className="space-y-1">
                        <p className="text-sm font-bold uppercase tracking-[0.06em] text-foreground">{t(row.titleKey)}</p>
                        <p className="text-sm text-muted-foreground">{t(row.bodyKey)}</p>
                      </div>
                    </div>
                  ))}
                </div>
                <div className="border-t border-border pt-4">
                  <div className="grid gap-3 sm:grid-cols-2">
                    <div className="border border-border bg-background p-4">
                      <p className="text-xs font-bold uppercase tracking-wider text-foreground">EN</p>
                      <p className="mt-2 text-sm text-muted-foreground">{t("landing.heroLocaleEn")}</p>
                    </div>
                    <div className="border border-border bg-background p-4">
                      <p className="text-xs font-bold uppercase tracking-wider text-foreground">AR / AM</p>
                      <p className="mt-2 text-sm text-muted-foreground">{t("landing.heroLocaleMultilingual")}</p>
                    </div>
                  </div>
                </div>
              </div>
            </aside>
          </div>
        </section>

        <section className="mx-auto max-w-7xl px-4 py-16 sm:px-6 lg:px-8">
          <div className="max-w-3xl space-y-3">
            <p className="section-kicker">{t("landing.featuresTitle")}</p>
            <p className="text-base text-muted-foreground sm:text-lg">{t("landing.featuresBody")}</p>
          </div>

          <div className="mt-10 border-t border-border">
            {laneRows.map((row) => (
              <div key={row.code} className="grid gap-5 border-b border-border py-6 md:grid-cols-[120px_minmax(0,1fr)_220px]">
                <div className="font-display text-5xl leading-none text-primary">{row.code}</div>
                <div className="space-y-2">
                  <h2 className="font-display text-3xl text-foreground">{t(row.titleKey)}</h2>
                  <p className="max-w-2xl text-sm text-muted-foreground sm:text-base">{t(row.bodyKey)}</p>
                </div>
                <div className="flex items-start justify-start md:justify-end">
                  <span className="inline-flex items-center gap-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">
                    <CheckCircle2 className="h-4 w-4 text-primary" aria-hidden />
                    {t("landing.featureLane")}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </section>
      </main>

      <footer className="border-t border-border bg-card">
        <div className="mx-auto flex max-w-7xl flex-col gap-8 px-4 py-10 sm:px-6 lg:flex-row lg:justify-between lg:px-8">
          <div className="space-y-3">
            <Logo size="sm" />
            <p className="max-w-sm text-sm text-muted-foreground">{t("landing.footerTagline")}</p>
          </div>

          <div className="grid gap-8 sm:grid-cols-3">
            <div className="space-y-2">
              <p className="text-xs font-bold uppercase tracking-wider text-muted-foreground">{t("landing.footerAuth")}</p>
              <div className="flex flex-col gap-2 text-sm">
                <Link href="/login" className="hover:text-primary">{t("landing.footerAgency")}</Link>
                <Link href="/register" className="hover:text-primary">{t("landing.footerRegister")}</Link>
              </div>
            </div>
            <div className="space-y-2">
              <p className="text-xs font-bold uppercase tracking-wider text-muted-foreground">{t("landing.footerAdminLabel")}</p>
              <div className="flex flex-col gap-2 text-sm">
                <Link href="/admin/login" className="hover:text-primary">{t("landing.footerAdmin")}</Link>
              </div>
            </div>
            <div className="space-y-2">
              <p className="text-xs font-bold uppercase tracking-wider text-muted-foreground">{t("landing.footerContact")}</p>
              <div className="flex flex-col gap-2 text-sm">
                <Link href="mailto:contact@recruitmatch.com" className="hover:text-primary">contact@recruitmatch.com</Link>
              </div>
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}
