"use client"

import * as React from "react"

import { AdminPageHeader } from "@/components/admin/admin-page-header"
import { AdminStatusBadge } from "@/components/admin/admin-status-badge"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useAdminSelections } from "@/hooks/use-admin-portal"
import { formatShortDate } from "@/lib/admin-utils"

export default function AdminSelectionsPage() {
  const [status, setStatus] = React.useState("all")
  const [search, setSearch] = React.useState("")

  const { data: selections = [], isLoading } = useAdminSelections(status === "all" ? undefined : status)

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

      <Card className="border-slate-200 bg-white/90">
        <CardContent className="grid gap-4 p-5 lg:grid-cols-[1.2fr_0.8fr]">
          <Input
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Search candidate, Ethiopian agency, or Foreign agency"
            className="bg-white"
          />
          <Select value={status} onValueChange={setStatus}>
            <SelectTrigger className="bg-white">
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
        </CardContent>
      </Card>

      <Card className="border-slate-200 bg-white/90">
        <CardContent className="p-0">
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
                  <TableCell colSpan={6} className="py-10 text-center text-sm text-slate-500">Loading selections...</TableCell>
                </TableRow>
              ) : null}
              {!isLoading && !filtered.length ? (
                <TableRow>
                  <TableCell colSpan={6} className="py-10 text-center text-sm text-slate-500">No selections matched the current filters.</TableCell>
                </TableRow>
              ) : null}
              {filtered.map((selection) => (
                <TableRow key={selection.id}>
                  <TableCell className="font-medium text-slate-950">{selection.candidate_name}</TableCell>
                  <TableCell className="text-slate-600">{selection.ethiopian_agency}</TableCell>
                  <TableCell className="text-slate-600">{selection.foreign_agency}</TableCell>
                  <TableCell><AdminStatusBadge status={selection.status} /></TableCell>
                  <TableCell className="text-slate-600">{formatShortDate(selection.selected_date)}</TableCell>
                  <TableCell><AdminStatusBadge status={selection.approval_status} /></TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
