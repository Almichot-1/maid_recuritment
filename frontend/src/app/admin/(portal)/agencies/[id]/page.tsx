"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { toast } from "sonner"
import { CheckCheck, ClipboardList, Link2, Loader2, ShieldCheck } from "lucide-react"

import { AdminEmptyState, AdminInfoTile, AdminMetricRow, AdminStatCard, AdminSurface } from "@/components/admin/admin-ui"
import { AdminPageHeader } from "@/components/admin/admin-page-header"
import { AdminStatusBadge } from "@/components/admin/admin-status-badge"
import { Button } from "@/components/ui/button"
import { CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"
import { useApproveAgency, useAgency, useRejectAgency, useUpdateAgencyStatus } from "@/hooks/use-admin-portal"
import { useAgencyPairings } from "@/hooks/use-pairings"
import { formatDateTime, titleize } from "@/lib/admin-utils"
import { AccountStatus } from "@/types"

export default function AdminAgencyDetailPage({ params }: { params: { id: string } }) {
  const router = useRouter()
  const { data, isLoading } = useAgency(params.id)
  const { data: pairings = [], isLoading: pairingsLoading } = useAgencyPairings(params.id)
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
    return (
      <AdminSurface className="p-10">
        <div className="flex items-center justify-center gap-3 text-sm text-slate-500 dark:text-slate-400">
          <Loader2 className="h-4 w-4 animate-spin" />
          Loading agency profile...
        </div>
      </AdminSurface>
    )
  }

  const agencyLabel = data.agency.company_name || data.agency.contact_person

  const handleApprove = async () => {
    const confirmed = window.confirm(`Approve ${agencyLabel}?`)
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
    setRejectReason("")
    setRejectNotes("")
    router.refresh()
  }

  const handleApplyStatus = async () => {
    if ((statusDraft === AccountStatus.SUSPENDED || statusDraft === AccountStatus.REJECTED) && !statusReason.trim()) {
      toast.error("A reason is required for suspended or rejected status.")
      return
    }

    await updateStatus({ agencyId: data.agency.id, status: statusDraft, reason: statusReason })
    toast.success("Agency status updated.")
    router.refresh()
  }

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title={agencyLabel}
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

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminStatCard label="Approval state" value={titleize(data.approval_status)} detail="Registration review stage" icon={ShieldCheck} />
        <AdminStatCard label="Active recruitments" value={data.activity_summary.active_recruitments} detail="Currently in motion" icon={CheckCheck} />
        <AdminStatCard label="Pair workspaces" value={pairings.length} detail="Private linked partners" icon={Link2} />
        <AdminStatCard label="Documents" value={data.submitted_documents.length} detail="Verification items on file" icon={ClipboardList} />
      </div>

      <div className="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
        <AdminSurface>
          <CardHeader>
            <CardTitle className="text-lg text-slate-950 dark:text-slate-50">Agency profile</CardTitle>
          </CardHeader>
          <CardContent className="grid gap-4 sm:grid-cols-2">
            <AdminInfoTile label="Company name" value={data.agency.company_name || "Not provided"} />
            <AdminInfoTile label="Contact person" value={data.agency.contact_person} />
            <AdminInfoTile label="Email" value={data.agency.email} />
            <AdminInfoTile label="Agency type" value={titleize(data.agency.role)} />
            <AdminInfoTile label="Registration date" value={formatDateTime(data.agency.registration_date)} />
            <AdminInfoTile label="Approval state" value={titleize(data.approval_status)} />
            <AdminInfoTile label="Rejection reason" value={data.rejection_reason || "None"} />
            <AdminInfoTile label="Internal notes" value={data.admin_notes || "No internal notes recorded yet."} className="sm:col-span-2" />
          </CardContent>
        </AdminSurface>

        <AdminSurface>
          <CardHeader>
            <CardTitle className="text-lg text-slate-950 dark:text-slate-50">Account controls</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-700 dark:text-slate-200">Account status</label>
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
              <label className="text-sm font-medium text-slate-700 dark:text-slate-200">Reason / admin note</label>
              <Textarea
                value={statusReason}
                onChange={(event) => setStatusReason(event.target.value)}
                placeholder="Explain the reason for this status change"
              />
            </div>

            <div className="rounded-2xl border border-slate-200/80 bg-white/80 px-4 py-3 text-sm text-slate-600 dark:border-slate-800 dark:bg-slate-900/78 dark:text-slate-300">
              Use this panel for operational restrictions only. No backend workflow rules are changed here, only the agency account state.
            </div>

            <Button className="w-full" disabled={updatingStatus} onClick={handleApplyStatus}>
              Apply Status Change
            </Button>
          </CardContent>
        </AdminSurface>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <AdminSurface>
          <CardHeader>
            <CardTitle className="text-lg text-slate-950 dark:text-slate-50">Activity summary</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <AdminMetricRow label="Total candidates" value={data.activity_summary.total_candidates} />
            <AdminMetricRow label="Active candidates" value={data.activity_summary.active_candidates} />
            <AdminMetricRow label="Completed recruitments" value={data.activity_summary.completed_recruitments} />
            <AdminMetricRow label="Total selections" value={data.activity_summary.total_selections} />
            <AdminMetricRow label="Approved selections" value={data.activity_summary.approved_selections} />
            <AdminMetricRow label="Active recruitments" value={data.activity_summary.active_recruitments} />
          </CardContent>
        </AdminSurface>

        <AdminSurface className="lg:col-span-2">
          <CardHeader>
            <CardTitle className="text-lg text-slate-950 dark:text-slate-50">Recent activity</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {data.recent_activity.map((item) => (
              <div key={item.id} className="flex items-center justify-between rounded-2xl border border-slate-200/80 bg-white/75 p-4 dark:border-slate-800 dark:bg-slate-900/82">
                <div>
                  <p className="font-medium text-slate-950 dark:text-slate-100">{item.title}</p>
                  <p className="text-sm text-slate-500 dark:text-slate-400">{titleize(item.type)} • {formatDateTime(item.occurred_at)}</p>
                </div>
                <AdminStatusBadge status={item.status} />
              </div>
            ))}
            {!data.recent_activity.length ? (
              <AdminEmptyState
                title="No recent activity yet"
                description="Recent operational events for this agency will appear here as soon as the first workflow begins."
              />
            ) : null}
          </CardContent>
        </AdminSurface>
      </div>

      <AdminSurface>
        <CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <CardTitle className="text-lg text-slate-950 dark:text-slate-50">Pair workspaces</CardTitle>
            <p className="text-sm text-slate-500 dark:text-slate-400">Private Ethiopian-to-Foreign relationships connected to this agency.</p>
          </div>
          <Button variant="outline" onClick={() => router.push("/admin/pairings")}>
            Open pairings manager
          </Button>
        </CardHeader>
        <CardContent className="space-y-3">
          {pairingsLoading ? (
            <AdminEmptyState title="Loading workspaces" description="Fetching the linked pair workspaces for this agency." />
          ) : pairings.length ? (
            pairings.map((pairing) => (
              <div key={pairing.id} className="flex flex-col gap-3 rounded-2xl border border-slate-200/80 bg-white/80 p-4 sm:flex-row sm:items-center sm:justify-between dark:border-slate-800 dark:bg-slate-900/82">
                <div className="space-y-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <AdminStatusBadge status={pairing.status} />
                    <p className="text-sm font-medium text-slate-950 dark:text-slate-100">
                      {pairing.ethiopian_agency.company_name || pairing.ethiopian_agency.full_name} {" -> "} {pairing.foreign_agency.company_name || pairing.foreign_agency.full_name}
                    </p>
                  </div>
                  <p className="text-sm text-slate-500 dark:text-slate-400">Approved {formatDateTime(pairing.approved_at)}</p>
                  {pairing.notes ? <p className="text-sm text-slate-500 dark:text-slate-400">{pairing.notes}</p> : null}
                </div>
              </div>
            ))
          ) : (
            <AdminEmptyState title="No pair workspaces yet" description="This agency is not connected to any private workspaces right now." />
          )}
        </CardContent>
      </AdminSurface>

      <AdminSurface>
        <CardHeader>
          <CardTitle className="text-lg text-slate-950 dark:text-slate-50">Verification documents</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {data.submitted_documents.length ? (
            data.submitted_documents.map((document, index) => (
              <div key={`${document.id ?? index}`} className="rounded-2xl border border-slate-200/80 bg-white/80 p-4 text-sm text-slate-600 dark:border-slate-800 dark:bg-slate-900/82 dark:text-slate-300">
                {Object.entries(document).map(([key, value]) => (
                  <p key={key}>
                    <span className="font-medium text-slate-950 dark:text-slate-100">{titleize(key)}:</span> {String(value)}
                  </p>
                ))}
              </div>
            ))
          ) : (
            <AdminEmptyState
              title="No verification documents captured"
              description="Agency verification upload is not wired into the current signup flow yet, so this panel stays empty for now."
            />
          )}
        </CardContent>
      </AdminSurface>

      <Dialog open={rejectOpen} onOpenChange={setRejectOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Reject agency application</DialogTitle>
            <DialogDescription>Provide a clear reason so the agency understands why access was denied.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-700 dark:text-slate-200">Rejection reason</label>
              <Input value={rejectReason} onChange={(event) => setRejectReason(event.target.value)} placeholder="Incomplete documentation, invalid license, etc." />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-700 dark:text-slate-200">Internal notes</label>
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
