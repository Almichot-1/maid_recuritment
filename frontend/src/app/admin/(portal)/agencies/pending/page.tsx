"use client"

import Link from "next/link"

import { AdminPageHeader } from "@/components/admin/admin-page-header"
import { AdminStatusBadge } from "@/components/admin/admin-status-badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Button } from "@/components/ui/button"
import { usePendingAgencies } from "@/hooks/use-admin-portal"
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
      <div className="rounded-3xl border border-dashed border-slate-300 bg-white p-10 text-center text-sm text-slate-500">
        No agencies are waiting in this queue right now.
      </div>
    )
  }

  return (
    <div className="grid gap-4 xl:grid-cols-2">
      {data.map((agency) => (
        <Card key={agency.id} className="border-slate-200 bg-white/90 shadow-sm">
          <CardHeader className="space-y-3">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <CardTitle className="text-xl text-slate-950">{agency.company_name || agency.contact_person}</CardTitle>
                <p className="mt-1 text-sm text-slate-500">{agency.email}</p>
              </div>
              <AdminStatusBadge status={agency.account_status} />
            </div>
          </CardHeader>
          <CardContent className="space-y-5">
            <dl className="grid gap-3 sm:grid-cols-2">
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wide text-slate-400">Contact Person</dt>
                <dd className="mt-1 text-sm text-slate-700">{agency.contact_person}</dd>
              </div>
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wide text-slate-400">Agency Type</dt>
                <dd className="mt-1 text-sm text-slate-700">{titleize(agency.role)}</dd>
              </div>
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wide text-slate-400">Registered</dt>
                <dd className="mt-1 text-sm text-slate-700">{formatRelative(agency.registration_date)}</dd>
              </div>
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wide text-slate-400">Verification Docs</dt>
                <dd className="mt-1 text-sm text-slate-700">Registration docs are not attached yet in the signup flow.</dd>
              </div>
            </dl>

            <div className="flex items-center justify-between rounded-2xl bg-slate-50 px-4 py-3 text-sm text-slate-600">
              <span>Review application details and decide whether this agency can access the platform.</span>
              <Link href={`/admin/agencies/${agency.id}`}>
                <Button className="bg-slate-950 hover:bg-slate-800">Review</Button>
              </Link>
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}

export default function PendingApprovalsPage() {
  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="Pending Approvals"
        description="Review newly registered Ethiopian and Foreign agencies before they can access the main platform."
      />

      <Tabs defaultValue="all" className="space-y-4">
        <TabsList className="rounded-2xl bg-white p-1 shadow-sm">
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
