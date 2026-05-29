"use client"

import * as React from "react"
import Link from "next/link"
import { ArrowRight, Building2, CheckSquare, Link2, Loader2, Route, Users } from "lucide-react"

import { useCurrentUser } from "@/hooks/use-auth"
import { useCandidates } from "@/hooks/use-candidates"
import { usePairingContext } from "@/hooks/use-pairings"
import { useMySelections } from "@/hooks/use-selections"
import { CandidateTable } from "@/components/candidates/candidate-table"
import { PageHeader } from "@/components/layout/page-header"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useI18n } from "@/lib/i18n"
import { cn } from "@/lib/utils"

function partnerName(companyName?: string, fullName?: string) {
  return companyName?.trim() || fullName?.trim() || "Partner agency"
}

export default function PartnersPage() {
  const { isEthiopianAgent } = useCurrentUser()
  const { t } = useI18n()
  const {
    context,
    activePairingId,
    activeWorkspace,
    hasActivePairs,
    isLoading: isPairingLoading,
    setActivePairingId,
  } = usePairingContext()
  const { data: selections = [], isLoading: isSelectionsLoading } = useMySelections()
  const { data: candidateData, isLoading: isCandidatesLoading } = useCandidates({
    page: 1,
    page_size: 24,
    ...(isEthiopianAgent ? { shared_only: true } : {}),
  })

  const candidates = candidateData?.data || []
  const activeSelectionCount = selections.filter((selection) => selection.status === "pending" || selection.status === "approved").length
  const inTrackingCount = selections.filter((selection) => selection.candidate?.status === "in_progress" || selection.candidate?.status === "completed").length

  const headerText = isEthiopianAgent ? t("partners.bodyEthiopian") : t("partners.bodyForeign")

  if (isPairingLoading) {
    return (
      <div className="flex h-[50vh] w-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (!hasActivePairs || !context?.workspaces.length) {
    return (
      <div className="space-y-6">
        <PageHeader heading={t("partners.heading")} text={t("partners.noneBody")} />
        <Card className="border-dashed">
          <CardContent className="flex min-h-[260px] flex-col items-center justify-center gap-4 text-center">
            <div className="flex h-14 w-14 items-center justify-center border border-border bg-muted/20 text-muted-foreground">
              <Link2 className="h-6 w-6" />
            </div>
            <div className="space-y-2">
              <h2 className="font-display text-3xl text-foreground">{t("partners.noneTitle")}</h2>
              <p className="max-w-lg text-sm text-muted-foreground">{t("partners.noneBody")}</p>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="space-y-6 animate-in">
      <PageHeader
        heading={t("partners.heading")}
        text={headerText}
        action={
          <Button variant="outline" asChild>
            <Link href="/candidates">
              <Users className="h-4 w-4" />
              {isEthiopianAgent ? t("partners.openLibrary") : t("partners.browseCandidates")}
            </Link>
          </Button>
        }
      />

      <div className="grid gap-6 xl:grid-cols-[320px_minmax(0,1fr)]">
        <Card>
          <CardHeader>
            <CardTitle>{t("partners.listTitle")}</CardTitle>
            <CardDescription>{t("partners.listBody")}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {context.workspaces.map((workspace) => {
              const isActive = workspace.id === activePairingId
              return (
                <button
                  key={workspace.id}
                  type="button"
                  onClick={() => setActivePairingId(workspace.id)}
                  className={cn(
                    "w-full border p-4 text-left transition-colors",
                    isActive ? "border-primary bg-primary/5" : "border-border bg-background hover:bg-muted/20"
                  )}
                >
                  <div className="flex items-start gap-3">
                    <div className="flex h-10 w-10 shrink-0 items-center justify-center border border-border text-primary">
                      <Building2 className="h-4 w-4" />
                    </div>
                    <div className="min-w-0 flex-1 space-y-2">
                      <div className="flex flex-wrap items-center gap-2">
                        <p className="truncate text-sm font-bold uppercase tracking-[0.06em] text-foreground">
                          {partnerName(workspace.partner_agency.company_name, workspace.partner_agency.full_name)}
                        </p>
                        {isActive ? <Badge variant="outline">{t("common.activeView")}</Badge> : null}
                      </div>
                      <p className="truncate text-xs text-muted-foreground">{workspace.partner_agency.email}</p>
                      <Badge variant="outline">{workspace.status.replaceAll("_", " ")}</Badge>
                    </div>
                  </div>
                </button>
              )
            })}
          </CardContent>
        </Card>

        <div className="space-y-6">
          <Card>
            <CardContent className="grid gap-6 p-6 lg:grid-cols-[minmax(0,1fr)_240px]">
              <div className="space-y-4">
                <p className="section-kicker">{t("partners.heroLabel")}</p>
                <div className="space-y-2">
                  <h2 className="font-display text-4xl text-foreground">
                    {activeWorkspace
                      ? partnerName(activeWorkspace.partner_agency.company_name, activeWorkspace.partner_agency.full_name)
                      : t("partners.heroDefaultTitle")}
                  </h2>
                  <p className="max-w-2xl text-sm text-muted-foreground sm:text-base">
                    {isEthiopianAgent ? t("partners.heroEthiopianBody") : t("partners.heroForeignBody")}
                  </p>
                </div>
                <div className="flex flex-wrap gap-3 text-xs text-muted-foreground">
                  <span className="border border-border bg-background px-3 py-2">
                    {candidateData?.meta.total ?? 0} {t("partners.visibleCandidates")}
                  </span>
                  <span className="border border-border bg-background px-3 py-2">
                    {activeSelectionCount} {t("partners.selections")}
                  </span>
                  <span className="border border-border bg-background px-3 py-2">
                    {inTrackingCount} {t("partners.tracking")}
                  </span>
                </div>
              </div>

              <div className="space-y-3 border-l border-border pl-0 lg:pl-6">
                <p className="section-kicker">{t("partners.shortcutTitle")}</p>
                <p className="text-sm text-muted-foreground">
                  {isEthiopianAgent ? t("partners.shortcutEthiopian") : t("partners.shortcutForeign")}
                </p>
                <Button className="w-full" asChild>
                  <Link href="/candidates">
                    <ArrowRight className="h-4 w-4" />
                    {isEthiopianAgent ? t("partners.goLibrary") : t("partners.goBrowser")}
                  </Link>
                </Button>
              </div>
            </CardContent>
          </Card>

          <div className="grid gap-4 md:grid-cols-3">
            <SummaryCard title={t("partners.visibleCandidates")} value={isCandidatesLoading ? "..." : String(candidateData?.meta.total ?? 0)} description={isEthiopianAgent ? t("partners.candidatesBodyEthiopian") : t("partners.candidatesBodyForeign")} />
            <SummaryCard title={t("partners.selections")} value={isSelectionsLoading ? "..." : String(activeSelectionCount)} description={t("partners.activityBody")} />
            <SummaryCard title={t("partners.tracking")} value={isSelectionsLoading ? "..." : String(inTrackingCount)} description={t("partners.activityBody")} />
          </div>

          <Card>
            <CardHeader className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <CardTitle>{t("partners.candidatesTitle")}</CardTitle>
                <CardDescription>
                  {isEthiopianAgent ? t("partners.candidatesBodyEthiopian") : t("partners.candidatesBodyForeign")}
                </CardDescription>
              </div>
              <Button variant="outline" asChild>
                <Link href="/selections">
                  <CheckSquare className="h-4 w-4" />
                  {t("partners.openSelections")}
                </Link>
              </Button>
            </CardHeader>
            <CardContent>
              {isCandidatesLoading ? (
                <div className="space-y-3">
                  {Array.from({ length: 4 }).map((_, index) => (
                    <Skeleton key={index} className="h-14 w-full" />
                  ))}
                </div>
              ) : candidates.length > 0 ? (
                <CandidateTable candidates={candidates} />
              ) : (
                <div className="border border-dashed border-border p-8 text-center">
                  <p className="text-sm font-bold uppercase tracking-[0.06em] text-foreground">{t("partners.emptyCandidatesTitle")}</p>
                  <p className="mt-2 text-sm text-muted-foreground">
                    {isEthiopianAgent ? t("partners.emptyCandidatesBodyEthiopian") : t("partners.emptyCandidatesBodyForeign")}
                  </p>
                </div>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>{t("partners.activityTitle")}</CardTitle>
              <CardDescription>{t("partners.activityBody")}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              {isSelectionsLoading ? (
                <div className="space-y-3">
                  {Array.from({ length: 3 }).map((_, index) => (
                    <Skeleton key={index} className="h-16 w-full" />
                  ))}
                </div>
              ) : selections.length > 0 ? (
                selections.slice(0, 5).map((selection) => (
                  <Link
                    key={selection.id}
                    href={`/selections/${selection.id}`}
                    className="flex items-center justify-between border border-border bg-background p-4 transition-colors hover:bg-muted/20"
                  >
                    <div className="space-y-1">
                      <p className="text-sm font-bold uppercase tracking-[0.06em] text-foreground">{selection.candidate?.full_name || "Candidate"}</p>
                      <p className="text-xs text-muted-foreground">
                        {selection.status.replaceAll("_", " ")} - {selection.created_at ? new Date(selection.created_at).toLocaleDateString() : "Recently updated"}
                      </p>
                    </div>
                    <div className="flex items-center gap-2">
                      <Badge variant="outline">{selection.candidate?.status?.replaceAll("_", " ") || "selection"}</Badge>
                      <Route className="h-4 w-4 text-muted-foreground" />
                    </div>
                  </Link>
                ))
              ) : (
                <div className="border border-dashed border-border p-8 text-center">
                  <p className="text-sm font-bold uppercase tracking-[0.06em] text-foreground">{t("partners.emptyActivityTitle")}</p>
                  <p className="mt-2 text-sm text-muted-foreground">{t("partners.emptyActivityBody")}</p>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}

function SummaryCard({ title, value, description }: { title: string; value: string; description: string }) {
  return (
    <Card>
      <CardContent className="space-y-2 p-5">
        <p className="route-stamp text-[11px] text-muted-foreground">{title}</p>
        <p className="font-display text-4xl text-foreground">{value}</p>
        <p className="text-sm text-muted-foreground">{description}</p>
      </CardContent>
    </Card>
  )
}
