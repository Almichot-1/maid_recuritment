"use client"

import * as React from "react"
import { CheckCheck, Clock3, ShieldAlert, TimerReset } from "lucide-react"

import { AdminPageHeader } from "@/components/admin/admin-page-header"
import { AdminStatusBadge } from "@/components/admin/admin-status-badge"
import { AdminEmptyState, AdminStatCard, AdminSurface, AdminToolbar } from "@/components/admin/admin-ui"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useAdminSelections } from "@/hooks/use-admin-portal"
import { formatShortDate } from "@/lib/admin-utils"

export default function AdminSelectionsPage() {
  const [status, setStatus] = React.useState("all")
  const [search, setSearch] = React.useState("")

  const { data: selections = [], isLoading } = useAdminSelections(status === "all" ? undefined : status)
  const pendingCount = selections.filter((selection) => selection.status === "pending").length
  const approvedCount = selections.filter((selection) => selection.status === "approved").length
  const expiredCount = selections.filter((selection) => selection.status === "expired").length

  const filtered = React.useMemo(
    () =>
      selections.filter((selection) => {
        if (!search) {
          return true
        }
        const needle = search.toLowerCase()
        return (
          selection.candidate_name.toLowerCase().includes(needle) ||
          selection.ethiopian_agency.toLowerCase().includes(needle) ||
          selection.foreign_agency.toLowerCase().includes(needle)
        )
      }),
    [search, selections]
  )

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="Selections & Approvals"
        description="Track all candidate selections across agencies, including approval outcomes and selection activity."
      />

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminStatCard label="Selection records" value={filtered.length} detail="Visible in the current result set" icon={CheckCheck} />
        <AdminStatCard label="Pending" value={pendingCount} detail="Waiting on action or confirmation" icon={Clock3} />
        <AdminStatCard label="Approved" value={approvedCount} detail="Cleared to proceed" icon={ShieldAlert} />
        <AdminStatCard label="Expired" value={expiredCount} detail="Selection windows already closed" icon={TimerReset} />
      </div>

      <AdminToolbar className="grid gap-4 lg:grid-cols-[1.2fr_0.8fr]">
          <Input
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Search candidate, Ethiopian agency, or Foreign agency"
            className="bg-white dark:bg-slate-950"
          />
          <Select value={status} onValueChange={setStatus}>
            <SelectTrigger className="bg-white dark:bg-slate-950">
              <SelectValue placeholder="Filter by status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All statuses</SelectItem>
              <SelectItem value="pending">Pending</SelectItem>
              <SelectItem value="approved">Approved</SelectItem>
              <SelectItem value="rejected">Rejected</SelectItem>
              <SelectItem value="expired">Expired</SelectItem>
            </SelectContent>
          </Select>
      </AdminToolbar>

      <AdminSurface className="overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Candidate</TableHead>
                <TableHead>Ethiopian Agency</TableHead>
                <TableHead>Foreign Agency</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Selected</TableHead>
                <TableHead>Approval State</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading ? (
                <TableRow>
                  <TableCell colSpan={6} className="py-10 text-center text-sm text-slate-500 dark:text-slate-400">Loading selections...</TableCell>
                </TableRow>
              ) : null}
              {!isLoading && !filtered.length ? (
                <TableRow>
                  <TableCell colSpan={6} className="p-6">
                    <AdminEmptyState
                      title="No selections matched"
                      description="Adjust the text search or the selection status filter to bring matching records back."
                    />
                  </TableCell>
                </TableRow>
              ) : null}
              {filtered.map((selection) => (
                <TableRow key={selection.id}>
                  <TableCell className="font-medium text-slate-950 dark:text-slate-100">{selection.candidate_name}</TableCell>
                  <TableCell className="text-slate-600 dark:text-slate-300">{selection.ethiopian_agency}</TableCell>
                  <TableCell className="text-slate-600 dark:text-slate-300">{selection.foreign_agency}</TableCell>
                  <TableCell><AdminStatusBadge status={selection.status} /></TableCell>
                  <TableCell className="text-slate-600 dark:text-slate-300">{formatShortDate(selection.selected_date)}</TableCell>
                  <TableCell><AdminStatusBadge status={selection.approval_status} /></TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
      </AdminSurface>
    </div>
  )
}
