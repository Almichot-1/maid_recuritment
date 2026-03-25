"use client"

import Link from "next/link"
import { Building2, Clock3, Globe2, ShieldCheck } from "lucide-react"

import { AdminPageHeader } from "@/components/admin/admin-page-header"
import { AdminStatusBadge } from "@/components/admin/admin-status-badge"
import { AdminEmptyState, AdminStatCard, AdminSurface } from "@/components/admin/admin-ui"
import { CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Button } from "@/components/ui/button"
import { useAgencies, usePendingAgencies } from "@/hooks/use-admin-portal"
import { UserRole } from "@/types"
import { formatRelative, titleize } from "@/lib/admin-utils"

function PendingAgencyCards({ role }: { role: UserRole | "all" }) {
  const { data = [], isLoading } = usePendingAgencies(role)

  if (isLoading) {
    return (
      <div className="grid gap-4 xl:grid-cols-2">
        {Array.from({ length: 4 }).map((_, index) => (
          <div key={index} className="h-52 rounded-3xl border border-slate-200 bg-white/70" />
        ))}
      </div>
    )
  }

  if (!data.length) {
    return (
      <AdminEmptyState
        title="Queue is clear"
        description="No agencies are waiting in this lane right now. New registrations will appear here automatically."
      />
    )
  }

  return (
    <div className="grid gap-4 xl:grid-cols-2">
      {data.map((agency) => (
        <AdminSurface key={agency.id} className="shadow-sm">
          <CardHeader className="space-y-3">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <CardTitle className="text-xl text-slate-950 dark:text-slate-50">{agency.company_name || agency.contact_person}</CardTitle>
                <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">{agency.email}</p>
              </div>
              <AdminStatusBadge status={agency.account_status} />
            </div>
          </CardHeader>
          <CardContent className="space-y-5">
            <dl className="grid gap-3 sm:grid-cols-2">
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wide text-slate-400">Contact Person</dt>
                <dd className="mt-1 text-sm text-slate-700 dark:text-slate-200">{agency.contact_person}</dd>
              </div>
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wide text-slate-400">Agency Type</dt>
                <dd className="mt-1 text-sm text-slate-700 dark:text-slate-200">{titleize(agency.role)}</dd>
              </div>
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wide text-slate-400">Registered</dt>
                <dd className="mt-1 text-sm text-slate-700 dark:text-slate-200">{formatRelative(agency.registration_date)}</dd>
              </div>
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wide text-slate-400">Verification Docs</dt>
                <dd className="mt-1 text-sm text-slate-700 dark:text-slate-200">Registration docs are not attached yet in the signup flow.</dd>
              </div>
            </dl>

            <div className="admin-muted-surface flex items-center justify-between px-4 py-3 text-sm">
              <span>Review application details and decide whether this agency can access the platform.</span>
              <Link href={`/admin/agencies/${agency.id}`}>
                <Button className="bg-slate-950 hover:bg-slate-800 dark:bg-amber-300 dark:text-slate-950 dark:hover:bg-amber-200">Review</Button>
              </Link>
            </div>
          </CardContent>
        </AdminSurface>
      ))}
    </div>
  )
}

export default function PendingApprovalsPage() {
  const { data: allPending = [] } = usePendingAgencies("all")
  const { data: agencies = [] } = useAgencies({ status: "all", role: "all", search: "" })
  const approvedCount = agencies.filter((agency) => agency.account_status === "active").length
  const ethiopianPending = allPending.filter((agency) => agency.role === UserRole.ETHIOPIAN_AGENT).length
  const foreignPending = allPending.filter((agency) => agency.role === UserRole.FOREIGN_AGENT).length

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="Pending Approvals"
        description="Review newly registered Ethiopian and Foreign agencies before they can access the main platform."
      />

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminStatCard label="Queue size" value={allPending.length} detail="Registrations waiting for review" icon={Clock3} />
        <AdminStatCard label="Ethiopian agencies" value={ethiopianPending} detail="Sourcing-side registrations pending" icon={Building2} />
        <AdminStatCard label="Foreign agencies" value={foreignPending} detail="Hiring-side registrations pending" icon={Globe2} />
        <AdminStatCard label="Approved agencies" value={approvedCount} detail="Already cleared for platform access" icon={ShieldCheck} />
      </div>

      <Tabs defaultValue="all" className="space-y-4">
        <TabsList className="rounded-2xl bg-white p-1 shadow-sm dark:bg-slate-950/80">
          <TabsTrigger value="all">All Agencies</TabsTrigger>
          <TabsTrigger value="ethiopian_agent">Ethiopian Agencies</TabsTrigger>
          <TabsTrigger value="foreign_agent">Foreign Agencies</TabsTrigger>
        </TabsList>
        <TabsContent value="all">
          <PendingAgencyCards role="all" />
        </TabsContent>
        <TabsContent value="ethiopian_agent">
          <PendingAgencyCards role={UserRole.ETHIOPIAN_AGENT} />
        </TabsContent>
        <TabsContent value="foreign_agent">
          <PendingAgencyCards role={UserRole.FOREIGN_AGENT} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
