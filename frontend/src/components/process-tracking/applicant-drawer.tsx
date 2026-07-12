"use client";

import * as React from "react";
import { motion, AnimatePresence } from "framer-motion";
import { X, Printer, Download, Globe, Hash, User } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { StatusBadge, APPLICANT_STATUS_MAP } from "./status-badge";
import { OverviewTab } from "./drawer-tabs/overview-tab";
import { TimelineTab } from "./drawer-tabs/timeline-tab";
import { DocumentsTab } from "./drawer-tabs/documents-tab";
import { NotesTab } from "./drawer-tabs/notes-tab";
import { ProcessStepsTab } from "./drawer-tabs/process-steps-tab";
import { StatusDropdown } from "./status-dropdown";
import { cn } from "@/lib/utils";
import type { Candidate, StepItem } from "@/types";

export interface DrawerApplicant {
  selectionId: string;
  candidateId: string;
  candidate: Candidate;
  photoUrl?: string;
  steps: StepItem[];
  candidateStatus: string;
  canEdit: boolean;
  savingSteps: Set<string>;
  onUpdateStatus: (stepName: string, status: string, extras?: { destination_country?: string; arrival_date?: string }) => void | Promise<void>;
  onUpdateOverallStatus?: (status: string) => Promise<void>;
}

interface ApplicantDrawerProps {
  applicant: DrawerApplicant | null;
  onClose: () => void;
}

function deriveApplicantStatus(candidateStatus: string): string {
  switch (candidateStatus) {
    case "in_progress": return "processing";
    case "completed":   return "completed";
    case "rejected":    return "rejected";
    default:            return "processing";
  }
}

