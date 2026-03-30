"use client"

import * as React from "react"
import Link from "next/link"
import { ArrowRight, CheckCheck, Clock3, Layers3, Users } from "lucide-react"

import { AdminPageHeader } from "@/components/admin/admin-page-header"
import { AdminStatusBadge } from "@/components/admin/admin-status-badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  useAdminCandidates,
  useAdminDashboard,
  useAdminSelections,
  useAgencies,
  usePendingAgencies,
} from "@/hooks/use-admin-portal"
import { UserRole } from "@/types"
import { buildDailySeries, formatPercent, formatRelative, formatShortDate, titleize } from "@/lib/admin-utils"

function MiniBars({ points }: { points: Array<{ label: string; value: number }> }) {
  const max = Math.max(...points.map((point) => point.value), 1)

  return (
    <div className="space-y-3">
      <div className="flex h-28 items-end gap-2">
        {points.map((point) => (
          <div key={point.label} className="flex flex-1 flex-col items-center gap-2">
            <div
              className="w-full rounded-t-2xl bg-gradient-to-t from-slate-950 via-slate-800 to-amber-400"
              style={{ height: `${Math.max((point.value / max) * 100, point.value > 0 ? 12 : 4)}%` }}
            />
            <span className="text-[11px] text-slate-500">{point.label}</span>
          </div>
        ))}
      </div>
    </div>
  )
}

