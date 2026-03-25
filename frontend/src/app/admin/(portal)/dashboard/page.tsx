"use client"

import * as React from "react"
import Link from "next/link"
import { ArrowRight, CheckCheck, Clock3, Layers3, Users } from "lucide-react"

import { AdminEmptyState, AdminStatCard, AdminSurface } from "@/components/admin/admin-ui"
import { AdminPageHeader } from "@/components/admin/admin-page-header"
import { AdminStatusBadge } from "@/components/admin/admin-status-badge"
import { Button } from "@/components/ui/button"
import { CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  useAdminCandidates,
  useAdminDashboard,
  useAdminSelections,
  useAgencies,
  usePendingAgencies,
} from "@/hooks/use-admin-portal"
import { buildDailySeries, formatPercent, formatRelative, formatShortDate, titleize } from "@/lib/admin-utils"
import { UserRole } from "@/types"

function MiniBars({ points }: { points: Array<{ label: string; value: number }> }) {
  const max = Math.max(...points.map((point) => point.value), 1)

  return (
    <div className="space-y-3">
      <div className="flex h-28 items-end gap-2">
        {points.map((point) => (
          <div key={point.label} className="flex flex-1 flex-col items-center gap-2">
            <div
              className="w-full rounded-t-2xl bg-gradient-to-t from-amber-500 via-cyan-400 to-emerald-300"
              style={{ height: `${Math.max((point.value / max) * 100, point.value > 0 ? 12 : 4)}%` }}
            />
            <span className="text-[11px] text-slate-500 dark:text-slate-400">{point.label}</span>
          </div>
        ))}
      </div>
    </div>
  )
}

