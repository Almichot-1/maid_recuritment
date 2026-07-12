"use client";

import { useState, useMemo, useCallback, type ReactNode } from "react";
import {
  RefreshCw,
  Download,
  Upload,
  UserPlus,
  Sparkles,
  Route,
  TimerReset,
  CheckCircle2,
  Loader2,
} from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { useCurrentUser } from "@/hooks/use-auth";
import { useMySelections } from "@/hooks/use-selections";
import { useSelectionProgress, useUpdateProgress, useBatchUpdateProgress } from "@/hooks/use-selection-progress";
import { useCandidate, useUpdateCandidateStatus } from "@/hooks/use-candidates";
import { SelectionStatus } from "@/types";
import type { StepItem, SelectionProgress } from "@/types";

import { SearchToolbar, DEFAULT_FILTERS, type TrackingFilters } from "./search-toolbar";
import { ApplicantTableRow, type ApplicantRowData } from "./applicant-table-row";
import { BulkToolbar } from "./bulk-toolbar";
import { ApplicantDrawer, type DrawerApplicant } from "./applicant-drawer";

// ─── Status mapping helpers ────────────────────────────────────────────────────
function mapNewStatusToOld(stepName: string, newStatus: string): string {
  switch (stepName) {
    case 'Medical':
    case 'CoC':
      switch (newStatus) {
        case 'done':       return 'completed';
        case 'pending':    return 'pending';
        case 'in_progress': return 'in_progress';
        case 'failed':     return 'failed';
        default:           return newStatus;
      }
    case 'Visa':
      switch (newStatus) {
        case 'approved':   return 'completed';
        case 'rejected':   return 'failed';
        case 'pending':    return 'pending';
        case 'in_progress': return 'in_progress';
        default:           return newStatus;
      }
    case 'Ticket':
      switch (newStatus) {
        case 'pending':    return 'pending';
        case 'booked':     return 'in_progress';
        case 'confirmed':  return 'completed';
        case 'arrived':    return 'completed';
        default:           return newStatus;
      }
    case 'Arrival City':
    case 'Arrival':
      switch (newStatus) {
        case 'not_arrived': return 'pending';
        case 'in_transit':  return 'in_progress';
        case 'arrived':     return 'completed';
        case 'failed':      return 'failed';
        default:            return newStatus;
      }
    default:
      return newStatus;
  }
}

export function mapOldStatusToNew(stepName: string, oldStatus: string): string {
  switch (stepName) {
    case 'Medical':
    case 'CoC':
      switch (oldStatus) {
        case 'completed':  return 'done';
        case 'pending':    return 'pending';
        case 'in_progress': return 'in_progress';
        case 'failed':     return 'failed';
        default:           return oldStatus;
      }
    case 'Visa':
      switch (oldStatus) {
        case 'completed':  return 'approved';
        case 'failed':     return 'rejected';
        case 'pending':    return 'pending';
        case 'in_progress': return 'in_progress';
        default:           return oldStatus;
      }
    case 'Ticket':
      switch (oldStatus) {
        case 'pending':    return 'pending';
        case 'in_progress': return 'booked';
        case 'completed':  return 'arrived';
        case 'failed':     return 'pending';
        default:           return oldStatus;
      }
    case 'Arrival City':
    case 'Arrival':
      switch (oldStatus) {
        case 'pending':    return 'not_arrived';
        case 'in_progress': return 'in_transit';
        case 'completed':  return 'arrived';
        case 'failed':     return 'failed';
        default:           return oldStatus;
      }
    default:
      return oldStatus;
  }
}

