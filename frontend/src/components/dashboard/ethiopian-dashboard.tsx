"use client"

import * as React from "react"
import Link from "next/link"
import { formatDistanceToNowStrict } from "date-fns"
import { ArrowRight, CheckSquare, CircleAlert, FileCheck, PlaneLanding, UserPlus } from "lucide-react"

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useDashboardHome, useSmartAlerts } from "@/hooks/use-dashboard"
import { useCurrentUser } from "@/hooks/use-auth"
import { usePairingContext } from "@/hooks/use-pairings"
import { Skeleton } from "@/components/ui/skeleton"
import { useI18n } from "@/lib/i18n"
import { cn } from "@/lib/utils"

export function EthiopianDashboard() {
  const { user } = useCurrentUser()
  const { activeWorkspace } = usePairingContext()
  const { t } = useI18n()
  const { data: home, isLoading } = useDashboardHome()
  const { data: smartAlerts, isLoading: isSmartAlertsLoading } = useSmartAlerts()

  if (!user) return null

  const stats = home?.stats
  const recentCandidates = home?.recent_candidates || []
  const incompleteProfiles = home?.pending_actions?.incompleteProfiles ?? 0

  return (
    <div className="space-y-6 animate-in">
      <Card>
        <CardContent className="grid gap-6 p-6 lg:grid-cols-[minmax(0,1fr)_300px]">
          <div className="space-y-4">
            <p className="section-kicker">{t("dashboard.ethiopianLabel")}</p>
            <h2 className="font-display text-5xl leading-none text-foreground">{t("dashboard.ethiopianTitle")}</h2>
            <p className="max-w-3xl text-sm text-muted-foreground sm:text-base">{t("dashboard.ethiopianBody")}</p>
            {activeWorkspace ? (
              <Badge variant="outline">{t("dashboard.activePartner", { name: activeWorkspace.partner_agency.company_name || activeWorkspace.partner_agency.full_name })}</Badge>
            ) : null}
            <div className="flex flex-wrap gap-3">
              <Button asChild>
                <Link href="/candidates/new">
                  <UserPlus className="h-4 w-4" />
                  {t("dashboard.addCandidate")}
                </Link>
              </Button>
              <Button variant="outline" asChild>
                <Link href="/selections">
                  <CheckSquare className="h-4 w-4" />
                  {t("dashboard.reviewSelections")}
                </Link>
              </Button>
            </div>
          </div>

          <div className="space-y-3 border-l border-border pl-0 lg:pl-6">
            <p className="section-kicker">{t("dashboard.todayAtGlance")}</p>
            <MiniMetric label={t("dashboard.sharedWorkspace")} value={stats?.totalCandidates ?? 0} />
            <MiniMetric label={t("dashboard.profilesTracking")} value={stats?.inProgress ?? 0} />
            <MiniMetric label={t("dashboard.approvalQueue")} value={stats?.activeSelections ?? 0} />
          </div>
        </CardContent>
      </Card>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <StatCard title="Total candidates" value={stats?.totalCandidates} isLoading={isLoading} detail="Agency library size" />
        <StatCard title="Available now" value={stats?.availableCandidates} isLoading={isLoading} detail="Ready to share" />
        <StatCard title="Selected" value={stats?.selectedCandidates} isLoading={isLoading} detail="Held by partner selection" />
        <StatCard title="In progress" value={stats?.inProgress} isLoading={isLoading} detail="Cases already moving" />
      </div>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1.35fr)_360px]">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between gap-4 space-y-0">
            <div>
              <CardTitle>Recent activity</CardTitle>
              <CardDescription>Latest updates to candidate records.</CardDescription>
            </div>
            <Button variant="outline" size="sm" asChild>
              <Link href="/candidates">
                {t("common.viewAll")}
                <ArrowRight className="h-4 w-4" />
              </Link>
            </Button>
          </CardHeader>
          <CardContent>
            <div className="border border-border">
              <Table>
                <TableHeader className="bg-muted/30">
                  <TableRow>
                    <TableHead>Candidate</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="text-right">Created</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {isLoading ? (
                    <TableRow>
                      <TableCell colSpan={3} className="py-6 text-center text-muted-foreground">
                        {t("common.loading")}
                      </TableCell>
                    </TableRow>
                  ) : recentCandidates.length > 0 ? (
                    recentCandidates.map((candidate) => (
                      <TableRow key={candidate.id}>
                        <TableCell className="font-medium">{candidate.full_name}</TableCell>
                        <TableCell>
                          <Badge variant="outline">{candidate.status.replace("_", " ")}</Badge>
                        </TableCell>
                        <TableCell className="text-right text-muted-foreground">{new Date(candidate.created_at).toLocaleDateString()}</TableCell>
                      </TableRow>
                    ))
                  ) : (
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

        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Pending actions</CardTitle>
              <CardDescription>Tasks that still need attention.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <PendingActionLink href="/selections" icon={<FileCheck className="h-5 w-5" />} title={`${stats?.activeSelections ?? 0} selections`} description="Waiting for your review" />
              <PendingActionLink href="/candidates" icon={<CircleAlert className="h-5 w-5" />} title={`${incompleteProfiles} profiles`} description="Missing required documents" />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Quick actions</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <Button className="w-full justify-start" asChild>
                <Link href="/candidates/new">
                  <UserPlus className="h-4 w-4" />
                  {t("dashboard.addCandidate")}
                </Link>
              </Button>
              <Button variant="outline" className="w-full justify-start" asChild>
                <Link href="/selections">
                  <CheckSquare className="h-4 w-4" />
                  {t("dashboard.reviewSelections")}
                </Link>
              </Button>
            </CardContent>
          </Card>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Smart alerts</CardTitle>
          <CardDescription>Expiring selections, document deadlines, and flight-stage activity in the active workspace.</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 xl:grid-cols-4">
          <SmartAlertColumn
            title="Passport expiry"
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
            title="Medical expiry"
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
            title="Selection expiry"
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
            title="Flight updates"
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

function PendingActionLink({
  href,
  icon,
  title,
  description,
}: {
  href: string
  icon: React.ReactNode
  title: string
  description: string
}) {
  return (
    <Link href={href} className="grid grid-cols-[40px_minmax(0,1fr)_20px] items-center gap-4 border border-border bg-background p-4 transition-colors hover:bg-muted/20">
      <div className="flex h-10 w-10 items-center justify-center border border-border text-primary">{icon}</div>
      <div className="min-w-0">
        <p className="text-sm font-bold uppercase tracking-[0.06em] text-foreground">{title}</p>
        <p className="text-xs text-muted-foreground">{description}</p>
      </div>
      <ArrowRight className="h-4 w-4 text-muted-foreground" />
    </Link>
  )
}

function SmartAlertColumn({
  title,
  loading,
  emptyText,
  items,
}: {
  title: string
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
    <div className="border border-border bg-background p-4">
      <p className="route-stamp text-[11px] text-muted-foreground">{title}</p>
      <div className="mt-4 space-y-3">
        {loading ? (
          <>
            <Skeleton className="h-20 w-full" />
            <Skeleton className="h-20 w-full" />
          </>
        ) : items.length > 0 ? (
          items.map((item) => (
            <Link key={item.id} href={item.href} className="block border border-border bg-card p-4 transition-colors hover:bg-muted/20">
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <p className="truncate text-sm font-bold uppercase tracking-[0.06em] text-foreground">{item.title}</p>
                  <p className="mt-1 text-xs text-muted-foreground">{item.meta}</p>
                </div>
                <Badge variant="outline" className={cn(alertTone(item.tone), item.arrived && "border-[color:var(--color-success)] text-[color:var(--color-success)]")}>
                  {item.arrived ? <PlaneLanding className="h-3.5 w-3.5" /> : null}
                  {item.arrived ? "Arrived" : item.tone}
                </Badge>
              </div>
              <p className="mt-3 text-sm text-muted-foreground">{item.detail}</p>
            </Link>
          ))
        ) : (
          <div className="border border-dashed border-border px-4 py-6 text-sm text-muted-foreground">{emptyText}</div>
        )}
      </div>
    </div>
  )
}

function alertTone(level: string) {
  switch (level) {
    case "critical":
      return "border-[color:var(--color-danger)] text-[color:var(--color-danger)]"
    case "high":
      return "border-[color:var(--color-warning)] text-[color:var(--color-warning)]"
    case "medium":
      return "border-[color:var(--color-info)] text-[color:var(--color-info)]"
    default:
      return "border-[color:var(--color-success)] text-[color:var(--color-success)]"
  }
}

function formatRelativeDate(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return "soon"
  }
  return formatDistanceToNowStrict(date, { addSuffix: true })
}
