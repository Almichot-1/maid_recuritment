"use client";

import * as React from "react";
import { motion } from "framer-motion";
import { Eye, Clock, Hash, Globe, User, Phone, Paperclip, Building2 } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { StatusBadge, TICKET_STATUS_MAP, ARRIVAL_STATUS_MAP, APPLICANT_STATUS_MAP } from "./status-badge";
import { StatusDropdown } from "./status-dropdown";
import { cn } from "@/lib/utils";
import { formatDistanceToNow } from "date-fns";
import type { StepItem } from "@/types";

// Derive overall applicant status from backend candidate status
function deriveApplicantStatus(candidateStatus: string): string {
  switch (candidateStatus) {
    case "in_progress": return "processing";
    case "completed":   return "completed";
    case "rejected":    return "rejected";
    default:            return "processing";
  }
}

// Get step status value from list of steps
function getStepStatus(steps: StepItem[], stepName: string): string {
  return steps.find((s) => s.step_name === stepName)?.step_status ?? "pending";
}

function getStepDoc(steps: StepItem[], stepName: string): string | undefined {
  return steps.find((s) => s.step_name === stepName)?.document_url;
}

export interface ApplicantCardData {
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

interface ApplicantCardProps {
  data: ApplicantCardData;
  isSelected: boolean;
  onToggleSelect: () => void;
  onViewDetails: () => void;
  onUpdateStatus: (stepName: string, status: string, extras?: { destination_country?: string; arrival_date?: string }) => Promise<void>;
  onUpdateOverallStatus?: (status: string) => Promise<void>;
  savingSteps: Set<string>;
  index: number;
}

export function ApplicantCard({
  data,
  isSelected,
  onToggleSelect,
  onViewDetails,
  onUpdateStatus,
  onUpdateOverallStatus,
  savingSteps,
  index,
}: ApplicantCardProps) {
  const overallStatus = deriveApplicantStatus(data.candidateStatus);

  const medicalStatus     = getStepStatus(data.steps, "Medical");
  const cocStatus         = getStepStatus(data.steps, "CoC");
  const visaStatus        = getStepStatus(data.steps, "Visa");
  const ticketStatus      = getStepStatus(data.steps, "Ticket");
  const arrivalStatus     = getStepStatus(data.steps, "Arrival City");

  const initials = data.candidateName
    .split(" ")
    .map((w) => w[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();

  const lastUpdated = React.useMemo(() => {
    try {
      return formatDistanceToNow(new Date(data.lastUpdatedAt), { addSuffix: true });
    } catch {
      return "—";
    }
  }, [data.lastUpdatedAt]);

  return (
    <motion.div
      initial={{ opacity: 0, y: 16 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.22, delay: Math.min(index * 0.04, 0.4), ease: "easeOut" }}
      whileHover={{ y: -2, boxShadow: "0 10px 40px -12px rgba(15,23,42,0.13)" }}
      className={cn(
        "group relative rounded-[18px] border bg-white p-5 transition-shadow duration-200",
        "shadow-[0_2px_12px_-4px_rgba(15,23,42,0.07)]",
        isSelected
          ? "border-primary/40 ring-1 ring-primary/20 bg-primary/[0.015]"
          : "border-border/60 hover:border-border/90",
      )}
    >
      <div className="flex items-start gap-4 min-w-0">
        {/* ── LEFT ── */}
        <div className="flex items-start gap-3 min-w-0 flex-1">
          {/* Checkbox */}
          <div className="pt-1 shrink-0">
            <Checkbox
              checked={isSelected}
              onCheckedChange={onToggleSelect}
              className="h-4 w-4 rounded-md"
              aria-label={`Select ${data.candidateName}`}
            />
          </div>

          {/* Avatar */}
          <div className="shrink-0">
            <Avatar className="h-12 w-12 rounded-2xl ring-2 ring-white shadow">
              <AvatarImage src={data.photoUrl} alt={data.candidateName} className="object-cover" />
              <AvatarFallback className="rounded-2xl bg-gradient-to-br from-teal-500 to-sky-500 text-white text-sm font-bold">
                {initials}
              </AvatarFallback>
            </Avatar>
          </div>

          {/* Info */}
          <div className="min-w-0 flex-1">
            <div className="flex flex-wrap items-center gap-2 mb-1">
              {data.canEdit && onUpdateOverallStatus ? (
                <StatusDropdown
                  value={overallStatus}
                  onSelect={(v) => onUpdateOverallStatus(v)}
                  options={APPLICANT_STATUS_MAP}
                />
              ) : (
                <StatusBadge status={overallStatus} type="applicant" />
              )}
            </div>
            <h3 className="text-sm font-bold text-foreground leading-snug truncate">
              {data.candidateName}
            </h3>
            <div className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-0.5">
              {data.selectedByName && (
                <span className="inline-flex items-center gap-1 text-[11px] text-muted-foreground">
                  <Building2 className="h-3 w-3" />
                  {data.selectedByName}
                </span>
              )}
              {data.passportNumber && (
                <span className="inline-flex items-center gap-1 text-[11px] text-muted-foreground">
                  <Hash className="h-3 w-3" />
                  {data.passportNumber}
                </span>
              )}
              {data.age != null && (
                <span className="inline-flex items-center gap-1 text-[11px] text-muted-foreground">
                  <User className="h-3 w-3" />
                  {data.age} yrs
                </span>
              )}
              {data.nationality && (
                <span className="inline-flex items-center gap-1 text-[11px] text-muted-foreground">
                  <Globe className="h-3 w-3" />
                  {data.nationality}
                </span>
              )}
              {data.phone && (
                <span className="inline-flex items-center gap-1 text-[11px] text-muted-foreground">
                  <Phone className="h-3 w-3" />
                  {data.phone}
                </span>
              )}
            </div>
          </div>
        </div>

        {/* ── CENTER: Statuses (desktop) ── */}
        <div className="hidden xl:flex items-start gap-6 shrink-0 pt-1">
          {(["Medical", "CoC", "Visa", "Ticket", "Arrival City"] as const).map((step) => {
            const statusMap: Record<string, string> = {
              Medical: medicalStatus,
              CoC:     cocStatus,
              Visa:    visaStatus,
              Ticket:  ticketStatus,
              "Arrival City": arrivalStatus,
            };
            const stepOptions: Record<string, Record<string, { label: string; className: string }>> = {
              Ticket: TICKET_STATUS_MAP,
              "Arrival City": ARRIVAL_STATUS_MAP,
            };
            const stepDocUrl = getStepDoc(data.steps, step);
            const stepLabel = step === "Arrival City" ? "Arrival" : step;
            return (
              <div key={step} className="flex flex-col items-start gap-1.5 min-w-[78px]">
                <span className="text-[9px] font-semibold uppercase tracking-widest text-muted-foreground">{stepLabel}</span>
                <div className="flex items-center gap-1">
                  <StatusDropdown
                    value={statusMap[step]}
                    disabled={!data.canEdit}
                    isSaving={savingSteps.has(step)}
                    onSelect={(v) => onUpdateStatus(step, v)}
                    options={stepOptions[step]}
                  />
                  {stepDocUrl && (
                    <a href={stepDocUrl} target="_blank" rel="noopener noreferrer" title="View uploaded document">
                      <div className="flex h-5 w-5 items-center justify-center rounded-full bg-emerald-100 text-emerald-700">
                        <Paperclip className="h-2.5 w-2.5" />
                      </div>
                    </a>
                  )}
                </div>
              </div>
            );
          })}
        </div>

        {/* ── RIGHT: View + Meta ── */}
        <div className="flex flex-col items-end gap-2 shrink-0 ml-2 pt-0.5">
          <Button
            onClick={onViewDetails}
            size="sm"
            className="h-8 rounded-xl px-4 text-xs font-semibold bg-gradient-to-r from-teal-600 to-sky-600 hover:from-teal-700 hover:to-sky-700 text-white shadow-sm"
          >
            <Eye className="h-3.5 w-3.5 mr-1.5" />
            View Details
          </Button>
          <span className="inline-flex items-center gap-1 text-[10px] text-muted-foreground whitespace-nowrap">
            <Clock className="h-3 w-3" />
            {lastUpdated}
          </span>
        </div>
      </div>

      {/* ── Mobile status row ── */}
      <div className="mt-4 flex flex-wrap items-center gap-4 xl:hidden border-t border-border/40 pt-3">
        {(["Medical", "CoC", "Visa", "Ticket", "Arrival City"] as const).map((step) => {
          const statusValues: Record<string, string> = {
            Medical: medicalStatus,
            CoC:     cocStatus,
            Visa:    visaStatus,
            Ticket:  ticketStatus,
            "Arrival City": arrivalStatus,
          };
          const stepOptions: Record<string, Record<string, { label: string; className: string }>> = {
            Ticket: TICKET_STATUS_MAP,
            "Arrival City": ARRIVAL_STATUS_MAP,
          };
          const stepDocUrl = getStepDoc(data.steps, step);
          return (
            <div key={step} className="flex flex-col gap-1">
              <span className="text-[9px] font-semibold uppercase tracking-widest text-muted-foreground">{step}</span>
              <div className="flex items-center gap-1">
                <StatusDropdown
                  value={statusValues[step]}
                  disabled={!data.canEdit}
                  isSaving={savingSteps.has(step)}
                  onSelect={(v) => onUpdateStatus(step, v)}
                  options={stepOptions[step]}
                />
                {stepDocUrl && (
                  <a href={stepDocUrl} target="_blank" rel="noopener noreferrer" title="View uploaded document">
                    <div className="flex h-5 w-5 items-center justify-center rounded-full bg-emerald-100 text-emerald-700">
                      <Paperclip className="h-2.5 w-2.5" />
                    </div>
                  </a>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </motion.div>
  );
}