export function mapProgressToSteps(progress: SelectionProgress): StepItem[] {
  const base = {
    notes: progress.notes || undefined,
    updated_at: progress.updated_at,
    updated_by: progress.updated_by
      ? { id: progress.updated_by, name: progress.updated_by_name ?? '' }
      : undefined,
  };

  const steps: StepItem[] = [];

  const addStep = (stepName: string, status: string, extras?: Partial<StepItem>) => {
    steps.push({
      id: `${progress.id}-${stepName}`,
      step_name: stepName,
      step_status: mapNewStatusToOld(stepName, status),
      ...base,
      ...extras,
    });
  };

  addStep('Medical', progress.medical_status, {
    medical_document_url: progress.medical_document?.file_url || undefined,
    document_url: progress.medical_document?.file_url || undefined,
    document_name: progress.medical_document?.file_name || undefined,
  });
  addStep('CoC', progress.coc_status, {
    coc_status: progress.coc_type || undefined,
    document_url: progress.coc_document?.file_url || undefined,
    document_name: progress.coc_document?.file_name || undefined,
  });
  addStep('Visa', progress.visa_status, {
    document_url: progress.visa_document?.file_url || undefined,
    document_name: progress.visa_document?.file_name || undefined,
  });
  addStep('Ticket', progress.ticket_status, {
    arrival_city: progress.arrival_city || undefined,
    document_url: progress.ticket_document?.file_url || undefined,
    document_name: progress.ticket_document?.file_name || undefined,
  });
  addStep('Arrival City', progress.arrival_status, {
    arrival_city: progress.arrival_city || undefined,
    document_url: progress.arrival_document?.file_url || undefined,
    document_name: progress.arrival_document?.file_name || undefined,
  });

  return steps;
}

export function stepToPayload(stepName: string, mappedStatus: string): Record<string, string> {
  switch (stepName) {
    case 'Medical':      return { medical_status: mappedStatus };
    case 'CoC':          return { coc_status: mappedStatus };
    case 'Visa':         return { visa_status: mappedStatus };
    case 'Ticket':       return { ticket_status: mappedStatus };
    case 'Arrival City':
    case 'Arrival':      return { arrival_status: mappedStatus };
    default:             return {};
  }
}

