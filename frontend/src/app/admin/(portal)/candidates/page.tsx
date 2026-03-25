"use client"

import * as React from "react"
import { CheckCheck, Layers3, LockKeyhole, SearchCheck } from "lucide-react"

import { AdminPageHeader } from "@/components/admin/admin-page-header"
import { AdminStatusBadge } from "@/components/admin/admin-status-badge"
import { AdminEmptyState, AdminStatCard, AdminSurface, AdminToolbar } from "@/components/admin/admin-ui"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useAdminCandidates } from "@/hooks/use-admin-portal"
import { formatShortDate } from "@/lib/admin-utils"

export default function AdminCandidatesPage() {
  const [search, setSearch] = React.useState("")
  const [status, setStatus] = React.useState("all")
  const [agency, setAgency] = React.useState("all")

  const { data: candidates = [], isLoading } = useAdminCandidates(status === "all" ? undefined : status)
  const availableCount = candidates.filter((candidate) => candidate.status === "available").length
  const reviewCount = candidates.filter((candidate) => candidate.status === "under_review").length
  const lockedCount = candidates.filter((candidate) => candidate.status === "locked").length

  const filtered = React.useMemo(() => {
    return candidates.filter((candidate) => {
      const matchesSearch =
        !search ||
        candidate.full_name.toLowerCase().includes(search.toLowerCase()) ||
        candidate.company_name.toLowerCase().includes(search.toLowerCase()) ||
        candidate.agency_name.toLowerCase().includes(search.toLowerCase())
      const matchesAgency =
        agency === "all" ||
        candidate.company_name.toLowerCase() === agency.toLowerCase() ||
        candidate.agency_name.toLowerCase() === agency.toLowerCase()
      return matchesSearch && matchesAgency
    })
  }, [agency, candidates, search])

  const agencyOptions = React.useMemo(() => {
    const values = new Set<string>()
    for (const candidate of candidates) {
      values.add(candidate.company_name || candidate.agency_name)
    }
    return Array.from(values).sort()
  }, [candidates])

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="Candidates Overview"
        description="A platform-wide read-only view of every candidate uploaded by Ethiopian agencies."
      />

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminStatCard label="Candidate records" value={filtered.length} detail="Visible in the current result set" icon={Layers3} />
        <AdminStatCard label="Available" value={availableCount} detail="Open for partner selection" icon={SearchCheck} />
        <AdminStatCard label="Under review" value={reviewCount} detail="In active processing stages" icon={CheckCheck} />
        <AdminStatCard label="Locked" value={lockedCount} detail="Temporarily reserved candidates" icon={LockKeyhole} />
      </div>

      <AdminToolbar className="grid gap-4 lg:grid-cols-[1.3fr_0.8fr_1fr]">
          <Input
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Search candidate or agency"
            className="bg-white dark:bg-slate-950"
          />
          <Select value={status} onValueChange={setStatus}>
            <SelectTrigger className="bg-white dark:bg-slate-950">
              <SelectValue placeholder="Filter by status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All statuses</SelectItem>
              <SelectItem value="draft">Draft</SelectItem>
              <SelectItem value="available">Available</SelectItem>
              <SelectItem value="locked">Locked</SelectItem>
              <SelectItem value="under_review">Under review</SelectItem>
              <SelectItem value="approved">Approved</SelectItem>
              <SelectItem value="in_progress">In progress</SelectItem>
              <SelectItem value="completed">Completed</SelectItem>
            </SelectContent>
          </Select>
          <Select value={agency} onValueChange={setAgency}>
            <SelectTrigger className="bg-white dark:bg-slate-950">
              <SelectValue placeholder="Filter by agency" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All agencies</SelectItem>
              {agencyOptions.map((option) => (
                <SelectItem key={option} value={option}>{option}</SelectItem>
              ))}
            </SelectContent>
          </Select>
      </AdminToolbar>

      <AdminSurface className="overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Candidate</TableHead>
                <TableHead>Age</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Agency</TableHead>
                <TableHead>Created</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading ? (
                <TableRow>
                  <TableCell colSpan={5} className="py-10 text-center text-sm text-slate-500 dark:text-slate-400">Loading candidates...</TableCell>
                </TableRow>
              ) : null}
              {!isLoading && !filtered.length ? (
                <TableRow>
                  <TableCell colSpan={5} className="p-6">
                    <AdminEmptyState
                      title="No candidates matched"
                      description="Try clearing the agency or status filter, or broaden the name search to bring candidate records back."
                    />
                  </TableCell>
                </TableRow>
              ) : null}
              {filtered.map((candidate) => (
                <TableRow key={candidate.id}>
                  <TableCell>
                    <div>
                      <p className="font-medium text-slate-950 dark:text-slate-100">{candidate.full_name}</p>
                      <p className="text-xs text-slate-500 dark:text-slate-400">{candidate.company_name || candidate.agency_name}</p>
                    </div>
                  </TableCell>
                  <TableCell className="text-slate-600 dark:text-slate-300">{candidate.age ?? "N/A"}</TableCell>
                  <TableCell><AdminStatusBadge status={candidate.status} /></TableCell>
                  <TableCell className="text-slate-600 dark:text-slate-300">{candidate.company_name || candidate.agency_name}</TableCell>
                  <TableCell className="text-slate-600 dark:text-slate-300">{formatShortDate(candidate.created_at)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
      </AdminSurface>
    </div>
  )
}
