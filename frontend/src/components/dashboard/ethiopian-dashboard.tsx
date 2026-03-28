"use client"

import * as React from "react"
import Link from "next/link"
import { formatDistanceToNowStrict } from "date-fns"
import { Users, CheckCircle, Lock, Loader, CircleAlert, ArrowRight, UserPlus, FileCheck, CheckSquare, Sparkles, AlertTriangle, ShieldAlert, PlaneTakeoff, PlaneLanding, HeartPulse } from "lucide-react"

import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { useDashboardHome, useSmartAlerts } from "@/hooks/use-dashboard"
import { useCurrentUser } from "@/hooks/use-auth"
import { usePairingContext } from "@/hooks/use-pairings"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/utils"

export function EthiopianDashboard() {
  const { user } = useCurrentUser()
  const { activeWorkspace } = usePairingContext()
  const { data: home, isLoading } = useDashboardHome()
  const { data: smartAlerts, isLoading: isSmartAlertsLoading } = useSmartAlerts()

  if (!user) return null

  const stats = home?.stats
  const recentCandidates = home?.recent_candidates || []
  const incompleteProfiles = home?.pending_actions?.incompleteProfiles ?? 0

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <Card className="overflow-hidden border-0 bg-[radial-gradient(circle_at_top_left,_rgba(16,185,129,0.16),_transparent_28%),radial-gradient(circle_at_top_right,_rgba(245,158,11,0.24),_transparent_24%),linear-gradient(135deg,rgba(15,23,42,0.98),rgba(30,41,59,0.98))] text-white shadow-xl">
        <CardContent className="grid gap-6 p-6 lg:grid-cols-[minmax(0,1fr)_280px]">
          <div className="space-y-4">
            <Badge className="w-fit rounded-full border-0 bg-white/15 px-3 py-1 text-[11px] uppercase tracking-[0.24em] text-emerald-200 hover:bg-white/15">
              Agency control room
            </Badge>
            <div className="space-y-2">
              <h2 className="text-3xl font-semibold tracking-tight">
                Manage candidates, approvals, and recruitment progress from one page
              </h2>
              <p className="max-w-2xl text-sm text-slate-200/90">
          Build polished candidate profiles in your master library, share them into the right private partner workspace, and keep every approved recruitment step visible from medical through arrival.
              </p>
            </div>
            {activeWorkspace ? (
              <Badge className="w-fit rounded-full border-0 bg-white/10 px-3 py-1 text-[11px] uppercase tracking-[0.24em] text-slate-100 hover:bg-white/10">
                Active partner: {activeWorkspace.partner_agency.company_name || activeWorkspace.partner_agency.full_name}
              </Badge>
            ) : null}
            <div className="flex flex-wrap gap-3">
              <Button className="bg-white text-slate-950 hover:bg-slate-100" asChild>
                <Link href="/candidates/new">
                  <UserPlus className="mr-2 h-4 w-4" />
                  Add a new candidate
                </Link>
              </Button>
              <Button variant="outline" className="border-white/20 bg-white/10 text-white hover:bg-white/15 hover:text-white" asChild>
                <Link href="/selections">
                  <CheckSquare className="mr-2 h-4 w-4" />
                  Review selections
                </Link>
              </Button>
            </div>
          </div>

          <div className="rounded-3xl border border-white/10 bg-white/10 p-5 backdrop-blur">
            <p className="text-xs uppercase tracking-[0.24em] text-emerald-200">Today at a glance</p>
            <div className="mt-4 space-y-4">
              <MiniMetric label="Shared in this workspace" value={stats?.totalCandidates ?? 0} />
              <MiniMetric label="Profiles in tracking" value={stats?.inProgress ?? 0} />
              <MiniMetric label="Approval queue" value={stats?.activeSelections ?? 0} />
            </div>
          </div>
        </CardContent>
      </Card>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard title="Total Candidates" icon={<Users className="h-5 w-5 text-blue-600 dark:text-blue-400" />} bg="bg-blue-100 dark:bg-blue-900/30" value={stats?.totalCandidates} isLoading={isLoading} trend={{ value: "+12%", label: "from last month", positive: true }} />
        <StatCard title="Available Candidates" icon={<CheckCircle className="h-5 w-5 text-green-600 dark:text-green-400" />} bg="bg-green-100 dark:bg-green-900/30" value={stats?.availableCandidates} isLoading={isLoading} trend={{ value: "+4", label: "new this week", positive: true }} />
        <StatCard title="Selected Candidates" icon={<Lock className="h-5 w-5 text-purple-600 dark:text-purple-400" />} bg="bg-purple-100 dark:bg-purple-900/30" value={stats?.selectedCandidates} isLoading={isLoading} trend={{ value: "-2%", label: "from last month", positive: false }} />
        <StatCard title="In Progress" icon={<Loader className="h-5 w-5 text-amber-600 dark:text-amber-400" />} bg="bg-amber-100 dark:bg-amber-900/30" value={stats?.inProgress} isLoading={isLoading} trend={{ value: "5", label: "awaiting docs", positive: false }} />
      </div>

      <div className="grid gap-6 md:grid-cols-7">
        <Card className="md:col-span-4 lg:col-span-5 flex flex-col shadow-sm">
          <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
            <div>
              <CardTitle>Recent Activity</CardTitle>
              <CardDescription>Latest updates to candidate profiles.</CardDescription>
            </div>
            <Button variant="outline" size="sm" asChild>
              <Link href="/candidates">
                View All <ArrowRight className="ml-2 h-4 w-4" />
              </Link>
            </Button>
          </CardHeader>
          <CardContent className="flex-1">
            <div className="rounded-md border mt-2">
              <Table>
                <TableHeader className="bg-muted/50">
                  <TableRow>
                    <TableHead>Candidate Name</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="text-right">Created Date</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {isLoading ? (
                    <TableRow>
                      <TableCell colSpan={3} className="py-6 text-center text-muted-foreground">
                        Loading recent candidates...
                      </TableCell>
                    </TableRow>
                  ) : recentCandidates.length > 0 ? recentCandidates.map((candidate) => (
                    <TableRow key={candidate.id}>
                      <TableCell className="font-medium">{candidate.full_name}</TableCell>
                      <TableCell>
                        <Badge variant={
                          candidate.status === "available" ? "default" :
                          candidate.status === "locked" ? "secondary" :
                          candidate.status === "approved" ? "outline" : "destructive"
                        } className={candidate.status === 'available' ? "bg-green-100 text-green-800 hover:bg-green-100 dark:bg-green-900 dark:text-green-300 pointer-events-none" : "pointer-events-none"}>
                          {candidate.status.replace("_", " ")}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-right text-muted-foreground">{new Date(candidate.created_at).toLocaleDateString()}</TableCell>
                    </TableRow>
                  )) : (
                    <TableRow>
                      <TableCell colSpan={3} className="py-6 text-center text-muted-foreground">
                        No recent candidates yet.
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>

        <div className="space-y-6 md:col-span-3 lg:col-span-2">
          <Card className="shadow-sm">
            <CardHeader className="pb-3">
              <CardTitle>Pending Actions</CardTitle>
              <CardDescription>Tasks that need your attention</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <PendingActionLink
                href="/selections"
                icon={<FileCheck className="h-5 w-5 text-amber-600 dark:text-amber-500" />}
                iconWrapperClassName="bg-amber-100 dark:bg-amber-900/40"
                title={`${stats?.activeSelections ?? 0} Selections`}
                description="Awaiting your approval"
              />
              <PendingActionLink
                href="/candidates"
                icon={<CircleAlert className="h-5 w-5 text-destructive dark:text-red-400" />}
                iconWrapperClassName="bg-destructive/10 dark:bg-destructive/20"
                title={`${incompleteProfiles} Profiles`}
                description="Missing required documents"
              />
            </CardContent>
          </Card>

          <Card className="shadow-sm">
            <CardHeader className="pb-3">
              <CardTitle>Quick Actions</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <Button className="w-full justify-start" asChild>
                <Link href="/candidates/new">
                  <UserPlus className="mr-2 h-4 w-4" />
                  Add New Candidate
                </Link>
              </Button>
              <Button variant="secondary" className="w-full justify-start" asChild>
                <Link href="/selections">
                  <CheckSquare className="mr-2 h-4 w-4" />
                  View All Selections
                </Link>
              </Button>
              <Button variant="outline" className="w-full justify-start" asChild>
                <Link href="/candidates">
                  <Sparkles className="mr-2 h-4 w-4" />
                  Open candidate workspace
                </Link>
              </Button>
            </CardContent>
          </Card>
        </div>
      </div>

      <Card className="shadow-sm">
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between gap-3">
            <div>
              <CardTitle>Smart Alerts</CardTitle>
              <CardDescription>
                Watch expiring selections, passport deadlines, and flight-stage candidates in the active workspace.
              </CardDescription>
            </div>
            {activeWorkspace ? (
              <Badge variant="outline" className="rounded-full px-3 py-1">
                {activeWorkspace.partner_agency.company_name || activeWorkspace.partner_agency.full_name}
              </Badge>
            ) : null}
          </div>
        </CardHeader>
        <CardContent className="grid gap-4 xl:grid-cols-4">
          <SmartAlertColumn
            title="Passport Expiry"
            icon={<ShieldAlert className="h-4 w-4 text-amber-600" />}
            emptyText="No passports are near expiry in this workspace."
            loading={isSmartAlertsLoading}
            items={(smartAlerts?.expiring_passports ?? []).map((passport) => ({
              id: passport.candidate_id,
              href: `/candidates/${passport.candidate_id}`,
              title: passport.candidate_name,
              meta: `Passport ${passport.passport_number || "not stored"}`,
              detail: `Expires ${formatRelativeDate(passport.expiry_date)}`,
              tone: passport.warning_level,
            }))}
          />
          <SmartAlertColumn
            title="Medical Expiry"
            icon={<HeartPulse className="h-4 w-4 text-rose-600" />}
            emptyText="No medical files are near expiry in this workspace."
            loading={isSmartAlertsLoading}
            items={(smartAlerts?.expiring_medicals ?? []).map((medical) => ({
              id: `${medical.candidate_id}-medical`,
              href: `/candidates/${medical.candidate_id}/tracking`,
              title: medical.candidate_name,
              meta: "Medical document",
              detail: `Expires ${formatRelativeDate(medical.expiry_date)}`,
              tone: medical.warning_level,
            }))}
          />
          <SmartAlertColumn
            title="Selection Expiry"
            icon={<AlertTriangle className="h-4 w-4 text-rose-600" />}
            emptyText="No selections are nearing expiry right now."
            loading={isSmartAlertsLoading}
            items={(smartAlerts?.expiring_selections ?? []).map((selection) => ({
              id: selection.selection_id,
              href: `/selections/${selection.selection_id}`,
              title: selection.candidate_name,
              meta: selection.remaining_label,
              detail: `Lock expires ${formatRelativeDate(selection.expires_at)}`,
              tone: selection.warning_level,
            }))}
          />
          <SmartAlertColumn
            title="Flight Updates"
            icon={<PlaneTakeoff className="h-4 w-4 text-sky-600" />}
            emptyText="No candidates are in the flight stages yet."
            loading={isSmartAlertsLoading}
            items={[
              ...(smartAlerts?.flight_updates ?? []).map((item) => ({
                id: `${item.candidate_id}-${item.stage}`,
                href: `/candidates/${item.candidate_id}/tracking`,
                title: item.candidate_name,
                meta: item.stage,
                detail: `Updated ${formatRelativeDate(item.updated_at)}`,
                tone: "medium",
              })),
              ...(smartAlerts?.recently_arrived ?? []).map((item) => ({
                id: `${item.candidate_id}-${item.stage}-arrived`,
                href: `/candidates/${item.candidate_id}/tracking`,
                title: item.candidate_name,
                meta: "Arrived",
                detail: `Arrived ${formatRelativeDate(item.updated_at)}`,
                tone: "low",
                arrived: true,
              })),
            ]}
          />
        </CardContent>
      </Card>
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

function StatCard({ title, icon, bg, value, isLoading, trend }: { title: string, icon: React.ReactNode, bg: string, value?: number, isLoading: boolean, trend: { value: string, label: string, positive: boolean } }) {
  return (
    <Card className="hover:shadow-md transition-shadow">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          {title}
        </CardTitle>
        <div className={cn("flex h-9 w-9 items-center justify-center rounded-full", bg)}>
          {icon}
        </div>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <Skeleton className="h-8 w-16 mb-1" />
        ) : (
          <div className="text-3xl font-bold">{value ?? 0}</div>
        )}
        <p className="text-xs text-muted-foreground mt-1 flex items-center gap-1">
          <span className={cn("font-medium", trend.positive ? "text-green-600 dark:text-green-400" : "text-amber-600 dark:text-amber-400")}>
            {trend.value}
          </span>
          <span className="truncate">{" "}{trend.label}</span>
        </p>
      </CardContent>
    </Card>
  )
}

function PendingActionLink({
  href,
  icon,
  iconWrapperClassName,
  title,
  description,
}: {
  href: string
  icon: React.ReactNode
  iconWrapperClassName: string
  title: string
  description: string
}) {
  return (
    <Link
      href={href}
      className="group flex items-center gap-3 rounded-lg border p-3 text-sm transition-all duration-200 hover:-translate-y-0.5 hover:bg-slate-50 hover:shadow-sm dark:hover:bg-slate-800/50"
    >
      <div className={cn("flex h-9 w-9 shrink-0 items-center justify-center rounded-full", iconWrapperClassName)}>
        {icon}
      </div>
      <div className="flex flex-1 flex-col">
        <span className="font-semibold text-foreground">{title}</span>
        <span className="text-xs text-muted-foreground">{description}</span>
      </div>
      <ArrowRight className="h-4 w-4 text-muted-foreground transition-transform duration-200 group-hover:translate-x-0.5" />
    </Link>
  )
}

function SmartAlertColumn({
  title,
  icon,
  loading,
  emptyText,
  items,
}: {
  title: string
  icon: React.ReactNode
  loading: boolean
  emptyText: string
  items: Array<{
    id: string
    href: string
    title: string
    meta: string
    detail: string
    tone: string
    arrived?: boolean
  }>
}) {
  return (
    <div className="rounded-3xl border border-border/70 bg-muted/15 p-4">
      <div className="flex items-center gap-2">
        {icon}
        <p className="text-sm font-semibold">{title}</p>
      </div>

      <div className="mt-4 space-y-3">
        {loading ? (
          <>
            <Skeleton className="h-20 w-full rounded-2xl" />
            <Skeleton className="h-20 w-full rounded-2xl" />
          </>
        ) : items.length > 0 ? (
          items.map((item) => (
            <Link
              key={item.id}
              href={item.href}
              className="group block rounded-2xl border border-border/60 bg-background/85 p-4 transition-all duration-200 hover:-translate-y-0.5 hover:border-primary/35 hover:shadow-sm"
            >
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <p className="truncate text-sm font-semibold text-foreground">{item.title}</p>
                  <p className="mt-1 text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">
                    {item.meta}
                  </p>
                </div>
                <Badge variant="outline" className={cn("shrink-0", alertTone(item.tone), item.arrived ? "border-emerald-300 bg-emerald-50 text-emerald-700" : "")}>
                  {item.arrived ? <PlaneLanding className="mr-1 h-3.5 w-3.5" /> : null}
                  {item.arrived ? "Arrived" : item.tone}
                </Badge>
              </div>
              <p className="mt-3 text-sm text-muted-foreground">{item.detail}</p>
              <div className="mt-3 flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.18em] text-primary">
                Open workspace item
                <ArrowRight className="h-3.5 w-3.5 transition-transform duration-200 group-hover:translate-x-0.5" />
              </div>
            </Link>
          ))
        ) : (
          <div className="rounded-2xl border border-dashed border-border/70 bg-background/70 px-4 py-6 text-sm text-muted-foreground">
            {emptyText}
          </div>
        )}
      </div>
    </div>
  )
}

function alertTone(level: string) {
  switch (level) {
    case "critical":
      return "border-rose-300 bg-rose-50 text-rose-700"
    case "high":
      return "border-amber-300 bg-amber-50 text-amber-700"
    case "medium":
      return "border-sky-300 bg-sky-50 text-sky-700"
    default:
      return "border-emerald-300 bg-emerald-50 text-emerald-700"
  }
}

function formatRelativeDate(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return "soon"
  }
  return formatDistanceToNowStrict(date, { addSuffix: true })
}
