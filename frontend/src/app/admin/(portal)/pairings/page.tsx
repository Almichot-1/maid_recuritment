"use client"

import * as React from "react"
import { toast } from "sonner"
import { Link2, PauseCircle, PlayCircle, ShieldCheck } from "lucide-react"

import { AdminPageHeader } from "@/components/admin/admin-page-header"
import { AdminStatusBadge } from "@/components/admin/admin-status-badge"
import { AdminEmptyState, AdminStatCard, AdminSurface } from "@/components/admin/admin-ui"
import { Button } from "@/components/ui/button"
import { CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"
import { useAgencies } from "@/hooks/use-admin-portal"
import { useAdminPairings, useCreateAdminPairing, useUpdateAdminPairing } from "@/hooks/use-pairings"
import { AccountStatus, AgencyPairingStatus, UserRole } from "@/types"
import { formatDateTime } from "@/lib/admin-utils"

function agencyLabel(companyName?: string, contactName?: string) {
  return companyName?.trim() || contactName?.trim() || "Agency"
}

export default function AdminPairingsPage() {
  const [statusFilter, setStatusFilter] = React.useState<string>("all")
  const [ethiopianAgencyId, setEthiopianAgencyId] = React.useState("")
  const [foreignAgencyId, setForeignAgencyId] = React.useState("")
  const [notes, setNotes] = React.useState("")

  const { data: pairings = [], isLoading } = useAdminPairings(statusFilter === "all" ? undefined : { status: statusFilter })
  const { data: ethiopianAgencies = [] } = useAgencies({
    status: AccountStatus.ACTIVE,
    role: UserRole.ETHIOPIAN_AGENT,
    search: "",
  })
  const { data: foreignAgencies = [] } = useAgencies({
    status: AccountStatus.ACTIVE,
    role: UserRole.FOREIGN_AGENT,
    search: "",
  })
  const createPairing = useCreateAdminPairing()
  const updatePairing = useUpdateAdminPairing()
  const activeCount = pairings.filter((pairing) => pairing.status === AgencyPairingStatus.ACTIVE).length
  const suspendedCount = pairings.filter((pairing) => pairing.status === AgencyPairingStatus.SUSPENDED).length
  const endedCount = pairings.filter((pairing) => pairing.status === AgencyPairingStatus.ENDED).length

  const handleCreatePairing = async () => {
    if (!ethiopianAgencyId || !foreignAgencyId) {
      toast.error("Select both agencies before creating a workspace.")
      return
    }

    try {
      await createPairing.mutateAsync({
        ethiopian_user_id: ethiopianAgencyId,
        foreign_user_id: foreignAgencyId,
        notes: notes.trim() || undefined,
      })
      toast.success("Pair workspace created.")
      setForeignAgencyId("")
      setNotes("")
    } catch {
      toast.error("Could not create the pairing. Check that it does not already exist.")
    }
  }

  const handleStatusUpdate = async (pairingId: string, status: AgencyPairingStatus, note: string) => {
    try {
      await updatePairing.mutateAsync({ id: pairingId, status, notes: note })
      toast.success("Pair workspace updated.")
    } catch {
      toast.error("Could not update the workspace status.")
    }
  }

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="Pair Workspaces"
        description="Create and manage the private Ethiopian-to-Foreign agency workspaces that govern candidate visibility, selections, and tracking."
      />

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminStatCard label="Workspaces" value={pairings.length} detail="Current pair relationships" icon={Link2} />
        <AdminStatCard label="Active" value={activeCount} detail="Currently usable by both partners" icon={ShieldCheck} />
        <AdminStatCard label="Suspended" value={suspendedCount} detail="Paused from operator controls" icon={PauseCircle} />
        <AdminStatCard label="Ended" value={endedCount} detail="Closed historical workspaces" icon={PlayCircle} />
      </div>

      <div className="grid gap-6 xl:grid-cols-[420px_minmax(0,1fr)]">
        <AdminSurface className="shadow-sm">
          <CardHeader>
            <CardTitle className="text-lg text-slate-950 dark:text-slate-50">Create a private workspace</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-700 dark:text-slate-200">Ethiopian agency</label>
              <Select value={ethiopianAgencyId} onValueChange={setEthiopianAgencyId}>
                <SelectTrigger>
                  <SelectValue placeholder="Choose an Ethiopian agency" />
                </SelectTrigger>
                <SelectContent>
                  {ethiopianAgencies.map((agency) => (
                    <SelectItem key={agency.id} value={agency.id}>
                      {agencyLabel(agency.company_name, agency.contact_person)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-700 dark:text-slate-200">Foreign agency</label>
              <Select value={foreignAgencyId} onValueChange={setForeignAgencyId}>
                <SelectTrigger>
                  <SelectValue placeholder="Choose a foreign agency" />
                </SelectTrigger>
                <SelectContent>
                  {foreignAgencies.map((agency) => (
                    <SelectItem key={agency.id} value={agency.id}>
                      {agencyLabel(agency.company_name, agency.contact_person)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-700 dark:text-slate-200">Internal note</label>
              <Textarea
                value={notes}
                onChange={(event) => setNotes(event.target.value)}
                placeholder="Optional context for why this workspace is being created"
              />
            </div>

            <Button className="w-full" onClick={handleCreatePairing} disabled={createPairing.isPending}>
              Create private workspace
            </Button>
          </CardContent>
        </AdminSurface>

        <AdminSurface className="shadow-sm">
          <CardHeader className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <CardTitle className="text-lg text-slate-950 dark:text-slate-50">Existing workspaces</CardTitle>
              <p className="text-sm text-slate-500 dark:text-slate-400">Every relationship below is a private operational workspace between one Ethiopian and one foreign agency.</p>
            </div>
            <div className="w-full sm:w-[220px]">
              <Select value={statusFilter} onValueChange={setStatusFilter}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All statuses</SelectItem>
                  <SelectItem value={AgencyPairingStatus.ACTIVE}>Active</SelectItem>
                  <SelectItem value={AgencyPairingStatus.SUSPENDED}>Suspended</SelectItem>
                  <SelectItem value={AgencyPairingStatus.ENDED}>Ended</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            {isLoading ? (
              <div className="admin-empty-state">
                Loading pair workspaces...
              </div>
            ) : pairings.length ? (
              pairings.map((pairing) => (
                <div key={pairing.id} className="space-y-4 rounded-3xl border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-800 dark:bg-slate-900/82">
                  <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                    <div className="space-y-3">
                      <div className="flex flex-wrap items-center gap-2">
                        <AdminStatusBadge status={pairing.status} />
                        <span className="text-xs uppercase tracking-[0.22em] text-slate-400">
                          Approved {formatDateTime(pairing.approved_at)}
                        </span>
                      </div>
                      <div className="grid gap-3 md:grid-cols-2">
                        <WorkspacePartyCard
                          label="Ethiopian agency"
                          title={agencyLabel(pairing.ethiopian_agency.company_name, pairing.ethiopian_agency.full_name)}
                          subtitle={pairing.ethiopian_agency.email}
                        />
                        <WorkspacePartyCard
                          label="Foreign agency"
                          title={agencyLabel(pairing.foreign_agency.company_name, pairing.foreign_agency.full_name)}
                          subtitle={pairing.foreign_agency.email}
                        />
                      </div>
                      {pairing.notes ? (
                        <p className="rounded-2xl bg-slate-50 px-4 py-3 text-sm text-slate-600 dark:bg-slate-950 dark:text-slate-300">{pairing.notes}</p>
                      ) : null}
                    </div>

                    <div className="flex flex-wrap gap-2">
                      {pairing.status !== AgencyPairingStatus.ACTIVE ? (
                        <Button
                          variant="outline"
                          onClick={() => handleStatusUpdate(pairing.id, AgencyPairingStatus.ACTIVE, "Reactivated from admin portal")}
                          disabled={updatePairing.isPending}
                        >
                          Reactivate
                        </Button>
                      ) : null}
                      {pairing.status === AgencyPairingStatus.ACTIVE ? (
                        <Button
                          variant="outline"
                          onClick={() => handleStatusUpdate(pairing.id, AgencyPairingStatus.SUSPENDED, "Suspended from admin portal")}
                          disabled={updatePairing.isPending}
                        >
                          Suspend
                        </Button>
                      ) : null}
                      {pairing.status !== AgencyPairingStatus.ENDED ? (
                        <Button
                          variant="destructive"
                          onClick={() => handleStatusUpdate(pairing.id, AgencyPairingStatus.ENDED, "Ended from admin portal")}
                          disabled={updatePairing.isPending}
                        >
                          End workspace
                        </Button>
                      ) : null}
                    </div>
                  </div>
                </div>
              ))
            ) : (
              <AdminEmptyState
                title="No workspaces matched"
                description="Try switching the status filter or create a new pairing to open a private workspace."
              />
            )}
          </CardContent>
        </AdminSurface>
      </div>
    </div>
  )
}

function WorkspacePartyCard({
  label,
  title,
  subtitle,
}: {
  label: string
  title: string
  subtitle: string
}) {
  return (
    <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-800 dark:bg-slate-950">
      <p className="text-xs font-semibold uppercase tracking-[0.22em] text-slate-400">{label}</p>
      <p className="mt-2 font-semibold text-slate-950 dark:text-slate-100">{title}</p>
      <p className="text-sm text-slate-500 dark:text-slate-400">{subtitle}</p>
    </div>
  )
}
