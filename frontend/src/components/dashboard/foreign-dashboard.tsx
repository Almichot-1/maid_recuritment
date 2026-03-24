"use client"

import * as React from "react"
import Link from "next/link"
import { ArrowRight, CheckCircle, Clock, Search, Sparkles, Star, Users } from "lucide-react"

import { useCurrentUser } from "@/hooks/use-auth"
import { useDashboardHome } from "@/hooks/use-dashboard"
import { usePairingContext } from "@/hooks/use-pairings"
import { CandidateStatus } from "@/types"
import { CandidateCard } from "@/components/candidates/candidate-card"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/utils"

export function ForeignDashboard() {
  const { user } = useCurrentUser()
  const { activeWorkspace } = usePairingContext()
  const { data: home, isLoading } = useDashboardHome()

  if (!user) return null

  const stats = home?.stats
  const availableCandidates = (home?.available_candidates || []).filter((candidate) => candidate.status === CandidateStatus.AVAILABLE)
  const activeSelections = home?.active_selections || []
  const approvedSelections = home?.approved_selections || []

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <Card className="overflow-hidden border-0 bg-[radial-gradient(circle_at_top_left,_rgba(14,165,233,0.18),_transparent_30%),radial-gradient(circle_at_top_right,_rgba(251,191,36,0.24),_transparent_25%),linear-gradient(135deg,rgba(15,23,42,0.98),rgba(30,41,59,0.98))] text-white shadow-xl">
        <CardContent className="grid gap-6 p-6 lg:grid-cols-[minmax(0,1fr)_280px]">
          <div className="space-y-4">
            <Badge className="w-fit rounded-full border-0 bg-white/15 px-3 py-1 text-[11px] uppercase tracking-[0.24em] text-amber-200 hover:bg-white/15">
              Employer workspace
            </Badge>
            <div className="space-y-2">
              <h2 className="text-3xl font-semibold tracking-tight">
                Select available maid candidates straight from your dashboard
              </h2>
              <p className="max-w-2xl text-sm text-slate-200/90">
                Review the newest candidates shared by your currently selected Ethiopian partner, make a selection immediately, and then track approvals and recruitment progress from the same workspace.
              </p>
            </div>
            {activeWorkspace ? (
              <Badge className="w-fit rounded-full border-0 bg-white/10 px-3 py-1 text-[11px] uppercase tracking-[0.24em] text-slate-100 hover:bg-white/10">
                Active partner: {activeWorkspace.partner_agency.company_name || activeWorkspace.partner_agency.full_name}
              </Badge>
            ) : null}
            <div className="flex flex-wrap gap-3">
              <Button className="bg-white text-slate-950 hover:bg-slate-100" asChild>
                <Link href="/candidates">
                  <Search className="mr-2 h-4 w-4" />
                  Browse full candidate pool
                </Link>
              </Button>
              <Button variant="outline" className="border-white/20 bg-white/10 text-white hover:bg-white/15 hover:text-white" asChild>
                <Link href="/selections">
                  <Sparkles className="mr-2 h-4 w-4" />
                  Open my selections
                </Link>
              </Button>
            </div>
          </div>

          <div className="rounded-3xl border border-white/10 bg-white/10 p-5 backdrop-blur">
            <p className="text-xs uppercase tracking-[0.24em] text-amber-200">Quick pulse</p>
            <div className="mt-4 space-y-4">
              <MiniMetric label="Available now" value={stats?.availableCandidates ?? 0} />
              <MiniMetric label="Pending approvals" value={stats?.activeSelections ?? 0} />
              <MiniMetric label="Approved profiles" value={stats?.approved ?? 0} />
            </div>
          </div>
        </CardContent>
      </Card>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <StatCard title="Available Candidates" icon={<Users className="h-5 w-5 text-sky-600" />} bg="bg-sky-100" value={stats?.availableCandidates} isLoading={isLoading} trend="Ready to review" />
        <StatCard title="Active Selections" icon={<Star className="h-5 w-5 text-amber-600" />} bg="bg-amber-100" value={stats?.activeSelections} isLoading={isLoading} trend="Awaiting approvals" />
        <StatCard title="Approved Profiles" icon={<CheckCircle className="h-5 w-5 text-emerald-600" />} bg="bg-emerald-100" value={stats?.approved} isLoading={isLoading} trend="In recruitment workflow" />
        <StatCard title="In Progress" icon={<Clock className="h-5 w-5 text-indigo-600" />} bg="bg-indigo-100" value={stats?.inProgress} isLoading={isLoading} trend="Shared tracking visible" />
      </div>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1.45fr)_360px]">
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-xl font-semibold tracking-tight">Select from available candidates</h3>
              <p className="text-sm text-muted-foreground">These cards are live. You can select directly here without leaving the dashboard.</p>
            </div>
            <Link href="/candidates" className="text-sm font-medium text-primary hover:underline underline-offset-4">
              View all available
            </Link>
          </div>

          {isLoading ? (
            <div className="grid gap-4 md:grid-cols-2">
              {Array.from({ length: 4 }).map((_, index) => (
                <div key={index} className="rounded-2xl border bg-card p-4 shadow-sm">
                  <Skeleton className="aspect-square w-full rounded-xl" />
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
                No available candidates are visible right now. Check back after the Ethiopian agency publishes more profiles.
              </CardContent>
            </Card>
          )}
        </div>

        <div className="space-y-6">
          <Card className="shadow-sm">
            <CardHeader className="border-b border-border/70 pb-4">
              <CardTitle className="text-lg">My live selections</CardTitle>
              <CardDescription>Open a selection to approve, reject, or continue tracking.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3 pt-4">
              {activeSelections.length > 0 ? (
                activeSelections.map((selection) => (
                  <div key={selection.id} className="rounded-2xl border border-border/70 bg-muted/20 p-4">
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <p className="font-semibold">{selection.candidate_name || "Candidate"}</p>
                        <p className="text-xs text-muted-foreground">
                          {selection.expires_at ? `Expires ${new Date(selection.expires_at).toLocaleString()}` : "Pending approval"}
                        </p>
                      </div>
                      <Badge className="bg-amber-500 text-white hover:bg-amber-500">Pending</Badge>
                    </div>
                    <Button className="mt-4 w-full" variant="outline" asChild>
                      <Link href={`/selections/${selection.id}`}>
                        Open selection
                        <ArrowRight className="ml-2 h-4 w-4" />
                      </Link>
                    </Button>
                  </div>
                ))
              ) : (
                <p className="text-sm text-muted-foreground">No active selections yet.</p>
              )}
            </CardContent>
            <CardFooter>
              <Button variant="secondary" className="w-full" asChild>
                <Link href="/selections">Manage all selections</Link>
              </Button>
            </CardFooter>
          </Card>

          <Card className="shadow-sm">
            <CardHeader className="pb-3">
              <CardTitle className="text-lg">Recently approved</CardTitle>
              <CardDescription>Jump back into tracking for profiles already approved by both parties.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              {approvedSelections.length > 0 ? (
                approvedSelections.map((selection) => (
                      <Link
                    key={selection.id}
                    href={`/selections/${selection.id}`}
                    className="flex items-center justify-between rounded-2xl border border-border/70 px-4 py-3 transition-colors hover:bg-muted/30"
                  >
                    <div>
                      <p className="font-semibold">{selection.candidate_name || "Candidate"}</p>
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
    <div className="rounded-2xl border border-white/10 bg-white/5 p-4">
      <p className="text-xs uppercase tracking-[0.2em] text-slate-300">{label}</p>
      <p className="mt-2 text-2xl font-semibold text-white">{value}</p>
    </div>
  )
}

function StatCard({
  title,
  icon,
  bg,
  value,
  isLoading,
  trend,
}: {
  title: string
  icon: React.ReactNode
  bg: string
  value?: number
  isLoading: boolean
  trend: string
}) {
  return (
    <Card className="hover:shadow-md transition-all duration-300 shadow-sm">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-semibold text-muted-foreground">{title}</CardTitle>
        <div className={cn("flex h-9 w-9 items-center justify-center rounded-full", bg)}>{icon}</div>
      </CardHeader>
      <CardContent>
        {isLoading ? <Skeleton className="mb-1 h-8 w-16" /> : <div className="text-3xl font-bold">{value ?? 0}</div>}
        <p className="mt-1 text-xs text-muted-foreground">{trend}</p>
      </CardContent>
    </Card>
  )
}
