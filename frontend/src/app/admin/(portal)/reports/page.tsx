"use client"

import * as React from "react"
import { BarChart3, CheckCheck, Clock3, Download, Layers3, Users } from "lucide-react"

import { AdminEmptyState, AdminMetricRow, AdminStatCard, AdminSurface } from "@/components/admin/admin-ui"
import { AdminPageHeader } from "@/components/admin/admin-page-header"
import { AdminStatusBadge } from "@/components/admin/admin-status-badge"
import { Button } from "@/components/ui/button"
import { CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { useAdminCandidates, useAdminDashboard, useAdminSelections, useAgencies } from "@/hooks/use-admin-portal"
import { downloadCsv, formatPercent, formatShortDate, titleize } from "@/lib/admin-utils"
import { UserRole } from "@/types"

function ProgressRow({
  label,
  value,
  total,
}: {
  label: string
  value: number
  total: number
}) {
  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between gap-3">
        <AdminStatusBadge status={label} />
        <span className="text-sm font-semibold text-slate-950 dark:text-slate-100">{value}</span>
      </div>
      <div className="h-2 rounded-full bg-slate-200 dark:bg-slate-800">
        <div
          className="h-2 rounded-full bg-gradient-to-r from-amber-400 via-sky-400 to-emerald-400"
          style={{ width: `${Math.max((value / Math.max(total, 1)) * 100, value > 0 ? 10 : 0)}%` }}
        />
      </div>
    </div>
  )
}

