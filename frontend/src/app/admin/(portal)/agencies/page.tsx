"use client"

import * as React from "react"
import Link from "next/link"
import { toast } from "sonner"

import { AdminPageHeader } from "@/components/admin/admin-page-header"
import { AdminStatusBadge } from "@/components/admin/admin-status-badge"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useAgencies, useUpdateAgencyStatus } from "@/hooks/use-admin-portal"
import { AccountStatus, UserRole } from "@/types"
import { formatShortDate, titleize } from "@/lib/admin-utils"

export default function AgenciesManagementPage() {
  const [status, setStatus] = React.useState<AccountStatus | "all">("all")
  const [role, setRole] = React.useState<UserRole | "all">("all")
  const [search, setSearch] = React.useState("")

  const { data: agencies = [], isLoading } = useAgencies({ status, role, search })
  const { mutateAsync: updateStatus, isPending } = useUpdateAgencyStatus()

  const handleStatusUpdate = async (agencyId: string, nextStatus: AccountStatus) => {
    const reason =
      nextStatus === AccountStatus.SUSPENDED || nextStatus === AccountStatus.REJECTED
        ? window.prompt(`Provide a reason for setting this agency to ${titleize(nextStatus)}:`) ?? ""
        : ""

    if ((nextStatus === AccountStatus.SUSPENDED || nextStatus === AccountStatus.REJECTED) && !reason.trim()) {
      toast.error("A reason is required for this status change.")
      return
    }

    const confirmed = window.confirm(`Are you sure you want to mark this agency as ${titleize(nextStatus)}?`)
    if (!confirmed) {
      return
    }

    await updateStatus({ agencyId, status: nextStatus, reason })
    toast.success(`Agency marked as ${titleize(nextStatus)}.`)
  }

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="Agencies Management"
        description="Monitor every registered agency, filter by status and role, and take operational actions without entering the agency portal."
      />

      <Card className="border-slate-200 bg-white/90">
        <CardContent className="grid gap-4 p-5 lg:grid-cols-[1.3fr_0.8fr_0.8fr]">
          <Input
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Search company, email, or contact person"
            className="bg-white"
          />
          <Select value={status} onValueChange={(value) => setStatus(value as AccountStatus | "all")}>
            <SelectTrigger className="bg-white">
              <SelectValue placeholder="Filter by status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All statuses</SelectItem>
              <SelectItem value={AccountStatus.ACTIVE}>Active</SelectItem>
              <SelectItem value={AccountStatus.PENDING_APPROVAL}>Pending approval</SelectItem>
              <SelectItem value={AccountStatus.SUSPENDED}>Suspended</SelectItem>
              <SelectItem value={AccountStatus.REJECTED}>Rejected</SelectItem>
            </SelectContent>
          </Select>
          <Select value={role} onValueChange={(value) => setRole(value as UserRole | "all")}>
            <SelectTrigger className="bg-white">
              <SelectValue placeholder="Filter by role" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All roles</SelectItem>
              <SelectItem value={UserRole.ETHIOPIAN_AGENT}>Ethiopian agencies</SelectItem>
              <SelectItem value={UserRole.FOREIGN_AGENT}>Foreign agencies</SelectItem>
            </SelectContent>
          </Select>
        </CardContent>
      </Card>

      <Card className="border-slate-200 bg-white/90">
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Company</TableHead>
                <TableHead>Email</TableHead>
                <TableHead>Role</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Registered</TableHead>
                <TableHead>Activity</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading ? (
                <TableRow>
                  <TableCell colSpan={7} className="py-10 text-center text-sm text-slate-500">
                    Loading agencies...
                  </TableCell>
                </TableRow>
              ) : null}

              {!isLoading && !agencies.length ? (
                <TableRow>
                  <TableCell colSpan={7} className="py-10 text-center text-sm text-slate-500">
                    No agencies matched the current filters.
                  </TableCell>
                </TableRow>
              ) : null}

              {agencies.map((agency) => (
                <TableRow key={agency.id}>
                  <TableCell>
                    <div>
                      <p className="font-medium text-slate-950">{agency.company_name || agency.contact_person}</p>
                      <p className="text-xs text-slate-500">{agency.contact_person}</p>
                    </div>
                  </TableCell>
                  <TableCell className="text-slate-600">{agency.email}</TableCell>
                  <TableCell>
                    <Badge variant="outline" className="rounded-full">{titleize(agency.role)}</Badge>
                  </TableCell>
                  <TableCell>
                    <AdminStatusBadge status={agency.account_status} />
                  </TableCell>
                  <TableCell className="text-slate-600">{formatShortDate(agency.registration_date)}</TableCell>
                  <TableCell>
                    {agency.role === UserRole.ETHIOPIAN_AGENT ? (
                      <span className="text-sm text-slate-600">{agency.total_candidates} candidates</span>
                    ) : (
                      <span className="text-sm text-slate-600">{agency.total_selections} selections</span>
                    )}
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center justify-end gap-2">
                      <Link href={`/admin/agencies/${agency.id}`}>
                        <Button variant="outline" size="sm">View</Button>
                      </Link>
                      {agency.account_status !== AccountStatus.SUSPENDED ? (
                        <Button
                          variant="outline"
                          size="sm"
                          disabled={isPending}
                          onClick={() => handleStatusUpdate(agency.id, AccountStatus.SUSPENDED)}
                        >
                          Suspend
                        </Button>
                      ) : (
                        <Button
                          size="sm"
                          disabled={isPending}
                          onClick={() => handleStatusUpdate(agency.id, AccountStatus.ACTIVE)}
                        >
                          Activate
                        </Button>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
