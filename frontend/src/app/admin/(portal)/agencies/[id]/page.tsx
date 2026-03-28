"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { toast } from "sonner"

import { AdminPageHeader } from "@/components/admin/admin-page-header"
import { AdminStatusBadge } from "@/components/admin/admin-status-badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"
import { useAgency, useApproveAgency, useRejectAgency, useUpdateAgencyStatus } from "@/hooks/use-admin-portal"
import { useAgencyPairings } from "@/hooks/use-pairings"
import { AccountStatus } from "@/types"
import { formatDateTime, titleize } from "@/lib/admin-utils"

interface AgencyDetailPageParams {
  id: string
}

interface AgencyDetailPageProps {
  params: Promise<AgencyDetailPageParams>
}

export default function AdminAgencyDetailPage({ params }: AgencyDetailPageProps) {
  const { id } = React.use(params)
  const router = useRouter()
  const { data, isLoading } = useAgency(id)
  const { data: pairings = [], isLoading: pairingsLoading } = useAgencyPairings(id)
  const { mutateAsync: approveAgency, isPending: approving } = useApproveAgency()
  const { mutateAsync: rejectAgency, isPending: rejecting } = useRejectAgency()
  const { mutateAsync: updateStatus, isPending: updatingStatus } = useUpdateAgencyStatus()

  const [statusDraft, setStatusDraft] = React.useState<AccountStatus>(AccountStatus.ACTIVE)
  const [statusReason, setStatusReason] = React.useState("")
  const [rejectOpen, setRejectOpen] = React.useState(false)
  const [rejectReason, setRejectReason] = React.useState("")
  const [rejectNotes, setRejectNotes] = React.useState("")

  React.useEffect(() => {
    if (data?.agency.account_status) {
      setStatusDraft(data.agency.account_status)
    }
  }, [data?.agency.account_status])

  if (isLoading || !data) {
    return <div className="rounded-3xl border border-slate-200 bg-white p-10 text-center text-sm text-slate-500">Loading agency profile...</div>
  }

  const handleApprove = async () => {
    const confirmed = window.confirm(`Approve ${data.agency.company_name || data.agency.contact_person}?`)
    if (!confirmed) {
      return
    }
    await approveAgency(data.agency.id)
    toast.success("Agency approved.")
    router.refresh()
  }

  const handleReject = async () => {
    if (!rejectReason.trim()) {
      toast.error("Rejection reason is required.")
      return
    }
    await rejectAgency({ agencyId: data.agency.id, reason: rejectReason, notes: rejectNotes })
    toast.success("Agency rejected.")
    setRejectOpen(false)
  }

  const handleApplyStatus = async () => {
    if ((statusDraft === AccountStatus.SUSPENDED || statusDraft === AccountStatus.REJECTED) && !statusReason.trim()) {
      toast.error("A reason is required for suspended or rejected status.")
      return
    }
    await updateStatus({ agencyId: data.agency.id, status: statusDraft, reason: statusReason })
    toast.success("Agency status updated.")
  }

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title={data.agency.company_name || data.agency.contact_person}
        description="Review registration details, operational history, and current account controls for this agency."
        action={
          <div className="flex flex-wrap items-center gap-3">
            <AdminStatusBadge status={data.agency.account_status} />
            {data.approval_status === "pending" ? (
              <>
                <Button className="bg-emerald-600 hover:bg-emerald-700" disabled={approving} onClick={handleApprove}>
                  Approve Agency
                </Button>
                <Button variant="destructive" disabled={rejecting} onClick={() => setRejectOpen(true)}>
                  Reject Agency
                </Button>
              </>
            ) : null}
          </div>
        }
      />

      <div className="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
        <Card className="border-slate-200 bg-white/90">
          <CardHeader>
            <CardTitle className="text-lg text-slate-950">Agency profile</CardTitle>
          </CardHeader>
          <CardContent className="grid gap-4 sm:grid-cols-2">
            <InfoItem label="Company name" value={data.agency.company_name || "Not provided"} />
            <InfoItem label="Contact person" value={data.agency.contact_person} />
            <InfoItem label="Email" value={data.agency.email} />
            <InfoItem label="Agency type" value={titleize(data.agency.role)} />
            <InfoItem label="Registration date" value={formatDateTime(data.agency.registration_date)} />
            <InfoItem label="Approval state" value={titleize(data.approval_status)} />
            <InfoItem label="Rejection reason" value={data.rejection_reason || "None"} />
            <InfoItem label="Internal notes" value={data.admin_notes || "No internal notes recorded yet."} />
          </CardContent>
        </Card>

        <Card className="border-slate-200 bg-white/90">
          <CardHeader>
            <CardTitle className="text-lg text-slate-950">Account controls</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-700">Account status</label>
              <Select value={statusDraft} onValueChange={(value) => setStatusDraft(value as AccountStatus)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={AccountStatus.ACTIVE}>Active</SelectItem>
                  <SelectItem value={AccountStatus.SUSPENDED}>Suspended</SelectItem>
                  <SelectItem value={AccountStatus.REJECTED}>Rejected</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-700">Reason / admin note</label>
              <Textarea
                value={statusReason}
                onChange={(event) => setStatusReason(event.target.value)}
                placeholder="Explain the reason for this status change"
              />
            </div>
            <Button className="w-full" disabled={updatingStatus} onClick={handleApplyStatus}>
              Apply Status Change
            </Button>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <Card className="border-slate-200 bg-white/90">
          <CardHeader>
            <CardTitle className="text-lg text-slate-950">Activity summary</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <MetricItem label="Total candidates" value={data.activity_summary.total_candidates} />
            <MetricItem label="Active candidates" value={data.activity_summary.active_candidates} />
            <MetricItem label="Completed recruitments" value={data.activity_summary.completed_recruitments} />
            <MetricItem label="Total selections" value={data.activity_summary.total_selections} />
            <MetricItem label="Approved selections" value={data.activity_summary.approved_selections} />
            <MetricItem label="Active recruitments" value={data.activity_summary.active_recruitments} />
          </CardContent>
        </Card>

        <Card className="border-slate-200 bg-white/90 lg:col-span-2">
          <CardHeader>
            <CardTitle className="text-lg text-slate-950">Recent activity</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {data.recent_activity.map((item) => (
              <div key={item.id} className="flex items-center justify-between rounded-2xl border border-slate-200 p-4">
                <div>
                  <p className="font-medium text-slate-950">{item.title}</p>
                  <p className="text-sm text-slate-500">{titleize(item.type)} • {formatDateTime(item.occurred_at)}</p>
                </div>
                <AdminStatusBadge status={item.status} />
              </div>
            ))}
            {!data.recent_activity.length ? (
              <div className="rounded-2xl border border-dashed border-slate-300 p-6 text-sm text-slate-500">
                No recent activity has been recorded for this agency yet.
              </div>
            ) : null}
          </CardContent>
        </Card>
      </div>

      <Card className="border-slate-200 bg-white/90">
        <CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <CardTitle className="text-lg text-slate-950">Pair workspaces</CardTitle>
            <p className="text-sm text-slate-500">Private Ethiopian–Foreign relationships connected to this agency.</p>
          </div>
          <Button variant="outline" onClick={() => router.push("/admin/pairings")}>
            Open pairings manager
          </Button>
        </CardHeader>
        <CardContent className="space-y-3">
          {pairingsLoading ? (
            <div className="rounded-2xl border border-dashed border-slate-300 p-6 text-sm text-slate-500">
              Loading pair workspaces...
            </div>
          ) : pairings.length ? (
            pairings.map((pairing) => (
              <div key={pairing.id} className="flex flex-col gap-3 rounded-2xl border border-slate-200 p-4 sm:flex-row sm:items-center sm:justify-between">
                <div className="space-y-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <AdminStatusBadge status={pairing.status} />
                    <p className="text-sm font-medium text-slate-950">
                      {pairing.ethiopian_agency.company_name || pairing.ethiopian_agency.full_name} ↔ {pairing.foreign_agency.company_name || pairing.foreign_agency.full_name}
                    </p>
                  </div>
                  <p className="text-sm text-slate-500">
                    Approved {formatDateTime(pairing.approved_at)}
                  </p>
                  {pairing.notes ? <p className="text-sm text-slate-500">{pairing.notes}</p> : null}
                </div>
              </div>
            ))
          ) : (
            <div className="rounded-2xl border border-dashed border-slate-300 p-6 text-sm text-slate-500">
              This agency is not connected to any private workspaces yet.
            </div>
          )}
        </CardContent>
      </Card>

      <Card className="border-slate-200 bg-white/90">
        <CardHeader>
          <CardTitle className="text-lg text-slate-950">Verification documents</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {data.submitted_documents.length ? (
            data.submitted_documents.map((document, index) => (
              <div key={`${document.id ?? index}`} className="rounded-2xl border border-slate-200 p-4 text-sm text-slate-600">
                {Object.entries(document).map(([key, value]) => (
                  <p key={key}>
                    <span className="font-medium text-slate-900">{titleize(key)}:</span> {value}
                  </p>
                ))}
              </div>
            ))
          ) : (
            <div className="rounded-2xl border border-dashed border-slate-300 p-6 text-sm text-slate-500">
              Agency verification document upload is not captured in the current signup flow yet, so this section stays empty for now.
            </div>
          )}
        </CardContent>
      </Card>

      <Dialog open={rejectOpen} onOpenChange={setRejectOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Reject agency application</DialogTitle>
            <DialogDescription>Provide a clear reason so the agency understands why access was denied.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Rejection reason</label>
              <Input value={rejectReason} onChange={(event) => setRejectReason(event.target.value)} placeholder="Incomplete documentation, invalid license, etc." />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Internal notes</label>
              <Textarea value={rejectNotes} onChange={(event) => setRejectNotes(event.target.value)} placeholder="Optional internal context for the admin team" />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRejectOpen(false)}>Cancel</Button>
            <Button variant="destructive" disabled={rejecting} onClick={handleReject}>Confirm Rejection</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

function InfoItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
      <p className="text-xs font-semibold uppercase tracking-wide text-slate-400">{label}</p>
      <p className="mt-2 text-sm text-slate-700">{value}</p>
    </div>
  )
}

function MetricItem({ label, value }: { label: string; value: number }) {
  return (
    <div className="flex items-center justify-between rounded-2xl border border-slate-200 px-4 py-3">
      <span className="text-sm text-slate-500">{label}</span>
      <span className="text-lg font-semibold text-slate-950">{value}</span>
    </div>
  )
}