export default function AdminReportsPage() {
  const { data: stats } = useAdminDashboard()
  const { data: agencies = [] } = useAgencies({ status: "all", role: "all", search: "" })
  const { data: candidates = [] } = useAdminCandidates()
  const { data: selections = [] } = useAdminSelections()

  const candidateStatusBreakdown = React.useMemo(() => {
    const counts = new Map<string, number>()
    for (const candidate of candidates) {
      counts.set(candidate.status, (counts.get(candidate.status) ?? 0) + 1)
    }
    return Array.from(counts.entries()).sort((left, right) => right[1] - left[1])
  }, [candidates])

  const selectionApprovalBreakdown = React.useMemo(() => {
    const counts = new Map<string, number>()
    for (const selection of selections) {
      counts.set(selection.approval_status, (counts.get(selection.approval_status) ?? 0) + 1)
    }
    return Array.from(counts.entries()).sort((left, right) => right[1] - left[1])
  }, [selections])

  const latestAgencies = React.useMemo(
    () =>
      [...agencies]
        .sort((left, right) => new Date(right.registration_date).getTime() - new Date(left.registration_date).getTime())
        .slice(0, 6),
    [agencies]
  )

  const ethiopianAgencies = agencies.filter((agency) => agency.role === UserRole.ETHIOPIAN_AGENT)
  const foreignAgencies = agencies.filter((agency) => agency.role === UserRole.FOREIGN_AGENT)
  const activeAgencies = agencies.filter((agency) => agency.account_status === "active")

  const summaryCards = [
    {
      label: "Platform reach",
      value: stats?.total_agencies ?? agencies.length,
      detail: `${activeAgencies.length} active agencies live`,
      icon: Layers3,
    },
    {
      label: "Approval backlog",
      value: stats?.pending_approvals ?? 0,
      detail: "Registrations waiting for admin review",
      icon: Clock3,
    },
    {
      label: "Candidate supply",
      value: stats?.total_candidates ?? candidates.length,
      detail: `${candidateStatusBreakdown.length} tracked candidate states`,
      icon: Users,
    },
    {
      label: "Selection success",
      value: stats ? formatPercent(stats.success_rate) : "0.0%",
      detail: `${stats?.active_selections ?? 0} active selections in progress`,
      icon: CheckCheck,
    },
  ]

  const exportAgencies = () => {
    downloadCsv(
      "admin-agencies-report.csv",
      agencies.map((agency) => ({
        company_name: agency.company_name || agency.contact_person,
        contact_person: agency.contact_person,
        email: agency.email,
        role: titleize(agency.role),
        status: titleize(agency.account_status),
        registration_date: formatShortDate(agency.registration_date),
        total_candidates: agency.total_candidates,
        total_selections: agency.total_selections,
      }))
    )
  }

  const exportSelections = () => {
    downloadCsv(
      "admin-selections-report.csv",
      selections.map((selection) => ({
        candidate_name: selection.candidate_name,
        ethiopian_agency: selection.ethiopian_agency,
        foreign_agency: selection.foreign_agency,
        selection_status: titleize(selection.status),
        approval_status: titleize(selection.approval_status),
        selected_date: formatShortDate(selection.selected_date),
      }))
    )
  }

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="Operational Reports"
        description="A clean reporting surface for super admins to export current platform activity, monitor backlog, and review the latest operational shifts."
        action={
          <>
            <Button variant="outline" className="gap-2" onClick={exportAgencies}>
              <Download className="h-4 w-4" />
              Export agencies
            </Button>
            <Button className="gap-2" onClick={exportSelections}>
              <Download className="h-4 w-4" />
              Export selections
            </Button>
          </>
        }
      />

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        {summaryCards.map((card) => (
          <AdminStatCard key={card.label} label={card.label} value={card.value} detail={card.detail} icon={card.icon} />
        ))}
      </div>

      <div className="grid gap-6 xl:grid-cols-2">
        <AdminSurface>
          <CardHeader>
            <CardTitle className="text-lg text-slate-950 dark:text-slate-50">Candidate pipeline</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {candidateStatusBreakdown.length ? (
              candidateStatusBreakdown.map(([status, count]) => (
                <ProgressRow key={status} label={status} value={count} total={candidates.length} />
              ))
            ) : (
              <AdminEmptyState
                title="No candidate pipeline data yet"
                description="Candidate state reporting will appear here once agencies begin publishing profiles."
              />
            )}
          </CardContent>
        </AdminSurface>

        <AdminSurface>
          <CardHeader>
            <CardTitle className="text-lg text-slate-950 dark:text-slate-50">Selection approval states</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {selectionApprovalBreakdown.length ? (
              selectionApprovalBreakdown.map(([status, count]) => (
                <ProgressRow key={status} label={status} value={count} total={selections.length} />
              ))
            ) : (
              <AdminEmptyState
                title="No selection approvals yet"
                description="Approval-state reporting will appear here once agencies begin locking and confirming candidates."
              />
            )}
          </CardContent>
        </AdminSurface>
      </div>

      <div className="grid gap-6 xl:grid-cols-[1.15fr_0.85fr]">
        <AdminSurface>
          <CardHeader>
            <CardTitle className="text-lg text-slate-950 dark:text-slate-50">Latest agency registrations</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {latestAgencies.length ? (
              latestAgencies.map((agency) => (
                <div
                  key={agency.id}
                  className="flex flex-col gap-3 rounded-2xl border border-slate-200/80 bg-white/80 p-4 sm:flex-row sm:items-center sm:justify-between dark:border-slate-800 dark:bg-slate-900/85"
                >
                  <div className="space-y-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <p className="font-medium text-slate-950 dark:text-slate-100">{agency.company_name || agency.contact_person}</p>
                      <AdminStatusBadge status={agency.account_status} />
                    </div>
                    <p className="text-sm text-slate-500 dark:text-slate-400">{agency.contact_person} • {agency.email}</p>
                  </div>
                  <div className="text-left sm:text-right">
                    <p className="text-sm font-medium text-slate-950 dark:text-slate-100">{titleize(agency.role)}</p>
                    <p className="text-xs uppercase tracking-[0.18em] text-slate-400 dark:text-slate-500">
                      {formatShortDate(agency.registration_date)}
                    </p>
                  </div>
                </div>
              ))
            ) : (
              <AdminEmptyState
                title="No agency registrations yet"
                description="The latest registrations feed will appear here once agencies begin signing up."
              />
            )}
          </CardContent>
        </AdminSurface>

        <AdminSurface>
          <CardHeader>
            <CardTitle className="text-lg text-slate-950 dark:text-slate-50">Agency mix</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="rounded-2xl border border-slate-200/80 bg-white/80 p-4 dark:border-slate-800 dark:bg-slate-900/85">
              <div className="flex items-center gap-3">
                <div className="flex h-10 w-10 items-center justify-center rounded-2xl bg-amber-100 text-amber-700 dark:bg-amber-400/15 dark:text-amber-300">
                  <BarChart3 className="h-4 w-4" />
                </div>
                <div>
                  <p className="text-sm font-semibold text-slate-950 dark:text-slate-100">Operator spread</p>
                  <p className="text-sm text-slate-500 dark:text-slate-400">Balance between sourcing and hiring-side partners.</p>
                </div>
              </div>
            </div>

            <div className="grid gap-3">
              <AdminMetricRow label="Ethiopian agencies" value={ethiopianAgencies.length} />
              <AdminMetricRow label="Foreign agencies" value={foreignAgencies.length} />
              <AdminMetricRow
                label="Average candidates per Ethiopian agency"
                value={
                  ethiopianAgencies.length
                    ? (
                        ethiopianAgencies.reduce((sum, agency) => sum + agency.total_candidates, 0) /
                        ethiopianAgencies.length
                      ).toFixed(1)
                    : "0.0"
                }
              />
              <AdminMetricRow
                label="Average selections per Foreign agency"
                value={
                  foreignAgencies.length
                    ? (
                        foreignAgencies.reduce((sum, agency) => sum + agency.total_selections, 0) /
                        foreignAgencies.length
                      ).toFixed(1)
                    : "0.0"
                }
              />
            </div>
          </CardContent>
        </AdminSurface>
      </div>
    </div>
  )
}
