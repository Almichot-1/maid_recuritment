"use client";

import * as React from "react";
import {
  User, Hash, Globe, Calendar, Flag,
  Briefcase, GraduationCap, Heart, Baby, Languages,
  Stethoscope, FileText, Building2, Ticket, MapPin, CheckCircle2, XCircle, Clock,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { format } from "date-fns";
import type { StepItem } from "@/types";
import type { Candidate } from "@/types";

const STEP_STATUS_CONFIG: Record<string, { icon: React.ElementType; label: string; colorClass: string }> = {
  completed:   { icon: CheckCircle2, label: "Completed",   colorClass: "text-emerald-600" },
  in_progress: { icon: Clock,        label: "In Progress", colorClass: "text-sky-600" },
  failed:      { icon: XCircle,      label: "Failed",      colorClass: "text-rose-500" },
  pending:     { icon: Clock,        label: "Pending",     colorClass: "text-muted-foreground" },
};

function InfoRow({ icon: Icon, label, value }: { icon: React.ElementType; label: string; value?: string | number | null }) {
  if (!value && value !== 0) return null;
  return (
    <div className="flex items-start gap-3 py-2.5 border-b border-border/50 last:border-0">
      <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-muted/60">
        <Icon className="h-3.5 w-3.5 text-muted-foreground" />
      </div>
      <div className="min-w-0 flex-1">
        <p className="text-[10px] font-semibold uppercase tracking-widest text-muted-foreground mb-0.5">{label}</p>
        <p className="text-sm font-medium text-foreground">{String(value)}</p>
      </div>
    </div>
  );
}

function StepSummaryCard({
  stepName,
  icon: Icon,
  steps,
  colorBg,
}: {
  stepName: string;
  icon: React.ElementType;
  steps: StepItem[];
  colorBg: string;
}) {
  const step = steps.find((s) => s.step_name === stepName);
  const status = step?.step_status ?? "pending";
  const config = STEP_STATUS_CONFIG[status] ?? STEP_STATUS_CONFIG.pending;
  const StatusIcon = config.icon;

  return (
    <div className={cn("flex items-center gap-3 rounded-2xl border p-3.5", colorBg)}>
      <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-white/60 shadow-sm">
        <Icon className="h-4 w-4" />
      </div>
      <div className="min-w-0 flex-1">
        <p className="text-[10px] font-semibold uppercase tracking-widest opacity-70 mb-0.5">{stepName}</p>
        <div className="flex items-center gap-1.5">
          <StatusIcon className={cn("h-3.5 w-3.5", config.colorClass)} />
          <span className="text-sm font-semibold">{config.label}</span>
        </div>
      </div>
    </div>
  );
}

interface OverviewTabProps {
  candidate: Candidate;
  steps: StepItem[];
}

export function OverviewTab({ candidate, steps }: OverviewTabProps) {
  const dob = candidate.date_of_birth
    ? (() => { try { return format(new Date(candidate.date_of_birth!), "MMM d, yyyy"); } catch { return candidate.date_of_birth; } })()
    : null;

  return (
    <div className="flex flex-col gap-6">
      {/* Progress summary */}
      <div>
        <h4 className="text-xs font-semibold uppercase tracking-widest text-muted-foreground mb-3">Process Status</h4>
        <div className="grid grid-cols-2 gap-2">
          <StepSummaryCard stepName="Medical" icon={Stethoscope} steps={steps} colorBg="border-emerald-100 bg-emerald-50/50 text-emerald-900" />
          <StepSummaryCard stepName="CoC"     icon={FileText}    steps={steps} colorBg="border-violet-100 bg-violet-50/50 text-violet-900" />
          <StepSummaryCard stepName="Visa"    icon={Building2}   steps={steps} colorBg="border-sky-100 bg-sky-50/50 text-sky-900" />
          <StepSummaryCard stepName="Ticket"  icon={Ticket}      steps={steps} colorBg="border-amber-100 bg-amber-50/50 text-amber-900" />
        </div>
      </div>

      {/* Personal info */}
      <div>
        <h4 className="text-xs font-semibold uppercase tracking-widest text-muted-foreground mb-3">Personal Information</h4>
        <div className="rounded-2xl border border-border/60 bg-white px-1 divide-y-0">
          <InfoRow icon={User}         label="Full Name"        value={candidate.full_name} />
          <InfoRow icon={Hash}         label="Passport Number"  value={candidate.passport_number} />
          <InfoRow icon={Globe}        label="Nationality"      value={candidate.nationality} />
          <InfoRow icon={Calendar}     label="Date of Birth"    value={dob} />
          <InfoRow icon={User}         label="Age"              value={candidate.age ? `${candidate.age} years` : null} />
          <InfoRow icon={Flag}         label="Place of Birth"   value={candidate.place_of_birth} />
          <InfoRow icon={Heart}        label="Marital Status"   value={candidate.marital_status} />
          <InfoRow icon={Baby}         label="Children"         value={candidate.children_count != null ? String(candidate.children_count) : null} />
          <InfoRow icon={GraduationCap}label="Education"        value={candidate.education_level} />
          <InfoRow icon={Briefcase}    label="Experience"       value={candidate.experience_years ? `${candidate.experience_years} years` : null} />
          <InfoRow icon={Globe}        label="Country of Exp."  value={candidate.country_of_experience} />
          <InfoRow icon={Languages}    label="Languages"        value={candidate.languages?.map((l) => l.language).join(", ")} />
        </div>
      </div>

      {/* Arrival city */}
      {(() => {
        const arrivalStep = steps.find((s) => s.step_name === "Arrival City");
        if (!arrivalStep?.arrival_city) return null;
        return (
          <div>
            <h4 className="text-xs font-semibold uppercase tracking-widest text-muted-foreground mb-3">Arrival Details</h4>
            <div className="rounded-2xl border border-border/60 bg-white px-1">
              <InfoRow icon={MapPin} label="Arrival City" value={arrivalStep.arrival_city} />
            </div>
          </div>
        );
      })()}
    </div>
  );
}
