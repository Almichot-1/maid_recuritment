"use client";

import * as React from "react";
import { Hash, Paperclip, Building2 } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { StatusBadge, TICKET_STATUS_MAP, ARRIVAL_STATUS_MAP, APPLICANT_STATUS_MAP } from "./status-badge";
import { StatusDropdown } from "./status-dropdown";
import { cn } from "@/lib/utils";
import type { StepItem } from "@/types";

function deriveApplicantStatus(candidateStatus: string): string {
  switch (candidateStatus) {
    case "in_progress": return "processing";
    case "completed":   return "completed";
    case "rejected":    return "rejected";
    default:            return "processing";
  }
}

function getStepStatus(steps: StepItem[], stepName: string): string {
  return steps.find((s) => s.step_name === stepName)?.step_status ?? "pending";
}

function getStepDoc(steps: StepItem[], stepName: string): string | undefined {
  return steps.find((s) => s.step_name === stepName)?.document_url;
}

export function formatApplicantName(name: string): string {
  if (!name) return '';
  return name
    .trim()
    .toLowerCase()
    .split(/\s+/)
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
}

export interface ApplicantRowData {
  selectionId: string;
  candidateId: string;
  candidateName: string;
  passportNumber?: string;
  age?: number;
  nationality?: string;
  phone?: string;
  photoUrl?: string;
  candidateStatus: string;
  steps: StepItem[];
  lastUpdatedAt: string;
  canEdit: boolean;
  selectedByName?: string;
}

interface ApplicantTableRowProps {
  data: ApplicantRowData;
  isSelected: boolean;
  onToggleSelect: () => void;
  onViewDetails: () => void;
  onUpdateStatus: (stepName: string, status: string, extras?: { destination_country?: string; arrival_date?: string }) => Promise<void>;
  onUpdateOverallStatus?: (status: string) => Promise<void>;
  savingSteps: Set<string>;
}

export function ApplicantTableRow({
  data,
  isSelected,
  onToggleSelect,
  onViewDetails,
  onUpdateStatus,
  onUpdateOverallStatus,
  savingSteps,
}: ApplicantTableRowProps) {
  const overallStatus = deriveApplicantStatus(data.candidateStatus);
  const medicalStatus = getStepStatus(data.steps, "Medical");
  const cocStatus     = getStepStatus(data.steps, "CoC");
  const visaStatus    = getStepStatus(data.steps, "Visa");
  const ticketStatus  = getStepStatus(data.steps, "Ticket");
  const arrivalStatus = getStepStatus(data.steps, "Arrival City");

  const normalizedName = formatApplicantName(data.candidateName);
  const initials = normalizedName
    .split(" ")
    .map((w) => w[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();

  const StepCell = ({ step, status, options }: { step: string; status: string; options?: Record<string, { label: string; className: string }> }) => {
    const docUrl = getStepDoc(data.steps, step);
    return (
      <td className="py-2.5 px-3" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center gap-1.5">
          <StatusDropdown
            value={status}
            disabled={!data.canEdit}
            isSaving={savingSteps.has(step)}
            onSelect={(v) => onUpdateStatus(step, v)}
            options={options}
          />
          {docUrl && (
            <a href={docUrl} target="_blank" rel="noopener noreferrer" title="View document" onClick={(e) => e.stopPropagation()}>
              <div className="flex h-5 w-5 items-center justify-center rounded-full bg-emerald-100 text-emerald-700 hover:bg-emerald-200 transition-colors">
                <Paperclip className="h-2.5 w-2.5" />
              </div>
            </a>
          )}
        </div>
      </td>
    );
  };

  return (
    <tr 
      onClick={onViewDetails}
      className={cn(
      "border-b border-border/50 hover:bg-slate-50/80 dark:hover:bg-slate-800/50 transition-colors text-sm cursor-pointer",
      isSelected && "bg-primary/[0.02] dark:bg-primary/[0.05]"
    )}>
      {/* Checkbox */}
      <td className="py-3 px-4 w-10" onClick={(e) => e.stopPropagation()}>
        <Checkbox
          checked={isSelected}
          onCheckedChange={onToggleSelect}
          className="h-4 w-4 rounded-md"
          aria-label={`Select ${data.candidateName}`}
        />
      </td>

      {/* Applicant Info */}
      <td className="py-3 px-3">
        <div className="flex items-center gap-3">
          <Avatar className="h-9 w-9 rounded-xl ring-1 ring-border shadow-sm">
            <AvatarImage src={data.photoUrl} alt={normalizedName} className="object-cover" />
            <AvatarFallback className="rounded-xl bg-gradient-to-br from-teal-500 to-sky-500 text-white text-[10px] font-bold">
              {initials}
            </AvatarFallback>
          </Avatar>
          <div className="flex flex-col min-w-0">
            <span className="font-bold text-slate-900 dark:text-slate-100 leading-tight truncate">{normalizedName}</span>
            <div className="flex items-center gap-2 mt-0.5 text-[10px] text-muted-foreground">
              {data.passportNumber && <span className="flex items-center gap-0.5"><Hash className="h-3 w-3" /> {data.passportNumber}</span>}
              {data.selectedByName && <span className="flex items-center gap-0.5"><Building2 className="h-3 w-3" /> {data.selectedByName}</span>}
            </div>
          </div>
        </div>
      </td>

      {/* Stages */}
      <StepCell step="Medical" status={medicalStatus} />
      <StepCell step="CoC" status={cocStatus} />
      <StepCell step="Visa" status={visaStatus} />
      <StepCell step="Ticket" status={ticketStatus} options={TICKET_STATUS_MAP} />
      <StepCell step="Arrival City" status={arrivalStatus} options={ARRIVAL_STATUS_MAP} />

      {/* Overall Status */}
      <td className="py-2.5 px-3" onClick={(e) => e.stopPropagation()}>
        {data.canEdit && onUpdateOverallStatus ? (
          <StatusDropdown
            value={overallStatus}
            onSelect={(v) => onUpdateOverallStatus(v)}
            options={APPLICANT_STATUS_MAP}
          />
        ) : (
          <StatusBadge status={overallStatus} type="applicant" />
        )}
      </td>

      {/* Actions */}
      <td className="py-3 px-4 text-right" onClick={(e) => e.stopPropagation()}>
        <Button
          onClick={onViewDetails}
          size="sm"
          className="h-8 rounded-lg px-3 text-xs font-semibold bg-blue-50 text-blue-700 hover:bg-blue-100 dark:bg-blue-900/30 dark:text-blue-400 dark:hover:bg-blue-900/50 shadow-none border border-blue-200 dark:border-blue-800/50"
        >
          View Details
        </Button>
      </td>
    </tr>
  );
}