// ─── Per-card controller (hooks per selectionId) ───────────────────────────────
function ApplicantTableRowController({
  selectionId,
  candidateId,
  photoUrl,
  selectedByName,
  isSelected,
  onToggleSelect,
  onViewDetails,
  canEdit,
  filters,
}: {
  selectionId: string;
  candidateId: string;
  photoUrl?: string;
  selectedByName?: string;
  isSelected: boolean;
  onToggleSelect: () => void;
  onViewDetails: (d: DrawerApplicant) => void;
  canEdit: boolean;
  filters: TrackingFilters;
}) {
  const { data: candidate } = useCandidate(candidateId);
  const { data: selectionProgress } = useSelectionProgress(selectionId);
  const updateProgress = useUpdateProgress(selectionId);
  const updateCandidateStatus = useUpdateCandidateStatus(candidateId);
  const [savingSteps, setSavingSteps] = useState<Set<string>>(new Set());

  const steps: StepItem[] = useMemo(
    () => (selectionProgress ? mapProgressToSteps(selectionProgress) : []),
    [selectionProgress],
  );

  const handleUpdateStatus = useCallback(
    async (stepName: string, status: string, extras?: { destination_country?: string; arrival_date?: string }) => {
      setSavingSteps((prev) => new Set(prev).add(stepName));
      try {
        const mappedStatus = mapOldStatusToNew(stepName, status);
        const payload: Record<string, string> = {
          ...stepToPayload(stepName, mappedStatus),
          ...(extras?.destination_country ? { destination_country: extras.destination_country } : {}),
          ...(extras?.arrival_date ? { arrival_date: extras.arrival_date } : {}),
        };
        await updateProgress.mutateAsync(payload);
      } catch {
        // Error is already handled by mutation's onError toast
      } finally {
        setSavingSteps((prev) => {
          const next = new Set(prev);
          next.delete(stepName);
          return next;
        });
      }
    },
    [updateProgress],
  );

  const handleUpdateOverallStatus = useCallback(
    async (status: string) => {
      try {
        await updateCandidateStatus.mutateAsync(status);
      } catch {
        // Error handled by mutation toast
      }
    },
    [updateCandidateStatus],
  );

  if (!candidate) return null;

  // Apply filters
  const searchLower = filters.search.toLowerCase();
  if (
    searchLower &&
    !candidate.full_name.toLowerCase().includes(searchLower) &&
    !(candidate.passport_number ?? "").toLowerCase().includes(searchLower) &&
    !(candidate.nationality ?? "").toLowerCase().includes(searchLower) &&
    !(selectedByName ?? "").toLowerCase().includes(searchLower)
  ) {
    return null;
  }

  const getStep = (name: string) => steps.find((s) => s.step_name === name)?.step_status ?? "pending";
  if (filters.medical && getStep("Medical") !== filters.medical) return null;
  if (filters.coc     && getStep("CoC")     !== filters.coc)     return null;
  if (filters.visa    && getStep("Visa")    !== filters.visa)    return null;
  if (filters.ticket  && getStep("Ticket")  !== filters.ticket)  return null;
  if (filters.arrival && getStep("Arrival City") !== filters.arrival) return null;
  if (filters.status) {
    let derived: string;
    switch (candidate.status) {
      case "in_progress": derived = "processing"; break;
      case "completed":   derived = "completed"; break;
      case "rejected":    derived = "rejected"; break;
      default:            derived = "processing";
    }
    if (derived !== filters.status) return null;
  }
  if (filters.foreignAgent && (!selectedByName || !selectedByName.toLowerCase().includes(filters.foreignAgent.toLowerCase()))) return null;

  const rowData: ApplicantRowData = {
    selectionId,
    candidateId,
    candidateName:   candidate.full_name,
    passportNumber:  candidate.passport_number,
    age:             candidate.age,
    nationality:     candidate.nationality,
    photoUrl:        photoUrl,
    candidateStatus: candidate.status,
    steps,
    lastUpdatedAt:   selectionProgress?.updated_at ?? candidate.updated_at,
    canEdit,
    selectedByName,
  };

  return (
    <ApplicantTableRow
      data={rowData}
      isSelected={isSelected}
      onToggleSelect={onToggleSelect}
      onViewDetails={() =>
        onViewDetails({
          selectionId,
          candidateId,
          candidate,
          photoUrl,
          steps,
          candidateStatus: candidate.status,
          canEdit,
          savingSteps,
          onUpdateStatus: handleUpdateStatus,
          onUpdateOverallStatus: canEdit ? handleUpdateOverallStatus : undefined,
        })
      }
      onUpdateStatus={handleUpdateStatus}
      onUpdateOverallStatus={canEdit ? handleUpdateOverallStatus : undefined}
      savingSteps={savingSteps}
    />
  );
}

// ─── Stats card ────────────────────────────────────────────────────────────────
function StatCard({
  label,
  value,
  gradient,
  icon,
}: {
  label: string;
  value: number;
  gradient: string;
  icon: ReactNode;
}) {
  return (
    <Card className={`overflow-hidden border-white/20 bg-gradient-to-br ${gradient} shadow-soft`}>
      <CardContent className="flex items-center justify-between p-4">
        <div>
          <p className="text-xs font-medium text-current/70">{label}</p>
          <p className="mt-0.5 text-2xl font-bold">{value}</p>
        </div>
        <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-white/60 shadow-sm">
          {icon}
        </div>
      </CardContent>
    </Card>
  );
}

