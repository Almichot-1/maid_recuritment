"use client";

import * as React from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Clock, Search, ArrowRight, Stethoscope, FileText, Building2, Ticket, MapPin, GitCommitHorizontal } from "lucide-react";
import { Input } from "@/components/ui/input";
import { formatDistanceToNow, format } from "date-fns";
import { cn } from "@/lib/utils";
import type { StepItem } from "@/types";

const STEP_CONFIG: Record<string, { label: string; icon: React.ElementType; color: string }> = {
  Medical:       { label: "Medical",  icon: Stethoscope, color: "text-emerald-600 bg-emerald-50 border-emerald-200" },
  CoC:           { label: "CoC",      icon: FileText,    color: "text-violet-600 bg-violet-50 border-violet-200" },
  Visa:          { label: "Visa",     icon: Building2,   color: "text-sky-600 bg-sky-50 border-sky-200" },
  Ticket:        { label: "Ticket",   icon: Ticket,      color: "text-amber-600 bg-amber-50 border-amber-200" },
  "Arrival City":{ label: "Arrival",  icon: MapPin,      color: "text-teal-600 bg-teal-50 border-teal-200" },
};

const STATUS_LABELS: Record<string, string> = {
  pending:     "Pending",
  in_progress: "In Progress",
  completed:   "Completed",
  failed:      "Failed",
};

interface TimelineTabProps {
  steps: StepItem[];
}

export function TimelineTab({ steps }: TimelineTabProps) {
  const [search, setSearch] = React.useState("");

  const events = React.useMemo(() => {
    return [...steps]
      .sort((a, b) => new Date(b.updated_at ?? 0).getTime() - new Date(a.updated_at ?? 0).getTime())
      .filter((s) => {
        if (!search) return true;
        const config = STEP_CONFIG[s.step_name];
        return (
          (config?.label ?? s.step_name).toLowerCase().includes(search.toLowerCase()) ||
          (s.updated_by?.name ?? "").toLowerCase().includes(search.toLowerCase()) ||
          s.step_status.toLowerCase().includes(search.toLowerCase())
        );
      });
  }, [steps, search]);

  return (
    <div className="flex flex-col gap-5">
      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground pointer-events-none" />
        <Input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search timeline…"
          className="pl-9 h-9 rounded-full text-sm"
        />
      </div>

      {events.length === 0 ? (
        <div className="flex flex-col items-center justify-center gap-3 py-16 text-center">
          <div className="h-12 w-12 rounded-full bg-muted flex items-center justify-center">
            <GitCommitHorizontal className="h-5 w-5 text-muted-foreground" />
          </div>
          <p className="text-sm text-muted-foreground">No activity recorded yet</p>
        </div>
      ) : (
        <div className="relative">
          {/* Vertical line */}
          <div className="absolute left-[19px] top-2 bottom-2 w-px bg-border/70" />

          <div className="flex flex-col gap-0">
            <AnimatePresence>
              {events.map((step, i) => {
                const config = STEP_CONFIG[step.step_name] ?? { label: step.step_name, icon: GitCommitHorizontal, color: "text-slate-600 bg-slate-50 border-slate-200" };
                const Icon = config.icon;
                const updatedBy = step.updated_by?.name ?? "System";
                const initials = updatedBy.split(" ").map((w: string) => w[0]).join("").slice(0, 2).toUpperCase();

                let relativeTime = "—";
                let fullDate = "—";
                try {
                  relativeTime = formatDistanceToNow(new Date(step.updated_at ?? 0), { addSuffix: true });
                  fullDate = format(new Date(step.updated_at ?? 0), "MMM d, yyyy 'at' h:mm a");
                } catch { /* noop */ }

                return (
                  <motion.div
                    key={step.id}
                    initial={{ opacity: 0, x: -12 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ delay: i * 0.045, duration: 0.2 }}
                    className="relative flex gap-4 pb-7"
                  >
                    {/* Icon */}
                    <div
                      className={cn(
                        "relative z-10 flex h-10 w-10 shrink-0 items-center justify-center rounded-full border-2 shadow-sm bg-white",
                        step.step_status === "completed"
                          ? "border-emerald-300"
                          : step.step_status === "failed"
                          ? "border-rose-300"
                          : "border-border",
                      )}
                    >
                      <Icon
                        className={cn(
                          "h-4 w-4",
                          step.step_status === "completed"
                            ? "text-emerald-600"
                            : step.step_status === "failed"
                            ? "text-rose-500"
                            : "text-muted-foreground",
                        )}
                      />
                    </div>

                    {/* Content */}
                    <div className="flex-1 min-w-0 pt-2">
                      <div className="flex flex-wrap items-center justify-between gap-x-4 gap-y-1">
                        <div className="flex items-center gap-2 flex-wrap text-xs">
                          {/* Employee avatar */}
                          <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-teal-500 to-sky-500 text-[9px] font-bold text-white">
                            {initials}
                          </span>
                          <span className="font-semibold text-foreground">{updatedBy}</span>
                          <span className="text-muted-foreground">updated</span>
                          <span
                            className={cn(
                              "inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-[10px] font-semibold",
                              config.color,
                            )}
                          >
                            <Icon className="h-3 w-3" />
                            {config.label}
                          </span>
                          <ArrowRight className="h-3 w-3 text-muted-foreground/50" />
                          <span
                            className={cn(
                              "inline-flex rounded-full border px-2 py-0.5 text-[10px] font-semibold",
                              step.step_status === "completed"
                                ? "bg-emerald-50 text-emerald-700 border-emerald-200"
                                : step.step_status === "failed"
                                ? "bg-rose-50 text-rose-700 border-rose-200"
                                : step.step_status === "in_progress"
                                ? "bg-sky-50 text-sky-700 border-sky-200"
                                : "bg-slate-100 text-slate-600 border-slate-200",
                            )}
                          >
                            {STATUS_LABELS[step.step_status] ?? step.step_status}
                          </span>
                        </div>
                        <span
                          className="text-[10px] text-muted-foreground whitespace-nowrap"
                          title={fullDate}
                        >
                          <Clock className="inline h-3 w-3 mr-0.5 -mt-0.5" />
                          {relativeTime}
                        </span>
                      </div>

                      {/* Notes / arrival city */}
                      {(step.notes || step.arrival_city) && (
                        <div className="mt-2 rounded-lg bg-muted/40 px-3 py-2 text-[11px] text-muted-foreground">
                          {step.notes && <p>&ldquo;{step.notes}&rdquo;</p>}
                          {step.arrival_city && (
                            <p className="mt-0.5">
                              City: <span className="font-medium text-foreground">{step.arrival_city}</span>
                            </p>
                          )}
                        </div>
                      )}
                    </div>
                  </motion.div>
                );
              })}
            </AnimatePresence>
          </div>
        </div>
      )}
    </div>
  );
}
