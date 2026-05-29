"use client"

import { toast } from "sonner"

import { AdminPageHeader } from "@/components/admin/admin-page-header"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { useAgencies, useAdminCandidates, useAdminDashboard, useAdminSelections, usePendingAgencies } from "@/hooks/use-admin-portal"
import { downloadCsv, formatPercent } from "@/lib/admin-utils"

export default function AdminReportsPage() {
  const { data: stats } = useAdminDashboard()
  const { data: agencies = [] } = useAgencies({ status: "all", role: "all", search: "" })
  const { data: pending = [] } = usePendingAgencies("all")
  const { data: candidates = [] } = useAdminCandidates()
  const { data: selections = [] } = useAdminSelections()

  const reportCards = [
    {
      title: "Agency Performance Report",
      description: "Exports registered agencies with role, status, and activity metrics.",
      action: () => {
        downloadCsv("agency-performance-report.csv", agencies.map((agency) => ({
          company_name: agency.company_name,
          contact_person: agency.contact_person,
          email: agency.email,
          role: agency.role,
          account_status: agency.account_status,
          registration_date: agency.registration_date,
          total_candidates: agency.total_candidates,
          total_selections: agency.total_selections,
        })))
        toast.success("Agency performance CSV downloaded.")
      },
    },
    {
      title: "Recruitment Activity Report",
      description: "Exports every selection with both agencies and approval status.",
      action: () => {
        downloadCsv("recruitment-activity-report.csv", selections.map((selection) => ({
          candidate_name: selection.candidate_name,
          ethiopian_agency: selection.ethiopian_agency,
          foreign_agency: selection.foreign_agency,
          status: selection.status,
          approval_status: selection.approval_status,
          selected_date: selection.selected_date,
        })))
        toast.success("Recruitment activity CSV downloaded.")
      },
    },
    {
      title: "Pending Approvals Report",
      description: "Exports the current review queue for agency onboarding.",
      action: () => {
        downloadCsv("pending-approvals-report.csv", pending.map((agency) => ({
          company_name: agency.company_name,
          contact_person: agency.contact_person,
          email: agency.email,
          role: agency.role,
          registration_date: agency.registration_date,
          account_status: agency.account_status,
        })))
        toast.success("Pending approvals CSV downloaded.")
      },
    },
    {
      title: "Candidate Status Summary",
      description: "Exports every candidate with status, age, and source agency.",
      action: () => {
        downloadCsv("candidate-status-summary.csv", candidates.map((candidate) => ({
          full_name: candidate.full_name,
          age: candidate.age ?? "",
          status: candidate.status,
          agency_name: candidate.company_name || candidate.agency_name,
          created_at: candidate.created_at,
        })))
        toast.success("Candidate status CSV downloaded.")
      },
    },
  ]

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="Reports & Analytics"
        description="Export operational reports and keep a quick read on platform metrics without leaving the admin portal."
      />

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <SummaryCard label="Pending approvals" value={stats?.pending_approvals ?? 0} />
        <SummaryCard label="Total candidates" value={stats?.total_candidates ?? 0} />
        <SummaryCard label="Active selections" value={stats?.active_selections ?? 0} />
        <SummaryCard label="Success rate" value={stats ? formatPercent(stats.success_rate) : "0.0%"} />
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        {reportCards.map((report) => (
          <Card key={report.title} className="border-border bg-card">
            <CardHeader>
              <CardTitle className="text-lg text-foreground">{report.title}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <p className="text-sm text-muted-foreground">{report.description}</p>
              <Button className="bg-slate-950 hover:bg-slate-800" onClick={report.action}>Download CSV</Button>
            </CardContent>
          </Card>
        ))}
      </div>

      <Card className="border-amber-200 bg-amber-50">
        <CardContent className="p-5 text-sm text-amber-900">
          CSV exports are live now. PDF and Excel exports, plus scheduled recurring report delivery, are staged for the next backend reporting phase.
        </CardContent>
      </Card>
    </div>
  )
}

function SummaryCard({ label, value }: { label: string; value: string | number }) {
  return (
    <Card className="border-border bg-card">
      <CardContent className="p-5">
        <p className="text-sm text-muted-foreground">{label}</p>
        <p className="mt-2 text-3xl font-semibold text-foreground">{value}</p>
      </CardContent>
    </Card>
  )
}
