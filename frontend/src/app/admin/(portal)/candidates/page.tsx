"use client"

import * as React from "react"

import { AdminPageHeader } from "@/components/admin/admin-page-header"
import { AdminStatusBadge } from "@/components/admin/admin-status-badge"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Button } from "@/components/ui/button"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useAdminCandidates } from "@/hooks/use-admin-portal"
import { formatShortDate } from "@/lib/admin-utils"

export default function AdminCandidatesPage() {
  const [search, setSearch] = React.useState("")
  const [status, setStatus] = React.useState("all")
  const [agency, setAgency] = React.useState("all")
  const [page, setPage] = React.useState(1)
  const [pageSize] = React.useState(20)

  const { data: candidatesData, isLoading } = useAdminCandidates(
    status === "all" ? undefined : status,
    page,
    pageSize
  )
  const candidates = candidatesData?.candidates ?? []
  const meta = candidatesData?.meta

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

  React.useEffect(() => {
    setPage(1)
  }, [status, agency])

  const totalPages = meta ? Math.ceil(meta.count / pageSize) : 1

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="Candidates Overview"
        description="A platform-wide read-only view of every candidate uploaded by Ethiopian agencies."
      />

      <Card className="border-slate-200 bg-white/90">
        <CardContent className="grid gap-4 p-5 lg:grid-cols-[1.3fr_0.8fr_1fr]">
          <Input
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Search candidate or agency"
            className="bg-white"
          />
          <Select value={status} onValueChange={setStatus}>
            <SelectTrigger className="bg-white">
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
            <SelectTrigger className="bg-white">
              <SelectValue placeholder="Filter by agency" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All agencies</SelectItem>
              {agencyOptions.map((option) => (
                <SelectItem key={option} value={option}>{option}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </CardContent>
      </Card>

      <Card className="border-slate-200 bg-white/90">
        <CardContent className="p-0 overflow-x-auto">
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
                  <TableCell colSpan={5} className="py-10 text-center text-sm text-slate-500">Loading candidates...</TableCell>
                </TableRow>
              ) : null}
              {!isLoading && !filtered.length ? (
                <TableRow>
                  <TableCell colSpan={5} className="py-10 text-center text-sm text-slate-500">No candidates matched the current filters.</TableCell>
                </TableRow>
              ) : null}
              {filtered.map((candidate) => (
                <TableRow key={candidate.id}>
                  <TableCell>
                    <div>
                      <p className="font-medium text-slate-950">{candidate.full_name}</p>
                      <p className="text-xs text-slate-500">{candidate.company_name || candidate.agency_name}</p>
                    </div>
                  </TableCell>
                  <TableCell className="text-slate-600">{candidate.age ?? "N/A"}</TableCell>
                  <TableCell><AdminStatusBadge status={candidate.status} /></TableCell>
                  <TableCell className="text-slate-600">{candidate.company_name || candidate.agency_name}</TableCell>
                  <TableCell className="text-slate-600">{formatShortDate(candidate.created_at)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
        {meta && (
          <div className="flex items-center justify-between border-t border-slate-200 px-5 py-4">
            <Button
              variant="outline"
              disabled={page === 1}
              onClick={() => setPage((p) => Math.max(1, p - 1))}
            >
              Previous
            </Button>
            <span className="text-sm text-slate-500">
              Page {page} of {totalPages}
            </span>
            <Button
              variant="outline"
              disabled={page >= totalPages}
              onClick={() => setPage((p) => p + 1)}
            >
              Next
            </Button>
          </div>
        )}
      </Card>
    </div>
  )
}
