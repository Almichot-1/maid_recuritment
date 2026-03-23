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
import { cn } from "@/lib/utils"

function partnerName(companyName?: string, fullName?: string) {
  return companyName?.trim() || fullName?.trim() || "Partner agency"
}

export default function PartnersPage() {
  const { isEthiopianAgent } = useCurrentUser()
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

  const headerText = isEthiopianAgent
    ? "Switch between foreign partner agencies and see exactly which candidates from your library are visible inside each private workspace."
    : "Switch between Ethiopian partner agencies and review the candidates and selections that belong to each private workspace."

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
        <PageHeader
          heading="Partner Workspaces"
          text="This page comes alive once an admin connects your agency to at least one partner agency."
        />
        <Card className="border-dashed">
          <CardContent className="flex min-h-[260px] flex-col items-center justify-center gap-4 text-center">
            <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-muted text-muted-foreground">
              <Link2 className="h-6 w-6" />
            </div>
            <div className="space-y-2">
              <h2 className="text-xl font-semibold">No partner workspaces yet</h2>
              <p className="max-w-lg text-sm text-muted-foreground">
                As soon as an admin approves a partnership, this page will show the related agency and the candidates shared into that workspace.
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <PageHeader
        heading="Partner Workspaces"
        text={headerText}
        action={
          <Button variant="outline" asChild>
            <Link href="/candidates">
              <Users className="mr-2 h-4 w-4" />
              {isEthiopianAgent ? "Open Candidate Library" : "Browse Candidates"}
            </Link>
          </Button>
        }
      />

      <div className="grid gap-6 xl:grid-cols-[320px_minmax(0,1fr)]">
        <Card className="border-border/70 shadow-sm">
          <CardHeader>
            <CardTitle>Partner agencies</CardTitle>
            <CardDescription>
              Choose a partner workspace to inspect the candidates and workflow inside it.
            </CardDescription>
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
                    "w-full rounded-2xl border p-4 text-left transition-all duration-200",
                    isActive
                      ? "border-sky-400 bg-sky-50 shadow-sm dark:border-sky-700 dark:bg-sky-950/30"
                      : "border-border/70 bg-card hover:-translate-y-0.5 hover:border-sky-200 hover:bg-slate-50 dark:hover:bg-slate-900/50"
                  )}
                >
                  <div className="flex items-start gap-3">
                    <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-primary/10 text-primary">
                      <Building2 className="h-4 w-4" />
                    </div>
                    <div className="min-w-0 flex-1 space-y-2">
                      <div className="flex flex-wrap items-center gap-2">
                        <p className="truncate font-semibold text-foreground">
                          {partnerName(workspace.partner_agency.company_name, workspace.partner_agency.full_name)}
                        </p>
                        {isActive ? <Badge>Active view</Badge> : null}
                      </div>
                      <p className="truncate text-xs text-muted-foreground">{workspace.partner_agency.email}</p>
                      <Badge variant="outline" className="capitalize">
                        {workspace.status.replaceAll("_", " ")}
                      </Badge>
                    </div>
                  </div>
                </button>
              )
            })}
          </CardContent>
        </Card>

        <div className="space-y-6">
          <Card className="overflow-hidden border-0 bg-[radial-gradient(circle_at_top_left,_rgba(14,165,233,0.18),_transparent_28%),radial-gradient(circle_at_top_right,_rgba(16,185,129,0.18),_transparent_24%),linear-gradient(135deg,rgba(15,23,42,0.98),rgba(30,41,59,0.96))] text-white shadow-xl">
            <CardContent className="grid gap-6 p-6 lg:grid-cols-[minmax(0,1fr)_240px]">
              <div className="space-y-4">
                <Badge className="w-fit rounded-full border-0 bg-white/15 px-3 py-1 text-[11px] uppercase tracking-[0.24em] text-sky-200 hover:bg-white/15">
                  Active workspace
                </Badge>
                <div className="space-y-2">
                  <h2 className="text-3xl font-semibold tracking-tight">
                    {activeWorkspace
                      ? partnerName(activeWorkspace.partner_agency.company_name, activeWorkspace.partner_agency.full_name)
                      : "Partner workspace"}
                  </h2>
                  <p className="max-w-2xl text-sm text-slate-200/90">
                    {isEthiopianAgent
                      ? "These are the candidates from your agency library that are currently visible to this foreign partner."
                      : "These are the candidates this Ethiopian partner has made visible inside your current workspace."}
                  </p>
                </div>
                <div className="flex flex-wrap gap-3 text-xs text-slate-100/90">
                  <span className="rounded-full border border-white/15 bg-white/10 px-3 py-1.5">
                    {candidateData?.meta.total ?? 0} candidate{candidateData?.meta.total === 1 ? "" : "s"} in this workspace
                  </span>
                  <span className="rounded-full border border-white/15 bg-white/10 px-3 py-1.5">
                    {activeSelectionCount} selection{activeSelectionCount === 1 ? "" : "s"} in motion
                  </span>
                  <span className="rounded-full border border-white/15 bg-white/10 px-3 py-1.5">
                    {inTrackingCount} candidate{inTrackingCount === 1 ? "" : "s"} in tracking
                  </span>
                </div>
              </div>

              <div className="rounded-3xl border border-white/10 bg-white/10 p-5 backdrop-blur">
                <p className="text-xs uppercase tracking-[0.24em] text-sky-200">Helpful shortcut</p>
                <p className="mt-3 text-sm text-slate-100/90">
                  {isEthiopianAgent
                    ? "Your full agency library is still on the Candidates page. This workspace page only shows what the selected partner can see."
                    : "Switch partners here to compare which Ethiopian agency shared which candidates with you."}
                </p>
                <Button className="mt-4 w-full bg-white text-slate-950 hover:bg-slate-100" asChild>
                  <Link href="/candidates">
                    <ArrowRight className="mr-2 h-4 w-4" />
                    {isEthiopianAgent ? "Go to full library" : "Open candidate browser"}
                  </Link>
                </Button>
              </div>
            </CardContent>
          </Card>

          <div className="grid gap-4 md:grid-cols-3">
            <SummaryCard
              title="Visible candidates"
              value={isCandidatesLoading ? "..." : String(candidateData?.meta.total ?? 0)}
              description={isEthiopianAgent ? "Profiles this partner can review" : "Profiles shared into this workspace"}
            />
            <SummaryCard
              title="Selections"
              value={isSelectionsLoading ? "..." : String(activeSelectionCount)}
              description="Selections tied to this workspace"
            />
            <SummaryCard
              title="Tracking"
              value={isSelectionsLoading ? "..." : String(inTrackingCount)}
              description="Candidates already moving through recruitment"
            />
          </div>

          <Card className="border-border/70 shadow-sm">
            <CardHeader className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <CardTitle>Candidates in this workspace</CardTitle>
                <CardDescription>
                  {isEthiopianAgent
                    ? "Only the candidates you have shared with this selected foreign partner appear here."
                    : "Only the candidates shared by this selected Ethiopian partner appear here."}
                </CardDescription>
              </div>
              <Button variant="outline" asChild>
                <Link href="/selections">
                  <CheckSquare className="mr-2 h-4 w-4" />
                  Open selections
                </Link>
              </Button>
            </CardHeader>
            <CardContent>
              {isCandidatesLoading ? (
                <div className="space-y-3">
                  {Array.from({ length: 4 }).map((_, index) => (
                    <Skeleton key={index} className="h-14 w-full rounded-xl" />
                  ))}
                </div>
              ) : candidates.length > 0 ? (
                <CandidateTable candidates={candidates} />
              ) : (
                <div className="rounded-2xl border border-dashed border-border p-8 text-center">
                  <p className="text-sm font-semibold text-foreground">No candidates in this workspace yet</p>
                  <p className="mt-2 text-sm text-muted-foreground">
                    {isEthiopianAgent
                      ? "Open your full candidate library and share the right profiles with this partner agency."
                      : "This Ethiopian partner has not shared any currently available candidates into this workspace yet."}
                  </p>
                </div>
              )}
            </CardContent>
          </Card>

          <Card className="border-border/70 shadow-sm">
            <CardHeader>
              <CardTitle>Current workspace activity</CardTitle>
              <CardDescription>
                Selections and tracking items connected to the partner currently in view.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              {isSelectionsLoading ? (
                <div className="space-y-3">
                  {Array.from({ length: 3 }).map((_, index) => (
                    <Skeleton key={index} className="h-16 w-full rounded-xl" />
                  ))}
                </div>
              ) : selections.length > 0 ? (
                selections.slice(0, 5).map((selection) => (
                  <Link
                    key={selection.id}
                    href={`/selections/${selection.id}`}
                    className="flex items-center justify-between rounded-2xl border border-border/70 bg-muted/20 p-4 transition-colors hover:bg-muted/40"
                  >
                    <div className="space-y-1">
                      <p className="font-semibold text-foreground">{selection.candidate?.full_name || "Candidate"}</p>
                      <p className="text-xs text-muted-foreground">
                        {selection.status.replaceAll("_", " ")} - {selection.created_at ? new Date(selection.created_at).toLocaleDateString() : "Recently updated"}
                      </p>
                    </div>
                    <div className="flex items-center gap-2">
                      <Badge variant="outline" className="capitalize">
                        {selection.candidate?.status?.replaceAll("_", " ") || "selection"}
                      </Badge>
                      <Route className="h-4 w-4 text-muted-foreground" />
                    </div>
                  </Link>
                ))
              ) : (
                <div className="rounded-2xl border border-dashed border-border p-8 text-center">
                  <p className="text-sm font-semibold text-foreground">No activity in this workspace yet</p>
                  <p className="mt-2 text-sm text-muted-foreground">
                    Once a candidate is selected here, the approvals and process tracking will appear in this panel.
                  </p>
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
    <Card className="border-border/70 shadow-sm">
      <CardContent className="space-y-2 p-5">
        <p className="text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">{title}</p>
        <p className="text-3xl font-semibold tracking-tight text-foreground">{value}</p>
        <p className="text-sm text-muted-foreground">{description}</p>
      </CardContent>
    </Card>
  )
}
