"use client";

import { cn } from "@/lib/utils";

// The backend strictly accepts these 4 values for ALL steps.
export type StepStatus = "pending" | "in_progress" | "completed" | "failed";

export const STEP_STATUS_MAP: Record<StepStatus, { label: string; className: string }> = {
  pending:     { label: "Pending",     className: "bg-slate-100 text-slate-800 border-slate-300 dark:bg-slate-500/10 dark:text-slate-300 dark:border-slate-500/20" },
  in_progress: { label: "In Progress", className: "bg-sky-100 text-sky-800 border-sky-300 dark:bg-sky-500/10 dark:text-sky-400 dark:border-sky-500/20" },
  completed:   { label: "Completed",   className: "bg-emerald-100 text-emerald-800 border-emerald-300 dark:bg-emerald-500/10 dark:text-emerald-400 dark:border-emerald-500/20" },
  failed:      { label: "Failed",      className: "bg-rose-100 text-rose-800 border-rose-300 dark:bg-rose-500/10 dark:text-rose-400 dark:border-rose-500/20" },
};

export const TICKET_STATUS_MAP: Record<string, { label: string; className: string }> = {
  pending:     { label: "Pending",     className: "bg-slate-100 text-slate-800 border-slate-300 dark:bg-slate-500/10 dark:text-slate-300 dark:border-slate-500/20" },
  in_progress: { label: "Booked",      className: "bg-sky-100 text-sky-800 border-sky-300 dark:bg-sky-500/10 dark:text-sky-400 dark:border-sky-500/20" },
  completed:   { label: "Arrived",     className: "bg-emerald-100 text-emerald-800 border-emerald-300 dark:bg-emerald-500/10 dark:text-emerald-400 dark:border-emerald-500/20" },
  failed:      { label: "Failed",      className: "bg-rose-100 text-rose-800 border-rose-300 dark:bg-rose-500/10 dark:text-rose-400 dark:border-rose-500/20" },
};

export const ARRIVAL_STATUS_MAP: Record<string, { label: string; className: string }> = {
  pending:     { label: "Not Arrived", className: "bg-slate-100 text-slate-800 border-slate-300 dark:bg-slate-500/10 dark:text-slate-300 dark:border-slate-500/20" },
  in_progress: { label: "In Transit",  className: "bg-sky-100 text-sky-800 border-sky-300 dark:bg-sky-500/10 dark:text-sky-400 dark:border-sky-500/20" },
  completed:   { label: "Arrived",     className: "bg-emerald-100 text-emerald-800 border-emerald-300 dark:bg-emerald-500/10 dark:text-emerald-400 dark:border-emerald-500/20" },
  failed:      { label: "Failed",      className: "bg-rose-100 text-rose-800 border-rose-300 dark:bg-rose-500/10 dark:text-rose-400 dark:border-rose-500/20" },
};

// ─── Applicant (Overall) Status ───────────────────────────────────────────────
export type ApplicantStatus = "processing" | "ready" | "completed" | "rejected";

export const APPLICANT_STATUS_MAP: Record<ApplicantStatus, { label: string; className: string }> = {
  processing: { label: "Processing", className: "bg-amber-100 text-amber-900 border-amber-300 dark:bg-amber-500/10 dark:text-amber-400 dark:border-amber-500/20" },
  ready:      { label: "Ready",      className: "bg-sky-100 text-sky-800 border-sky-300 dark:bg-sky-500/10 dark:text-sky-400 dark:border-sky-500/20" },
  completed:  { label: "Completed",  className: "bg-emerald-100 text-emerald-800 border-emerald-300 dark:bg-emerald-500/10 dark:text-emerald-400 dark:border-emerald-500/20" },
  rejected:   { label: "Rejected",   className: "bg-rose-100 text-rose-800 border-rose-300 dark:bg-rose-500/10 dark:text-rose-400 dark:border-rose-500/20" },
};

// ─── StatusBadge Component ────────────────────────────────────────────────────
interface StatusBadgeProps {
  status: string;
  type?: "applicant" | "step";
  className?: string;
}

export function StatusBadge({ status, type = "step", className }: StatusBadgeProps) {
  let info = { label: status, className: "bg-slate-100 text-slate-800 border-slate-300 dark:bg-slate-500/10 dark:text-slate-300 dark:border-slate-500/20" };

  if (type === "applicant") {
    info = APPLICANT_STATUS_MAP[status as ApplicantStatus] ?? info;
  } else {
    info = STEP_STATUS_MAP[status as StepStatus] ?? info;
  }

  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full border px-2.5 py-0.5 text-[11px] font-bold tracking-wide select-none transition-colors",
        info.className,
        className,
      )}
    >
      {info.label}
    </span>
  );
}