// ─── Main Page Component ───────────────────────────────────────────────────────
export function ProcessTrackingPage() {
  const { isEthiopianAgent } = useCurrentUser();
  const { data: selectionsData, isLoading, refetch, isFetching } = useMySelections();

  const [filters, setFilters] = useState<TrackingFilters>(DEFAULT_FILTERS);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [drawerApplicant, setDrawerApplicant] = useState<DrawerApplicant | null>(null);

  // All approved selections that have a candidate
  const allSelections = useMemo(
    () =>
      (selectionsData?.selections ?? []).filter(
        (s) => s.status === SelectionStatus.APPROVED && s.candidate,
      ),
    [selectionsData],
  );

  const toggleSelect = useCallback((id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }, []);

  const clearSelection = useCallback(() => setSelectedIds(new Set()), []);
  const batchUpdate = useBatchUpdateProgress();

  const handleBulkStepUpdate = useCallback(
    (stepName: string, status: string) => {
      const selectionIds = allSelections
        .filter((s) => selectedIds.has(s.candidate_id))
        .map((s) => s.id);

      if (selectionIds.length === 0) {
        toast.error('No valid selections found for update');
        return;
      }

      const mappedStatus = mapOldStatusToNew(stepName, status);
      const payload = stepToPayload(stepName, mappedStatus);

      batchUpdate.mutate(
        { selection_ids: selectionIds, ...payload },
        { onSuccess: () => clearSelection() },
      );
    },
    [selectedIds, allSelections, batchUpdate, clearSelection],
  );

  // Sort selections
  const sortedSelections = useMemo(() => {
    const arr = [...allSelections];
    switch (filters.sort) {
      case "name_desc":
        return arr.sort((a, b) => (b.candidate?.full_name ?? "").localeCompare(a.candidate?.full_name ?? ""));
      case "updated_desc":
        return arr.sort((a, b) => new Date(b.updated_at ?? 0).getTime() - new Date(a.updated_at ?? 0).getTime());
      case "updated_asc":
        return arr.sort((a, b) => new Date(a.updated_at ?? 0).getTime() - new Date(b.updated_at ?? 0).getTime());
      default: // name_asc
        return arr.sort((a, b) => (a.candidate?.full_name ?? "").localeCompare(b.candidate?.full_name ?? ""));
    }
  }, [allSelections, filters.sort]);

  return (
    <div className="flex flex-col min-h-0 -mx-6 -mt-2">
      {/* ── Page Header Actions ── */}
      <div className="px-6 pt-4 pb-3 flex items-center justify-end gap-2">
        <Button
          size="sm"
          variant="outline"
          className="h-8 rounded-xl text-xs gap-1.5"
          onClick={() => refetch()}
          disabled={isFetching}
        >
          {isFetching ? (
            <Loader2 className="h-3.5 w-3.5 animate-spin" />
          ) : (
            <RefreshCw className="h-3.5 w-3.5" />
          )}
          Refresh
        </Button>
        <Button size="sm" variant="outline" className="h-8 rounded-xl text-xs gap-1.5">
          <Download className="h-3.5 w-3.5" />
          Export
        </Button>
        <Button size="sm" variant="outline" className="h-8 rounded-xl text-xs gap-1.5">
          <Upload className="h-3.5 w-3.5" />
          Import
        </Button>
        {isEthiopianAgent && (
          <Button size="sm" className="h-8 rounded-xl text-xs gap-1.5 bg-gradient-to-r from-teal-600 to-sky-600 hover:from-teal-700 hover:to-sky-700 text-white">
            <UserPlus className="h-3.5 w-3.5" />
            New Applicant
          </Button>
        )}
      </div>

      {/* ── Stats row ── */}
      <div className="px-6 pb-5">
        <div className="grid grid-cols-3 gap-3">
          <StatCard
            label="Total Approved"
            value={allSelections.length}
            gradient="from-sky-500/20 to-sky-400/5 text-sky-950 dark:text-sky-100"
            icon={<Sparkles className="h-4 w-4" />}
          />
          <StatCard
            label="In Progress"
            value={allSelections.filter((s) => s.candidate?.status === "in_progress").length}
            gradient="from-amber-500/20 to-amber-400/5 text-amber-950 dark:text-amber-100"
            icon={<TimerReset className="h-4 w-4" />}
          />
          <StatCard
            label="Completed"
            value={allSelections.filter((s) => s.candidate?.status === "completed").length}
            gradient="from-emerald-500/20 to-emerald-400/5 text-emerald-950 dark:text-emerald-100"
            icon={<CheckCircle2 className="h-4 w-4" />}
          />
        </div>
      </div>

      {/* ── Search toolbar ── */}
      <SearchToolbar
        filters={filters}
        onChange={setFilters}
        totalCount={allSelections.length}
        filteredCount={sortedSelections.length}
        foreignAgentOptions={isEthiopianAgent ? [...new Set(allSelections.map((s) => s.selected_by_name).filter(Boolean))] as string[] : undefined}
      />

      {/* ── Data Table ── */}
      <div className="flex-1 overflow-x-auto px-6 py-5">
        {isLoading ? (
          <div className="flex items-center justify-center py-24">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : sortedSelections.length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-4 py-24 text-center">
            <div className="flex h-20 w-20 items-center justify-center rounded-full bg-muted">
              <Route className="h-10 w-10 text-muted-foreground" />
            </div>
            <div className="space-y-1.5">
              <h2 className="text-xl font-semibold">No applicants found</h2>
              <p className="max-w-sm text-sm text-muted-foreground">
                Once both parties approve a selection, it will appear here for tracking.
              </p>
            </div>
          </div>
        ) : (
          <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm overflow-hidden">
            <table className="w-full text-left text-sm text-slate-600 dark:text-slate-300">
              <thead className="bg-slate-50 dark:bg-slate-800/50 text-xs uppercase text-slate-500 dark:text-slate-400 border-b dark:border-slate-800">
                <tr>
                  <th className="py-3 px-4 w-10">
                    <Checkbox
                      checked={selectedIds.size > 0 && selectedIds.size === allSelections.length}
                      onCheckedChange={(checked) => {
                        if (checked) {
                          setSelectedIds(new Set(allSelections.map((s) => s.candidate_id)));
                        } else {
                          clearSelection();
                        }
                      }}
                      className="h-4 w-4 rounded-md"
                    />
                  </th>
                  <th className="py-3 px-3 font-semibold tracking-wider">Applicant</th>
                  <th className="py-3 px-3 font-semibold tracking-wider">Medical</th>
                  <th className="py-3 px-3 font-semibold tracking-wider">CoC</th>
                  <th className="py-3 px-3 font-semibold tracking-wider">Visa</th>
                  <th className="py-3 px-3 font-semibold tracking-wider">Ticket</th>
                  <th className="py-3 px-3 font-semibold tracking-wider">Overall Status</th>
                  <th className="py-3 px-4 text-right font-semibold tracking-wider">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 dark:divide-slate-800/50">
                {sortedSelections.map((selection) => (
                  <ApplicantTableRowController
                    key={selection.id}
                    selectionId={selection.id}
                    candidateId={selection.candidate_id}
                    photoUrl={selection.candidate?.photo_url}
                    selectedByName={selection.selected_by_name}
                    isSelected={selectedIds.has(selection.candidate_id)}
                    onToggleSelect={() => toggleSelect(selection.candidate_id)}
                    onViewDetails={setDrawerApplicant}
                    canEdit={isEthiopianAgent}
                    filters={filters}
                  />
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* ── Bulk toolbar ── */}
      <BulkToolbar
        selectedCount={selectedIds.size}
        onClear={clearSelection}
        onDelete={() => { toast.info(`Delete ${selectedIds.size} applicants`); clearSelection(); }}
        onExport={() => { toast.info("Exporting…"); clearSelection(); }}
        onPrint={() => { toast.info("Printing…"); clearSelection(); }}
        onAssignEmployee={() => { toast.info("Assign employee"); clearSelection(); }}
        onSendNotification={() => { toast.info("Sending notification…"); clearSelection(); }}
        onChangeMedical={(s) => handleBulkStepUpdate("Medical", s)}
        onChangeCoC={(s)     => handleBulkStepUpdate("CoC",     s)}
        onChangeVisa={(s)    => handleBulkStepUpdate("Visa",    s)}
        onChangeTicket={(s)  => handleBulkStepUpdate("Ticket",  s)}
        onChangeArrival={(s) => handleBulkStepUpdate("Arrival City", s)}
      />

      {/* ── Applicant Drawer ── */}
      <ApplicantDrawer
        applicant={drawerApplicant}
        onClose={() => setDrawerApplicant(null)}
      />
    </div>
  );
}