function HeroMetric({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-2xl border border-slate-200/80 bg-white/82 p-4 dark:border-white/10 dark:bg-white/5">
      <p className="text-xs uppercase tracking-[0.2em] text-slate-500 dark:text-slate-300">{label}</p>
      <p className="mt-2 text-2xl font-semibold text-slate-950 dark:text-white">{value}</p>
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
            <Button className="gap-2">
              Review queue
              <ArrowRight className="h-4 w-4" />
            </Button>
          </Link>
        }
      />

      <AdminSurface className="overflow-hidden">
        <CardContent className="grid gap-6 p-6 lg:grid-cols-[minmax(0,1fr)_320px]">
          <div className="space-y-3">
            <div className="text-xs font-semibold uppercase tracking-[0.24em] text-amber-700 dark:text-amber-300">Operator pulse</div>
            <h2 className="text-3xl font-semibold tracking-tight text-slate-950 dark:text-slate-50">A calmer view of the whole platform</h2>
            <p className="max-w-3xl text-sm text-slate-600 dark:text-slate-300">
              Approvals, candidate supply, selection throughput, and recruitment momentum are visible here so platform operators can intervene quickly when something drifts.
            </p>
          </div>

          <div className="grid gap-3 sm:grid-cols-3 lg:grid-cols-1">
            <HeroMetric label="Pending queue" value={pendingAgencies.length} />
            <HeroMetric label="Candidate states" value={candidateDistribution.length} />
            <HeroMetric label="Selection stream" value={selections.length} />
          </div>
        </CardContent>
      </AdminSurface>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        {metricCards.map((card) => (
          <AdminStatCard
            key={card.label}
            label={card.label}
            value={statsLoading ? "..." : card.value}
            detail={card.detail}
            icon={card.icon}
            className="transition-transform duration-200 hover:-translate-y-1"
          />
        ))}
      </div>

      <div className="grid gap-6 xl:grid-cols-[1.35fr_0.95fr]">
        <AdminSurface>
          <CardHeader className="space-y-2">
            <CardTitle className="text-lg text-slate-950 dark:text-slate-50">Pending approval queue</CardTitle>
            <p className="text-sm text-slate-500 dark:text-slate-400">Newest agency applications that still need an admin decision.</p>
          </CardHeader>
          <CardContent className="space-y-3">
            {pendingAgencies.slice(0, 5).map((agency) => (
              <div key={agency.id} className="flex flex-col gap-3 rounded-2xl border border-slate-200/80 bg-white/80 p-4 sm:flex-row sm:items-center sm:justify-between dark:border-slate-800 dark:bg-slate-900/85">
                <div className="space-y-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <p className="font-medium text-slate-950 dark:text-slate-100">{agency.company_name || agency.contact_person}</p>
                    <AdminStatusBadge status={agency.account_status} />
                  </div>
                  <p className="text-sm text-slate-500 dark:text-slate-400">{agency.contact_person} • {agency.email}</p>
                  <p className="text-xs uppercase tracking-wide text-slate-400 dark:text-slate-500">
                    {titleize(agency.role)} • registered {formatRelative(agency.registration_date)}
                  </p>
                </div>
                <Link href={`/admin/agencies/${agency.id}`}>
                  <Button variant="outline">Review</Button>
                </Link>
              </div>
            ))}
            {!pendingAgencies.length ? (
              <AdminEmptyState
                title="Queue is clear"
                description="No pending approvals right now. Newly registered agencies will appear here automatically."
              />
            ) : null}
          </CardContent>
        </AdminSurface>

        <AdminSurface>
          <CardHeader>
            <CardTitle className="text-lg text-slate-950 dark:text-slate-50">Candidate status distribution</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {candidateDistribution.map(([status, count]) => (
              <div key={status} className="space-y-2">
                <div className="flex items-center justify-between text-sm">
                  <AdminStatusBadge status={status} />
                  <span className="font-semibold text-slate-950 dark:text-slate-100">{count}</span>
                </div>
                <div className="h-2 rounded-full bg-slate-200 dark:bg-slate-800">
                  <div
                    className="h-2 rounded-full bg-gradient-to-r from-amber-400 via-sky-400 to-emerald-400"
                    style={{ width: `${Math.max((count / Math.max(candidates.length, 1)) * 100, 8)}%` }}
                  />
                </div>
              </div>
            ))}
            {!candidateDistribution.length ? (
              <AdminEmptyState
                title="No candidate data yet"
                description="Candidate status tracking will appear here once agencies begin publishing profiles."
              />
            ) : null}
          </CardContent>
        </AdminSurface>
      </div>

      <div className="grid gap-6 xl:grid-cols-2">
        <AdminSurface>
          <CardHeader>
            <CardTitle className="text-lg text-slate-950 dark:text-slate-50">Agency registrations over the last 7 days</CardTitle>
          </CardHeader>
          <CardContent>
            <MiniBars points={registrationTrend} />
          </CardContent>
        </AdminSurface>

        <AdminSurface>
          <CardHeader>
            <CardTitle className="text-lg text-slate-950 dark:text-slate-50">Selections over the last 7 days</CardTitle>
          </CardHeader>
          <CardContent>
            <MiniBars points={selectionTrend} />
          </CardContent>
        </AdminSurface>
      </div>

      <div className="grid gap-6 xl:grid-cols-2">
        <AdminSurface>
          <CardHeader>
            <CardTitle className="text-lg text-slate-950 dark:text-slate-50">Top Ethiopian agencies</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {topEthiopian.map((agency, index) => (
              <div key={agency.id} className="flex items-center justify-between rounded-2xl border border-slate-200/80 bg-white/80 p-4 dark:border-slate-800 dark:bg-slate-900/85">
                <div>
                  <p className="font-medium text-slate-950 dark:text-slate-100">
                    {index + 1}. {agency.company_name || agency.contact_person}
                  </p>
                  <p className="text-sm text-slate-500 dark:text-slate-400">Registered {formatShortDate(agency.registration_date)}</p>
                </div>
                <div className="text-right">
                  <p className="text-lg font-semibold text-slate-950 dark:text-slate-100">{agency.total_candidates}</p>
                  <p className="text-xs uppercase tracking-wide text-slate-400 dark:text-slate-500">Candidates</p>
                </div>
              </div>
            ))}
            {!topEthiopian.length ? (
              <AdminEmptyState
                title="No Ethiopian agency rankings yet"
                description="Agency performance will populate here once candidate supply starts moving."
              />
            ) : null}
          </CardContent>
        </AdminSurface>

        <AdminSurface>
          <CardHeader>
            <CardTitle className="text-lg text-slate-950 dark:text-slate-50">Top Foreign agencies</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {topForeign.map((agency, index) => (
              <div key={agency.id} className="flex items-center justify-between rounded-2xl border border-slate-200/80 bg-white/80 p-4 dark:border-slate-800 dark:bg-slate-900/85">
                <div>
                  <p className="font-medium text-slate-950 dark:text-slate-100">
                    {index + 1}. {agency.company_name || agency.contact_person}
                  </p>
                  <p className="text-sm text-slate-500 dark:text-slate-400">Registered {formatShortDate(agency.registration_date)}</p>
                </div>
                <div className="text-right">
                  <p className="text-lg font-semibold text-slate-950 dark:text-slate-100">{agency.total_selections}</p>
                  <p className="text-xs uppercase tracking-wide text-slate-400 dark:text-slate-500">Selections</p>
                </div>
              </div>
            ))}
            {!topForeign.length ? (
              <AdminEmptyState
                title="No Foreign agency rankings yet"
                description="Selection activity will populate here once employer-side demand grows."
              />
            ) : null}
          </CardContent>
        </AdminSurface>
      </div>
    </div>
  )
}
