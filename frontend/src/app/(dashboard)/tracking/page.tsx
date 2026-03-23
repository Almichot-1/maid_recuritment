"use client"

import * as React from "react"
import Link from "next/link"
import { ArrowRight, CheckCircle2, Home, Loader2, Route, Sparkles, TimerReset } from "lucide-react"

import { PageHeader } from "@/components/layout/page-header"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { useCurrentUser } from "@/hooks/use-auth"
import { usePairingContext } from "@/hooks/use-pairings"
import { useMySelections } from "@/hooks/use-selections"
import { CandidateStatus, SelectionStatus } from "@/types"

export default function TrackingHubPage() {
  const { isEthiopianAgent } = useCurrentUser()
  const { activeWorkspace } = usePairingContext()
  const { data: selections, isLoading } = useMySelections()

  const trackingSelections = React.useMemo(
    () => (selections || []).filter((selection) => selection.status === SelectionStatus.APPROVED && selection.candidate),
    [selections]
  )

  const activeRecruitments = trackingSelections.filter((selection) => selection.candidate?.status === CandidateStatus.IN_PROGRESS)
  const completedRecruitments = trackingSelections.filter((selection) => selection.candidate?.status === CandidateStatus.COMPLETED)

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500 pb-10">
      <nav className="flex items-center text-sm font-medium text-muted-foreground">
        <Link href="/dashboard" className="flex items-center transition-colors hover:text-primary">
          <Home className="mr-1.5 h-4 w-4" />
          Dashboard
        </Link>
        <span className="mx-2 text-muted-foreground/50">/</span>
        <span className="font-semibold text-foreground">Process Tracking</span>
      </nav>

      <Card className="overflow-hidden border-0 bg-[radial-gradient(circle_at_top_left,_rgba(56,189,248,0.18),_transparent_28%),radial-gradient(circle_at_top_right,_rgba(16,185,129,0.18),_transparent_22%),linear-gradient(135deg,rgba(255,255,255,0.88),rgba(244,250,249,0.96))] shadow-glow dark:bg-[radial-gradient(circle_at_top_left,_rgba(56,189,248,0.24),_transparent_28%),radial-gradient(circle_at_top_right,_rgba(16,185,129,0.22),_transparent_22%),linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,118,110,0.24))]">
        <CardContent className="grid gap-6 p-6 lg:grid-cols-[minmax(0,1fr)_320px]">
          <div className="space-y-4">
            <PageHeader
              className="pb-0"
              heading="Process Tracking"
              text={
                isEthiopianAgent
                  ? "Open any approved recruitment and update the shared progress timeline from medical through arrival."
                  : "Monitor every approved recruitment from one place so you always know what is happening with each candidate."
              }
            />
            {activeWorkspace ? (
              <Badge variant="outline" className="w-fit rounded-full px-3 py-1 text-[11px] uppercase tracking-[0.2em]">
                {activeWorkspace.partner_agency.company_name || activeWorkspace.partner_agency.full_name}
              </Badge>
            ) : null}
            <div className="flex flex-wrap gap-3">
              <Button asChild>
                <Link href="/selections">
                  <Sparkles className="mr-2 h-4 w-4" />
                  Open My Selections
                </Link>
              </Button>
              <Button variant="outline" asChild>
                <Link href="/candidates">
                  <Route className="mr-2 h-4 w-4" />
                  Browse Candidates
                </Link>
              </Button>
            </div>
          </div>

          <div className="grid gap-3">
            <SummaryCard label="Approved recruitments" value={trackingSelections.length} tone="from-sky-500/20 to-sky-400/5 text-sky-950 dark:text-sky-100" icon={<Sparkles className="h-5 w-5" />} />
            <SummaryCard label="In progress" value={activeRecruitments.length} tone="from-amber-500/20 to-amber-400/5 text-amber-950 dark:text-amber-100" icon={<TimerReset className="h-5 w-5" />} />
            <SummaryCard label="Completed" value={completedRecruitments.length} tone="from-emerald-500/20 to-emerald-400/5 text-emerald-950 dark:text-emerald-100" icon={<CheckCircle2 className="h-5 w-5" />} />
          </div>
        </CardContent>
      </Card>

      <Card className="overflow-hidden shadow-sm">
        <CardContent className="p-6">
          {isLoading ? (
            <div className="flex items-center justify-center py-16">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : trackingSelections.length > 0 ? (
            <div className="space-y-4">
              {trackingSelections.map((selection) => (
                <div
                  key={selection.id}
                  className="group flex flex-col gap-4 rounded-[1.6rem] border border-border/70 bg-card/95 p-5 shadow-soft transition-all duration-300 hover:-translate-y-1 hover:shadow-glow md:flex-row md:items-center md:justify-between"
                >
                  <div className="space-y-2">
                    <div className="flex flex-wrap items-center gap-2">
                      <h2 className="text-lg font-semibold">{selection.candidate?.full_name || "Candidate"}</h2>
                      <Badge className="bg-emerald-500 text-white hover:bg-emerald-500">Approved</Badge>
                      <Badge variant="outline" className="capitalize">
                        {selection.candidate?.status?.replaceAll("_", " ") || "tracking"}
                      </Badge>
                    </div>
                    <p className="text-sm text-muted-foreground">
                      {isEthiopianAgent
                        ? "Use the tracking page to update the recruitment milestones for this candidate."
                        : "Use the tracking page to follow the shared recruitment milestones for this candidate."}
                    </p>
                    <div className="inline-flex items-center gap-2 rounded-full bg-muted/60 px-3 py-1 text-xs font-medium text-muted-foreground">
                      <span className="h-2 w-2 rounded-full bg-primary shadow-[0_0_0_6px_rgba(20,184,166,0.14)]" />
                      Shared progress workspace is active
                    </div>
                  </div>

                  <div className="flex flex-wrap gap-3">
                    <Button variant="outline" asChild>
                      <Link href={`/selections/${selection.id}`}>
                        <Sparkles className="mr-2 h-4 w-4" />
                        Open Selection
                      </Link>
                    </Button>
                    <Button asChild>
                      <Link href={`/candidates/${selection.candidate_id}/tracking`}>
                        <Route className="mr-2 h-4 w-4" />
                        Open Process Tracking
                        <ArrowRight className="h-4 w-4 transition-transform duration-200 group-hover:translate-x-0.5" />
                      </Link>
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center gap-4 py-16 text-center">
              <div className="flex h-20 w-20 items-center justify-center rounded-full bg-muted">
                <CheckCircle2 className="h-10 w-10 text-muted-foreground" />
              </div>
              <div className="space-y-2">
                <h2 className="text-xl font-semibold">No approved recruitments yet</h2>
                <p className="max-w-md text-muted-foreground">
                  Once both the employer and the Ethiopian agency approve a candidate, that recruitment will appear here.
                </p>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

function SummaryCard({
  label,
  value,
  tone,
  icon,
}: {
  label: string
  value: number
  tone: string
  icon: React.ReactNode
}) {
  return (
    <Card className={`overflow-hidden border-white/20 bg-gradient-to-br ${tone} shadow-soft`}>
      <CardContent className="flex items-center justify-between p-5">
        <div>
          <p className="text-sm text-muted-foreground">{label}</p>
          <p className="mt-1 text-3xl font-semibold">{value}</p>
        </div>
        <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-white/70 text-current shadow-sm dark:bg-white/10">
          {icon}
        </div>
      </CardContent>
    </Card>
  )
}
