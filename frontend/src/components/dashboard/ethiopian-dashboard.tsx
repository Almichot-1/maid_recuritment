"use client"

import * as React from "react"
import Link from "next/link"
import { Users, CheckCircle, Lock, Loader, CircleAlert, ArrowRight, UserPlus, FileCheck, CheckSquare, Sparkles } from "lucide-react"

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
import { useCandidates } from "@/hooks/use-candidates"
import { useDashboardStats } from "@/hooks/use-dashboard"
import { useCurrentUser } from "@/hooks/use-auth"
import { usePairingContext } from "@/hooks/use-pairings"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/utils"

export function EthiopianDashboard() {
  const { user } = useCurrentUser()
  const { activeWorkspace } = usePairingContext()
  const { data: stats, isLoading } = useDashboardStats()
  const { data: candidateData, isLoading: isCandidatesLoading } = useCandidates({ page: 1, page_size: 5 })

  if (!user) return null

  const recentCandidates = candidateData?.data || []
  const incompleteProfiles = (candidateData?.data || []).filter((candidate) => {
    const documentTypes = new Set(candidate.documents.map((document) => document.document_type))
    return !documentTypes.has("passport") || !documentTypes.has("photo")
  }).length

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
                  {isCandidatesLoading ? (
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