export function ApplicantDrawer({ applicant, onClose }: ApplicantDrawerProps) {
  // Close on Escape
  React.useEffect(() => {
    const onKey = (e: KeyboardEvent) => { if (e.key === "Escape") onClose(); };
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, [onClose]);

  return (
    <AnimatePresence>
      {applicant && (
        <>
          {/* Backdrop */}
          <motion.div
            key="drawer-backdrop"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.2 }}
            className="fixed inset-0 z-40 bg-black/30 backdrop-blur-sm"
            onClick={onClose}
          />

          {/* Drawer panel */}
          <motion.div
            key="drawer-panel"
            initial={{ x: "100%" }}
            animate={{ x: 0 }}
            exit={{ x: "100%" }}
            transition={{ type: "spring", damping: 30, stiffness: 300 }}
            className="fixed right-0 top-0 bottom-0 z-50 flex w-full max-w-[680px] flex-col bg-background shadow-2xl"
          >
            {/* ── Header ── */}
            <div className="flex items-start gap-4 border-b border-border/60 px-6 py-5 bg-gradient-to-b from-background to-muted/20">
              {/* Close */}
              <button
                onClick={onClose}
                className="mt-0.5 rounded-full p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground transition-colors shrink-0"
                aria-label="Close drawer"
              >
                <X className="h-4.5 w-4.5" />
              </button>

              {/* Avatar */}
              <Avatar className="h-14 w-14 rounded-2xl ring-2 ring-border shadow-sm shrink-0">
                <AvatarImage src={applicant.photoUrl} alt={applicant.candidate.full_name} className="object-cover" />
                <AvatarFallback className="rounded-2xl bg-gradient-to-br from-teal-500 to-sky-500 text-white text-lg font-bold">
                  {applicant.candidate.full_name
                    .split(" ")
                    .map((w) => w[0])
                    .join("")
                    .slice(0, 2)
                    .toUpperCase()}
                </AvatarFallback>
              </Avatar>

              {/* Identity */}
              <div className="min-w-0 flex-1">
                <div className="flex flex-wrap items-center gap-2 mb-1">
                  {applicant.canEdit ? (
                    <StatusDropdown
                      value={deriveApplicantStatus(applicant.candidateStatus)}
                      onSelect={(v) => applicant.onUpdateOverallStatus?.(v)}
                      options={APPLICANT_STATUS_MAP}
                    />
                  ) : (
                    <StatusBadge
                      status={deriveApplicantStatus(applicant.candidateStatus)}
                      type="applicant"
                    />
                  )}
                </div>
                <h2 className="text-lg font-bold text-foreground leading-snug truncate">
                  {applicant.candidate.full_name}
                </h2>
                <div className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-0.5">
                  {applicant.candidate.passport_number && (
                    <span className="inline-flex items-center gap-1 text-[11px] text-muted-foreground">
                      <Hash className="h-3 w-3" />
                      {applicant.candidate.passport_number}
                    </span>
                  )}
                  {applicant.candidate.nationality && (
                    <span className="inline-flex items-center gap-1 text-[11px] text-muted-foreground">
                      <Globe className="h-3 w-3" />
                      {applicant.candidate.nationality}
                    </span>
                  )}
                  {applicant.candidate.age != null && (
                    <span className="inline-flex items-center gap-1 text-[11px] text-muted-foreground">
                      <User className="h-3 w-3" />
                      {applicant.candidate.age} yrs
                    </span>
                  )}
                </div>
              </div>

              {/* Quick actions */}
              <div className="flex items-center gap-1.5 shrink-0">
                <Button size="icon" variant="ghost" className="h-8 w-8 rounded-xl" title="Print">
                  <Printer className="h-3.5 w-3.5" />
                </Button>
                <Button size="icon" variant="ghost" className="h-8 w-8 rounded-xl" title="Download PDF">
                  <Download className="h-3.5 w-3.5" />
                </Button>
              </div>
            </div>

            {/* ── Tabs ── */}
            <Tabs defaultValue="overview" className="flex flex-col flex-1 min-h-0">
              <div className="border-b border-border/60 px-6 bg-background">
                <TabsList className="h-10 bg-transparent p-0 gap-0 rounded-none justify-start">
                  {[
                    { value: "overview",  label: "Overview" },
                    { value: "timeline",  label: "Timeline" },
                    { value: "process",   label: "Process" },
                    { value: "documents", label: "Documents" },
                    { value: "notes",     label: "Notes" },
                  ].map((tab) => (
                    <TabsTrigger
                      key={tab.value}
                      value={tab.value}
                      className={cn(
                        "relative h-10 rounded-none px-4 text-xs font-semibold transition-colors",
                        "data-[state=active]:text-foreground data-[state=active]:shadow-none data-[state=active]:bg-transparent",
                        "data-[state=inactive]:text-muted-foreground data-[state=inactive]:bg-transparent",
                        "after:absolute after:bottom-0 after:left-0 after:right-0 after:h-0.5 after:rounded-full after:bg-primary after:scale-x-0",
                        "data-[state=active]:after:scale-x-100 after:transition-transform after:duration-200",
                      )}
                    >
                      {tab.label}
                    </TabsTrigger>
                  ))}
                </TabsList>
              </div>

              {/* Tab content area */}
              <div className="flex-1 overflow-y-auto">
                <TabsContent value="overview" className="mt-0 p-6 focus-visible:outline-none">
                  <OverviewTab candidate={applicant.candidate} steps={applicant.steps} />
                </TabsContent>

                <TabsContent value="timeline" className="mt-0 p-6 focus-visible:outline-none">
                  <TimelineTab steps={applicant.steps} />
                </TabsContent>

                <TabsContent value="process" className="mt-0 p-6 focus-visible:outline-none">
                  <ProcessStepsTab
                    selectionId={applicant.selectionId}
                    steps={applicant.steps}
                    canEdit={applicant.canEdit}
                    savingSteps={applicant.savingSteps}
                    onUpdate={applicant.onUpdateStatus}
                  />
                </TabsContent>

                <TabsContent value="documents" className="mt-0 p-6 focus-visible:outline-none">
                  <DocumentsTab
                    documents={applicant.candidate.documents ?? []}
                    canEdit={applicant.canEdit}
                  />
                </TabsContent>

                <TabsContent value="notes" className="mt-0 p-6 focus-visible:outline-none">
                  <NotesTab candidateId={applicant.candidateId} canEdit={applicant.canEdit} />
                </TabsContent>
              </div>
            </Tabs>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  );
}
