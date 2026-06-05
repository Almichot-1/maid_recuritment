"use client"

import * as React from "react"
import Link from "next/link"
import { ArrowRight, Search } from "lucide-react"

import { useCurrentUser } from "@/hooks/use-auth"
import { useDashboardHome } from "@/hooks/use-dashboard"
import { usePairingContext } from "@/hooks/use-pairings"
import { CandidateStatus } from "@/types"
import { CandidateCard } from "@/components/candidates/candidate-card"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import { useI18n } from "@/lib/i18n"

export function ForeignDashboard() {
  const { user } = useCurrentUser()
  const { activeWorkspace } = usePairingContext()
  const { t } = useI18n()
  const { data: home, isLoading } = useDashboardHome()

  if (!user) return null

  const stats = home?.stats
  const availableCandidates = (home?.available_candidates || []).filter((candidate) => candidate.status === CandidateStatus.AVAILABLE)
  const activeSelections = home?.active_selections || []
  const approvedSelections = home?.approved_selections || []

  return (
    <div className="space-y-6 animate-in">
      <Card>
        <CardContent className="grid gap-6 p-6 lg:grid-cols-[minmax(0,1fr)_300px]">
          <div className="space-y-4">
            <p className="section-kicker">{t("dashboard.foreignLabel")}</p>
            <h2 className="font-display text-5xl leading-none text-foreground">{t("dashboard.foreignTitle")}</h2>
            <p className="max-w-3xl text-sm text-muted-foreground sm:text-base">{t("dashboard.foreignBody")}</p>
            {activeWorkspace ? (
              <Badge variant="outline">{t("dashboard.activePartner", { name: activeWorkspace.partner_agency.company_name || activeWorkspace.partner_agency.full_name })}</Badge>
            ) : null}
            <div className="flex flex-wrap gap-3">
              <Button asChild>
                <Link href="/candidates">
                  <Search className="h-4 w-4" />
                  {t("dashboard.browsePool")}
                </Link>
              </Button>
              <Button variant="outline" asChild>
                <Link href="/selections">
                  <ArrowRight className="h-4 w-4" />
                  {t("dashboard.openSelections")}
                </Link>
              </Button>
            </div>
          </div>

          <div className="space-y-3 border-l border-border pl-0 lg:pl-6">
            <p className="section-kicker">{t("dashboard.quickPulse")}</p>
            <MiniMetric label={t("dashboard.availableNow")} value={stats?.availableCandidates ?? 0} />
            <MiniMetric label={t("dashboard.pendingApprovals")} value={stats?.activeSelections ?? 0} />
            <MiniMetric label={t("dashboard.approvedProfiles")} value={stats?.approved ?? 0} />
          </div>
        </CardContent>
      </Card>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <StatCard title={t("dashboard.availableNow")} value={stats?.availableCandidates} isLoading={isLoading} detail="Ready to review" />
        <StatCard title={t("dashboard.pendingApprovals")} value={stats?.activeSelections} isLoading={isLoading} detail="Waiting for approval" />
        <StatCard title={t("dashboard.approvedProfiles")} value={stats?.approved} isLoading={isLoading} detail="Both parties confirmed" />
        <StatCard title="In progress" value={stats?.inProgress} isLoading={isLoading} detail="Shared tracking visible" />
      </div>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1.45fr)_360px]">
        <div className="space-y-4">
          <div className="border-b border-border pb-4">
            <p className="section-kicker">{t("dashboard.availableNow")}</p>
            <h3 className="font-display text-4xl text-foreground">Select from live candidates</h3>
            <p className="text-sm text-muted-foreground">Choose directly from the shared candidate list without leaving the dashboard.</p>
          </div>

          {isLoading ? (
            <div className="grid gap-4 md:grid-cols-2">
              {Array.from({ length: 4 }).map((_, index) => (
                <div key={index} className="border border-border bg-card p-4">
                  <Skeleton className="aspect-square w-full" />
                  <Skeleton className="mt-4 h-6 w-2/3" />
                  <Skeleton className="mt-2 h-4 w-1/2" />
                  <Skeleton className="mt-4 h-10 w-full" />
                </div>
              ))}
            </div>
          ) : availableCandidates.length > 0 ? (
            <div className="grid gap-4 md:grid-cols-2">
              {availableCandidates.map((candidate) => (
                <CandidateCard key={candidate.id} candidate={candidate} />
              ))}
            </div>
          ) : (
            <Card className="border-dashed">
              <CardContent className="py-12 text-center text-sm text-muted-foreground">
                No available candidates are visible right now.
              </CardContent>
            </Card>
          )}
        </div>

        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>My live selections</CardTitle>
              <CardDescription>Open a selection to approve, reject, or keep tracking.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              {activeSelections.length > 0 ? (
                activeSelections.map((selection) => (
                  <div key={selection.id} className="border border-border bg-background p-4">
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <p className="text-sm font-bold uppercase tracking-[0.06em] text-foreground">{selection.candidate_name || "Candidate"}</p>
                        <p className="text-xs text-muted-foreground">
                          {selection.expires_at ? `Expires ${new Date(selection.expires_at).toLocaleString()}` : "Pending approval"}
                        </p>
                      </div>
                      <Badge variant="outline">Pending</Badge>
                    </div>
                    <Button className="mt-4 w-full" variant="outline" asChild>
                      <Link href={`/selections/${selection.id}`}>
                        Open selection
                        <ArrowRight className="h-4 w-4" />
                      </Link>
                    </Button>
                  </div>
                ))
              ) : (
                <p className="text-sm text-muted-foreground">No active selections yet.</p>
              )}
            </CardContent>
            <CardFooter>
              <Button variant="outline" className="w-full" asChild>
                <Link href="/selections">Manage all selections</Link>
              </Button>
            </CardFooter>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Recently approved</CardTitle>
              <CardDescription>Jump back into tracking for profiles already approved by both parties.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              {approvedSelections.length > 0 ? (
                approvedSelections.map((selection) => (
                  <Link key={selection.id} href={`/selections/${selection.id}`} className="flex items-center justify-between border border-border bg-background px-4 py-3 transition-colors hover:bg-muted/20">
                    <div>
                      <p className="text-sm font-bold uppercase tracking-[0.06em] text-foreground">{selection.candidate_name || "Candidate"}</p>
                      <p className="text-xs text-muted-foreground">Shared recruitment tracking is active</p>
                    </div>
                    <ArrowRight className="h-4 w-4 text-muted-foreground" />
                  </Link>
                ))
              ) : (
                <p className="text-sm text-muted-foreground">Approved selections will appear here once both parties confirm.</p>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}

function MiniMetric({ label, value }: { label: string; value: number }) {
  return (
    <div className="border border-border bg-background p-4">
      <p className="route-stamp text-[11px] text-muted-foreground">{label}</p>
      <p className="mt-2 font-display text-4xl text-foreground">{value}</p>
    </div>
  )
}

function StatCard({
  title,
  value,
  isLoading,
  detail,
}: {
  title: string
  value?: number
  isLoading: boolean
  detail: string
}) {
  return (
    <Card>
      <CardContent className="space-y-2 p-5">
        <p className="route-stamp text-[11px] text-muted-foreground">{title}</p>
        {isLoading ? <Skeleton className="h-10 w-20" /> : <p className="font-display text-5xl text-foreground">{value ?? 0}</p>}
        <p className="text-sm text-muted-foreground">{detail}</p>
      </CardContent>
    </Card>
  )
}