export default function AdminDashboardPage() {
  const { data: stats, isLoading: statsLoading } = useAdminDashboard()
  const { data: pendingAgencies = [] } = usePendingAgencies("all")
  const { data: agencies = [] } = useAgencies({ status: "all", role: "all", search: "" })
  const { data: candidates = [] } = useAdminCandidates()
  const { data: selections = [] } = useAdminSelections()

  const registrationTrend = React.useMemo(
    () => buildDailySeries(agencies.map((agency) => agency.registration_date), 7),
    [agencies]
  )
  const selectionTrend = React.useMemo(
    () => buildDailySeries(selections.map((selection) => selection.selected_date), 7),
    [selections]
  )

  const topEthiopian = React.useMemo(
    () =>
      agencies
        .filter((agency) => agency.role === UserRole.ETHIOPIAN_AGENT)
        .sort((left, right) => right.total_candidates - left.total_candidates)
        .slice(0, 5),
    [agencies]
  )

  const topForeign = React.useMemo(
    () =>
      agencies
        .filter((agency) => agency.role === UserRole.FOREIGN_AGENT)
        .sort((left, right) => right.total_selections - left.total_selections)
        .slice(0, 5),
    [agencies]
  )

  const candidateDistribution = React.useMemo(() => {
    const counts = new Map<string, number>()
    for (const candidate of candidates) {
      counts.set(candidate.status, (counts.get(candidate.status) ?? 0) + 1)
    }
    return Array.from(counts.entries()).sort((left, right) => right[1] - left[1])
  }, [candidates])

  const metricCards = [
    {
      label: "Total Agencies",
      value: stats?.total_agencies ?? 0,
      detail: `${stats?.ethiopian_agencies ?? 0} Ethiopian / ${stats?.foreign_agencies ?? 0} Foreign`,
      icon: Layers3,
    },
    {
      label: "Pending Approvals",
      value: stats?.pending_approvals ?? 0,
      detail: "Registrations waiting for admin decision",
      icon: Clock3,
    },
    {
      label: "Total Candidates",
      value: stats?.total_candidates ?? 0,
      detail: "Visible across all Ethiopian agencies",
      icon: Users,
    },
    {
      label: "Success Rate",
      value: stats ? formatPercent(stats.success_rate) : "0.0%",
      detail: `${stats?.active_selections ?? 0} active selections in flight`,
      icon: CheckCheck,
    },
  ]

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="Platform Overview"
        description="A live control center for approvals, agency activity, candidate supply, and recruitment progress across the platform."
        action={
          <Link href="/admin/agencies/pending">
            <Button className="gap-2 bg-amber-400 text-slate-950 hover:bg-amber-300">
              Review queue
              <ArrowRight className="h-4 w-4" />
            </Button>
          </Link>
        }
      />

      <Card className="overflow-hidden border-0 bg-[radial-gradient(circle_at_top_left,_rgba(251,191,36,0.22),_transparent_28%),radial-gradient(circle_at_bottom_right,_rgba(15,23,42,0.12),_transparent_30%),linear-gradient(135deg,rgba(15,23,42,0.98),rgba(30,41,59,0.98))] text-white shadow-xl">
        <CardContent className="grid gap-6 p-6 lg:grid-cols-[minmax(0,1fr)_320px]">
          <div className="space-y-3">
            <div className="text-xs font-semibold uppercase tracking-[0.24em] text-amber-200">Operator pulse</div>
            <h2 className="text-3xl font-semibold tracking-tight">The platform, above every agency workflow</h2>
            <p className="max-w-3xl text-sm text-slate-200/90">
              Approvals, candidate supply, selection throughput, and recruitment momentum are visible here so platform operators can intervene quickly when something drifts.
            </p>
          </div>

          <div className="grid gap-3 sm:grid-cols-3 lg:grid-cols-1">
            <HeroMetric label="Pending queue" value={pendingAgencies.length} />
            <HeroMetric label="Candidate states" value={candidateDistribution.length} />
            <HeroMetric label="Selection stream" value={selections.length} />
          </div>
        </CardContent>
      </Card>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        {metricCards.map((card) => (
          <Card key={card.label} className="border-slate-200 bg-white/90 shadow-sm transition-transform duration-200 hover:-translate-y-1 hover:shadow-lg">
            <CardContent className="flex items-start justify-between p-6">
              <div className="space-y-2">
                <p className="text-sm font-medium text-slate-500">{card.label}</p>
                <p className="text-3xl font-semibold tracking-tight text-slate-950">
                  {statsLoading ? "..." : card.value}
                </p>
                <p className="text-sm text-slate-500">{card.detail}</p>
              </div>
              <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-amber-100 text-amber-700">
                <card.icon className="h-5 w-5" />
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="grid gap-6 xl:grid-cols-[1.35fr_0.95fr]">
        <Card className="border-slate-200 bg-white/90 shadow-sm">
          <CardHeader className="space-y-2">
            <CardTitle className="text-lg text-slate-950">Pending approval queue</CardTitle>
            <p className="text-sm text-slate-500">Newest agency applications that still need an admin decision.</p>
          </CardHeader>
          <CardContent className="space-y-3">
            {pendingAgencies.slice(0, 5).map((agency) => (
              <div key={agency.id} className="flex flex-col gap-3 rounded-2xl border border-slate-200 bg-white p-4 sm:flex-row sm:items-center sm:justify-between">
                <div className="space-y-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <p className="font-medium text-slate-950">{agency.company_name || agency.contact_person}</p>
                    <AdminStatusBadge status={agency.account_status} />
                  </div>
                  <p className="text-sm text-slate-500">{agency.contact_person} • {agency.email}</p>
                  <p className="text-xs uppercase tracking-wide text-slate-400">
                    {titleize(agency.role)} • registered {formatRelative(agency.registration_date)}
                  </p>
                </div>
                <Link href={`/admin/agencies/${agency.id}`}>
                  <Button variant="outline" className="border-slate-700 bg-slate-900 text-slate-100 hover:bg-slate-800">Review</Button>
                </Link>
              </div>
            ))}
            {!pendingAgencies.length ? (
              <div className="rounded-2xl border border-dashed border-slate-300 p-6 text-sm text-slate-500">
                No pending approvals right now. Newly registered agencies will appear here automatically.
              </div>
            ) : null}
          </CardContent>
        </Card>

        <Card className="border-slate-200 bg-white/90 shadow-sm">
          <CardHeader>
            <CardTitle className="text-lg text-slate-950">Candidate status distribution</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {candidateDistribution.map(([status, count]) => (
              <div key={status} className="space-y-2">
                <div className="flex items-center justify-between text-sm">
                  <AdminStatusBadge status={status} />
                  <span className="font-semibold text-slate-950">{count}</span>
                </div>
                <div className="h-2 rounded-full bg-slate-100">
                  <div
                    className="h-2 rounded-full bg-slate-950"
                    style={{ width: `${Math.max((count / Math.max(candidates.length, 1)) * 100, 8)}%` }}
                  />
                </div>
              </div>
            ))}
            {!candidateDistribution.length ? <p className="text-sm text-slate-500">No candidate data available yet.</p> : null}
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 xl:grid-cols-2">
        <Card className="border-slate-200 bg-white/90 shadow-sm">
          <CardHeader>
            <CardTitle className="text-lg text-slate-950">Agency registrations over the last 7 days</CardTitle>
          </CardHeader>
          <CardContent>
            <MiniBars points={registrationTrend} />
          </CardContent>
        </Card>
        <Card className="border-slate-200 bg-white/90 shadow-sm">
          <CardHeader>
            <CardTitle className="text-lg text-slate-950">Selections over the last 7 days</CardTitle>
          </CardHeader>
          <CardContent>
            <MiniBars points={selectionTrend} />
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 xl:grid-cols-2">
        <Card className="border-slate-200 bg-white/90 shadow-sm">
          <CardHeader>
            <CardTitle className="text-lg text-slate-950">Top Ethiopian agencies</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {topEthiopian.map((agency, index) => (
              <div key={agency.id} className="flex items-center justify-between rounded-2xl border border-slate-200 p-4">
                <div>
                  <p className="font-medium text-slate-950">
                    {index + 1}. {agency.company_name || agency.contact_person}
                  </p>
                  <p className="text-sm text-slate-500">Registered {formatShortDate(agency.registration_date)}</p>
                </div>
                <div className="text-right">
                  <p className="text-lg font-semibold text-slate-950">{agency.total_candidates}</p>
                  <p className="text-xs uppercase tracking-wide text-slate-400">Candidates</p>
                </div>
              </div>
            ))}
          </CardContent>
        </Card>

        <Card className="border-slate-200 bg-white/90 shadow-sm">
          <CardHeader>
            <CardTitle className="text-lg text-slate-950">Top Foreign agencies</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {topForeign.map((agency, index) => (
              <div key={agency.id} className="flex items-center justify-between rounded-2xl border border-slate-200 p-4">
                <div>
                  <p className="font-medium text-slate-950">
                    {index + 1}. {agency.company_name || agency.contact_person}
                  </p>
                  <p className="text-sm text-slate-500">Registered {formatShortDate(agency.registration_date)}</p>
                </div>
                <div className="text-right">
                  <p className="text-lg font-semibold text-slate-950">{agency.total_selections}</p>
                  <p className="text-xs uppercase tracking-wide text-slate-400">Selections</p>
                </div>
              </div>
            ))}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

function HeroMetric({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-2xl border border-white/10 bg-white/10 p-4">
      <p className="text-xs uppercase tracking-[0.2em] text-slate-300">{label}</p>
      <p className="mt-2 text-2xl font-semibold text-white">{value}</p>
    </div>
  )
}
